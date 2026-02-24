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
	loopSleepMinSecs = 2
	loopSleepMaxSecs = 4
	minSyncInterval  = time.Hour
	throttleBase     = time.Hour
	throttleMax      = 4 * time.Hour
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
	return s.syncEmpresa(empresa, false)
}

func (s *SyncService) SyncEmpresaForce(empresa model.Empresa) error {
	return s.syncEmpresa(empresa, true)
}

func (s *SyncService) syncEmpresa(empresa model.Empresa, force bool) error {
	if s.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	ctx := context.Background()
	siglaUF := normalizeSiglaUF(firstNonEmpty(empresa.SiglaUF, empresa.Estado))

	if !force {
		if empresa.SyncState != nil && empresa.SyncState.BlockedUntil != nil {
			if time.Now().Before(*empresa.SyncState.BlockedUntil) {
				s.log.Debug().Str("cnpj", empresa.CNPJ).
					Time("blocked_until", *empresa.SyncState.BlockedUntil).
					Msg("skipping, still blocked by SEFAZ")
				return nil
			}
		}

		if empresa.UltimaSincronizacao != nil {
			if time.Since(*empresa.UltimaSincronizacao) < minSyncInterval {
				s.log.Debug().Str("cnpj", empresa.CNPJ).Msg("skipping, synced recently")
				return nil
			}
		}
	}

	if !force {
		if err := s.rateLimiter.Allow(empresa.CNPJ); err != nil {
			s.log.Warn().Str("cnpj", empresa.CNPJ).Msg("rate limited, skipping sync")
			return nil
		}
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

	// Persist progress on any exit path so we never re-request already-seen NSUs.
	progressSaved := false
	defer func() {
		if progressSaved {
			return
		}
		now := time.Now()
		_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
			UltNSU:              &currentNSU,
			MaxNSU:              &lastMaxNSU,
			UltimaSincronizacao: &now,
			UltimoCStat:         &lastCStat,
			UltimoXMotivo:       &lastXMotivo,
		})
	}()

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
			Str("ult_nsu", currentNSU).
			Str("max_nsu", lastMaxNSU).
			Msg("sefaz response")

		if cStat == "656" {
			blockDuration := s.computeThrottleBlock(empresa)

			s.rateLimiter.MarkThrottled(empresa.CNPJ, int(blockDuration.Seconds()))

			now := time.Now()
			blockedUntil := now.Add(blockDuration)
			progressSaved = true
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

			s.log.Warn().
				Str("cnpj", empresa.CNPJ).
				Dur("block_duration", blockDuration).
				Time("blocked_until", blockedUntil).
				Msg("sefaz throttle (656), backing off")

			return fmt.Errorf("sefaz throttle: %s", xMotivo)
		}

		if cStat == "137" {
			now := time.Now()
			blockedUntil := now.Add(time.Hour)
			progressSaved = true
			_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
				UltNSU:              &currentNSU,
				UltimaSincronizacao: &now,
				BlockedUntil:        &blockedUntil,
				SetBlockedUntil:     true,
				UltimoCStat:         &cStat,
				UltimoXMotivo:       &xMotivo,
			})

			s.log.Info().Str("cnpj", empresa.CNPJ).Msg("no documents available (137), blocked for 1h")
			break
		}

		var docs []model.DocumentoFiscal
		if len(parsed.Documents) > 0 {
			docs = make([]model.DocumentoFiscal, 0, len(parsed.Documents))
		}
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

		// Persist NSU after each successful iteration so progress is never lost.
		_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
			UltNSU:        &currentNSU,
			MaxNSU:        &lastMaxNSU,
			UltimoCStat:   &lastCStat,
			UltimoXMotivo: &lastXMotivo,
		})

		if parsed.UltNSU == parsed.MaxNSU || parsed.UltNSU == "" {
			break
		}

		if iteration < maxSyncLoops {
			sleepWithJitter(loopSleepMinSecs, loopSleepMaxSecs)
		}
	}

	now := time.Now()
	progressSaved = true
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

	s.enviarCienciasPendentes(ctx, empresa, siglaUF)
	s.baixarDocsBloqueados(ctx, empresa, siglaUF, force)

	return nil
}

