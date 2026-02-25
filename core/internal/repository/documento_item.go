package repository

import (
	"context"

	"docapp/core/internal/model"

	"github.com/uptrace/bun"
)

type DocumentoItemRepository struct {
	db *bun.DB
}

func NewDocumentoItemRepository(db *bun.DB) *DocumentoItemRepository {
	return &DocumentoItemRepository{db: db}
}

func (r *DocumentoItemRepository) UpsertItens(ctx context.Context, documentoID uint, itens []model.DocumentoItem) error {
	if len(itens) == 0 {
		return nil
	}

	for i := range itens {
		itens[i].DocumentoID = documentoID
	}

	_, err := r.db.NewInsert().Model(&itens).
		On("CONFLICT (documento_id, n_item) DO UPDATE").
		Set("c_prod = EXCLUDED.c_prod").
		Set("c_ean = EXCLUDED.c_ean").
		Set("x_prod = EXCLUDED.x_prod").
		Set("ncm = EXCLUDED.ncm").
		Set("cest = EXCLUDED.cest").
		Set("cfop = EXCLUDED.cfop").
		Set("u_com = EXCLUDED.u_com").
		Set("q_com = EXCLUDED.q_com").
		Set("v_un_com = EXCLUDED.v_un_com").
		Set("v_prod = EXCLUDED.v_prod").
		Set("v_desc = EXCLUDED.v_desc").
		Set("v_frete = EXCLUDED.v_frete").
		Set("v_seg = EXCLUDED.v_seg").
		Set("v_outro = EXCLUDED.v_outro").
		Set("x_ped = EXCLUDED.x_ped").
		Set("n_item_ped = EXCLUDED.n_item_ped").
		Set("inf_ad_prod = EXCLUDED.inf_ad_prod").
		Set("icms_orig = EXCLUDED.icms_orig").
		Set("icms_cst = EXCLUDED.icms_cst").
		Set("icms_mod_bc = EXCLUDED.icms_mod_bc").
		Set("icms_p_red_bc = EXCLUDED.icms_p_red_bc").
		Set("icms_v_bc = EXCLUDED.icms_v_bc").
		Set("icms_p_icms = EXCLUDED.icms_p_icms").
		Set("icms_v_icms = EXCLUDED.icms_v_icms").
		Set("icms_v_bc_st = EXCLUDED.icms_v_bc_st").
		Set("icms_p_st = EXCLUDED.icms_p_st").
		Set("icms_v_icms_st = EXCLUDED.icms_v_icms_st").
		Set("icms_v_icms_deson = EXCLUDED.icms_v_icms_deson").
		Set("ipi_cst = EXCLUDED.ipi_cst").
		Set("ipi_v_bc = EXCLUDED.ipi_v_bc").
		Set("ipi_p_ipi = EXCLUDED.ipi_p_ipi").
		Set("ipi_v_ipi = EXCLUDED.ipi_v_ipi").
		Set("pis_cst = EXCLUDED.pis_cst").
		Set("pis_v_bc = EXCLUDED.pis_v_bc").
		Set("pis_p_pis = EXCLUDED.pis_p_pis").
		Set("pis_v_pis = EXCLUDED.pis_v_pis").
		Set("cofins_cst = EXCLUDED.cofins_cst").
		Set("cofins_v_bc = EXCLUDED.cofins_v_bc").
		Set("cofins_p_cofins = EXCLUDED.cofins_p_cofins").
		Set("cofins_v_cofins = EXCLUDED.cofins_v_cofins").
		Set("ibscbs_cst = EXCLUDED.ibscbs_cst").
		Set("ibscbs_c_class_trib = EXCLUDED.ibscbs_c_class_trib").
		Set("ibscbs_v_bc = EXCLUDED.ibscbs_v_bc").
		Set("ibscbs_p_ibs_uf = EXCLUDED.ibscbs_p_ibs_uf").
		Set("ibscbs_v_ibs_uf = EXCLUDED.ibscbs_v_ibs_uf").
		Set("ibscbs_p_ibs_mun = EXCLUDED.ibscbs_p_ibs_mun").
		Set("ibscbs_v_ibs_mun = EXCLUDED.ibscbs_v_ibs_mun").
		Set("ibscbs_p_cbs = EXCLUDED.ibscbs_p_cbs").
		Set("ibscbs_v_cbs = EXCLUDED.ibscbs_v_cbs").
		Set("v_tot_trib = EXCLUDED.v_tot_trib").
		Exec(ctx)
	return err
}

func (r *DocumentoItemRepository) ListByDocumentoID(ctx context.Context, documentoID uint) ([]model.DocumentoItem, error) {
	itens := make([]model.DocumentoItem, 0)
	err := r.db.NewSelect().Model(&itens).
		Where("di.documento_id = ?", documentoID).
		OrderExpr("di.n_item ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return itens, nil
}
