<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../api'
import type { ReportRow } from '../types'

const reports = [
  ['daily-sales', 'Daily sales'],
  ['monthly-sales', 'Monthly sales'],
  ['best-selling', 'Best selling products'],
  ['profit-products', 'Profit per product'],
  ['stock', 'Stock report'],
  ['valuation', 'Inventory valuation'],
  ['payments', 'Payment summary'],
] as const

const active = ref<(typeof reports)[number][0]>('daily-sales')
const rows = ref<ReportRow[]>([])
const error = ref('')

async function load() {
  error.value = ''
  try {
    rows.value = await api<ReportRow[]>(`/reports/${active.value}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load report'
  }
}

async function exportCsv() {
  const csv = await api<string>(`/export/${active.value}`)
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${active.value}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(load)
</script>

<template>
  <section class="space-y-5">
    <div><p class="label">Revenue, profit, stock</p><h2 class="text-2xl font-bold">Reports</h2></div>
    <div class="panel flex flex-col gap-3 sm:flex-row">
      <select v-model="active" class="input" @change="load">
        <option v-for="[key, label] in reports" :key="key" :value="key">{{ label }}</option>
      </select>
      <button class="btn-primary" @click="exportCsv">Export CSV</button>
    </div>
    <p v-if="error" class="panel border-red-200 bg-red-50 text-red-700">{{ error }}</p>
    <div v-else-if="rows.length === 0" class="panel text-sm text-slate-500">No rows for this report yet.</div>
    <div v-else class="table-wrap">
      <table class="table">
        <thead><tr><th v-for="key in Object.keys(rows[0])" :key="key">{{ key }}</th></tr></thead>
        <tbody>
          <tr v-for="(row, idx) in rows" :key="idx">
            <td v-for="key in Object.keys(rows[0])" :key="key">{{ row[key] }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <section class="panel">
      <h3 class="font-bold">Product Import Template</h3>
      <p class="mt-2 text-sm text-slate-600">CSV columns: sku, barcode, name, unit, price, cost, threshold, reorder_point.</p>
    </section>
  </section>
</template>
