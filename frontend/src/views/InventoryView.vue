<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, postJSON } from '../api'
import StatusBox from '../components/StatusBox.vue'
import type { Alert, Location, Product, ProductStock, StockMovement } from '../types'

const products = ref<Product[]>([])
const locations = ref<Location[]>([])
const stocks = ref<ProductStock[]>([])
const movements = ref<StockMovement[]>([])
const alerts = ref<Alert[]>([])
const loading = ref(true)
const error = ref('')
const locationForm = reactive({ id: 0, name: '', active: true })
const stockForm = reactive({ productId: 0, locationId: 0, quantity: 1, unitCost: 0, reason: 'demo stock operation' })
const transferForm = reactive({ productId: 0, fromLocationId: 0, toLocationId: 0, quantity: 1, reason: 'demo transfer' })

const selectedProduct = computed(() => products.value.find((p) => p.id === stockForm.productId))

async function load() {
  loading.value = true
  error.value = ''
  try {
    products.value = await api<Product[]>('/products')
    locations.value = await api<Location[]>('/locations')
    stocks.value = await api<ProductStock[]>('/stocks')
    movements.value = await api<StockMovement[]>('/stock/movements')
    alerts.value = await api<Alert[]>('/alerts')
    if (!stockForm.productId && products.value[0]) stockForm.productId = products.value[0].id
    if (!stockForm.locationId && locations.value[0]) stockForm.locationId = locations.value[0].id
    if (!transferForm.productId && products.value[0]) transferForm.productId = products.value[0].id
    if (!transferForm.fromLocationId && locations.value[0]) transferForm.fromLocationId = locations.value[0].id
    if (!transferForm.toLocationId && locations.value[1]) transferForm.toLocationId = locations.value[1].id
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load inventory'
  } finally {
    loading.value = false
  }
}

async function saveLocation() {
  await postJSON('/locations', locationForm)
  Object.assign(locationForm, { id: 0, name: '', active: true })
  await load()
}

async function mutateStock(path: string) {
  await postJSON(path, stockForm)
  await load()
}

async function transfer() {
  await postJSON('/stock/transfer', transferForm)
  await load()
}

onMounted(load)
</script>

<template>
  <section class="space-y-5">
    <div>
      <p class="label">Multi-location stock</p>
      <h2 class="text-2xl font-bold">Inventory</h2>
    </div>
    <StatusBox :loading="loading" :error="error">
      <div class="grid gap-4 lg:grid-cols-3">
        <form class="panel space-y-3" @submit.prevent="saveLocation">
          <h3 class="font-bold">Locations</h3>
          <input v-model="locationForm.name" class="input" placeholder="Location name" />
          <label class="flex items-center gap-2 text-sm"><input v-model="locationForm.active" type="checkbox" /> Active</label>
          <button class="btn-primary w-full">Save Location</button>
          <ul class="text-sm">
            <li v-for="l in locations" :key="l.id" class="border-b border-emerald-50 py-2">{{ l.name }}</li>
          </ul>
        </form>

        <form class="panel space-y-3" @submit.prevent="mutateStock('/stock/restock')">
          <h3 class="font-bold">Restock / Adjust</h3>
          <select v-model.number="stockForm.productId" class="input">
            <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }} · {{ p.sku }}</option>
          </select>
          <select v-model.number="stockForm.locationId" class="input">
            <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option>
          </select>
          <div class="grid grid-cols-2 gap-3">
            <input v-model.number="stockForm.quantity" class="input" type="number" />
            <input v-model.number="stockForm.unitCost" class="input" type="number" step="0.01" placeholder="Unit cost" />
          </div>
          <input v-model="stockForm.reason" class="input" />
          <button class="btn-primary w-full">Restock {{ selectedProduct?.unit ?? '' }}</button>
          <button class="btn-soft w-full" type="button" @click="mutateStock('/stock/adjust')">Apply Adjustment</button>
        </form>

        <form class="panel space-y-3" @submit.prevent="transfer">
          <h3 class="font-bold">Transfer Stock</h3>
          <select v-model.number="transferForm.productId" class="input">
            <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <div class="grid grid-cols-2 gap-3">
            <select v-model.number="transferForm.fromLocationId" class="input">
              <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option>
            </select>
            <select v-model.number="transferForm.toLocationId" class="input">
              <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option>
            </select>
          </div>
          <input v-model.number="transferForm.quantity" class="input" type="number" />
          <input v-model="transferForm.reason" class="input" />
          <button class="btn-primary w-full">Transfer</button>
        </form>
      </div>

      <div class="grid gap-4 lg:grid-cols-[1fr_360px]">
        <div class="table-wrap">
          <table class="table">
            <thead><tr><th>Location</th><th>Product</th><th>SKU</th><th>Quantity</th></tr></thead>
            <tbody>
              <tr v-for="s in stocks" :key="`${s.productId}-${s.locationId}`">
                <td>{{ s.locationName }}</td><td class="font-semibold">{{ s.productName }}</td><td>{{ s.sku }}</td><td>{{ s.quantity }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <aside class="panel">
          <h3 class="font-bold">Alerts</h3>
          <ul class="mt-3 space-y-2 text-sm">
            <li v-for="a in alerts" :key="a.id" class="rounded-md bg-amber-50 p-2">
              <b>{{ a.type }}</b> · {{ a.productName }} at {{ a.locationName }}
            </li>
          </ul>
        </aside>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead><tr><th>Time</th><th>Product</th><th>Location</th><th>Type</th><th>Delta</th><th>Note</th></tr></thead>
          <tbody>
            <tr v-for="m in movements" :key="m.id">
              <td>{{ new Date(m.createdAt).toLocaleString() }}</td><td>{{ m.productName }}</td><td>{{ m.locationName }}</td><td>{{ m.referenceType }}</td><td>{{ m.quantityChange }}</td><td>{{ m.note }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </StatusBox>
  </section>
</template>
