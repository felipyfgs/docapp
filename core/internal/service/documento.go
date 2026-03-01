package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/model"
	"docapp/core/internal/repository"

	"github.com/rs/zerolog"
)

const (
	ExportFormatXML   = "xml"
	ExportFormatDanfe = "danfe"
	ExportFormatBoth  = "ambos"

	ExportDeliveryProxy     = "proxy"
	ExportDeliveryPresigned = "presigned"
)

type DocumentoListFilter struct {
	Search     string
	Tipo       string
	Status     string
	EmpresaID  uint
	XMLResumo  *bool
	DataInicio *time.Time
	DataFim    *time.Time
	Page       int
	PageSize   int
}

type DocumentoExportOptions struct {
	IDs          []uint
	Format       string
	Organization string
	DeliveryMode string
}

type DocumentoExportResult struct {
	FileName     string `json:"file_name"`
	Mode         string `json:"mode"`
	PresignedURL string `json:"presigned_url,omitempty"`
	Content      []byte `json:"-"`
	Total        int    `json:"total"`
	XMLCount     int    `json:"xml_count"`
	DanfeCount   int    `json:"danfe_count"`
	SkippedDanfe int    `json:"skipped_danfe"`
}

type DocumentoBackfillResult struct {
	Processed int `json:"processed"`
	Uploaded  int `json:"uploaded"`
	Skipped   int `json:"skipped"`
}

type DocumentoService struct {
	repo     *repository.DocumentoRepository
	itemRepo *repository.DocumentoItemRepository
	storage  DocumentStorage
	client   *client.Client
	log      zerolog.Logger
}

func NewDocumentoService(repo *repository.DocumentoRepository, itemRepo *repository.DocumentoItemRepository, storage DocumentStorage, c *client.Client, log zerolog.Logger) *DocumentoService {
	return &DocumentoService{repo: repo, itemRepo: itemRepo, storage: storage, client: c, log: log}
}

func (s *DocumentoService) List(filter DocumentoListFilter) ([]model.DocumentoFiscal, int64, error) {
	return s.repo.List(context.Background(), repository.DocumentoListFilter{
		Search:     filter.Search,
		Tipo:       filter.Tipo,
		Status:     filter.Status,
		EmpresaID:  filter.EmpresaID,
		XMLResumo:  filter.XMLResumo,
		DataInicio: filter.DataInicio,
		DataFim:    filter.DataFim,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	})
}

func (s *DocumentoService) GetByID(id uint) (*model.DocumentoFiscal, error) {
	return s.repo.GetByID(context.Background(), id)
}

func (s *DocumentoService) ReadXML(ctx context.Context, doc *model.DocumentoFiscal) (string, error) {
	if doc == nil {
		return "", errors.New("documento inválido")
	}

	if s.storage == nil {
		return "", errors.New("storage indisponível")
	}

	if strings.TrimSpace(doc.XMLObjectKey) == "" {
		return "", errors.New("xml object key não encontrado")
	}

	content, err := s.storage.GetObject(ctx, doc.XMLObjectKey)
	if err != nil {
		s.log.Warn().Err(err).Uint("documento_id", doc.ID).Str("object_key", doc.XMLObjectKey).Msg("failed to read xml from storage")
		return "", errors.New("xml não encontrado")
	}

	return string(content), nil
}

