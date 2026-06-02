<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('admin')
const password = ref('password')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    router.push('/')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="grid min-h-screen place-items-center bg-[radial-gradient(circle_at_top_left,#e8f6ec,transparent_30%),#fffdf5] px-4">
    <form class="w-full max-w-sm rounded-lg border border-emerald-100 bg-white p-6 shadow-sm" @submit.prevent="submit">
      <p class="label">Demo system</p>
      <h1 class="mt-1 text-2xl font-bold text-leaf">Grocery POS Login</h1>
      <p class="mt-2 text-sm text-slate-600">Use admin, manager, or cashier with password <b>password</b>.</p>
      <label class="mt-6 block">
        <span class="label">Username</span>
        <input v-model="username" class="input mt-1" autocomplete="username" />
      </label>
      <label class="mt-4 block">
        <span class="label">Password</span>
        <input v-model="password" class="input mt-1" type="password" autocomplete="current-password" />
      </label>
      <p v-if="error" class="mt-4 rounded-md bg-red-50 p-3 text-sm text-red-700">{{ error }}</p>
      <button class="btn-primary mt-6 w-full" :disabled="loading">{{ loading ? 'Signing in...' : 'Login' }}</button>
    </form>
  </main>
</template>
