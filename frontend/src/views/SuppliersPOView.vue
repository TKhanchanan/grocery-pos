<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, postJSON } from '../api'
import type { Location, Product, PurchaseOrder, Supplier } from '../types'

const suppliers = ref<Supplier[]>([])
const products = ref<Product[]>([])
const locations = ref<Location[]>([])
const orders = ref<PurchaseOrder[]>([])
const supplier = reactive({ id: 0, name: '', phone: '', email: '', address: '' })
const po = reactive({ supplierId: 0, locationId: 0, productId: 0, quantity: 10, unitCost: 1 })

async function load() {
  suppliers.value = await api<Supplier[]>('/suppliers')
  products.value = await api<Product[]>('/products')
  locations.value = await api<Location[]>('/locations')
  orders.value = await api<PurchaseOrder[]>('/purchase-orders')
  if (!po.supplierId && suppliers.value[0]) po.supplierId = suppliers.value[0].id
  if (!po.locationId && locations.value[0]) po.locationId = locations.value[0].id
  if (!po.productId && products.value[0]) po.productId = products.value[0].id
}

async function saveSupplier() {
  await postJSON('/suppliers', supplier)
  Object.assign(supplier, { id: 0, name: '', phone: '', email: '', address: '' })
  await load()
}

async function createPO() {
  await postJSON('/purchase-orders', {
    supplierId: po.supplierId,
    locationId: po.locationId,
    items: [{ productId: po.productId, quantity: po.quantity, unitCost: po.unitCost }],
  })
  await load()
}

async function receive(id: number) {
  await postJSON(`/purchase-orders/${id}/receive`, {})
  await load()
}

onMounted(load)
</script>

<template>
  <section class="grid gap-5 lg:grid-cols-[360px_1fr]">
    <div class="space-y-4">
      <form class="panel space-y-3" @submit.prevent="saveSupplier">
        <h2 class="font-bold">Supplier</h2>
        <input v-model="supplier.name" class="input" placeholder="Name" />
        <input v-model="supplier.phone" class="input" placeholder="Phone" />
        <input v-model="supplier.email" class="input" placeholder="Email" />
        <input v-model="supplier.address" class="input" placeholder="Address" />
        <button class="btn-primary w-full">Save Supplier</button>
      </form>
      <form class="panel space-y-3" @submit.prevent="createPO">
        <h2 class="font-bold">Create Purchase Order</h2>
        <select v-model.number="po.supplierId" class="input"><option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option></select>
        <select v-model.number="po.locationId" class="input"><option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option></select>
        <select v-model.number="po.productId" class="input"><option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option></select>
        <div class="grid grid-cols-2 gap-3"><input v-model.number="po.quantity" class="input" type="number" /><input v-model.number="po.unitCost" class="input" type="number" step="0.01" /></div>
        <button class="btn-primary w-full">Create PO</button>
      </form>
    </div>
    <div class="space-y-4">
      <div class="table-wrap">
        <table class="table">
          <thead><tr><th>Supplier</th><th>Phone</th><th>Email</th></tr></thead>
          <tbody><tr v-for="s in suppliers" :key="s.id"><td>{{ s.name }}</td><td>{{ s.phone }}</td><td>{{ s.email }}</td></tr></tbody>
        </table>
      </div>
      <div class="table-wrap">
        <table class="table">
          <thead><tr><th>PO</th><th>Supplier</th><th>Location</th><th>Total</th><th>Status</th><th></th></tr></thead>
          <tbody>
            <tr v-for="order in orders" :key="order.id">
              <td>{{ order.poNumber }}</td><td>{{ order.supplier }}</td><td>{{ order.location }}</td><td>{{ order.totalCost.toFixed(2) }}</td><td>{{ order.status }}</td>
              <td><button v-if="order.status === 'OPEN'" class="btn-soft" @click="receive(order.id)">Receive to Stock</button></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>
