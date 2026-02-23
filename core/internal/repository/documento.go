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
		On("CONFLICT (empresa_id, nsu) DO NOTHING").
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
