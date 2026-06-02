<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { apiClient } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import type { Receipt } from '../types/navigation'

const route = useRoute()
const receipt = ref<Receipt | null>(null)
const loading = ref(false)
const error = ref('')
const receiptID = computed(() => String(route.query.id ?? ''))

function money(value: number) {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function loadReceipt() {
  receipt.value = null
  error.value = ''
  if (!receiptID.value) return
  loading.value = true
  try {
    receipt.value = await apiClient<Receipt>(`/v1/sales/${receiptID.value}/receipt`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load receipt'
  } finally {
    loading.value = false
  }
}

function printReceipt() {
  window.print()
}

watch(receiptID, loadReceipt)
onMounted(loadReceipt)
</script>

<template>
  <section>
    <PageHeader title="Receipt Detail" eyebrow="Sales" description="Printable sale receipt with item snapshots, payment, and change.">
      <div class="flex flex-wrap gap-2 print:hidden">
        <RouterLink to="/pos" class="inline-flex min-h-10 items-center justify-center rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700">
          Back to POS
        </RouterLink>
        <AppButton :disabled="!receipt" @click="printReceipt">Print</AppButton>
      </div>
    </PageHeader>

    <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
    <div v-if="loading" class="rounded-lg border border-slate-200 bg-white p-6 text-sm text-slate-500">Loading receipt...</div>
    <AppEmptyState v-else-if="!receipt" title="No receipt selected" description="Open a receipt from the POS success modal or sales history." />

    <AppCard v-else class="mx-auto max-w-3xl print:border-0 print:shadow-none">
      <div class="flex flex-wrap items-start justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <p class="text-sm font-semibold uppercase text-brand-700">Grocery POS</p>
          <h2 class="text-2xl font-bold">{{ receipt.receipt_no }}</h2>
          <p class="text-sm text-slate-500">{{ new Date(receipt.created_at).toLocaleString('th-TH') }}</p>
        </div>
        <div class="text-right text-sm">
          <p><b>Location:</b> {{ receipt.location_name }}</p>
          <p><b>Cashier:</b> {{ receipt.cashier_name }}</p>
          <p><b>Status:</b> <span :class="receipt.status === 'CANCELLED' ? 'text-slate-500' : 'text-brand-700'">{{ receipt.status }}</span></p>
        </div>
      </div>

      <div v-if="receipt.status === 'CANCELLED'" class="mt-4 rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm">
        <p class="font-bold">Cancelled receipt</p>
        <p class="text-slate-600">Reason: {{ receipt.cancel_reason || '-' }}</p>
        <p v-if="receipt.cancelled_at" class="text-slate-500">Cancelled at {{ new Date(receipt.cancelled_at).toLocaleString('th-TH') }}</p>
      </div>

      <div class="mt-4 overflow-x-auto">
        <table class="min-w-full divide-y divide-slate-200 text-sm">
          <thead class="bg-slate-50">
            <tr>
              <th class="px-3 py-2 text-left">Product</th>
              <th class="px-3 py-2 text-right">Qty</th>
              <th class="px-3 py-2 text-right">Price</th>
              <th class="px-3 py-2 text-right">Total</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="item in receipt.items" :key="item.id">
              <td class="px-3 py-2">
                <p class="font-semibold">{{ item.product_name }}</p>
                <p class="text-xs text-slate-500">{{ item.sku }}</p>
              </td>
              <td class="px-3 py-2 text-right">{{ item.quantity }}</td>
              <td class="px-3 py-2 text-right">{{ money(item.price) }}</td>
              <td class="px-3 py-2 text-right font-semibold">{{ money(item.line_total) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <dl class="mt-4 ml-auto grid max-w-sm gap-2 text-sm">
        <div class="flex justify-between"><dt class="text-slate-500">Subtotal</dt><dd class="font-semibold">{{ money(receipt.subtotal) }} บาท</dd></div>
        <div class="flex justify-between"><dt class="text-slate-500">Total</dt><dd class="font-bold">{{ money(receipt.total_amount) }} บาท</dd></div>
        <div class="flex justify-between"><dt class="text-slate-500">Paid</dt><dd class="font-semibold">{{ money(receipt.paid_amount) }} บาท</dd></div>
        <div class="flex justify-between"><dt class="text-slate-500">Change</dt><dd class="font-bold">{{ money(receipt.change_amount) }} บาท</dd></div>
        <div class="flex justify-between"><dt class="text-slate-500">Payment</dt><dd class="font-semibold">{{ receipt.payment_method }}</dd></div>
      </dl>
    </AppCard>
  </section>
</template>
