<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppIcon from '../components/AppIcon.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import ProductAvatar from '../components/ProductAvatar.vue'
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { IconName } from '../types/icons'
import type { Location, PaymentSummaryReport, ProductSalesReport, SalesPeriodReport, StockReport, StockStatus } from '../types/navigation'

type ReportKey = 'daily-sales' | 'monthly-sales' | 'best-selling' | 'profit-by-product' | 'stock' | 'inventory-valuation' | 'payment-summary' | 'low-stock' | 'reorder'
type ReportRow = SalesPeriodReport | ProductSalesReport | StockReport | PaymentSummaryReport
type ExportKind = 'inventory' | 'products' | 'sales' | 'profit'
type CsvCell = string | number | boolean | null | undefined

interface ReportTab {
  key: ReportKey
  labelKey: TranslationKey
  endpoint: string
  icon: IconName
}

interface KpiCard {
  label: string
  value: string | number
  helper: string
  icon: IconName
  tone?: 'brand' | 'success' | 'warning' | 'danger' | 'info'
}

const app = useAppStore()
const auth = useAuthStore()

const tabs: ReportTab[] = [
  { key: 'daily-sales', labelKey: 'reports.tab.dailySales', endpoint: '/v1/reports/daily-sales', icon: 'receipt-text' },
  { key: 'monthly-sales', labelKey: 'reports.tab.monthlySales', endpoint: '/v1/reports/monthly-sales', icon: 'history' },
  { key: 'best-selling', labelKey: 'reports.tab.bestSelling', endpoint: '/v1/reports/best-selling', icon: 'package' },
  { key: 'profit-by-product', labelKey: 'reports.tab.profitByProduct', endpoint: '/v1/reports/profit-by-product', icon: 'banknote' },
  { key: 'stock', labelKey: 'reports.tab.stock', endpoint: '/v1/reports/stock', icon: 'package' },
  { key: 'inventory-valuation', labelKey: 'reports.tab.valuation', endpoint: '/v1/reports/inventory-valuation', icon: 'chart-column' },
  { key: 'payment-summary', labelKey: 'reports.tab.payments', endpoint: '/v1/reports/payment-summary', icon: 'qr-code' },
  { key: 'low-stock', labelKey: 'reports.tab.lowStock', endpoint: '/v1/reports/low-stock', icon: 'triangle-alert' },
  { key: 'reorder', labelKey: 'reports.tab.reorder', endpoint: '/v1/reports/reorder', icon: 'clipboard-list' },
]

