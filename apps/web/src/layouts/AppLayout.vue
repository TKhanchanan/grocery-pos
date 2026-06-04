<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { NavigationItem } from '../types/navigation'
import AppBadge from '../components/AppBadge.vue'
import AppDrawer from '../components/AppDrawer.vue'
import AppIcon from '../components/AppIcon.vue'
import ProfileDropdown from '../components/ProfileDropdown.vue'

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()
const languageOpen = ref(false)
const textSizeOpen = ref(false)
const topbarControls = ref<HTMLElement | null>(null)

const languageOptions = [
  { value: 'th', label: 'ไทย', flag: '🇹🇭' },
  { value: 'en', label: 'English', flag: '🇺🇸' },
] as const

const textSizeOptions = [
  { value: 'sm', label: 'Small', sample: 'A-' },
  { value: 'base', label: 'Default', sample: 'A' },
  { value: 'lg', label: 'Large', sample: 'A+' },
  { value: 'xl', label: 'Extra large', sample: 'A++' },
] as const

const navItems: NavigationItem[] = [
  { labelKey: 'nav.dashboard', to: '/dashboard', permission: 'dashboard.view', icon: 'layout-dashboard' },
  { labelKey: 'nav.pos', to: '/pos', permission: 'pos.view', icon: 'shopping-cart' },
  { labelKey: 'nav.products', to: '/products', permission: 'products.view', icon: 'package' },
  { labelKey: 'nav.categories', to: '/categories', permission: 'categories.view', icon: 'tags' },
  { labelKey: 'nav.restock', to: '/restock', permission: 'stock.restock', icon: 'package-plus' },
  { labelKey: 'nav.stockMovements', to: '/stock-movements', permission: 'stock.movements.view', icon: 'history' },
  { labelKey: 'nav.locations', to: '/locations', permission: 'locations.view', icon: 'map-pin' },
  { labelKey: 'nav.transfers', to: '/transfers', permission: 'transfers.view', icon: 'arrow-left-right' },
  { labelKey: 'nav.salesHistory', to: '/sales-history', permission: 'sales.view', icon: 'purchase-order' },
  { labelKey: 'nav.receiptDetail', to: '/receipt-detail', permission: 'sales.receipt.view', icon: 'receipt-text' },
  { labelKey: 'nav.alerts', to: '/alerts', permission: 'alerts.view', icon: 'bell' },
  { labelKey: 'nav.reports', to: '/reports', permission: 'reports.view', icon: 'chart-column' },
  { labelKey: 'nav.exports', to: '/exports', permission: 'exports.view', icon: 'download' },
  { labelKey: 'nav.imports', to: '/imports', permission: 'imports.view', icon: 'upload' },
  { labelKey: 'nav.purchaseOrders', to: '/purchase-orders', permission: 'purchase_orders.view', icon: 'clipboard-list' },
  { labelKey: 'nav.suppliers', to: '/suppliers', permission: 'suppliers.view', icon: 'truck' },
  { labelKey: 'nav.users', to: '/users', permission: 'users.view', icon: 'users' },
  { labelKey: 'nav.roles', to: '/roles', permission: 'roles.view', icon: 'role' },
  { labelKey: 'nav.settings', to: '/settings', permission: 'settings.view', icon: 'settings' },
]

const visibleNavItems = computed(() => navItems.filter((item) => auth.canViewMenu(item)))
const currentLanguage = computed(() => languageOptions.find((option) => option.value === app.language) ?? languageOptions[0])
const currentTextSize = computed(() => textSizeOptions.find((option) => option.value === app.textSize) ?? textSizeOptions[1])

function isActive(to: string) {
  return route.path === to
}

function selectLanguage(value: 'th' | 'en') {
  app.setLanguage(value)
  languageOpen.value = false
}

function selectTextSize(value: 'sm' | 'base' | 'lg' | 'xl') {
  app.setTextSize(value)
  textSizeOpen.value = false
}

function closeTopbarDropdowns(event: MouseEvent) {
  if (topbarControls.value?.contains(event.target as Node)) return
  languageOpen.value = false
  textSizeOpen.value = false
}

onMounted(() => {
  app.loadAlertCount()
  document.addEventListener('mousedown', closeTopbarDropdowns)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', closeTopbarDropdowns)
})

watch(() => route.path, () => {
  app.loadAlertCount()
})
</script>

