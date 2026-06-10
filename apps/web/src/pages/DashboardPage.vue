<script setup lang="ts">
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, ref } from 'vue'
import type { ApexOptions } from 'apexcharts'
import { RouterLink } from 'vue-router'
import { apiClient } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppIcon from '../components/AppIcon.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import DashboardKpiCard from '../components/DashboardKpiCard.vue'
import PageHeader from '../components/PageHeader.vue'
import ProductAvatar from '../components/ProductAvatar.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { DashboardSummary, StockReport, StockStatus } from '../types/navigation'
import { appLocale, formatAppDateTime } from '../utils/date'

type Period = '7D' | '30D' | 'MONTH'

const VueApexCharts = defineAsyncComponent(() => import('vue3-apexcharts'))
const app = useAppStore()
const auth = useAuthStore()
const summary = ref<DashboardSummary | null>(null)
const loading = ref(false)
const error = ref('')
const period = ref<Period>('7D')
let liveRefreshTimer: number | undefined

const periods: Array<{ value: Period; labelKey: TranslationKey }> = [
  { value: '7D', labelKey: 'dashboard.period.7d' },
  { value: '30D', labelKey: 'dashboard.period.30d' },
  { value: 'MONTH', labelKey: 'dashboard.period.month' },
]

const alertTotal = computed(() => (summary.value?.low_stock_count ?? 0) + (summary.value?.out_of_stock_count ?? 0) + (summary.value?.reorder_count ?? 0))
const reorderCount = computed(() => summary.value?.reorder_count ?? 0)
const paymentTotal = computed(() => (summary.value?.payment_method_summary ?? []).reduce((sum, item) => sum + item.revenue, 0))
const paymentReceiptCount = computed(() => (summary.value?.payment_method_summary ?? []).reduce((sum, item) => sum + item.receipt_count, 0))
const topProductMax = computed(() => Math.max(...(summary.value?.top_products ?? []).map((item) => item.quantity), 1))
const stockRiskItems = computed(() => (summary.value?.low_stock_items ?? []).slice(0, 6))
const recentSales = computed(() => (summary.value?.recent_sales ?? []).slice(0, 6))
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')
const currencySuffix = computed(() => ` ${app.t('dashboard.currency.bahtAmount').replace('{amount}', '').trim()}`)
const canViewPOS = computed(() => auth.hasPermission('pos.view'))
const canViewReports = computed(() => auth.hasPermission('reports.view'))
const canRestock = computed(() => auth.hasPermission('stock.restock'))
const canViewAlerts = computed(() => auth.hasPermission('alerts.view'))
const canViewSalesHistory = computed(() => auth.hasPermission('sales.view'))

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) {
    text = text.replaceAll(`{${name}}`, String(value))
  }
  return text
}

