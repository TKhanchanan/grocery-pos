<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { apiClient, postJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppIcon from '../components/AppIcon.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppModal from '../components/AppModal.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import POSCartPanel from '../components/POSCartPanel.vue'
import ProductAvatar from '../components/ProductAvatar.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import { useCartStore } from '../stores/cart'
import type { Location, POSProduct, Receipt, StockStatus } from '../types/navigation'

const app = useAppStore()
const auth = useAuthStore()
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
const scannerMessageKey = ref<TranslationKey>('pos.scannerStarting')
const videoRef = ref<HTMLVideoElement | null>(null)
let stream: MediaStream | null = null
let scanFrame = 0
let loadTimer: number | undefined

const selectedLocation = computed(() => locations.value.find((location) => location.id === Number(selectedLocationID.value)) ?? null)
const canSubmit = computed(() => cart.items.length > 0 && cart.receivedAmount >= cart.totalAmount && !cart.isSubmitting)
const canViewReceipt = computed(() => auth.hasPermission('sales.receipt.view'))
const receiptPath = computed(() => successReceipt.value ? `/sales/${successReceipt.value.id}/receipt` : '/receipt-detail')
const scannerMessage = computed(() => app.t(scannerMessageKey.value))
const productCountLabel = computed(() => t('pos.productsCount', { count: products.value.length }))

function money(value: number) {
  const locale = app.language === 'th' ? 'th-TH' : 'en-US'
  const amount = value.toLocaleString(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return t('pos.currency', { amount })
}

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) {
    text = text.replaceAll(`{${name}}`, String(value))
  }
  return text
}

function stockTone(status: StockStatus) {
  return {
    in_stock: 'success',
    low_stock: 'warning',
    out_of_stock: 'danger',
    reorder_point: 'info',
  }[status] as 'success' | 'warning' | 'danger' | 'info'
}

function stockLabel(status: StockStatus) {
  const labels: Record<StockStatus, TranslationKey> = {
    in_stock: 'pos.inStock',
    low_stock: 'pos.lowStock',
    out_of_stock: 'pos.outOfStock',
    reorder_point: 'pos.reorderPoint',
  }
  return app.t(labels[status])
}

function paymentLabel(method: string) {
  if (method === 'QR') return app.t('pos.qr')
  return app.t('pos.cash')
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
    error.value = err instanceof Error ? err.message : app.t('pos.noProducts')
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
    error.value = t('pos.stockUnavailable', { name: product.name })
    return
  }
  cart.addItem(product)
  app.pushToast({ type: 'success', message: app.t('pos.addedToCart'), description: product.name, resultModal: false })
  if (cart.paymentMethod === 'QR') cart.setReceivedAmount(cart.totalAmount)
  if (cart.paymentMethod === 'CASH' && cart.receivedAmount < cart.totalAmount) cart.setReceivedAmount(cart.totalAmount)
}

function addBarcode() {
  const code = barcodeInput.value.trim()
  if (!code) return
  const product = products.value.find((item) => item.barcode === code || item.sku.toLowerCase() === code.toLowerCase())
  if (!product) {
    search.value = code
    scheduleLoadProducts()
    error.value = app.t('pos.productNotFoundDescription')
    app.pushToast({ type: 'warning', message: app.t('pos.productNotFound'), description: code })
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
    app.pushToast({ type: 'success', message: app.t('pos.saleCompleted'), description: t('pos.saleSuccessDescription', { receipt: receipt.receipt_no }) })
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('pos.saleFailed')
    app.pushToast({ type: 'error', message: app.t('pos.saleFailed'), description: error.value })
  } finally {
    cart.isSubmitting = false
  }
}

async function openScanner() {
  scannerOpen.value = true
  scannerMessageKey.value = 'pos.scannerStarting'
  await nextTick()
  const video = videoRef.value
  const detectorCtor = (window as unknown as { BarcodeDetector?: new (options?: object) => { detect(video: HTMLVideoElement): Promise<Array<{ rawValue: string }>> } }).BarcodeDetector
  if (!video || !navigator.mediaDevices?.getUserMedia || !detectorCtor) {
    scannerMessageKey.value = 'pos.scannerUnavailable'
    return
  }
  try {
    stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
    video.srcObject = stream
    await video.play()
    const detector = new detectorCtor({ formats: ['ean_13', 'code_128', 'qr_code'] })
    scannerMessageKey.value = 'pos.scannerReady'
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
    scannerMessageKey.value = 'pos.scannerBlocked'
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
  if (cart.paymentMethod === 'CASH' && total > 0 && cart.receivedAmount < total) cart.setReceivedAmount(total)
})

