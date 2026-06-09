<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { apiClient, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppDateRangeFilter from '../components/AppDateRangeFilter.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppIcon from '../components/AppIcon.vue'
import AppInput from '../components/AppInput.vue'
import AppPageSizeSelect from '../components/AppPageSizeSelect.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTextarea from '../components/AppTextarea.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { Location, Receipt } from '../types/navigation'
import { formatAppDateTime } from '../utils/date'

const app = useAppStore()
const auth = useAuthStore()
const sales = ref<Receipt[]>([])
const locations = ref<Location[]>([])
const loading = ref(false)
const error = ref('')
const cancelTarget = ref<Receipt | null>(null)
const cancelReason = ref('')
const page = ref(1)
const pageSize = ref(10)

const filters = reactive({
  date_from: '',
  date_to: '',
  cashier_id: '',
  location_id: '',
  payment_method: '',
  status: '',
  receipt_no: '',
})

const canCancel = computed(() => auth.user?.role === 'ADMIN' || auth.user?.role === 'MANAGER')
const canViewReceipt = computed(() => auth.hasPermission('sales.receipt.view'))
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')
const totals = computed(() => {
  const completed = sales.value.filter((sale) => sale.status === 'COMPLETED')
  return {
    completedCount: completed.length,
    cancelledCount: sales.value.filter((sale) => sale.status === 'CANCELLED').length,
    completedTotal: completed.reduce((sum, sale) => sum + sale.total_amount, 0),
  }
})
const totalPages = computed(() => Math.max(1, Math.ceil(sales.value.length / pageSize.value)))
const visibleSales = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return sales.value.slice(start, start + pageSize.value)
})

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function money(value: number) {
  return value.toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function moneyWithCurrency(value: number) {
  return t('sales.currency', { amount: money(value) })
}

function formatDate(value: string) {
  return formatAppDateTime(value, app.language)
}

function statusClass(status: Receipt['status']) {
  return status === 'CANCELLED'
    ? 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-100'
    : 'bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-100'
}

function statusLabel(status: Receipt['status']) {
  return app.t(status === 'CANCELLED' ? 'sales.status.cancelled' : 'sales.status.completed')
}

function paymentLabel(method: Receipt['payment_method']) {
  return method === 'QR' ? app.t('sales.payment.qr') : app.t('sales.payment.cash')
}

function buildQuery() {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filters)) {
    if (String(value).trim()) params.set(key, String(value).trim())
  }
  return params.toString()
}

async function loadSales() {
  loading.value = true
  error.value = ''
  try {
    const query = buildQuery()
    sales.value = await apiClient<Receipt[]>(`/v1/sales${query ? `?${query}` : ''}`)
    page.value = 1
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('sales.loadFailed')
  } finally {
    loading.value = false
  }
}

async function loadLocations() {
  locations.value = await apiClient<Location[]>('/v1/locations')
}

function resetFilters() {
  filters.date_from = ''
  filters.date_to = ''
  filters.cashier_id = ''
  filters.location_id = ''
  filters.payment_method = ''
  filters.status = ''
  filters.receipt_no = ''
  loadSales()
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
}

function previousPage() {
  if (page.value > 1) page.value -= 1
}

function nextPage() {
  if (page.value < totalPages.value) page.value += 1
}

function openCancel(sale: Receipt) {
  cancelTarget.value = sale
  cancelReason.value = ''
}

function receiptRoute(sale: Receipt) {
  return `/sales/${sale.id}/receipt`
}

async function cancelSale() {
  if (!cancelTarget.value) return
  error.value = ''
  try {
    await postJSON<Receipt>(`/v1/sales/${cancelTarget.value.id}/cancel`, { reason: cancelReason.value })
    cancelTarget.value = null
    cancelReason.value = ''
    await loadSales()
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('sales.cancelFailed')
  }
}

onMounted(async () => {
  await Promise.all([loadLocations(), loadSales()])
})
</script>