// enviarCienciasPendentes sends Ciência da Operação for all resNFe summaries with no manifestation yet.
// The full procNFe XML will be returned by SEFAZ on the next distNSU call (no consChNFe needed).
func (s *SyncService) enviarCienciasPendentes(ctx context.Context, empresa model.Empresa, siglaUF string) {
	pendentes, err := s.documentoRepo.ListPendingCiencia(ctx, empresa.ID)
	if err != nil {
		s.log.Warn().Err(err).Uint("empresa_id", empresa.ID).Msg("failed to list pending ciencia")
		return
	}

	if len(pendentes) == 0 {
		return
	}

	s.log.Info().
		Uint("empresa_id", empresa.ID).
		Int("pendentes", len(pendentes)).
		Msg("sending ciencia for pending resNFe")

	for _, doc := range pendentes {
		manifResp, err := s.client.Manifesta(
			ctx,
			empresa.CertificadoPFX,
			empresa.CertificadoSenha,
			empresa.CNPJ,
			empresa.RazaoSocial,
			siglaUF,
			empresa.TpAmb,
			doc.ChaveAcesso,
			"210210",
			"",
		)
		if err != nil {
			s.log.Warn().Err(err).Str("chave", doc.ChaveAcesso).Msg("ciencia request failed")
			continue
		}

		manifCStat := firstNonEmpty(manifResp.CStat)
		s.log.Debug().
			Str("chave", doc.ChaveAcesso).
			Str("cstat", manifCStat).
			Str("xmotivo", manifResp.XMotivo).
			Msg("ciencia response")

		// 128 = lote processado (batch ok), 135 = evento registrado, 573 = duplicidade
		if manifCStat != "128" && manifCStat != "135" && manifCStat != "573" {
			s.log.Warn().
				Str("chave", doc.ChaveAcesso).
				Str("cstat", manifCStat).
				Str("xmotivo", manifResp.XMotivo).
				Msg("ciencia rejected by SEFAZ")
			continue
		}

		_ = s.documentoRepo.UpdateManifestacaoStatus(ctx, doc.ID, "ciencia", time.Now())
		sleepWithJitter(loopSleepMinSecs, loopSleepMaxSecs)
	}
}

