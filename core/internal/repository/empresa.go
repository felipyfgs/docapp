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

type SyncStatePatch struct {
	Ativo               *bool
	LookbackDays        *int
	UltNSU              *string
	MaxNSU              *string
	UltimaSincronizacao *time.Time
	BlockedUntil            *time.Time
	SetBlockedUntil         bool
	DownloadBlockedUntil    *time.Time
	SetDownloadBlockedUntil bool
	UltimoCStat             *string
	UltimoXMotivo           *string
}

type EmpresaRepository struct {
	db *bun.DB
}

func NewEmpresaRepository(db *bun.DB) *EmpresaRepository {
	return &EmpresaRepository{db: db}
}

func (r *EmpresaRepository) List(ctx context.Context) ([]model.Empresa, error) {
	empresas := make([]model.Empresa, 0)
	query := r.baseSelect(ctx, &empresas)

	if err := query.OrderExpr("e.created_at DESC").Scan(ctx); err != nil {
		return nil, err
	}

	hydrateEmpresas(empresas)
	return empresas, nil
}

func (r *EmpresaRepository) Create(ctx context.Context, e *model.Empresa) error {
	lookbackDays := e.LookbackDays
	if lookbackDays <= 0 {
		lookbackDays = 90
	}

	ultNSU := strings.TrimSpace(e.UltNSU)
	if ultNSU == "" {
		ultNSU = "000000000000000"
	}

	ativo := e.Ativo
	if !ativo {
		ativo = true
	}

	now := time.Now()

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		e.CreatedAt = now
		e.UpdatedAt = now

		if _, err := tx.NewInsert().Model(e).Exec(ctx); err != nil {
			return err
		}

		syncState := &model.EmpresaSyncState{
			EmpresaID:    e.ID,
			CreatedAt:    now,
			UpdatedAt:    now,
			Ativo:        ativo,
			LookbackDays: lookbackDays,
			UltNSU:       ultNSU,
		}

		if _, err := tx.NewInsert().Model(syncState).Exec(ctx); err != nil {
			return err
		}

		e.SyncState = syncState
		e.HydrateFromRelations()
		return nil
	})

	return err
}

