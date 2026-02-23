ALTER TABLE documentos_fiscais
    ADD COLUMN IF NOT EXISTS manifestacao_status VARCHAR(20) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS manifestacao_at TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_documentos_fiscais_manifestacao_pending
    ON documentos_fiscais (empresa_id, data_emissao DESC)
    WHERE deleted_at IS NULL AND xml_resumo = TRUE AND manifestacao_status IS NULL;