const activeTab = ref<ReportKey>('daily-sales')
const rows = ref<ReportRow[]>([])
const locations = ref<Location[]>([])
const loading = ref(false)
const exportLoading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = ref(20)

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
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')
const totalPages = computed(() => Math.max(1, Math.ceil(rows.value.length / pageSize.value)))
const visibleRows = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return rows.value.slice(start, start + pageSize.value)
})
const exportKind = computed<ExportKind>(() => {
  if (['stock', 'inventory-valuation', 'low-stock', 'reorder'].includes(activeTab.value)) return 'inventory'
  if (activeTab.value === 'profit-by-product') return 'profit'
  if (activeTab.value === 'best-selling') return 'products'
  return 'sales'
})
const exportPermission = computed(() => ({
  inventory: 'exports.inventory',
  products: 'exports.products',
  sales: 'exports.sales',
  profit: 'exports.profit',
}[exportKind.value]))
const canExport = computed(() => auth.hasPermission(exportPermission.value))
const summary = computed(() => {
  if (isStockReport.value) {
    const stockRows = rows.value as StockReport[]
    return {
      rows: stockRows.length,
      receipts: 0,
      revenue: 0,
      cost: stockRows.reduce((sum, row) => sum + row.unit_cost * row.quantity, 0),
      profit: 0,
      quantity: stockRows.reduce((sum, row) => sum + row.quantity, 0),
      value: stockRows.reduce((sum, row) => sum + row.total_value, 0),
    }
  }
  if (isProductReport.value) {
    const productRows = rows.value as ProductSalesReport[]
    return {
      rows: productRows.length,
      receipts: 0,
      revenue: productRows.reduce((sum, row) => sum + row.revenue, 0),
      cost: productRows.reduce((sum, row) => sum + row.cost, 0),
      profit: productRows.reduce((sum, row) => sum + row.profit, 0),
      quantity: productRows.reduce((sum, row) => sum + row.quantity, 0),
      value: 0,
    }
  }
  if (isPaymentReport.value) {
    const paymentRows = rows.value as PaymentSummaryReport[]
    return {
      rows: paymentRows.length,
      receipts: paymentRows.reduce((sum, row) => sum + row.receipt_count, 0),
      revenue: paymentRows.reduce((sum, row) => sum + row.revenue, 0),
      cost: 0,
      profit: 0,
      quantity: 0,
      value: 0,
    }
  }
  const periodRows = rows.value as SalesPeriodReport[]
  return {
    rows: periodRows.length,
    receipts: periodRows.reduce((sum, row) => sum + row.receipt_count, 0),
    revenue: periodRows.reduce((sum, row) => sum + row.revenue, 0),
    cost: periodRows.reduce((sum, row) => sum + row.cost, 0),
    profit: periodRows.reduce((sum, row) => sum + row.profit, 0),
    quantity: 0,
    value: 0,
  }
})
const kpiCards = computed<KpiCard[]>(() => {
  if (isStockReport.value) {
    return [
      { label: app.t('reports.kpi.products'), value: summary.value.rows, helper: app.t('reports.kpi.productsHelper'), icon: 'package' },
      { label: app.t('reports.kpi.stockQuantity'), value: quantity(summary.value.quantity), helper: app.t('reports.kpi.stockQuantityHelper'), icon: 'clipboard-list', tone: 'info' },
      { label: app.t('reports.kpi.stockValue'), value: money(summary.value.value), helper: app.t('reports.kpi.stockValueHelper'), icon: 'banknote', tone: 'success' },
    ]
  }
  if (isPaymentReport.value) {
    return [
      { label: app.t('reports.kpi.transactions'), value: summary.value.receipts, helper: app.t('reports.kpi.transactionsHelper'), icon: 'receipt-text' },
      { label: app.t('reports.kpi.totalPayment'), value: money(summary.value.revenue), helper: app.t('reports.kpi.totalPaymentHelper'), icon: 'banknote', tone: 'success' },
      { label: app.t('reports.kpi.rows'), value: summary.value.rows, helper: app.t('reports.kpi.rowsHelper'), icon: 'chart-column', tone: 'info' },
    ]
  }
  if (isProductReport.value) {
    return [
      { label: app.t('reports.kpi.rows'), value: summary.value.rows, helper: app.t('reports.kpi.rowsHelper'), icon: 'package' },
      { label: app.t('reports.kpi.quantity'), value: quantity(summary.value.quantity), helper: app.t('reports.kpi.quantityHelper'), icon: 'clipboard-list', tone: 'info' },
      { label: app.t('reports.kpi.revenue'), value: money(summary.value.revenue), helper: app.t('reports.kpi.revenueHelper'), icon: 'banknote', tone: 'success' },
      { label: app.t('reports.kpi.profit'), value: money(summary.value.profit), helper: app.t('reports.kpi.profitHelper'), icon: 'chart-column', tone: summary.value.profit < 0 ? 'danger' : 'success' },
    ]
  }
  return [
    { label: app.t('reports.kpi.receipts'), value: summary.value.receipts, helper: app.t('reports.kpi.receiptsHelper'), icon: 'receipt-text' },
    { label: app.t('reports.kpi.revenue'), value: money(summary.value.revenue), helper: app.t('reports.kpi.revenueHelper'), icon: 'banknote', tone: 'success' },
    { label: app.t('reports.kpi.cost'), value: money(summary.value.cost), helper: app.t('reports.kpi.costHelper'), icon: 'package', tone: 'info' },
    { label: app.t('reports.kpi.profit'), value: money(summary.value.profit), helper: app.t('reports.kpi.profitHelper'), icon: 'chart-column', tone: summary.value.profit < 0 ? 'danger' : 'success' },
  ]
})

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function money(value: number) {
  return t('reports.currency', { amount: value.toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) })
}

function quantity(value: number) {
  return value.toLocaleString(locale.value)
}

