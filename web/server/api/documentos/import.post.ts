export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()

  const body = await readRawBody(event, false)
  const contentType = getHeader(event, 'content-type') ?? ''

  return $fetch(`${config.coreApiUrl}/documentos/import`, {
    method: 'POST',
    body,
    headers: { 'content-type': contentType },
  })
})
