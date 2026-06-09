const AUTH_STORAGE_KEYS = ['auth_token', 'auth_user', 'auth_roles', 'auth_permissions']

let redirectingToLogin = false

export class AuthSessionError extends Error {
  constructor(message = 'Invalid or expired token') {
    super(message)
    this.name = 'AuthSessionError'
  }
}

export function readAuthToken() {
  return localStorage.getItem('auth_token')
}

export function clearAuthSession() {
  AUTH_STORAGE_KEYS.forEach((key) => localStorage.removeItem(key))
}

export function redirectToLogin() {
  if (typeof window === 'undefined' || redirectingToLogin || window.location.pathname === '/login') return
  redirectingToLogin = true
  window.location.replace('/login')
}

export function handleAuthFailure(message?: string) {
  clearAuthSession()
  redirectToLogin()
  return new AuthSessionError(message)
}

export function authHeaders(headersInit?: HeadersInit) {
  const token = readAuthToken()
  if (!token) throw handleAuthFailure('Missing authentication token')

  const headers = new Headers(headersInit)
  headers.set('Authorization', `Bearer ${token}`)
  return headers
}
