export default defineEventHandler(async (event) => {
  if (event.method !== 'GET') {
    throw createError({ statusCode: 405, message: 'Method Not Allowed' })
  }

  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')

  return $fetch(`${config.coreApiUrl}/documentos/${id}/xml`, {
    method: 'GET'
  })
})
