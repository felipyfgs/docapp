export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')
  const formData = await readFormData(event)

  return $fetch(`${config.coreApiUrl}/empresas/${id}/certificado`, {
    method: 'POST',
    body: formData
  })
})
