import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

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

  return { sidebarOpen, alertCount, userName, userInitials, openSidebar, closeSidebar }
})
