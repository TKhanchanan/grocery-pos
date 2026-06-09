<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiClient, postJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppDateRangeFilter from '../components/AppDateRangeFilter.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppModal from '../components/AppModal.vue'
import AppPageSizeSelect from '../components/AppPageSizeSelect.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTabs from '../components/AppTabs.vue'
import AppTextarea from '../components/AppTextarea.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
import ProductAvatar from '../components/ProductAvatar.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { Location, Product, ProductStock, StockMovement } from '../types/navigation'
import { formatAppDateTime } from '../utils/date'

type StockTab = 'restock' | 'movements'

interface StockMovementPage {
  items: StockMovement[]
  total: number
  page: number
  page_size: number
}

interface StockOperationOptions {
  products: Product[]
  locations: Location[]
  stocks: ProductStock[]
}

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const products = ref<Product[]>([])
const locations = ref<Location[]>([])
const movementProducts = ref<Product[]>([])
const movementLocations = ref<Location[]>([])
const stocks = ref<ProductStock[]>([])
const movements = ref<StockMovement[]>([])
const latestMovement = ref<StockMovement | null>(null)
const loadingData = ref(false)
const loadingMovements = ref(false)
const submitting = ref(false)
const adjustSubmitting = ref(false)
const error = ref('')
const adjustError = ref('')
const movementError = ref('')
const adjustOpen = ref(false)
const restockConfirmOpen = ref(false)
const adjustConfirmOpen = ref(false)
const activeTab = ref<StockTab>('restock')
const page = ref(1)
const pageSize = ref(20)
const totalMovements = ref(0)

const form = reactive({
  product_id: '',
  location_id: '',
  quantity: 1,
  total_cost: 0,
  unit_cost: 0,
  note: '',
})

const adjustment = reactive({
  quantity: -1,
  note: '',
})

const movementFilters = reactive({
  product_id: '',
  location_id: '',
  type: '',
  date_from: '',
  date_to: '',
})

const canRestock = computed(() => auth.hasPermission('stock.restock'))
const canAdjust = computed(() => auth.hasPermission('stock.adjust'))
const canViewMovements = computed(() => auth.hasPermission('stock.movements.view'))
const canUseRestockTab = computed(() => canRestock.value || canAdjust.value)
const tabs = computed(() => {
  const items: Array<{ key: StockTab; label: string }> = []
  if (canUseRestockTab.value) items.push({ key: 'restock', label: app.t('stockOps.tab.restock') })
  if (canViewMovements.value) items.push({ key: 'movements', label: app.t('stockOps.tab.history') })
  return items
})
const selectedProduct = computed(() => products.value.find((product) => product.id === Number(form.product_id)) ?? null)
const selectedLocation = computed(() => locations.value.find((location) => location.id === Number(form.location_id)) ?? null)
const currentStock = computed(() => stocks.value.find((stock) => stock.product_id === Number(form.product_id) && stock.location_id === Number(form.location_id))?.quantity ?? 0)
const unitCostPreview = computed(() => form.total_cost > 0 && form.quantity > 0 ? Number(form.total_cost) / Number(form.quantity) : Number(form.unit_cost || 0))
const afterRestockPreview = computed(() => currentStock.value + Number(form.quantity || 0))
const afterAdjustmentPreview = computed(() => currentStock.value + Number(adjustment.quantity || 0))
const totalPages = computed(() => Math.max(1, Math.ceil(totalMovements.value / pageSize.value)))
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function money(value: number) {
  const amount = value.toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return t('stockOps.baht', { amount })
}

function stockLine(quantity: number) {
  return t('stockOps.stockLine', { quantity, unit: selectedProduct.value?.unit ?? '' })
}

function signed(value: number) {
  return value > 0 ? `+${value}` : String(value)
}

function movementProduct(movement: StockMovement) {
  return products.value.find((product) => product.id === movement.product_id) ?? null
}

function movementImageURL(movement: StockMovement) {
  return movement.image_url ?? movementProduct(movement)?.image_url ?? null
}

function movementImageUpdatedAt(movement: StockMovement) {
  return movement.image_updated_at ?? movementProduct(movement)?.image_updated_at ?? null
}

