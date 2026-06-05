import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { apiClient } from '../api/client'
import { applyDocumentPreferences, readStoredLanguage, readStoredTheme, translateMessage, type AppLanguage, type AppTheme, type TranslationKey } from '../i18n'
import type { InventoryAlert } from '../types/navigation'

type TextSize = 'sm' | 'base' | 'lg' | 'xl'
type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface ToastMessage {
  id: number
  type: ToastType
  message: string
  description?: string
  resultModal?: boolean
}

function readStoredTextSize(): TextSize {
  const value = localStorage.getItem('app_text_size')
  return value === 'sm' || value === 'lg' || value === 'xl' ? value : 'base'
}

function readStoredSidebarCollapsed() {
  return localStorage.getItem('sidebar_collapsed') === 'true'
}

export const useAppStore = defineStore('app', () => {
  const sidebarOpen = ref(false)
  const sidebarCollapsed = ref(readStoredSidebarCollapsed())
  const alertCount = ref(0)
  const userName = ref('Demo User')
  const language = ref<AppLanguage>(readStoredLanguage())
  const theme = ref<AppTheme>(readStoredTheme())
  const textSize = ref<TextSize>(readStoredTextSize())
  const toasts = ref<ToastMessage[]>([])
  let toastID = 0

  const userInitials = computed(() => userName.value.split(' ').map((part) => part[0]).join('').slice(0, 2).toUpperCase())
  const isDark = computed(() => theme.value === 'dark')

  function openSidebar() {
    sidebarOpen.value = true
  }

  function closeSidebar() {
    sidebarOpen.value = false
  }

  function toggleSidebarCollapsed() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  async function loadAlertCount() {
    const alerts = await apiClient<InventoryAlert[]>('/v1/alerts?unread=true').catch(() => [])
    alertCount.value = alerts.length
  }

  function setLanguage(value: AppLanguage) {
    language.value = value
  }

  function setTheme(value: AppTheme) {
    theme.value = value
  }

  function setTextSize(value: TextSize) {
    textSize.value = value
  }

  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  function t(key: TranslationKey) {
    return translateMessage(language.value, key)
  }

  function pushToast(input: ToastType | { type?: ToastType; message: string; description?: string; resultModal?: boolean }, message?: string, description?: string) {
    const toast = typeof input === 'string'
      ? { type: input, message: message ?? '', description }
      : { type: input.type ?? 'success', message: input.message, description: input.description, resultModal: input.resultModal }
    const id = ++toastID
    toasts.value.push({ id, ...toast })
    window.setTimeout(() => removeToast(id), 4200)
    return id
  }

  function removeToast(id: number) {
    toasts.value = toasts.value.filter((toast) => toast.id !== id)
  }

  watch([language, theme, textSize], () => {
    localStorage.setItem('app_language', language.value)
    localStorage.setItem('app_theme', theme.value)
    localStorage.setItem('app_text_size', textSize.value)
    applyDocumentPreferences(language.value, theme.value)
    document.documentElement.dataset.textSize = textSize.value
  }, { immediate: true })

  watch(sidebarCollapsed, () => {
    localStorage.setItem('sidebar_collapsed', String(sidebarCollapsed.value))
  }, { immediate: true })

  return {
    sidebarOpen,
    sidebarCollapsed,
    alertCount,
    userName,
    userInitials,
    language,
    theme,
    textSize,
    toasts,
    isDark,
    openSidebar,
    closeSidebar,
    toggleSidebarCollapsed,
    loadAlertCount,
    setLanguage,
    setTheme,
    setTextSize,
    toggleTheme,
    t,
    pushToast,
    removeToast,
  }
})
