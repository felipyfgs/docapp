export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const base = config.coreApiUrl
  const cnpj = getRouterParam(event, 'cnpj')

  return $fetch(`${base}/cnpj/${cnpj}`, { method: 'GET' })
})