function moneyAmount(value: number) {
  return value.toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function money(value: number) {
  return t('dashboard.currency.bahtAmount', { amount: moneyAmount(value) })
}

function stockClass(status: StockStatus) {
  return {
    in_stock: 'bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-200',
    low_stock: 'bg-amber-100 text-amber-800 dark:bg-amber-500/20 dark:text-amber-200',
    out_of_stock: 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-200',
    reorder_point: 'bg-blue-100 text-blue-700 dark:bg-sky-500/20 dark:text-sky-200',
  }[status]
}

function stockTone(status: StockStatus) {
  return {
    in_stock: 'success',
    low_stock: 'warning',
    out_of_stock: 'danger',
    reorder_point: 'info',
  }[status] as 'success' | 'warning' | 'danger' | 'info'
}

function stockProgress(item: StockReport) {
  const target = Math.max(item.threshold, item.reorder_point, item.quantity, 1)
  return Math.max(item.quantity > 0 ? 8 : 2, Math.min(100, Math.round((item.quantity / target) * 100)))
}

function topProductProgress(quantity: number) {
  if (topProductMax.value <= 0 || quantity <= 0) return 2
  return Math.max(10, Math.min(100, Math.round((quantity / topProductMax.value) * 100)))
}

function periodDays() {
  if (period.value === '7D') return 7
  if (period.value === '30D') return 30
  return new Date(new Date().getFullYear(), new Date().getMonth() + 1, 0).getDate()
}

function dateKey(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function normalizeDateKey(value: string) {
  const match = value.match(/^(\d{4})-(\d{2})-(\d{2})/)
  if (match) return `${match[1]}-${match[2]}-${match[3]}`
  return dateKey(new Date(value))
}

function shortDate(key: string) {
  const [year, month, day] = key.split('-').map(Number)
  return new Intl.DateTimeFormat(appLocale(app.language), { day: '2-digit', month: 'short' }).format(new Date(year, month - 1, day))
}

function trendStartDate(today: Date) {
  if (period.value === 'MONTH') return new Date(today.getFullYear(), today.getMonth(), 1)
  const date = new Date(today)
  date.setDate(today.getDate() - (periodDays() - 1))
  return date
}

function trendEndDate(today: Date) {
  if (period.value === 'MONTH') return new Date(today.getFullYear(), today.getMonth() + 1, 0)
  return today
}

function stockStatusLabel(status: StockStatus) {
  const key = `dashboard.stockStatus.${status}` as TranslationKey
  return app.t(key)
}

function paymentMethodLabel(method: string) {
  const labels: Record<string, TranslationKey> = {
    CASH: 'dashboard.payment.cash',
    QR: 'dashboard.payment.qr',
  }
  const key = labels[method]
  return key ? app.t(key) : method
}

const salesTrend = computed(() => {
  const today = new Date()
  const rows = new Map<string, { revenue: number; profit: number; receipts: number }>()
  const start = trendStartDate(today)
  const end = trendEndDate(today)
  for (const date = new Date(start); date <= end; date.setDate(date.getDate() + 1)) {
    rows.set(dateKey(date), { revenue: 0, profit: 0, receipts: 0 })
  }
  for (const report of summary.value?.sales_trend ?? []) {
    const key = normalizeDateKey(report.period)
    const row = rows.get(key)
    if (!row) continue
    row.revenue += report.revenue
    row.profit += report.profit
    row.receipts += report.receipt_count
  }
  return {
    labels: [...rows.keys()].map(shortDate),
    revenue: [...rows.values()].map((row) => Number(row.revenue.toFixed(2))),
    profit: [...rows.values()].map((row) => Number(row.profit.toFixed(2))),
    receipts: [...rows.values()].map((row) => row.receipts),
  }
})

const salesChartOptions = computed<ApexOptions>(() => ({
  chart: {
    type: 'area',
    toolbar: { show: false },
    animations: { enabled: true, speed: 650, dynamicAnimation: { enabled: true, speed: 450 } },
    foreColor: app.isDark ? '#cbd5e1' : '#475569',
  },
  colors: app.isDark ? ['#5eead4', '#fbbf24'] : ['#0f766e', '#ea580c'],
  dataLabels: { enabled: false },
  stroke: { curve: 'smooth', width: 3 },
  fill: {
    type: 'gradient',
    gradient: { shadeIntensity: 0.8, opacityFrom: 0.35, opacityTo: 0.05, stops: [0, 80, 100] },
  },
  grid: { borderColor: app.isDark ? '#475569' : '#cbd5e1', strokeDashArray: 4 },
  legend: { position: 'top', horizontalAlign: 'right', fontWeight: 700 },
  xaxis: { categories: salesTrend.value.labels, axisBorder: { show: false }, axisTicks: { show: false } },
  yaxis: {
    labels: { formatter: (value) => `${Math.round(value).toLocaleString(locale.value)}` },
  },
  tooltip: {
    y: { formatter: (value) => money(Number(value)) },
  },
}))

const salesSeries = computed(() => [
  { name: app.t('dashboard.chart.revenue'), data: salesTrend.value.revenue },
  { name: app.t('dashboard.chart.profit'), data: salesTrend.value.profit },
])

const receiptChartOptions = computed<ApexOptions>(() => ({
  chart: { type: 'bar', toolbar: { show: false }, sparkline: { enabled: true }, animations: { enabled: true, speed: 700 } },
  colors: [app.isDark ? '#5eead4' : '#5eead4'],
  plotOptions: { bar: { borderRadius: 7, columnWidth: '52%' } },
  tooltip: { y: { formatter: (value) => t('dashboard.chart.receiptTooltip', { count: Number(value).toLocaleString(locale.value) }) } },
}))

const receiptSeries = computed(() => [{ name: app.t('dashboard.chart.receipts'), data: salesTrend.value.receipts }])

const paymentChartOptions = computed<ApexOptions>(() => ({
  chart: { type: 'donut', animations: { enabled: true, speed: 700 } },
  labels: (summary.value?.payment_method_summary ?? []).map((item) => paymentMethodLabel(item.payment_method)),
  colors: app.isDark ? ['#5eead4', '#fbbf24', '#93c5fd', '#fb7185'] : ['#0f766e', '#ea580c', '#2563eb', '#dc2626'],
  dataLabels: { enabled: false },
  legend: { show: false },
  stroke: { width: 0 },
  plotOptions: {
    pie: {
      donut: {
        size: '72%',
        labels: {
          show: true,
          name: {
            color: app.isDark ? '#cbd5e1' : '#475569',
            fontWeight: 800,
          },
          value: {
            color: app.isDark ? '#f8fafc' : '#0f172a',
            fontWeight: 900,
          },
          total: {
            show: true,
            label: app.t('dashboard.chart.revenue'),
            color: app.isDark ? '#cbd5e1' : '#475569',
            fontWeight: 800,
            formatter: () => moneyAmount(paymentTotal.value),
          },
        },
      },
    },
  },
  tooltip: { y: { formatter: (value) => money(Number(value)) } },
}))

const paymentSeries = computed(() => (summary.value?.payment_method_summary ?? []).map((item) => item.revenue))

async function loadDashboard() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await apiClient<DashboardSummary>('/v1/dashboard/summary')
    app.alertCount = summary.value.unread_alert_count
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('dashboard.error.load')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
  liveRefreshTimer = window.setInterval(loadDashboard, 15000)
})

