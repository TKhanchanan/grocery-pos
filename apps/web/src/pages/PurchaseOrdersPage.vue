<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTextarea from '../components/AppTextarea.vue'
import PageHeader from '../components/PageHeader.vue'
import type { Location, Product, PurchaseOrder, PurchaseOrderItem, Supplier } from '../types/navigation'

const route = useRoute()
const purchaseOrders = ref<PurchaseOrder[]>([])
const suppliers = ref<Supplier[]>([])
const locations = ref<Location[]>([])
const products = ref<Product[]>([])
const selectedPO = ref<PurchaseOrder | null>(null)
const editingID = ref<number | null>(null)
const loading = ref(false)
const error = ref('')

const form = reactive({
  supplier_id: '',
  location_id: '',
  note: '',
  items: [] as PurchaseOrderItem[],
})

const totalCost = computed(() => form.items.reduce((sum, item) => sum + Number(item.quantity || 0) * Number(item.unit_cost || 0), 0))

function money(value: number) {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function statusClass(status: PurchaseOrder['status']) {
  return {
    DRAFT: 'bg-slate-100 text-slate-700',
    SENT: 'bg-blue-100 text-blue-700',
    RECEIVED: 'bg-brand-100 text-brand-700',
    CANCELLED: 'bg-red-100 text-red-700',
  }[status]
}

function defaultItem(): PurchaseOrderItem {
  const product = products.value.find((item) => item.id === Number(route.query.product_id)) ?? products.value[0]
  return {
    product_id: product?.id ?? 0,
    quantity: 1,
    received_quantity: 0,
    unit_cost: product?.unit_cost ?? 0,
    line_cost: product?.unit_cost ?? 0,
  }
}

function resetForm() {
  editingID.value = null
  form.supplier_id = suppliers.value[0] ? String(suppliers.value[0].id) : ''
  form.location_id = route.query.location_id ? String(route.query.location_id) : locations.value[0] ? String(locations.value[0].id) : ''
  form.note = route.query.product_id ? 'Created from reorder alert' : ''
  form.items = [defaultItem()]
}

function addItem() {
  form.items.push(defaultItem())
}

function removeItem(index: number) {
  form.items.splice(index, 1)
  if (form.items.length === 0) addItem()
}

function syncProductCost(item: PurchaseOrderItem) {
  const product = products.value.find((candidate) => candidate.id === Number(item.product_id))
  if (product && !item.unit_cost) item.unit_cost = product.unit_cost
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [poRows, supplierRows, locationRows, productRows] = await Promise.all([
      apiClient<PurchaseOrder[]>('/v1/purchase-orders'),
      apiClient<Supplier[]>('/v1/suppliers'),
      apiClient<Location[]>('/v1/locations'),
      apiClient<Product[]>('/v1/products'),
    ])
    purchaseOrders.value = poRows
    suppliers.value = supplierRows.filter((item) => item.is_active)
    locations.value = locationRows.filter((item) => item.is_active)
    products.value = productRows.filter((item) => item.is_active)
    if (form.items.length === 0) resetForm()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load purchase orders'
  } finally {
    loading.value = false
  }
}

function editPO(po: PurchaseOrder) {
  editingID.value = po.id
  selectedPO.value = po
  form.supplier_id = String(po.supplier_id)
  form.location_id = String(po.location_id)
  form.note = po.note
  form.items = po.items.map((item) => ({ ...item }))
}

async function savePO() {
  error.value = ''
  try {
    const payload = {
      supplier_id: Number(form.supplier_id),
      location_id: Number(form.location_id),
      note: form.note,
      items: form.items.map((item) => ({
        product_id: Number(item.product_id),
        quantity: Number(item.quantity),
        unit_cost: Number(item.unit_cost),
      })),
    }
    if (editingID.value) {
      selectedPO.value = await patchJSON<PurchaseOrder>(`/v1/purchase-orders/${editingID.value}`, payload)
    } else {
      selectedPO.value = await postJSON<PurchaseOrder>('/v1/purchase-orders', payload)
    }
    await load()
    resetForm()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save purchase order'
  }
}

async function showPO(po: PurchaseOrder) {
  selectedPO.value = await apiClient<PurchaseOrder>(`/v1/purchase-orders/${po.id}`)
}

