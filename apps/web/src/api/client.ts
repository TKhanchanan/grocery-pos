import { authHeaders, handleAuthFailure, readAuthToken } from './session'

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

interface CacheEntry {
  expiresAt: number
  value: unknown
}

const responseCache = new Map<string, CacheEntry>()
const pendingRequests = new Map<string, Promise<unknown>>()
let cacheGeneration = 0

const TTL = {
  fiveMinutes: 5 * 60_000,
  thirtySeconds: 30_000,
  fortyFiveSeconds: 45_000,
  tenSeconds: 10_000,
}

export function assetURL(path?: string | null) {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://') || path.startsWith('data:') || path.startsWith('blob:')) return path
  return `${API_ORIGIN}${path.startsWith('/') ? path : `/${path}`}`
}

function requiresAuth(path: string) {
  return path !== '/v1/auth/login'
}

function cacheTTL(path: string) {
  const pathname = path.split('?')[0]
  if (pathname === '/v1/locations' || pathname === '/v1/categories' || pathname === '/v1/pos/categories' || pathname === '/v1/receipt-settings') {
    return TTL.fiveMinutes
  }
  if (pathname === '/v1/alerts' || pathname === '/v1/alerts/unread-count') return TTL.thirtySeconds
  if (pathname === '/v1/products' || pathname === '/v1/pos/products') return TTL.fortyFiveSeconds
  if (pathname === '/v1/dashboard/summary') return TTL.tenSeconds
  return 0
}

function requestKey(path: string) {
  return `${readAuthToken() ?? 'public'}:${path}`
}

function invalidationPrefixes(path: string) {
  if (path.startsWith('/v1/categories')) {
    return ['/v1/categories', '/v1/pos/categories', '/v1/products', '/v1/pos/products', '/v1/dashboard/summary']
  }
  if (path.startsWith('/v1/locations')) {
    return ['/v1/locations', '/v1/products', '/v1/pos/products', '/v1/dashboard/summary']
  }
  if (path.startsWith('/v1/settings')) {
    return ['/v1/settings', '/v1/receipt-settings']
  }
  if (path.startsWith('/v1/alerts')) {
    return ['/v1/alerts', '/v1/dashboard/summary']
  }
  if (path.startsWith('/v1/products') || path.startsWith('/v1/imports/products')) {
    return ['/v1/products', '/v1/pos/products', '/v1/categories', '/v1/pos/categories', '/v1/locations', '/v1/alerts', '/v1/dashboard/summary', '/v1/reports']
  }
  if (path.startsWith('/v1/sales') || path.startsWith('/v1/stock-transfers')) {
    return ['/v1/products', '/v1/pos/products', '/v1/alerts', '/v1/dashboard/summary', '/v1/reports', '/v1/sales', '/v1/stock-transfers']
  }
  return []
}

export function invalidateApiCache(...prefixes: string[]) {
  cacheGeneration += 1
  if (prefixes.length === 0) {
    responseCache.clear()
    pendingRequests.clear()
    return
  }
  for (const key of responseCache.keys()) {
    if (prefixes.some((prefix) => key.includes(`:${prefix}`))) responseCache.delete(key)
  }
  for (const key of pendingRequests.keys()) {
    if (prefixes.some((prefix) => key.includes(`:${prefix}`))) pendingRequests.delete(key)
  }
}

export async function apiClient<T>(path: string, init: RequestInit = {}): Promise<T> {
  const shouldRequireAuth = requiresAuth(path)
  const method = (init.method ?? 'GET').toUpperCase()
  const key = requestKey(path)
  const requestGeneration = cacheGeneration
  const ttl = method === 'GET' ? cacheTTL(path) : 0
  const cached = ttl > 0 ? responseCache.get(key) : undefined
  if (cached && cached.expiresAt > Date.now()) return cached.value as T
  if (cached) responseCache.delete(key)
  if (method === 'GET') {
    const pending = pendingRequests.get(key)
    if (pending) return pending as Promise<T>
  }

  const headers = shouldRequireAuth ? authHeaders(init.headers) : new Headers(init.headers)
  if (!(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  const request = (async () => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers,
      cache: init.cache ?? (method === 'GET' ? 'no-store' : undefined),
    })

    const envelope = (await response.json().catch(() => ({ success: false, error: { message: response.statusText } }))) as ApiEnvelope<T>
    if (shouldRequireAuth && response.status === 401) {
      invalidateApiCache()
      throw handleAuthFailure(envelope.error?.message)
    }
    if (!response.ok || !envelope.success) {
      throw new Error(envelope.error?.message ?? 'API request failed')
    }

    const result = envelope.data as T
    if (ttl > 0 && requestGeneration === cacheGeneration) {
      responseCache.set(key, { expiresAt: Date.now() + ttl, value: result })
    }
    if (method !== 'GET') invalidateApiCache(...invalidationPrefixes(path))
    return result
  })()

  if (method === 'GET') pendingRequests.set(key, request)
  try {
    return await request
  } finally {
    if (method === 'GET' && pendingRequests.get(key) === request) pendingRequests.delete(key)
  }
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
