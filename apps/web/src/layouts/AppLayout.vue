<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { apiClient } from '../api/client'
import logoUrl from '../assets/logo.png'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { AlertType, InventoryAlert, NavigationItem } from '../types/navigation'
import AppBadge from '../components/AppBadge.vue'
import AppDrawer from '../components/AppDrawer.vue'
import AppIcon from '../components/AppIcon.vue'
import ProfileDropdown from '../components/ProfileDropdown.vue'

interface NavigationGroup {
  labelKey: 'nav.group.overview' | 'nav.group.sales' | 'nav.group.catalog' | 'nav.group.inventory' | 'nav.group.admin'
  items: NavigationItem[]
}

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()
const languageOpen = ref(false)
const textSizeOpen = ref(false)
const notificationsOpen = ref(false)
const notificationsLoading = ref(false)
const notificationError = ref('')
const latestAlerts = ref<InventoryAlert[]>([])
const topbarControls = ref<HTMLElement | null>(null)

const languageOptions = [
  { value: 'th', label: 'ไทย', flag: '🇹🇭' },
  { value: 'en', label: 'English', flag: '🇺🇸' },
] as const

const textSizeOptions = [
  { value: 'sm', labelKey: 'settings.small', sample: 'Aa-' },
  { value: 'base', labelKey: 'settings.default', sample: 'Aa' },
  { value: 'lg', labelKey: 'settings.large', sample: 'Aa+' },
  { value: 'xl', labelKey: 'settings.extraLarge', sample: 'Aa++' },
] as const

const navGroups: NavigationGroup[] = [
  {
    labelKey: 'nav.group.overview',
    items: [
      { labelKey: 'nav.dashboard', to: '/dashboard', permission: 'dashboard.view', icon: 'layout-dashboard' },
      { labelKey: 'nav.reports', to: '/reports', permission: 'reports.view', icon: 'chart-column' },
    ],
  },
  {
    labelKey: 'nav.group.sales',
    items: [
      { labelKey: 'nav.pos', to: '/pos', permission: 'pos.view', icon: 'shopping-cart' },
      { labelKey: 'nav.salesHistory', to: '/sales-history', permission: 'sales.view', icon: 'purchase-order' },
    ],
  },
  {
    labelKey: 'nav.group.catalog',
    items: [
      { labelKey: 'nav.products', to: '/products', permission: 'products.view', icon: 'package' },
      { labelKey: 'nav.categories', to: '/categories', permission: 'categories.view', icon: 'tags' },
      { labelKey: 'nav.procurement', to: '/procurement', permissions: ['purchase_orders.view', 'suppliers.view'], icon: 'clipboard-list' },
    ],
  },
  {
    labelKey: 'nav.group.inventory',
    items: [
      { labelKey: 'nav.stockOperations', to: '/stock-operations', permissions: ['stock.restock', 'stock.adjust', 'stock.movements.view'], icon: 'package-plus' },
      { labelKey: 'nav.inventoryManagement', to: '/inventory-management', permissions: ['locations.view', 'transfers.view'], icon: 'map-pin' },
    ],
  },
  {
    labelKey: 'nav.group.admin',
    items: [
      { labelKey: 'nav.users', to: '/users', permission: 'users.view', icon: 'users' },
      { labelKey: 'nav.roles', to: '/roles', permission: 'roles.view', icon: 'role' },
      { labelKey: 'nav.settings', to: '/settings', permission: 'settings.view', icon: 'settings' },
    ],
  },
]

const visibleNavGroups = computed(() => navGroups
  .map((group) => ({ ...group, items: group.items.filter((item) => auth.canViewMenu(item)) }))
  .filter((group) => group.items.length > 0))
const currentLanguage = computed(() => languageOptions.find((option) => option.value === app.language) ?? languageOptions[0])
const currentTextSize = computed(() => textSizeOptions.find((option) => option.value === app.textSize) ?? textSizeOptions[1])
const activeNavPath = computed(() => route.name === 'receipt-detail' || route.name === 'sale-receipt' ? '/sales-history' : route.path)
const canViewAlerts = computed(() => auth.hasPermission('alerts.view'))
const alertBadgeLabel = computed(() => app.alertCount > 99 ? '99+' : String(app.alertCount))
const dropdownAlerts = computed(() => latestAlerts.value.slice(0, 5))

