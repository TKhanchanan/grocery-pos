<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient } from '../api/client'
import { downloadFile } from '../api/download'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import type { Location, PaymentSummaryReport, ProductSalesReport, SalesPeriodReport, StockReport } from '../types/navigation'

type ReportKey = 'daily-sales' | 'monthly-sales' | 'best-selling' | 'profit-by-product' | 'stock' | 'inventory-valuation' | 'payment-summary' | 'low-stock' | 'reorder'
type ReportRow = SalesPeriodReport | ProductSalesReport | StockReport | PaymentSummaryReport

const tabs: { key: ReportKey; label: string; endpoint: string }[] = [
  { key: 'daily-sales', label: 'Daily Sales', endpoint: '/v1/reports/daily-sales' },
  { key: 'monthly-sales', label: 'Monthly Sales', endpoint: '/v1/reports/monthly-sales' },
  { key: 'best-selling', label: 'Best Selling', endpoint: '/v1/reports/best-selling' },
  { key: 'profit-by-product', label: 'Profit by Product', endpoint: '/v1/reports/profit-by-product' },
  { key: 'stock', label: 'Stock', endpoint: '/v1/reports/stock' },
  { key: 'inventory-valuation', label: 'Valuation', endpoint: '/v1/reports/inventory-valuation' },
  { key: 'payment-summary', label: 'Payments', endpoint: '/v1/reports/payment-summary' },
  { key: 'low-stock', label: 'Low Stock', endpoint: '/v1/reports/low-stock' },
  { key: 'reorder', label: 'Reorder', endpoint: '/v1/reports/reorder' },
]

const activeTab = ref<ReportKey>('daily-sales')
const rows = ref<ReportRow[]>([])
const locations = ref<Location[]>([])
const loading = ref(false)
const exportLoading = ref(false)
const error = ref('')

const filters = reactive({
  date_from: '',
  date_to: '',
  month: '',
  location_id: '',
})

const active = computed(() => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0])
const isStockReport = computed(() => ['stock', 'inventory-valuation', 'low-stock', 'reorder'].includes(activeTab.value))
const isProductReport = computed(() => ['best-selling', 'profit-by-product'].includes(activeTab.value))
const isPeriodReport = computed(() => ['daily-sales', 'monthly-sales'].includes(activeTab.value))
const isPaymentReport = computed(() => activeTab.value === 'payment-summary')
const exportKind = computed<'inventory' | 'products' | 'sales' | 'profit'>(() => {
  if (['stock', 'inventory-valuation', 'low-stock', 'reorder'].includes(activeTab.value)) return 'inventory'
  if (activeTab.value === 'profit-by-product') return 'profit'
  if (activeTab.value === 'best-selling') return 'products'
  return 'sales'
})

const summary = computed(() => {
  if (isStockReport.value) {
    const stockRows = rows.value as StockReport[]
    return {
      count: stockRows.length,
      revenue: stockRows.reduce((sum, row) => sum + row.total_value, 0),
      profit: 0,
      quantity: stockRows.reduce((sum, row) => sum + row.quantity, 0),
    }
  }
  if (isProductReport.value) {
    const productRows = rows.value as ProductSalesReport[]
    return {
      count: productRows.length,
      revenue: productRows.reduce((sum, row) => sum + row.revenue, 0),
      profit: productRows.reduce((sum, row) => sum + row.profit, 0),
      quantity: productRows.reduce((sum, row) => sum + row.quantity, 0),
    }
  }
  if (isPaymentReport.value) {
    const paymentRows = rows.value as PaymentSummaryReport[]
    return {
      count: paymentRows.reduce((sum, row) => sum + row.receipt_count, 0),
      revenue: paymentRows.reduce((sum, row) => sum + row.revenue, 0),
      profit: 0,
      quantity: 0,
    }
  }
  const periodRows = rows.value as SalesPeriodReport[]
  return {
    count: periodRows.reduce((sum, row) => sum + row.receipt_count, 0),
    revenue: periodRows.reduce((sum, row) => sum + row.revenue, 0),
    profit: periodRows.reduce((sum, row) => sum + row.profit, 0),
    quantity: 0,
  }
})

