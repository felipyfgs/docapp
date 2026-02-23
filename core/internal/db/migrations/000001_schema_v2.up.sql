CREATE EXTENSION IF NOT EXISTS pg_trgm;

DROP TABLE IF EXISTS documentos_fiscais CASCADE;
DROP TABLE IF EXISTS documento_fiscals CASCADE;
DROP TABLE IF EXISTS empresa_sync_states CASCADE;
DROP TABLE IF EXISTS empresa_certificados CASCADE;
DROP TABLE IF EXISTS empresas CASCADE;

CREATE TABLE empresas (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    cnpj VARCHAR(14) NOT NULL,
    razao_social TEXT NOT NULL,
    nome_fantasia TEXT,
    situacao_cadastral TEXT,
    logradouro TEXT,
    numero TEXT,
    complemento TEXT,
    bairro TEXT,
    cep TEXT,
    cidade TEXT,
    estado TEXT,
    telefone TEXT,
    email TEXT,
    cnae TEXT,
    porte TEXT,
    natureza_juridica TEXT,
    data_inicio_atividade TEXT,
    CONSTRAINT uq_empresas_cnpj UNIQUE (cnpj),
    CONSTRAINT ck_empresas_cnpj CHECK (cnpj ~ '^[0-9]{14}$')
);

CREATE INDEX idx_empresas_deleted_at ON empresas (deleted_at);

CREATE TABLE empresa_certificados (
    empresa_id BIGINT PRIMARY KEY REFERENCES empresas(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    certificado_pfx BYTEA NOT NULL,
    certificado_senha TEXT NOT NULL,
    sigla_uf VARCHAR(2) NOT NULL,
    tp_amb SMALLINT NOT NULL DEFAULT 1,
    certificado_valido_ate TIMESTAMPTZ,
    CONSTRAINT ck_empresa_certificados_tp_amb CHECK (tp_amb IN (1, 2)),
    CONSTRAINT ck_empresa_certificados_sigla_uf CHECK (sigla_uf ~ '^[A-Z]{2}$'),
    CONSTRAINT ck_empresa_certificados_pfx CHECK (octet_length(certificado_pfx) > 0)
);

CREATE INDEX idx_empresa_certificados_deleted_at ON empresa_certificados (deleted_at);
CREATE INDEX idx_empresa_certificados_sigla_uf ON empresa_certificados (sigla_uf) WHERE deleted_at IS NULL;

CREATE TABLE empresa_sync_states (
    empresa_id BIGINT PRIMARY KEY REFERENCES empresas(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    lookback_days SMALLINT NOT NULL DEFAULT 90,
    ult_nsu VARCHAR(15) NOT NULL DEFAULT '000000000000000',
    max_nsu VARCHAR(15),
    ultima_sincronizacao TIMESTAMPTZ,
    blocked_until TIMESTAMPTZ,
    ultimo_cstat VARCHAR(3),
    ultimo_xmotivo TEXT,
    CONSTRAINT ck_empresa_sync_states_lookback CHECK (lookback_days >= 1 AND lookback_days <= 3650),
    CONSTRAINT ck_empresa_sync_states_ult_nsu CHECK (ult_nsu ~ '^[0-9]{15}$'),
    CONSTRAINT ck_empresa_sync_states_max_nsu CHECK (max_nsu IS NULL OR max_nsu ~ '^[0-9]{15}$')
);

CREATE INDEX idx_empresa_sync_states_deleted_at ON empresa_sync_states (deleted_at);
CREATE INDEX idx_empresa_sync_states_worker_pick
    ON empresa_sync_states (ativo, blocked_until, ultima_sincronizacao)
    WHERE deleted_at IS NULL;

CREATE TABLE documentos_fiscais (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    empresa_id BIGINT NOT NULL REFERENCES empresas(id) ON DELETE CASCADE,
    nsu VARCHAR(15) NOT NULL,
    chave_acesso VARCHAR(44),
    tipo_documento VARCHAR(12) NOT NULL,
    status_documento VARCHAR(16) NOT NULL DEFAULT 'desconhecido',
    numero_documento VARCHAR(32),
    emitente_nome TEXT,
    emitente_cnpj VARCHAR(14),
    destinatario_nome TEXT,
    destinatario_cnpj VARCHAR(14),
    competencia VARCHAR(7),
    schema_nome VARCHAR(64),
    data_emissao TIMESTAMPTZ,
    xml_object_key TEXT NOT NULL,
    xml_sha256 VARCHAR(64),
    xml_size_bytes INTEGER,
    xml_resumo BOOLEAN NOT NULL DEFAULT FALSE,
    danfe_object_key TEXT,
    danfe_generated_at TIMESTAMPTZ,
    search_text TEXT NOT NULL DEFAULT '',
    CONSTRAINT uq_documentos_fiscais_empresa_nsu UNIQUE (empresa_id, nsu),
    CONSTRAINT uq_documentos_fiscais_empresa_chave UNIQUE (empresa_id, chave_acesso),
    CONSTRAINT ck_documentos_fiscais_nsu CHECK (nsu ~ '^[0-9]{15}$'),
    CONSTRAINT ck_documentos_fiscais_chave CHECK (chave_acesso IS NULL OR chave_acesso ~ '^[0-9]{44}$'),
    CONSTRAINT ck_documentos_fiscais_emitente_cnpj CHECK (emitente_cnpj IS NULL OR emitente_cnpj ~ '^[0-9]{14}$'),
    CONSTRAINT ck_documentos_fiscais_destinatario_cnpj CHECK (destinatario_cnpj IS NULL OR destinatario_cnpj ~ '^[0-9]{14}$')
);

CREATE INDEX idx_documentos_fiscais_deleted_at ON documentos_fiscais (deleted_at);
CREATE INDEX idx_documentos_fiscais_list
    ON documentos_fiscais (empresa_id, data_emissao DESC, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_documentos_fiscais_filters
    ON documentos_fiscais (empresa_id, tipo_documento, status_documento, data_emissao DESC, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_documentos_fiscais_chave_acesso ON documentos_fiscais (chave_acesso) WHERE deleted_at IS NULL;
CREATE INDEX idx_documentos_fiscais_emitente_cnpj ON documentos_fiscais (emitente_cnpj) WHERE deleted_at IS NULL;
CREATE INDEX idx_documentos_fiscais_destinatario_cnpj ON documentos_fiscais (destinatario_cnpj) WHERE deleted_at IS NULL;
CREATE INDEX idx_documentos_fiscais_competencia
    ON documentos_fiscais (empresa_id, competencia, data_emissao DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_documentos_fiscais_xml_resumo
    ON documentos_fiscais (empresa_id, xml_resumo, data_emissao DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_documentos_fiscais_danfe_pending
    ON documentos_fiscais (empresa_id, data_emissao DESC)
    WHERE deleted_at IS NULL AND xml_resumo = FALSE AND danfe_object_key IS NULL;
CREATE INDEX idx_documentos_fiscais_search_trgm
    ON documentos_fiscais USING GIN (search_text gin_trgm_ops)
    WHERE deleted_at IS NULL;