function isActive(to: string) {
  return activeNavPath.value === to
}

function selectLanguage(value: 'th' | 'en') {
  app.setLanguage(value)
  languageOpen.value = false
}

function selectTextSize(value: 'sm' | 'base' | 'lg' | 'xl') {
  app.setTextSize(value)
  textSizeOpen.value = false
}

function alertTypeLabel(type: AlertType) {
  const labels: Record<AlertType, 'alerts.type.lowStock' | 'alerts.type.outOfStock' | 'alerts.type.reorder'> = {
    LOW_STOCK: 'alerts.type.lowStock',
    OUT_OF_STOCK: 'alerts.type.outOfStock',
    REORDER_POINT: 'alerts.type.reorder',
  }
  return app.t(labels[type])
}

function formatMessage(template: string, params: Record<string, string | number>) {
  let text = template
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function alertMessage(alert: InventoryAlert) {
  const keys: Record<AlertType, 'alerts.message.lowStock' | 'alerts.message.outOfStock' | 'alerts.message.reorder'> = {
    LOW_STOCK: 'alerts.message.lowStock',
    OUT_OF_STOCK: 'alerts.message.outOfStock',
    REORDER_POINT: 'alerts.message.reorder',
  }
  return formatMessage(app.t(keys[alert.type]), { product: alert.product_name, location: alert.location_name })
}

function alertIcon(type: AlertType) {
  return {
    LOW_STOCK: 'triangle-alert',
    OUT_OF_STOCK: 'circle-x',
    REORDER_POINT: 'clipboard-list',
  }[type] as 'triangle-alert' | 'circle-x' | 'clipboard-list'
}

function alertIconClass(type: AlertType) {
  return {
    LOW_STOCK: 'bg-amber-100 text-amber-800 dark:bg-amber-500/20 dark:text-amber-100',
    OUT_OF_STOCK: 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-100',
    REORDER_POINT: 'bg-blue-100 text-blue-700 dark:bg-sky-500/20 dark:text-sky-100',
  }[type]
}

function notificationTime(value: string) {
  return new Date(value).toLocaleString(app.language === 'th' ? 'th-TH' : 'en-US', { dateStyle: 'short', timeStyle: 'short' })
}

async function loadLatestAlerts() {
  if (!canViewAlerts.value) return
  notificationsLoading.value = true
  notificationError.value = ''
  try {
    latestAlerts.value = await apiClient<InventoryAlert[]>('/v1/alerts?limit=5')
    await app.loadAlertCount()
  } catch (err) {
    notificationError.value = err instanceof Error ? err.message : app.t('alerts.loadFailed')
  } finally {
    notificationsLoading.value = false
  }
}

async function toggleNotifications() {
  notificationsOpen.value = !notificationsOpen.value
  languageOpen.value = false
  textSizeOpen.value = false
  if (notificationsOpen.value) await loadLatestAlerts()
}

function closeNotifications() {
  notificationsOpen.value = false
}

function closeTopbarDropdowns(event: MouseEvent) {
  if (topbarControls.value?.contains(event.target as Node)) return
  languageOpen.value = false
  textSizeOpen.value = false
  notificationsOpen.value = false
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
  notificationsOpen.value = false
  languageOpen.value = false
  textSizeOpen.value = false
})
</script>