function movementTone(type: string) {
  if (type.includes('SALE') || type.includes('OUT')) return 'danger'
  if (type.includes('RESTOCK') || type.includes('IN') || type.includes('RECEIVE')) return 'success'
  return 'info'
}

function movementLabel(type: string) {
  const keys: Record<string, TranslationKey> = {
    RESTOCK: 'stockOps.movement.RESTOCK',
    ADJUSTMENT: 'stockOps.movement.ADJUSTMENT',
    SALE: 'stockOps.movement.SALE',
    CANCEL_SALE: 'stockOps.movement.CANCEL_SALE',
    PO_RECEIVE: 'stockOps.movement.PO_RECEIVE',
    TRANSFER_IN: 'stockOps.movement.TRANSFER_IN',
    TRANSFER_OUT: 'stockOps.movement.TRANSFER_OUT',
    IMPORT: 'stockOps.movement.IMPORT',
    SEED: 'stockOps.movement.SEED',
  }
  const key = keys[type]
  return key ? app.t(key) : type
}

function friendlyError(err: unknown, fallback: TranslationKey) {
  const message = err instanceof Error ? err.message : app.t(fallback)
  if (message.toLowerCase().includes('permission')) return app.t('stockOps.noPermission')
  if (message.toLowerCase().includes('insufficient stock') || message.toLowerCase().includes('stock cannot become negative')) {
    return app.t('stockOps.insufficientStock')
  }
  return message
}

function calculateSuggestedTotalCost() {
  const quantity = Math.max(0, Number(form.quantity || 0))
  const unitCost = Math.max(0, Number(selectedProduct.value?.unit_cost || 0))
  form.total_cost = Number((quantity * unitCost).toFixed(2))
}

function applySelectedProductCost() {
  form.unit_cost = Math.max(0, Number(selectedProduct.value?.unit_cost || 0))
  calculateSuggestedTotalCost()
}

function syncActiveTabFromRoute() {
  const requested = route.query.tab === 'movements' ? 'movements' : 'restock'
  if (requested === 'movements' && canViewMovements.value) activeTab.value = 'movements'
  else if (requested === 'restock' && canUseRestockTab.value) activeTab.value = 'restock'
  else activeTab.value = tabs.value[0]?.key ?? 'restock'
}

function setActiveTab(tab: StockTab) {
  activeTab.value = tab
  router.replace({ path: '/stock-operations', query: { ...route.query, tab } })
  if (tab === 'movements') {
    Object.assign(movementFilters, {
      product_id: '',
      location_id: '',
      type: '',
      date_from: '',
      date_to: '',
    })
    page.value = 1
    loadMovements()
  }
}

async function loadData() {
  loadingData.value = true
  error.value = ''
  try {
    const options = await apiClient<StockOperationOptions>('/v1/stock-operations/options')
    movementProducts.value = options.products
    movementLocations.value = options.locations
    products.value = options.products.filter((product) => product.is_active)
    locations.value = options.locations.filter((location) => location.is_active)
    stocks.value = options.stocks
    if (!form.product_id && products.value[0]) form.product_id = String(products.value[0].id)
    if (!form.location_id && locations.value[0]) form.location_id = String(locations.value[0].id)
  } catch (err) {
    error.value = friendlyError(err, 'stockOps.loadFailed')
  } finally {
    loadingData.value = false
  }
}

async function loadMovements() {
  if (!canViewMovements.value) return
  loadingMovements.value = true
  movementError.value = ''
  try {
    const params = new URLSearchParams({ page: String(page.value), page_size: String(pageSize.value) })
    if (movementFilters.product_id) params.set('product_id', movementFilters.product_id)
    if (movementFilters.location_id) params.set('location_id', movementFilters.location_id)
    if (movementFilters.type) params.set('type', movementFilters.type)
    if (movementFilters.date_from) params.set('date_from', movementFilters.date_from)
    if (movementFilters.date_to) params.set('date_to', movementFilters.date_to)
    const result = await apiClient<StockMovementPage>(`/v1/stock-movements?${params.toString()}`)
    movements.value = result.items
    totalMovements.value = result.total
    page.value = result.page
    pageSize.value = result.page_size
  } catch (err) {
    movementError.value = friendlyError(err, 'stockOps.historyFailed')
  } finally {
    loadingMovements.value = false
  }
}

