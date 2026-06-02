import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { apiClient } from '../api/client'
import { applyDocumentPreferences, readStoredLanguage, readStoredTheme, translateMessage, type AppLanguage, type AppTheme, type TranslationKey } from '../i18n'
import type { InventoryAlert } from '../types/navigation'

export const useAppStore = defineStore('app', () => {
  const sidebarOpen = ref(false)
  const alertCount = ref(0)
  const userName = ref('Demo User')
  const language = ref<AppLanguage>(readStoredLanguage())
  const theme = ref<AppTheme>(readStoredTheme())

  const userInitials = computed(() => userName.value.split(' ').map((part) => part[0]).join('').slice(0, 2).toUpperCase())
  const isDark = computed(() => theme.value === 'dark')

  function openSidebar() {
    sidebarOpen.value = true
  }

  function closeSidebar() {
    sidebarOpen.value = false
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

  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  function t(key: TranslationKey) {
    return translateMessage(language.value, key)
  }

  watch([language, theme], () => {
    localStorage.setItem('app_language', language.value)
    localStorage.setItem('app_theme', theme.value)
    applyDocumentPreferences(language.value, theme.value)
  }, { immediate: true })

  return {
    sidebarOpen,
    alertCount,
    userName,
    userInitials,
    language,
    theme,
    isDark,
    openSidebar,
    closeSidebar,
    loadAlertCount,
    setLanguage,
    setTheme,
    toggleTheme,
    t,
  }
})
