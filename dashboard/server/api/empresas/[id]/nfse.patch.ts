export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')
  const body = await readBody(event)

  return $fetch(`${config.coreApiUrl}/empresas/${id}/nfse`, {
    method: 'PATCH',
    body
  })
})
