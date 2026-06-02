<script setup lang="ts">
import { computed, defineComponent, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { apiClient, postJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppModal from '../components/AppModal.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import { useCartStore } from '../stores/cart'
import type { Location, POSProduct, Receipt, StockStatus } from '../types/navigation'

const cart = useCartStore()
const locations = ref<Location[]>([])
const products = ref<POSProduct[]>([])
const selectedLocationID = ref('')
const search = ref('')
const barcodeInput = ref('')
const loading = ref(false)
const error = ref('')
const successReceipt = ref<Receipt | null>(null)
const scannerOpen = ref(false)
const scannerMessage = ref('')
const videoRef = ref<HTMLVideoElement | null>(null)
let stream: MediaStream | null = null
let scanFrame = 0
let loadTimer: number | undefined

const selectedLocation = computed(() => locations.value.find((location) => location.id === Number(selectedLocationID.value)) ?? null)
const canSubmit = computed(() => cart.items.length > 0 && cart.receivedAmount >= cart.totalAmount && !cart.isSubmitting)
const receiptPath = computed(() => successReceipt.value ? `/receipt-detail?id=${successReceipt.value.id}` : '/receipt-detail')

const CartPanel = defineComponent({
  components: { AppBadge, AppButton, AppCard, AppInput, AppSelect },
  props: {
    submitDisabled: { type: Boolean, default: false },
  },
  emits: ['submitSale'],
  setup() {
    const cart = useCartStore()
    const money = (value: number) => value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
    const insufficientPayment = computed(() => cart.items.length > 0 && cart.receivedAmount < cart.totalAmount)
    const setPaymentMethod = (value: string) => cart.setPaymentMethod(value === 'QR' ? 'QR' : 'CASH')
    return { cart, money, insufficientPayment, setPaymentMethod }
  },
  template: `
    <AppCard>
      <div class="flex items-center justify-between gap-3">
        <div>
          <h2 class="font-bold">Cart</h2>
          <p class="text-sm text-slate-500">{{ cart.totalItems }} items</p>
        </div>
        <AppButton variant="secondary" :disabled="cart.items.length === 0 || cart.isSubmitting" @click="cart.clearCart()">Clear</AppButton>
      </div>

      <div v-if="cart.items.length === 0" class="mt-4 rounded-lg border border-dashed border-slate-300 p-6 text-center text-sm text-slate-500">
        Add products to start a sale.
      </div>

      <div v-else class="mt-4 grid max-h-[42vh] gap-3 overflow-y-auto pr-1 xl:max-h-[45vh]">
        <article v-for="item in cart.items" :key="item.productId" class="rounded-lg border border-slate-200 p-3">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <h3 class="truncate font-bold">{{ item.name }}</h3>
              <p class="text-xs text-slate-500">{{ item.sku }} · stock {{ item.stock }} {{ item.unit }}</p>
            </div>
            <button class="rounded-md px-2 py-1 text-sm font-semibold text-red-600" @click="cart.removeItem(item.productId)">Remove</button>
          </div>
          <div class="mt-3 grid grid-cols-[1fr_120px] items-end gap-3">
            <div>
              <p class="text-sm text-slate-500">Line total</p>
              <p class="font-bold">{{ money(item.price * item.quantity) }} บาท</p>
            </div>
            <AppInput
              label="Qty"
              type="number"
              :model-value="item.quantity"
              @update:model-value="cart.updateQuantity(item.productId, Number($event))"
            />
          </div>
        </article>
      </div>

      <div class="mt-4 grid gap-3 border-t border-slate-200 pt-4">
        <div class="flex items-center justify-between text-sm">
          <span class="text-slate-500">Total</span>
          <span class="text-xl font-bold">{{ money(cart.totalAmount) }} บาท</span>
        </div>
        <AppSelect
          label="Payment method"
          :model-value="cart.paymentMethod"
          @update:model-value="setPaymentMethod"
        >
          <option value="CASH">CASH</option>
          <option value="QR">QR</option>
        </AppSelect>
        <AppInput
          label="Received amount"
          type="number"
          :model-value="cart.receivedAmount"
          @update:model-value="cart.setReceivedAmount(Number($event))"
        />
        <div class="grid grid-cols-2 gap-3 text-sm">
          <div class="rounded-lg bg-slate-50 p-3">
            <p class="text-slate-500">Change</p>
            <p class="font-bold">{{ money(cart.changeAmount) }} บาท</p>
          </div>
          <div class="rounded-lg bg-slate-50 p-3">
            <p class="text-slate-500">Status</p>
            <AppBadge>{{ insufficientPayment ? 'payment short' : 'ready' }}</AppBadge>
          </div>
        </div>
        <p v-if="insufficientPayment" class="text-sm font-semibold text-red-600">Payment is insufficient.</p>
        <AppButton class="w-full" :disabled="submitDisabled" :loading="cart.isSubmitting" @click="$emit('submitSale')">
          Confirm sale
        </AppButton>
      </div>
    </AppCard>
  `,
})

function money(value: number) {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function stockBadgeClass(status: StockStatus) {
  return {
    in_stock: 'bg-brand-100 text-brand-700',
    low_stock: 'bg-amber-100 text-amber-800',
    out_of_stock: 'bg-red-100 text-red-700',
    reorder_point: 'bg-blue-100 text-blue-700',
  }[status]
}

async function loadLocations() {
  locations.value = await apiClient<Location[]>('/v1/locations')
  const firstActive = locations.value.find((location) => location.is_active) ?? locations.value[0]
  if (!selectedLocationID.value && firstActive) selectedLocationID.value = String(firstActive.id)
}

async function loadProducts() {
  if (!selectedLocationID.value) return
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ location_id: selectedLocationID.value })
    if (search.value.trim()) params.set('q', search.value.trim())
    products.value = await apiClient<POSProduct[]>(`/v1/pos/products?${params.toString()}`)
    cart.refreshStock(products.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load POS products'
  } finally {
    loading.value = false
  }
}

function scheduleLoadProducts() {
  window.clearTimeout(loadTimer)
  loadTimer = window.setTimeout(loadProducts, 180)
}

function addProduct(product: POSProduct) {
  error.value = ''
  if (product.stock <= 0) {
    error.value = `${product.name} stock is not available at this location`
    return
  }
  cart.addItem(product)
  if (cart.paymentMethod === 'QR') cart.setReceivedAmount(cart.totalAmount)
}

function addBarcode() {
  const code = barcodeInput.value.trim()
  if (!code) return
  const product = products.value.find((item) => item.barcode === code || item.sku.toLowerCase() === code.toLowerCase())
  if (!product) {
    search.value = code
    scheduleLoadProducts()
    error.value = 'No matching item loaded yet. Search results updated.'
    return
  }
  addProduct(product)
  barcodeInput.value = ''
}

async function submitSale() {
  if (!selectedLocationID.value || cart.isSubmitting) return
  error.value = ''
  successReceipt.value = null
  cart.isSubmitting = true
  try {
    const receipt = await postJSON<Receipt>('/v1/sales', {
      location_id: Number(selectedLocationID.value),
      payment_method: cart.paymentMethod,
      received_amount: cart.receivedAmount,
      items: cart.items.map((item) => ({ product_id: item.productId, quantity: item.quantity })),
    })
    successReceipt.value = receipt
    cart.clearCart()
    await loadProducts()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Sale failed'
  } finally {
    cart.isSubmitting = false
  }
}

async function openScanner() {
  scannerOpen.value = true
  scannerMessage.value = 'Starting camera...'
  await nextTick()
  const video = videoRef.value
  const detectorCtor = (window as unknown as { BarcodeDetector?: new (options?: object) => { detect(video: HTMLVideoElement): Promise<Array<{ rawValue: string }>> } }).BarcodeDetector
  if (!video || !navigator.mediaDevices?.getUserMedia || !detectorCtor) {
    scannerMessage.value = 'Camera scan is not available in this browser. Manual barcode input is ready.'
    return
  }
  try {
    stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
    video.srcObject = stream
    await video.play()
    const detector = new detectorCtor({ formats: ['ean_13', 'code_128', 'qr_code'] })
    scannerMessage.value = 'Point the camera at a barcode.'
    const scan = async () => {
      if (!scannerOpen.value || !videoRef.value) return
      const codes = await detector.detect(videoRef.value).catch(() => [])
      if (codes[0]?.rawValue) {
        barcodeInput.value = codes[0].rawValue
        addBarcode()
        closeScanner()
        return
      }
      scanFrame = window.requestAnimationFrame(scan)
    }
    scanFrame = window.requestAnimationFrame(scan)
  } catch {
    scannerMessage.value = 'Camera permission was blocked or unavailable. Use manual barcode input instead.'
  }
}

function closeScanner() {
  scannerOpen.value = false
  window.cancelAnimationFrame(scanFrame)
  if (stream) {
    stream.getTracks().forEach((track) => track.stop())
    stream = null
  }
}

watch([selectedLocationID, search], scheduleLoadProducts)
watch(() => cart.paymentMethod, (method) => {
  if (method === 'QR') cart.setReceivedAmount(cart.totalAmount)
})
watch(() => cart.totalAmount, (total) => {
  if (cart.paymentMethod === 'QR') cart.setReceivedAmount(total)
})

onMounted(async () => {
  await loadLocations()
  await loadProducts()
})

onBeforeUnmount(() => {
  window.clearTimeout(loadTimer)
  closeScanner()
})
</script>

<template>
  <section class="pb-24 lg:pb-0">
    <PageHeader title="POS" eyebrow="Sale transaction" description="Search products, scan barcodes, collect payment, and deduct stock from the selected location." icon="shopping-cart" />

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_390px]">
      <div class="grid gap-4">
        <AppCard hover>
          <div class="grid gap-3 lg:grid-cols-[240px_1fr]">
            <AppSelect v-model="selectedLocationID" label="Location">
              <option v-for="location in locations" :key="location.id" :value="String(location.id)" :disabled="!location.is_active">
                {{ location.name }}{{ location.is_active ? '' : ' (inactive)' }}
              </option>
            </AppSelect>
            <AppInput v-model="search" label="Search product" placeholder="Name, SKU, or barcode" />
          </div>
          <div class="mt-3 grid gap-3 md:grid-cols-[1fr_auto]">
            <AppInput v-model="barcodeInput" label="Barcode manual input" placeholder="Scan or type barcode/SKU" @keyup.enter="addBarcode" />
            <div class="flex items-end gap-2">
              <AppButton class="flex-1 md:flex-none" variant="secondary" icon="scan-barcode" @click="addBarcode">Add barcode</AppButton>
              <AppButton class="flex-1 md:flex-none" variant="secondary" icon="qr-code" @click="openScanner">Camera scan</AppButton>
            </div>
          </div>
          <p class="mt-3 text-sm text-slate-500">Selected location: <b>{{ selectedLocation?.name ?? '-' }}</b></p>
        </AppCard>

        <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-semibold text-red-700">{{ error }}</div>
        <AppLoadingState v-if="loading" label="Loading products..." />
        <AppEmptyState v-else-if="products.length === 0" title="No products found" description="Try another name, SKU, barcode, or location." />

        <div v-else class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <article v-for="product in products" :key="product.id" class="premium-card-hover rounded-2xl border border-slate-200 bg-white/80 p-4 shadow-sm">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <h2 class="truncate font-bold">{{ product.name }}</h2>
                <p class="text-xs text-slate-500">{{ product.sku }}<span v-if="product.barcode"> · {{ product.barcode }}</span></p>
              </div>
              <span class="rounded-full px-2 py-1 text-xs font-bold" :class="stockBadgeClass(product.stock_status)">
                {{ product.stock_status.replaceAll('_', ' ') }}
              </span>
            </div>
            <div class="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <p class="text-slate-500">Price</p>
                <p class="text-lg font-bold">{{ money(product.selling_price) }} บาท</p>
              </div>
              <div>
                <p class="text-slate-500">Stock</p>
                <p class="text-lg font-bold">{{ product.stock }} {{ product.unit }}</p>
              </div>
            </div>
            <AppButton class="mt-4 w-full" :disabled="product.stock <= 0" icon="plus" @click="addProduct(product)">
              {{ product.stock <= 0 ? 'Out of stock' : 'Add to cart' }}
            </AppButton>
          </article>
        </div>
      </div>

      <aside class="hidden xl:block">
        <div class="sticky top-24">
          <CartPanel :submit-disabled="!canSubmit" @submit-sale="submitSale" />
        </div>
      </aside>
    </div>

    <div class="mt-4 xl:hidden">
      <CartPanel :submit-disabled="!canSubmit" @submit-sale="submitSale" />
    </div>

    <div class="fixed inset-x-0 bottom-0 z-30 border-t border-slate-200 bg-white/90 p-3 shadow-lg backdrop-blur-xl xl:hidden">
      <div class="mx-auto flex max-w-3xl items-center justify-between gap-3">
        <div>
          <p class="text-xs text-slate-500">Cart</p>
          <p class="font-bold">{{ cart.totalItems }} items · {{ money(cart.totalAmount) }} บาท</p>
        </div>
        <AppButton :disabled="!canSubmit" :loading="cart.isSubmitting" @click="submitSale">
          Confirm
        </AppButton>
      </div>
    </div>

    <AppModal :open="scannerOpen" title="Camera scan" @close="closeScanner">
      <div class="grid gap-3">
        <video ref="videoRef" class="aspect-video w-full rounded-lg bg-slate-950 object-cover" muted playsinline />
        <p class="text-sm text-slate-600">{{ scannerMessage }}</p>
        <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
          <AppInput v-model="barcodeInput" label="Manual fallback" placeholder="Barcode or SKU" @keyup.enter="addBarcode" />
          <div class="flex items-end"><AppButton class="w-full" @click="addBarcode">Add</AppButton></div>
        </div>
      </div>
    </AppModal>

    <AppModal :open="Boolean(successReceipt)" title="Sale completed" @close="successReceipt = null">
      <div v-if="successReceipt" class="grid gap-4">
        <div class="rounded-lg bg-brand-50 p-4">
          <p class="text-sm text-brand-700">Receipt</p>
          <h2 class="text-xl font-bold">{{ successReceipt.receipt_no }}</h2>
        </div>
        <dl class="grid grid-cols-2 gap-3 text-sm">
          <div><dt class="text-slate-500">Total</dt><dd class="font-bold">{{ money(successReceipt.total_amount) }} บาท</dd></div>
          <div><dt class="text-slate-500">Received</dt><dd class="font-bold">{{ money(successReceipt.paid_amount) }} บาท</dd></div>
          <div><dt class="text-slate-500">Change</dt><dd class="font-bold">{{ money(successReceipt.change_amount) }} บาท</dd></div>
          <div><dt class="text-slate-500">Payment</dt><dd class="font-bold">{{ successReceipt.payment_method }}</dd></div>
        </dl>
        <div class="flex justify-end gap-2">
          <AppButton variant="secondary" @click="successReceipt = null">Close</AppButton>
          <RouterLink :to="receiptPath" class="focus-ring inline-flex min-h-11 items-center justify-center rounded-xl bg-brand-600 px-4 py-2 text-sm font-bold text-white shadow-sm shadow-brand-600/20">
            View receipt
          </RouterLink>
        </div>
      </div>
    </AppModal>
  </section>
</template>
