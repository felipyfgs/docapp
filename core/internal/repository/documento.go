package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"docapp/core/internal/model"

	"github.com/uptrace/bun"
)

type DocumentoListFilter struct {
	Search    string
	Tipo      string
	Status    string
	EmpresaID uint
	XMLResumo *bool
	Page      int
	PageSize  int
}

type DocumentoRepository struct {
	db *bun.DB
}

func NewDocumentoRepository(db *bun.DB) *DocumentoRepository {
	return &DocumentoRepository{db: db}
}

func (r *DocumentoRepository) List(ctx context.Context, filter DocumentoListFilter) ([]model.DocumentoFiscal, int64, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	countQuery := r.db.NewSelect().Model((*model.DocumentoFiscal)(nil)).Where("df.deleted_at IS NULL")
	countQuery = applyDocumentoFilters(countQuery, filter)

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	docs := make([]model.DocumentoFiscal, 0)
	query := r.db.NewSelect().Model(&docs).
		Relation("Empresa", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"empresa".deleted_at IS NULL`)
		}).
		Where("df.deleted_at IS NULL")
	query = applyDocumentoFilters(query, filter)

	err = query.
		OrderExpr("df.data_emissao DESC NULLS LAST, df.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return docs, int64(total), nil
}

func (r *DocumentoRepository) GetByID(ctx context.Context, id uint) (*model.DocumentoFiscal, error) {
	var doc model.DocumentoFiscal
	err := r.db.NewSelect().Model(&doc).
		Relation("Empresa", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"empresa".deleted_at IS NULL`)
		}).
		Where("df.id = ?", id).
		Where("df.deleted_at IS NULL").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &doc, nil
}

func (r *DocumentoRepository) ListByIDs(ctx context.Context, ids []uint) ([]model.DocumentoFiscal, error) {
	if len(ids) == 0 {
		return []model.DocumentoFiscal{}, nil
	}

	docs := make([]model.DocumentoFiscal, 0, len(ids))
	err := r.db.NewSelect().Model(&docs).
		Relation("Empresa", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"empresa".deleted_at IS NULL`)
		}).
		Where("df.id IN (?)", bun.In(ids)).
		Where("df.deleted_at IS NULL").
		OrderExpr("df.created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *DocumentoRepository) UpsertMany(ctx context.Context, docs []model.DocumentoFiscal) error {
	if len(docs) == 0 {
		return nil
	}

	now := time.Now()
	for i := range docs {
		docs[i].CreatedAt = now
		docs[i].UpdatedAt = now
	}

	_, err := r.db.NewInsert().Model(&docs).
		On("CONFLICT (empresa_id, nsu) DO UPDATE").
		Set("chave_acesso = EXCLUDED.chave_acesso").
		Set("tipo_documento = EXCLUDED.tipo_documento").
		Set("status_documento = EXCLUDED.status_documento").
		Set("numero_documento = EXCLUDED.numero_documento").
		Set("emitente_nome = EXCLUDED.emitente_nome").
		Set("emitente_cnpj = EXCLUDED.emitente_cnpj").
		Set("destinatario_nome = EXCLUDED.destinatario_nome").
		Set("destinatario_cnpj = EXCLUDED.destinatario_cnpj").
		Set("competencia = EXCLUDED.competencia").
		Set("schema_nome = EXCLUDED.schema_nome").
		Set("xml_object_key = EXCLUDED.xml_object_key").
		Set("xml_sha256 = EXCLUDED.xml_sha256").
		Set("xml_size_bytes = EXCLUDED.xml_size_bytes").
		Set("xml_resumo = EXCLUDED.xml_resumo").
		Set("data_emissao = EXCLUDED.data_emissao").
		Set("search_text = EXCLUDED.search_text").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *DocumentoRepository) UpdateDanfeMetadata(ctx context.Context, id uint, danfeObjectKey string, generatedAt time.Time) error {
	res, err := r.db.NewUpdate().Model((*model.DocumentoFiscal)(nil)).
		Set("danfe_object_key = ?", danfeObjectKey).
		Set("danfe_generated_at = ?", generatedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *DocumentoRepository) UpgradeFromResumo(ctx context.Context, id uint, doc model.DocumentoFiscal) error {
	_, err := r.db.NewUpdate().Model((*model.DocumentoFiscal)(nil)).
		Set("xml_resumo = ?", false).
		Set("xml_object_key = ?", doc.XMLObjectKey).
		Set("xml_sha256 = ?", doc.XMLSHA256).
		Set("xml_size_bytes = ?", doc.XMLSizeBytes).
		Set("emitente_nome = ?", doc.EmitenteNome).
		Set("emitente_cnpj = ?", doc.EmitenteCNPJ).
		Set("destinatario_nome = ?", doc.DestinatarioNome).
		Set("destinatario_cnpj = ?", doc.DestinatarioCNPJ).
		Set("numero_documento = ?", doc.NumeroDocumento).
		Set("status_documento = ?", doc.StatusDocumento).
		Set("schema_nome = ?", doc.Schema).
		Set("search_text = ?", doc.SearchText).
		Set("manifestacao_status = ?", doc.ManifestacaoStatus).
		Set("manifestacao_at = ?", doc.ManifestacaoAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r *DocumentoRepository) UpdateManifestacaoStatus(ctx context.Context, id uint, status string, at time.Time) error {
	_, err := r.db.NewUpdate().Model((*model.DocumentoFiscal)(nil)).
		Set("manifestacao_status = ?", status).
		Set("manifestacao_at = ?", at).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r *DocumentoRepository) ListPendingManifestacao(ctx context.Context, empresaID uint) ([]model.DocumentoFiscal, error) {
	docs := make([]model.DocumentoFiscal, 0)
	err := r.db.NewSelect().Model(&docs).
		Where("df.empresa_id = ?", empresaID).
		Where("df.deleted_at IS NULL").
		Where("df.xml_resumo = TRUE").
		Where("df.manifestacao_status IS NULL").
		Where("df.chave_acesso IS NOT NULL").
		Where("df.chave_acesso != ''").
		OrderExpr("df.data_emissao DESC NULLS LAST").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func applyDocumentoFilters(query *bun.SelectQuery, filter DocumentoListFilter) *bun.SelectQuery {
	if filter.EmpresaID > 0 {
		query = query.Where("df.empresa_id = ?", filter.EmpresaID)
	}

	if tipo := strings.TrimSpace(filter.Tipo); tipo != "" {
		query = query.Where("df.tipo_documento = ?", tipo)
	}

	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("df.status_documento = ?", status)
	}

	if filter.XMLResumo != nil {
		query = query.Where("df.xml_resumo = ?", *filter.XMLResumo)
	}

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(df.search_text) LIKE ?", like)
	}

	return query
}
