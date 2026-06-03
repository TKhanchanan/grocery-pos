<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { ApexOptions } from 'apexcharts'
import VueApexCharts from 'vue3-apexcharts'
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
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import type { DashboardSummary, StockReport, StockStatus } from '../types/navigation'

type Period = '7D' | '30D' | 'MONTH'

const app = useAppStore()
const summary = ref<DashboardSummary | null>(null)
const loading = ref(false)
const error = ref('')
const period = ref<Period>('7D')

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
  return Math.max(5, Math.min(100, Math.round((item.quantity / target) * 100)))
}

function periodDays() {
  if (period.value === '7D') return 7
  if (period.value === '30D') return 30
  return new Date(new Date().getFullYear(), new Date().getMonth() + 1, 0).getDate()
}

function dateKey(date: Date) {
  return date.toISOString().slice(0, 10)
}

function shortDate(key: string) {
  return new Date(`${key}T00:00:00`).toLocaleDateString(locale.value, { day: '2-digit', month: 'short' })
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
  const days = periodDays()
  const today = new Date()
  const rows = new Map<string, { revenue: number; profit: number; receipts: number }>()
  for (let index = days - 1; index >= 0; index--) {
    const date = new Date(today)
    date.setDate(today.getDate() - index)
    rows.set(dateKey(date), { revenue: 0, profit: 0, receipts: 0 })
  }
  for (const sale of summary.value?.recent_sales ?? []) {
    const key = dateKey(new Date(sale.created_at))
    const row = rows.get(key)
    if (!row || sale.status === 'CANCELLED') continue
    row.revenue += sale.total_amount
    row.profit += sale.profit
    row.receipts += 1
  }
  return {
    labels: [...rows.keys()].map(shortDate),
    revenue: [...rows.values()].map((row) => Number(row.revenue.toFixed(2))),
    profit: [...rows.values()].map((row) => Number(row.profit.toFixed(2))),
    receipts: [...rows.values()].map((row) => row.receipts),
  }
})

const hasSalesTrend = computed(() => salesTrend.value.revenue.some((value) => value > 0) || salesTrend.value.profit.some((value) => value > 0))

const salesChartOptions = computed<ApexOptions>(() => ({
  chart: {
    type: 'area',
    toolbar: { show: false },
    animations: { enabled: true, speed: 650, dynamicAnimation: { enabled: true, speed: 450 } },
    foreColor: app.isDark ? '#cbd5e1' : '#475569',
  },
  colors: ['#16a34a', '#0ea5e9'],
  dataLabels: { enabled: false },
  stroke: { curve: 'smooth', width: 3 },
  fill: {
    type: 'gradient',
    gradient: { shadeIntensity: 0.8, opacityFrom: 0.35, opacityTo: 0.05, stops: [0, 80, 100] },
  },
  grid: { borderColor: app.isDark ? '#334155' : '#e2e8f0', strokeDashArray: 4 },
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
  colors: ['#16a34a'],
  plotOptions: { bar: { borderRadius: 7, columnWidth: '52%' } },
  tooltip: { y: { formatter: (value) => t('dashboard.chart.receiptTooltip', { count: Number(value).toLocaleString(locale.value) }) } },
}))

const receiptSeries = computed(() => [{ name: app.t('dashboard.chart.receipts'), data: salesTrend.value.receipts }])

