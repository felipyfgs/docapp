package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/model"
	"docapp/core/internal/repository"

	"github.com/rs/zerolog"
)

const (
	maxSyncLoops     = 12
	loopSleepSeconds = 2
	minSyncInterval  = time.Hour
)

type SyncService struct {
	empresaRepo   *repository.EmpresaRepository
	documentoRepo *repository.DocumentoRepository
	client        *client.Client
	storage       DocumentStorage
	log           zerolog.Logger
	rateLimiter   *RateLimiter
}

func NewSyncService(empresaRepo *repository.EmpresaRepository, documentoRepo *repository.DocumentoRepository, c *client.Client, storage DocumentStorage, log zerolog.Logger) *SyncService {
	return &SyncService{
		empresaRepo:   empresaRepo,
		documentoRepo: documentoRepo,
		client:        c,
		storage:       storage,
		log:           log,
		rateLimiter:   NewRateLimiter(20),
	}
}

func (s *SyncService) SyncEmpresa(empresa model.Empresa) error {
	if s.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	ctx := context.Background()
	siglaUF := normalizeSiglaUF(firstNonEmpty(empresa.SiglaUF, empresa.Estado))

	if empresa.UltimaSincronizacao != nil {
		if time.Since(*empresa.UltimaSincronizacao) < minSyncInterval {
			s.log.Debug().Str("cnpj", empresa.CNPJ).Msg("skipping, synced recently")
			return nil
		}
	}

	if err := s.rateLimiter.Allow(empresa.CNPJ); err != nil {
		s.log.Warn().Str("cnpj", empresa.CNPJ).Msg("rate limited, skipping sync")
		return nil
	}

	if len(empresa.CertificadoPFX) == 0 || empresa.CertificadoSenha == "" {
		s.log.Warn().Str("cnpj", empresa.CNPJ).Msg("no certificate configured, skipping")
		return nil
	}

	if siglaUF == "" {
		s.log.Warn().Str("cnpj", empresa.CNPJ).Msg("no UF configured, skipping")
		return nil
	}

	if strings.TrimSpace(empresa.SiglaUF) == "" {
		if err := s.empresaRepo.UpdateCertificadoUF(ctx, empresa.ID, siglaUF); err != nil {
			s.log.Warn().Err(err).Uint("empresa_id", empresa.ID).Msg("failed to persist fallback sigla_uf")
		} else {
			empresa.SiglaUF = siglaUF
		}
	}

	currentNSU := empresa.UltNSU
	if strings.TrimSpace(currentNSU) == "" {
		currentNSU = "000000000000000"
	}

	cutoff := time.Now().AddDate(0, 0, -empresa.LookbackDays)
	totalDocs := 0
	lastCStat := ""
	lastXMotivo := ""
	lastMaxNSU := ""

	for iteration := 1; iteration <= maxSyncLoops; iteration++ {
		resp, err := s.client.DistDFe(
			ctx,
			empresa.CertificadoPFX,
			empresa.CertificadoSenha,
			empresa.CNPJ,
			empresa.RazaoSocial,
			siglaUF,
			empresa.TpAmb,
			currentNSU,
		)
		if err != nil {
			if strings.Contains(err.Error(), "throttled") && resp != nil && resp.RetryAfter > 0 {
				s.rateLimiter.MarkThrottled(empresa.CNPJ, resp.RetryAfter)
			}
			return fmt.Errorf("calling distdfe (iteration %d): %w", iteration, err)
		}

		parsed, err := ParseDistDFeResponse(resp.RawXML)
		if err != nil {
			return fmt.Errorf("parsing sefaz response: %w", err)
		}

		if parsed.UltNSU != "" {
			currentNSU = parsed.UltNSU
		}

		cStat := firstNonEmpty(resp.CStat, parsed.CStat)
		xMotivo := firstNonEmpty(resp.XMotivo, parsed.XMotivo)
		lastCStat = cStat
		lastXMotivo = xMotivo
		if parsed.MaxNSU != "" {
			lastMaxNSU = parsed.MaxNSU
		}

		s.log.Debug().
			Int("iteration", iteration).
			Str("cstat", cStat).
			Str("xmotivo", xMotivo).
			Msg("sefaz response")

		if cStat == "656" {
			retryAfter := resp.RetryAfter
			if retryAfter <= 0 {
				retryAfter = int(time.Hour.Seconds())
			}

			s.rateLimiter.MarkThrottled(empresa.CNPJ, retryAfter)

			now := time.Now()
			blockedUntil := now.Add(time.Duration(retryAfter) * time.Second)
			if err := s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
				UltNSU:              &currentNSU,
				UltimaSincronizacao: &now,
				BlockedUntil:        &blockedUntil,
				SetBlockedUntil:     true,
				UltimoCStat:         &cStat,
				UltimoXMotivo:       &xMotivo,
			}); err != nil {
				s.log.Warn().Err(err).Uint("empresa_id", empresa.ID).Msg("failed to persist sync state after throttle")
			}

			return fmt.Errorf("sefaz throttle: %s", xMotivo)
		}

		if cStat == "137" {
			now := time.Now()
			blockedUntil := now.Add(time.Hour)
			_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
				UltNSU:              &currentNSU,
				UltimaSincronizacao: &now,
				BlockedUntil:        &blockedUntil,
				SetBlockedUntil:     true,
				UltimoCStat:         &cStat,
				UltimoXMotivo:       &xMotivo,
			})

			s.log.Info().Str("cnpj", empresa.CNPJ).Msg("no documents available")
			break
		}

		var docs []model.DocumentoFiscal
		for _, d := range parsed.Documents {
			if d.DataEmissao != nil && d.DataEmissao.Before(cutoff) {
				continue
			}

			xmlObjectKey := ""
			baseName := firstNonEmpty(d.ChaveAcesso, d.NSU)
			if baseName == "" {
				baseName = fmt.Sprintf("doc_%d", time.Now().UnixNano())
			}

			xmlObjectKey = s.storage.BuildDocumentKey(d.DocumentType, d.Competencia, empresa.CNPJ, baseName+".xml")
			if err := s.storage.PutObject(ctx, xmlObjectKey, "application/xml", []byte(d.XML)); err != nil {
				s.log.Warn().Err(err).
					Uint("empresa_id", empresa.ID).
					Str("nsu", d.NSU).
					Str("chave", d.ChaveAcesso).
					Msg("failed to upload xml to storage")
				continue
			}

			xmlHash := sha256.Sum256([]byte(d.XML))
			searchText := buildDocumentSearchText(empresa.CNPJ, d)

			docs = append(docs, model.DocumentoFiscal{
				EmpresaID:        empresa.ID,
				NSU:              d.NSU,
				ChaveAcesso:      d.ChaveAcesso,
				TipoDocumento:    d.DocumentType,
				StatusDocumento:  d.StatusDocumento,
				NumeroDocumento:  d.NumeroDocumento,
				EmitenteNome:     d.EmitenteNome,
				EmitenteCNPJ:     d.EmitenteCNPJ,
				DestinatarioNome: d.DestinatarioNome,
				DestinatarioCNPJ: d.DestinatarioCNPJ,
				Competencia:      d.Competencia,
				Schema:           d.Schema,
				XMLObjectKey:     xmlObjectKey,
				XMLSHA256:        hex.EncodeToString(xmlHash[:]),
				XMLSizeBytes:     len(d.XML),
				XMLResumo:        d.XMLResumo,
				DataEmissao:      d.DataEmissao,
				SearchText:       searchText,
			})
		}

		if len(docs) > 0 {
			if err := s.documentoRepo.UpsertMany(ctx, docs); err != nil {
				s.log.Error().Err(err).Uint("empresa_id", empresa.ID).Msg("persist documentos failed")
			}
			totalDocs += len(docs)
		}

		if parsed.UltNSU == parsed.MaxNSU || parsed.UltNSU == "" {
			break
		}

		if iteration < maxSyncLoops {
			time.Sleep(loopSleepSeconds * time.Second)
		}
	}

	now := time.Now()
	if err := s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
		UltNSU:              &currentNSU,
		MaxNSU:              &lastMaxNSU,
		UltimaSincronizacao: &now,
		BlockedUntil:        nil,
		SetBlockedUntil:     true,
		UltimoCStat:         &lastCStat,
		UltimoXMotivo:       &lastXMotivo,
	}); err != nil {
		return fmt.Errorf("updating empresa sync state: %w", err)
	}

	s.log.Info().
		Uint("empresa_id", empresa.ID).
		Str("cnpj", empresa.CNPJ).
		Int("documentos", totalDocs).
		Str("ult_nsu", currentNSU).
		Msg("sync completed")

	return nil
}

func buildDocumentSearchText(empresaCNPJ string, d Document) string {
	parts := []string{
		empresaCNPJ,
		d.ChaveAcesso,
		d.NSU,
		d.NumeroDocumento,
		d.EmitenteNome,
		d.EmitenteCNPJ,
		d.DestinatarioNome,
		d.DestinatarioCNPJ,
		d.DocumentType,
		d.StatusDocumento,
		d.Competencia,
	}

	b := strings.Builder{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strings.ToLower(trimmed))
	}

	return b.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func normalizeSiglaUF(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	upper := strings.ToUpper(trimmed)
	if len(upper) > 2 {
		return upper[:2]
	}

	return upper
}
