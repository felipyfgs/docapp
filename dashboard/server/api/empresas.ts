export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const base = config.coreApiUrl

  if (event.method === 'GET') {
    return $fetch(`${base}/empresas/`, { method: 'GET' })
  }

  if (event.method === 'POST') {
    const body = await readBody(event)
    return $fetch(`${base}/empresas/`, { method: 'POST', body })
  }

  throw createError({ statusCode: 405, message: 'Method Not Allowed' })
})
