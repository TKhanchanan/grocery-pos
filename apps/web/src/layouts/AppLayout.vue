<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { NavigationItem } from '../types/navigation'
import AppBadge from '../components/AppBadge.vue'
import AppDrawer from '../components/AppDrawer.vue'
import AppIcon from '../components/AppIcon.vue'

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()

const navItems: NavigationItem[] = [
  { labelKey: 'nav.dashboard', to: '/dashboard', permission: 'dashboard.view', icon: 'layout-dashboard' },
  { labelKey: 'nav.pos', to: '/pos', permission: 'pos.view', icon: 'shopping-cart' },
  { labelKey: 'nav.products', to: '/products', permission: 'products.view', icon: 'package' },
  { labelKey: 'nav.categories', to: '/categories', permission: 'categories.view', icon: 'tags' },
  { labelKey: 'nav.restock', to: '/restock', permission: 'stock.restock', icon: 'package-plus' },
  { labelKey: 'nav.stockMovements', to: '/stock-movements', permission: 'stock.movements.view', icon: 'history' },
  { labelKey: 'nav.locations', to: '/locations', permission: 'locations.view', icon: 'map-pin' },
  { labelKey: 'nav.transfers', to: '/transfers', permission: 'transfers.view', icon: 'arrow-left-right' },
  { labelKey: 'nav.salesHistory', to: '/sales-history', permission: 'sales.view', icon: 'receipt-text' },
  { labelKey: 'nav.receiptDetail', to: '/receipt-detail', permission: 'sales.receipt.view', icon: 'receipt-text' },
  { labelKey: 'nav.alerts', to: '/alerts', permission: 'alerts.view', icon: 'bell' },
  { labelKey: 'nav.reports', to: '/reports', permission: 'reports.view', icon: 'chart-column' },
  { labelKey: 'nav.exports', to: '/exports', permission: 'exports.view', icon: 'download' },
  { labelKey: 'nav.imports', to: '/imports', permission: 'imports.view', icon: 'upload' },
  { labelKey: 'nav.purchaseOrders', to: '/purchase-orders', permission: 'purchase_orders.view', icon: 'clipboard-list' },
  { labelKey: 'nav.suppliers', to: '/suppliers', permission: 'suppliers.view', icon: 'truck' },
  { labelKey: 'nav.users', to: '/users', permission: 'users.view', icon: 'users' },
  { labelKey: 'nav.roles', to: '/roles', permission: 'roles.view', icon: 'settings' },
  { labelKey: 'nav.settings', to: '/settings', permission: 'settings.view', icon: 'settings' },
]

