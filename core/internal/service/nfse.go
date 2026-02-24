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
	nfseMaxSyncLoops = 10
	nfseADNBaseURL   = "https://adn.nfse.gov.br/contribuintes"
)

type NFSeSyncService struct {
	empresaRepo   *repository.EmpresaRepository
	documentoRepo *repository.DocumentoRepository
	storage       DocumentStorage
	log           zerolog.Logger
}

func NewNFSeSyncService(
	empresaRepo *repository.EmpresaRepository,
	documentoRepo *repository.DocumentoRepository,
	storage DocumentStorage,
	log zerolog.Logger,
) *NFSeSyncService {
	return &NFSeSyncService{
		empresaRepo:   empresaRepo,
		documentoRepo: documentoRepo,
		storage:       storage,
		log:           log,
	}
}

func (s *NFSeSyncService) SyncEmpresaNFSe(empresa model.Empresa) error {
	if s.storage == nil {
		return fmt.Errorf("storage not configured")
	}
	if len(empresa.CertificadoPFX) == 0 || empresa.CertificadoSenha == "" {
		return fmt.Errorf("no certificate for empresa %d", empresa.ID)
	}
	if !empresa.NFSeHabilitada {
		return nil
	}

	ctx := context.Background()

	if empresa.SyncState != nil && empresa.SyncState.NFSeBlockedUntil != nil {
		if time.Now().Before(*empresa.SyncState.NFSeBlockedUntil) {
			s.log.Debug().Str("cnpj", empresa.CNPJ).Msg("nfse: skipping, still blocked")
			return nil
		}
	}

	if empresa.SyncState != nil && empresa.SyncState.UltimaSyncNFSe != nil {
		if time.Since(*empresa.SyncState.UltimaSyncNFSe) < time.Hour {
			s.log.Debug().Str("cnpj", empresa.CNPJ).Msg("nfse: skipping, synced recently")
			return nil
		}
	}

	nfseClient, err := client.NewNFSeClient(nfseADNBaseURL, empresa.CertificadoPFX, empresa.CertificadoSenha)
	if err != nil {
		return fmt.Errorf("creating nfse client: %w", err)
	}

	currentNSU := empresa.UltNSUNFSe
	if strings.TrimSpace(currentNSU) == "" || currentNSU == "0" {
		currentNSU = "0"
	}

	totalDocs := 0

	for iteration := 1; iteration <= nfseMaxSyncLoops; iteration++ {
		resp, rawBody, err := nfseClient.DistDFe(ctx, currentNSU)
		if err != nil {
			return fmt.Errorf("nfse distdfe (iteration %d): %w", iteration, err)
		}

		s.log.Debug().
			Int("iteration", iteration).
			Str("cstat", resp.CStat).
			Str("ult_nsu", resp.UltNSU).
			Str("max_nsu", resp.MaxNSU).
			Int("docs", len(resp.Documentos)).
			Msg("nfse adn response")

		if resp.CStat == "" && resp.UltNSU == "" && len(resp.Documentos) == 0 {
			bodyPreview := rawBody
			if len(bodyPreview) > 500 {
				bodyPreview = bodyPreview[:500] + "..."
			}
			s.log.Warn().
				Str("cnpj", empresa.CNPJ).
				Str("raw_body", bodyPreview).
				Msg("nfse: ADN returned empty/unexpected response")
			break
		}

		if resp.CStat == "137" || len(resp.Documentos) == 0 {
			break
		}

		docs := make([]model.DocumentoFiscal, 0, len(resp.Documentos))
		for _, d := range resp.Documentos {
			xmlContent := d.XMLBase64
			if xmlContent == "" {
				continue
			}

			parsed := ParseNFSeXML(xmlContent)

			chave := firstNonEmpty(parsed.ChaveAcesso, d.ChNFSe)
			nsu := d.NSU
			baseName := firstNonEmpty(chave, nsu)
			if baseName == "" {
				baseName = fmt.Sprintf("nfse_%d", time.Now().UnixNano())
			}

			objectKey := s.storage.BuildDocumentKey("nfs-e", parsed.Competencia, empresa.CNPJ, baseName+".xml")
			if err := s.storage.PutObject(ctx, objectKey, "application/xml", []byte(xmlContent)); err != nil {
				s.log.Warn().Err(err).Str("nsu", nsu).Msg("nfse: failed to upload xml")
				continue
			}

			xmlHash := sha256.Sum256([]byte(xmlContent))

			docs = append(docs, model.DocumentoFiscal{
				EmpresaID:        empresa.ID,
				NSU:              nsu,
				ChaveAcesso:      chave,
				TipoDocumento:    "nfs-e",
				StatusDocumento:  parsed.StatusDocumento,
				NumeroDocumento:  parsed.NumeroNFSe,
				EmitenteNome:     parsed.PrestadorNome,
				EmitenteCNPJ:     parsed.PrestadorCNPJ,
				DestinatarioNome: parsed.TomadorNome,
				DestinatarioCNPJ: parsed.TomadorCNPJ,
				Competencia:      parsed.Competencia,
				Schema:           d.Schema,
				XMLObjectKey:     objectKey,
				XMLSHA256:        hex.EncodeToString(xmlHash[:]),
				XMLSizeBytes:     len(xmlContent),
				XMLResumo:        false,
				DataEmissao:      parsed.DataEmissao,
				ValorTotal:       parsed.ValorLiquido,
				ValorProdutos:    parsed.ValorServico,
				SearchText: strings.Join([]string{
					empresa.CNPJ, chave, parsed.NumeroNFSe,
					parsed.PrestadorNome, parsed.PrestadorCNPJ,
					parsed.TomadorNome, parsed.TomadorCNPJ,
				}, " "),
			})
		}

		if len(docs) > 0 {
			if err := s.documentoRepo.UpsertMany(ctx, docs); err != nil {
				s.log.Error().Err(err).Uint("empresa_id", empresa.ID).Msg("nfse: persist docs failed")
			}
			totalDocs += len(docs)
		}

		if resp.UltNSU != "" {
			currentNSU = resp.UltNSU
		}

		_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
			UltNSUNFSe: &currentNSU,
		})

		if resp.UltNSU == resp.MaxNSU || resp.UltNSU == "" {
			break
		}

		sleepWithJitter(loopSleepMinSecs, loopSleepMaxSecs)
	}

	now := time.Now()
	_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
		UltNSUNFSe:     &currentNSU,
		UltimaSyncNFSe: &now,
	})

	s.log.Info().
		Str("cnpj", empresa.CNPJ).
		Int("docs", totalDocs).
		Str("ult_nsu", currentNSU).
		Msg("nfse: sync completed")

	return nil
}