const paymentChartOptions = computed<ApexOptions>(() => ({
  chart: { type: 'donut', animations: { enabled: true, speed: 700 } },
  labels: (summary.value?.payment_method_summary ?? []).map((item) => paymentMethodLabel(item.payment_method)),
  colors: ['#16a34a', '#0ea5e9', '#f59e0b', '#ef4444'],
  dataLabels: { enabled: false },
  legend: { show: false },
  stroke: { width: 0 },
  plotOptions: {
    pie: {
      donut: {
        size: '72%',
        labels: {
          show: true,
          total: {
            show: true,
            label: app.t('dashboard.chart.revenue'),
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
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('dashboard.error.load')
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboard)
</script>

<template>
  <section class="dashboard-page">
    <PageHeader :title="app.t('dashboard.title')" :eyebrow="app.t('dashboard.eyebrow')" :description="app.t('dashboard.description')" icon="layout-dashboard">
      <AppButton variant="secondary" icon="history" :loading="loading" @click="loadDashboard">{{ app.t('dashboard.refresh') }}</AppButton>
    </PageHeader>

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
      <div class="relative overflow-hidden rounded-3xl border border-brand-100 bg-gradient-to-br from-brand-600 via-emerald-600 to-sky-600 p-6 text-white shadow-2xl shadow-brand-950/20 dark:border-emerald-400/20 dark:from-emerald-950 dark:via-slate-900 dark:to-sky-950 dark:shadow-black/30">
        <div class="absolute -right-16 -top-20 h-56 w-56 rounded-full bg-white/15 blur-3xl" />
        <div class="absolute bottom-0 right-16 h-28 w-28 rounded-full bg-white/10 blur-2xl" />
        <div class="relative grid gap-5 xl:grid-cols-[minmax(0,1fr)_320px] xl:items-center">
          <div>
            <p class="text-sm font-black uppercase text-white/75">{{ app.t('dashboard.hero.eyebrow') }}</p>
            <h2 class="mt-2 max-w-2xl text-3xl font-black md:text-4xl">{{ app.t('dashboard.hero.title') }}</h2>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-white/80">{{ app.t('dashboard.hero.description') }}</p>
            <div class="mt-5 flex flex-wrap gap-2">
              <RouterLink to="/pos" class="focus-ring inline-flex min-h-11 items-center gap-2 rounded-xl bg-white px-4 text-sm font-black text-brand-700 shadow-lg dark:bg-emerald-200 dark:text-emerald-950">
                <AppIcon name="shopping-cart" :size="18" />{{ app.t('dashboard.hero.pos') }}
              </RouterLink>
              <RouterLink to="/reports" class="focus-ring inline-flex min-h-11 items-center gap-2 rounded-xl border border-white/30 px-4 text-sm font-black text-white hover:bg-white/10">
                <AppIcon name="chart-column" :size="18" />{{ app.t('dashboard.hero.reports') }}
              </RouterLink>
              <RouterLink to="/restock" class="focus-ring inline-flex min-h-11 items-center gap-2 rounded-xl border border-white/30 px-4 text-sm font-black text-white hover:bg-white/10">
                <AppIcon name="package-plus" :size="18" />{{ app.t('dashboard.hero.restock') }}
              </RouterLink>
            </div>
          </div>
          <div class="rounded-2xl border border-white/20 bg-white/10 p-4 backdrop-blur">
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
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.todaySales')" :value="summary.today_sales" :decimals="2" :locale="locale" :suffix="currencySuffix" :helper="app.t('dashboard.helper.revenueCompleted')" :trend="app.t('dashboard.trend.livePos')" icon="banknote" tone="success" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.monthProfit')" :value="summary.gross_profit_this_month" :decimals="2" :locale="locale" :suffix="currencySuffix" :helper="app.t('dashboard.helper.cancelledExcluded')" :trend="app.t('dashboard.trend.snapshot')" icon="chart-column" tone="info" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.topProduct')" :text-value="summary.top_product_this_month?.product_name ?? app.t('dashboard.empty.noTopProduct')" :helper="summary.top_product_this_month ? t('dashboard.helper.soldThisMonth', { quantity: summary.top_product_this_month.quantity }) : app.t('dashboard.helper.noSalesYet')" icon="package" tone="brand" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.pendingAlerts')" :value="alertTotal" :locale="locale" :helper="t('dashboard.helper.alertBreakdown', { out: summary.out_of_stock_count, low: summary.low_stock_count })" icon="bell" tone="warning" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.receiptsToday')" :value="summary.today_receipts" :locale="locale" :helper="app.t('dashboard.helper.receiptsCompleted')" icon="receipt-text" tone="success" />
        <DashboardKpiCard class="xl:col-span-2" :label="app.t('dashboard.kpi.reorderItems')" :value="reorderCount" :locale="locale" :helper="app.t('dashboard.helper.reorderPoint')" icon="clipboard-list" tone="danger" />
      </div>

      <div class="mt-5 grid gap-4 xl:grid-cols-[minmax(0,1fr)_380px]">
        <AppCard class="dashboard-section">
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
          <div v-if="!hasSalesTrend" class="mt-2">
            <AppEmptyState :title="app.t('dashboard.empty.noTrendTitle')" :description="app.t('dashboard.empty.noTrendDescription')" />
          </div>
        </AppCard>

        <AppCard class="dashboard-section">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.payment.eyebrow') }}</p>
              <h2 class="mt-1 text-xl font-black">{{ app.t('dashboard.payment.title') }}</h2>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('dashboard.payment.summary', { count: paymentReceiptCount, amount: money(paymentTotal) }) }}</p>
            </div>
            <AppBadge tone="info">{{ app.t('dashboard.payment.live') }}</AppBadge>
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
        <AppCard class="dashboard-section">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.product.eyebrow') }}</p>
              <h2 class="mt-1 text-xl font-black">{{ app.t('dashboard.product.title') }}</h2>
            </div>
            <RouterLink to="/reports" class="text-sm font-black text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.product.viewReport') }}</RouterLink>
          </div>
          <AppEmptyState v-if="summary.top_products.length === 0" class="mt-4" :title="app.t('dashboard.empty.noProductSalesTitle')" :description="app.t('dashboard.empty.noProductSalesDescription')" />
          <div v-else class="mt-5 grid gap-3">
            <article v-for="(product, index) in summary.top_products.slice(0, 6)" :key="product.product_id" class="rounded-2xl border border-slate-200 bg-white/70 p-4 dark:border-slate-700 dark:bg-slate-950/50">
              <div class="flex items-center justify-between gap-3">
                <div class="flex min-w-0 items-center gap-3">
                  <span class="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-100 text-sm font-black text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-200">#{{ index + 1 }}</span>
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
                <div class="h-full rounded-full bg-gradient-to-r from-brand-500 to-sky-400 transition-all duration-700" :style="{ width: `${Math.max(8, Math.round((product.quantity / topProductMax) * 100))}%` }" />
              </div>
            </article>
          </div>
        </AppCard>

        <AppCard class="dashboard-section">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.inventory.eyebrow') }}</p>
              <h2 class="mt-1 text-xl font-black">{{ app.t('dashboard.inventory.title') }}</h2>
            </div>
            <RouterLink to="/alerts" class="text-sm font-black text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.inventory.openAlerts') }}</RouterLink>
          </div>
          <AppEmptyState v-if="stockRiskItems.length === 0" class="mt-4" :title="app.t('dashboard.empty.stockHealthyTitle')" :description="app.t('dashboard.empty.stockHealthyDescription')" />
          <div v-else class="mt-5 grid gap-3">
            <article v-for="item in stockRiskItems" :key="`${item.product_id}-${item.location_id}`" class="rounded-2xl border border-slate-200 bg-white/70 p-4 dark:border-slate-700 dark:bg-slate-950/50">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <h3 class="truncate font-black">{{ item.product_name }}</h3>
                  <p class="text-sm text-slate-500 dark:text-slate-400">{{ item.location_name }} · {{ item.sku }}</p>
                </div>
                <AppBadge :tone="stockTone(item.stock_status)">{{ stockStatusLabel(item.stock_status) }}</AppBadge>
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

      <div class="mt-5 grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <AppCard class="dashboard-section">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.activity.eyebrow') }}</p>
              <h2 class="mt-1 text-xl font-black">{{ app.t('dashboard.activity.title') }}</h2>
            </div>
            <RouterLink to="/sales-history" class="text-sm font-black text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.activity.viewAll') }}</RouterLink>
          </div>
          <AppEmptyState v-if="recentSales.length === 0" class="mt-4" :title="app.t('dashboard.empty.noSalesTitle')" :description="app.t('dashboard.empty.noSalesDescription')" />
          <div v-else class="mt-5 grid gap-3">
            <article v-for="sale in recentSales" :key="sale.id" class="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white/70 p-4 sm:flex-row sm:items-center sm:justify-between dark:border-slate-700 dark:bg-slate-950/50">
              <div class="flex min-w-0 items-center gap-3">
                <span class="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-200"><AppIcon name="receipt-text" /></span>
                <div class="min-w-0">
                  <p class="truncate font-black">{{ sale.receipt_no }}</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">{{ sale.cashier_name }} · {{ new Date(sale.created_at).toLocaleString(locale) }}</p>
                </div>
              </div>
              <div class="flex items-center justify-between gap-3 sm:block sm:text-right">
                <AppBadge :tone="sale.payment_method === 'CASH' ? 'success' : 'info'">{{ paymentMethodLabel(sale.payment_method) }}</AppBadge>
                <p class="font-black">{{ money(sale.total_amount) }}</p>
              </div>
            </article>
          </div>
        </AppCard>

        <AppCard class="dashboard-section">
          <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('dashboard.insight.eyebrow') }}</p>
          <h2 class="mt-1 text-xl font-black">{{ app.t('dashboard.insight.title') }}</h2>
          <div class="mt-5 grid gap-3">
            <RouterLink to="/alerts" class="rounded-2xl border border-amber-200 bg-amber-50 p-4 transition hover:shadow-md dark:border-amber-400/30 dark:bg-amber-500/10">
              <div class="flex items-center justify-between gap-3">
                <span class="font-black text-amber-900 dark:text-amber-100">{{ app.t('dashboard.insight.openAlerts') }}</span>
                <span class="text-2xl font-black text-amber-800 dark:text-amber-200">{{ alertTotal }}</span>
              </div>
              <p class="mt-1 text-sm text-amber-800 dark:text-amber-200">{{ app.t('dashboard.insight.openAlertsDescription') }}</p>
            </RouterLink>
            <RouterLink to="/purchase-orders" class="rounded-2xl border border-sky-200 bg-sky-50 p-4 transition hover:shadow-md dark:border-sky-400/30 dark:bg-sky-500/10">
              <div class="flex items-center justify-between gap-3">
                <span class="font-black text-sky-900 dark:text-sky-100">{{ app.t('dashboard.insight.purchasePlanning') }}</span>
                <span class="text-2xl font-black text-sky-800 dark:text-sky-200">{{ reorderCount }}</span>
              </div>
              <p class="mt-1 text-sm text-sky-800 dark:text-sky-200">{{ app.t('dashboard.insight.purchasePlanningDescription') }}</p>
            </RouterLink>
            <RouterLink to="/exports" class="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 transition hover:shadow-md dark:border-emerald-400/30 dark:bg-emerald-500/10">
              <div class="flex items-center justify-between gap-3">
                <span class="font-black text-emerald-900 dark:text-emerald-100">{{ app.t('dashboard.insight.readyExport') }}</span>
                <AppIcon name="download" class="text-emerald-700 dark:text-emerald-200" />
              </div>
              <p class="mt-1 text-sm text-emerald-800 dark:text-emerald-200">{{ app.t('dashboard.insight.readyExportDescription') }}</p>
            </RouterLink>
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
