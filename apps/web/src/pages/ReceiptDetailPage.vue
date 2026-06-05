<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { apiClient } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAppStore } from '../stores/app'
import type { Receipt } from '../types/navigation'
import { formatThaiDateTime } from '../utils/date'
import { prepareReceiptPrintArea, resetReceiptPrintArea } from '../utils/print'

const app = useAppStore()
const route = useRoute()
const receipt = ref<Receipt | null>(null)
const loading = ref(false)
const error = ref('')
const receiptID = computed(() => {
  const paramID = route.params.id
  if (Array.isArray(paramID)) return paramID[0] ?? ''
  return String(paramID ?? route.query.id ?? '')
})
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')
const handleBeforePrint = () => prepareReceiptPrintArea()
const handleAfterPrint = () => resetReceiptPrintArea()

function money(value: number) {
  return value.toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatDate(value: string | null) {
  return formatThaiDateTime(value)
}

function paymentLabel(method: string) {
  return method === 'QR' ? app.t('pos.qr') : app.t('pos.cash')
}

async function loadReceipt() {
  receipt.value = null
  error.value = ''
  if (!receiptID.value) return
  loading.value = true
  try {
    receipt.value = await apiClient<Receipt>(`/v1/sales/${receiptID.value}/receipt`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('receipt.loadFailed')
  } finally {
    loading.value = false
  }
}

function printReceipt() {
  if (!receipt.value || loading.value) return
  prepareReceiptPrintArea()
  window.print()
}

watch(receiptID, loadReceipt)
onMounted(() => {
  window.addEventListener('beforeprint', handleBeforePrint)
  window.addEventListener('afterprint', handleAfterPrint)
  loadReceipt()
})
onBeforeUnmount(() => {
  window.removeEventListener('beforeprint', handleBeforePrint)
  window.removeEventListener('afterprint', handleAfterPrint)
})
</script>

<template>
  <section class="receipt-page">
    <div class="no-print">
      <PageHeader :title="app.t('receipt.title')" :eyebrow="app.t('receipt.eyebrow')" :description="app.t('receipt.description')" icon="receipt-text">
        <div class="flex flex-wrap gap-2">
          <RouterLink to="/pos" class="focus-ring inline-flex min-h-11 items-center justify-center rounded-xl bg-white/85 px-4 py-2.5 text-sm font-bold text-slate-700 shadow-sm hover:bg-brand-50 dark:bg-slate-900/85 dark:text-slate-100 dark:hover:bg-teal-400/10">
            {{ app.t('receipt.backToPOS') }}
          </RouterLink>
          <RouterLink to="/sales-history" class="focus-ring inline-flex min-h-11 items-center justify-center rounded-xl bg-white/85 px-4 py-2.5 text-sm font-bold text-slate-700 shadow-sm hover:bg-brand-50 dark:bg-slate-900/85 dark:text-slate-100 dark:hover:bg-teal-400/10">
            {{ app.t('receipt.backToSales') }}
          </RouterLink>
          <AppButton :disabled="!receipt || loading" icon="receipt-text" @click="printReceipt">{{ app.t('receipt.print') }}</AppButton>
        </div>
      </PageHeader>

      <div v-if="error" class="mb-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-400/30 dark:bg-red-500/10 dark:text-red-100">{{ error }}</div>
      <div v-if="loading" class="mb-4 rounded-xl bg-white/85 p-6 text-sm text-slate-500 shadow-sm dark:bg-slate-900/85 dark:text-slate-300">{{ app.t('receipt.loading') }}</div>
      <AppEmptyState v-else-if="!receipt" :title="app.t('receipt.noReceipt')" :description="app.t('receipt.noReceiptDescription')" />
    </div>

    <div v-if="receipt" class="receipt-preview">
      <article id="receipt-print-area" class="receipt-print-area">
        <header class="receipt-shop">
          <p class="receipt-shop-name">Grocery POS</p>
          <p>{{ app.t('receipt.receipt') }}</p>
        </header>

        <dl class="receipt-meta">
          <div><dt>{{ app.t('receipt.receiptNo') }}</dt><dd>{{ receipt.receipt_no }}</dd></div>
          <div><dt>{{ app.t('receipt.date') }}</dt><dd>{{ formatDate(receipt.created_at) }}</dd></div>
          <div><dt>{{ app.t('receipt.cashier') }}</dt><dd>{{ receipt.cashier_name }}</dd></div>
          <div><dt>{{ app.t('receipt.location') }}</dt><dd>{{ receipt.location_name }}</dd></div>
          <div><dt>{{ app.t('receipt.paymentMethod') }}</dt><dd>{{ paymentLabel(receipt.payment_method) }}</dd></div>
          <div><dt>{{ app.t('receipt.status') }}</dt><dd>{{ receipt.status }}</dd></div>
        </dl>

        <div v-if="receipt.status === 'CANCELLED'" class="receipt-cancelled">
          <p><b>{{ app.t('receipt.cancelledReceipt') }}</b></p>
          <p>{{ app.t('receipt.reason') }}: {{ receipt.cancel_reason || '-' }}</p>
          <p v-if="receipt.cancelled_at">{{ app.t('receipt.cancelledAt') }} {{ formatDate(receipt.cancelled_at) }}</p>
        </div>

        <section class="receipt-lines">
          <h2>{{ app.t('receipt.items') }}</h2>
          <div class="receipt-line receipt-line-head">
            <span>{{ app.t('receipt.items') }}</span>
            <span>{{ app.t('receipt.quantity') }}</span>
            <span>{{ app.t('receipt.price') }}</span>
            <span>{{ app.t('receipt.subtotal') }}</span>
          </div>
          <div v-for="item in receipt.items" :key="item.id" class="receipt-line">
            <span>
              <b>{{ item.product_name }}</b>
              <small>{{ item.sku }}</small>
            </span>
            <span>{{ item.quantity }}</span>
            <span>{{ money(item.price) }}</span>
            <span>{{ money(item.line_total) }}</span>
          </div>
        </section>

        <dl class="receipt-totals">
          <div><dt>{{ app.t('receipt.subtotal') }}</dt><dd>{{ money(receipt.subtotal) }}</dd></div>
          <div class="receipt-total-row"><dt>{{ app.t('receipt.total') }}</dt><dd>{{ money(receipt.total_amount) }}</dd></div>
          <div><dt>{{ app.t('receipt.received') }}</dt><dd>{{ money(receipt.paid_amount) }}</dd></div>
          <div><dt>{{ app.t('receipt.change') }}</dt><dd>{{ money(receipt.change_amount) }}</dd></div>
        </dl>

        <footer class="receipt-footer">
          {{ app.t('receipt.thankYou') }}
        </footer>
      </article>
    </div>
  </section>
</template>

<style scoped>
.receipt-preview {
  display: flex;
  justify-content: center;
  padding-bottom: 2rem;
}

.receipt-print-area {
  width: min(100%, 420px);
  background: #ffffff;
  color: #111827;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.14);
}

.receipt-shop {
  border-bottom: 1px dashed #94a3b8;
  padding-bottom: 10px;
  text-align: center;
}

.receipt-shop-name {
  font-size: 18px;
  font-weight: 900;
}

.receipt-meta,
.receipt-totals {
  display: grid;
  gap: 6px;
  margin-top: 12px;
  font-size: 13px;
}

.receipt-meta div,
.receipt-totals div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.receipt-meta dt,
.receipt-totals dt {
  color: #64748b;
}

.receipt-meta dd,
.receipt-totals dd {
  margin: 0;
  text-align: right;
  font-weight: 700;
}

.receipt-cancelled {
  margin-top: 12px;
  border: 1px dashed #94a3b8;
  padding: 8px;
  font-size: 12px;
}

.receipt-lines {
  margin-top: 14px;
  border-top: 1px dashed #94a3b8;
  border-bottom: 1px dashed #94a3b8;
  padding: 10px 0;
}

.receipt-lines h2 {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 900;
}

.receipt-line {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 34px 58px 64px;
  gap: 8px;
  padding: 5px 0;
  font-size: 12px;
  align-items: start;
}

.receipt-line span:not(:first-child) {
  text-align: right;
}

.receipt-line b,
.receipt-line small {
  display: block;
  min-width: 0;
  overflow-wrap: anywhere;
}

.receipt-line small {
  color: #64748b;
}

.receipt-line-head {
  color: #64748b;
  font-size: 11px;
  font-weight: 800;
}

.receipt-total-row {
  border-top: 1px dashed #94a3b8;
  margin-top: 4px;
  padding-top: 8px;
  font-size: 15px;
  font-weight: 900;
}

.receipt-footer {
  border-top: 1px dashed #94a3b8;
  margin-top: 14px;
  padding-top: 12px;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
}

@media print {
  .receipt-preview {
    display: block;
    padding: 0;
  }
}
</style>
