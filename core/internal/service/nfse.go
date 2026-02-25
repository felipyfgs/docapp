package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/model"
	"docapp/core/internal/repository"

	"github.com/rs/zerolog"
)

const nfseMaxSyncLoops = 10

type NFSeSyncService struct {
	empresaRepo   *repository.EmpresaRepository
	documentoRepo *repository.DocumentoRepository
	storage       DocumentStorage
	log           zerolog.Logger
	adnBaseURL    string
}

func NewNFSeSyncService(
	empresaRepo *repository.EmpresaRepository,
	documentoRepo *repository.DocumentoRepository,
	storage DocumentStorage,
	log zerolog.Logger,
	adnBaseURL string,
) *NFSeSyncService {
	return &NFSeSyncService{
		empresaRepo:   empresaRepo,
		documentoRepo: documentoRepo,
		storage:       storage,
		log:           log,
		adnBaseURL:    adnBaseURL,
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

	nfseClient, err := client.NewNFSeClient(s.adnBaseURL, empresa.CertificadoPFX, empresa.CertificadoSenha)
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
			Str("status", resp.StatusProcessamento).
			Int("docs", len(resp.LoteDFe)).
			Msg("nfse adn response")

		if !resp.HasDocuments() {
			if resp.StatusProcessamento == "NENHUM_DOCUMENTO_LOCALIZADO" {
				s.log.Info().Str("cnpj", empresa.CNPJ).Msg("nfse: no more documents at ADN")
			} else if resp.StatusProcessamento == "" {
				bodyPreview := rawBody
				if len(bodyPreview) > 500 {
					bodyPreview = bodyPreview[:500] + "..."
				}
				s.log.Warn().
					Str("cnpj", empresa.CNPJ).
					Str("raw_body", bodyPreview).
					Msg("nfse: ADN returned empty/unexpected response")
			}
			break
		}

		var maxNSU int
		docs := make([]model.DocumentoFiscal, 0, len(resp.LoteDFe))
		for _, d := range resp.LoteDFe {
			if d.NSU > maxNSU {
				maxNSU = d.NSU
			}

			xmlContent, err := decompressGzipBase64(d.ArquivoXml)
			if err != nil {
				s.log.Warn().Err(err).Int("nsu", d.NSU).Msg("nfse: failed to decompress xml")
				continue
			}
			if xmlContent == "" {
				continue
			}

			if d.TipoDocumento == "EVENTO" {
				chave := extractTagValue(xmlContent, "chNFSe")
				if chave != "" {
					if newStatus := NFSeStatusFromEvento(xmlContent); newStatus != "" {
						if err := s.documentoRepo.UpdateStatusByChave(ctx, empresa.ID, chave, newStatus); err != nil {
							s.log.Warn().Err(err).Str("chave", chave).Str("status", newStatus).Msg("nfse: failed to apply event status")
						} else {
							s.log.Info().Str("chave", chave).Str("status", newStatus).Msg("nfse: applied event to document")
						}
					}
				}
				continue
			}

			parsed := ParseNFSeXML(xmlContent)

			chave := firstNonEmpty(parsed.ChaveAcesso, d.ChaveAcesso)
			nsu := strconv.Itoa(d.NSU)
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
				Schema:           d.TipoDocumento,
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

		if maxNSU > 0 {
			currentNSU = strconv.Itoa(maxNSU)
		}

		_ = s.empresaRepo.UpdateSyncState(ctx, empresa.ID, repository.SyncStatePatch{
			UltNSUNFSe: &currentNSU,
		})

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

func decompressGzipBase64(data string) (string, error) {
	if data == "" {
		return "", nil
	}

	compressed, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return data, nil
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("reading gzip data: %w", err)
	}

	return string(decompressed), nil
}