<template>
  <section class="min-w-0 max-w-full">
    <PageHeader :title="app.t('sales.title')" :eyebrow="app.t('sales.eyebrow')" :description="app.t('sales.description')" icon="purchase-order" />

    <div class="grid min-w-0 max-w-full gap-4">
      <div class="grid gap-3 md:grid-cols-3">
        <StatCard :label="app.t('sales.completedSales')" :value="totals.completedCount" :helper="app.t('sales.completedHelper')" icon="receipt-text" />
        <StatCard :label="app.t('sales.completedTotal')" :value="moneyWithCurrency(totals.completedTotal)" :helper="app.t('sales.totalHelper')" icon="banknote" tone="success" />
        <StatCard :label="app.t('sales.cancelled')" :value="totals.cancelledCount" :helper="app.t('sales.cancelledHelper')" icon="circle-x" tone="danger" />
      </div>

      <AppCard class="relative z-20 min-w-0 max-w-full overflow-visible dark:bg-slate-900/80">
        <div class="grid min-w-0 max-w-full gap-3 lg:grid-cols-2 2xl:grid-cols-4">
          <div class="grid min-w-0 max-w-full gap-3 sm:grid-cols-2 lg:col-span-2 2xl:col-span-4 2xl:grid-cols-4">
            <AppInput v-model="filters.receipt_no" :label="app.t('sales.receiptNo')" :placeholder="app.t('sales.receiptPlaceholder')" />
            <!-- <AppInput v-model="filters.cashier_id" :label="app.t('sales.cashierId')" :placeholder="app.t('sales.cashierPlaceholder')" /> -->
            <AppSelect v-model="filters.location_id" :label="app.t('sales.location')">
              <option value="">{{ app.t('sales.locationPlaceholder') }}</option>
              <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
            </AppSelect>
            <AppSelect v-model="filters.payment_method" :label="app.t('sales.payment')">
              <option value="">{{ app.t('sales.paymentPlaceholder') }}</option>
              <option value="CASH">{{ app.t('sales.payment.cash') }}</option>
              <option value="QR">{{ app.t('sales.payment.qr') }}</option>
            </AppSelect>
            <AppSelect v-model="filters.status" :label="app.t('sales.status')">
              <option value="">{{ app.t('sales.statusPlaceholder') }}</option>
              <option value="COMPLETED">{{ app.t('sales.status.completed') }}</option>
              <option value="CANCELLED">{{ app.t('sales.status.cancelled') }}</option>
            </AppSelect>
          </div>
          <div class="grid min-w-0 max-w-full gap-3 lg:col-span-2 lg:grid-cols-[minmax(0,1fr)_240px] lg:items-end 2xl:col-span-4">
            <AppDateRangeFilter class="min-w-0 max-w-full" v-model:date-from="filters.date_from" v-model:date-to="filters.date_to"
              :date-from-label="app.t('sales.dateFrom')" :date-to-label="app.t('sales.dateTo')"
              :date-placeholder="app.t('reports.selectDate')"
              :month-placeholder="app.t('reports.selectMonth')"
              :today-label="app.t('reports.today')"
              :this-month-label="app.t('reports.thisMonth')"
              :locale="app.language === 'th' ? 'th-TH-u-ca-buddhist' : 'en-US'"
              :show-shortcuts="false" />

            <div class="grid grid-cols-2 gap-2 lg:w-60">
              <AppButton class="w-full whitespace-nowrap" icon="search" @click="loadSales">
                {{ app.t('sales.apply') }}
              </AppButton>

              <AppButton class="w-full whitespace-nowrap" variant="secondary" @click="resetFilters">
                {{ app.t('sales.reset') }}
              </AppButton>
            </div>
          </div>
        </div>
      </AppCard>

      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
      <div v-if="loading" class="rounded-lg border border-slate-200 bg-white p-6 text-sm text-slate-500 dark:bg-slate-900 dark:text-slate-300">{{ app.t('sales.loading') }}</div>
      <AppEmptyState v-else-if="sales.length === 0" :title="app.t('sales.empty')" :description="app.t('sales.emptyDescription')" />

      <AppCard v-else class="min-w-0 max-w-full overflow-hidden dark:bg-slate-900/80">
        <div class="min-w-0 max-w-full">
        <div class="hidden w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain pb-2 [scrollbar-gutter:stable] md:block">
          <table class="w-full min-w-[1280px] divide-y divide-slate-200 text-sm dark:divide-slate-800">
            <thead class="bg-slate-50 dark:bg-slate-950/60">
              <tr>
                <th class="px-3 py-3 text-left font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.receipt') }}</th>
                <th class="px-3 py-3 text-left font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.date') }}</th>
                <th class="px-3 py-3 text-left font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.location') }}</th>
                <th class="px-3 py-3 text-left font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.cashier') }}</th>
                <th class="px-3 py-3 text-right font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.total') }}</th>
                <th class="px-3 py-3 text-right font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.payment') }}</th>
                <th class="px-3 py-3 text-right font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.status') }}</th>
                <th class="px-3 py-3 text-right font-black text-slate-600 dark:text-slate-300">{{ app.t('sales.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="sale in visibleSales" :key="sale.id" class="transition hover:bg-slate-50/80 dark:hover:bg-slate-950/50">
                <td class="px-3 py-3 font-semibold">{{ sale.receipt_no }}</td>
                <td class="px-3 py-3">{{ formatDate(sale.created_at) }}</td>
                <td class="px-3 py-3">{{ sale.location_name }}</td>
                <td class="px-3 py-3">{{ sale.cashier_name }}</td>
                <td class="px-3 py-3 text-right font-semibold">{{ money(sale.total_amount) }}</td>
                <td class="px-3 py-3 text-right">{{ paymentLabel(sale.payment_method) }}</td>
                <td class="px-3 py-3 text-right"><span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(sale.status)">{{ statusLabel(sale.status) }}</span></td>
                <td class="px-3 py-3 text-right">
                  <div class="flex justify-end gap-2 whitespace-nowrap">
                    <RouterLink v-if="canViewReceipt" :to="receiptRoute(sale)"
                      class="inline-flex box-border h-10 min-h-10 w-32 min-w-32 shrink-0 items-center justify-center gap-2 rounded-md border border-slate-200 bg-white px-3 py-0 text-sm font-semibold text-slate-700 transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800">
                      <AppIcon name="receipt-text" :size="16" />
                      <span class="truncate">{{ app.t('sales.viewReceipt') }}</span>
                    </RouterLink>

                    <AppButton v-if="canCancel && sale.status === 'COMPLETED'"
                      class="!box-border !h-10 !min-h-10 !w-32 !min-w-32 !shrink-0 !px-3 !py-0" variant="danger"
                      @click="openCancel(sale)">
                      <span class="truncate">{{ app.t('sales.cancel') }}</span>
                    </AppButton>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-3 md:hidden">
          <article v-for="sale in visibleSales" :key="sale.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-950/60">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="font-bold">{{ sale.receipt_no }}</h2>
                <p class="text-sm text-slate-500">{{ formatDate(sale.created_at) }}</p>
              </div>
              <span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(sale.status)">{{ statusLabel(sale.status) }}</span>
            </div>
            <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
              <div><dt class="text-slate-500">{{ app.t('sales.location') }}</dt><dd class="font-semibold">{{ sale.location_name }}</dd></div>
              <div><dt class="text-slate-500">{{ app.t('sales.cashier') }}</dt><dd class="font-semibold">{{ sale.cashier_name }}</dd></div>
              <div><dt class="text-slate-500">{{ app.t('sales.payment') }}</dt><dd class="font-semibold">{{ paymentLabel(sale.payment_method) }}</dd></div>
              <div><dt class="text-slate-500">{{ app.t('sales.total') }}</dt><dd class="font-bold">{{ moneyWithCurrency(sale.total_amount) }}</dd></div>
            </dl>
            <p v-if="sale.status === 'CANCELLED'" class="mt-2 text-sm text-slate-500">{{ app.t('sales.reason') }}: {{ sale.cancel_reason || '-' }}</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <RouterLink v-if="canViewReceipt" :to="receiptRoute(sale)" class="inline-flex min-h-10 items-center justify-center gap-2 rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100">
                <AppIcon name="receipt-text" :size="16" />
                {{ app.t('sales.viewReceipt') }}
              </RouterLink>
              <AppButton v-if="canCancel && sale.status === 'COMPLETED'" variant="danger" @click="openCancel(sale)">{{ app.t('sales.cancelSale') }}</AppButton>
            </div>
          </article>
        </div>

        <div class="mt-4 flex flex-col gap-3 border-t border-slate-200 pt-4 text-sm dark:border-slate-800 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-slate-500 dark:text-slate-400">{{ app.t('sales.show') }}</span>
            <AppPageSizeSelect :model-value="pageSize" @update:model-value="changePageSize" />
            <span class="text-slate-500 dark:text-slate-400">{{ app.t('sales.perPage') }}</span>
            <span class="text-slate-500 dark:text-slate-400">{{ app.t('sales.totalRows') }} {{ sales.length }}</span>
          </div>
          <div class="flex items-center justify-end gap-2">
            <AppButton variant="secondary" :disabled="page <= 1" @click="previousPage">{{ app.t('sales.previous') }}</AppButton>
            <span class="font-bold text-slate-600 dark:text-slate-300">{{ t('sales.page', { page, total: totalPages }) }}</span>
            <AppButton variant="secondary" :disabled="page >= totalPages" @click="nextPage">{{ app.t('sales.next') }}</AppButton>
          </div>
        </div>
        </div>
      </AppCard>
    </div>

    <ConfirmDialog
      :open="Boolean(cancelTarget)"
      :title="app.t('sales.cancelSale')"
      :message="cancelTarget ? t('sales.cancelMessage', { receipt: cancelTarget.receipt_no, location: cancelTarget.location_name }) : ''"
      :confirm-label="app.t('sales.cancel')"
      :cancel-label="app.t('sales.close')"
      @close="cancelTarget = null"
      @confirm="cancelSale"
    >
      <AppTextarea v-model="cancelReason" :label="app.t('sales.cancelReason')" :placeholder="app.t('sales.cancelReasonPlaceholder')" />
    </ConfirmDialog>
  </section>
</template>