func (r *EmpresaRepository) GetByID(ctx context.Context, id uint) (*model.Empresa, error) {
	var empresa model.Empresa

	err := r.baseSelect(ctx, &empresa).
		Where("e.id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	empresa.HydrateFromRelations()
	return &empresa, nil
}

func (r *EmpresaRepository) FindByCNPJ(ctx context.Context, cnpj string) (*model.Empresa, error) {
	var empresa model.Empresa
	err := r.baseSelect(ctx, &empresa).
		Where("e.cnpj = ?", cnpj).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	empresa.HydrateFromRelations()
	return &empresa, nil
}

func (r *EmpresaRepository) Update(ctx context.Context, id uint, updates *model.Empresa) (*model.Empresa, error) {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		current, err := r.getByIDTx(ctx, tx, id)
		if err != nil {
			return err
		}

		current.CNPJ = updates.CNPJ
		current.RazaoSocial = updates.RazaoSocial
		current.NomeFantasia = updates.NomeFantasia
		current.SituacaoCadastral = updates.SituacaoCadastral
		current.Logradouro = updates.Logradouro
		current.Numero = updates.Numero
		current.Complemento = updates.Complemento
		current.Bairro = updates.Bairro
		current.CEP = updates.CEP
		current.Cidade = updates.Cidade
		current.Estado = updates.Estado
		current.Telefone = updates.Telefone
		current.Email = updates.Email
		current.CNAE = updates.CNAE
		current.Porte = updates.Porte
		current.NaturezaJuridica = updates.NaturezaJuridica
		current.DataInicioAtividade = updates.DataInicioAtividade
		current.UpdatedAt = time.Now()

		if _, err := tx.NewUpdate().Model(current).
			Column(
				"cnpj",
				"razao_social",
				"nome_fantasia",
				"situacao_cadastral",
				"logradouro",
				"numero",
				"complemento",
				"bairro",
				"cep",
				"cidade",
				"estado",
				"telefone",
				"email",
				"cnae",
				"porte",
				"natureza_juridica",
				"data_inicio_atividade",
				"updated_at",
			).
			Where("id = ?", id).
			Exec(ctx); err != nil {
			return err
		}

		if updates.LookbackDays > 0 {
			if _, err := tx.NewUpdate().Model((*model.EmpresaSyncState)(nil)).
				Set("lookback_days = ?", updates.LookbackDays).
				Set("updated_at = ?", time.Now()).
				Where("empresa_id = ?", id).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *EmpresaRepository) Delete(ctx context.Context, id uint) error {
	res, err := r.db.NewDelete().Model((*model.Empresa)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *EmpresaRepository) ListAtivas(ctx context.Context) ([]model.Empresa, error) {
	empresas := make([]model.Empresa, 0)
	query := r.baseSelect(ctx, &empresas).
		Join("JOIN empresa_sync_states AS ess_filter ON ess_filter.empresa_id = e.id AND ess_filter.deleted_at IS NULL").
		Where("ess_filter.ativo = TRUE")

	if err := query.OrderExpr("e.created_at DESC").Scan(ctx); err != nil {
		return nil, err
	}

	hydrateEmpresas(empresas)
	return empresas, nil
}

func (r *EmpresaRepository) ListAtivasComCertificado(ctx context.Context) ([]model.Empresa, error) {
	now := time.Now()
	empresas := make([]model.Empresa, 0)
	query := r.baseSelect(ctx, &empresas).
		Join("JOIN empresa_sync_states AS ess_filter ON ess_filter.empresa_id = e.id AND ess_filter.deleted_at IS NULL").
		Join("JOIN empresa_certificados AS ec_filter ON ec_filter.empresa_id = e.id AND ec_filter.deleted_at IS NULL").
		Where("ess_filter.ativo = TRUE").
		Where("octet_length(ec_filter.certificado_pfx) > 0").
		Where("ec_filter.certificado_senha <> ''").
		Where("(ec_filter.sigla_uf <> '' OR e.estado <> '')").
		Where("(ess_filter.blocked_until IS NULL OR ess_filter.blocked_until <= ?)", now)

	if err := query.OrderExpr("e.created_at DESC").Scan(ctx); err != nil {
		return nil, err
	}

	hydrateEmpresas(empresas)
	return empresas, nil
}

func (r *EmpresaRepository) UpdateSyncState(ctx context.Context, empresaID uint, patch SyncStatePatch) error {
	query := r.db.NewUpdate().Model((*model.EmpresaSyncState)(nil)).Where("empresa_id = ?", empresaID)

	if patch.Ativo != nil {
		query = query.Set("ativo = ?", *patch.Ativo)
	}
	if patch.LookbackDays != nil {
		query = query.Set("lookback_days = ?", *patch.LookbackDays)
	}
	if patch.UltNSU != nil {
		query = query.Set("ult_nsu = ?", strings.TrimSpace(*patch.UltNSU))
	}
	if patch.MaxNSU != nil {
		query = query.Set("max_nsu = ?", strings.TrimSpace(*patch.MaxNSU))
	}
	if patch.UltimaSincronizacao != nil {
		query = query.Set("ultima_sincronizacao = ?", *patch.UltimaSincronizacao)
	}
	if patch.SetBlockedUntil {
		if patch.BlockedUntil == nil {
			query = query.Set("blocked_until = NULL")
		} else {
			query = query.Set("blocked_until = ?", *patch.BlockedUntil)
		}
	}
	if patch.SetDownloadBlockedUntil {
		if patch.DownloadBlockedUntil == nil {
			query = query.Set("download_blocked_until = NULL")
		} else {
			query = query.Set("download_blocked_until = ?", *patch.DownloadBlockedUntil)
		}
	}
	if patch.UltimoCStat != nil {
		query = query.Set("ultimo_cstat = ?", strings.TrimSpace(*patch.UltimoCStat))
	}
	if patch.UltimoXMotivo != nil {
		query = query.Set("ultimo_xmotivo = ?", strings.TrimSpace(*patch.UltimoXMotivo))
	}

	query = query.Set("updated_at = ?", time.Now())

	res, err := query.Exec(ctx)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *EmpresaRepository) UpdateCertificadoUF(ctx context.Context, empresaID uint, siglaUF string) error {
	res, err := r.db.NewUpdate().Model((*model.EmpresaCertificado)(nil)).
		Set("sigla_uf = ?", strings.ToUpper(strings.TrimSpace(siglaUF))).
		Set("updated_at = ?", time.Now()).
		Where("empresa_id = ?", empresaID).
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

func (r *EmpresaRepository) UpsertCertificado(ctx context.Context, id uint, pfx []byte, senha, siglaUF string, tpAmb int, validoAte *time.Time) error {
	now := time.Now()
	cert := &model.EmpresaCertificado{
		EmpresaID:            id,
		CreatedAt:            now,
		UpdatedAt:            now,
		CertificadoPFX:       pfx,
		CertificadoSenha:     senha,
		SiglaUF:              strings.ToUpper(strings.TrimSpace(siglaUF)),
		TpAmb:                tpAmb,
		CertificadoValidoAte: validoAte,
	}

	_, err := r.db.NewInsert().Model(cert).
		On("CONFLICT (empresa_id) DO UPDATE").
		Set("updated_at = EXCLUDED.updated_at").
		Set("deleted_at = NULL").
		Set("certificado_pfx = EXCLUDED.certificado_pfx").
		Set("certificado_senha = EXCLUDED.certificado_senha").
		Set("sigla_uf = EXCLUDED.sigla_uf").
		Set("tp_amb = EXCLUDED.tp_amb").
		Set("certificado_valido_ate = EXCLUDED.certificado_valido_ate").
		Exec(ctx)

	return err
}

func (r *EmpresaRepository) ListCertificadosSemValidade(ctx context.Context) ([]model.EmpresaCertificado, error) {
	certs := make([]model.EmpresaCertificado, 0)
	err := r.db.NewSelect().Model(&certs).
		Where("deleted_at IS NULL").
		Where("certificado_valido_ate IS NULL").
		Where("octet_length(certificado_pfx) > 0").
		Where("certificado_senha <> ''").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return certs, nil
}

func (r *EmpresaRepository) UpdateCertificadoValidade(ctx context.Context, empresaID uint, validoAte *time.Time) error {
	res, err := r.db.NewUpdate().Model((*model.EmpresaCertificado)(nil)).
		Set("certificado_valido_ate = ?", validoAte).
		Set("updated_at = ?", time.Now()).
		Where("empresa_id = ?", empresaID).
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

func (r *EmpresaRepository) getByIDTx(ctx context.Context, tx bun.Tx, id uint) (*model.Empresa, error) {
	var empresa model.Empresa
	err := tx.NewSelect().Model(&empresa).
		Relation("SyncState", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"sync_state".deleted_at IS NULL`)
		}).
		Relation("Certificado", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"certificado".deleted_at IS NULL`)
		}).
		Where("e.id = ?", id).
		Where("e.deleted_at IS NULL").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	empresa.HydrateFromRelations()
	return &empresa, nil
}

func (r *EmpresaRepository) baseSelect(_ context.Context, target any) *bun.SelectQuery {
	return r.db.NewSelect().Model(target).
		Relation("SyncState", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"sync_state".deleted_at IS NULL`)
		}).
		Relation("Certificado", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"certificado".deleted_at IS NULL`)
		}).
		Where("e.deleted_at IS NULL")
}

func hydrateEmpresas(empresas []model.Empresa) {
	for i := range empresas {
		empresas[i].HydrateFromRelations()
	}
}
