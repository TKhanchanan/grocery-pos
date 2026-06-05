<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { API_BASE_URL, apiClient, deleteJSON, patchJSON, postJSON } from '../api/client'
import { downloadFile } from '../api/download'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppCheckbox from '../components/AppCheckbox.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppModal from '../components/AppModal.vue'
import AppSelect from '../components/AppSelect.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
import ProductAvatar from '../components/ProductAvatar.vue'
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { Category, ImportJob, ImportJobRow, Product, ProductStock, StockStatus } from '../types/navigation'

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

interface ProductImageResponse {
  product: Product
  content_type: string
}

const maxImageBytes = 2 * 1024 * 1024
const allowedImageTypes = ['image/jpeg', 'image/png', 'image/webp']

const app = useAppStore()
const auth = useAuthStore()
const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const selectedStocks = ref<ProductStock[]>([])
const selectedProduct = ref<Product | null>(null)
const loading = ref(false)
const saving = ref(false)
const imageSaving = ref(false)
const error = ref('')
const modalOpen = ref(false)
const saveConfirmOpen = ref(false)
const importModalOpen = ref(false)
const stockTooltipProductID = ref<number | null>(null)
const stockTooltipLoading = ref(false)
const stockTooltipRef = ref<HTMLElement | null>(null)
const stockTooltipPosition = reactive({ top: 0, left: 0 })
const stockTooltipAnchor = ref<HTMLElement | null>(null)
const barcodeScanOpen = ref(false)
const barcodeScanValue = ref('')
const barcodeScanInput = ref<{ focus: () => void } | null>(null)
const cameraOpen = ref(false)
const cameraMessageKey = ref<TranslationKey>('products.scannerStarting')
const videoRef = ref<HTMLVideoElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const imageFile = ref<File | null>(null)
const imagePreviewUrl = ref('')
const importPreviewJob = ref<ImportJob | null>(null)
const selectedImportFile = ref<File | null>(null)
const importUploading = ref(false)
const importConfirming = ref(false)
const importError = ref('')
const fieldErrors = reactive({ sku: '', name: '' })
const filters = reactive({ q: '', category_id: '', status: '', stock_status: '' })
const form = reactive<ProductForm>({
  id: 0,
  sku: '',
  name: '',
  barcode: '',
  category_id: '',
  selling_price: 1,
  unit_cost: 0,
  unit: '',
  threshold: 0,
  reorder_point: 0,
  is_active: true,
})

const canCreate = computed(() => auth.hasPermission('products.create'))
const canUpdate = computed(() => auth.hasPermission('products.update'))
const canDeactivate = computed(() => auth.hasPermission('products.deactivate'))
const canImportProducts = computed(() => auth.hasAnyPermission(['imports.view', 'imports.products.preview', 'imports.products.confirm']))
const canDownloadImportTemplate = computed(() => auth.hasPermission('imports.template.download'))
const canPreviewImport = computed(() => auth.hasPermission('imports.products.preview'))
const canConfirmImport = computed(() => auth.hasPermission('imports.products.confirm'))
const activeCount = computed(() => products.value.filter((product) => product.is_active).length)
const lowSignalCount = computed(() => products.value.filter((product) => product.stock_status !== 'in_stock').length)
const totalStock = computed(() => products.value.reduce((sum, product) => sum + product.total_stock, 0))
const validImportRows = computed(() => importPreviewJob.value?.rows?.filter((row) => row.status === 'PENDING').length ?? 0)
const invalidImportRows = computed(() => importPreviewJob.value?.rows?.filter((row) => row.status === 'FAILED').length ?? 0)
const modalTitle = computed(() => form.id ? app.t('products.edit') : app.t('products.add'))
const saveConfirmTitle = computed(() => form.id ? app.t('products.confirmUpdateTitle') : app.t('products.confirmCreateTitle'))
const saveConfirmMessage = computed(() => t('products.confirmSaveMessage', { name: form.name.trim() || app.t('products.name') }))
const currentImage = computed(() => imagePreviewUrl.value || selectedProduct.value?.image_url || '')
const currentImageUpdatedAt = computed(() => imagePreviewUrl.value ? '' : selectedProduct.value?.image_updated_at ?? '')
const cameraMessage = computed(() => app.t(cameraMessageKey.value))
const stockTooltipStyle = computed(() => ({
  top: `${stockTooltipPosition.top}px`,
  left: `${stockTooltipPosition.left}px`,
}))
let cameraStream: MediaStream | null = null
let scanFrame = 0

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function formatFileSize(size?: number) {
  if (!size) return ''
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function defaultUnit() {
  return app.language === 'th' ? 'ชิ้น' : 'piece'
}

function money(value: number) {
  return value.toLocaleString(app.language === 'th' ? 'th-TH' : 'en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function stockLabel(status: StockStatus) {
  const labels: Record<StockStatus, TranslationKey> = {
    in_stock: 'products.inStock',
    low_stock: 'products.lowStock',
    out_of_stock: 'products.outOfStock',
    reorder_point: 'products.reorderStock',
  }
  return app.t(labels[status])
}

function stockClass(status: StockStatus) {
  return {
    in_stock: 'bg-brand-100 text-brand-700 dark:bg-emerald-500/15 dark:text-emerald-100',
    low_stock: 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-100',
    out_of_stock: 'bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-100',
    reorder_point: 'bg-blue-100 text-blue-700 dark:bg-blue-500/15 dark:text-blue-100',
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
    const [productRows, categoryRows] = await Promise.all([
      apiClient<Product[]>(`/v1/products${qs ? `?${qs}` : ''}`),
      apiClient<Category[]>('/v1/categories'),
    ])
    products.value = productRows
    categories.value = categoryRows
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('products.notFound')
  } finally {
    loading.value = false
  }
}

function revokePreview() {
  if (imagePreviewUrl.value) URL.revokeObjectURL(imagePreviewUrl.value)
  imagePreviewUrl.value = ''
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
    unit: defaultUnit(),
    threshold: 0,
    reorder_point: 0,
    is_active: true,
  })
  selectedProduct.value = null
  imageFile.value = null
  revokePreview()
  fieldErrors.sku = ''
  fieldErrors.name = ''
  if (fileInput.value) fileInput.value.value = ''
}

