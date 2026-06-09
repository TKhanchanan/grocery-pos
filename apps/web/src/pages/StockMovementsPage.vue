<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiClient } from '../api/client'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import type { StockMovement } from '../types/navigation'
import { formatAppDateTime } from '../utils/date'

const movements = ref<StockMovement[]>([])
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    movements.value = await apiClient<StockMovement[]>('/v1/stock-movements')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load stock movements'
  } finally {
    loading.value = false
  }
}

function signed(value: number) {
  return value > 0 ? `+${value}` : String(value)
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="Stock Movements" eyebrow="Audit trail" description="Every restock and adjustment records before and after stock." />
    <AppCard>
      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      <div v-else-if="loading" class="text-sm text-slate-500">Loading movements...</div>
      <AppEmptyState v-else-if="movements.length === 0" title="No movements" description="Restock or adjust stock to create movement history." />
      <div v-else>
        <div class="hidden overflow-x-auto md:block">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50">
              <tr>
                <th class="px-3 py-2 text-left">Time</th>
                <th class="px-3 py-2 text-left">Product</th>
                <th class="px-3 py-2 text-left">Location</th>
                <th class="px-3 py-2 text-left">Type</th>
                <th class="px-3 py-2 text-right">Change</th>
                <th class="px-3 py-2 text-right">Before</th>
                <th class="px-3 py-2 text-right">After</th>
                <th class="px-3 py-2 text-left">Note</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="movement in movements" :key="movement.id">
                <td class="px-3 py-2">{{ formatAppDateTime(movement.created_at, 'en') }}</td>
                <td class="px-3 py-2"><b>{{ movement.product_name }}</b><br /><span class="text-xs text-slate-500">{{ movement.sku }}</span></td>
                <td class="px-3 py-2">{{ movement.location_name }}</td>
                <td class="px-3 py-2">{{ movement.reference_type }}</td>
                <td class="px-3 py-2 text-right font-bold" :class="movement.quantity_change < 0 ? 'text-red-600' : 'text-brand-700'">{{ signed(movement.quantity_change) }}</td>
                <td class="px-3 py-2 text-right">{{ movement.before_stock }}</td>
                <td class="px-3 py-2 text-right">{{ movement.after_stock }}</td>
                <td class="px-3 py-2">{{ movement.note || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-3 md:hidden">
          <article v-for="movement in movements" :key="movement.id" class="rounded-lg border border-slate-200 p-3">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h3 class="font-bold">{{ movement.product_name }}</h3>
                <p class="text-sm text-slate-500">{{ movement.location_name }} · {{ movement.reference_type }}</p>
              </div>
              <span class="font-bold" :class="movement.quantity_change < 0 ? 'text-red-600' : 'text-brand-700'">{{ signed(movement.quantity_change) }}</span>
            </div>
            <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
              <div><dt class="text-slate-500">Before</dt><dd class="font-semibold">{{ movement.before_stock }}</dd></div>
              <div><dt class="text-slate-500">After</dt><dd class="font-semibold">{{ movement.after_stock }}</dd></div>
              <div class="col-span-2"><dt class="text-slate-500">Note</dt><dd class="font-semibold">{{ movement.note || '-' }}</dd></div>
            </dl>
          </article>
        </div>
      </div>
    </AppCard>
  </section>
</template>
