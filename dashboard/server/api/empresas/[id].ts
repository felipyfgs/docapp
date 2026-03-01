export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const base = config.coreApiUrl
  const id = getRouterParam(event, 'id')

  if (event.method === 'GET') {
    return $fetch(`${base}/empresas/${id}`, { method: 'GET' })
  }

  if (event.method === 'PUT') {
    const body = await readBody(event)
    return $fetch(`${base}/empresas/${id}`, { method: 'PUT', body })
  }

  if (event.method === 'DELETE') {
    await $fetch(`${base}/empresas/${id}`, { method: 'DELETE' })
    return null
  }

  throw createError({ statusCode: 405, message: 'Method Not Allowed' })
})