function validateRestock() {
  if (Number(form.quantity) <= 0) return app.t('stockOps.quantityRequired')
  if (Number(form.total_cost || 0) < 0 || Number(form.unit_cost || 0) < 0) return app.t('stockOps.costRequired')
  if (Number(form.total_cost || 0) === 0 && Number(form.unit_cost || 0) === 0) return app.t('stockOps.costRequired')
  return ''
}

function requestRestock() {
  const validation = validateRestock()
  if (validation) {
    error.value = validation
    return
  }
  error.value = ''
  restockConfirmOpen.value = true
}

function closeRestockConfirm() {
  if (submitting.value) return
  restockConfirmOpen.value = false
}

async function restock() {
  submitting.value = true
  error.value = ''
  latestMovement.value = null
  try {
    latestMovement.value = await postJSON<StockMovement>(`/v1/products/${form.product_id}/restock`, {
      location_id: Number(form.location_id),
      quantity: Number(form.quantity),
      total_cost: form.total_cost ? Number(form.total_cost) : null,
      unit_cost: Number(unitCostPreview.value),
      note: form.note,
    })
    restockConfirmOpen.value = false
    app.pushToast({ type: 'success', message: app.t('stockOps.restockSuccess'), description: selectedProduct.value?.name })
    await loadData()
    if (canViewMovements.value) {
      page.value = 1
      await loadMovements()
    }
  } catch (err) {
    error.value = friendlyError(err, 'stockOps.restockFailed')
    app.pushToast({ type: 'error', message: app.t('stockOps.restockFailed'), description: error.value })
  } finally {
    submitting.value = false
  }
}

function openAdjust() {
  adjustment.quantity = -1
  adjustment.note = ''
  adjustError.value = ''
  adjustOpen.value = true
}

function closeAdjust(force = false) {
  if (adjustSubmitting.value && !force) return
  adjustOpen.value = false
  adjustConfirmOpen.value = false
}

function requestAdjustStock() {
  if (Number(adjustment.quantity) === 0) {
    adjustError.value = app.t('stockOps.adjustQuantityRequired')
    app.pushToast({ type: 'error', message: app.t('stockOps.adjustFailed'), description: adjustError.value })
    return
  }
  if (Number(adjustment.quantity) < 0 && afterAdjustmentPreview.value < 0) {
    adjustError.value = app.t('stockOps.insufficientStock')
    app.pushToast({ type: 'error', message: app.t('stockOps.adjustFailed'), description: adjustError.value })
    return
  }
  if (!adjustment.note.trim()) {
    adjustError.value = app.t('stockOps.noteRequired')
    app.pushToast({ type: 'error', message: app.t('stockOps.adjustFailed'), description: adjustError.value })
    return
  }
  adjustError.value = ''
  adjustConfirmOpen.value = true
}

function closeAdjustConfirm() {
  if (adjustSubmitting.value) return
  adjustConfirmOpen.value = false
}

async function adjustStock() {
  adjustSubmitting.value = true
  adjustError.value = ''
  const productName = selectedProduct.value?.name
  try {
    latestMovement.value = await postJSON<StockMovement>(`/v1/products/${form.product_id}/adjust-stock`, {
      location_id: Number(form.location_id),
      quantity: Number(adjustment.quantity),
      note: adjustment.note.trim(),
    })
    adjustConfirmOpen.value = false
    closeAdjust(true)
    app.pushToast({ type: 'success', message: app.t('stockOps.adjustSuccess'), description: productName })
    await loadData()
    if (canViewMovements.value) {
      page.value = 1
      await loadMovements()
    }
  } catch (err) {
    adjustError.value = friendlyError(err, 'stockOps.adjustFailed')
    app.pushToast({ type: 'error', message: app.t('stockOps.adjustFailed'), description: adjustError.value })
  } finally {
    adjustSubmitting.value = false
  }
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  loadMovements()
}

function applyMovementFilters() {
  page.value = 1
  loadMovements()
}

function resetMovementFilters() {
  Object.assign(movementFilters, {
    product_id: '',
    location_id: '',
    type: '',
    date_from: '',
    date_to: '',
  })
  page.value = 1
  loadMovements()
}

