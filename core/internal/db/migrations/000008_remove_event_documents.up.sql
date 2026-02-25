UPDATE documentos_fiscais
SET deleted_at = NOW()
WHERE schema_nome IN ('procEventoNFe_v1.00.xsd', 'resEvento_v1.01.xsd')
  AND deleted_at IS NULL;
