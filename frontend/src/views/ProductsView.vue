<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, postJSON } from '../api'
import StatusBox from '../components/StatusBox.vue'
import type { Category, Product } from '../types'

const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const query = ref('')
const loading = ref(true)
const error = ref('')
const form = reactive({
  id: 0,
  categoryId: null as number | null,
  sku: '',
  barcode: '',
  name: '',
  unit: 'ชิ้น',
  price: 0,
  cost: 0,
  threshold: 0,
  reorderPoint: 0,
  active: true,
})

function edit(p: Product) {
  Object.assign(form, { ...p, barcode: p.barcode ?? '' })
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    products.value = await api<Product[]>(`/products?q=${encodeURIComponent(query.value)}`)
    categories.value = await api<Category[]>('/categories')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load products'
  } finally {
    loading.value = false
  }
}

async function save() {
  await postJSON('/products', { ...form, barcode: form.barcode || null })
  Object.assign(form, { id: 0, sku: '', barcode: '', name: '', unit: 'ชิ้น', price: 0, cost: 0, threshold: 0, reorderPoint: 0, active: true })
  await load()
}

onMounted(load)
</script>

<template>
  <section class="grid gap-5 lg:grid-cols-[360px_1fr]">
    <form class="panel space-y-3" @submit.prevent="save">
      <h2 class="text-xl font-bold">{{ form.id ? 'Edit Product' : 'Create Product' }}</h2>
      <input v-model="form.name" class="input" placeholder="Product name" />
      <input v-model="form.sku" class="input" placeholder="Unique SKU" />
      <input v-model="form.barcode" class="input" placeholder="Barcode optional" />
      <div class="grid grid-cols-2 gap-3">
        <input v-model="form.unit" class="input" placeholder="Unit" />
        <select v-model="form.categoryId" class="input">
          <option :value="null">No category</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <input v-model.number="form.price" class="input" type="number" step="0.01" placeholder="Price" />
        <input v-model.number="form.cost" class="input" type="number" step="0.01" placeholder="Cost" />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <input v-model.number="form.threshold" class="input" type="number" placeholder="Low threshold" />
        <input v-model.number="form.reorderPoint" class="input" type="number" placeholder="Reorder point" />
      </div>
      <label class="flex items-center gap-2 text-sm"><input v-model="form.active" type="checkbox" /> Active</label>
      <button class="btn-primary w-full">Save Product</button>
    </form>
    <section class="space-y-3">
      <div class="flex flex-col gap-2 sm:flex-row">
        <input v-model="query" class="input" placeholder="Search by name, SKU, or barcode" @keyup.enter="load" />
        <button class="btn-soft" @click="load">Search</button>
      </div>
      <StatusBox :loading="loading" :error="error" :empty="products.length === 0">
        <div class="table-wrap">
          <table class="table">
            <thead><tr><th>Name</th><th>SKU</th><th>Barcode</th><th>Price</th><th>Stock</th><th>Alerts</th><th></th></tr></thead>
            <tbody>
              <tr v-for="p in products" :key="p.id">
                <td class="font-semibold">{{ p.name }}</td>
                <td>{{ p.sku }}</td>
                <td>{{ p.barcode ?? '-' }}</td>
                <td>{{ p.price.toFixed(2) }}</td>
                <td>{{ p.totalStock }} {{ p.unit }}</td>
                <td>Low {{ p.threshold }} · Reorder {{ p.reorderPoint }}</td>
                <td><button class="btn-soft" @click="edit(p)">Edit</button></td>
              </tr>
            </tbody>
          </table>
        </div>
      </StatusBox>
    </section>
  </section>
</template>
