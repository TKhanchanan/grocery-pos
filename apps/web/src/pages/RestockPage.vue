<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppInput from '../components/AppInput.vue'
import AppModal from '../components/AppModal.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTextarea from '../components/AppTextarea.vue'
import PageHeader from '../components/PageHeader.vue'
import type { Location, Product, ProductStock, StockMovement } from '../types/navigation'

const products = ref<Product[]>([])
const locations = ref<Location[]>([])
const stocks = ref<ProductStock[]>([])
const result = ref<StockMovement | null>(null)
const error = ref('')
const adjustOpen = ref(false)

const form = reactive({
  product_id: '',
  location_id: '',
  quantity: 100,
  total_cost: 200,
  unit_cost: 0,
  note: 'Restock demo',
})

const adjustment = reactive({
  quantity: -1,
  note: '',
})

const selectedProduct = computed(() => products.value.find((product) => product.id === Number(form.product_id)) ?? null)
const selectedLocation = computed(() => locations.value.find((location) => location.id === Number(form.location_id)) ?? null)
const currentStock = computed(() => stocks.value.find((stock) => stock.product_id === Number(form.product_id) && stock.location_id === Number(form.location_id))?.quantity ?? 0)
const unitCostPreview = computed(() => form.total_cost > 0 && form.quantity > 0 ? form.total_cost / form.quantity : Number(form.unit_cost || 0))
const afterRestockPreview = computed(() => currentStock.value + Number(form.quantity || 0))

async function load() {
  products.value = await apiClient<Product[]>('/v1/products')
  locations.value = await apiClient<Location[]>('/v1/locations')
  stocks.value = await apiClient<ProductStock[]>('/v1/product-stocks')
  if (!form.product_id && products.value[0]) form.product_id = String(products.value[0].id)
  if (!form.location_id && locations.value[0]) form.location_id = String(locations.value[0].id)
}

async function restock() {
  error.value = ''
  result.value = null
  try {
    result.value = await postJSON<StockMovement>(`/v1/products/${form.product_id}/restock`, {
      location_id: Number(form.location_id),
      quantity: Number(form.quantity),
      total_cost: form.total_cost ? Number(form.total_cost) : null,
      unit_cost: Number(unitCostPreview.value),
      note: form.note,
    })
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Restock failed'
  }
}

async function adjustStock() {
  error.value = ''
  try {
    result.value = await postJSON<StockMovement>(`/v1/products/${form.product_id}/adjust-stock`, {
      location_id: Number(form.location_id),
      quantity: Number(adjustment.quantity),
      note: adjustment.note,
    })
    adjustOpen.value = false
    adjustment.note = ''
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Adjustment failed'
  }
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="Restock" eyebrow="Stock operations" description="Receive stock into a selected location and record before/after stock in movement history." />
    <div class="grid gap-4 lg:grid-cols-[420px_1fr]">
      <AppCard>
        <form class="grid gap-3" @submit.prevent="restock">
          <AppSelect v-model="form.product_id" label="Product">
            <option v-for="product in products" :key="product.id" :value="String(product.id)">{{ product.name }} · {{ product.sku }}</option>
          </AppSelect>
          <AppSelect v-model="form.location_id" label="Location">
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <div class="rounded-lg bg-slate-50 p-3 text-sm">
            <p><b>Current stock:</b> {{ currentStock }} {{ selectedProduct?.unit }}</p>
            <p><b>After restock:</b> {{ afterRestockPreview }} {{ selectedProduct?.unit }}</p>
            <p><b>Unit cost preview:</b> {{ unitCostPreview.toFixed(2) }} บาท</p>
          </div>
          <AppInput v-model="form.quantity" label="Quantity" type="number" />
          <AppInput v-model="form.total_cost" label="Total cost" type="number" />
          <AppInput v-model="form.unit_cost" label="Unit cost fallback" type="number" />
          <AppTextarea v-model="form.note" label="Note" />
          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <div class="flex flex-wrap gap-2">
            <AppButton type="submit">Restock</AppButton>
            <AppButton variant="secondary" @click="adjustOpen = true">Adjust stock</AppButton>
          </div>
        </form>
      </AppCard>

      <div class="grid gap-4">
        <AppCard>
          <h2 class="font-bold">Selected stock</h2>
          <div class="mt-3 grid gap-3 sm:grid-cols-3">
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Product</p><p class="font-bold">{{ selectedProduct?.name ?? '-' }}</p></div>
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Location</p><p class="font-bold">{{ selectedLocation?.name ?? '-' }}</p></div>
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Current stock</p><p class="font-bold">{{ currentStock }}</p></div>
          </div>
        </AppCard>
        <AppCard v-if="result">
          <h2 class="font-bold">Latest movement</h2>
          <dl class="mt-3 grid gap-3 sm:grid-cols-3">
            <div><dt class="text-sm text-slate-500">Type</dt><dd class="font-bold">{{ result.reference_type }}</dd></div>
            <div><dt class="text-sm text-slate-500">Before</dt><dd class="font-bold">{{ result.before_stock }}</dd></div>
            <div><dt class="text-sm text-slate-500">After</dt><dd class="font-bold">{{ result.after_stock }}</dd></div>
          </dl>
        </AppCard>
      </div>
    </div>

    <AppModal :open="adjustOpen" title="Stock adjustment" @close="adjustOpen = false">
      <form class="grid gap-3" @submit.prevent="adjustStock">
        <p class="text-sm text-slate-600">Current stock: <b>{{ currentStock }}</b>. Negative adjustments cannot make stock below zero.</p>
        <AppInput v-model="adjustment.quantity" label="Adjustment quantity" type="number" />
        <AppTextarea v-model="adjustment.note" label="Required note" />
        <div class="flex justify-end gap-2">
          <AppButton variant="secondary" @click="adjustOpen = false">Cancel</AppButton>
          <AppButton type="submit">Apply adjustment</AppButton>
        </div>
      </form>
    </AppModal>
  </section>
</template>
