<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { apiClient, patchJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppIcon from '../components/AppIcon.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { AlertType, InventoryAlert, Location } from '../types/navigation'
import { formatAppDateTime } from '../utils/date'

type AlertFilter = 'all' | 'unread' | AlertType

const app = useAppStore()
const auth = useAuthStore()
const alerts = ref<InventoryAlert[]>([])
const locations = ref<Location[]>([])
const loading = ref(false)
const error = ref('')

const filters = reactive({
  mode: 'all' as AlertFilter,
  location_id: '',
})

const canMarkRead = computed(() => auth.hasPermission('alerts.mark_read'))
const canCreatePO = computed(() => auth.hasPermission('alerts.create_po') || auth.hasPermission('purchase_orders.create_from_alert'))
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')
const allCount = computed(() => alerts.value.length)
const unreadCount = computed(() => alerts.value.filter((alert) => !alert.read_at).length)
const lowStockCount = computed(() => alerts.value.filter((alert) => alert.type === 'LOW_STOCK').length)
const outOfStockCount = computed(() => alerts.value.filter((alert) => alert.type === 'OUT_OF_STOCK').length)
const reorderCount = computed(() => alerts.value.filter((alert) => alert.type === 'REORDER_POINT').length)

const filterChips = computed<Array<{ value: AlertFilter; label: string; count: number }>>(() => [
  { value: 'all', label: app.t('alerts.filter.all'), count: allCount.value },
  { value: 'unread', label: app.t('alerts.unread'), count: unreadCount.value },
  { value: 'LOW_STOCK', label: alertTypeLabel('LOW_STOCK'), count: lowStockCount.value },
  { value: 'OUT_OF_STOCK', label: alertTypeLabel('OUT_OF_STOCK'), count: outOfStockCount.value },
  { value: 'REORDER_POINT', label: alertTypeLabel('REORDER_POINT'), count: reorderCount.value },
])

const visibleAlerts = computed(() => alerts.value.filter((alert) => {
  if (filters.mode === 'unread') return !alert.read_at
  if (filters.mode === 'LOW_STOCK' || filters.mode === 'OUT_OF_STOCK' || filters.mode === 'REORDER_POINT') return alert.type === filters.mode
  return true
}))

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function alertTypeLabel(type: AlertType) {
  const labels: Record<AlertType, TranslationKey> = {
    LOW_STOCK: 'alerts.type.lowStock',
    OUT_OF_STOCK: 'alerts.type.outOfStock',
    REORDER_POINT: 'alerts.type.reorder',
  }
  return app.t(labels[type])
}

function alertMessage(alert: InventoryAlert) {
  const keys: Record<AlertType, TranslationKey> = {
    LOW_STOCK: 'alerts.message.lowStock',
    OUT_OF_STOCK: 'alerts.message.outOfStock',
    REORDER_POINT: 'alerts.message.reorder',
  }
  return t(keys[alert.type], { product: alert.product_name, location: alert.location_name })
}

function alertTone(type: AlertType) {
  return {
    LOW_STOCK: 'warning',
    OUT_OF_STOCK: 'danger',
    REORDER_POINT: 'info',
  }[type] as 'warning' | 'danger' | 'info'
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

function buildQuery() {
  const params = new URLSearchParams()
  if (filters.location_id) params.set('location_id', filters.location_id)
  return params.toString()
}

function relativeDate(value: string) {
  const date = new Date(value)
  const today = new Date()
  const yesterday = new Date()
  yesterday.setDate(today.getDate() - 1)
  const day = date.toDateString()
  if (day === today.toDateString()) return `${app.t('alerts.today')} ${date.toLocaleTimeString(locale.value, { hour: '2-digit', minute: '2-digit' })}`
  if (day === yesterday.toDateString()) return `${app.t('alerts.yesterday')} ${date.toLocaleTimeString(locale.value, { hour: '2-digit', minute: '2-digit' })}`
  return formatAppDateTime(value, app.language)
}

async function loadAlerts() {
  loading.value = true
  error.value = ''
  try {
    const query = buildQuery()
    alerts.value = await apiClient<InventoryAlert[]>(`/v1/alerts${query ? `?${query}` : ''}`)
    await app.loadAlertCount()
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('alerts.loadFailed')
  } finally {
    loading.value = false
  }
}

async function loadLocations() {
  locations.value = await apiClient<Location[]>('/v1/locations').catch(() => [])
}

async function markRead(alert: InventoryAlert) {
  if (!canMarkRead.value) return
  try {
    await patchJSON<InventoryAlert>(`/v1/alerts/${alert.id}/read`, {})
    await loadAlerts()
  } catch (err) {
    app.pushToast({ type: 'error', message: app.t('alerts.markReadFailed'), description: err instanceof Error ? err.message : '' })
  }
}

async function markAllRead() {
  if (!canMarkRead.value) return
  try {
    await patchJSON<InventoryAlert[]>('/v1/alerts/read-all', {})
    await loadAlerts()
  } catch (err) {
    app.pushToast({ type: 'error', message: app.t('alerts.markReadFailed'), description: err instanceof Error ? err.message : '' })
  }
}

onMounted(async () => {
  await Promise.all([loadLocations(), loadAlerts()])
})
</script>

<template>
  <section>
    <PageHeader :title="app.t('alerts.title')" :eyebrow="app.t('alerts.eyebrow')" :description="app.t('alerts.description')" icon="bell">
      <div class="flex flex-wrap gap-2">
        <!-- <AppButton variant="secondary" icon="history" :loading="loading" @click="loadAlerts">{{ app.t('alerts.tryAgain') }}</AppButton> -->
        <AppButton v-if="canMarkRead" :disabled="unreadCount === 0" icon="check-circle" @click="markAllRead">{{ app.t('alerts.markAllRead') }}</AppButton>
      </div>
    </PageHeader>

    <div class="mb-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
      <StatCard :label="app.t('alerts.total')" :value="allCount" :helper="app.t('alerts.totalHelper')" icon="bell" />
      <StatCard :label="app.t('alerts.unread')" :value="unreadCount" :helper="app.t('alerts.unreadHelper')" icon="sparkles" tone="warning" />
      <StatCard :label="alertTypeLabel('LOW_STOCK')" :value="lowStockCount" :helper="app.t('alerts.lowStockHelper')" icon="triangle-alert" tone="warning" />
      <StatCard :label="alertTypeLabel('OUT_OF_STOCK')" :value="outOfStockCount" :helper="app.t('alerts.outOfStockHelper')" icon="circle-x" tone="danger" />
      <StatCard :label="alertTypeLabel('REORDER_POINT')" :value="reorderCount" :helper="app.t('alerts.reorderHelper')" icon="clipboard-list" tone="info" />
    </div>

    <AppCard class="mb-4 dark:bg-slate-900/80">
      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_260px] lg:items-end">
        <div>
          <p class="text-sm font-black text-slate-700 dark:text-slate-200">{{ app.t('alerts.status') }}</p>
          <div class="mt-2 flex gap-2 overflow-x-auto pb-1">
            <button
              v-for="chip in filterChips"
              :key="chip.value"
              class="inline-flex min-h-10 shrink-0 items-center gap-2 rounded-full px-4 text-sm font-black transition"
              :class="filters.mode === chip.value ? 'bg-brand-600 text-white dark:bg-teal-300 dark:text-slate-950' : 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700'"
              @click="filters.mode = chip.value"
            >
              <span>{{ chip.label }}</span>
              <span class="rounded-full bg-white/20 px-2 py-0.5 text-xs">{{ chip.count }}</span>
            </button>
          </div>
        </div>
        <AppSelect v-model="filters.location_id" :label="app.t('alerts.location')" @update:model-value="loadAlerts">
          <option value="">{{ app.t('alerts.allLocations') }}</option>
          <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
        </AppSelect>
      </div>
    </AppCard>

    <div v-if="error" class="mb-4 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <span>{{ error }}</span>
        <!-- <AppButton variant="secondary" @click="loadAlerts">{{ app.t('alerts.tryAgain') }}</AppButton> -->
      </div>
    </div>
    <AppLoadingState v-if="loading" class="mb-4" :label="app.t('alerts.loading')" />
    <AppEmptyState v-else-if="visibleAlerts.length === 0" :title="app.t('alerts.empty')" :description="app.t('alerts.emptyDescription')" icon="bell" />

    <div v-else class="grid gap-3">
      <article v-for="alert in visibleAlerts" :key="alert.id" class="rounded-2xl border border-slate-200 bg-white/80 p-4 shadow-sm transition hover:border-brand-200 hover:shadow-md dark:border-slate-700 dark:bg-slate-900/80 dark:hover:border-teal-400/40">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="flex min-w-0 gap-4">
            <div class="grid h-12 w-12 shrink-0 place-items-center rounded-2xl" :class="alertIconClass(alert.type)">
              <AppIcon :name="alertIcon(alert.type)" :size="22" />
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <AppBadge :tone="alertTone(alert.type)">{{ alertTypeLabel(alert.type) }}</AppBadge>
                <AppBadge v-if="!alert.read_at" tone="brand">{{ app.t('alerts.unread') }}</AppBadge>
                <AppBadge v-else tone="neutral">{{ app.t('alerts.read') }}</AppBadge>
              </div>
              <h2 class="mt-2 text-lg font-black">{{ alert.product_name }}</h2>
              <p class="mt-1 text-sm text-slate-600 dark:text-slate-300">{{ alertMessage(alert) }}</p>
              <div class="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-xs font-semibold text-slate-500 dark:text-slate-400">
                <span>{{ alert.sku }}</span>
                <span>{{ alert.location_name }}</span>
                <span>{{ app.t('alerts.date') }}: {{ relativeDate(alert.created_at) }}</span>
              </div>
            </div>
          </div>
          <div class="flex flex-wrap gap-2 lg:justify-end">
            <RouterLink :to="alert.links.product" class="focus-ring inline-flex min-h-10 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-3 text-sm font-black text-slate-700 transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200 dark:hover:bg-slate-800">
              <AppIcon name="package" :size="16" />{{ app.t('alerts.goToProduct') }}
            </RouterLink>
            <RouterLink :to="alert.links.restock" class="focus-ring inline-flex min-h-10 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-3 text-sm font-black text-slate-700 transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200 dark:hover:bg-slate-800">
              <AppIcon name="package-plus" :size="16" />{{ app.t('alerts.restock') }}
            </RouterLink>
            <RouterLink v-if="canCreatePO" :to="alert.links.purchase_order" class="focus-ring inline-flex min-h-10 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-3 text-sm font-black text-slate-700 transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200 dark:hover:bg-slate-800">
              <AppIcon name="clipboard-list" :size="16" />{{ app.t('alerts.createPO') }}
            </RouterLink>
            <AppButton v-if="canMarkRead && !alert.read_at" icon="check-circle" @click="markRead(alert)">{{ app.t('alerts.markRead') }}</AppButton>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>