function openCreate() {
  resetForm()
  modalOpen.value = true
}

function openEdit(product: Product) {
  resetForm()
  selectedProduct.value = product
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
  modalOpen.value = true
}

function closeModal() {
  if (saving.value || imageSaving.value) return
  closeCameraScanner()
  barcodeScanOpen.value = false
  saveConfirmOpen.value = false
  modalOpen.value = false
  resetForm()
}

function openSaveConfirm() {
  if (saving.value || imageSaving.value) return
  if (!validateForm()) return
  saveConfirmOpen.value = true
}

function openProductImport() {
  selectedImportFile.value = null
  importPreviewJob.value = null
  importError.value = ''
  importModalOpen.value = true
}

function closeProductImport() {
  if (importUploading.value || importConfirming.value) return
  resetProductImport()
}

function resetProductImport() {
  importModalOpen.value = false
  selectedImportFile.value = null
  importPreviewJob.value = null
  importError.value = ''
}

function chooseImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  selectedImportFile.value = input.files?.[0] ?? null
  importPreviewJob.value = null
  importError.value = ''
}

function importRowClass(row: ImportJobRow) {
  if (row.status === 'FAILED') return 'bg-red-50 dark:bg-red-950/30'
  if (row.status === 'IMPORTED') return 'bg-brand-50 dark:bg-emerald-500/10'
  return ''
}

async function downloadImportTemplate() {
  try {
    await downloadFile('/v1/imports/products/template', 'product-import-template.csv')
  } catch (err) {
    importError.value = friendlyError(err, 'products.importFailed')
    app.pushToast({ type: 'error', message: app.t('products.importFailed'), description: importError.value })
  }
}

async function previewProductImport() {
  if (!selectedImportFile.value) return
  importUploading.value = true
  importError.value = ''
  const token = localStorage.getItem('auth_token')
  const body = new FormData()
  body.append('file', selectedImportFile.value)
  try {
    const response = await fetch(`${API_BASE_URL}/v1/imports/products/preview`, {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
      body,
    })
    const envelope = await response.json()
    if (!response.ok || !envelope.success) throw new Error(envelope.error?.message ?? app.t('products.importPreviewFailed'))
    importPreviewJob.value = envelope.data as ImportJob
  } catch (err) {
    importError.value = friendlyError(err, 'products.importPreviewFailed')
    app.pushToast({ type: 'error', message: app.t('products.importPreviewFailed'), description: importError.value })
  } finally {
    importUploading.value = false
  }
}

async function confirmProductImport() {
  if (!importPreviewJob.value) return
  importConfirming.value = true
  importError.value = ''
  const importedCount = validImportRows.value
  try {
    importPreviewJob.value = await postJSON<ImportJob>('/v1/imports/products/confirm', { job_id: importPreviewJob.value.id })
    await load()
    resetProductImport()
    app.pushToast({ type: 'success', message: app.t('products.importSuccess'), description: t('products.importSuccessDescription', { count: importedCount }) })
  } catch (err) {
    importError.value = friendlyError(err, 'products.importConfirmFailed')
    app.pushToast({ type: 'error', message: app.t('products.importConfirmFailed'), description: importError.value })
  } finally {
    importConfirming.value = false
  }
}

function payload() {
  return {
    sku: form.sku.trim(),
    name: form.name.trim(),
    barcode: form.barcode.trim() ? form.barcode.trim() : null,
    category_id: form.category_id ? Number(form.category_id) : null,
    selling_price: Number(form.selling_price),
    unit_cost: Number(form.unit_cost),
    unit: form.unit.trim() || defaultUnit(),
    threshold: Number(form.threshold || 0),
    reorder_point: Number(form.reorder_point || 0),
    is_active: form.is_active,
  }
}

