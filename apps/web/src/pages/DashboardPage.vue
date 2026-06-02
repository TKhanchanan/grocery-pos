<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { apiClient } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import type { DashboardSummary, StockStatus } from '../types/navigation'

const summary = ref<DashboardSummary | null>(null)
const loading = ref(false)
const error = ref('')

const alertTotal = computed(() => (summary.value?.low_stock_count ?? 0) + (summary.value?.out_of_stock_count ?? 0) + (summary.value?.reorder_count ?? 0))

function money(value: number) {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function stockClass(status: StockStatus) {
  return {
    in_stock: 'bg-brand-100 text-brand-700',
    low_stock: 'bg-amber-100 text-amber-800',
    out_of_stock: 'bg-red-100 text-red-700',
    reorder_point: 'bg-blue-100 text-blue-700',
  }[status]
}

async function loadDashboard() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await apiClient<DashboardSummary>('/v1/dashboard/summary')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load dashboard'
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboard)
</script>

<template>
  <section>
    <PageHeader title="Dashboard" eyebrow="Overview" description="Daily sales, monthly profit, alerts, and stock signals from live POS data.">
      <AppButton variant="secondary" @click="loadDashboard">Refresh</AppButton>
    </PageHeader>

    <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
    <div v-if="loading" class="rounded-lg border border-slate-200 bg-white p-6 text-sm text-slate-500">Loading dashboard...</div>

    <template v-if="summary">
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Today sales" :value="`${money(summary.today_sales)} บาท`" :helper="`${summary.today_receipts} receipts today`" />
        <StatCard label="This month profit" :value="`${money(summary.gross_profit_this_month)} บาท`" helper="Completed sales only" />
        <StatCard label="Top product" :value="summary.top_product_this_month?.product_name ?? '-'" :helper="summary.top_product_this_month ? `${summary.top_product_this_month.quantity} sold this month` : 'No sales yet'" />
        <StatCard label="Open alerts" :value="alertTotal" :helper="`${summary.out_of_stock_count} out · ${summary.reorder_count} reorder`" />
      </div>

      <div class="mt-4 grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <AppCard>
          <div class="flex items-center justify-between gap-3">
            <h2 class="font-bold">Recent sales</h2>
            <RouterLink to="/sales-history" class="text-sm font-semibold text-brand-700">View all</RouterLink>
          </div>
          <AppEmptyState v-if="summary.recent_sales.length === 0" class="mt-4" title="No sales yet" description="Completed POS sales will appear here." />
          <div v-else class="mt-4 overflow-x-auto">
            <table class="min-w-full divide-y divide-slate-200 text-sm">
              <thead class="bg-slate-50">
                <tr>
                  <th class="px-3 py-2 text-left">Receipt</th>
                  <th class="px-3 py-2 text-left">Cashier</th>
                  <th class="px-3 py-2 text-left">Payment</th>
                  <th class="px-3 py-2 text-right">Total</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="sale in summary.recent_sales" :key="sale.id">
                  <td class="px-3 py-2 font-semibold">{{ sale.receipt_no }}</td>
                  <td class="px-3 py-2">{{ sale.cashier_name }}</td>
                  <td class="px-3 py-2">{{ sale.payment_method }}</td>
                  <td class="px-3 py-2 text-right font-semibold">{{ money(sale.total_amount) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </AppCard>

        <AppCard>
          <h2 class="font-bold">Payment today</h2>
          <AppEmptyState v-if="summary.payment_method_summary.length === 0" class="mt-4" title="No payments" description="Payment mix appears after today sales." />
          <div v-else class="mt-4 grid gap-3">
            <div v-for="payment in summary.payment_method_summary" :key="payment.payment_method" class="rounded-lg border border-slate-200 p-3">
              <div class="flex items-center justify-between">
                <span class="font-bold">{{ payment.payment_method }}</span>
                <AppBadge>{{ payment.receipt_count }} receipts</AppBadge>
              </div>
              <p class="mt-2 text-xl font-bold">{{ money(payment.revenue) }} บาท</p>
            </div>
          </div>
        </AppCard>
      </div>

      <div class="mt-4 grid gap-4 xl:grid-cols-2">
        <AppCard>
          <div class="flex items-center justify-between gap-3">
            <h2 class="font-bold">Low stock</h2>
            <RouterLink to="/alerts" class="text-sm font-semibold text-brand-700">Open alerts</RouterLink>
          </div>
          <AppEmptyState v-if="summary.low_stock_items.length === 0" class="mt-4" title="No low stock" description="Products below threshold will appear here." />
          <div v-else class="mt-4 grid gap-3 sm:grid-cols-2">
            <article v-for="item in summary.low_stock_items" :key="`${item.product_id}-${item.location_id}`" class="rounded-lg border border-slate-200 p-3">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <h3 class="truncate font-bold">{{ item.product_name }}</h3>
                  <p class="text-sm text-slate-500">{{ item.location_name }} · {{ item.sku }}</p>
                </div>
                <span class="rounded-full px-2 py-1 text-xs font-bold" :class="stockClass(item.stock_status)">{{ item.stock_status.replaceAll('_', ' ') }}</span>
              </div>
              <p class="mt-3 text-2xl font-bold">{{ item.quantity }}</p>
              <p class="text-xs text-slate-500">Threshold {{ item.threshold }} · Reorder {{ item.reorder_point }}</p>
            </article>
          </div>
        </AppCard>

        <AppCard>
          <h2 class="font-bold">Top products this month</h2>
          <AppEmptyState v-if="summary.top_products.length === 0" class="mt-4" title="No product sales" description="Best sellers appear after completed sales." />
          <div v-else class="mt-4 grid gap-3">
            <div v-for="product in summary.top_products" :key="product.product_id" class="rounded-lg border border-slate-200 p-3">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <h3 class="font-bold">{{ product.product_name }}</h3>
                  <p class="text-sm text-slate-500">{{ product.sku }}</p>
                </div>
                <div class="text-right">
                  <p class="font-bold">{{ product.quantity }} sold</p>
                  <p class="text-sm text-slate-500">{{ money(product.revenue) }} บาท</p>
                </div>
              </div>
            </div>
          </div>
        </AppCard>
      </div>
    </template>
  </section>
</template>
