CREATE TABLE documento_itens (
    id            BIGSERIAL PRIMARY KEY,
    documento_id  BIGINT NOT NULL REFERENCES documentos_fiscais(id) ON DELETE CASCADE,
    n_item        SMALLINT NOT NULL,

    -- produto
    c_prod        VARCHAR(60),
    c_ean         VARCHAR(14),
    x_prod        TEXT NOT NULL,
    ncm           VARCHAR(8),
    cest          VARCHAR(7),
    cfop          VARCHAR(4),
    u_com         VARCHAR(6),
    q_com         NUMERIC(15,4)  NOT NULL DEFAULT 0,
    v_un_com      NUMERIC(21,10) NOT NULL DEFAULT 0,
    v_prod        NUMERIC(15,2)  NOT NULL DEFAULT 0,
    v_desc        NUMERIC(15,2)  NOT NULL DEFAULT 0,
    v_frete       NUMERIC(15,2)  NOT NULL DEFAULT 0,
    v_seg         NUMERIC(15,2)  NOT NULL DEFAULT 0,
    v_outro       NUMERIC(15,2)  NOT NULL DEFAULT 0,
    x_ped         VARCHAR(60),
    n_item_ped    VARCHAR(6),
    inf_ad_prod   TEXT,

    -- icms
    icms_orig       VARCHAR(1),
    icms_cst        VARCHAR(3),
    icms_mod_bc     VARCHAR(1),
    icms_p_red_bc   NUMERIC(7,4),
    icms_v_bc       NUMERIC(15,2),
    icms_p_icms     NUMERIC(7,4),
    icms_v_icms     NUMERIC(15,2)  NOT NULL DEFAULT 0,
    icms_v_bc_st    NUMERIC(15,2),
    icms_p_st       NUMERIC(7,4),
    icms_v_icms_st  NUMERIC(15,2),
    icms_v_icms_deson NUMERIC(15,2),

    -- ipi
    ipi_cst       VARCHAR(2),
    ipi_v_bc      NUMERIC(15,2),
    ipi_p_ipi     NUMERIC(7,4),
    ipi_v_ipi     NUMERIC(15,2) NOT NULL DEFAULT 0,

    -- pis
    pis_cst       VARCHAR(2),
    pis_v_bc      NUMERIC(15,2),
    pis_p_pis     NUMERIC(7,4),
    pis_v_pis     NUMERIC(15,2) NOT NULL DEFAULT 0,

    -- cofins
    cofins_cst      VARCHAR(2),
    cofins_v_bc     NUMERIC(15,2),
    cofins_p_cofins NUMERIC(7,4),
    cofins_v_cofins NUMERIC(15,2) NOT NULL DEFAULT 0,

    -- ibs/cbs (reforma tributaria)
    ibscbs_cst          VARCHAR(3),
    ibscbs_c_class_trib VARCHAR(6),
    ibscbs_v_bc         NUMERIC(15,2),
    ibscbs_p_ibs_uf     NUMERIC(7,4),
    ibscbs_v_ibs_uf     NUMERIC(15,2),
    ibscbs_p_ibs_mun    NUMERIC(7,4),
    ibscbs_v_ibs_mun    NUMERIC(15,2),
    ibscbs_p_cbs        NUMERIC(7,4),
    ibscbs_v_cbs        NUMERIC(15,2),

    -- total tributos item
    v_tot_trib    NUMERIC(15,2),

    CONSTRAINT uq_documento_itens UNIQUE (documento_id, n_item)
);

CREATE INDEX idx_documento_itens_doc ON documento_itens (documento_id);
CREATE INDEX idx_documento_itens_ncm ON documento_itens (ncm) WHERE ncm IS NOT NULL;
CREATE INDEX idx_documento_itens_cfop ON documento_itens (cfop) WHERE cfop IS NOT NULL;
CREATE INDEX idx_documento_itens_cprod ON documento_itens (c_prod) WHERE c_prod IS NOT NULL;
CREATE INDEX idx_documento_itens_xprod ON documento_itens USING GIN (x_prod gin_trgm_ops);
