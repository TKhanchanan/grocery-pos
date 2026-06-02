import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { apiClient } from '../api/client'
import type { InventoryAlert } from '../types/navigation'

export const useAppStore = defineStore('app', () => {
  const sidebarOpen = ref(false)
  const alertCount = ref(0)
  const userName = ref('Demo User')

  const userInitials = computed(() => userName.value.split(' ').map((part) => part[0]).join('').slice(0, 2).toUpperCase())

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

  return { sidebarOpen, alertCount, userName, userInitials, openSidebar, closeSidebar, loadAlertCount }
})
