import { authHeaders, handleAuthFailure } from './session'

export interface ApiEnvelope<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
}

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
export const API_ORIGIN = API_BASE_URL.replace(/\/api\/?$/, '')

export function assetURL(path?: string | null) {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://') || path.startsWith('data:') || path.startsWith('blob:')) return path
  return `${API_ORIGIN}${path.startsWith('/') ? path : `/${path}`}`
}

function requiresAuth(path: string) {
  return path !== '/v1/auth/login'
}

export async function apiClient<T>(path: string, init: RequestInit = {}): Promise<T> {
  const shouldRequireAuth = requiresAuth(path)
  const headers = shouldRequireAuth ? authHeaders(init.headers) : new Headers(init.headers)
  if (!(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
  })

  const envelope = (await response.json().catch(() => ({ success: false, error: { message: response.statusText } }))) as ApiEnvelope<T>
  if (shouldRequireAuth && response.status === 401) {
    throw handleAuthFailure(envelope.error?.message)
  }
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

export function deleteJSON<T>(path: string): Promise<T> {
  return apiClient<T>(path, { method: 'DELETE' })
}