<template>
  <div class="min-h-screen lg:grid" :class="app.sidebarCollapsed ? 'lg:grid-cols-[88px_1fr]' : 'lg:grid-cols-[292px_1fr]'">
    <aside class="hidden bg-white/80 shadow-xl shadow-teal-950/5 backdrop-blur-xl transition-[width] dark:bg-slate-950/80 dark:shadow-black/25 lg:block">
      <div class="sticky top-0 flex h-screen flex-col">
        <div class="p-4">
          <div v-if="!app.sidebarCollapsed" class="flex min-h-12 items-center justify-between gap-3">
            <div class="flex min-w-0 flex-1 items-center gap-3">
              <div class="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-brand-600 text-white shadow-lg shadow-brand-600/20">
                <AppIcon name="shopping-cart" />
              </div>
              <div class="min-w-0">
                <p class="truncate text-xs font-semibold uppercase text-brand-700">{{ app.t('app.name') }}</p>
                <h1 class="truncate text-base font-black">{{ app.t('app.subtitle') }}</h1>
              </div>
            </div>
            <button class="grid h-10 w-10 shrink-0 place-items-center rounded-xl text-slate-500 hover:bg-brand-50 dark:bg-slate-900/80 dark:text-slate-300 dark:hover:bg-slate-800" aria-label="ย่อเมนู" @click="app.toggleSidebarCollapsed">
              <AppIcon name="panel-left" :size="18" />
            </button>
          </div>
          <div v-else class="grid min-h-12 place-items-center">
            <button class="grid h-10 w-10 place-items-center rounded-xl text-slate-500 hover:bg-brand-50 dark:bg-slate-900/80 dark:text-slate-300 dark:hover:bg-slate-800" aria-label="ขยายเมนู" @click="app.toggleSidebarCollapsed">
              <AppIcon name="panel-left" :size="18" />
            </button>
          </div>
        </div>
        <nav class="flex-1 overflow-y-auto p-3" :class="app.sidebarCollapsed ? 'px-2' : ''">
          <RouterLink v-for="item in visibleNavItems" :key="item.to" :to="item.to"
            class="group relative mb-1 flex min-h-11 items-center gap-3 rounded-xl px-3 py-2 text-sm font-bold text-slate-600 transition hover:bg-brand-50 hover:text-brand-700 dark:text-slate-300 dark:hover:bg-teal-400/10 dark:hover:text-teal-100"
            :class="[
              isActive(item.to)
                ? 'bg-brand-600 text-white shadow-lg shadow-brand-600/20 hover:!bg-brand-600 hover:!text-white dark:bg-teal-300 dark:!text-slate-950 dark:hover:!bg-teal-300 dark:hover:!text-slate-950'
                : '',
              app.sidebarCollapsed ? 'justify-center px-2' : '',
            ]" :title="app.t(item.labelKey)" :aria-label="app.sidebarCollapsed ? app.t(item.labelKey) : undefined">
            <AppIcon v-if="item.icon" :name="item.icon" :size="20" />
            <span :class="app.sidebarCollapsed ? 'sr-only' : 'truncate'">{{ app.t(item.labelKey) }}</span>
            <span v-if="item.to === '/alerts' && app.alertCount" :class="app.sidebarCollapsed ? 'absolute right-1 top-1' : 'ml-auto'">
              <AppBadge>{{ app.alertCount }}</AppBadge>
            </span>
          </RouterLink>
        </nav>
      </div>
    </aside>

    <AppDrawer :open="app.sidebarOpen" @close="app.closeSidebar">
      <div class="mb-4 flex items-center gap-3 rounded-xl bg-brand-50 p-3 shadow-sm dark:bg-slate-900">
        <div class="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-brand-600 text-white shadow-lg shadow-brand-600/20">
          <AppIcon name="shopping-cart" />
        </div>
        <div class="min-w-0">
          <p class="truncate text-xs font-semibold uppercase text-brand-700">{{ app.t('app.name') }}</p>
          <h1 class="truncate text-base font-black">{{ app.t('app.subtitle') }}</h1>
        </div>
      </div>
      <nav class="grid gap-1">
        <RouterLink
          v-for="item in visibleNavItems"
          :key="item.to"
          :to="item.to"
          class="rounded-xl px-3 py-2 text-sm font-medium text-slate-600 hover:bg-brand-50 dark:text-slate-300 dark:hover:bg-teal-400/10"
          :class="{ 'bg-brand-600 text-white dark:bg-teal-300 dark:!text-slate-950': isActive(item.to) }"
          @click="app.closeSidebar"
        >
          <span class="flex items-center justify-start gap-3">
            <AppIcon v-if="item.icon" :name="item.icon" :size="19" />
            <span>{{ app.t(item.labelKey) }}</span>
            <AppBadge v-if="item.to === '/alerts' && app.alertCount">{{ app.alertCount }}</AppBadge>
          </span>
        </RouterLink>
      </nav>
    </AppDrawer>

    <div class="min-w-0">
      <header class="sticky top-0 z-20 bg-white/75 shadow-sm backdrop-blur-xl dark:bg-slate-950/75">
        <div class="flex min-h-16 items-center justify-between gap-2 px-3 sm:gap-3 sm:px-4 lg:px-6">
          <button class="shrink-0 rounded-xl px-3 py-2 text-sm font-bold dark:bg-slate-900/80 lg:hidden" @click="app.openSidebar">
            <AppIcon name="menu" :size="20" />
          </button>
          <div class="hidden min-w-0 sm:block"></div>
          <div ref="topbarControls" class="flex min-w-0 items-center gap-2 sm:gap-2">
            <div class="relative hidden md:block">
              <button
                class="inline-flex min-h-10 min-w-11 items-center justify-center rounded-xl px-3 text-sm font-black text-slate-700 transition hover:bg-brand-50 dark:bg-slate-900/80 dark:text-slate-200 dark:hover:bg-teal-400/10"
                aria-label="ปรับขนาดตัวอักษร"
                @click="textSizeOpen = !textSizeOpen; languageOpen = false"
              >
                {{ currentTextSize.sample }}
              </button>
              <div v-if="textSizeOpen" class="premium-surface absolute right-0 z-40 mt-2 w-44 rounded-2xl border p-2 shadow-xl">
                <button
                  v-for="option in textSizeOptions"
                  :key="option.value"
                  class="flex min-h-11 w-full items-center justify-start gap-3 rounded-xl px-3 text-left text-sm font-bold transition hover:bg-brand-50 dark:hover:bg-slate-800"
                  :class="app.textSize === option.value ? 'bg-brand-600 text-white hover:bg-brand-600 dark:bg-emerald-500 dark:!text-slate-950 dark:hover:bg-emerald-400' : 'text-slate-700 dark:text-slate-200'"
                  @click="selectTextSize(option.value)"
                >
                  <span>{{ option.label }}</span>
                  <span class="font-black">{{ option.sample }}</span>
                </button>
              </div>
            </div>
            <div class="relative hidden sm:block">
              <button
                class="inline-flex min-h-10 items-center gap-2 rounded-xl px-3 text-sm font-bold text-slate-700 transition hover:bg-brand-50 dark:bg-slate-900/80 dark:text-slate-200 dark:hover:bg-teal-400/10"
                aria-label="เลือกภาษา"
                @click="languageOpen = !languageOpen; textSizeOpen = false"
              >
                <span class="text-base leading-none uppercase">{{ currentLanguage.value }}</span>
              </button>
              <div v-if="languageOpen" class="premium-surface absolute right-0 z-40 mt-2 w-44 rounded-2xl border p-2 shadow-xl">
                <button
                  v-for="option in languageOptions"
                  :key="option.value"
                  class="flex min-h-11 w-full items-center gap-3 rounded-xl px-3 text-left text-sm font-bold transition hover:bg-brand-50 dark:hover:bg-slate-800"
                  :class="app.language === option.value ? 'bg-brand-600 text-white hover:bg-brand-600 dark:bg-emerald-500 dark:!text-slate-950 dark:hover:bg-emerald-400' : 'text-slate-700 dark:text-slate-200'"
                  @click="selectLanguage(option.value)"
                >
                  <span class="text-base leading-none">{{ option.flag }}</span>
                  <span>{{ option.label }}</span>
                </button>
              </div>
            </div>
            <RouterLink to="/alerts" class="relative grid h-10 w-10 place-items-center rounded-xl text-slate-700 transition hover:bg-brand-50 dark:bg-slate-900/80 dark:text-slate-200 dark:hover:bg-teal-400/10" aria-label="Alerts" @click="languageOpen = false; textSizeOpen = false">
              <AppIcon name="bell" :size="20" />
              <AppBadge v-if="app.alertCount" class="absolute -right-2 -top-2">{{ app.alertCount }}</AppBadge>
            </RouterLink>
            <button class="grid h-10 w-10 place-items-center rounded-xl text-slate-700 transition hover:bg-brand-50 dark:bg-slate-900/80 dark:text-slate-200 dark:hover:bg-teal-400/10" aria-label="Toggle theme" @click="languageOpen = false; textSizeOpen = false; app.toggleTheme()">
              <AppIcon :name="app.isDark ? 'sun' : 'moon'" :size="20" />
            </button>
            <ProfileDropdown />
          </div>
        </div>
      </header>
      <main class="px-4 py-7 lg:px-8">
        <RouterView />
      </main>
    </div>
  </div>
</template>
