<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
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
  { labelKey: 'nav.dashboard', to: '/dashboard' },
  { labelKey: 'nav.pos', to: '/pos', roles: ['ADMIN', 'CASHIER'] },
  { labelKey: 'nav.products', to: '/products', roles: ['ADMIN', 'MANAGER', 'CASHIER'] },
  { labelKey: 'nav.categories', to: '/categories', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.restock', to: '/restock', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.stockMovements', to: '/stock-movements', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.locations', to: '/locations', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.transfers', to: '/transfers', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.salesHistory', to: '/sales-history', roles: ['ADMIN', 'MANAGER', 'CASHIER'] },
  { labelKey: 'nav.receiptDetail', to: '/receipt-detail', roles: ['ADMIN', 'MANAGER', 'CASHIER'] },
  { labelKey: 'nav.alerts', to: '/alerts', roles: ['ADMIN', 'MANAGER', 'CASHIER'] },
  { labelKey: 'nav.reports', to: '/reports', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.exports', to: '/exports', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.imports', to: '/imports', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.purchaseOrders', to: '/purchase-orders', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.suppliers', to: '/suppliers', roles: ['ADMIN', 'MANAGER'] },
  { labelKey: 'nav.users', to: '/users', roles: ['ADMIN'] },
  { labelKey: 'nav.settings', to: '/settings', roles: ['ADMIN'] },
]

const visibleNavItems = computed(() => navItems.filter((item) => auth.can(item.roles)))

function isActive(to: string) {
  return route.path === to
}

async function logout() {
  await auth.logout()
  window.location.assign('/login')
}

onMounted(() => {
  app.loadAlertCount()
})

watch(() => route.path, () => {
  app.loadAlertCount()
})
</script>

<template>
  <div class="min-h-screen lg:grid lg:grid-cols-[280px_1fr]">
    <aside class="hidden border-r border-slate-200 bg-white lg:block">
      <div class="sticky top-0 flex h-screen flex-col">
        <div class="border-b border-slate-200 p-5">
          <p class="text-xs font-semibold uppercase text-brand-700">{{ app.t('app.name') }}</p>
          <h1 class="mt-1 text-lg font-bold">{{ app.t('app.subtitle') }}</h1>
        </div>
        <nav class="flex-1 overflow-y-auto p-3">
          <RouterLink
            v-for="item in visibleNavItems"
            :key="item.to"
            :to="item.to"
            class="block rounded-md px-3 py-2 text-sm font-medium text-slate-600 hover:bg-brand-50 hover:text-brand-700"
            :class="{ 'bg-brand-600 text-white hover:bg-brand-600 hover:text-white': isActive(item.to) }"
          >
            <span class="flex items-center justify-between gap-2">
              <span>{{ app.t(item.labelKey) }}</span>
              <AppBadge v-if="item.to === '/alerts' && app.alertCount">{{ app.alertCount }}</AppBadge>
            </span>
          </RouterLink>
        </nav>
      </div>
    </aside>

    <AppDrawer :open="app.sidebarOpen" :title="app.t('topbar.navigation')" :close-label="app.t('topbar.close')" @close="app.closeSidebar">
      <nav class="grid gap-1">
        <RouterLink
          v-for="item in visibleNavItems"
          :key="item.to"
          :to="item.to"
          class="rounded-md px-3 py-2 text-sm font-medium text-slate-600 hover:bg-brand-50"
          :class="{ 'bg-brand-600 text-white': isActive(item.to) }"
          @click="app.closeSidebar"
        >
          <span class="flex items-center justify-between gap-2">
            <span>{{ app.t(item.labelKey) }}</span>
            <AppBadge v-if="item.to === '/alerts' && app.alertCount">{{ app.alertCount }}</AppBadge>
          </span>
        </RouterLink>
      </nav>
    </AppDrawer>

    <div class="min-w-0">
      <header class="sticky top-0 z-20 border-b border-slate-200 bg-white/95 backdrop-blur">
        <div class="flex min-h-16 items-center justify-between gap-3 px-4 lg:px-6">
          <button class="rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold lg:hidden" @click="app.openSidebar">
            {{ app.t('topbar.menu') }}
          </button>
          <div class="min-w-0">
            <p class="truncate text-sm text-slate-500">{{ app.t('topbar.kicker') }}</p>
            <p class="truncate font-semibold">{{ app.t('topbar.workspace') }}</p>
          </div>
          <div class="flex items-center gap-3">
            <RouterLink to="/alerts" class="relative rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold">
              {{ app.t('topbar.alerts') }}
              <AppBadge v-if="app.alertCount" class="absolute -right-2 -top-2">{{ app.alertCount }}</AppBadge>
            </RouterLink>
            <div class="hidden items-center gap-1 rounded-md border border-slate-200 bg-white p-1 sm:flex">
              <button
                class="rounded px-2 py-1 text-xs font-bold"
                :class="app.language === 'th' ? 'bg-brand-600 text-white' : 'text-slate-600'"
                @click="app.setLanguage('th')"
              >
                TH
              </button>
              <button
                class="rounded px-2 py-1 text-xs font-bold"
                :class="app.language === 'en' ? 'bg-brand-600 text-white' : 'text-slate-600'"
                @click="app.setLanguage('en')"
              >
                EN
              </button>
            </div>
            <button class="rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold" @click="app.toggleTheme">
              {{ app.isDark ? app.t('settings.light') : app.t('settings.dark') }}
            </button>
            <button class="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold">
              <span class="grid h-7 w-7 place-items-center rounded-full bg-brand-100 text-xs text-brand-700">{{ auth.userInitials }}</span>
              <span class="hidden sm:inline">{{ auth.user?.username }} · {{ auth.user?.role }}</span>
            </button>
            <button class="rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold" @click="logout">{{ app.t('topbar.logout') }}</button>
          </div>
        </div>
      </header>
      <main class="px-4 py-6 lg:px-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>
