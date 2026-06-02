import type { LoginResponse } from './types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api'

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token')
  const headers = new Headers(options.headers)
  if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
    headers.set('Content-Type', 'application/json')
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  const res = await fetch(`${API_URL}${path}`, { ...options, headers })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(String(err.error ?? res.statusText))
  }
  if (res.headers.get('Content-Type')?.includes('text/csv')) {
    return (await res.text()) as T
  }
  return (await res.json()) as T
}

export function postJSON<T>(path: string, body: object): Promise<T> {
  return api<T>(path, { method: 'POST', body: JSON.stringify(body) })
}

export function putJSON<T>(path: string, body: object): Promise<T> {
  return api<T>(path, { method: 'PUT', body: JSON.stringify(body) })
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  return postJSON<LoginResponse>('/auth/login', { username, password })
}