<template>
  <div class="min-h-screen lg:grid" :class="app.sidebarCollapsed ? 'lg:grid-cols-[88px_1fr]' : 'lg:grid-cols-[292px_1fr]'">
    <aside class="hidden bg-white/80 shadow-xl shadow-teal-950/5 backdrop-blur-xl transition-[width] dark:bg-slate-950/80 dark:shadow-black/25 lg:block">
      <div class="sticky top-0 flex h-screen flex-col">
        <div class="p-4">
          <div v-if="!app.sidebarCollapsed" class="flex min-h-12 items-center justify-between gap-3">
            <div class="flex min-w-0 flex-1 items-center gap-3">
              <img class="h-11 w-11 shrink-0 object-contain" :src="logoUrl" :alt="app.t('app.name')" />
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
          <section v-for="group in visibleNavGroups" :key="group.labelKey" class="mb-4 last:mb-0">
            <p v-if="!app.sidebarCollapsed" class="mb-2 px-3 text-[11px] font-black uppercase tracking-wide text-slate-400 dark:text-slate-500">{{ app.t(group.labelKey) }}</p>
            <div v-else class="mx-auto mb-2 h-px w-8 bg-slate-200 dark:bg-slate-800" />
            <RouterLink v-for="item in group.items" :key="item.to" :to="item.to"
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
          </section>
        </nav>
      </div>
    </aside>

    <AppDrawer :open="app.sidebarOpen" @close="app.closeSidebar">
      <div class="mb-4 flex items-center gap-3 rounded-xl bg-brand-50 p-3 shadow-sm dark:bg-slate-900">
          <AppIcon name="shopping-cart" />
        </div>
        <div class="min-w-0">
          <p class="truncate text-xs font-semibold uppercase text-brand-700">{{ app.t('app.name') }}</p>
          <h1 class="truncate text-base font-black">{{ app.t('app.subtitle') }}</h1>
        </div>
      </div>
      <nav class="grid gap-3">
        <section v-for="group in visibleNavGroups" :key="group.labelKey" class="grid gap-1">
          <p class="px-3 text-[11px] font-black uppercase tracking-wide text-slate-400 dark:text-slate-500">{{ app.t(group.labelKey) }}</p>
          <RouterLink
            v-for="item in group.items"
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
        </section>
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
                :aria-label="app.t('settings.adjustTextSize')"
                @click="textSizeOpen = !textSizeOpen; languageOpen = false"
              >
                <AppIcon name="text-size" :size="20" />
                <!-- <span class="sr-only">{{ currentTextSize.sample }}</span> -->
              </button>
              <div v-if="textSizeOpen" class="dark:bg-slate-900/80 absolute right-0 z-40 mt-2 w-48 rounded-2xl p-2 shadow-xl">
                <p class="px-3 py-2 text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('settings.textSize') }}</p>
                <button
                  v-for="option in textSizeOptions"
                  :key="option.value"
                  class="flex min-h-11 w-full items-center justify-start gap-3 rounded-xl px-3 text-left text-sm font-bold transition hover:bg-brand-50 dark:hover:bg-slate-800"
                  :class="app.textSize === option.value ? 'bg-brand-600 text-white hover:bg-brand-600 dark:bg-teal-300 dark:!text-slate-950 dark:hover:bg-emerald-400' : 'text-slate-700 dark:text-slate-200'"
                  @click="selectTextSize(option.value)"
                >
                  <span>{{ app.t(option.labelKey) }}</span>
                  <!-- <span class="font-black">{{ option.sample }}</span> -->
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
              <div v-if="languageOpen" class="dark:bg-slate-900/80 absolute right-0 z-40 mt-2 w-44 rounded-2xl p-2 shadow-xl">
                <button
                  v-for="option in languageOptions"
                  :key="option.value"
                  class="flex min-h-11 w-full items-center gap-3 rounded-xl px-3 text-left text-sm font-bold transition hover:bg-brand-50 dark:hover:bg-slate-800"
                  :class="app.language === option.value ? 'bg-brand-600 text-white hover:bg-brand-600 dark:bg-teal-300 dark:!text-slate-950 dark:hover:bg-emerald-400' : 'text-slate-700 dark:text-slate-200'"
                  @click="selectLanguage(option.value)"
                >
                  <span class="text-base leading-none">{{ option.flag }}</span>
                  <span>{{ option.label }}</span>
                </button>
              </div>
            </div>
            <div v-if="canViewAlerts" class="relative">
              <button
                class="relative grid h-10 w-10 place-items-center rounded-xl text-slate-700 transition hover:bg-brand-50 focus:outline-none focus:ring-2 focus:ring-brand-400/50 dark:bg-slate-900/80 dark:text-slate-200 dark:hover:bg-teal-400/10"
                :aria-label="app.t('topbar.openNotifications')"
                :aria-expanded="notificationsOpen"
                aria-haspopup="menu"
                @click="toggleNotifications"
              >
                <AppIcon name="bell" :size="20" />
                <span v-if="app.alertCount > 0" class="absolute -right-1 -top-1 grid h-5 min-w-5 place-items-center rounded-full bg-red-600 px-1.5 text-xs font-black leading-none text-white shadow-sm shadow-red-900/30">
                  {{ alertBadgeLabel }}
                </span>
              </button>
              <div v-if="notificationsOpen" class="absolute right-0 z-40 mt-2 w-[calc(100vw-2rem)] max-w-[360px] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl shadow-slate-950/15 dark:border-slate-700 dark:bg-slate-900 dark:shadow-black/30">
                <div class="flex items-start justify-between gap-3 border-b border-slate-100 p-4 dark:border-slate-800">
                  <div>
                    <p class="text-sm font-black">{{ app.t('topbar.notifications') }}</p>
                    <p class="mt-0.5 text-xs text-slate-500 dark:text-slate-400">{{ app.t('topbar.unreadCount').replace('{count}', String(app.alertCount)) }}</p>
                  </div>
                  <button class="grid h-8 w-8 place-items-center rounded-xl text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800" :aria-label="app.t('topbar.close')" @click="closeNotifications">
                    <AppIcon name="x" :size="16" />
                  </button>
                </div>
                <div class="max-h-[360px] overflow-auto p-2">
                  <div v-if="notificationsLoading" class="p-4 text-sm font-semibold text-slate-500 dark:text-slate-400">{{ app.t('alerts.loading') }}</div>
                  <div v-else-if="notificationError" class="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ notificationError }}</div>
                  <div v-else-if="dropdownAlerts.length === 0" class="p-5 text-center">
                    <div class="mx-auto grid h-11 w-11 place-items-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-emerald-500/15 dark:text-emerald-100">
                      <AppIcon name="bell" />
                    </div>
                    <p class="mt-3 text-sm font-black">{{ app.t('topbar.noNotifications') }}</p>
                  </div>
                  <template v-else>
                    <RouterLink
                      v-for="alert in dropdownAlerts"
                      :key="alert.id"
                      to="/alerts"
                      class="flex gap-3 rounded-xl p-3 text-left transition hover:bg-slate-50 dark:hover:bg-slate-800/80"
                      @click="closeNotifications"
                    >
                      <span class="grid h-10 w-10 shrink-0 place-items-center rounded-xl" :class="alertIconClass(alert.type)">
                        <AppIcon :name="alertIcon(alert.type)" :size="18" />
                      </span>
                      <span class="min-w-0 flex-1">
                        <span class="flex items-start justify-between gap-2">
                          <span class="truncate text-sm font-black">{{ alertTypeLabel(alert.type) }}</span>
                          <span v-if="!alert.read_at" class="mt-1 h-2 w-2 shrink-0 rounded-full bg-red-500" />
                        </span>
                        <span class="mt-1 block truncate text-sm text-slate-600 dark:text-slate-300">{{ alertMessage(alert) }}</span>
                        <span class="mt-1 block truncate text-xs text-slate-500 dark:text-slate-400">{{ alert.location_name }} · {{ notificationTime(alert.created_at) }}</span>
                      </span>
                    </RouterLink>
                  </template>
                </div>
                <RouterLink to="/alerts" class="block border-t border-slate-100 px-4 py-3 text-center text-sm font-black text-brand-700 hover:bg-brand-50 dark:border-slate-800 dark:text-emerald-300 dark:hover:bg-slate-800" @click="closeNotifications">
                  {{ app.t('topbar.viewAllNotifications') }}
                </RouterLink>
              </div>
            </div>
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
