ALTER TABLE empresa_sync_states
    DROP COLUMN IF EXISTS nfse_habilitada,
    DROP COLUMN IF EXISTS ult_nsu_nfse,
    DROP COLUMN IF EXISTS ultima_sync_nfse,
    DROP COLUMN IF EXISTS nfse_blocked_until;
