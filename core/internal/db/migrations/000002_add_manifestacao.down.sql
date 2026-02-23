DROP INDEX IF EXISTS idx_documentos_fiscais_manifestacao_pending;
ALTER TABLE documentos_fiscais
    DROP COLUMN IF EXISTS manifestacao_status,
    DROP COLUMN IF EXISTS manifestacao_at;
