<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { downloadFile } from '../api/download'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppInput from '../components/AppInput.vue'
import PageHeader from '../components/PageHeader.vue'

const loadingKey = ref('')
const error = ref('')

const filters = reactive({
  month: new Date().toISOString().slice(0, 7),
  date_from: '',
  date_to: '',
})

const canExportMonth = computed(() => Boolean(filters.month))

async function runExport(key: string, path: string, filename: string) {
  loadingKey.value = key
  error.value = ''
  try {
    await downloadFile(path, filename)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Export failed'
  } finally {
    loadingKey.value = ''
  }
}

function exportInventory() {
  runExport('inventory', `/v1/exports/inventory-monthly?month=${filters.month}&format=csv`, `inventory-monthly-${filters.month}.csv`)
}

function exportProducts() {
  runExport('products', '/v1/exports/products?format=csv', 'products.csv')
}

function exportSales() {
  const params = new URLSearchParams({ format: 'csv' })
  if (filters.date_from) params.set('date_from', filters.date_from)
  if (filters.date_to) params.set('date_to', filters.date_to)
  runExport('sales', `/v1/exports/sales?${params.toString()}`, 'sales.csv')
}

function exportProfit() {
  runExport('profit', `/v1/exports/profit?month=${filters.month}&format=csv`, `profit-${filters.month}.csv`)
}
</script>

<template>
  <section>
    <PageHeader title="Exports" eyebrow="CSV downloads" description="Download inventory, product, sales, and profit reports as Excel-friendly CSV files." />

    <div class="grid gap-4">
      <AppCard>
        <div class="grid gap-3 md:grid-cols-3">
          <AppInput v-model="filters.month" label="Month" type="month" />
          <AppInput v-model="filters.date_from" label="Sales date from" type="date" />
          <AppInput v-model="filters.date_to" label="Sales date to" type="date" />
        </div>
        <div v-if="error" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      </AppCard>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AppCard>
          <h2 class="font-bold">Inventory monthly</h2>
          <p class="mt-1 text-sm text-slate-500">Current stock valuation labeled by selected month.</p>
          <AppButton class="mt-4 w-full" :disabled="!canExportMonth || Boolean(loadingKey)" @click="exportInventory">
            {{ loadingKey === 'inventory' ? 'Downloading...' : 'Download CSV' }}
          </AppButton>
        </AppCard>
        <AppCard>
          <h2 class="font-bold">Product list</h2>
          <p class="mt-1 text-sm text-slate-500">SKU, barcode, category, price, cost, and stock totals.</p>
          <AppButton class="mt-4 w-full" :disabled="Boolean(loadingKey)" @click="exportProducts">
            {{ loadingKey === 'products' ? 'Downloading...' : 'Download CSV' }}
          </AppButton>
        </AppCard>
        <AppCard>
          <h2 class="font-bold">Sales</h2>
          <p class="mt-1 text-sm text-slate-500">Completed sales only, with date filters.</p>
          <AppButton class="mt-4 w-full" :disabled="Boolean(loadingKey)" @click="exportSales">
            {{ loadingKey === 'sales' ? 'Downloading...' : 'Download CSV' }}
          </AppButton>
        </AppCard>
        <AppCard>
          <h2 class="font-bold">Profit</h2>
          <p class="mt-1 text-sm text-slate-500">Monthly profit by product from sale snapshots.</p>
          <AppButton class="mt-4 w-full" :disabled="!canExportMonth || Boolean(loadingKey)" @click="exportProfit">
            {{ loadingKey === 'profit' ? 'Downloading...' : 'Download CSV' }}
          </AppButton>
        </AppCard>
      </div>
    </div>
  </section>
</template>
