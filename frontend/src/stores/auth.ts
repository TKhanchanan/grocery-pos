import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { login as loginApi } from '../api'
import type { LoginResponse, Role, User } from '../types'

export type { LoginResponse }

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') ?? 'null') as User | null)
  const isLoggedIn = computed(() => Boolean(token.value && user.value))

  async function login(username: string, password: string) {
    const out = await loginApi(username, password)
    token.value = out.token
    user.value = out.user
    localStorage.setItem('token', out.token)
    localStorage.setItem('user', JSON.stringify(out.user))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function can(roles: Role[]) {
    if (!user.value) return false
    return roles.includes(user.value.role)
  }

  return { token, user, isLoggedIn, login, logout, can }
})
