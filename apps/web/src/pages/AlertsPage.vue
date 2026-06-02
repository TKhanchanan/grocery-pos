<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { apiClient, patchJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAppStore } from '../stores/app'
import type { AlertType, InventoryAlert, Location } from '../types/navigation'

const app = useAppStore()
const alerts = ref<InventoryAlert[]>([])
const locations = ref<Location[]>([])
const loading = ref(false)
const error = ref('')

const filters = reactive({
  unread: 'true',
  type: '',
  location_id: '',
})

const unreadCount = computed(() => alerts.value.filter((alert) => !alert.read_at).length)

function typeClass(type: AlertType) {
  return {
    LOW_STOCK: 'bg-amber-100 text-amber-800',
    OUT_OF_STOCK: 'bg-red-100 text-red-700',
    REORDER_POINT: 'bg-blue-100 text-blue-700',
  }[type]
}

function buildQuery() {
  const params = new URLSearchParams()
  if (filters.unread) params.set('unread', filters.unread)
  if (filters.type) params.set('type', filters.type)
  if (filters.location_id) params.set('location_id', filters.location_id)
  return params.toString()
}

async function loadAlerts() {
  loading.value = true
  error.value = ''
  try {
    const query = buildQuery()
    alerts.value = await apiClient<InventoryAlert[]>(`/v1/alerts${query ? `?${query}` : ''}`)
    await app.loadAlertCount()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load alerts'
  } finally {
    loading.value = false
  }
}

async function loadLocations() {
  locations.value = await apiClient<Location[]>('/v1/locations')
}

async function markRead(alert: InventoryAlert) {
  await patchJSON<InventoryAlert>(`/v1/alerts/${alert.id}/read`, {})
  await loadAlerts()
}

async function markAllRead() {
  await patchJSON<InventoryAlert[]>('/v1/alerts/read-all', {})
  await loadAlerts()
}

onMounted(async () => {
  await Promise.all([loadLocations(), loadAlerts()])
})
</script>

<template>
  <section>
    <PageHeader title="Alerts" eyebrow="Reorder point" description="Low stock, out-of-stock, and reorder point alerts for each location.">
      <div class="flex flex-wrap gap-2">
        <AppButton variant="secondary" @click="loadAlerts">Refresh</AppButton>
        <AppButton :disabled="unreadCount === 0" @click="markAllRead">Mark all read</AppButton>
      </div>
    </PageHeader>

    <div class="grid gap-4">
      <AppCard>
        <div class="grid gap-3 md:grid-cols-4">
          <AppSelect v-model="filters.unread" label="Unread">
            <option value="">All active</option>
            <option value="true">Unread only</option>
          </AppSelect>
          <AppSelect v-model="filters.type" label="Type">
            <option value="">All types</option>
            <option value="LOW_STOCK">LOW_STOCK</option>
            <option value="OUT_OF_STOCK">OUT_OF_STOCK</option>
            <option value="REORDER_POINT">REORDER_POINT</option>
          </AppSelect>
          <AppSelect v-model="filters.location_id" label="Location">
            <option value="">All locations</option>
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <div class="flex items-end">
            <AppButton class="w-full" @click="loadAlerts">Apply filters</AppButton>
          </div>
        </div>
      </AppCard>

      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      <div v-if="loading" class="rounded-lg border border-slate-200 bg-white p-6 text-sm text-slate-500">Loading alerts...</div>
      <AppEmptyState v-else-if="alerts.length === 0" title="No alerts" description="Stock alerts will appear when products reach threshold or reorder point." />

      <div v-else class="grid gap-3">
        <article v-for="alert in alerts" :key="alert.id" class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
          <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full px-2 py-1 text-xs font-bold" :class="typeClass(alert.type)">{{ alert.type }}</span>
                <AppBadge v-if="!alert.read_at">Unread</AppBadge>
                <span v-else class="rounded-full bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600">Read</span>
              </div>
              <h2 class="mt-2 font-bold">{{ alert.product_name }} · {{ alert.location_name }}</h2>
              <p class="mt-1 text-sm text-slate-600">{{ alert.message }}</p>
              <p class="mt-1 text-xs text-slate-500">{{ alert.sku }} · {{ new Date(alert.created_at).toLocaleString('th-TH') }}</p>
            </div>
            <div class="flex flex-wrap gap-2 md:justify-end">
              <RouterLink :to="alert.links.product" class="inline-flex min-h-10 items-center justify-center rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">
                Product
              </RouterLink>
              <RouterLink :to="alert.links.restock" class="inline-flex min-h-10 items-center justify-center rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">
                Restock
              </RouterLink>
              <RouterLink :to="alert.links.purchase_order" class="inline-flex min-h-10 items-center justify-center rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">
                Create PO
              </RouterLink>
              <AppButton v-if="!alert.read_at" @click="markRead(alert)">Mark read</AppButton>
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