function profitClass(value: number) {
  return value < 0 ? 'text-red-600 dark:text-red-300' : 'text-brand-700 dark:text-emerald-200'
}

function paymentLabel(method: string) {
  return method === 'QR' ? app.t('reports.payment.qr') : app.t('reports.payment.cash')
}

function stockStatusLabel(status: StockStatus) {
  const key = `reports.stockStatus.${status}` as TranslationKey
  return app.t(key)
}

function stockStatusTone(status: StockStatus) {
  if (status === 'out_of_stock') return 'danger'
  if (status === 'low_stock' || status === 'reorder_point') return 'warning'
  return 'success'
}

function suggestedAction(row: StockReport) {
  if (row.quantity <= 0) return app.t('reports.action.restockNow')
  if (row.quantity <= row.reorder_point) return app.t('reports.action.createPO')
  return app.t('reports.action.monitor')
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
    page.value = 1
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('reports.loadFailed')
  } finally {
    loading.value = false
  }
}

function csvEscape(value: CsvCell) {
  const text = value === null || value === undefined ? '' : String(value)
  return /[",\n\r]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text
}

function downloadCSV(filename: string, data: CsvCell[][]) {
  const csv = data.map((row) => row.map(csvEscape).join(',')).join('\r\n')
  const blob = new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function activeExportFilename() {
  const suffix = new Date().toISOString().slice(0, 10)
  return `${activeTab.value}-${suffix}.csv`
}

function activeExportRows(): CsvCell[][] {
  if (isPeriodReport.value) {
    const periodLabel = activeTab.value === 'monthly-sales' ? app.t('reports.month') : app.t('reports.period')
    return [
      [periodLabel, app.t('reports.receipts'), app.t('reports.revenue'), app.t('reports.cost'), app.t('reports.profit')],
      ...(rows.value as SalesPeriodReport[]).map((row) => [row.period, row.receipt_count, row.revenue.toFixed(2), row.cost.toFixed(2), row.profit.toFixed(2)]),
    ]
  }
  if (isProductReport.value) {
    const headers = [app.t('reports.product'), app.t('reports.sku'), app.t('reports.soldQuantity'), app.t('reports.revenue')]
    if (activeTab.value === 'profit-by-product') headers.push(app.t('reports.cost'), app.t('reports.profit'))
    return [
      headers,
      ...(rows.value as ProductSalesReport[]).map((row) => {
        const values: CsvCell[] = [row.product_name, row.sku, row.quantity, row.revenue.toFixed(2)]
        if (activeTab.value === 'profit-by-product') values.push(row.cost.toFixed(2), row.profit.toFixed(2))
        return values
      }),
    ]
  }
  if (isPaymentReport.value) {
    return [
      [app.t('reports.paymentMethod'), app.t('reports.transactions'), app.t('reports.totalPayment')],
      ...(rows.value as PaymentSummaryReport[]).map((row) => [paymentLabel(row.payment_method), row.receipt_count, row.revenue.toFixed(2)]),
    ]
  }
  if (activeTab.value === 'inventory-valuation') {
    return [
      [app.t('reports.product'), app.t('reports.sku'), app.t('reports.stockQuantity'), app.t('reports.unitCost'), app.t('reports.totalValue')],
      ...(rows.value as StockReport[]).map((row) => [row.product_name, row.sku, row.quantity, row.unit_cost.toFixed(2), row.total_value.toFixed(2)]),
    ]
  }
  const headers = [app.t('reports.product'), app.t('reports.sku'), app.t('reports.location'), app.t('reports.stockQuantity')]
  if (activeTab.value === 'low-stock') headers.push(app.t('reports.threshold'))
  if (activeTab.value === 'reorder') headers.push(app.t('reports.reorderPoint'))
  headers.push(activeTab.value === 'reorder' ? app.t('reports.suggestedAction') : app.t('reports.status'))
  return [
    headers,
    ...(rows.value as StockReport[]).map((row) => {
      const values: CsvCell[] = [row.product_name, row.sku, row.location_name, row.quantity]
      if (activeTab.value === 'low-stock') values.push(row.threshold)
      if (activeTab.value === 'reorder') values.push(row.reorder_point)
      values.push(activeTab.value === 'reorder' ? suggestedAction(row) : stockStatusLabel(row.stock_status))
      return values
    }),
  ]
}

async function exportCSV() {
  if (!canExport.value) return
  exportLoading.value = true
  error.value = ''
  try {
    downloadCSV(activeExportFilename(), activeExportRows())
    app.pushToast({ type: 'success', message: app.t('reports.exportSuccess') })
  } catch (err) {
    const message = err instanceof Error ? err.message : app.t('reports.exportFailed')
    const friendly = message.toLowerCase().includes('permission') ? app.t('reports.noPermission') : message
    error.value = friendly
    app.pushToast({ type: 'error', message: app.t('reports.exportFailed'), description: friendly })
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

function changePageSize(value: string) {
  pageSize.value = Number(value)
  page.value = 1
}

function previousPage() {
  if (page.value > 1) page.value -= 1
}

function nextPage() {
  if (page.value < totalPages.value) page.value += 1
}

onMounted(async () => {
  await loadLocations()
  await loadReport()
})
</script>

<template>
  <section>
    <PageHeader :title="app.t('reports.title')" :eyebrow="app.t('reports.eyebrow')" :description="app.t('reports.description')" icon="chart-column">
      <AppButton variant="secondary" icon="history" @click="loadReport">{{ app.t('reports.refresh') }}</AppButton>
    </PageHeader>

    <div class="grid gap-4">
      <div class="flex gap-2 overflow-x-auto pb-1">
        <button v-for="tab in tabs" :key="tab.key"
          class="focus-ring inline-flex min-h-11 shrink-0 items-center gap-2 rounded-xl px-3.5 py-2 text-sm font-bold transition"
          :class="activeTab === tab.key ? 'bg-brand-600 text-white shadow-sm dark:bg-teal-300 dark:text-slate-950' : 'bg-white/80 text-slate-600 hover:bg-brand-50 dark:bg-slate-950/80 dark:text-slate-300 dark:hover:bg-teal-400/10'"
          @click="setTab(tab.key)">
          <AppIcon :name="tab.icon" :size="17" />
          {{ app.t(tab.labelKey) }}
        </button>
      </div>

      <AppCard class="dark:bg-slate-900/80">
        <div class="mb-4 flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('reports.filters') }}</p>
            <h2 class="text-lg font-black text-slate-950 dark:text-slate-50">{{ app.t(active.labelKey) }}</h2>
          </div>
          <AppBadge tone="info">{{ app.t('reports.reportCenter') }}</AppBadge>
        </div>
        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_1fr_auto]">
          <AppInput v-model="filters.date_from" :label="app.t('reports.dateFrom')" type="date" :disabled="isStockReport" />
          <AppInput v-model="filters.date_to" :label="app.t('reports.dateTo')" type="date" :disabled="isStockReport" />
          <AppInput v-model="filters.month" :label="app.t('reports.month')" type="month" :disabled="isStockReport" />
          <AppSelect v-model="filters.location_id" :label="app.t('reports.location')">
            <option value="">{{ app.t('reports.allLocations') }}</option>
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <div class="flex items-end gap-2">
            <AppButton class="flex-1 xl:flex-none" icon="search" @click="loadReport">{{ app.t('reports.apply') }}</AppButton>
            <AppButton class="flex-1 xl:flex-none" variant="secondary" icon="x" @click="clearFilters">{{ app.t('reports.reset') }}</AppButton>
          </div>
        </div>
      </AppCard>

      <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <StatCard v-for="card in kpiCards" :key="card.label" :label="card.label" :value="card.value" :helper="card.helper" :icon="card.icon" :tone="card.tone" />
      </div>

      <div v-if="error" class="rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <span>{{ error }}</span>
          <AppButton variant="secondary" @click="loadReport">{{ app.t('reports.retry') }}</AppButton>
        </div>
      </div>
      <AppLoadingState v-if="loading" :label="app.t('reports.loading')" />
      <AppEmptyState v-else-if="rows.length === 0" :title="app.t('reports.empty')" :description="app.t('reports.emptyDescription')" icon="chart-column" />

      <AppCard v-else class="dark:bg-slate-900/80">
        <div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('reports.rows') }}</p>
            <h2 class="text-lg font-black text-slate-950 dark:text-slate-50">{{ app.t(active.labelKey) }}</h2>
          </div>
          <div class="flex flex-wrap items-center gap-2 sm:justify-end">
            <AppBadge tone="neutral">{{ rows.length.toLocaleString(locale) }} {{ app.t('reports.rows') }}</AppBadge>
            <AppButton v-if="canExport" variant="secondary" icon="download" :loading="exportLoading" @click="exportCSV">{{ app.t('reports.exportCSV') }}</AppButton>
            <AppButton v-if="canExport" variant="secondary" icon="download" disabled>{{ app.t('reports.exportExcel') }}</AppButton>
          </div>
        </div>

        <div class="hidden overflow-x-auto lg:block">
          <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
            <thead class="bg-slate-50 dark:bg-slate-950/70">
              <tr v-if="isPeriodReport">
                <th class="px-3 py-3 text-left">{{ activeTab === 'monthly-sales' ? app.t('reports.month') : app.t('reports.period') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.receipts') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.revenue') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.cost') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.profit') }}</th>
              </tr>
              <tr v-else-if="isProductReport">
                <th class="px-3 py-3 text-left">{{ app.t('reports.product') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('reports.sku') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.soldQuantity') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.revenue') }}</th>
                <th v-if="activeTab === 'profit-by-product'" class="px-3 py-3 text-right">{{ app.t('reports.cost') }}</th>
                <th v-if="activeTab === 'profit-by-product'" class="px-3 py-3 text-right">{{ app.t('reports.profit') }}</th>
              </tr>
              <tr v-else-if="isPaymentReport">
                <th class="px-3 py-3 text-left">{{ app.t('reports.paymentMethod') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.transactions') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.totalPayment') }}</th>
              </tr>
              <tr v-else-if="activeTab === 'inventory-valuation'">
                <th class="px-3 py-3 text-left">{{ app.t('reports.product') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('reports.sku') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.stockQuantity') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.unitCost') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.totalValue') }}</th>
              </tr>
              <tr v-else>
                <th class="px-3 py-3 text-left">{{ app.t('reports.product') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('reports.sku') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('reports.location') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('reports.stockQuantity') }}</th>
                <th v-if="activeTab === 'low-stock'" class="px-3 py-3 text-right">{{ app.t('reports.threshold') }}</th>
                <th v-if="activeTab === 'reorder'" class="px-3 py-3 text-right">{{ app.t('reports.reorderPoint') }}</th>
                <th class="px-3 py-3 text-left">{{ activeTab === 'reorder' ? app.t('reports.suggestedAction') : app.t('reports.status') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="(row, index) in visibleRows" :key="index" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                <template v-if="isPeriodReport">
                  <td class="px-3 py-3 font-semibold">{{ (row as SalesPeriodReport).period }}</td>
                  <td class="px-3 py-3 text-right">{{ (row as SalesPeriodReport).receipt_count.toLocaleString(locale) }}</td>
                  <td class="px-3 py-3 text-right">{{ money((row as SalesPeriodReport).revenue) }}</td>
                  <td class="px-3 py-3 text-right">{{ money((row as SalesPeriodReport).cost) }}</td>
                  <td class="px-3 py-3 text-right font-semibold" :class="profitClass((row as SalesPeriodReport).profit)">{{ money((row as SalesPeriodReport).profit) }}</td>
                </template>
                <template v-else-if="isProductReport">
                  <td class="px-3 py-3">
                    <div class="flex min-w-0 items-center gap-3">
                      <ProductAvatar :src="(row as ProductSalesReport).image_url" :updated-at="(row as ProductSalesReport).image_updated_at" :name="(row as ProductSalesReport).product_name" size="sm" shape="square" />
                      <p class="truncate font-semibold">{{ (row as ProductSalesReport).product_name }}</p>
                    </div>
                  </td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400">{{ (row as ProductSalesReport).sku }}</td>
                  <td class="px-3 py-3 text-right">{{ (row as ProductSalesReport).quantity.toLocaleString(locale) }}</td>
                  <td class="px-3 py-3 text-right">{{ money((row as ProductSalesReport).revenue) }}</td>
                  <td v-if="activeTab === 'profit-by-product'" class="px-3 py-3 text-right">{{ money((row as ProductSalesReport).cost) }}</td>
                  <td v-if="activeTab === 'profit-by-product'" class="px-3 py-3 text-right font-semibold" :class="profitClass((row as ProductSalesReport).profit)">{{ money((row as ProductSalesReport).profit) }}</td>
                </template>
                <template v-else-if="isPaymentReport">
                  <td class="px-3 py-3 font-semibold">{{ paymentLabel((row as PaymentSummaryReport).payment_method) }}</td>
                  <td class="px-3 py-3 text-right">{{ (row as PaymentSummaryReport).receipt_count.toLocaleString(locale) }}</td>
                  <td class="px-3 py-3 text-right font-semibold">{{ money((row as PaymentSummaryReport).revenue) }}</td>
                </template>
                <template v-else-if="activeTab === 'inventory-valuation'">
                  <td class="px-3 py-3">
                    <div class="flex min-w-0 items-center gap-3">
                      <ProductAvatar :src="(row as StockReport).image_url" :updated-at="(row as StockReport).image_updated_at" :name="(row as StockReport).product_name" size="sm" shape="square" />
                      <p class="truncate font-semibold">{{ (row as StockReport).product_name }}</p>
                    </div>
                  </td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400">{{ (row as StockReport).sku }}</td>
                  <td class="px-3 py-3 text-right">{{ (row as StockReport).quantity.toLocaleString(locale) }}</td>
                  <td class="px-3 py-3 text-right">{{ money((row as StockReport).unit_cost) }}</td>
                  <td class="px-3 py-3 text-right font-semibold">{{ money((row as StockReport).total_value) }}</td>
                </template>
                <template v-else>
                  <td class="px-3 py-3">
                    <div class="flex min-w-0 items-center gap-3">
                      <ProductAvatar :src="(row as StockReport).image_url" :updated-at="(row as StockReport).image_updated_at" :name="(row as StockReport).product_name" size="sm" shape="square" />
                      <p class="truncate font-semibold">{{ (row as StockReport).product_name }}</p>
                    </div>
                  </td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400">{{ (row as StockReport).sku }}</td>
                  <td class="px-3 py-3">{{ (row as StockReport).location_name }}</td>
                  <td class="px-3 py-3 text-right">{{ (row as StockReport).quantity.toLocaleString(locale) }}</td>
                  <td v-if="activeTab === 'low-stock'" class="px-3 py-3 text-right">{{ (row as StockReport).threshold.toLocaleString(locale) }}</td>
                  <td v-if="activeTab === 'reorder'" class="px-3 py-3 text-right">{{ (row as StockReport).reorder_point.toLocaleString(locale) }}</td>
                  <td class="px-3 py-3">
                    <AppBadge v-if="activeTab !== 'reorder'" :tone="stockStatusTone((row as StockReport).stock_status)">{{ stockStatusLabel((row as StockReport).stock_status) }}</AppBadge>
                    <span v-else class="font-semibold text-brand-700 dark:text-emerald-200">{{ suggestedAction(row as StockReport) }}</span>
                  </td>
                </template>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-3 lg:hidden">
          <article v-for="(row, index) in visibleRows" :key="index" class="rounded-2xl border border-slate-200 bg-white/65 p-4 dark:border-slate-700 dark:bg-slate-950/60">
            <template v-if="isPeriodReport">
              <h3 class="font-black">{{ (row as SalesPeriodReport).period }}</h3>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ (row as SalesPeriodReport).receipt_count.toLocaleString(locale) }} {{ app.t('reports.receipts') }}</p>
              <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                <div><dt class="text-slate-500">{{ app.t('reports.revenue') }}</dt><dd class="font-bold">{{ money((row as SalesPeriodReport).revenue) }}</dd></div>
                <div><dt class="text-slate-500">{{ app.t('reports.profit') }}</dt><dd class="font-bold" :class="profitClass((row as SalesPeriodReport).profit)">{{ money((row as SalesPeriodReport).profit) }}</dd></div>
              </dl>
            </template>
            <template v-else-if="isProductReport">
              <div class="flex min-w-0 items-center gap-3">
                <ProductAvatar :src="(row as ProductSalesReport).image_url" :updated-at="(row as ProductSalesReport).image_updated_at" :name="(row as ProductSalesReport).product_name" size="sm" shape="square" />
                <div class="min-w-0">
                  <h3 class="truncate font-black">{{ (row as ProductSalesReport).product_name }}</h3>
                  <p class="text-sm text-slate-500 dark:text-slate-400">{{ (row as ProductSalesReport).sku }}</p>
                </div>
              </div>
              <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                <div><dt class="text-slate-500">{{ app.t('reports.soldQuantity') }}</dt><dd class="font-bold">{{ (row as ProductSalesReport).quantity.toLocaleString(locale) }}</dd></div>
                <div><dt class="text-slate-500">{{ app.t('reports.revenue') }}</dt><dd class="font-bold">{{ money((row as ProductSalesReport).revenue) }}</dd></div>
                <div v-if="activeTab === 'profit-by-product'"><dt class="text-slate-500">{{ app.t('reports.profit') }}</dt><dd class="font-bold" :class="profitClass((row as ProductSalesReport).profit)">{{ money((row as ProductSalesReport).profit) }}</dd></div>
              </dl>
            </template>
            <template v-else-if="isPaymentReport">
              <h3 class="font-black">{{ paymentLabel((row as PaymentSummaryReport).payment_method) }}</h3>
              <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                <div><dt class="text-slate-500">{{ app.t('reports.transactions') }}</dt><dd class="font-bold">{{ (row as PaymentSummaryReport).receipt_count.toLocaleString(locale) }}</dd></div>
                <div><dt class="text-slate-500">{{ app.t('reports.totalPayment') }}</dt><dd class="font-bold">{{ money((row as PaymentSummaryReport).revenue) }}</dd></div>
              </dl>
            </template>
            <template v-else>
              <div class="flex min-w-0 items-center gap-3">
                <ProductAvatar :src="(row as StockReport).image_url" :updated-at="(row as StockReport).image_updated_at" :name="(row as StockReport).product_name" size="sm" shape="square" />
                <div class="min-w-0">
                  <h3 class="truncate font-black">{{ (row as StockReport).product_name }}</h3>
                  <p class="text-sm text-slate-500 dark:text-slate-400">{{ (row as StockReport).sku }} · {{ (row as StockReport).location_name }}</p>
                </div>
              </div>
              <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                <div><dt class="text-slate-500">{{ app.t('reports.stockQuantity') }}</dt><dd class="font-bold">{{ (row as StockReport).quantity.toLocaleString(locale) }}</dd></div>
                <div v-if="activeTab === 'inventory-valuation'"><dt class="text-slate-500">{{ app.t('reports.totalValue') }}</dt><dd class="font-bold">{{ money((row as StockReport).total_value) }}</dd></div>
                <div v-else><dt class="text-slate-500">{{ app.t('reports.status') }}</dt><dd><AppBadge :tone="stockStatusTone((row as StockReport).stock_status)">{{ stockStatusLabel((row as StockReport).stock_status) }}</AppBadge></dd></div>
              </dl>
            </template>
          </article>
        </div>

        <div class="mt-4 flex flex-col gap-3 border-t border-slate-200 pt-4 text-sm dark:border-slate-800 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-slate-500 dark:text-slate-400">{{ app.t('reports.show') }}</span>
            <AppSelect :model-value="pageSize" hide-arrow @update:model-value="changePageSize">
              <option value="10">10</option>
              <option value="20">20</option>
              <option value="50">50</option>
            </AppSelect>
            <span class="text-slate-500 dark:text-slate-400">{{ app.t('reports.perPage') }}</span>
            <span class="text-slate-500 dark:text-slate-400">{{ app.t('reports.totalRows') }} {{ rows.length.toLocaleString(locale) }}</span>
          </div>
          <div class="flex items-center justify-end gap-2">
            <AppButton variant="secondary" :disabled="page <= 1" @click="previousPage">{{ app.t('reports.previous') }}</AppButton>
            <span class="font-bold text-slate-600 dark:text-slate-300">{{ t('reports.page', { page, total: totalPages }) }}</span>
            <AppButton variant="secondary" :disabled="page >= totalPages" @click="nextPage">{{ app.t('reports.next') }}</AppButton>
          </div>
        </div>
      </AppCard>
    </div>
  </section>
</template>
