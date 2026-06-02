<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { NavigationItem } from '../types/navigation'
import AppBadge from '../components/AppBadge.vue'
import AppDrawer from '../components/AppDrawer.vue'

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()

const navItems: NavigationItem[] = [
  { label: 'Dashboard', to: '/dashboard' },
  { label: 'POS', to: '/pos', roles: ['ADMIN', 'CASHIER'] },
  { label: 'Products', to: '/products', roles: ['ADMIN', 'MANAGER', 'CASHIER'] },
  { label: 'Categories', to: '/categories', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Restock', to: '/restock', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Stock Movements', to: '/stock-movements', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Locations', to: '/locations', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Transfers', to: '/transfers', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Sales History', to: '/sales-history', roles: ['ADMIN', 'CASHIER'] },
  { label: 'Receipt Detail', to: '/receipt-detail', roles: ['ADMIN', 'CASHIER'] },
  { label: 'Alerts', to: '/alerts', roles: ['ADMIN', 'MANAGER', 'CASHIER'] },
  { label: 'Reports', to: '/reports', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Exports', to: '/exports', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Imports', to: '/imports', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Purchase Orders', to: '/purchase-orders', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Suppliers', to: '/suppliers', roles: ['ADMIN', 'MANAGER'] },
  { label: 'Users', to: '/users', roles: ['ADMIN'] },
  { label: 'Settings', to: '/settings', roles: ['ADMIN'] },
]

const visibleNavItems = computed(() => navItems.filter((item) => auth.can(item.roles)))

function isActive(to: string) {
  return route.path === to
}

async function logout() {
  await auth.logout()
  window.location.assign('/login')
}
</script>

<template>
  <div class="min-h-screen lg:grid lg:grid-cols-[280px_1fr]">
    <aside class="hidden border-r border-slate-200 bg-white lg:block">
      <div class="sticky top-0 flex h-screen flex-col">
        <div class="border-b border-slate-200 p-5">
          <p class="text-xs font-semibold uppercase text-brand-700">Grocery POS</p>
          <h1 class="mt-1 text-lg font-bold">Inventory System</h1>
        </div>
        <nav class="flex-1 overflow-y-auto p-3">
          <RouterLink
            v-for="item in visibleNavItems"
            :key="item.to"
            :to="item.to"
            class="block rounded-md px-3 py-2 text-sm font-medium text-slate-600 hover:bg-brand-50 hover:text-brand-700"
            :class="{ 'bg-brand-600 text-white hover:bg-brand-600 hover:text-white': isActive(item.to) }"
          >
            {{ item.label }}
          </RouterLink>
        </nav>
      </div>
    </aside>

    <AppDrawer :open="app.sidebarOpen" title="Navigation" @close="app.closeSidebar">
      <nav class="grid gap-1">
        <RouterLink
          v-for="item in visibleNavItems"
          :key="item.to"
          :to="item.to"
          class="rounded-md px-3 py-2 text-sm font-medium text-slate-600 hover:bg-brand-50"
          :class="{ 'bg-brand-600 text-white': isActive(item.to) }"
          @click="app.closeSidebar"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
    </AppDrawer>

    <div class="min-w-0">
      <header class="sticky top-0 z-20 border-b border-slate-200 bg-white/95 backdrop-blur">
        <div class="flex min-h-16 items-center justify-between gap-3 px-4 lg:px-6">
          <button class="rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold lg:hidden" @click="app.openSidebar">
            Menu
          </button>
          <div class="min-w-0">
            <p class="truncate text-sm text-slate-500">Small grocery operations</p>
            <p class="truncate font-semibold">Foundation workspace</p>
          </div>
          <div class="flex items-center gap-3">
            <button class="relative rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold">
              Alerts
              <AppBadge v-if="app.alertCount" class="absolute -right-2 -top-2">{{ app.alertCount }}</AppBadge>
            </button>
            <button class="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold">
              <span class="grid h-7 w-7 place-items-center rounded-full bg-brand-100 text-xs text-brand-700">{{ auth.userInitials }}</span>
              <span class="hidden sm:inline">{{ auth.user?.username }} · {{ auth.user?.role }}</span>
            </button>
            <button class="rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold" @click="logout">Logout</button>
          </div>
        </div>
      </header>
      <main class="px-4 py-6 lg:px-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>
