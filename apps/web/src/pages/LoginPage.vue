<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import logoUrl from '../assets/logo.png'
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
    <template #brand>
      <div class="grid gap-4">
        <div>
          <p class="text-sm font-bold uppercase text-teal-100/90">{{ app.t('app.name') }}</p>
          <h2 class="mt-2 max-w-sm text-4xl font-black leading-tight">{{ app.t('app.subtitle') }}</h2>
        </div>
      </div>
    </template>

    <form class="grid gap-6" @submit.prevent="submit">
      <div class="flex items-start justify-between gap-4">
        <div class="flex min-w-0 items-start gap-3">
          <span class="grid h-14 w-14 shrink-0 place-items-center rounded-xl bg-brand-50 shadow-sm dark:bg-teal-300/10">
            <img class="h-11 w-11 object-contain" :src="logoUrl" :alt="app.t('app.name')" />
          </span>
          <div class="min-w-0">
            <h1 class="mt-1 text-3xl font-black tracking-normal">{{ app.t('login.title') }}</h1>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1 rounded-xl bg-slate-100/90 p-1 shadow-sm dark:bg-slate-900/80">
          <button type="button" class="grid h-9 min-w-10 place-items-center rounded-lg px-2 text-xs font-black transition" :class="app.language === 'th' ? 'bg-white text-brand-700 shadow-sm dark:bg-teal-300 dark:text-slate-950' : 'text-slate-600 hover:bg-white/70 dark:text-slate-300 dark:hover:bg-slate-800'" @click="app.setLanguage('th')">TH</button>
          <button type="button" class="grid h-9 min-w-10 place-items-center rounded-lg px-2 text-xs font-black transition" :class="app.language === 'en' ? 'bg-white text-brand-700 shadow-sm dark:bg-teal-300 dark:text-slate-950' : 'text-slate-600 hover:bg-white/70 dark:text-slate-300 dark:hover:bg-slate-800'" @click="app.setLanguage('en')">EN</button>
          <button type="button" class="grid h-9 w-9 place-items-center rounded-lg text-slate-600 transition hover:bg-white/70 dark:text-slate-300 dark:hover:bg-slate-800" :aria-label="app.isDark ? app.t('settings.light') : app.t('settings.dark')" @click="app.toggleTheme">
            <AppIcon :name="app.isDark ? 'sun' : 'moon'" :size="18" />
          </button>
        </div>
      </div>

      <div class="grid gap-4">
        <p class="rounded-xl bg-brand-50 px-4 py-3 text-sm font-semibold text-slate-600 shadow-sm dark:bg-slate-900/80 dark:text-slate-300">{{ app.t('login.demoUsers') }}</p>
        <div class="grid gap-4">
          <AppInput v-model="username" :label="app.t('login.username')" autocomplete="username" />
          <AppInput v-model="password" :label="app.t('login.password')" type="password" autocomplete="current-password" />
        </div>
        <div v-if="error" class="rounded-xl bg-red-50 px-4 py-3 text-sm font-semibold text-red-700 dark:bg-red-500/15 dark:text-red-200">{{ error }}</div>
        <AppButton class="w-full" type="submit" :loading="loading" :disabled="loading" icon="check-circle">{{ app.t('login.submit') }}</AppButton>
      </div>
    </form>
  </AuthLayout>
</template>
