package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/model"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SyncService struct {
	db     *gorm.DB
	client *client.Client
	log    zerolog.Logger
}

func NewSyncService(db *gorm.DB, c *client.Client, log zerolog.Logger) *SyncService {
	return &SyncService{db: db, client: c, log: log}
}

func (s *SyncService) SyncEmpresa(empresa model.Empresa) error {
	payload, err := json.Marshal(map[string]any{
		"tenant":      strings.ReplaceAll(empresa.CNPJ, "/", ""),
		"ult_nsu":     empresa.UltNSU,
		"max_loops":   12,
		"include_xml": true,
	})
	if err != nil {
		return fmt.Errorf("building payload: %w", err)
	}

	_, body, err := s.client.DistDFe(context.Background(), payload)
	if err != nil {
		return fmt.Errorf("calling distdfe: %w", err)
	}

	var result struct {
		UltNSU    string           `json:"ult_nsu"`
		Documents []map[string]any `json:"documents"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -empresa.LookbackDays)

	var docs []model.DocumentoFiscal
	for _, d := range result.Documents {
		doc := model.DocumentoFiscal{
			EmpresaID:   empresa.ID,
			NSU:         stringVal(d, "nsu"),
			Tipo:        stringVal(d, "document_type"),
			Schema:      stringVal(d, "schema"),
			XML:         stringVal(d, "xml"),
			ChaveAcesso: extractChaveAcesso(stringVal(d, "xml")),
		}

		if emissao := extractDataEmissao(stringVal(d, "xml")); emissao != nil {
			if emissao.Before(cutoff) {
				continue
			}

			doc.DataEmissao = emissao
		}

		docs = append(docs, doc)
	}

	if len(docs) > 0 {
		if err := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "nsu"}, {Name: "empresa_id"}},
			DoNothing: true,
		}).Create(&docs).Error; err != nil {
			s.log.Error().Err(err).Uint("empresa_id", empresa.ID).Msg("persist documentos failed")
		}
	}

	if result.UltNSU != "" && result.UltNSU != empresa.UltNSU {
		if err := s.db.Model(&model.Empresa{}).Where("id = ?", empresa.ID).
			Update("ult_nsu", result.UltNSU).Error; err != nil {
			return fmt.Errorf("updating ult_nsu: %w", err)
		}
	}

	s.log.Info().
		Uint("empresa_id", empresa.ID).
		Str("cnpj", empresa.CNPJ).
		Int("documentos", len(docs)).
		Str("ult_nsu", result.UltNSU).
		Msg("sync completed")

	return nil
}

func stringVal(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func extractChaveAcesso(xml string) string {
	const tag = "<chNFe>"
	start := strings.Index(xml, tag)
	if start == -1 {
		return ""
	}

	start += len(tag)
	end := strings.Index(xml[start:], "</chNFe>")

	if end == -1 || end != 44 {
		return ""
	}

	return xml[start : start+end]
}

func extractDataEmissao(xml string) *time.Time {
	for _, tag := range []string{"<dhEmi>", "<dEmi>"} {
		start := strings.Index(xml, tag)
		if start == -1 {
			continue
		}

		start += len(tag)
		end := strings.Index(xml[start:], tag[:1]+"/"+tag[1:])
		if end == -1 {
			continue
		}

		raw := strings.TrimSpace(xml[start : start+end])

		for _, layout := range []string{time.RFC3339, "2006-01-02"} {
			if t, err := time.Parse(layout, raw); err == nil {
				return &t
			}
		}
	}

	return nil
}
