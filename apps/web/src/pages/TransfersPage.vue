<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTextarea from '../components/AppTextarea.vue'
import PageHeader from '../components/PageHeader.vue'
import type { Location, Product, ProductStock, StockTransfer } from '../types/navigation'

const transfers = ref<StockTransfer[]>([])
const products = ref<Product[]>([])
const locations = ref<Location[]>([])
const stocks = ref<ProductStock[]>([])
const selectedTransfer = ref<StockTransfer | null>(null)
const loading = ref(false)
const error = ref('')

const form = reactive({
  from_location_id: '',
  to_location_id: '',
  product_id: '',
  quantity: 1,
  note: 'Demo stock transfer',
})

const selectedProduct = computed(() => products.value.find((product) => product.id === Number(form.product_id)) ?? null)
const sourceStock = computed(() => stocks.value.find((stock) => stock.product_id === Number(form.product_id) && stock.location_id === Number(form.from_location_id))?.quantity ?? 0)
const afterSource = computed(() => sourceStock.value - Number(form.quantity || 0))

async function load() {
  loading.value = true
  error.value = ''
  try {
    transfers.value = await apiClient<StockTransfer[]>('/v1/stock-transfers')
    products.value = await apiClient<Product[]>('/v1/products')
    locations.value = await apiClient<Location[]>('/v1/locations')
    stocks.value = await apiClient<ProductStock[]>('/v1/product-stocks')
    if (!form.from_location_id && locations.value[0]) form.from_location_id = String(locations.value[0].id)
    if (!form.to_location_id && locations.value[1]) form.to_location_id = String(locations.value[1].id)
    if (!form.product_id && products.value[0]) form.product_id = String(products.value[0].id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load transfers'
  } finally {
    loading.value = false
  }
}

async function createTransfer() {
  error.value = ''
  try {
    selectedTransfer.value = await postJSON<StockTransfer>('/v1/stock-transfers', {
      from_location_id: Number(form.from_location_id),
      to_location_id: Number(form.to_location_id),
      note: form.note,
      items: [{ product_id: Number(form.product_id), quantity: Number(form.quantity) }],
    })
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not create transfer'
  }
}

async function completeTransfer(transfer: StockTransfer) {
  error.value = ''
  try {
    selectedTransfer.value = await postJSON<StockTransfer>(`/v1/stock-transfers/${transfer.id}/complete`, {})
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not complete transfer'
  }
}

async function cancelTransfer(transfer: StockTransfer) {
  selectedTransfer.value = await postJSON<StockTransfer>(`/v1/stock-transfers/${transfer.id}/cancel`, {})
  await load()
}

async function showTransfer(transfer: StockTransfer) {
  selectedTransfer.value = await apiClient<StockTransfer>(`/v1/stock-transfers/${transfer.id}`)
}

function statusClass(status: StockTransfer['status']) {
  return {
    DRAFT: 'bg-amber-100 text-amber-800',
    COMPLETED: 'bg-brand-100 text-brand-700',
    CANCELLED: 'bg-slate-100 text-slate-600',
  }[status]
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="Transfers" eyebrow="Multi-location stock" description="Move stock between locations with source validation and movement history." />

    <div class="grid gap-4 xl:grid-cols-[420px_1fr]">
      <AppCard>
        <form class="grid gap-3" @submit.prevent="createTransfer">
          <h2 class="font-bold">Create transfer</h2>
          <AppSelect v-model="form.from_location_id" label="Source location">
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <AppSelect v-model="form.to_location_id" label="Destination location">
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <AppSelect v-model="form.product_id" label="Product">
            <option v-for="product in products" :key="product.id" :value="String(product.id)">{{ product.name }} · {{ product.sku }}</option>
          </AppSelect>
          <div class="rounded-lg bg-slate-50 p-3 text-sm">
            <p><b>Available source stock:</b> {{ sourceStock }} {{ selectedProduct?.unit }}</p>
            <p><b>Source after transfer:</b> {{ afterSource }} {{ selectedProduct?.unit }}</p>
          </div>
          <AppInput v-model="form.quantity" label="Quantity" type="number" />
          <AppTextarea v-model="form.note" label="Note" />
          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <AppButton type="submit">Create draft transfer</AppButton>
        </form>
      </AppCard>

      <div class="grid gap-4">
        <AppCard>
          <div class="flex items-center justify-between gap-3">
            <h2 class="font-bold">Transfer list</h2>
            <AppButton variant="secondary" @click="load">Refresh</AppButton>
          </div>
          <div v-if="loading" class="mt-4 text-sm text-slate-500">Loading transfers...</div>
          <AppEmptyState v-else-if="transfers.length === 0" class="mt-4" title="No transfers" description="Create a transfer draft from the form." />

          <div v-else>
            <div class="mt-4 hidden overflow-x-auto md:block">
              <table class="min-w-full divide-y divide-slate-200 text-sm">
                <thead class="bg-slate-50">
                  <tr>
                    <th class="px-3 py-2 text-left">No.</th>
                    <th class="px-3 py-2 text-left">Route</th>
                    <th class="px-3 py-2 text-left">Items</th>
                    <th class="px-3 py-2 text-left">Status</th>
                    <th class="px-3 py-2 text-left">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="transfer in transfers" :key="transfer.id">
                    <td class="px-3 py-2 font-semibold">{{ transfer.transfer_no }}</td>
                    <td class="px-3 py-2">{{ transfer.from_location_name }} → {{ transfer.to_location_name }}</td>
                    <td class="px-3 py-2">
                      <span v-for="item in transfer.items" :key="item.id">{{ item.product_name }} x {{ item.quantity }}</span>
                    </td>
                    <td class="px-3 py-2"><span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(transfer.status)">{{ transfer.status }}</span></td>
                    <td class="px-3 py-2">
                      <div class="flex flex-wrap gap-2">
                        <AppButton variant="secondary" @click="showTransfer(transfer)">Detail</AppButton>
                        <AppButton v-if="transfer.status === 'DRAFT'" @click="completeTransfer(transfer)">Complete</AppButton>
                        <AppButton v-if="transfer.status === 'DRAFT'" variant="danger" @click="cancelTransfer(transfer)">Cancel</AppButton>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="mt-4 grid gap-3 md:hidden">
              <article v-for="transfer in transfers" :key="transfer.id" class="rounded-lg border border-slate-200 p-3">
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <h3 class="font-bold">{{ transfer.transfer_no }}</h3>
                    <p class="text-sm text-slate-500">{{ transfer.from_location_name }} → {{ transfer.to_location_name }}</p>
                  </div>
                  <span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(transfer.status)">{{ transfer.status }}</span>
                </div>
                <p class="mt-2 text-sm" v-for="item in transfer.items" :key="item.id">{{ item.product_name }} x {{ item.quantity }}</p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <AppButton variant="secondary" @click="showTransfer(transfer)">Detail</AppButton>
                  <AppButton v-if="transfer.status === 'DRAFT'" @click="completeTransfer(transfer)">Complete</AppButton>
                </div>
              </article>
            </div>
          </div>
        </AppCard>

        <AppCard v-if="selectedTransfer">
          <h2 class="font-bold">Transfer detail: {{ selectedTransfer.transfer_no }}</h2>
          <p class="mt-1 text-sm text-slate-500">{{ selectedTransfer.from_location_name }} → {{ selectedTransfer.to_location_name }}</p>
          <div class="mt-4 grid gap-3 sm:grid-cols-3">
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Status</p><p class="font-bold">{{ selectedTransfer.status }}</p></div>
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Created</p><p class="font-bold">{{ new Date(selectedTransfer.created_at).toLocaleDateString() }}</p></div>
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Items</p><p class="font-bold">{{ selectedTransfer.items.length }}</p></div>
          </div>
          <ul class="mt-4 grid gap-2 text-sm">
            <li v-for="item in selectedTransfer.items" :key="item.id" class="rounded-md bg-slate-50 p-3">
              <b>{{ item.product_name }}</b> · {{ item.sku }} · quantity {{ item.quantity }}
            </li>
          </ul>
        </AppCard>
      </div>
    </div>
  </section>
</template>