function validateForm() {
  fieldErrors.sku = form.sku.trim() ? '' : app.t('products.sku')
  fieldErrors.name = form.name.trim() ? '' : app.t('products.name')
  return !fieldErrors.sku && !fieldErrors.name
}

function friendlyError(err: unknown, fallback: TranslationKey) {
  const message = err instanceof Error ? err.message : app.t(fallback)
  if (message.toLowerCase().includes('permission')) return app.t('products.noPermission')
  return message
}

function updateProductRow(product: Product) {
  const index = products.value.findIndex((item) => item.id === product.id)
  if (index >= 0) products.value[index] = product
  selectedProduct.value = product
}

function chooseImage() {
  fileInput.value?.click()
}

async function openBarcodeScan() {
  barcodeScanValue.value = form.barcode
  barcodeScanOpen.value = true
  await nextTick()
  barcodeScanInput.value?.focus()
}

function closeBarcodeScan() {
  barcodeScanOpen.value = false
  barcodeScanValue.value = ''
}

function applyBarcodeScan() {
  const value = barcodeScanValue.value.trim()
  if (!value) return
  form.barcode = value
  closeCameraScanner()
  closeBarcodeScan()
}

async function openCameraScanner() {
  barcodeScanValue.value = form.barcode
  cameraOpen.value = true
  cameraMessageKey.value = 'products.scannerStarting'
  await nextTick()
  const video = videoRef.value
  const detectorCtor = (window as unknown as { BarcodeDetector?: new (options?: object) => { detect(video: HTMLVideoElement): Promise<Array<{ rawValue: string }>> } }).BarcodeDetector
  if (!video || !navigator.mediaDevices?.getUserMedia || !detectorCtor) {
    cameraMessageKey.value = 'products.scannerUnavailable'
    return
  }
  try {
    cameraStream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
    video.srcObject = cameraStream
    await video.play()
    const detector = new detectorCtor({ formats: ['ean_13', 'code_128', 'qr_code'] })
    cameraMessageKey.value = 'products.scannerReady'
    const scan = async () => {
      if (!cameraOpen.value || !videoRef.value) return
      const codes = await detector.detect(videoRef.value).catch(() => [])
      if (codes[0]?.rawValue) {
        form.barcode = codes[0].rawValue
        closeCameraScanner()
        return
      }
      scanFrame = window.requestAnimationFrame(scan)
    }
    scanFrame = window.requestAnimationFrame(scan)
  } catch {
    cameraMessageKey.value = 'products.scannerBlocked'
  }
}

function closeCameraScanner() {
  cameraOpen.value = false
  window.cancelAnimationFrame(scanFrame)
  if (cameraStream) {
    cameraStream.getTracks().forEach((track) => track.stop())
    cameraStream = null
  }
}

function onImageSelected(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (!allowedImageTypes.includes(file.type)) {
    app.pushToast({ type: 'error', message: app.t('products.invalidImage') })
    if (fileInput.value) fileInput.value.value = ''
    return
  }
  if (file.size > maxImageBytes) {
    app.pushToast({ type: 'error', message: app.t('products.imageTooLarge') })
    if (fileInput.value) fileInput.value.value = ''
    return
  }
  imageFile.value = file
  revokePreview()
  imagePreviewUrl.value = URL.createObjectURL(file)
}

function clearSelectedImage() {
  imageFile.value = null
  revokePreview()
  if (fileInput.value) fileInput.value.value = ''
}

async function uploadSelectedImage(productID: number) {
  if (!imageFile.value) return
  imageSaving.value = true
  try {
    const body = new FormData()
    body.append('image', imageFile.value)
    const result = await apiClient<ProductImageResponse>(`/v1/products/${productID}/image`, { method: 'POST', body })
    updateProductRow(result.product)
    clearSelectedImage()
    app.pushToast({ type: 'success', message: app.t('products.imageUploaded') })
  } catch (err) {
    throw new Error(friendlyError(err, 'products.uploadFailed'))
  } finally {
    imageSaving.value = false
  }
}

async function removeProductImage() {
  if (!form.id) {
    clearSelectedImage()
    return
  }
  imageSaving.value = true
  try {
    const product = await deleteJSON<Product>(`/v1/products/${form.id}/image`)
    updateProductRow(product)
    clearSelectedImage()
    await load()
    app.pushToast({ type: 'success', message: app.t('products.imageRemoved') })
  } catch (err) {
    error.value = friendlyError(err, 'products.removeFailed')
    app.pushToast({ type: 'error', message: app.t('products.removeFailed'), description: error.value })
  } finally {
    imageSaving.value = false
  }
}

async function save() {
  if (!validateForm()) return
  saving.value = true
  error.value = ''
  const toastMessage = form.id ? app.t('products.updated') : app.t('products.created')
  try {
    const product = form.id
      ? await patchJSON<Product>(`/v1/products/${form.id}`, payload())
      : await postJSON<Product>('/v1/products', payload())
    updateProductRow(product)
    if (imageFile.value) await uploadSelectedImage(product.id)
    await load()
    saveConfirmOpen.value = false
    modalOpen.value = false
    resetForm()
    app.pushToast({ type: 'success', message: toastMessage })
  } catch (err) {
    error.value = friendlyError(err, 'products.saveFailed')
    app.pushToast({ type: 'error', message: app.t('products.saveFailed'), description: error.value })
  } finally {
    saving.value = false
  }
}