function nextPage() {
  if (page.value >= totalPages.value) return
  page.value += 1
  loadMovements()
}

function previousPage() {
  if (page.value <= 1) return
  page.value -= 1
  loadMovements()
}

watch(() => route.query.tab, syncActiveTabFromRoute)
watch(() => form.product_id, applySelectedProductCost)
watch(() => form.quantity, calculateSuggestedTotalCost)

onMounted(async () => {
  syncActiveTabFromRoute()
  await loadData()
  if (canViewMovements.value) await loadMovements()
})
</script>

<template>
  <section>
    <PageHeader :title="app.t('stockOps.title')" :eyebrow="app.t('stockOps.eyebrow')" :description="app.t('stockOps.description')" icon="package-plus" />

    <div class="grid gap-4">
      <AppTabs v-if="tabs.length > 1" :tabs="tabs" :model-value="activeTab" @update:model-value="setActiveTab" />

      <div v-if="activeTab === 'restock' && canUseRestockTab" class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <AppCard class="dark:bg-slate-900/80">
          <AppLoadingState v-if="loadingData" :label="app.t('stockOps.loadingData')" />
          <form v-else class="grid gap-4" @submit.prevent="requestRestock">
            <div class="grid gap-4">
              <div class="grid gap-3 lg:grid-cols-2">
              <AppSelect v-model="form.product_id" :label="app.t('stockOps.product')">
                <option v-for="product in products" :key="product.id" :value="String(product.id)">{{ product.name }} · {{ product.sku }}</option>
              </AppSelect>
              <AppSelect v-model="form.location_id" :label="app.t('stockOps.location')">
                <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
              </AppSelect>
              </div>
              <div class="grid gap-3 lg:grid-cols-3">
                <AppInput v-model="form.quantity" :label="app.t('stockOps.quantity')" type="number" min="1" step="1" :placeholder="app.t('stockOps.quantityPlaceholder')" />
                <AppInput v-model="form.total_cost" :label="app.t('stockOps.totalCost')" type="number" min="0" step="0.01" :placeholder="app.t('stockOps.totalCostPlaceholder')" />
                <AppInput v-model="form.unit_cost" :label="app.t('stockOps.unitCostFallback')" type="number" min="0" step="0.01" :placeholder="app.t('stockOps.unitCostPlaceholder')" />
              </div>
              <AppTextarea v-model="form.note" :label="app.t('stockOps.note')" :placeholder="app.t('stockOps.notePlaceholder')" />
            </div>

            <div class="grid gap-3 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/50 md:hidden">
              <h2 class="font-black">{{ app.t('stockOps.stockPreview') }}</h2>
              <div class="grid grid-cols-3 gap-2 text-sm">
                <div><p class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.currentStock') }}</p><p class="font-black">{{ stockLine(currentStock) }}</p></div>
                <div><p class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.afterRestock') }}</p><p class="font-black">{{ stockLine(afterRestockPreview) }}</p></div>
                <div><p class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.unitCost') }}</p><p class="font-black">{{ money(unitCostPreview) }}</p></div>
              </div>
            </div>

            <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
            <div class="flex flex-wrap gap-2">
              <AppButton v-if="canRestock" type="submit" :loading="submitting" :disabled="submitting" icon="package-plus">{{ app.t('stockOps.submitRestock') }}</AppButton>
              <AppButton v-if="canAdjust" type="button" variant="secondary" icon="settings" @click="openAdjust">{{ app.t('stockOps.adjustStock') }}</AppButton>
            </div>
          </form>
        </AppCard>

        <div class="grid gap-4">
          <AppCard class="hidden dark:bg-slate-900/80 md:block">
            <h2 class="font-black">{{ app.t('stockOps.stockPreview') }}</h2>
            <div class="mt-4 flex min-w-0 items-center gap-3">
              <ProductAvatar :src="selectedProduct?.image_url" :updated-at="selectedProduct?.image_updated_at" :name="selectedProduct?.name" size="lg" shape="square" />
              <div class="min-w-0">
                <p class="truncate font-black">{{ selectedProduct?.name ?? '-' }}</p>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ selectedProduct?.sku ?? '-' }} · {{ selectedLocation?.name ?? '-' }}</p>
              </div>
            </div>
            <dl class="mt-4 grid gap-3">
              <div class="rounded-xl bg-slate-50 p-3 dark:bg-slate-950/60"><dt class="text-sm text-slate-500 dark:text-slate-400">{{ app.t('stockOps.currentStock') }}</dt><dd class="text-xl font-black">{{ stockLine(currentStock) }}</dd></div>
              <div class="rounded-xl bg-slate-50 p-3 dark:bg-slate-950/60"><dt class="text-sm text-slate-500 dark:text-slate-400">{{ app.t('stockOps.afterRestock') }}</dt><dd class="text-xl font-black text-brand-700 dark:text-emerald-200">{{ stockLine(afterRestockPreview) }}</dd></div>
              <div class="rounded-xl bg-slate-50 p-3 dark:bg-slate-950/60"><dt class="text-sm text-slate-500 dark:text-slate-400">{{ app.t('stockOps.unitCost') }}</dt><dd class="text-xl font-black">{{ money(unitCostPreview) }}</dd></div>
            </dl>
          </AppCard>

          <AppCard v-if="latestMovement" class="dark:bg-slate-900/80">
            <h2 class="font-black">{{ app.t('stockOps.latestMovement') }}</h2>
            <dl class="mt-3 grid grid-cols-3 gap-3 text-sm">
              <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.type') }}</dt><dd class="font-black">{{ movementLabel(latestMovement.reference_type) }}</dd></div>
              <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.before') }}</dt><dd class="font-black">{{ latestMovement.before_stock }}</dd></div>
              <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.after') }}</dt><dd class="font-black">{{ latestMovement.after_stock }}</dd></div>
            </dl>
          </AppCard>
        </div>
      </div>

      <div v-if="activeTab === 'movements' && canViewMovements" class="grid min-w-0 gap-4">
        <AppCard :padded="false" class="min-w-0 p-3 dark:bg-slate-900/80 sm:p-4">
          <div class="grid min-w-0 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-[minmax(170px,1.1fr)_minmax(140px,0.75fr)_minmax(160px,0.85fr)_minmax(300px,1.45fr)_auto] xl:items-end">
            <AppSelect v-model="movementFilters.product_id" :label="app.t('stockOps.filterProduct')">
              <option value="">{{ app.t('stockOps.allProducts') }}</option>
              <option v-for="product in movementProducts" :key="product.id" :value="String(product.id)">{{ product.name }} · {{ product.sku }}</option>
            </AppSelect>
            <AppSelect v-model="movementFilters.location_id" :label="app.t('stockOps.filterLocation')">
              <option value="">{{ app.t('stockOps.allLocations') }}</option>
              <option v-for="location in movementLocations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
            </AppSelect>
            <AppSelect v-model="movementFilters.type" :label="app.t('stockOps.filterMovementType')">
              <option value="">{{ app.t('stockOps.allMovementTypes') }}</option>
              <option v-for="type in ['RESTOCK', 'ADJUSTMENT', 'SALE', 'CANCEL_SALE', 'PO_RECEIVE', 'TRANSFER_IN', 'TRANSFER_OUT', 'IMPORT', 'SEED']" :key="type" :value="type">{{ movementLabel(type) }}</option>
            </AppSelect>
            <AppDateRangeFilter
              class="sm:col-span-2 lg:col-span-2 xl:col-span-1"
              v-model:date-from="movementFilters.date_from"
              v-model:date-to="movementFilters.date_to"
              :date-from-label="app.t('stockOps.dateFrom')"
              :date-to-label="app.t('stockOps.dateTo')"
              :date-placeholder="app.t('stockOps.selectDate')"
              :today-label="app.t('stockOps.today')"
              :locale="app.language === 'th' ? 'th-TH-u-ca-buddhist' : 'en-US'"
              :show-shortcuts="false"
            />
            <div class="flex items-end gap-2 sm:col-span-2 lg:col-span-1 xl:col-span-1 xl:self-end">
              <AppButton class="!h-11 !min-h-11 flex-1 !px-3 !py-0 whitespace-nowrap xl:flex-none" icon="search" :disabled="loadingMovements" @click="applyMovementFilters">{{ app.t('stockOps.applyFilters') }}</AppButton>
              <AppButton class="!h-11 !min-h-11 flex-1 !px-3 !py-0 whitespace-nowrap xl:flex-none" variant="secondary" icon="x" :disabled="loadingMovements" @click="resetMovementFilters">{{ app.t('stockOps.resetFilters') }}</AppButton>
            </div>
          </div>
        </AppCard>

        <AppCard class="min-w-0 overflow-hidden dark:bg-slate-900/80">
        <div v-if="movementError" class="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ movementError }}</div>
        <AppLoadingState v-if="loadingMovements" :label="app.t('stockOps.loadingHistory')" />
        <AppEmptyState v-else-if="movements.length === 0" :title="app.t('stockOps.noHistory')" :description="app.t('stockOps.noHistoryDescription')" />
        <div v-else class="min-w-0 max-w-full">
          <div class="hidden w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain pb-2 [scrollbar-gutter:stable] md:block">
            <table class="w-full min-w-[1120px] divide-y divide-slate-200 text-sm dark:divide-slate-800">
              <thead class="bg-slate-50 dark:bg-slate-950/70">
                <tr>
                  <th class="whitespace-nowrap px-3 py-3 text-left">{{ app.t('stockOps.time') }}</th>
                  <th class="px-3 py-3 text-left">{{ app.t('stockOps.product') }}</th>
                  <th class="whitespace-nowrap px-3 py-3 text-left">{{ app.t('stockOps.location') }}</th>
                  <th class="px-3 py-3 text-left">{{ app.t('stockOps.type') }}</th>
                  <th class="px-3 py-3 text-right">{{ app.t('stockOps.change') }}</th>
                  <th class="px-3 py-3 text-right">{{ app.t('stockOps.before') }}</th>
                  <th class="px-3 py-3 text-right">{{ app.t('stockOps.after') }}</th>
                  <th class="px-3 py-3 text-left">{{ app.t('stockOps.note') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                <tr v-for="movement in movements" :key="movement.id" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                  <td class="whitespace-nowrap px-3 py-3">{{ formatAppDateTime(movement.created_at, app.language) }}</td>
                  <td class="px-3 py-3">
                    <div class="flex min-w-0 items-center gap-3">
                      <ProductAvatar :src="movementImageURL(movement)" :updated-at="movementImageUpdatedAt(movement)" :name="movement.product_name" size="sm" shape="square" />
                      <div class="min-w-0">
                        <b class="block truncate">{{ movement.product_name }}</b>
                        <span class="text-xs text-slate-500 dark:text-slate-400">{{ movement.sku }}</span>
                      </div>
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-3 py-3">{{ movement.location_name }}</td>
                  <td class="px-3 py-3"><AppBadge :tone="movementTone(movement.reference_type)">{{ movementLabel(movement.reference_type) }}</AppBadge></td>
                  <td class="px-3 py-3 text-right font-black" :class="movement.quantity_change < 0 ? 'text-red-600 dark:text-red-300' : 'text-brand-700 dark:text-emerald-200'">{{ signed(movement.quantity_change) }}</td>
                  <td class="px-3 py-3 text-right">{{ movement.before_stock }}</td>
                  <td class="px-3 py-3 text-right">{{ movement.after_stock }}</td>
                  <td class="max-w-[220px] truncate px-3 py-3" :title="movement.note">{{ movement.note || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="grid gap-3 md:hidden">
            <article v-for="movement in movements" :key="movement.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 dark:border-slate-700 dark:bg-slate-950/60">
              <div class="flex items-start justify-between gap-3">
                <div class="flex min-w-0 items-center gap-3">
                  <ProductAvatar :src="movementImageURL(movement)" :updated-at="movementImageUpdatedAt(movement)" :name="movement.product_name" size="sm" shape="square" />
                  <div class="min-w-0">
                    <h3 class="truncate font-black">{{ movement.product_name }}</h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400">{{ movement.location_name }} · {{ movement.sku }}</p>
                  </div>
                </div>
                <span class="font-black" :class="movement.quantity_change < 0 ? 'text-red-600 dark:text-red-300' : 'text-brand-700 dark:text-emerald-200'">{{ signed(movement.quantity_change) }}</span>
              </div>
              <div class="mt-3 flex flex-wrap gap-2"><AppBadge :tone="movementTone(movement.reference_type)">{{ movementLabel(movement.reference_type) }}</AppBadge><span class="text-xs text-slate-500 dark:text-slate-400">{{ formatAppDateTime(movement.created_at, app.language) }}</span></div>
              <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.before') }}</dt><dd class="font-semibold">{{ movement.before_stock }}</dd></div>
                <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.after') }}</dt><dd class="font-semibold">{{ movement.after_stock }}</dd></div>
                <div class="col-span-2"><dt class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.note') }}</dt><dd class="font-semibold">{{ movement.note || '-' }}</dd></div>
              </dl>
            </article>
          </div>

          <div class="mt-4 flex flex-col gap-3 border-t border-slate-200 pt-4 text-sm dark:border-slate-800 md:flex-row md:items-center md:justify-between">
            <div class="flex flex-wrap items-center gap-2">
              <span>{{ app.t('stockOps.show') }}</span>
              <AppPageSizeSelect :model-value="pageSize" @update:model-value="changePageSize" />
              <span>{{ app.t('stockOps.perPage') }}</span>
              <span class="text-slate-500 dark:text-slate-400">{{ app.t('stockOps.total') }} {{ totalMovements }}</span>
            </div>
            <div class="flex items-center justify-between gap-2 md:justify-end">
              <AppButton variant="secondary" :disabled="page <= 1 || loadingMovements" @click="previousPage">{{ app.t('stockOps.previous') }}</AppButton>
              <span class="font-bold">{{ app.t('stockOps.page') }} {{ page }} / {{ totalPages }}</span>
              <AppButton variant="secondary" :disabled="page >= totalPages || loadingMovements" @click="nextPage">{{ app.t('stockOps.next') }}</AppButton>
            </div>
          </div>
        </div>
        </AppCard>
      </div>
    </div>

    <AppModal :open="adjustOpen" :title="app.t('stockOps.adjustTitle')" :description="app.t('stockOps.adjustDescription')" @close="closeAdjust">
      <form class="grid gap-3" @submit.prevent="requestAdjustStock">
        <div class="rounded-2xl bg-slate-50 p-4 text-sm dark:bg-slate-950/60">
          <p><b>{{ app.t('stockOps.currentStock') }}:</b> {{ stockLine(currentStock) }}</p>
          <p><b>{{ app.t('stockOps.afterAdjustment') }}:</b> {{ stockLine(afterAdjustmentPreview) }}</p>
        </div>
        <AppInput v-model="adjustment.quantity" :label="app.t('stockOps.quantity')" type="number" :placeholder="app.t('stockOps.adjustQuantityPlaceholder')" />
        <AppTextarea v-model="adjustment.note" :label="app.t('stockOps.note')" :placeholder="app.t('stockOps.adjustNotePlaceholder')" />
        <div v-if="adjustError" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ adjustError }}</div>
        <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <AppButton type="button" variant="secondary" :disabled="adjustSubmitting" @click="closeAdjust">{{ app.t('stockOps.cancel') }}</AppButton>
          <AppButton type="submit" :loading="adjustSubmitting" :disabled="adjustSubmitting">{{ app.t('stockOps.submitAdjustment') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="restockConfirmOpen"
      :title="app.t('stockOps.confirmRestockTitle')"
      :message="t('stockOps.confirmRestockMessage', { product: selectedProduct?.name ?? app.t('stockOps.product'), quantity: Number(form.quantity || 0).toLocaleString(locale) })"
      :confirm-label="app.t('stockOps.submitRestock')"
      :cancel-label="app.t('stockOps.cancel')"
      :loading="submitting"
      @close="closeRestockConfirm"
      @confirm="restock"
    />

    <ConfirmDialog
      :open="adjustConfirmOpen"
      :title="app.t('stockOps.confirmAdjustTitle')"
      :message="t('stockOps.confirmAdjustMessage', { product: selectedProduct?.name ?? app.t('stockOps.product'), quantity: signed(Number(adjustment.quantity || 0)) })"
      :confirm-label="app.t('stockOps.submitAdjustment')"
      :cancel-label="app.t('stockOps.cancel')"
      :loading="adjustSubmitting"
      @close="closeAdjustConfirm"
      @confirm="adjustStock"
    />
  </section>
</template>
