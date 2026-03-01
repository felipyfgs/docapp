export default defineEventHandler(async (event) => {
  if (event.method !== 'POST') {
    throw createError({ statusCode: 405, message: 'Method Not Allowed' })
  }

  const config = useRuntimeConfig()
  const payload = await readBody(event)

  return $fetch(`${config.coreApiUrl}/documentos/backfill`, {
    method: 'POST',
    body: payload
  })
})
