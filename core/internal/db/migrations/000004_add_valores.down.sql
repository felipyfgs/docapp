ALTER TABLE documentos_fiscais
  DROP COLUMN IF EXISTS valor_total,
  DROP COLUMN IF EXISTS valor_produtos;
