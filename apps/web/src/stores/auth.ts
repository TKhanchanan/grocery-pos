import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { apiClient, postJSON } from '../api/client'
import type { AssignedRole, AuthMeResponse, PermissionCode, Role, User } from '../types/navigation'

interface LoginResponse {
  token: string
  user: User
  roles?: AssignedRole[]
  permissions?: PermissionCode[]
}

function readStoredJSON<T>(key: string, fallback: T): T {
  const raw = localStorage.getItem(key)
  if (!raw || raw === 'undefined' || raw === 'null') return fallback
  try {
    return JSON.parse(raw) as T
  } catch {
    localStorage.removeItem(key)
    return fallback
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('auth_token'))
  const user = ref<User | null>(readStoredJSON<User | null>('auth_user', null))
  const roles = ref<AssignedRole[]>(readStoredJSON<AssignedRole[]>('auth_roles', []))
  const permissions = ref<PermissionCode[]>(readStoredJSON<PermissionCode[]>('auth_permissions', []))
  const isAuthenticated = computed(() => Boolean(token.value && user.value))
  const userInitials = computed(() => {
    const name = user.value?.fullName || user.value?.username || 'User'
    return name.split(' ').map((part) => part[0]).join('').slice(0, 2).toUpperCase()
  })

  async function login(username: string, password: string) {
    const result = await postJSON<LoginResponse>('/v1/auth/login', { username, password })
    token.value = result.token
    user.value = result.user
    roles.value = result.roles ?? result.user.roles ?? []
    permissions.value = result.permissions ?? []
    localStorage.setItem('auth_token', result.token)
    localStorage.setItem('auth_user', JSON.stringify(result.user))
    localStorage.setItem('auth_roles', JSON.stringify(roles.value))
    localStorage.setItem('auth_permissions', JSON.stringify(permissions.value))
  }

  async function logout() {
    if (token.value) {
      await postJSON('/v1/auth/logout', {}).catch(() => undefined)
    }
    token.value = null
    user.value = null
    roles.value = []
    permissions.value = []
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
    localStorage.removeItem('auth_roles')
    localStorage.removeItem('auth_permissions')
  }

  async function loadMe() {
    if (!token.value) return
    const result = await apiClient<AuthMeResponse>('/v1/auth/me')
    user.value = result.user
    roles.value = result.roles
    permissions.value = result.permissions
    localStorage.setItem('auth_user', JSON.stringify(user.value))
    localStorage.setItem('auth_roles', JSON.stringify(roles.value))
    localStorage.setItem('auth_permissions', JSON.stringify(permissions.value))
  }

  function can(roles?: Role[]) {
    if (!roles || roles.length === 0) return true
    return Boolean(user.value && roles.includes(user.value.role))
  }

  function hasPermission(code?: PermissionCode) {
    if (!code) return true
    return permissions.value.includes(code)
  }

  function hasAnyPermission(codes?: PermissionCode[]) {
    if (!codes || codes.length === 0) return true
    return codes.some((code) => hasPermission(code))
  }

  function hasAllPermissions(codes?: PermissionCode[]) {
    if (!codes || codes.length === 0) return true
    return codes.every((code) => hasPermission(code))
  }

  function canViewMenu(item: { roles?: Role[]; permission?: PermissionCode; permissions?: PermissionCode[] }) {
    if (item.permission) return hasPermission(item.permission)
    if (item.permissions) return hasAnyPermission(item.permissions)
    return can(item.roles)
  }

  return { token, user, roles, permissions, isAuthenticated, userInitials, login, logout, loadMe, can, hasPermission, hasAnyPermission, hasAllPermissions, canViewMenu }
})
