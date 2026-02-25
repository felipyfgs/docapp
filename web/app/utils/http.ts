export function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message) {
    return error.message
  }

  if (typeof error === 'object' && error !== null) {
    const maybeData = error as { data?: { message?: string }, statusMessage?: string, message?: string }
    return maybeData.data?.message || maybeData.statusMessage || maybeData.message || fallback
  }

  return fallback
}

export function extractFileName(contentDisposition: string | null, fallbackPrefix = 'download'): string {
  if (!contentDisposition) {
    return `${fallbackPrefix}_${Date.now()}.zip`
  }

  const match = contentDisposition.match(/filename="?([^";]+)"?/i)
  return match?.[1] || `${fallbackPrefix}_${Date.now()}.zip`
}

export function downloadBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
