<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppButton from '../components/AppButton.vue'
import AppIcon from '../components/AppIcon.vue'
import AppInput from '../components/AppInput.vue'
import AuthLayout from '../layouts/AuthLayout.vue'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'

const app = useAppStore()
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
        <div class="flex items-start justify-between gap-3">
          <div class="flex items-start gap-3">
            <div class="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-brand-600 text-white shadow-lg shadow-brand-600/20">
              <AppIcon name="shopping-cart" />
            </div>
            <div>
            <p class="text-xs font-bold uppercase text-brand-700">{{ app.t('login.secureAccess') }}</p>
            <h1 class="mt-1 text-2xl font-bold">{{ app.t('login.title') }}</h1>
            </div>
          </div>
          <div class="flex items-center gap-1 rounded-md border border-slate-200 bg-white p-1">
            <button type="button" class="rounded px-2 py-1 text-xs font-bold" :class="app.language === 'th' ? 'bg-brand-600 text-white' : 'text-slate-600'" @click="app.setLanguage('th')">TH</button>
            <button type="button" class="rounded px-2 py-1 text-xs font-bold" :class="app.language === 'en' ? 'bg-brand-600 text-white' : 'text-slate-600'" @click="app.setLanguage('en')">EN</button>
          </div>
        </div>
        <p class="mt-2 text-sm text-slate-500">{{ app.t('login.demoUsers') }}</p>
      </div>
      <AppInput v-model="username" :label="app.t('login.username')" autocomplete="username" />
      <AppInput v-model="password" :label="app.t('login.password')" type="password" autocomplete="current-password" />
      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
        <AppButton type="submit" :loading="loading" :disabled="loading" icon="check-circle">{{ app.t('login.submit') }}</AppButton>
        <AppButton variant="secondary" @click="app.toggleTheme">{{ app.isDark ? app.t('settings.light') : app.t('settings.dark') }}</AppButton>
      </div>
    </form>
  </AuthLayout>
</template>
