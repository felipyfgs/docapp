ALTER TABLE documentos_fiscais ALTER COLUMN chave_acesso TYPE VARCHAR(44);
ALTER TABLE documentos_fiscais ALTER COLUMN nsu TYPE VARCHAR(15);

ALTER TABLE documentos_fiscais DROP CONSTRAINT IF EXISTS ck_documentos_fiscais_chave;
ALTER TABLE documentos_fiscais ADD CONSTRAINT ck_documentos_fiscais_chave CHECK (chave_acesso IS NULL OR chave_acesso ~ '^[0-9]{44}$');

ALTER TABLE documentos_fiscais DROP CONSTRAINT IF EXISTS ck_documentos_fiscais_nsu;
ALTER TABLE documentos_fiscais ADD CONSTRAINT ck_documentos_fiscais_nsu CHECK (nsu ~ '^[0-9]{15}$');
