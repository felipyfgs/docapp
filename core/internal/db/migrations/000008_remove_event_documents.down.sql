UPDATE documentos_fiscais
SET deleted_at = NULL
WHERE schema_nome IN ('procEventoNFe_v1.00.xsd', 'resEvento_v1.01.xsd')
  AND deleted_at IS NOT NULL;
