<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
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
import { useReferenceDataStore } from '../stores/referenceData'
import type { Category, Location, POSProduct, POSProductPage, Receipt, StockStatus } from '../types/navigation'
import { formatAppDateTime } from '../utils/date'
import { prepareReceiptPrintArea, resetReceiptPrintArea } from '../utils/print'
import { defaultReceiptSettings } from '../utils/receiptSettings'

const app = useAppStore()
const auth = useAuthStore()
const cart = useCartStore()
const referenceData = useReferenceDataStore()
const { locations, posCategories: categories, receiptSettings } = storeToRefs(referenceData)
const products = ref<POSProduct[]>([])
const selectedLocationID = ref('')
const selectedCategoryID = ref('')
const search = ref('')
const barcodeInput = ref('')
const productPage = ref(1)
const productPageSize = ref(12)
const productTotal = ref(0)
const productTotalPages = ref(1)
const loading = ref(false)
const error = ref('')
const successReceipt = ref<Receipt | null>(null)
const saleConfirmOpen = ref(false)
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
const productCountLabel = computed(() => t('pos.productsCount', { count: productTotal.value }))
const productPageLabel = computed(() => t('pos.page', { page: productPage.value, total: productTotalPages.value }))
const receiptPreviewDate = computed(() => saleConfirmOpen.value ? formatAppDateTime(new Date(), app.language) : '-')
const receiptShopName = computed(() => receiptSettings.value.shop_name.trim() || defaultReceiptSettings.shop_name)
const receiptShopPhone = computed(() => receiptSettings.value.shop_phone.trim())
const receiptShopAddress = computed(() => receiptSettings.value.shop_address.trim())
const receiptFooter = computed(() => receiptSettings.value.receipt_footer.trim() || app.t('receipt.thankYou'))
const activeCategories = computed(() => categories.value.filter((category) => category.is_active))
const selectedCategoryName = computed(() => {
  if (!selectedCategoryID.value) return app.t('pos.allCategories')
  return categories.value.find((category) => category.id === Number(selectedCategoryID.value))?.name ?? app.t('pos.allCategories')
})

const handleBeforePrint = () => prepareReceiptPrintArea()
const handleAfterPrint = () => resetReceiptPrintArea()

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
  await referenceData.loadLocations()
}

async function loadCategories() {
  await referenceData.loadPOSCategories().catch(() => [])
}

async function loadReceiptProfile() {
  await referenceData.loadReceiptSettings().catch(() => ({ ...defaultReceiptSettings }))
}

function selectInitialLocation() {
  const currentLocation = locations.value.find((location) => location.id === Number(selectedLocationID.value) && location.is_active)
  if (currentLocation) return
  const defaultLocation = locations.value.find((location) => location.id === receiptSettings.value.default_location_id && location.is_active)
  const firstActive = locations.value.find((location) => location.is_active)
  selectedLocationID.value = defaultLocation || firstActive ? String((defaultLocation ?? firstActive)!.id) : ''
}

function readProductPage(result: POSProductPage | POSProduct[]) {
  if (Array.isArray(result)) {
    products.value = result
    productTotal.value = result.length
    productTotalPages.value = 1
    return
  }
  products.value = result.items
  productTotal.value = result.total
  productTotalPages.value = Math.max(result.total_pages, 1)
  productPage.value = result.page
  productPageSize.value = result.page_size
}

async function loadProducts() {
  if (!selectedLocationID.value) return
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      location_id: selectedLocationID.value,
      page: String(productPage.value),
      page_size: String(productPageSize.value),
    })
    if (search.value.trim()) params.set('q', search.value.trim())
    if (selectedCategoryID.value) params.set('category_id', selectedCategoryID.value)
    readProductPage(await apiClient<POSProductPage | POSProduct[]>(`/v1/pos/products?${params.toString()}`))
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

function selectCategory(id: string) {
  selectedCategoryID.value = id
}

function previousProductPage() {
  if (productPage.value <= 1) return
  productPage.value -= 1
}

function nextProductPage() {
  if (productPage.value >= productTotalPages.value) return
  productPage.value += 1
}

function setProductPageSize(value: string) {
  productPageSize.value = Number(value) || 12
}

async function findBarcodeProduct(code: string) {
  if (!selectedLocationID.value) return null
  const params = new URLSearchParams({
    location_id: selectedLocationID.value,
    q: code,
    page: '1',
    page_size: '120',
  })
  const result = await apiClient<POSProductPage | POSProduct[]>(`/v1/pos/products?${params.toString()}`)
  const candidates = Array.isArray(result) ? result : result.items
  return candidates.find((item) => item.barcode === code || item.sku.toLowerCase() === code.toLowerCase()) ?? null
}

function addProduct(product: POSProduct) {
  error.value = ''
  if (product.stock <= 0) {
    error.value = t('pos.stockUnavailable', { name: product.name })
    return
  }
  cart.addItem(product)
  app.pushToast({ type: 'success', message: app.t('pos.addedToCart'), description: product.name })
  if (cart.paymentMethod === 'QR') cart.setReceivedAmount(cart.totalAmount)
  if (cart.paymentMethod === 'CASH' && cart.receivedAmount < cart.totalAmount) cart.setReceivedAmount(cart.totalAmount)
}

