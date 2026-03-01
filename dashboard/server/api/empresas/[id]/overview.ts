export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')

  return $fetch(`${config.coreApiUrl}/empresas/${id}/overview`, { method: 'GET' })
})
