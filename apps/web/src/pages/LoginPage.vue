<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import AuthLayout from '../layouts/AuthLayout.vue'
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
    router.push('/dashboard')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthLayout>
    <form class="grid gap-4" @submit.prevent="submit">
      <div>
        <p class="text-xs font-bold uppercase text-brand-700">Secure access</p>
        <h1 class="mt-1 text-2xl font-bold">Login</h1>
        <p class="mt-2 text-sm text-slate-500">Demo users: admin, manager, cashier. Password: password.</p>
      </div>
      <AppInput v-model="username" label="Username" autocomplete="username" />
      <AppInput v-model="password" label="Password" type="password" autocomplete="current-password" />
      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      <AppButton type="submit" :disabled="loading">{{ loading ? 'Logging in...' : 'Login' }}</AppButton>
    </form>
  </AuthLayout>
</template>