async function actionPO(po: PurchaseOrder, action: 'send' | 'receive' | 'cancel') {
  error.value = ''
  try {
    selectedPO.value = await postJSON<PurchaseOrder>(`/v1/purchase-orders/${po.id}/${action}`, {})
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : `Could not ${action} purchase order`
  }
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="Purchase Orders" eyebrow="Purchasing" description="Create supplier purchase orders and receive items into stock." />

    <div class="grid gap-4 xl:grid-cols-[440px_1fr]">
      <AppCard>
        <form class="grid gap-3" @submit.prevent="savePO">
          <div class="flex items-center justify-between gap-3">
            <h2 class="font-bold">{{ editingID ? 'Edit purchase order' : 'Create purchase order' }}</h2>
            <AppButton variant="secondary" @click="resetForm">Clear</AppButton>
          </div>
          <AppSelect v-model="form.supplier_id" label="Supplier">
            <option value="">Select supplier</option>
            <option v-for="supplier in suppliers" :key="supplier.id" :value="String(supplier.id)">{{ supplier.name }}</option>
          </AppSelect>
          <AppSelect v-model="form.location_id" label="Target location">
            <option value="">Select location</option>
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <AppTextarea v-model="form.note" label="Note" />

          <div class="grid gap-3">
            <div class="flex items-center justify-between gap-3">
              <h3 class="font-bold">Items</h3>
              <AppButton variant="secondary" @click="addItem">Add item</AppButton>
            </div>
            <article v-for="(item, index) in form.items" :key="index" class="rounded-lg border border-slate-200 p-3">
              <div class="grid gap-3 md:grid-cols-[1fr_90px_110px_auto] md:items-end">
                <AppSelect v-model="item.product_id" label="Product" @update:model-value="syncProductCost(item)">
                  <option v-for="product in products" :key="product.id" :value="product.id">{{ product.name }} · {{ product.sku }}</option>
                </AppSelect>
                <AppInput v-model="item.quantity" label="Qty" type="number" />
                <AppInput v-model="item.unit_cost" label="Cost" type="number" />
                <AppButton variant="danger" @click="removeItem(index)">Remove</AppButton>
              </div>
              <p class="mt-2 text-sm text-slate-500">Line cost: {{ money(Number(item.quantity || 0) * Number(item.unit_cost || 0)) }} บาท</p>
            </article>
          </div>

          <div class="rounded-lg bg-slate-50 p-3 text-sm">
            <b>Total:</b> {{ money(totalCost) }} บาท
          </div>
          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <AppButton type="submit">{{ editingID ? 'Update PO' : 'Create PO' }}</AppButton>
        </form>
      </AppCard>

      <div class="grid gap-4">
        <AppCard>
          <div class="flex items-center justify-between gap-3">
            <h2 class="font-bold">PO list</h2>
            <AppButton variant="secondary" @click="load">Refresh</AppButton>
          </div>
          <div v-if="loading" class="mt-4 text-sm text-slate-500">Loading purchase orders...</div>
          <AppEmptyState v-else-if="purchaseOrders.length === 0" class="mt-4" title="No purchase orders" description="Create a PO from the form or a reorder alert." />
          <div v-else class="mt-4 grid gap-3">
            <article v-for="po in purchaseOrders" :key="po.id" class="rounded-lg border border-slate-200 p-3">
              <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="font-bold">{{ po.po_number }}</h3>
                    <span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(po.status)">{{ po.status }}</span>
                  </div>
                  <p class="mt-1 text-sm text-slate-500">{{ po.supplier_name }} · {{ po.location_name }} · {{ money(po.total_cost) }} บาท</p>
                  <p class="mt-1 text-xs text-slate-500">{{ po.items.length }} items</p>
                </div>
                <div class="flex flex-wrap gap-2 md:justify-end">
                  <AppButton variant="secondary" @click="showPO(po)">Detail</AppButton>
                  <AppButton v-if="po.status === 'DRAFT'" variant="secondary" @click="editPO(po)">Edit</AppButton>
                  <AppButton v-if="po.status === 'DRAFT'" @click="actionPO(po, 'send')">Send</AppButton>
                  <AppButton v-if="po.status === 'DRAFT' || po.status === 'SENT'" @click="actionPO(po, 'receive')">Receive</AppButton>
                  <AppButton v-if="po.status === 'DRAFT' || po.status === 'SENT'" variant="danger" @click="actionPO(po, 'cancel')">Cancel</AppButton>
                </div>
              </div>
            </article>
          </div>
        </AppCard>

        <AppCard v-if="selectedPO">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h2 class="font-bold">PO detail: {{ selectedPO.po_number }}</h2>
              <p class="text-sm text-slate-500">{{ selectedPO.supplier_name }} · {{ selectedPO.location_name }}</p>
            </div>
            <span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(selectedPO.status)">{{ selectedPO.status }}</span>
          </div>
          <div class="mt-4 overflow-x-auto">
            <table class="min-w-full divide-y divide-slate-200 text-sm">
              <thead class="bg-slate-50">
                <tr>
                  <th class="px-3 py-2 text-left">Product</th>
                  <th class="px-3 py-2 text-right">Qty</th>
                  <th class="px-3 py-2 text-right">Received</th>
                  <th class="px-3 py-2 text-right">Cost</th>
                  <th class="px-3 py-2 text-right">Line</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="item in selectedPO.items" :key="item.id">
                  <td class="px-3 py-2"><p class="font-semibold">{{ item.product_name }}</p><p class="text-xs text-slate-500">{{ item.sku }}</p></td>
                  <td class="px-3 py-2 text-right">{{ item.quantity }}</td>
                  <td class="px-3 py-2 text-right">{{ item.received_quantity }}</td>
                  <td class="px-3 py-2 text-right">{{ money(item.unit_cost) }}</td>
                  <td class="px-3 py-2 text-right font-semibold">{{ money(item.line_cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </AppCard>
      </div>
    </div>
  </section>
</template>
