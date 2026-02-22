package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/model"
	"docapp/core/internal/sefaz"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxSyncLoops     = 12
	loopSleepSeconds = 2
	minSyncInterval  = time.Hour
)

type SyncService struct {
	db          *gorm.DB
	client      *client.Client
	log         zerolog.Logger
	rateLimiter *sefaz.RateLimiter
}

func NewSyncService(db *gorm.DB, c *client.Client, log zerolog.Logger) *SyncService {
	return &SyncService{
		db:          db,
		client:      c,
		log:         log,
		rateLimiter: sefaz.NewRateLimiter(20),
	}
}

func (s *SyncService) SyncEmpresa(empresa model.Empresa) error {
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

	if empresa.SiglaUF == "" {
		s.log.Warn().Str("cnpj", empresa.CNPJ).Msg("no UF configured, skipping")
		return nil
	}

	currentNSU := empresa.UltNSU
	cutoff := time.Now().AddDate(0, 0, -empresa.LookbackDays)
	totalDocs := 0

	for iteration := 1; iteration <= maxSyncLoops; iteration++ {
		resp, err := s.client.DistDFe(
			context.Background(),
			empresa.CertificadoPFX,
			empresa.CertificadoSenha,
			empresa.CNPJ,
			empresa.RazaoSocial,
			empresa.SiglaUF,
			empresa.TpAmb,
			currentNSU,
		)
		if err != nil {
			if strings.Contains(err.Error(), "throttled") && resp != nil && resp.RetryAfter > 0 {
				s.rateLimiter.MarkThrottled(empresa.CNPJ, resp.RetryAfter)
			}
			return fmt.Errorf("calling distdfe (iteration %d): %w", iteration, err)
		}

		s.log.Debug().
			Int("iteration", iteration).
			Str("cstat", resp.CStat).
			Str("xmotivo", resp.XMotivo).
			Msg("sefaz response")

		if resp.CStat == "656" {
			s.rateLimiter.MarkThrottled(empresa.CNPJ, resp.RetryAfter)
			return fmt.Errorf("sefaz throttle: %s", resp.XMotivo)
		}

		if resp.CStat == "137" {
			s.log.Info().Str("cnpj", empresa.CNPJ).Msg("no documents available")
			break
		}

		parsed, err := sefaz.ParseDistDFeResponse(resp.RawXML)
		if err != nil {
			return fmt.Errorf("parsing sefaz response: %w", err)
		}

		var docs []model.DocumentoFiscal
		for _, d := range parsed.Documents {
			if d.DataEmissao != nil && d.DataEmissao.Before(cutoff) {
				continue
			}

			docs = append(docs, model.DocumentoFiscal{
				EmpresaID:   empresa.ID,
				NSU:         d.NSU,
				Tipo:        d.DocumentType,
				Schema:      d.Schema,
				XML:         d.XML,
				ChaveAcesso: d.ChaveAcesso,
				DataEmissao: d.DataEmissao,
			})
		}

		if len(docs) > 0 {
			if err := s.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "nsu"}, {Name: "empresa_id"}},
				DoNothing: true,
			}).Create(&docs).Error; err != nil {
				s.log.Error().Err(err).Uint("empresa_id", empresa.ID).Msg("persist documentos failed")
			}
			totalDocs += len(docs)
		}

		if parsed.UltNSU != "" {
			currentNSU = parsed.UltNSU
		}

		if parsed.UltNSU == parsed.MaxNSU || parsed.UltNSU == "" {
			break
		}

		if iteration < maxSyncLoops {
			time.Sleep(loopSleepSeconds * time.Second)
		}
	}

	now := time.Now()
	if err := s.db.Model(&model.Empresa{}).Where("id = ?", empresa.ID).
		Updates(map[string]any{
			"ult_nsu":              currentNSU,
			"ultima_sincronizacao": now,
		}).Error; err != nil {
		return fmt.Errorf("updating empresa: %w", err)
	}

	s.log.Info().
		Uint("empresa_id", empresa.ID).
		Str("cnpj", empresa.CNPJ).
		Int("documentos", totalDocs).
		Str("ult_nsu", currentNSU).
		Msg("sync completed")

	return nil
}