const visibleNavItems = computed(() => navItems.filter((item) => auth.canViewMenu(item)))

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
  <div class="min-h-screen lg:grid" :class="app.sidebarCollapsed ? 'lg:grid-cols-[88px_1fr]' : 'lg:grid-cols-[292px_1fr]'">
    <aside class="hidden border-r border-slate-200 bg-white/80 backdrop-blur-xl transition-[width] lg:block">
      <div class="sticky top-0 flex h-screen flex-col">
        <div class="border-b border-slate-200 p-4">
          <div class="flex items-center justify-between gap-2">
            <div class="flex min-w-0 items-center gap-3">
              <div class="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-brand-600 text-white shadow-lg shadow-brand-600/20">
                <AppIcon name="shopping-cart" />
              </div>
              <div v-if="!app.sidebarCollapsed" class="min-w-0">
                <p class="truncate text-xs font-semibold uppercase text-brand-700">{{ app.t('app.name') }}</p>
                <h1 class="truncate text-base font-black">{{ app.t('app.subtitle') }}</h1>
              </div>
            </div>
            <button class="focus-ring hidden rounded-xl p-2 text-slate-500 hover:bg-slate-100 xl:block" :aria-label="app.sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'" @click="app.toggleSidebarCollapsed">
              <AppIcon name="panel-left" :size="18" />
            </button>
          </div>
        </div>
        <nav class="flex-1 overflow-y-auto p-3" :class="app.sidebarCollapsed ? 'px-2' : ''">
          <RouterLink
            v-for="item in visibleNavItems"
            :key="item.to"
            :to="item.to"
            class="group relative mb-1 flex min-h-11 items-center gap-3 rounded-2xl px-3 py-2 text-sm font-bold text-slate-600 transition hover:bg-brand-50 hover:text-brand-700"
            :class="[
              { 'bg-brand-600 text-white shadow-lg shadow-brand-600/20 hover:bg-brand-600 hover:text-white': isActive(item.to) },
              app.sidebarCollapsed ? 'justify-center px-2' : '',
            ]"
            :title="app.t(item.labelKey)"
          >
            <AppIcon v-if="item.icon" :name="item.icon" :size="20" />
            <span :class="app.sidebarCollapsed ? 'sr-only' : 'truncate'">{{ app.t(item.labelKey) }}</span>
            <span v-if="item.to === '/alerts' && app.alertCount" :class="app.sidebarCollapsed ? 'absolute right-1 top-1' : 'ml-auto'">
              <AppBadge>{{ app.alertCount }}</AppBadge>
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
          <span class="flex items-center justify-between gap-3">
            <AppIcon v-if="item.icon" :name="item.icon" :size="19" />
            <span>{{ app.t(item.labelKey) }}</span>
            <AppBadge v-if="item.to === '/alerts' && app.alertCount">{{ app.alertCount }}</AppBadge>
          </span>
        </RouterLink>
      </nav>
    </AppDrawer>

    <div class="min-w-0">
      <header class="sticky top-0 z-20 border-b border-slate-200 bg-white/75 backdrop-blur-xl">
        <div class="flex min-h-16 items-center justify-between gap-2 px-3 sm:gap-3 sm:px-4 lg:px-6">
          <button class="focus-ring shrink-0 rounded-xl border border-slate-200 bg-white/80 px-3 py-2 text-sm font-bold lg:hidden" @click="app.openSidebar">
            <span class="inline-flex items-center gap-2"><AppIcon name="menu" :size="18" /><span class="hidden sm:inline">{{ app.t('topbar.menu') }}</span></span>
          </button>
          <div class="hidden min-w-0 sm:block">
            <p class="truncate text-sm text-slate-500">{{ app.t('topbar.kicker') }}</p>
            <p class="truncate font-semibold">{{ app.t('topbar.workspace') }}</p>
          </div>
          <div class="flex min-w-0 items-center gap-2 sm:gap-3">
            <RouterLink to="/alerts" class="relative shrink-0 rounded-xl border border-slate-200 bg-white/80 px-3 py-2 text-sm font-bold">
              <span class="inline-flex items-center gap-2"><AppIcon name="bell" :size="17" /><span class="hidden sm:inline">{{ app.t('topbar.alerts') }}</span></span>
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
            <select v-model="app.textSize" class="focus-ring hidden min-h-10 rounded-xl border border-slate-200 bg-white/80 px-2 text-sm font-bold text-slate-600 md:block">
              <option value="sm">A-</option>
              <option value="base">A</option>
              <option value="lg">A+</option>
              <option value="xl">A++</option>
            </select>
            <button class="shrink-0 rounded-xl border border-slate-200 bg-white/80 px-3 py-2 text-sm font-bold" @click="app.toggleTheme">
              <span class="inline-flex items-center gap-2"><AppIcon :name="app.isDark ? 'sun' : 'moon'" :size="17" /><span class="hidden sm:inline">{{ app.isDark ? app.t('settings.light') : app.t('settings.dark') }}</span></span>
            </button>
            <button class="hidden items-center gap-2 rounded-xl border border-slate-200 bg-white/80 px-3 py-2 text-sm font-bold sm:flex">
              <span class="grid h-7 w-7 place-items-center rounded-full bg-brand-100 text-xs text-brand-700">{{ auth.userInitials }}</span>
              <span class="hidden sm:inline">{{ auth.user?.username }} · {{ auth.user?.role }}</span>
            </button>
            <button class="rounded-xl border border-slate-200 bg-white/80 px-3 py-2 text-sm font-bold" @click="logout">
              <span class="inline-flex items-center gap-2"><AppIcon name="log-out" :size="17" /><span class="hidden sm:inline">{{ app.t('topbar.logout') }}</span></span>
            </button>
          </div>
        </div>
      </header>
      <main class="px-4 py-7 lg:px-8">
        <RouterView />
      </main>
    </div>
  </div>
</template>