async function addBarcode() {
  const code = barcodeInput.value.trim()
  if (!code) return
  const product = products.value.find((item) => item.barcode === code || item.sku.toLowerCase() === code.toLowerCase()) ?? await findBarcodeProduct(code)
  if (!product) {
    search.value = code
    productPage.value = 1
    scheduleLoadProducts()
    error.value = app.t('pos.productNotFoundDescription')
    app.pushToast({ type: 'warning', message: app.t('pos.productNotFound'), description: code })
    return
  }
  addProduct(product)
  barcodeInput.value = ''
}

function openSaleConfirm() {
  if (!canSubmit.value) return
  error.value = ''
  saleConfirmOpen.value = true
}

function closeSaleConfirm() {
  if (cart.isSubmitting) return
  saleConfirmOpen.value = false
}

function printReceiptPreview() {
  if (!canSubmit.value || cart.isSubmitting) return
  prepareReceiptPrintArea()
  window.print()
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
    saleConfirmOpen.value = false
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
        await addBarcode()
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

watch(selectedLocationID, () => {
  productPage.value = 1
  scheduleLoadProducts()
})
watch(selectedCategoryID, () => {
  productPage.value = 1
  scheduleLoadProducts()
})
watch(search, () => {
  productPage.value = 1
  scheduleLoadProducts()
})
watch(productPageSize, () => {
  productPage.value = 1
  scheduleLoadProducts()
})
watch(productPage, scheduleLoadProducts)
watch(() => cart.paymentMethod, (method) => {
  if (method === 'QR') cart.setReceivedAmount(cart.totalAmount)
})
watch(() => cart.totalAmount, (total) => {
  if (cart.paymentMethod === 'QR') cart.setReceivedAmount(total)
  if (cart.paymentMethod === 'CASH' && total > 0 && cart.receivedAmount < total) cart.setReceivedAmount(total)
})

onMounted(async () => {
  window.addEventListener('beforeprint', handleBeforePrint)
  window.addEventListener('afterprint', handleAfterPrint)
  await Promise.all([loadLocations(), loadCategories(), loadReceiptProfile()])
  selectInitialLocation()
  await loadProducts()
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeprint', handleBeforePrint)
  window.removeEventListener('afterprint', handleAfterPrint)
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
            <div class="grid min-w-36 gap-1">
              <AppSelect :model-value="productPageSize" :label="app.t('pos.pageSize')" hide-arrow @update:model-value="setProductPageSize">
                <option :value="12">12</option>
                <option :value="24">24</option>
                <option :value="48">48</option>
              </AppSelect>
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

          <div class="flex flex-col gap-3 rounded-2xl bg-white/80 p-3 shadow-sm dark:bg-slate-950/60 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-center text-sm font-bold text-slate-600 dark:text-slate-300 sm:text-left">{{ productPageLabel }}</p>
            <div class="grid grid-cols-2 gap-2 sm:flex sm:items-center">
              <AppButton variant="secondary" icon="chevron-left" :disabled="productPage <= 1 || loading" @click="previousProductPage">{{ app.t('pos.previous') }}</AppButton>
              <AppButton variant="secondary" icon="chevron-right" :disabled="productPage >= productTotalPages || loading" @click="nextProductPage">{{ app.t('pos.next') }}</AppButton>
            </div>
          </div>
        </div>
      </div>

      <aside class="hidden xl:block">
        <div class="sticky top-24">
          <POSCartPanel :submit-disabled="!canSubmit" @submit-sale="openSaleConfirm" />
        </div>
      </aside>
    </div>

    <div class="mt-4 xl:hidden">
      <POSCartPanel :submit-disabled="!canSubmit" @submit-sale="openSaleConfirm" />
    </div>

    <div class="fixed inset-x-0 bottom-0 z-30 border-t border-slate-200 bg-white/90 p-3 shadow-lg backdrop-blur-xl dark:border-slate-700 dark:bg-slate-950/90 xl:hidden">
      <div class="mx-auto flex max-w-3xl items-center justify-between gap-3">
        <div>
          <p class="text-xs text-slate-500 dark:text-slate-400">{{ app.t('pos.cart') }}</p>
          <p class="font-bold">{{ t('pos.mobileSummary', { count: cart.totalItems, amount: money(cart.totalAmount) }) }}</p>
        </div>
        <AppButton :disabled="!canSubmit" :loading="cart.isSubmitting" @click="openSaleConfirm">
          {{ app.t('pos.confirm') }}
        </AppButton>
      </div>
    </div>

    <AppModal :open="scannerOpen" :title="app.t('pos.cameraTitle')" size="xl" @close="closeScanner">
      <div class="grid gap-3">
        <video ref="videoRef" class="aspect-video w-full rounded-lg bg-slate-950 object-cover" muted playsinline />
        <p class="text-sm text-slate-600 dark:text-slate-300">{{ scannerMessage }}</p>
        <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
          <AppInput v-model="barcodeInput" :label="app.t('pos.manualFallback')" :placeholder="app.t('pos.barcodePlaceholder')" @keyup.enter="addBarcode" />
          <div class="flex items-end"><AppButton class="w-full" @click="addBarcode">{{ app.t('pos.add') }}</AppButton></div>
        </div>
      </div>
    </AppModal>

    <AppModal :open="saleConfirmOpen" :title="app.t('pos.confirmSaleTitle')" :description="app.t('pos.confirmSaleDescription')" size="lg" @close="closeSaleConfirm">
      <div class="grid gap-4">
        <article id="receipt-print-area" class="mx-auto w-full max-w-[420px] rounded-xl bg-white p-4 text-slate-950 shadow-sm ring-1 ring-slate-200">
          <header class="border-b border-dashed border-slate-300 pb-3 text-center">
            <p class="text-lg font-black">{{ receiptShopName }}</p>
            <p v-if="receiptShopAddress" class="whitespace-pre-line text-xs font-bold text-slate-600">{{ receiptShopAddress }}</p>
            <p v-if="receiptShopPhone" class="text-xs font-bold text-slate-600">{{ receiptShopPhone }}</p>
            <p class="text-sm font-semibold text-slate-600">{{ app.t('pos.receiptPreview') }}</p>
          </header>

          <dl class="mt-3 grid gap-1.5 text-xs">
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.receiptNo') }}</dt><dd class="text-right font-bold">{{ app.t('pos.previewReceiptNo') }}</dd></div>
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.date') }}</dt><dd class="text-right font-bold">{{ receiptPreviewDate }}</dd></div>
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.cashier') }}</dt><dd class="text-right font-bold">{{ auth.user?.fullName ?? auth.user?.username ?? '-' }}</dd></div>
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.location') }}</dt><dd class="text-right font-bold">{{ selectedLocation?.name ?? '-' }}</dd></div>
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.paymentMethod') }}</dt><dd class="text-right font-bold">{{ paymentLabel(cart.paymentMethod) }}</dd></div>
          </dl>

          <section class="mt-3 border-y border-dashed border-slate-300 py-3">
            <div class="grid grid-cols-[minmax(0,1fr)_34px_64px_72px] gap-2 text-xs font-black text-slate-500">
              <span>{{ app.t('receipt.items') }}</span>
              <span class="text-right">{{ app.t('receipt.quantity') }}</span>
              <span class="text-right">{{ app.t('receipt.price') }}</span>
              <span class="text-right">{{ app.t('receipt.subtotal') }}</span>
            </div>
            <div v-for="item in cart.items" :key="item.productId" class="grid grid-cols-[minmax(0,1fr)_34px_64px_72px] gap-2 py-1.5 text-xs">
              <span class="min-w-0">
                <b class="block overflow-wrap-anywhere">{{ item.name }}</b>
                <small class="block text-slate-500">{{ item.sku }}</small>
              </span>
              <span class="text-right">{{ item.quantity }}</span>
              <span class="text-right">{{ money(item.price) }}</span>
              <span class="text-right font-bold">{{ money(item.price * item.quantity) }}</span>
            </div>
          </section>

          <dl class="mt-3 grid gap-1.5 text-sm">
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.subtotal') }}</dt><dd class="text-right font-bold">{{ money(cart.totalAmount) }}</dd></div>
            <div class="flex justify-between gap-3 border-t border-dashed border-slate-300 pt-2 text-base font-black"><dt>{{ app.t('receipt.total') }}</dt><dd class="text-right">{{ money(cart.totalAmount) }}</dd></div>
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.received') }}</dt><dd class="text-right font-bold">{{ money(cart.receivedAmount) }}</dd></div>
            <div class="flex justify-between gap-3"><dt class="text-slate-500">{{ app.t('receipt.change') }}</dt><dd class="text-right font-bold">{{ money(cart.changeAmount) }}</dd></div>
          </dl>

          <footer class="mt-3 whitespace-pre-line border-t border-dashed border-slate-300 pt-3 text-center text-xs font-bold text-slate-600">
            {{ receiptFooter }}
          </footer>
        </article>

        <div class="flex flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end">
          <AppButton class="sm:w-auto" variant="secondary" :disabled="cart.isSubmitting" @click="closeSaleConfirm">{{ app.t('pos.cancel') }}</AppButton>
          <AppButton class="sm:w-40" variant="secondary" icon="receipt-text" :disabled="!canSubmit || cart.isSubmitting" @click="printReceiptPreview">{{ app.t('pos.printReceiptPreview') }}</AppButton>
          <AppButton class="sm:min-w-52" :loading="cart.isSubmitting" :disabled="!canSubmit" @click="submitSale">{{ app.t('pos.confirmSale') }}</AppButton>
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
          <RouterLink v-if="canViewReceipt" :to="receiptPath" class="focus-ring inline-flex min-h-11 items-center justify-center rounded-xl bg-brand-600 px-4 py-2 text-sm font-bold text-white shadow-sm shadow-brand-600/20 dark:bg-teal-500 dark:text-slate-950">
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

.overflow-wrap-anywhere {
  overflow-wrap: anywhere;
}

.pos-category-shelf {
  scrollbar-width: thin;
}
</style>
