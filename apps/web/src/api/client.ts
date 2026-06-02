export interface ApiEnvelope<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'

export async function apiClient<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('auth_token')
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
  })

  const envelope = (await response.json().catch(() => ({ success: false, error: { message: response.statusText } }))) as ApiEnvelope<T>
  if (!response.ok || !envelope.success) {
    throw new Error(envelope.error?.message ?? 'API request failed')
  }
  return envelope.data as T
}

export function postJSON<T>(path: string, body: object): Promise<T> {
  return apiClient<T>(path, { method: 'POST', body: JSON.stringify(body) })
}

export function patchJSON<T>(path: string, body: object): Promise<T> {
  return apiClient<T>(path, { method: 'PATCH', body: JSON.stringify(body) })
}
