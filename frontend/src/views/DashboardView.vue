<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../api'
import StatusBox from '../components/StatusBox.vue'
import type { Alert, DashboardSummary } from '../types'

const summary = ref<DashboardSummary | null>(null)
const alerts = ref<Alert[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await api<DashboardSummary>('/dashboard')
    alerts.value = await api<Alert[]>('/alerts')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load dashboard'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="space-y-5">
    <div>
      <p class="label">Today</p>
      <h2 class="text-2xl font-bold">Dashboard</h2>
    </div>
    <StatusBox :loading="loading" :error="error">
      <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <article v-for="[label, value] in [
          ['Revenue', summary?.revenue.toFixed(2)],
          ['Profit', summary?.profit.toFixed(2)],
          ['Sales', summary?.salesCount],
          ['Inventory value', summary?.inventoryValue.toFixed(2)]
        ]" :key="label" class="panel">
          <p class="label">{{ label }}</p>
          <p class="mt-2 text-2xl font-bold text-leaf">{{ value }}</p>
        </article>
      </div>
      <div class="grid gap-4 lg:grid-cols-[1fr_1fr]">
        <section class="panel">
          <h3 class="font-bold">Demo Checklist</h3>
          <ol class="mt-3 grid gap-2 text-sm text-slate-700">
            <li>1. Create/edit product with SKU, barcode, threshold, reorder point.</li>
            <li>2. Restock ไข่เค็ม 100 ฟอง into หน้าร้าน.</li>
            <li>3. Sell ไข่เค็ม 3 ฟอง from POS and print receipt.</li>
            <li>4. Cancel sale, export reports, import CSV, receive PO, transfer stock.</li>
          </ol>
        </section>
        <section class="panel">
          <div class="flex items-center justify-between">
            <h3 class="font-bold">Active Alerts</h3>
            <span class="rounded-full bg-red-50 px-3 py-1 text-xs font-bold text-red-700">{{ alerts.length }}</span>
          </div>
          <ul class="mt-3 space-y-2 text-sm">
            <li v-for="alert in alerts.slice(0, 6)" :key="alert.id" class="rounded-md bg-amber-50 p-3">
              <b>{{ alert.type }}</b> · {{ alert.productName }} at {{ alert.locationName }} ({{ alert.currentStock }})
            </li>
            <li v-if="alerts.length === 0" class="text-slate-500">No active alerts.</li>
          </ul>
        </section>
      </div>
    </StatusBox>
  </section>
</template>