async function setActive(product: Product, active: boolean) {
  try {
    await patchJSON<Product>(`/v1/products/${product.id}/status`, { is_active: active })
    await load()
  } catch (err) {
    error.value = friendlyError(err, 'products.saveFailed')
    app.pushToast({ type: 'error', message: app.t('products.saveFailed'), description: error.value })
  }
}

function closeStockTooltip() {
  stockTooltipProductID.value = null
  stockTooltipAnchor.value = null
}

function positionStockTooltip(target: HTMLElement, tooltip?: HTMLElement | null) {
  const rect = target.getBoundingClientRect()
  const width = 288
  const height = tooltip?.offsetHeight || 190
  const gap = 8
  stockTooltipPosition.left = Math.max(16, Math.min(window.innerWidth - width - 16, rect.right - width))
  stockTooltipPosition.top = Math.max(16, Math.min(window.innerHeight - height - 16, rect.top))
  if (rect.bottom + height + gap <= window.innerHeight) stockTooltipPosition.top = rect.bottom + gap
}

function handleDocumentPointerDown(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  if (!target || !stockTooltipProductID.value) return
  if (stockTooltipRef.value?.contains(target) || target.closest('[data-stock-tooltip-trigger]')) return
  closeStockTooltip()
}

function handleTooltipViewportChange() {
  if (!stockTooltipProductID.value || !stockTooltipAnchor.value) return
  positionStockTooltip(stockTooltipAnchor.value, stockTooltipRef.value)
}

async function showStocks(event: MouseEvent, product: Product) {
  const target = event.currentTarget as HTMLElement
  if (stockTooltipProductID.value === product.id) {
    closeStockTooltip()
    return
  }
  selectedProduct.value = product
  stockTooltipAnchor.value = target
  positionStockTooltip(target)
  stockTooltipProductID.value = product.id
  stockTooltipLoading.value = true
  await nextTick()
  positionStockTooltip(target, stockTooltipRef.value)
  try {
    selectedStocks.value = await apiClient<ProductStock[]>(`/v1/products/${product.id}/stocks`)
    await nextTick()
    if (stockTooltipAnchor.value) positionStockTooltip(stockTooltipAnchor.value, stockTooltipRef.value)
  } catch (err) {
    error.value = friendlyError(err, 'products.notFound')
    stockTooltipProductID.value = null
    app.pushToast({ type: 'error', message: app.t('products.notFound'), description: error.value })
  } finally {
    stockTooltipLoading.value = false
  }
}

onMounted(() => {
  document.addEventListener('mousedown', handleDocumentPointerDown)
  window.addEventListener('resize', handleTooltipViewportChange)
  window.addEventListener('scroll', handleTooltipViewportChange, true)
  resetForm()
  load()
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleDocumentPointerDown)
  window.removeEventListener('resize', handleTooltipViewportChange)
  window.removeEventListener('scroll', handleTooltipViewportChange, true)
  closeCameraScanner()
  revokePreview()
})
</script>