// baixarDocsBloqueados is a fallback for docs where ciência was sent but procNFe never arrived
// via distNSU after 2+ hours. Uses consChNFe directly, limited to 20/hour — stops immediately on 656.
func (s *SyncService) baixarDocsBloqueados(ctx context.Context, empresa model.Empresa, siglaUF string, force bool) {
	if !force {
		if empresa.SyncState != nil && empresa.SyncState.DownloadBlockedUntil != nil {
			if time.Now().Before(*empresa.SyncState.DownloadBlockedUntil) {
				s.log.Debug().
					Str("cnpj", empresa.CNPJ).
					Time("blocked_until", *empresa.SyncState.DownloadBlockedUntil).
					Msg("consChNFe downloads blocked, skipping fallback")
				return
			}
		}
	}

	docs, err := s.documentoRepo.ListDocsBloqueadosSemXML(ctx, empresa.ID)
	if err != nil {
		s.log.Warn().Err(err).Uint("empresa_id", empresa.ID).Msg("failed to list stuck docs for fallback download")
		return
	}

	if len(docs) == 0 {
		return
	}

	// Limit fallback downloads to avoid SEFAZ 656 block (max 10 per cycle)
	limit := 10
	if len(docs) > limit {
		s.log.Info().
			Uint("empresa_id", empresa.ID).
			Int("total_docs", len(docs)).
			Int("limit", limit).
			Msg("fallback: limiting downloads to prevent 656 block")
		docs = docs[:limit]
	}

	s.log.Info().
		Uint("empresa_id", empresa.ID).
		Int("docs", len(docs)).
		Msg("fallback: downloading stuck resNFe via consChNFe")

	for _, doc := range docs {
		dlResp, err := s.client.DownloadByKey(
			ctx,
			empresa.CertificadoPFX,
			empresa.CertificadoSenha,
			empresa.CNPJ,
			empresa.RazaoSocial,
			siglaUF,
			empresa.TpAmb,
			doc.ChaveAcesso,
		)
		if err != nil {
			s.log.Warn().Err(err).Str("chave", doc.ChaveAcesso).Msg("fallback download request failed")
			continue
		}

		dlCStat := firstNonEmpty(dlResp.CStat)

		if dlCStat == "656" {
			blockedUntil := time.Now().Add(time.Hour)
			_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
				DownloadBlockedUntil:    &blockedUntil,
				SetDownloadBlockedUntil: true,
			})
			s.log.Warn().
				Str("cnpj", empresa.CNPJ).
				Time("blocked_until", blockedUntil).
				Msg("consChNFe rate limit (656), stopping fallback downloads for 1h")
			break
		}

		if dlCStat != "138" && dlCStat != "140" {
			s.log.Warn().
				Str("chave", doc.ChaveAcesso).
				Str("cstat", dlCStat).
				Str("xmotivo", dlResp.XMotivo).
				Msg("fallback download unexpected cstat")
			continue
		}

		dlParsed, err := ParseDistDFeResponse(dlResp.RawXML)
		if err != nil || len(dlParsed.Documents) == 0 {
			s.log.Warn().Err(err).Str("chave", doc.ChaveAcesso).Msg("fallback: failed to parse downloaded XML")
			continue
		}

		fullDoc := dlParsed.Documents[0]

		// Se a SEFAZ retornou apenas o resumo (resNFe) no fallback, não podemos dar upgrade
		if fullDoc.XMLResumo {
			s.log.Warn().Str("chave", doc.ChaveAcesso).Msg("fallback: received resNFe instead of full XML, skipping upgrade")
			continue
		}

		xmlObjectKey := s.storage.BuildDocumentKey(
			fullDoc.DocumentType,
			fullDoc.Competencia,
			empresa.CNPJ,
			doc.ChaveAcesso+".xml",
		)
		if err := s.storage.PutObject(ctx, xmlObjectKey, "application/xml", []byte(fullDoc.XML)); err != nil {
			s.log.Warn().Err(err).Str("chave", doc.ChaveAcesso).Msg("fallback: failed to upload full xml")
			continue
		}

		xmlHash := sha256.Sum256([]byte(fullDoc.XML))
		now := time.Now()

		if err := s.documentoRepo.UpgradeFromResumo(ctx, doc.ID, model.DocumentoFiscal{
			XMLObjectKey:       xmlObjectKey,
			XMLSHA256:          hex.EncodeToString(xmlHash[:]),
			XMLSizeBytes:       len(fullDoc.XML),
			EmitenteNome:       fullDoc.EmitenteNome,
			EmitenteCNPJ:       fullDoc.EmitenteCNPJ,
			DestinatarioNome:   fullDoc.DestinatarioNome,
			DestinatarioCNPJ:   fullDoc.DestinatarioCNPJ,
			NumeroDocumento:    fullDoc.NumeroDocumento,
			StatusDocumento:    fullDoc.StatusDocumento,
			Schema:             fullDoc.Schema,
			SearchText:         buildDocumentSearchText(empresa.CNPJ, fullDoc),
			ManifestacaoStatus: "ciencia",
			ManifestacaoAt:     &now,
		}); err != nil {
			s.log.Warn().Err(err).Str("chave", doc.ChaveAcesso).Msg("fallback: failed to upgrade doc from resumo")
			continue
		}

		// Clear download block on first successful download
		_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
			SetDownloadBlockedUntil: true,
			DownloadBlockedUntil:    nil,
		})

		s.log.Info().
			Str("chave", doc.ChaveAcesso).
			Str("emitente", fullDoc.EmitenteNome).
			Str("numero", fullDoc.NumeroDocumento).
			Msg("fallback: upgraded resNFe to full XML via consChNFe")

		sleepWithJitter(loopSleepMinSecs, loopSleepMaxSecs)
	}
}

// computeThrottleBlock returns a progressive block duration.
// If the previous cStat was already 656, doubles the base (capped at throttleMax).
func (s *SyncService) computeThrottleBlock(empresa model.Empresa) time.Duration {
	block := throttleBase

	if empresa.SyncState != nil && empresa.SyncState.UltimoCStat == "656" {
		if empresa.SyncState.BlockedUntil != nil {
			prevBlock := empresa.SyncState.BlockedUntil.Sub(empresa.SyncState.UpdatedAt)
			if prevBlock > 0 {
				block = prevBlock * 2
			}
		} else {
			block = throttleBase * 2
		}
	}

	if block > throttleMax {
		block = throttleMax
	}
	if block < throttleBase {
		block = throttleBase
	}

	return block
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