onBeforeUnmount(() => {
  if (liveRefreshTimer) window.clearInterval(liveRefreshTimer)
})
</script>

<template>
  <section class="dashboard-page">
    <PageHeader :title="app.t('dashboard.title')" :eyebrow="app.t('dashboard.eyebrow')" :description="app.t('dashboard.description')" icon="layout-dashboard" />

    <div v-if="error" class="mb-4 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <span>{{ error }}</span>
        <AppButton variant="secondary" @click="loadDashboard">{{ app.t('dashboard.tryAgain') }}</AppButton>
      </div>
    </div>

    <div v-if="loading && !summary" class="grid gap-4">
      <AppLoadingState :label="app.t('dashboard.loadingAnalytics')" />
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div v-for="item in 4" :key="item" class="h-40 animate-pulse rounded-2xl border border-slate-200 bg-white/70 dark:border-slate-700 dark:bg-slate-900/70" />
      </div>
      <div class="h-80 animate-pulse rounded-2xl border border-slate-200 bg-white/70 dark:border-slate-700 dark:bg-slate-900/70" />
    </div>

    <template v-if="summary">
      <div
        class="relative overflow-hidden rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(204,251,241,0.35),transparent_26rem),linear-gradient(135deg,#009a9a_0%,#087f83_48%,#104145_100%)] p-6 text-white shadow-2xl shadow-brand-900/20 dark:bg-[linear-gradient(135deg,#0f172a_0%,#115e59_48%,#312e81_100%)] dark:shadow-black/30">
        <div class="relative grid gap-5 xl:grid-cols-[minmax(0,1fr)_320px] xl:items-center">
          <div>
            <p class="text-sm font-black uppercase text-white/75">{{ app.t('dashboard.hero.eyebrow') }}</p>
            <h2 class="mt-2 max-w-2xl text-3xl font-black md:text-4xl">{{ app.t('dashboard.hero.title') }}</h2>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-white/80">{{ app.t('dashboard.hero.description') }}</p>
            <div class="mt-5 flex flex-wrap gap-2">
              <RouterLink v-if="canViewPOS" to="/pos"
                class="focus-ring inline-flex min-h-11 items-center gap-2 rounded-xl bg-white px-4 text-sm font-black text-brand-700 shadow-lg dark:bg-emerald-200 dark:text-emerald-950">
                <AppIcon name="shopping-cart" :size="18" />{{ app.t('dashboard.hero.pos') }}
              </RouterLink>
              <RouterLink v-if="canViewReports" to="/reports"
                class="focus-ring inline-flex min-h-11 items-center gap-2 rounded-xl border border-white/30 px-4 text-sm font-black text-white hover:bg-white/10">
                <AppIcon name="chart-column" :size="18" />{{ app.t('dashboard.hero.reports') }}
              </RouterLink>
              <RouterLink v-if="canRestock" to="/restock"
                class="focus-ring inline-flex min-h-11 items-center gap-2 rounded-xl border border-white/30 px-4 text-sm font-black text-white hover:bg-white/10">
                <AppIcon name="package-plus" :size="18" />{{ app.t('dashboard.hero.restock') }}
              </RouterLink>
            </div>
          </div>
          <div class="rounded-2xl bg-white/10 p-4 shadow-xl shadow-slate-950/10 backdrop-blur">
            <div class="flex items-center justify-between">
              <span class="text-sm font-bold text-white/75">{{ app.t('dashboard.todayReceipts') }}</span>
              <AppIcon name="receipt-text" />
            </div>
            <p class="mt-3 text-4xl font-black">{{ summary.today_receipts }}</p>
            <VueApexCharts class="mt-2" height="72" type="bar" :options="receiptChartOptions" :series="receiptSeries" />
          </div>
        </div>
      </div>

      <div class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-6">
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.todaySales')" :value="summary.today_sales" :decimals="2" :locale="locale" :suffix="currencySuffix" :helper="app.t('dashboard.helper.revenueCompleted')" icon="banknote" tone="success" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.monthProfit')" :value="summary.gross_profit_this_month" :decimals="2" :locale="locale" :suffix="currencySuffix" :helper="app.t('dashboard.helper.cancelledExcluded')" icon="chart-column" tone="info" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.topProduct')" :text-value="summary.top_product_this_month?.product_name ?? app.t('dashboard.empty.noTopProduct')" :helper="summary.top_product_this_month ? t('dashboard.helper.soldThisMonth', { quantity: summary.top_product_this_month.quantity }) : app.t('dashboard.helper.noSalesYet')" icon="package" tone="brand" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.pendingAlerts')" :value="alertTotal" :locale="locale" :helper="t('dashboard.helper.alertBreakdown', { out: summary.out_of_stock_count, low: summary.low_stock_count })" icon="bell" tone="warning" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.receiptsToday')" :value="summary.today_receipts" :locale="locale" :helper="app.t('dashboard.helper.receiptsCompleted')" icon="receipt-text" tone="success" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.reorderItems')" :value="reorderCount" :locale="locale" :helper="app.t('dashboard.helper.reorderPoint')" icon="clipboard-list" tone="danger" />
      </div>

      <div class="mt-5 grid gap-4 xl:grid-cols-[minmax(0,1fr)_380px]">
        <AppCard class="dashboard-section dark:bg-slate-900/80">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.chart.performanceEyebrow') }}</p>
              <h2 class="mt-1 text-xl font-black">{{ app.t('dashboard.chart.salesProfitTitle') }}</h2>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ app.t('dashboard.chart.salesProfitDescription') }}</p>
            </div>
            <div class="flex rounded-xl border border-slate-200 bg-white/75 p-1 dark:border-slate-700 dark:bg-slate-950/60">
              <button
                v-for="item in periods"
                :key="item.value"
                class="rounded-lg px-3 py-1.5 text-xs font-black transition"
                :class="period === item.value ? 'bg-brand-600 text-white shadow-sm' : 'text-slate-600 hover:bg-brand-50 dark:text-slate-300 dark:hover:bg-slate-800'"
                @click="period = item.value"
              >
                {{ app.t(item.labelKey) }}
              </button>
            </div>
          </div>
          <div class="mt-5 min-h-[320px]">
            <VueApexCharts height="320" type="area" :options="salesChartOptions" :series="salesSeries" />
          </div>
        </AppCard>

        <AppCard class="dashboard-section dark:bg-slate-900/80">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.payment.eyebrow') }}</p>
              <h2 class="mt-1 text-xl font-black">{{ app.t('dashboard.payment.title') }}</h2>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('dashboard.payment.summary', { count: paymentReceiptCount, amount: money(paymentTotal) }) }}</p>
            </div>
            <AppBadge tone="success"><span class="mr-1.5 h-2 w-2 rounded-full bg-current animate-pulse" />{{ app.t('dashboard.payment.live') }}</AppBadge>
          </div>
          <div v-if="paymentSeries.length" class="mt-5">
            <VueApexCharts height="270" type="donut" :options="paymentChartOptions" :series="paymentSeries" />
            <div class="mt-4 grid gap-2">
              <div v-for="payment in summary.payment_method_summary" :key="payment.payment_method" class="flex items-center justify-between rounded-xl bg-slate-50 p-3 text-sm dark:bg-slate-950/70">
                <span class="font-black">{{ paymentMethodLabel(payment.payment_method) }}</span>
                <span class="text-slate-500 dark:text-slate-400">{{ t('dashboard.payment.rowSummary', { count: payment.receipt_count, amount: money(payment.revenue) }) }}</span>
              </div>
            </div>
          </div>
          <AppEmptyState v-else class="mt-5" :title="app.t('dashboard.empty.noPaymentTitle')" :description="app.t('dashboard.empty.noPaymentDescription')" />
        </AppCard>
      </div>

      <div class="mt-5 grid gap-4 xl:grid-cols-[minmax(0,1fr)_420px]">
        <AppCard class="dashboard-section dark:bg-slate-900/80">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <!-- <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.product.eyebrow') }}</p> -->
              <h2 class="text-xl font-black">{{ app.t('dashboard.product.title') }}</h2>
            </div>
            <RouterLink v-if="canViewReports" to="/reports" class="text-sm font-black text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.product.viewReport') }}</RouterLink>
          </div>
          <AppEmptyState v-if="summary.top_products.length === 0" class="mt-4" :title="app.t('dashboard.empty.noProductSalesTitle')" :description="app.t('dashboard.empty.noProductSalesDescription')" />
          <div v-else class="mt-5 grid gap-3">
            <article v-for="(product, index) in summary.top_products.slice(0, 6)" :key="product.product_id" class="nested-border-card rounded-2xl border border-slate-200 bg-white/70 p-4 dark:border-slate-700 dark:bg-slate-950/50">
              <div class="flex items-center justify-between gap-3">
                <div class="flex min-w-0 items-center gap-3">
                  <span class="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-100 text-sm font-black text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-200">#{{ index + 1 }}</span>
                  <ProductAvatar :src="product.image_url" :updated-at="product.image_updated_at" :name="product.product_name" size="sm" shape="square" />
                  <div class="min-w-0">
                    <h3 class="truncate font-black">{{ product.product_name }}</h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400">{{ product.sku }}</p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="font-black">{{ t('dashboard.product.sold', { quantity: product.quantity }) }}</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">{{ money(product.revenue) }}</p>
                </div>
              </div>
              <div class="mt-3 h-2 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
                <div class="h-full rounded-full bg-gradient-to-r from-teal-500 via-emerald-400 to-amber-400 transition-all duration-700 dark:from-teal-300 dark:via-emerald-300 dark:to-amber-300" :style="{ width: `${topProductProgress(product.quantity)}%` }" />
              </div>
            </article>
          </div>
        </AppCard>

        <AppCard class="dashboard-section dark:bg-slate-900/80">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <!-- <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.inventory.eyebrow') }}</p> -->
              <h2 class="text-xl font-black">{{ app.t('dashboard.inventory.title') }}</h2>
            </div>
            <RouterLink v-if="canViewAlerts" to="/alerts" class="text-sm font-black text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.inventory.openAlerts') }}</RouterLink>
          </div>
          <AppEmptyState v-if="stockRiskItems.length === 0" class="mt-4" :title="app.t('dashboard.empty.stockHealthyTitle')" :description="app.t('dashboard.empty.stockHealthyDescription')" />
          <div v-else class="mt-5 grid gap-3">
            <article v-for="item in stockRiskItems" :key="`${item.product_id}-${item.location_id}`" class="nested-border-card overflow-hidden rounded-2xl border p-4">
              <div class="flex min-w-0 items-start justify-between gap-3">
                <div class="flex min-w-0 flex-1 items-start gap-3 overflow-hidden">
                  <ProductAvatar class="shrink-0" :src="item.image_url" :updated-at="item.image_updated_at"
                    :name="item.product_name" size="sm" shape="square" />

                  <div class="min-w-0 flex-1 overflow-hidden">
                    <h3 class="truncate font-black" :title="item.product_name">
                      {{ item.product_name }}
                    </h3>

                    <p class="truncate text-sm text-slate-500 dark:text-slate-400"
                      :title="`${item.location_name} · ${item.sku}`">
                      {{ item.location_name }} · {{ item.sku }}
                    </p>
                  </div>
                </div>

                <AppBadge class="shrink-0 whitespace-nowrap" :tone="stockTone(item.stock_status)">
                  {{ stockStatusLabel(item.stock_status) }}
                </AppBadge>
              </div>
              <div class="mt-4 flex items-end justify-between gap-3">
                <div>
                  <p class="text-xs text-slate-500 dark:text-slate-400">{{ app.t('dashboard.inventory.currentStock') }}</p>
                  <p class="text-3xl font-black" :class="item.quantity === 0 ? 'text-red-700 dark:text-red-300' : 'text-slate-950 dark:text-slate-50'">{{ item.quantity }}</p>
                </div>
                <p class="whitespace-pre-line text-right text-xs text-slate-500 dark:text-slate-400">{{ t('dashboard.inventory.thresholdReorder', { threshold: item.threshold, reorder: item.reorder_point }) }}</p>
              </div>
              <div class="mt-3 h-2 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
                <div class="h-full rounded-full transition-all duration-700" :class="stockClass(item.stock_status)" :style="{ width: `${stockProgress(item)}%` }" />
              </div>
            </article>
          </div>
        </AppCard>
      </div>

      <div class="mt-5 grid gap-4">
        <AppCard class="dashboard-section dark:bg-slate-900/80">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <!-- <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.activity.eyebrow') }}</p> -->
              <h2 class="text-xl font-black">{{ app.t('dashboard.activity.title') }}</h2>
            </div>
            <RouterLink v-if="canViewSalesHistory" to="/sales-history" class="text-sm font-black text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.activity.viewAll') }}</RouterLink>
          </div>
          <AppEmptyState v-if="recentSales.length === 0" class="mt-4" :title="app.t('dashboard.empty.noSalesTitle')" :description="app.t('dashboard.empty.noSalesDescription')" />
          <div v-else class="mt-5 grid gap-3">
            <article v-for="sale in recentSales" :key="sale.id" class="nested-border-card flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white/70 p-4 sm:flex-row sm:items-center sm:justify-between dark:border-slate-700 dark:bg-slate-950/50">
              <div class="flex min-w-0 items-center gap-3">
                <span class="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-200"><AppIcon name="receipt-text" /></span>
                <div class="min-w-0">
                  <p class="truncate font-black">{{ sale.receipt_no }}</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">{{ sale.cashier_name }} · {{ formatAppDateTime(sale.created_at, app.language) }}</p>
                </div>
              </div>
              <div class="flex items-center justify-between gap-3 sm:block sm:text-right">
                <AppBadge :tone="sale.payment_method === 'CASH' ? 'success' : 'info'">{{ paymentMethodLabel(sale.payment_method) }}</AppBadge>
                <p class="font-black">{{ money(sale.total_amount) }}</p>
              </div>
            </article>
          </div>
        </AppCard>
      </div>
    </template>
  </section>
</template>

<style scoped>
.dashboard-page {
  animation: dashboard-page-in 420ms ease both;
}

.dashboard-section {
  animation: dashboard-card-in 520ms ease both;
}

@keyframes dashboard-page-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes dashboard-card-in {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
