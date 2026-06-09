import { API_BASE_URL } from './client'
import { authHeaders, handleAuthFailure } from './session'

export async function downloadFile(path: string, fallbackName: string) {
  const headers = authHeaders()

  const response = await fetch(`${API_BASE_URL}${path}`, { headers })
  if (!response.ok) {
    const payload = await response.json().catch(() => null)
    if (response.status === 401) throw handleAuthFailure(payload?.error?.message)
    throw new Error(payload?.error?.message ?? response.statusText)
  }

  const blob = await response.blob()
  const disposition = response.headers.get('content-disposition') ?? ''
  const match = disposition.match(/filename="([^"]+)"/)
  const filename = match?.[1] ?? fallbackName
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