func (s *DocumentoService) Export(ctx context.Context, opts DocumentoExportOptions) (*DocumentoExportResult, error) {
	if len(opts.IDs) == 0 {
		return nil, errors.New("nenhum documento selecionado")
	}

	format := normalizeExportFormat(opts.Format)
	if format == "" {
		return nil, errors.New("formato de exportação inválido")
	}

	organization := normalizeOrganization(opts.Organization)
	delivery := normalizeDeliveryMode(opts.DeliveryMode)

	docs, err := s.repo.ListByIDs(ctx, opts.IDs)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return nil, errors.New("documentos não encontrados")
	}

	now := time.Now()
	fileName := buildExportFileName(organization, docs, now)

	buf := bytes.NewBuffer(nil)
	zipWriter := zip.NewWriter(buf)

	xmlCount := 0
	danfeCount := 0
	skippedDanfe := 0
	skippedReasons := make([]string, 0)

	for _, doc := range docs {
		xmlContent, err := s.ReadXML(ctx, &doc)
		if err != nil {
			skippedReasons = append(skippedReasons, fmt.Sprintf("[%d] XML indisponível", doc.ID))
			continue
		}

		if format == ExportFormatXML || format == ExportFormatBoth {
			xmlPath := buildExportPath(organization, doc, "xml")
			if err := writeZipFile(zipWriter, xmlPath, []byte(xmlContent)); err != nil {
				return nil, err
			}
			xmlCount++
		}

		if format == ExportFormatDanfe || format == ExportFormatBoth {
			if doc.XMLResumo {
				skippedDanfe++
				skippedReasons = append(skippedReasons, fmt.Sprintf("[%d] DANFE ignorado: documento em resumo", doc.ID))
				continue
			}

			if !supportsDanfe(doc.TipoDocumento) {
				skippedDanfe++
				skippedReasons = append(skippedReasons, fmt.Sprintf("[%d] DANFE ignorado: tipo %s não suportado", doc.ID, doc.TipoDocumento))
				continue
			}

			pdfContent, err := s.loadOrGenerateDanfe(ctx, &doc, xmlContent)
			if err != nil {
				skippedDanfe++
				skippedReasons = append(skippedReasons, fmt.Sprintf("[%d] DANFE erro: %s", doc.ID, err.Error()))
				continue
			}

			pdfPath := buildExportPath(organization, doc, "pdf")
			if err := writeZipFile(zipWriter, pdfPath, pdfContent); err != nil {
				return nil, err
			}
			danfeCount++
		}
	}

	if len(skippedReasons) > 0 {
		report := strings.Join(skippedReasons, "\n") + "\n"
		if err := writeZipFile(zipWriter, "_relatorio_exportacao.txt", []byte(report)); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	result := &DocumentoExportResult{
		FileName:     fileName,
		Mode:         delivery,
		Total:        len(docs),
		XMLCount:     xmlCount,
		DanfeCount:   danfeCount,
		SkippedDanfe: skippedDanfe,
	}

	if delivery == ExportDeliveryPresigned && s.storage != nil {
		exportKey := s.storage.BuildDocumentKey("exports", now.Format("2006/01"), "lote", fileName)
		if err := s.storage.PutObject(ctx, exportKey, "application/zip", buf.Bytes()); err != nil {
			return nil, err
		}

		url, err := s.storage.PresignGetObject(ctx, exportKey)
		if err != nil {
			return nil, err
		}

		result.PresignedURL = url
		return result, nil
	}

	result.Content = buf.Bytes()
	return result, nil
}

func (s *DocumentoService) Backfill(ctx context.Context, limit int) (*DocumentoBackfillResult, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}

	docs, err := s.repo.ListSemValor(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("listing docs sem valor: %w", err)
	}

	result := &DocumentoBackfillResult{}

	for i := range docs {
		doc := &docs[i]
		result.Processed++

		if s.storage == nil || strings.TrimSpace(doc.XMLObjectKey) == "" {
			result.Skipped++
			continue
		}

		xmlBytes, err := s.storage.GetObject(ctx, doc.XMLObjectKey)
		if err != nil {
			s.log.Warn().Err(err).Uint("id", doc.ID).Msg("backfill: failed to read xml from storage")
			result.Skipped++
			continue
		}

		xmlContent := string(xmlBytes)
		valorTotal := extractValorTotal(xmlContent)
		valorProdutos := extractValorProdutos(xmlContent)

		if valorTotal == 0 && valorProdutos == 0 {
			result.Skipped++
			continue
		}

		if err := s.repo.UpdateValores(ctx, doc.ID, valorTotal, valorProdutos); err != nil {
			s.log.Warn().Err(err).Uint("id", doc.ID).Msg("backfill: failed to update valores")
			result.Skipped++
			continue
		}

		result.Uploaded++
	}

	return result, nil
}

func (s *DocumentoService) BackfillItens(ctx context.Context, limit int) (*DocumentoBackfillResult, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}

	docs, err := s.repo.ListSemItens(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("listing docs sem itens: %w", err)
	}

	result := &DocumentoBackfillResult{}

	for i := range docs {
		doc := &docs[i]
		result.Processed++

		if s.storage == nil || strings.TrimSpace(doc.XMLObjectKey) == "" {
			result.Skipped++
			continue
		}

		xmlBytes, err := s.storage.GetObject(ctx, doc.XMLObjectKey)
		if err != nil {
			s.log.Warn().Err(err).Uint("id", doc.ID).Msg("backfill-itens: failed to read xml")
			result.Skipped++
			continue
		}

		itens := ExtractItens(string(xmlBytes))
		if len(itens) == 0 {
			result.Skipped++
			continue
		}

		if err := s.itemRepo.UpsertItens(ctx, doc.ID, itens); err != nil {
			s.log.Warn().Err(err).Uint("id", doc.ID).Msg("backfill-itens: failed to upsert")
			result.Skipped++
			continue
		}

		result.Uploaded++
	}

	return result, nil
}

func (s *DocumentoService) ListItens(ctx context.Context, documentoID uint) ([]model.DocumentoItem, error) {
	return s.itemRepo.ListByDocumentoID(ctx, documentoID)
}

