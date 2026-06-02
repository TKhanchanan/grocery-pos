<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, postJSON } from '../api'
import { useCartStore } from '../stores/cart'
import type { Location, Product, Sale } from '../types'

const cart = useCartStore()
const products = ref<Product[]>([])
const locations = ref<Location[]>([])
const locationId = ref(0)
const search = ref('')
const paidAmount = ref(0)
const paymentMethod = ref('CASH')
const error = ref('')
const receipt = ref<Sale | null>(null)

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  return products.value.filter((p) => [p.name, p.sku, p.barcode ?? ''].some((value) => value.toLowerCase().includes(q)))
})

async function load() {
  products.value = await api<Product[]>('/products')
  locations.value = await api<Location[]>('/locations')
  locationId.value = locations.value[0]?.id ?? 0
}

async function checkout() {
  error.value = ''
  try {
    const out = await postJSON<{ id: number }>('/sales', {
      locationId: locationId.value,
      paymentMethod: paymentMethod.value,
      paidAmount: paidAmount.value,
      items: cart.lines.map((line) => ({ productId: line.product.id, quantity: line.quantity })),
    })
    receipt.value = await api<Sale>(`/sales/${out.id}`)
    cart.clear()
    products.value = await api<Product[]>('/products')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Checkout failed'
  }
}

onMounted(load)
</script>

<template>
  <section class="grid gap-5 lg:grid-cols-[1fr_420px]">
    <div class="space-y-3">
      <div class="panel grid gap-3 sm:grid-cols-[1fr_220px]">
        <input v-model="search" class="input" placeholder="Search or scan barcode manually" />
        <select v-model.number="locationId" class="input">
          <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option>
        </select>
      </div>
      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        <button v-for="p in filtered" :key="p.id" class="panel text-left hover:border-leaf" @click="cart.add(p)">
          <p class="font-bold">{{ p.name }}</p>
          <p class="text-sm text-slate-500">{{ p.sku }} · {{ p.barcode ?? 'no barcode' }}</p>
          <p class="mt-3 text-xl font-bold text-leaf">{{ p.price.toFixed(2) }} THB</p>
          <p class="text-xs text-slate-500">Total stock {{ p.totalStock }}</p>
        </button>
      </div>
      <div class="panel">
        <h3 class="font-bold">Camera Barcode Fallback</h3>
        <p class="mt-2 text-sm text-slate-600">Manual barcode input is fully supported. Camera scanning can be connected by sending decoded barcode text into the search field.</p>
      </div>
    </div>

    <aside class="panel sticky top-32 h-fit space-y-4">
      <h2 class="text-xl font-bold">Cart</h2>
      <div v-if="cart.lines.length === 0" class="text-sm text-slate-500">Cart is empty.</div>
      <div v-for="line in cart.lines" :key="line.product.id" class="rounded-md border border-emerald-100 p-3">
        <div class="flex justify-between gap-3">
          <b>{{ line.product.name }}</b>
          <button class="text-sm font-bold text-red-600" @click="cart.remove(line.product.id)">Remove</button>
        </div>
        <div class="mt-2 flex items-center gap-2">
          <input class="input w-24" type="number" :value="line.quantity" @input="cart.setQty(line.product.id, Number(($event.target as HTMLInputElement).value))" />
          <span class="font-bold">{{ (line.product.price * line.quantity).toFixed(2) }}</span>
        </div>
      </div>
      <div class="border-t border-emerald-100 pt-4">
        <div class="flex justify-between text-lg font-bold"><span>Total</span><span>{{ cart.total.toFixed(2) }}</span></div>
        <select v-model="paymentMethod" class="input mt-3">
          <option>CASH</option><option>QR</option><option>CARD</option>
        </select>
        <input v-model.number="paidAmount" class="input mt-3" type="number" step="0.01" placeholder="Paid amount" />
        <p v-if="error" class="mt-3 rounded-md bg-red-50 p-3 text-sm text-red-700">{{ error }}</p>
        <button class="btn-primary mt-3 w-full" :disabled="cart.lines.length === 0" @click="checkout">Confirm Sale</button>
      </div>
      <div v-if="receipt" class="rounded-md bg-mint p-4 text-sm">
        <h3 class="font-bold">Receipt {{ receipt.receiptNo }}</h3>
        <p>{{ receipt.locationName }} · {{ new Date(receipt.createdAt).toLocaleString() }}</p>
        <ul class="mt-2">
          <li v-for="item in receipt.items" :key="item.id">{{ item.productNameSnapshot }} x {{ item.quantity }} = {{ item.lineTotal.toFixed(2) }}</li>
        </ul>
        <p class="mt-2 font-bold">Total {{ receipt.totalAmount.toFixed(2) }} · Change {{ receipt.changeAmount.toFixed(2) }}</p>
      </div>
    </aside>
  </section>
</template>
