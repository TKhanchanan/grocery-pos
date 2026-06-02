<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { apiClient, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTextarea from '../components/AppTextarea.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAuthStore } from '../stores/auth'
import type { Location, Receipt } from '../types/navigation'

const auth = useAuthStore()
const sales = ref<Receipt[]>([])
const locations = ref<Location[]>([])
const loading = ref(false)
const error = ref('')
const cancelTarget = ref<Receipt | null>(null)
const cancelReason = ref('')

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
const totals = computed(() => {
  const completed = sales.value.filter((sale) => sale.status === 'COMPLETED')
  return {
    completedCount: completed.length,
    cancelledCount: sales.value.filter((sale) => sale.status === 'CANCELLED').length,
    completedTotal: completed.reduce((sum, sale) => sum + sale.total_amount, 0),
  }
})

function money(value: number) {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function statusClass(status: Receipt['status']) {
  return status === 'CANCELLED' ? 'bg-slate-100 text-slate-600' : 'bg-brand-100 text-brand-700'
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
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load sales'
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

function openCancel(sale: Receipt) {
  cancelTarget.value = sale
  cancelReason.value = ''
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
    error.value = err instanceof Error ? err.message : 'Could not cancel sale'
  }
}

onMounted(async () => {
  await Promise.all([loadLocations(), loadSales()])
})
</script>

<template>
  <section>
    <PageHeader title="Sales History" eyebrow="Receipts" description="Search sales, inspect receipts, and cancel completed sales without deleting evidence.">
      <AppButton variant="secondary" @click="loadSales">Refresh</AppButton>
    </PageHeader>

    <div class="grid gap-4">
      <AppCard>
        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <AppInput v-model="filters.date_from" label="Date from" type="date" />
          <AppInput v-model="filters.date_to" label="Date to" type="date" />
          <AppInput v-model="filters.receipt_no" label="Receipt no." placeholder="RC..." />
          <AppInput v-model="filters.cashier_id" label="Cashier ID" type="number" />
          <AppSelect v-model="filters.location_id" label="Location">
            <option value="">All locations</option>
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <AppSelect v-model="filters.payment_method" label="Payment">
            <option value="">All payments</option>
            <option value="CASH">CASH</option>
            <option value="QR">QR</option>
          </AppSelect>
          <AppSelect v-model="filters.status" label="Status">
            <option value="">All statuses</option>
            <option value="COMPLETED">COMPLETED</option>
            <option value="CANCELLED">CANCELLED</option>
          </AppSelect>
          <div class="flex items-end gap-2">
            <AppButton class="flex-1" @click="loadSales">Apply</AppButton>
            <AppButton class="flex-1" variant="secondary" @click="resetFilters">Reset</AppButton>
          </div>
        </div>
      </AppCard>

      <div class="grid gap-3 md:grid-cols-3">
        <AppCard>
          <p class="text-sm text-slate-500">Completed sales</p>
          <p class="mt-1 text-2xl font-bold">{{ totals.completedCount }}</p>
        </AppCard>
        <AppCard>
          <p class="text-sm text-slate-500">Completed total</p>
          <p class="mt-1 text-2xl font-bold">{{ money(totals.completedTotal) }} บาท</p>
        </AppCard>
        <AppCard>
          <p class="text-sm text-slate-500">Cancelled</p>
          <p class="mt-1 text-2xl font-bold">{{ totals.cancelledCount }}</p>
        </AppCard>
      </div>

      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      <div v-if="loading" class="rounded-lg border border-slate-200 bg-white p-6 text-sm text-slate-500">Loading sales...</div>
      <AppEmptyState v-else-if="sales.length === 0" title="No sales found" description="Try changing filters or make a sale from POS." />

      <AppCard v-else>
        <div class="hidden overflow-x-auto lg:block">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50">
              <tr>
                <th class="px-3 py-2 text-left">Receipt</th>
                <th class="px-3 py-2 text-left">Date</th>
                <th class="px-3 py-2 text-left">Location</th>
                <th class="px-3 py-2 text-left">Cashier</th>
                <th class="px-3 py-2 text-right">Total</th>
                <th class="px-3 py-2 text-left">Payment</th>
                <th class="px-3 py-2 text-left">Status</th>
                <th class="px-3 py-2 text-left">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="sale in sales" :key="sale.id">
                <td class="px-3 py-2 font-semibold">{{ sale.receipt_no }}</td>
                <td class="px-3 py-2">{{ new Date(sale.created_at).toLocaleString('th-TH') }}</td>
                <td class="px-3 py-2">{{ sale.location_name }}</td>
                <td class="px-3 py-2">{{ sale.cashier_name }}</td>
                <td class="px-3 py-2 text-right font-semibold">{{ money(sale.total_amount) }}</td>
                <td class="px-3 py-2">{{ sale.payment_method }}</td>
                <td class="px-3 py-2"><span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(sale.status)">{{ sale.status }}</span></td>
                <td class="px-3 py-2">
                  <div class="flex flex-wrap gap-2">
                    <RouterLink :to="`/receipt-detail?id=${sale.id}`" class="inline-flex min-h-10 items-center justify-center rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">
                      Receipt
                    </RouterLink>
                    <AppButton v-if="canCancel && sale.status === 'COMPLETED'" variant="danger" @click="openCancel(sale)">Cancel</AppButton>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-3 lg:hidden">
          <article v-for="sale in sales" :key="sale.id" class="rounded-lg border border-slate-200 p-3">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="font-bold">{{ sale.receipt_no }}</h2>
                <p class="text-sm text-slate-500">{{ new Date(sale.created_at).toLocaleString('th-TH') }}</p>
              </div>
              <span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(sale.status)">{{ sale.status }}</span>
            </div>
            <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
              <div><dt class="text-slate-500">Location</dt><dd class="font-semibold">{{ sale.location_name }}</dd></div>
              <div><dt class="text-slate-500">Cashier</dt><dd class="font-semibold">{{ sale.cashier_name }}</dd></div>
              <div><dt class="text-slate-500">Payment</dt><dd class="font-semibold">{{ sale.payment_method }}</dd></div>
              <div><dt class="text-slate-500">Total</dt><dd class="font-bold">{{ money(sale.total_amount) }} บาท</dd></div>
            </dl>
            <p v-if="sale.status === 'CANCELLED'" class="mt-2 text-sm text-slate-500">Reason: {{ sale.cancel_reason || '-' }}</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <RouterLink :to="`/receipt-detail?id=${sale.id}`" class="inline-flex min-h-10 items-center justify-center rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">
                Receipt
              </RouterLink>
              <AppButton v-if="canCancel && sale.status === 'COMPLETED'" variant="danger" @click="openCancel(sale)">Cancel sale</AppButton>
            </div>
          </article>
        </div>
      </AppCard>
    </div>

    <ConfirmDialog
      :open="Boolean(cancelTarget)"
      title="Cancel sale"
      :message="cancelTarget ? `Cancel receipt ${cancelTarget.receipt_no}? Stock will be restored to ${cancelTarget.location_name}.` : ''"
      @close="cancelTarget = null"
      @confirm="cancelSale"
    >
      <AppTextarea v-model="cancelReason" label="Cancel reason" placeholder="Required reason" />
    </ConfirmDialog>
  </section>
</template>
