import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { apiClient, postJSON } from '../api/client'
import type { Role, User } from '../types/navigation'

interface LoginResponse {
  token: string
  user: User
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('auth_token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('auth_user') ?? 'null') as User | null)
  const isAuthenticated = computed(() => Boolean(token.value && user.value))
  const userInitials = computed(() => {
    const name = user.value?.fullName || user.value?.username || 'User'
    return name.split(' ').map((part) => part[0]).join('').slice(0, 2).toUpperCase()
  })

  async function login(username: string, password: string) {
    const result = await postJSON<LoginResponse>('/v1/auth/login', { username, password })
    token.value = result.token
    user.value = result.user
    localStorage.setItem('auth_token', result.token)
    localStorage.setItem('auth_user', JSON.stringify(result.user))
  }

  async function logout() {
    if (token.value) {
      await postJSON('/v1/auth/logout', {}).catch(() => undefined)
    }
    token.value = null
    user.value = null
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
  }

  async function loadMe() {
    if (!token.value) return
    user.value = await apiClient<User>('/v1/auth/me')
    localStorage.setItem('auth_user', JSON.stringify(user.value))
  }

  function can(roles?: Role[]) {
    if (!roles || roles.length === 0) return true
    return Boolean(user.value && roles.includes(user.value.role))
  }

  return { token, user, isAuthenticated, userInitials, login, logout, loadMe, can }
})
