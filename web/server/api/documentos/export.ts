type ExportErrorResponse = {
  message?: string
  error?: string
}

export default defineEventHandler(async (event) => {
  if (event.method !== 'POST') {
    throw createError({ statusCode: 405, message: 'Method Not Allowed' })
  }

  const config = useRuntimeConfig()
  const payload = await readBody(event)

  const response = await fetch(`${config.coreApiUrl}/documentos/export`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  const contentType = response.headers.get('content-type') || ''

  if (contentType.includes('application/json')) {
    const body = await response.json() as ExportErrorResponse

    if (!response.ok) {
      throw createError({
        statusCode: response.status,
        message: body.message || body.error || 'Erro ao exportar documentos',
        data: body
      })
    }

    return body
  }

  if (!response.ok) {
    throw createError({
      statusCode: response.status,
      message: 'Erro ao exportar documentos'
    })
  }

  const binary = new Uint8Array(await response.arrayBuffer())

  setResponseStatus(event, response.status)

  for (const headerName of ['content-type', 'content-disposition', 'x-export-total', 'x-export-xml', 'x-export-danfe', 'x-export-skipped-danfe']) {
    const headerValue = response.headers.get(headerName)
    if (headerValue) {
      setHeader(event, headerName, headerValue)
    }
  }

  return binary
})
