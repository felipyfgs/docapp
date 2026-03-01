export default defineEventHandler(async (event) => {
  if (event.method !== 'POST') {
    throw createError({ statusCode: 405, message: 'Method Not Allowed' })
  }

  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')

  return $fetch(`${config.coreApiUrl}/empresas/${id}/sync`, { method: 'POST' })
})
