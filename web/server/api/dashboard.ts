export default defineEventHandler(async (event) => {
  if (event.method !== 'GET') {
    throw createError({ statusCode: 405, message: 'Method Not Allowed' })
  }

  const config = useRuntimeConfig()
  const query = getQuery(event)

  return $fetch(`${config.coreApiUrl}/documentos/dashboard`, {
    method: 'GET',
    query
  })
})
