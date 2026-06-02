<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppCheckbox from '../components/AppCheckbox.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import { useAuthStore } from '../stores/auth'
import type { Category, Product, ProductStock, StockStatus } from '../types/navigation'

interface ProductForm {
  id: number
  sku: string
  name: string
  barcode: string
  category_id: string
  selling_price: number
  unit_cost: number
  unit: string
  threshold: number
  reorder_point: number
  is_active: boolean
}

const auth = useAuthStore()
const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const selectedStocks = ref<ProductStock[]>([])
const selectedProduct = ref<Product | null>(null)
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const filters = reactive({ q: '', category_id: '', status: '', stock_status: '' })
const form = reactive<ProductForm>({
  id: 0,
  sku: '',
  name: '',
  barcode: '',
  category_id: '',
  selling_price: 1,
  unit_cost: 0,
  unit: 'ชิ้น',
  threshold: 0,
  reorder_point: 0,
  is_active: true,
})

const canCreate = computed(() => auth.hasPermission('products.create'))
const canUpdate = computed(() => auth.hasPermission('products.update'))
const canDeactivate = computed(() => auth.hasPermission('products.deactivate'))
const canManage = computed(() => canCreate.value || canUpdate.value)
const activeCount = computed(() => products.value.filter((product) => product.is_active).length)
const lowSignalCount = computed(() => products.value.filter((product) => product.stock_status !== 'in_stock').length)
const totalStock = computed(() => products.value.reduce((sum, product) => sum + product.total_stock, 0))

function stockLabel(status: StockStatus) {
  return {
    in_stock: 'In stock',
    low_stock: 'Low stock',
    out_of_stock: 'Out of stock',
    reorder_point: 'Reorder point',
  }[status]
}

function stockClass(status: StockStatus) {
  return {
    in_stock: 'bg-brand-100 text-brand-700',
    low_stock: 'bg-amber-100 text-amber-800',
    out_of_stock: 'bg-red-100 text-red-700',
    reorder_point: 'bg-blue-100 text-blue-700',
  }[status]
}

function queryString() {
  const params = new URLSearchParams()
  if (filters.q) params.set('q', filters.q)
  if (filters.category_id) params.set('category_id', filters.category_id)
  if (filters.status) params.set('status', filters.status)
  if (filters.stock_status) params.set('stock_status', filters.stock_status)
  return params.toString()
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const qs = queryString()
    products.value = await apiClient<Product[]>(`/v1/products${qs ? `?${qs}` : ''}`)
    categories.value = await apiClient<Category[]>('/v1/categories')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load products'
  } finally {
    loading.value = false
  }
}

function resetForm() {
  Object.assign(form, {
    id: 0,
    sku: '',
    name: '',
    barcode: '',
    category_id: '',
    selling_price: 1,
    unit_cost: 0,
    unit: 'ชิ้น',
    threshold: 0,
    reorder_point: 0,
    is_active: true,
  })
}

function edit(product: Product) {
  Object.assign(form, {
    id: product.id,
    sku: product.sku,
    name: product.name,
    barcode: product.barcode ?? '',
    category_id: product.category_id ? String(product.category_id) : '',
    selling_price: product.selling_price,
    unit_cost: product.unit_cost,
    unit: product.unit,
    threshold: product.threshold,
    reorder_point: product.reorder_point,
    is_active: product.is_active,
  })
}

function payload() {
  return {
    sku: form.sku,
    name: form.name,
    barcode: form.barcode.trim() ? form.barcode.trim() : null,
    category_id: form.category_id ? Number(form.category_id) : null,
    selling_price: Number(form.selling_price),
    unit_cost: Number(form.unit_cost),
    unit: form.unit,
    threshold: Number(form.threshold || 0),
    reorder_point: Number(form.reorder_point || 0),
    is_active: form.is_active,
  }
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    if (form.id) await patchJSON<Product>(`/v1/products/${form.id}`, payload())
    else await postJSON<Product>('/v1/products', payload())
    resetForm()
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save product'
  } finally {
    saving.value = false
  }
}

async function setActive(product: Product, active: boolean) {
  await patchJSON<Product>(`/v1/products/${product.id}/status`, { is_active: active })
  await load()
}