function money(value: number) {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function buildQuery() {
  const params = new URLSearchParams()
  if (!isStockReport.value) {
    if (filters.date_from) params.set('date_from', filters.date_from)
    if (filters.date_to) params.set('date_to', filters.date_to)
    if (filters.month) params.set('month', filters.month)
  }
  if (filters.location_id) params.set('location_id', filters.location_id)
  return params.toString()
}

async function loadReport() {
  loading.value = true
  error.value = ''
  try {
    const query = buildQuery()
    rows.value = await apiClient<ReportRow[]>(`${active.value.endpoint}${query ? `?${query}` : ''}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load report'
  } finally {
    loading.value = false
  }
}

async function exportCSV() {
  exportLoading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ format: 'csv' })
    if (exportKind.value === 'inventory') {
      params.set('month', filters.month || new Date().toISOString().slice(0, 7))
      await downloadFile(`/v1/exports/inventory-monthly?${params.toString()}`, 'inventory-monthly.csv')
      return
    }
    if (exportKind.value === 'products') {
      await downloadFile('/v1/exports/products?format=csv', 'products.csv')
      return
    }
    if (exportKind.value === 'profit') {
      params.set('month', filters.month || new Date().toISOString().slice(0, 7))
      if (filters.location_id) params.set('location_id', filters.location_id)
      await downloadFile(`/v1/exports/profit?${params.toString()}`, 'profit.csv')
      return
    }
    if (filters.date_from) params.set('date_from', filters.date_from)
    if (filters.date_to) params.set('date_to', filters.date_to)
    if (filters.month) params.set('month', filters.month)
    if (filters.location_id) params.set('location_id', filters.location_id)
    await downloadFile(`/v1/exports/sales?${params.toString()}`, 'sales.csv')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Export failed'
  } finally {
    exportLoading.value = false
  }
}

async function loadLocations() {
  locations.value = await apiClient<Location[]>('/v1/locations')
}

function setTab(key: ReportKey) {
  activeTab.value = key
  loadReport()
}

function clearFilters() {
  filters.date_from = ''
  filters.date_to = ''
  filters.month = ''
  filters.location_id = ''
  loadReport()
}

onMounted(async () => {
  await loadLocations()
  await loadReport()
})
</script>

<template>
  <section>
    <PageHeader title="Reports" eyebrow="Analytics" description="Sales, profit, stock, valuation, payments, low stock, and reorder reports.">
      <AppButton variant="secondary" @click="loadReport">Refresh</AppButton>
    </PageHeader>

    <div class="grid gap-4">
      <AppCard>
        <div class="flex gap-2 overflow-x-auto pb-1">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="whitespace-nowrap rounded-md px-3 py-2 text-sm font-semibold"
            :class="activeTab === tab.key ? 'bg-brand-600 text-white' : 'border border-slate-200 bg-white text-slate-700'"
            @click="setTab(tab.key)"
          >
            {{ tab.label }}
          </button>
        </div>
      </AppCard>

      <AppCard>
        <div class="grid gap-3 md:grid-cols-5">
          <AppInput v-model="filters.date_from" label="Date from" type="date" :disabled="isStockReport" />
          <AppInput v-model="filters.date_to" label="Date to" type="date" :disabled="isStockReport" />
          <AppInput v-model="filters.month" label="Month" type="month" :disabled="isStockReport" />
          <AppSelect v-model="filters.location_id" label="Location">
            <option value="">All locations</option>
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <div class="flex items-end gap-2">
            <AppButton class="flex-1" @click="loadReport">Apply</AppButton>
            <AppButton class="flex-1" variant="secondary" @click="clearFilters">Reset</AppButton>
          </div>
        </div>
      </AppCard>

      <div class="grid gap-3 md:grid-cols-4">
        <AppCard>
          <p class="text-sm text-slate-500">{{ isStockReport ? 'Rows' : 'Receipts/Rows' }}</p>
          <p class="mt-1 text-2xl font-bold text-brand-700">{{ summary.count }}</p>
        </AppCard>
        <AppCard>
          <p class="text-sm text-slate-500">{{ isStockReport ? 'Inventory value' : 'Revenue' }}</p>
          <p class="mt-1 text-2xl font-bold text-brand-700">{{ money(summary.revenue) }}</p>
        </AppCard>
        <AppCard>
          <p class="text-sm text-slate-500">{{ isStockReport ? 'Quantity' : 'Profit' }}</p>
          <p class="mt-1 text-2xl font-bold text-brand-700">{{ isStockReport ? summary.quantity : money(summary.profit) }}</p>
        </AppCard>
        <AppCard>
          <p class="text-sm text-slate-500">Export</p>
          <div class="mt-2 flex gap-2">
            <AppButton variant="secondary" :disabled="exportLoading" @click="exportCSV">{{ exportLoading ? 'Downloading...' : 'CSV' }}</AppButton>
            <AppButton variant="secondary" disabled>Excel</AppButton>
          </div>
        </AppCard>
      </div>

      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      <div v-if="loading" class="rounded-lg border border-slate-200 bg-white p-6 text-sm text-slate-500">Loading report...</div>
      <AppEmptyState v-else-if="rows.length === 0" title="No report data" description="Try changing filters or create more sales and stock activity." />

      <AppCard v-else>
        <div class="hidden overflow-x-auto lg:block">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50">
              <tr v-if="isPeriodReport">
                <th class="px-3 py-2 text-left">Period</th>
                <th class="px-3 py-2 text-right">Receipts</th>
                <th class="px-3 py-2 text-right">Revenue</th>
                <th class="px-3 py-2 text-right">Cost</th>
                <th class="px-3 py-2 text-right">Profit</th>
              </tr>
              <tr v-else-if="isProductReport">
                <th class="px-3 py-2 text-left">Product</th>
                <th class="px-3 py-2 text-right">Quantity</th>
                <th class="px-3 py-2 text-right">Revenue</th>
                <th class="px-3 py-2 text-right">Cost</th>
                <th class="px-3 py-2 text-right">Profit</th>
              </tr>
              <tr v-else-if="isPaymentReport">
                <th class="px-3 py-2 text-left">Payment</th>
                <th class="px-3 py-2 text-right">Receipts</th>
                <th class="px-3 py-2 text-right">Revenue</th>
              </tr>
              <tr v-else>
                <th class="px-3 py-2 text-left">Product</th>
                <th class="px-3 py-2 text-left">Location</th>
                <th class="px-3 py-2 text-right">Qty</th>
                <th class="px-3 py-2 text-right">Unit cost</th>
                <th class="px-3 py-2 text-right">Value</th>
                <th class="px-3 py-2 text-left">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="(row, index) in rows" :key="index">
                <template v-if="isPeriodReport">
                  <td class="px-3 py-2 font-semibold">{{ (row as SalesPeriodReport).period }}</td>
                  <td class="px-3 py-2 text-right">{{ (row as SalesPeriodReport).receipt_count }}</td>
                  <td class="px-3 py-2 text-right">{{ money((row as SalesPeriodReport).revenue) }}</td>
                  <td class="px-3 py-2 text-right">{{ money((row as SalesPeriodReport).cost) }}</td>
                  <td class="px-3 py-2 text-right font-semibold">{{ money((row as SalesPeriodReport).profit) }}</td>
                </template>
                <template v-else-if="isProductReport">
                  <td class="px-3 py-2"><p class="font-semibold">{{ (row as ProductSalesReport).product_name }}</p><p class="text-xs text-slate-500">{{ (row as ProductSalesReport).sku }}</p></td>
                  <td class="px-3 py-2 text-right">{{ (row as ProductSalesReport).quantity }}</td>
                  <td class="px-3 py-2 text-right">{{ money((row as ProductSalesReport).revenue) }}</td>
                  <td class="px-3 py-2 text-right">{{ money((row as ProductSalesReport).cost) }}</td>
                  <td class="px-3 py-2 text-right font-semibold">{{ money((row as ProductSalesReport).profit) }}</td>
                </template>
                <template v-else-if="isPaymentReport">
                  <td class="px-3 py-2 font-semibold">{{ (row as PaymentSummaryReport).payment_method }}</td>
                  <td class="px-3 py-2 text-right">{{ (row as PaymentSummaryReport).receipt_count }}</td>
                  <td class="px-3 py-2 text-right font-semibold">{{ money((row as PaymentSummaryReport).revenue) }}</td>
                </template>
                <template v-else>
                  <td class="px-3 py-2"><p class="font-semibold">{{ (row as StockReport).product_name }}</p><p class="text-xs text-slate-500">{{ (row as StockReport).sku }}</p></td>
                  <td class="px-3 py-2">{{ (row as StockReport).location_name }}</td>
                  <td class="px-3 py-2 text-right">{{ (row as StockReport).quantity }}</td>
                  <td class="px-3 py-2 text-right">{{ money((row as StockReport).unit_cost) }}</td>
                  <td class="px-3 py-2 text-right font-semibold">{{ money((row as StockReport).total_value) }}</td>
                  <td class="px-3 py-2">{{ (row as StockReport).stock_status.replaceAll('_', ' ') }}</td>
                </template>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-3 lg:hidden">
          <article v-for="(row, index) in rows" :key="index" class="rounded-lg border border-slate-200 p-3">
            <template v-if="isPeriodReport">
              <h2 class="font-bold">{{ (row as SalesPeriodReport).period }}</h2>
              <p class="mt-2 text-sm text-slate-500">{{ (row as SalesPeriodReport).receipt_count }} receipts</p>
              <p class="mt-2 text-xl font-bold">{{ money((row as SalesPeriodReport).revenue) }} บาท</p>
              <p class="text-sm text-slate-500">Profit {{ money((row as SalesPeriodReport).profit) }} บาท</p>
            </template>
            <template v-else-if="isProductReport">
              <h2 class="font-bold">{{ (row as ProductSalesReport).product_name }}</h2>
              <p class="text-sm text-slate-500">{{ (row as ProductSalesReport).sku }}</p>
              <p class="mt-2 text-xl font-bold">{{ (row as ProductSalesReport).quantity }} sold</p>
              <p class="text-sm text-slate-500">Revenue {{ money((row as ProductSalesReport).revenue) }} · Profit {{ money((row as ProductSalesReport).profit) }}</p>
            </template>
            <template v-else-if="isPaymentReport">
              <h2 class="font-bold">{{ (row as PaymentSummaryReport).payment_method }}</h2>
              <p class="mt-2 text-xl font-bold">{{ money((row as PaymentSummaryReport).revenue) }} บาท</p>
              <p class="text-sm text-slate-500">{{ (row as PaymentSummaryReport).receipt_count }} receipts</p>
            </template>
            <template v-else>
              <h2 class="font-bold">{{ (row as StockReport).product_name }}</h2>
              <p class="text-sm text-slate-500">{{ (row as StockReport).location_name }} · {{ (row as StockReport).sku }}</p>
              <p class="mt-2 text-xl font-bold">{{ (row as StockReport).quantity }} units</p>
              <p class="text-sm text-slate-500">Value {{ money((row as StockReport).total_value) }} · {{ (row as StockReport).stock_status.replaceAll('_', ' ') }}</p>
            </template>
          </article>
        </div>
      </AppCard>
    </div>
  </section>
</template>
