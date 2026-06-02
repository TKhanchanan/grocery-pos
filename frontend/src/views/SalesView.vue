<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, postJSON } from '../api'
import StatusBox from '../components/StatusBox.vue'
import type { Sale } from '../types'

const sales = ref<Sale[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    sales.value = await api<Sale[]>('/sales')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load sales'
  } finally {
    loading.value = false
  }
}

async function cancelSale(id: number) {
  await postJSON(`/sales/${id}/cancel`, {})
  await load()
}

onMounted(load)
</script>

<template>
  <section class="space-y-5">
    <div><p class="label">Evidence and reversals</p><h2 class="text-2xl font-bold">Sales History</h2></div>
    <StatusBox :loading="loading" :error="error" :empty="sales.length === 0">
      <div class="grid gap-4">
        <article v-for="sale in sales" :key="sale.id" class="panel">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h3 class="font-bold">Receipt {{ sale.receiptNo }}</h3>
              <p class="text-sm text-slate-500">{{ sale.locationName }} · {{ new Date(sale.createdAt).toLocaleString() }}</p>
            </div>
            <div class="flex items-center gap-2">
              <span class="rounded-full px-3 py-1 text-xs font-bold" :class="sale.status === 'CANCELLED' ? 'bg-red-50 text-red-700' : 'bg-mint text-leaf'">{{ sale.status }}</span>
              <button v-if="sale.status !== 'CANCELLED'" class="btn-danger" @click="cancelSale(sale.id)">Cancel & Restore Stock</button>
            </div>
          </div>
          <div class="mt-4 table-wrap">
            <table class="table">
              <thead><tr><th>Product</th><th>SKU</th><th>Qty</th><th>Total</th></tr></thead>
              <tbody>
                <tr v-for="item in sale.items" :key="item.id">
                  <td>{{ item.productNameSnapshot }}</td><td>{{ item.skuSnapshot }}</td><td>{{ item.quantity }}</td><td>{{ item.lineTotal.toFixed(2) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="mt-3 text-right font-bold">Total {{ sale.totalAmount.toFixed(2) }} · Profit {{ sale.profit.toFixed(2) }} · {{ sale.paymentMethod }}</p>
        </article>
      </div>
    </StatusBox>
  </section>
</template>
