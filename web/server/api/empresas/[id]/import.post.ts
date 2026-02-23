export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')

  const body = await readRawBody(event, false)
  const contentType = getHeader(event, 'content-type') ?? ''

  return $fetch(`${config.coreApiUrl}/empresas/${id}/import`, {
    method: 'POST',
    body,
    headers: { 'content-type': contentType },
  })
})