onMounted(async () => {
  await loadLocations()
  await loadProducts()
})

onBeforeUnmount(() => {
  window.clearTimeout(loadTimer)
  closeScanner()
  cart.clearCart()
})
</script>

<template>
  <section class="w-full pb-24 lg:pb-0">
    <PageHeader :title="app.t('pos.title')" :eyebrow="app.t('pos.eyebrow')" :description="app.t('pos.description')" icon="shopping-cart" />

    <div class="grid w-full items-start gap-4 xl:grid-cols-[minmax(0,1fr)_420px] 2xl:grid-cols-[minmax(0,1fr)_460px]">
      <div class="grid gap-4">
        <AppCard hover class="dark:bg-slate-900/80">
          <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('pos.controlTitle') }}</p>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ app.t('pos.controlDescription') }}</p>
            </div>
            <div class="rounded-full bg-brand-100 px-3 py-1 text-xs font-black text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-100">
              {{ app.t('pos.locationSelected') }}: {{ selectedLocation?.name ?? '-' }}
            </div>
          </div>
          <div class="grid gap-3 lg:grid-cols-[240px_1fr]">
            <AppSelect v-model="selectedLocationID" :label="app.t('pos.location')" hide-arrow>
              <option v-for="location in locations" :key="location.id" :value="String(location.id)" :disabled="!location.is_active">
                {{ location.name }}
              </option>
            </AppSelect>
            <AppInput v-model="search" :label="app.t('pos.search')" :placeholder="app.t('pos.searchPlaceholder')" />
          </div>
          <div class="mt-3 grid gap-3 md:grid-cols-[1fr_auto]">
            <AppInput v-model="barcodeInput" :label="app.t('pos.barcode')" :placeholder="app.t('pos.barcodePlaceholder')" @keyup.enter="addBarcode" />
            <div class="flex items-end gap-2">
              <AppButton class="flex-1 md:flex-none" variant="secondary" icon="scan-barcode" @click="addBarcode">{{ app.t('pos.addBarcode') }}</AppButton>
              <AppButton class="flex-1 md:flex-none" variant="secondary" icon="qr-code" @click="openScanner">{{ app.t('pos.cameraScan') }}</AppButton>
            </div>
          </div>
        </AppCard>

        <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
        <AppLoadingState v-if="loading" :label="app.t('pos.loadingProducts')" />
        <AppEmptyState v-else-if="products.length === 0" :title="app.t('pos.noProducts')" :description="app.t('pos.noProductsDescription')" />

        <div v-else class="grid gap-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300 mb-1">{{ app.t('pos.productsTitle') }}</p>
              <p class="text-sm text-slate-500 dark:text-slate-400">{{ productCountLabel }}</p>
            </div>
          </div>

          <div class="grid w-full gap-3 sm:grid-cols-2 2xl:grid-cols-3">
            <article
              v-for="product in products"
              :key="product.id"
              class="pos-product-card premium-card-hover flex h-full flex-col rounded-2xl border border-slate-200 bg-white/80 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-950/60"
              :class="product.stock <= 0 ? 'opacity-65 grayscale' : ''"
            >
              <div class="mb-3 aspect-[4/3] w-full overflow-hidden">
                <ProductAvatar :src="product.image_url" :updated-at="product.image_updated_at" :name="product.name" size="full" shape="square" :muted="product.stock <= 0" />
              </div>
              <div class="flex items-start gap-3">
                <div class="flex min-w-0 flex-1 items-start gap-3">
                  <div class="min-w-0">
                    <h2 class="pos-product-name font-black">{{ product.name }}</h2>
                    <p class="truncate text-xs text-slate-500 dark:text-slate-400">{{ product.sku }}<span v-if="product.barcode"> · {{ product.barcode }}</span></p>
                  </div>
                </div>
                <AppBadge :tone="stockTone(product.stock_status)">{{ stockLabel(product.stock_status) }}</AppBadge>
              </div>

              <div class="mt-4 grid grid-cols-2 gap-3 text-sm">
                <div class="flex min-h-20 flex-col justify-between rounded-xl bg-slate-50 p-3 dark:bg-slate-950/60">
                  <p class="text-slate-500 dark:text-slate-400">{{ app.t('pos.price') }}</p>
                  <p class="text-lg font-black">{{ money(product.selling_price) }}</p>
                </div>
                <div class="flex min-h-20 flex-col justify-between rounded-xl bg-slate-50 p-3 dark:bg-slate-950/60">
                  <p class="text-slate-500 dark:text-slate-400">{{ app.t('pos.stock') }}</p>
                  <p class="truncate text-lg font-black">{{ t('pos.stockLine', { stock: product.stock, unit: product.unit }) }}</p>
                </div>
              </div>
              <div class="mt-auto pt-4">
                <AppButton class="w-full" :disabled="product.stock <= 0" icon="plus" @click="addProduct(product)">
                  {{ product.stock <= 0 ? app.t('pos.outOfStock') : app.t('pos.addToCart') }}
                </AppButton>
              </div>
            </article>
          </div>
        </div>
      </div>

      <aside class="hidden xl:block">
        <div class="sticky top-24">
          <POSCartPanel :submit-disabled="!canSubmit" @submit-sale="submitSale" />
        </div>
      </aside>
    </div>

    <div class="mt-4 xl:hidden">
      <POSCartPanel :submit-disabled="!canSubmit" @submit-sale="submitSale" />
    </div>

    <div class="fixed inset-x-0 bottom-0 z-30 border-t border-slate-200 bg-white/90 p-3 shadow-lg backdrop-blur-xl dark:border-slate-700 dark:bg-slate-950/90 xl:hidden">
      <div class="mx-auto flex max-w-3xl items-center justify-between gap-3">
        <div>
          <p class="text-xs text-slate-500 dark:text-slate-400">{{ app.t('pos.cart') }}</p>
          <p class="font-bold">{{ t('pos.mobileSummary', { count: cart.totalItems, amount: money(cart.totalAmount) }) }}</p>
        </div>
        <AppButton :disabled="!canSubmit" :loading="cart.isSubmitting" @click="submitSale">
          {{ app.t('pos.confirm') }}
        </AppButton>
      </div>
    </div>

    <AppModal :open="scannerOpen" :title="app.t('pos.cameraTitle')" @close="closeScanner">
      <div class="grid gap-3">
        <video ref="videoRef" class="aspect-video w-full rounded-lg bg-slate-950 object-cover" muted playsinline />
        <p class="text-sm text-slate-600 dark:text-slate-300">{{ scannerMessage }}</p>
        <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
          <AppInput v-model="barcodeInput" :label="app.t('pos.manualFallback')" :placeholder="app.t('pos.barcodePlaceholder')" @keyup.enter="addBarcode" />
          <div class="flex items-end"><AppButton class="w-full" @click="addBarcode">{{ app.t('pos.add') }}</AppButton></div>
        </div>
      </div>
    </AppModal>

    <AppModal :open="Boolean(successReceipt)" :title="app.t('pos.saleCompleted')" @close="successReceipt = null">
      <div v-if="successReceipt" class="grid gap-4">
        <div class="rounded-lg bg-brand-50 p-4 dark:bg-emerald-500/10">
          <p class="text-sm text-brand-700 dark:text-emerald-200">{{ app.t('pos.receipt') }}</p>
          <h2 class="text-xl font-bold">{{ successReceipt.receipt_no }}</h2>
        </div>
        <dl class="grid grid-cols-2 gap-3 text-sm">
          <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('pos.total') }}</dt><dd class="font-bold">{{ money(successReceipt.total_amount) }}</dd></div>
          <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('pos.received') }}</dt><dd class="font-bold">{{ money(successReceipt.paid_amount) }}</dd></div>
          <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('pos.change') }}</dt><dd class="font-bold">{{ money(successReceipt.change_amount) }}</dd></div>
          <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('pos.payment') }}</dt><dd class="font-bold">{{ paymentLabel(successReceipt.payment_method) }}</dd></div>
        </dl>
        <div class="flex justify-end gap-2">
          <AppButton variant="secondary" @click="successReceipt = null">{{ app.t('pos.continueSelling') }}</AppButton>
          <RouterLink v-if="canViewReceipt" :to="receiptPath" class="focus-ring inline-flex min-h-11 items-center justify-center rounded-xl bg-brand-600 px-4 py-2 text-sm font-bold text-white shadow-sm shadow-brand-600/20 dark:bg-emerald-500 dark:text-slate-950">
            {{ app.t('pos.viewReceipt') }}
          </RouterLink>
        </div>
      </div>
    </AppModal>
  </section>
</template>

<style scoped>
.pos-product-name {
  display: -webkit-box;
  min-height: 2.5rem;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
</style>