async function showStocks(product: Product) {
  selectedProduct.value = product
  selectedStocks.value = await apiClient<ProductStock[]>(`/v1/products/${product.id}/stocks`)
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="Products" eyebrow="Catalog" description="Search by product name, SKU, or barcode. Product stock is read from product_stocks by location." icon="package">
      <AppButton variant="secondary" icon="history" @click="load">Refresh</AppButton>
    </PageHeader>

    <div class="mb-4 grid gap-3 sm:grid-cols-3">
      <StatCard label="Products" :value="products.length" :helper="`${activeCount} active`" icon="package" />
      <StatCard label="Total stock" :value="totalStock" helper="Across visible products" icon="map-pin" tone="success" />
      <StatCard label="Needs attention" :value="lowSignalCount" helper="Low, out, or reorder" icon="bell" tone="warning" />
    </div>

    <div class="grid gap-4 xl:grid-cols-[380px_1fr]">
      <AppCard v-if="canManage" hover>
        <form class="grid gap-3" @submit.prevent="save">
          <div>
            <h2 class="font-bold">{{ form.id ? 'Edit product' : 'Create product' }}</h2>
            <p class="mt-1 text-sm text-slate-500">SKU stays unique; product names can be reused.</p>
          </div>
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
            <AppInput v-model="form.sku" label="SKU" />
            <AppInput v-model="form.name" label="Name" />
            <AppInput v-model="form.barcode" label="Barcode" />
            <AppSelect v-model="form.category_id" label="Category">
              <option value="">No category</option>
              <option v-for="category in categories" :key="category.id" :value="String(category.id)">{{ category.name }}</option>
            </AppSelect>
            <AppInput v-model="form.unit" label="Unit" />
            <AppInput v-model="form.selling_price" label="Selling price" type="number" />
            <AppInput v-model="form.unit_cost" label="Unit cost" type="number" />
            <AppInput v-model="form.threshold" label="Threshold" type="number" />
            <AppInput v-model="form.reorder_point" label="Reorder point" type="number" />
          </div>
          <AppCheckbox v-model="form.is_active" label="Active" description="Inactive products stay in history but cannot be sold." />
          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <div class="flex gap-2">
            <AppButton type="submit" :loading="saving" :disabled="saving || (!form.id && !canCreate) || (Boolean(form.id) && !canUpdate)" icon="check-circle">{{ form.id ? 'Save' : 'Create' }}</AppButton>
            <AppButton v-if="form.id" variant="secondary" icon="x" @click="resetForm">Cancel</AppButton>
          </div>
        </form>
      </AppCard>

      <div class="space-y-4">
        <AppCard>
          <div class="grid gap-3 lg:grid-cols-5">
            <AppInput v-model="filters.q" label="Search" placeholder="Name, SKU, barcode" />
            <AppSelect v-model="filters.category_id" label="Category">
              <option value="">All</option>
              <option v-for="category in categories" :key="category.id" :value="String(category.id)">{{ category.name }}</option>
            </AppSelect>
            <AppSelect v-model="filters.status" label="Status">
              <option value="">All</option>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </AppSelect>
            <AppSelect v-model="filters.stock_status" label="Stock">
              <option value="">All</option>
              <option value="in_stock">In stock</option>
              <option value="low_stock">Low stock</option>
              <option value="out_of_stock">Out of stock</option>
              <option value="reorder_point">Reorder point</option>
            </AppSelect>
            <div class="flex items-end"><AppButton class="w-full" icon="search" @click="load">Apply</AppButton></div>
          </div>
        </AppCard>

        <AppCard>
          <AppLoadingState v-if="loading" label="Loading products..." />
          <AppEmptyState v-else-if="products.length === 0" title="No products" description="Try adjusting filters or create a new product." />

          <div v-else>
            <div class="hidden overflow-x-auto md:block">
              <table class="min-w-full divide-y divide-slate-200 text-sm">
                <thead class="bg-slate-50">
                  <tr>
                    <th class="px-3 py-2 text-left">Product</th>
                    <th class="px-3 py-2 text-left">SKU / Barcode</th>
                    <th class="px-3 py-2 text-left">Category</th>
                    <th class="px-3 py-2 text-right">Price</th>
                    <th class="px-3 py-2 text-right">Stock</th>
                    <th class="px-3 py-2 text-left">Status</th>
                    <th class="px-3 py-2 text-left">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="product in products" :key="product.id" class="hover:bg-slate-50/80">
                    <td class="px-3 py-2">
                      <p class="font-semibold">{{ product.name }}</p>
                      <p class="text-xs text-slate-500">{{ product.unit }}</p>
                    </td>
                    <td class="px-3 py-2">{{ product.sku }}<br /><span class="text-xs text-slate-500">{{ product.barcode || 'No barcode' }}</span></td>
                    <td class="px-3 py-2">{{ product.category_name || '-' }}</td>
                    <td class="px-3 py-2 text-right">{{ product.selling_price.toFixed(2) }}</td>
                    <td class="px-3 py-2 text-right">{{ product.total_stock }}</td>
                    <td class="px-3 py-2">
                      <span class="rounded-full px-2 py-1 text-xs font-bold" :class="stockClass(product.stock_status)">{{ stockLabel(product.stock_status) }}</span>
                      <span class="ml-2 text-xs text-slate-500">{{ product.is_active ? 'Active' : 'Inactive' }}</span>
                    </td>
                    <td class="px-3 py-2">
                      <div class="flex flex-wrap gap-2">
                        <AppButton variant="secondary" icon="map-pin" @click="showStocks(product)">Stocks</AppButton>
                        <AppButton v-if="canUpdate" variant="secondary" icon="settings" @click="edit(product)">Edit</AppButton>
                        <AppButton v-if="canDeactivate" :variant="product.is_active ? 'danger' : 'secondary'" @click="setActive(product, !product.is_active)">
                          {{ product.is_active ? 'Deactivate' : 'Activate' }}
                        </AppButton>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="grid gap-3 md:hidden">
              <article v-for="product in products" :key="product.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 shadow-sm">
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <h3 class="font-bold">{{ product.name }}</h3>
                    <p class="text-sm text-slate-500">{{ product.sku }} · {{ product.barcode || 'No barcode' }}</p>
                  </div>
                  <span class="rounded-full px-2 py-1 text-xs font-bold" :class="stockClass(product.stock_status)">{{ stockLabel(product.stock_status) }}</span>
                </div>
                <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                  <div><dt class="text-slate-500">Price</dt><dd class="font-semibold">{{ product.selling_price.toFixed(2) }}</dd></div>
                  <div><dt class="text-slate-500">Stock</dt><dd class="font-semibold">{{ product.total_stock }}</dd></div>
                  <div><dt class="text-slate-500">Category</dt><dd class="font-semibold">{{ product.category_name || '-' }}</dd></div>
                  <div><dt class="text-slate-500">Status</dt><dd class="font-semibold">{{ product.is_active ? 'Active' : 'Inactive' }}</dd></div>
                </dl>
                <div class="mt-3 flex flex-wrap gap-2">
                  <AppButton variant="secondary" icon="map-pin" @click="showStocks(product)">Stocks</AppButton>
                  <AppButton v-if="canUpdate" variant="secondary" icon="settings" @click="edit(product)">Edit</AppButton>
                </div>
              </article>
            </div>
          </div>
        </AppCard>

        <AppCard>
          <h2 class="font-bold">Product stock by location</h2>
          <p class="mt-1 text-sm text-slate-500">{{ selectedProduct ? selectedProduct.name : 'Select a product to view location stock.' }}</p>
          <div v-if="selectedStocks.length" class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <article v-for="stock in selectedStocks" :key="stock.location_id" class="rounded-2xl border border-slate-200 bg-white/65 p-4">
              <p class="font-semibold">{{ stock.location_name }}</p>
              <p class="mt-2 text-2xl font-bold text-brand-700">{{ stock.quantity }}</p>
              <span class="mt-2 inline-flex rounded-full px-2 py-1 text-xs font-bold" :class="stockClass(stock.stock_status)">{{ stockLabel(stock.stock_status) }}</span>
            </article>
          </div>
        </AppCard>
      </div>
    </div>
  </section>
</template>