func (s *DocumentoService) loadOrGenerateDanfe(ctx context.Context, doc *model.DocumentoFiscal, xmlContent string) ([]byte, error) {
	if doc == nil {
		return nil, errors.New("documento inválido")
	}

	if s.storage != nil && strings.TrimSpace(doc.DanfeObjectKey) != "" {
		content, err := s.storage.GetObject(ctx, doc.DanfeObjectKey)
		if err == nil {
			return content, nil
		}
	}

	pdf, _, err := s.client.GenerateDanfe(ctx, doc.TipoDocumento, xmlContent)
	if err != nil {
		return nil, err
	}

	if s.storage != nil {
		fileBase := firstNonEmpty(doc.ChaveAcesso, doc.NSU, fmt.Sprintf("doc_%d", doc.ID))
		empresaCNPJ := ""
		if doc.Empresa != nil {
			empresaCNPJ = doc.Empresa.CNPJ
		}
		cnpj := firstNonEmpty(empresaCNPJ, doc.DestinatarioCNPJ, doc.EmitenteCNPJ)
		objectKey := s.storage.BuildDocumentKey(doc.TipoDocumento, firstNonEmpty(doc.Competencia, "sem_competencia"), cnpj, fileBase+".pdf")

		if err := s.storage.PutObject(ctx, objectKey, "application/pdf", pdf); err == nil {
			now := time.Now()
			_ = s.repo.UpdateDanfeMetadata(ctx, doc.ID, objectKey, now)
			doc.DanfeObjectKey = objectKey
			doc.DanfeGeneratedAt = &now
		}
	}

	return pdf, nil
}

func normalizeExportFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case ExportFormatXML:
		return ExportFormatXML
	case ExportFormatDanfe:
		return ExportFormatDanfe
	case ExportFormatBoth, "both":
		return ExportFormatBoth
	default:
		return ""
	}
}

func normalizeDeliveryMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ExportDeliveryPresigned:
		return ExportDeliveryPresigned
	default:
		return ExportDeliveryProxy
	}
}

func normalizeOrganization(organization string) string {
	value := strings.ToLower(strings.TrimSpace(organization))
	switch value {
	case "cnpj/ano/mes", "ano/mes/cnpj":
		return value
	default:
		return "cnpj/ano/mes"
	}
}

func buildExportFileName(organization string, docs []model.DocumentoFiscal, now time.Time) string {
	cnpjSet := make(map[string]struct{})
	for _, doc := range docs {
		c := ""
		if doc.Empresa != nil {
			c = doc.Empresa.CNPJ
		}
		if c == "" {
			c = firstNonEmpty(doc.DestinatarioCNPJ, doc.EmitenteCNPJ)
		}
		if c != "" {
			cnpjSet[c] = struct{}{}
		}
	}

	cnpjPart := "DIVERSOS"
	if len(cnpjSet) == 1 {
		for k := range cnpjSet {
			cnpjPart = k
		}
	}

	ano := now.Format("2006")
	mes := strings.ToUpper(monthNames[now.Format("01")])
	if mes == "" {
		mes = now.Format("01")
	}

	switch organization {
	case "ano/mes/cnpj":
		return fmt.Sprintf("%s_%s_%s.zip", ano, mes, cnpjPart)
	default:
		return fmt.Sprintf("%s_%s_%s.zip", cnpjPart, ano, mes)
	}
}

var monthNames = map[string]string{
	"01": "JANEIRO", "02": "FEVEREIRO", "03": "MARCO", "04": "ABRIL",
	"05": "MAIO", "06": "JUNHO", "07": "JULHO", "08": "AGOSTO",
	"09": "SETEMBRO", "10": "OUTUBRO", "11": "NOVEMBRO", "12": "DEZEMBRO",
}

func competenciaToAnoParts(competencia string) (string, string) {
	parts := strings.SplitN(competencia, "-", 2)
	if len(parts) != 2 {
		return "SEM_ANO", "SEM_MES"
	}
	ano := parts[0]
	mes := monthNames[parts[1]]
	if mes == "" {
		mes = "SEM_MES"
	}
	return ano, mes
}

func buildExportPath(organization string, doc model.DocumentoFiscal, extension string) string {
	competencia := firstNonEmpty(doc.Competencia, "sem_competencia")
	ano, mes := competenciaToAnoParts(competencia)

	empresaCNPJ := ""
	if doc.Empresa != nil {
		empresaCNPJ = doc.Empresa.CNPJ
	}
	cnpj := strings.ToUpper(sanitizePathPart(firstNonEmpty(empresaCNPJ, doc.DestinatarioCNPJ, doc.EmitenteCNPJ, "sem_cnpj")))

	var dir string
	switch organization {
	case "ano/mes/cnpj":
		dir = path.Join(ano, mes, cnpj)
	default:
		dir = path.Join(cnpj, ano, mes)
	}

	fileBase := strings.ToUpper(sanitizeFileName(firstNonEmpty(doc.ChaveAcesso, doc.NumeroDocumento, doc.NSU, fmt.Sprintf("doc_%d", doc.ID))))
	return path.Join(dir, fileBase+"."+strings.ToUpper(extension))
}

func writeZipFile(zipWriter *zip.Writer, filePath string, content []byte) error {
	entry, err := zipWriter.Create(filePath)
	if err != nil {
		return err
	}

	if _, err := entry.Write(content); err != nil {
		return err
	}

	return nil
}

func supportsDanfe(tipo string) bool {
	switch strings.ToLower(strings.TrimSpace(tipo)) {
	case "nf-e", "nfc-e", "ct-e":
		return true
	default:
		return false
	}
}