<template>
  <section>
    <PageHeader :title="app.t('products.title')" :eyebrow="app.t('products.eyebrow')" :description="app.t('products.description')" icon="package">
      <div class="flex flex-wrap gap-2">
        <AppButton v-if="canImportProducts" variant="secondary" icon="upload" @click="openProductImport">{{ app.t('products.importData') }}</AppButton>
        <AppButton v-if="canCreate" icon="plus" @click="openCreate">{{ app.t('products.add') }}</AppButton>
      </div>
    </PageHeader>

    <div class="mb-4 grid gap-3 sm:grid-cols-3">
      <StatCard :label="app.t('products.title')" :value="products.length" :helper="t('products.activeHelper', { count: activeCount })" icon="package" />
      <StatCard :label="app.t('products.totalStock')" :value="totalStock" :helper="app.t('products.totalStockHelper')" icon="map-pin" tone="success" />
      <StatCard :label="app.t('products.needsAttention')" :value="lowSignalCount" :helper="app.t('products.attentionHelper')" icon="bell" tone="warning" />
    </div>

    <div class="grid gap-4">
      <AppCard class="dark:bg-slate-900/80">
        <div class="grid gap-3 lg:grid-cols-5">
          <AppInput v-model="filters.q" :label="app.t('products.search')" :placeholder="app.t('products.searchPlaceholder')" />
          <AppSelect v-model="filters.category_id" :label="app.t('products.category')">
            <option value="">{{ app.t('products.all') }}</option>
            <option v-for="category in categories" :key="category.id" :value="String(category.id)">{{ category.name }}</option>
          </AppSelect>
          <AppSelect v-model="filters.status" :label="app.t('products.status')">
            <option value="">{{ app.t('products.all') }}</option>
            <option value="active">{{ app.t('products.active') }}</option>
            <option value="inactive">{{ app.t('products.inactive') }}</option>
          </AppSelect>
          <AppSelect v-model="filters.stock_status" :label="app.t('products.stock')">
            <option value="">{{ app.t('products.all') }}</option>
            <option value="in_stock">{{ app.t('products.inStock') }}</option>
            <option value="low_stock">{{ app.t('products.lowStock') }}</option>
            <option value="out_of_stock">{{ app.t('products.outOfStock') }}</option>
            <option value="reorder_point">{{ app.t('products.reorderStock') }}</option>
          </AppSelect>
          <div class="flex items-end"><AppButton class="w-full" icon="search" @click="load">{{ app.t('products.apply') }}</AppButton></div>
        </div>
      </AppCard>

      <AppCard class="dark:bg-slate-900/80">
        <AppLoadingState v-if="loading" :label="app.t('products.loading')" />
        <AppEmptyState v-else-if="products.length === 0" :title="app.t('products.none')" :description="app.t('products.emptyDescription')">
          <template v-if="canCreate">
            <AppButton icon="plus" @click="openCreate">{{ app.t('products.addFirst') }}</AppButton>
          </template>
        </AppEmptyState>

        <div v-else>
          <div class="hidden overflow-x-auto md:block">
            <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
              <thead class="bg-slate-50 dark:bg-slate-950/70">
                <tr>
                  <th class="px-3 py-2 text-left">{{ app.t('products.name') }}</th>
                  <th class="px-3 py-2 text-left">{{ app.t('products.sku') }} / {{ app.t('products.barcode') }}</th>
                  <th class="px-3 py-2 text-left">{{ app.t('products.category') }}</th>
                  <th class="px-3 py-2 text-right">{{ app.t('products.sellingPrice') }}</th>
                  <th class="px-3 py-2 text-right">{{ app.t('products.stock') }}</th>
                  <th class="px-3 py-2 text-right">{{ app.t('products.status') }}</th>
                  <th class="px-3 py-2 text-right">{{ app.t('products.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                <tr v-for="product in products" :key="product.id" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                  <td class="px-3 py-2">
                    <div class="flex min-w-0 items-center gap-3">
                      <ProductAvatar :src="product.image_url" :updated-at="product.image_updated_at" :name="product.name" size="md" :muted="!product.is_active" />
                      <div class="min-w-0">
                        <p class="truncate font-semibold">{{ product.name }}</p>
                        <p class="text-xs text-slate-500 dark:text-slate-400">{{ product.unit }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="px-3 py-2">{{ product.sku }}<br /><span class="text-xs text-slate-500 dark:text-slate-400">{{ product.barcode || app.t('products.noBarcode') }}</span></td>
                  <td class="px-3 py-2">{{ product.category_name || app.t('products.noCategory') }}</td>
                  <td class="px-3 py-2 text-right">{{ money(product.selling_price) }}</td>
                  <td class="px-3 py-2 text-right">{{ product.total_stock }}</td>
                  <td class="px-3 py-2 text-right">
                    <span class="rounded-full px-2 py-1 text-xs font-bold" :class="stockClass(product.stock_status)">{{ stockLabel(product.stock_status) }}</span>
                    <span class="ml-2 text-xs text-slate-500 dark:text-slate-400">{{ product.is_active ? app.t('products.active') : app.t('products.inactive') }}</span>
                  </td>
                  <td class="px-3 py-2 text-right">
                    <div class="flex justify-end gap-2 whitespace-nowrap">
                      <AppButton data-stock-tooltip-trigger
                        class="!box-border !h-10 !min-h-10 !w-10 !min-w-10 !shrink-0 !px-0 !py-0" variant="secondary"
                        icon="map-pin" :title="app.t('products.stocks')" :aria-label="app.t('products.stocks')"
                        @click="showStocks($event, product)" />

                      <AppButton v-if="canUpdate"
                        class="!box-border !h-10 !min-h-10 !w-10 !min-w-10 !shrink-0 !px-0 !py-0" variant="secondary"
                        icon="settings" :title="app.t('products.edit')" :aria-label="app.t('products.edit')"
                        @click="openEdit(product)" />

                      <AppButton v-if="canDeactivate"
                        class="!box-border !h-10 !min-h-10 !w-28 !min-w-28 !shrink-0 !px-3 !py-0"
                        :variant="product.is_active ? 'danger' : 'secondary'"
                        @click="setActive(product, !product.is_active)">
                        {{ product.is_active ? app.t('products.deactivate') : app.t('products.activate') }}
                      </AppButton>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="grid gap-3 md:hidden">
            <article v-for="product in products" :key="product.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-950/60">
              <div class="flex items-start justify-between gap-3">
                <div class="flex min-w-0 items-start gap-3">
                  <ProductAvatar :src="product.image_url" :updated-at="product.image_updated_at" :name="product.name" size="lg" :muted="!product.is_active" />
                  <div class="min-w-0">
                    <h3 class="truncate font-bold">{{ product.name }}</h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400">{{ product.sku }} · {{ product.barcode || app.t('products.noBarcode') }}</p>
                  </div>
                </div>
                <span class="shrink-0 rounded-full px-2 py-1 text-xs font-bold" :class="stockClass(product.stock_status)">{{ stockLabel(product.stock_status) }}</span>
              </div>
              <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('products.sellingPrice') }}</dt><dd class="font-semibold">{{ money(product.selling_price) }}</dd></div>
                <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('products.stock') }}</dt><dd class="font-semibold">{{ product.total_stock }}</dd></div>
                <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('products.category') }}</dt><dd class="font-semibold">{{ product.category_name || app.t('products.noCategory') }}</dd></div>
                <div><dt class="text-slate-500 dark:text-slate-400">{{ app.t('products.status') }}</dt><dd class="font-semibold">{{ product.is_active ? app.t('products.active') : app.t('products.inactive') }}</dd></div>
              </dl>
              <div class="mt-3 flex flex-wrap gap-2">
                <AppButton data-stock-tooltip-trigger class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="map-pin" :title="app.t('products.stocks')" :aria-label="app.t('products.stocks')" @click="showStocks($event, product)" />
                <AppButton v-if="canUpdate" class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('products.edit')" :aria-label="app.t('products.edit')" @click="openEdit(product)" />
              </div>
            </article>
          </div>
        </div>
      </AppCard>
    </div>

    <AppModal :open="modalOpen" :title="modalTitle" :description="app.t('products.formDescription')" :close-label="app.t('products.cancel')" size="xl" @close="closeModal">
      <form class="grid gap-4" @submit.prevent="openSaveConfirm">
        <section class="grid gap-3 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/45">
          <h3 class="text-sm font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('products.basicInfo') }}</h3>
          <div class="grid gap-2 md:grid-cols-[116px_1fr] md:items-end">
            <div class="grid gap-2">
              <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ app.t('products.image') }}</span>
              <ProductAvatar :src="currentImage" :updated-at="currentImageUpdatedAt" :name="form.name" size="xl" shape="square" />
              <input ref="fileInput" class="hidden" type="file" accept="image/jpeg,image/png,image/webp" @change="onImageSelected" />
            </div>
            <div class="grid content-end gap-3">
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ app.t('products.imageHelp') }}</p>
              <div class="flex flex-wrap gap-2">
                <AppButton class="w-full sm:w-auto" type="button" variant="secondary" icon="upload" :disabled="!canUpdate" @click="chooseImage">{{ app.t('products.chooseImage') }}</AppButton>
                <AppButton v-if="imageFile || selectedProduct?.image_url" type="button" variant="danger" icon="x" :loading="imageSaving" :disabled="imageSaving || !canUpdate" @click="removeProductImage">
                  {{ app.t('products.removeImage') }}
                </AppButton>
              </div>
            </div>
          </div>
          <div class="grid gap-3 md:grid-cols-2">
            <AppInput v-model="form.sku" :label="app.t('products.sku')" :placeholder="app.t('products.skuPlaceholder')" :error="fieldErrors.sku" />
            <AppInput v-model="form.name" :label="app.t('products.name')" :placeholder="app.t('products.namePlaceholder')" :error="fieldErrors.name" />
            <div class="grid gap-2 md:col-span-2">
              <AppInput v-model="form.barcode" :label="app.t('products.barcode')" :placeholder="app.t('products.barcodePlaceholder')" />
              <div class="grid gap-2 sm:grid-cols-2">
                <AppButton class="w-full" type="button" variant="secondary" icon="scan-barcode" @click="openBarcodeScan">{{ app.t('products.scanBarcode') }}</AppButton>
                <AppButton class="w-full" type="button" variant="secondary" icon="qr-code" @click="openCameraScanner">{{ app.t('products.cameraScan') }}</AppButton>
              </div>
            </div>
            <AppSelect v-model="form.category_id" :label="app.t('products.category')">
              <option value="">{{ app.t('products.noCategory') }}</option>
              <option v-for="category in categories" :key="category.id" :value="String(category.id)">{{ category.name }}</option>
            </AppSelect>
            <AppInput v-model="form.unit" :label="app.t('products.unit')" :placeholder="app.t('products.unitPlaceholder')" />
          </div>
        </section>

        <section class="grid gap-3 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/45">
          <h3 class="text-sm font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('products.pricing') }}</h3>
          <div class="grid gap-3 md:grid-cols-2">
            <AppInput v-model="form.selling_price" :label="app.t('products.sellingPrice')" type="number" :placeholder="app.t('products.sellingPricePlaceholder')" />
            <AppInput v-model="form.unit_cost" :label="app.t('products.unitCost')" type="number" :placeholder="app.t('products.unitCostPlaceholder')" />
          </div>
        </section>

        <section class="grid gap-3 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/45">
          <h3 class="text-sm font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('products.stockRules') }}</h3>
          <div class="grid gap-3 md:grid-cols-2">
            <AppInput v-model="form.threshold" :label="app.t('products.threshold')" type="number" :placeholder="app.t('products.thresholdPlaceholder')" />
            <AppInput v-model="form.reorder_point" :label="app.t('products.reorderPoint')" type="number" :placeholder="app.t('products.reorderPointPlaceholder')" />
          </div>
          <AppCheckbox v-model="form.is_active" :label="app.t('products.active')" :description="app.t('products.status')" />
        </section>

        <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
        <div class="sticky -bottom-6 -mx-5 -mb-5 flex flex-col-reverse gap-2 border-t border-slate-200 bg-white/90 p-4 backdrop-blur sm:-mx-6 sm:-mb-6 sm:flex-row sm:justify-end dark:border-slate-800 dark:bg-slate-900/90">
          <AppButton class="w-full sm:w-auto" type="button" variant="secondary" :disabled="saving || imageSaving" @click="closeModal">{{ app.t('products.cancel') }}</AppButton>
          <AppButton class="w-full sm:w-auto" type="submit" :loading="saving || imageSaving" :disabled="saving || imageSaving || (!form.id && !canCreate) || (Boolean(form.id) && !canUpdate)" icon="check-circle">{{ app.t('products.save') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="saveConfirmOpen"
      :title="saveConfirmTitle"
      :message="saveConfirmMessage"
      :confirm-label="app.t('products.confirm')"
      :cancel-label="app.t('products.cancel')"
      :destructive="false"
      :loading="saving || imageSaving"
      @close="saveConfirmOpen = false"
      @confirm="save"
    />

    <AppModal :open="importModalOpen" :title="app.t('products.importTitle')" :description="app.t('products.importDescription')" :close-label="app.t('products.cancel')" size="xl" @close="closeProductImport">
      <div class="grid gap-4">
        <section class="grid gap-5 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/45">
          <label class="grid gap-2 text-sm">
            <span class="font-semibold text-slate-700 dark:text-slate-200">{{ app.t('products.uploadCSV') }}</span>
            <label
              class="group flex min-h-36 cursor-pointer flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-brand-700/25 bg-white/80 px-5 py-6 text-center transition hover:border-brand-600 hover:bg-brand-50/70 dark:border-brand-100/15 dark:bg-slate-950/45 dark:hover:border-brand-100/35 dark:hover:bg-brand-900/20">
              <input class="sr-only" type="file" accept=".csv,text/csv" :disabled="importUploading || importConfirming"
                @change="chooseImportFile" />
              <span class="grid gap-1">
                <span class="text-sm font-black text-slate-900 dark:text-slate-50">
                  {{
                    selectedImportFile
                      ? selectedImportFile.name
                      : app.t('products.chooseCSVFile')
                  }}
                </span>

                <span class="text-xs font-semibold text-slate-500 dark:text-slate-400">
                  {{
                    selectedImportFile
                      ? formatFileSize(selectedImportFile.size)
                      : app.t('products.dragOrClickUpload')
                  }}
                </span>
              </span>

              <span
                class="rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 dark:bg-brand-100/10 dark:text-brand-100">
                CSV only
              </span>
            </label>
            <span class="max-w-3xl text-xs leading-5 text-slate-500 dark:text-slate-400">{{ app.t('products.importColumnsHelp') }}</span>
          </label>
          <div class="flex flex-col gap-2 border-t border-slate-200 pt-4 dark:border-slate-800 sm:flex-row sm:flex-wrap sm:justify-end">
            <AppButton type="button" variant="secondary" icon="download" :disabled="!canDownloadImportTemplate" @click="downloadImportTemplate">{{ app.t('products.downloadTemplate') }}</AppButton>
            <AppButton type="button" variant="secondary" icon="search" :loading="importUploading" :disabled="!selectedImportFile || !canPreviewImport || importUploading" @click="previewProductImport">
              {{ importUploading ? app.t('products.previewing') : app.t('products.previewFile') }}
            </AppButton>
            <AppButton type="button" icon="check-circle" :loading="importConfirming" :disabled="!importPreviewJob || importPreviewJob.status !== 'PENDING' || validImportRows === 0 || !canConfirmImport || importConfirming" @click="confirmProductImport">
              {{ importConfirming ? app.t('products.importing') : app.t('products.confirmImport') }}
            </AppButton>
          </div>
        </section>

        <div v-if="importError" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ importError }}</div>

        <section v-if="importPreviewJob" class="rounded-2xl border border-slate-200 bg-white/70 p-4 dark:border-slate-700 dark:bg-slate-950/50">
          <div class="mb-3 flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 class="font-black">{{ app.t('products.importPreview') }}: {{ importPreviewJob.file_name }}</h3>
              <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('products.importPreviewSummary', { valid: validImportRows, invalid: invalidImportRows }) }}</p>
            </div>
            <span class="rounded-full bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600 dark:bg-slate-800 dark:text-slate-200">{{ importPreviewJob.status }}</span>
          </div>
          <div class="max-h-[360px] overflow-auto">
            <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
              <thead class="bg-slate-50 dark:bg-slate-900">
                <tr>
                  <th class="px-3 py-2 text-left">{{ app.t('products.importRow') }}</th>
                  <th class="px-3 py-2 text-left">{{ app.t('products.sku') }}</th>
                  <th class="px-3 py-2 text-left">{{ app.t('products.name') }}</th>
                  <th class="px-3 py-2 text-left">{{ app.t('products.category') }}</th>
                  <th class="px-3 py-2 text-right">{{ app.t('products.sellingPrice') }}</th>
                  <th class="px-3 py-2 text-right">{{ app.t('products.importInitialStock') }}</th>
                  <th class="px-3 py-2 text-left">{{ app.t('products.importStatusError') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                <tr v-for="row in importPreviewJob.rows" :key="row.id" :class="importRowClass(row)">
                  <td class="px-3 py-2">{{ row.row_index }}</td>
                  <td class="px-3 py-2 font-semibold">{{ row.raw_data.sku }}</td>
                  <td class="px-3 py-2">{{ row.raw_data.name }}</td>
                  <td class="px-3 py-2">{{ row.raw_data.category || '-' }}</td>
                  <td class="px-3 py-2 text-right">{{ row.raw_data.selling_price }}</td>
                  <td class="px-3 py-2 text-right">{{ row.raw_data.initial_stock ?? '-' }}</td>
                  <td class="px-3 py-2">
                    <span class="font-bold" :class="row.status === 'FAILED' ? 'text-red-700 dark:text-red-300' : 'text-brand-700 dark:text-emerald-200'">{{ row.status }}</span>
                    <p v-if="row.error_message" class="mt-1 text-xs text-red-700 dark:text-red-300">{{ row.error_message }}</p>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </AppModal>

    <AppModal :open="barcodeScanOpen" :title="app.t('products.barcodeScanTitle')" :description="app.t('products.barcodeScanDescription')" :close-label="app.t('products.cancel')" @close="closeBarcodeScan">
      <div class="grid gap-4">
        <AppInput ref="barcodeScanInput" v-model="barcodeScanValue" :label="app.t('products.barcode')" :placeholder="app.t('products.barcodeScanPlaceholder')" @keyup.enter="applyBarcodeScan" />
        <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <AppButton type="button" variant="secondary" @click="closeBarcodeScan">{{ app.t('products.cancel') }}</AppButton>
          <AppButton type="button" icon="check-circle" :disabled="!barcodeScanValue.trim()" @click="applyBarcodeScan">{{ app.t('products.useBarcode') }}</AppButton>
        </div>
      </div>
    </AppModal>

    <AppModal :open="cameraOpen" :title="app.t('products.cameraTitle')" :close-label="app.t('products.cancel')" @close="closeCameraScanner">
      <div class="grid gap-3">
        <video ref="videoRef" class="aspect-video w-full rounded-lg bg-slate-950 object-cover" muted playsinline />
        <p class="text-sm text-slate-600 dark:text-slate-300">{{ cameraMessage }}</p>
        <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
          <AppInput v-model="barcodeScanValue" :label="app.t('products.manualFallback')" :placeholder="app.t('products.barcodeScanPlaceholder')" @keyup.enter="applyBarcodeScan" />
          <div class="flex items-end"><AppButton class="w-full" type="button" @click="applyBarcodeScan">{{ app.t('products.useBarcode') }}</AppButton></div>
        </div>
      </div>
    </AppModal>

    <Teleport to="body">
      <div
        v-if="stockTooltipProductID && selectedProduct"
        ref="stockTooltipRef"
        class="fixed z-[120] w-72 rounded-xl bg-white p-3 text-left shadow-2xl shadow-slate-950/20 dark:border-slate-700 dark:bg-slate-900 dark:shadow-black/30"
        :style="stockTooltipStyle"
        role="dialog"
        :aria-label="app.t('products.locationStock')"
      >
        <p v-if="stockTooltipLoading" class="py-4 text-center text-sm text-slate-500 dark:text-slate-400">{{ app.t('products.loading') }}</p>
        <div v-else-if="selectedStocks.length" class="grid gap-2">
          <div v-for="stock in selectedStocks" :key="stock.location_id" class="rounded-lg bg-slate-50 p-2 dark:bg-slate-950/60">
            <div class="flex items-center justify-between gap-2 text-sm">
              <span class="min-w-0 truncate font-semibold text-slate-800 dark:text-slate-100">{{ stock.location_name }}</span>
              <span class="font-black text-brand-700 dark:text-emerald-200">{{ stock.quantity }}</span>
            </div>
            <span class="mt-1 inline-flex rounded-full px-2 py-0.5 text-[11px] font-bold" :class="stockClass(stock.stock_status)">{{ stockLabel(stock.stock_status) }}</span>
          </div>
        </div>
        <p v-else class="py-4 text-center text-sm text-slate-500 dark:text-slate-400">{{ app.t('products.selectStock') }}</p>
      </div>
    </Teleport>
  </section>
</template>
