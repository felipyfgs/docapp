package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"docapp/core/internal/model"
	"docapp/core/internal/repository"

	"github.com/rs/zerolog"
)

const importBatchSize = 100

type ImportFile struct {
	Filename string
	Content  []byte
}

type ImportResult struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

type AutoImportResult struct {
	ByEmpresa map[string]ImportResult `json:"by_empresa"` // keyed by empresa razao_social
	Unknown   int                     `json:"unknown"`    // files with no matching empresa
}

type ImportService struct {
	documentoRepo *repository.DocumentoRepository
	empresaRepo   *repository.EmpresaRepository
	storage       DocumentStorage
	log           zerolog.Logger
}

func NewImportService(documentoRepo *repository.DocumentoRepository, empresaRepo *repository.EmpresaRepository, storage DocumentStorage, log zerolog.Logger) *ImportService {
	return &ImportService{
		documentoRepo: documentoRepo,
		empresaRepo:   empresaRepo,
		storage:       storage,
		log:           log,
	}
}

// ExtractFiles unpacks a ZIP or returns the single XML as a slice.
func ExtractFiles(filename string, content []byte) ([]ImportFile, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".zip" {
		r, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
		if err != nil {
			return nil, fmt.Errorf("opening zip: %w", err)
		}
		var files []ImportFile
		for _, f := range r.File {
			if f.FileInfo().IsDir() {
				continue
			}
			if strings.ToLower(filepath.Ext(f.Name)) != ".xml" {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				continue
			}
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(rc)
			rc.Close()
			files = append(files, ImportFile{Filename: f.Name, Content: buf.Bytes()})
		}
		return files, nil
	}

	if ext == ".xml" {
		return []ImportFile{{Filename: filename, Content: content}}, nil
	}

	return nil, fmt.Errorf("unsupported file type: %s", ext)
}

func (s *ImportService) ImportDocumentos(ctx context.Context, empresa model.Empresa, files []ImportFile) ImportResult {
	result := ImportResult{}
	var batch []model.DocumentoFiscal

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := s.documentoRepo.UpsertMany(ctx, batch); err != nil {
			s.log.Error().Err(err).Int("batch_size", len(batch)).Msg("import: upsert batch failed")
			result.Failed += len(batch)
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Imported += len(batch)
		}
		batch = batch[:0]
	}

	for _, f := range files {
		xmlContent := string(f.Content)

		nsu := NSUFromFilename(f.Filename)
		doc := ParseNFeProcXML(xmlContent, nsu)

		if doc.ChaveAcesso == "" && nsu == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: no chave_acesso found", f.Filename))
			continue
		}

		// Use deterministic NSU derived from chave when not in filename
		if nsu == "" {
			nsu = nsuFromChave(doc.ChaveAcesso)
			doc.NSU = nsu
		}

		baseName := doc.ChaveAcesso
		if baseName == "" {
			baseName = nsu
		}

		xmlObjectKey := ""
		if s.storage != nil {
			xmlObjectKey = s.storage.BuildDocumentKey(doc.DocumentType, doc.Competencia, empresa.CNPJ, baseName+".xml")
			if err := s.storage.PutObject(ctx, xmlObjectKey, "application/xml", f.Content); err != nil {
				s.log.Warn().Err(err).Str("chave", doc.ChaveAcesso).Msg("import: failed to upload xml")
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("%s: storage error: %v", f.Filename, err))
				continue
			}
		}

		xmlHash := sha256.Sum256(f.Content)
		searchText := buildDocumentSearchText(empresa.CNPJ, doc)

		batch = append(batch, model.DocumentoFiscal{
			EmpresaID:        empresa.ID,
			NSU:              doc.NSU,
			ChaveAcesso:      doc.ChaveAcesso,
			TipoDocumento:    doc.DocumentType,
			StatusDocumento:  doc.StatusDocumento,
			NumeroDocumento:  doc.NumeroDocumento,
			EmitenteNome:     doc.EmitenteNome,
			EmitenteCNPJ:     doc.EmitenteCNPJ,
			DestinatarioNome: doc.DestinatarioNome,
			DestinatarioCNPJ: doc.DestinatarioCNPJ,
			Competencia:      doc.Competencia,
			Schema:           doc.Schema,
			XMLObjectKey:     xmlObjectKey,
			XMLSHA256:        hex.EncodeToString(xmlHash[:]),
			XMLSizeBytes:     len(f.Content),
			XMLResumo:        false,
			DataEmissao:      doc.DataEmissao,
			SearchText:       searchText,
		})

		if len(batch) >= importBatchSize {
			flush()
		}
	}

	flush()
	return result
}

// ImportDocumentosAuto detects each XML's empresa by CNPJ (emitente first, then destinatário)
// and imports each group. Supports multiple files extracted from ZIPs.
func (s *ImportService) ImportDocumentosAuto(ctx context.Context, files []ImportFile) AutoImportResult {
	// empresaCache avoids repeated DB lookups for the same CNPJ
	empresaCache := map[string]*model.Empresa{}
	lookupEmpresa := func(cnpj string) *model.Empresa {
		if cnpj == "" {
			return nil
		}
		if e, ok := empresaCache[cnpj]; ok {
			return e
		}
		e, err := s.empresaRepo.FindByCNPJ(ctx, cnpj)
		if err != nil {
			empresaCache[cnpj] = nil
			return nil
		}
		empresaCache[cnpj] = e
		return e
	}

	// Group files by empresa_id
	type group struct {
		empresa model.Empresa
		files   []ImportFile
	}
	grouped := map[uint]*group{}
	unknown := 0

	for _, f := range files {
		xmlContent := string(f.Content)
		chave := ""

		// Try to get chave from filename first (fastest), then from XML
		nsu := NSUFromFilename(f.Filename)
		doc := ParseNFeProcXML(xmlContent, nsu)
		chave = doc.ChaveAcesso

		// Detect empresa: emitente CNPJ from chave → destinatário CNPJ from XML
		var empresa *model.Empresa
		if cnpj := CNPJFromChave(chave); cnpj != "" {
			empresa = lookupEmpresa(cnpj)
		}
		if empresa == nil && doc.DestinatarioCNPJ != "" {
			empresa = lookupEmpresa(doc.DestinatarioCNPJ)
		}
		if empresa == nil && doc.EmitenteCNPJ != "" {
			empresa = lookupEmpresa(doc.EmitenteCNPJ)
		}

		if empresa == nil {
			s.log.Warn().Str("filename", f.Filename).Str("chave", chave).Msg("import auto: empresa not found")
			unknown++
			continue
		}

		if g, ok := grouped[empresa.ID]; ok {
			g.files = append(g.files, f)
		} else {
			grouped[empresa.ID] = &group{empresa: *empresa, files: []ImportFile{f}}
		}
	}

	result := AutoImportResult{
		ByEmpresa: make(map[string]ImportResult),
		Unknown:   unknown,
	}

	for _, g := range grouped {
		r := s.ImportDocumentos(ctx, g.empresa, g.files)
		result.ByEmpresa[g.empresa.RazaoSocial] = r
	}

	return result
}
