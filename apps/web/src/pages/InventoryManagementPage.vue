<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
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
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { Location, Product, ProductStock, StockTransfer } from '../types/navigation'
import { formatThaiDateTime } from '../utils/date'

type InventoryTab = 'locations' | 'transfers'
type TransferAction = 'complete' | 'cancel'

interface StockTransferPage {
  items: StockTransfer[]
  total: number
  page: number
  page_size: number
}

interface StockTransferOptions {
  products: Product[]
  locations: Location[]
  stocks: ProductStock[]
}

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeTab = ref<InventoryTab>('locations')
const locations = ref<Location[]>([])
const products = ref<Product[]>([])
const stocks = ref<ProductStock[]>([])
const transfers = ref<StockTransfer[]>([])
const selectedTransfer = ref<StockTransfer | null>(null)

const loadingLocations = ref(false)
const loadingTransfers = ref(false)
const loadingOptions = ref(false)
const locationError = ref('')
const transferError = ref('')
const locationModalOpen = ref(false)
const transferModalOpen = ref(false)
const locationSaveConfirmOpen = ref(false)
const transferSaveConfirmOpen = ref(false)
const transferTooltipID = ref<number | null>(null)
const transferTooltipLoading = ref(false)
const transferTooltipRef = ref<HTMLElement | null>(null)
const transferTooltipAnchor = ref<HTMLElement | null>(null)
const transferTooltipPosition = reactive({ top: 0, left: 0 })
const locationSubmitting = ref(false)
const transferSubmitting = ref(false)
const confirmSubmitting = ref(false)
const pendingLocation = ref<Location | null>(null)
const pendingLocationActive = ref(false)
const pendingTransfer = ref<StockTransfer | null>(null)
const pendingTransferAction = ref<TransferAction>('complete')
const transferPage = ref(1)
const transferPageSize = ref(20)
const transferTotal = ref(0)

const locationForm = reactive({
  id: 0,
  name: '',
  description: '',
})

const transferForm = reactive({
  from_location_id: '',
  to_location_id: '',
  product_id: '',
  quantity: 1,
  note: '',
})

const canViewLocations = computed(() => auth.hasPermission('locations.view'))
const canCreateLocation = computed(() => auth.hasPermission('locations.create'))
const canUpdateLocation = computed(() => auth.hasPermission('locations.update'))
const canDeactivateLocation = computed(() => auth.hasPermission('locations.deactivate'))
const canViewTransfers = computed(() => auth.hasPermission('transfers.view'))
const canCreateTransfer = computed(() => auth.hasPermission('transfers.create'))
const canCompleteTransfer = computed(() => auth.hasPermission('transfers.complete'))
const canCancelTransfer = computed(() => auth.hasPermission('transfers.cancel'))
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')
const activeLocations = computed(() => locations.value.filter((location) => location.is_active))
const activeProducts = computed(() => products.value.filter((product) => product.is_active))
const tabs = computed(() => {
  const items: Array<{ key: InventoryTab; label: string }> = []
  if (canViewLocations.value) items.push({ key: 'locations', label: app.t('inventory.tabs.locations') })
  if (canViewTransfers.value) items.push({ key: 'transfers', label: app.t('inventory.tabs.transfers') })
  return items
})
const totalTransferPages = computed(() => Math.max(1, Math.ceil(transferTotal.value / transferPageSize.value)))
const selectedProduct = computed(() => activeProducts.value.find((product) => product.id === Number(transferForm.product_id)) ?? null)
const selectedSourceLocation = computed(() => activeLocations.value.find((location) => location.id === Number(transferForm.from_location_id)) ?? null)
const selectedDestinationLocation = computed(() => activeLocations.value.find((location) => location.id === Number(transferForm.to_location_id)) ?? null)
const sourceStock = computed(() => stockAt(Number(transferForm.product_id), Number(transferForm.from_location_id)))
const destinationStock = computed(() => stockAt(Number(transferForm.product_id), Number(transferForm.to_location_id)))
const sourceAfterTransfer = computed(() => sourceStock.value - Number(transferForm.quantity || 0))
const destinationAfterTransfer = computed(() => destinationStock.value + Number(transferForm.quantity || 0))
const transferSummary = computed(() => ({
  draft: transfers.value.filter((transfer) => transfer.status === 'DRAFT').length,
  completed: transfers.value.filter((transfer) => transfer.status === 'COMPLETED').length,
  cancelled: transfers.value.filter((transfer) => transfer.status === 'CANCELLED').length,
}))
const locationConfirmOpen = computed(() => Boolean(pendingLocation.value))
const transferConfirmOpen = computed(() => Boolean(pendingTransfer.value))
const locationSaveConfirmTitle = computed(() => locationForm.id ? app.t('inventory.locations.confirmUpdateTitle') : app.t('inventory.locations.confirmCreateTitle'))
const locationSaveConfirmMessage = computed(() => t('inventory.locations.confirmSaveMessage', { name: locationForm.name.trim() || app.t('inventory.locations.name') }))
const transferTooltipStyle = computed(() => ({
  top: `${transferTooltipPosition.top}px`,
  left: `${transferTooltipPosition.left}px`,
}))

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function friendlyError(err: unknown, fallback: TranslationKey) {
  const message = err instanceof Error ? err.message : app.t(fallback)
  if (message.toLowerCase().includes('permission')) return app.t('inventory.noPermission')
  return message
}

function formatDate(value: string | null) {
  return formatThaiDateTime(value)
}

function stockAt(productID: number, locationID: number) {
  if (!productID || !locationID) return 0
  return stocks.value.find((stock) => stock.product_id === productID && stock.location_id === locationID)?.quantity ?? 0
}

function productByID(productID: number) {
  return products.value.find((product) => product.id === productID) ?? null
}

function transferStatusTone(status: StockTransfer['status']) {
  if (status === 'COMPLETED') return 'success'
  if (status === 'CANCELLED') return 'danger'
  return 'warning'
}

function transferStatusLabel(status: StockTransfer['status']) {
  return app.t(`inventory.transfer.status.${status.toLowerCase()}` as TranslationKey)
}

function syncActiveTabFromRoute() {
  const requested = route.query.tab === 'transfers' ? 'transfers' : 'locations'
  if (requested === 'transfers' && canViewTransfers.value) activeTab.value = 'transfers'
  else if (requested === 'locations' && canViewLocations.value) activeTab.value = 'locations'
  else activeTab.value = tabs.value[0]?.key ?? 'locations'
}

function setActiveTab(tab: InventoryTab) {
  if (!tabs.value.some((item) => item.key === tab)) return
  activeTab.value = tab
  router.replace({ path: '/inventory-management', query: { ...route.query, tab } })
}

function setTransferPageSize(value: number) {
  transferPageSize.value = value
  transferPage.value = 1
}

async function loadActiveTab() {
  if (activeTab.value === 'locations') await loadLocations()
  if (activeTab.value === 'transfers') {
    await loadTransferOptions()
    await loadTransfers()
  }
}

async function loadLocations() {
  if (!canViewLocations.value) return
  loadingLocations.value = true
  locationError.value = ''
  try {
    locations.value = await apiClient<Location[]>('/v1/locations')
  } catch (err) {
    locationError.value = friendlyError(err, 'inventory.locations.loadFailed')
  } finally {
    loadingLocations.value = false
  }
}

async function loadTransferOptions() {
  if (!canViewTransfers.value || loadingOptions.value) return
  loadingOptions.value = true
  transferError.value = ''
  try {
    const options = await apiClient<StockTransferOptions>('/v1/stock-transfers/options')
    products.value = options.products
    locations.value = options.locations
    stocks.value = options.stocks
    if (!transferForm.product_id && activeProducts.value[0]) transferForm.product_id = String(activeProducts.value[0].id)
    if (!transferForm.from_location_id && activeLocations.value[0]) transferForm.from_location_id = String(activeLocations.value[0].id)
    if (!transferForm.to_location_id && activeLocations.value[1]) transferForm.to_location_id = String(activeLocations.value[1].id)
  } catch (err) {
    transferError.value = friendlyError(err, 'inventory.transfers.optionsFailed')
  } finally {
    loadingOptions.value = false
  }
}

async function loadTransfers() {
  if (!canViewTransfers.value) return
  loadingTransfers.value = true
  transferError.value = ''
  try {
    const params = new URLSearchParams({ page: String(transferPage.value), page_size: String(transferPageSize.value) })
    const result = await apiClient<StockTransferPage>(`/v1/stock-transfers?${params.toString()}`)
    transfers.value = result.items
    transferTotal.value = result.total
    transferPage.value = result.page
    transferPageSize.value = result.page_size
  } catch (err) {
    transferError.value = friendlyError(err, 'inventory.transfers.loadFailed')
  } finally {
    loadingTransfers.value = false
  }
}

function resetLocationForm() {
  locationForm.id = 0
  locationForm.name = ''
  locationForm.description = ''
}

function openCreateLocation() {
  resetLocationForm()
  locationModalOpen.value = true
}

function openEditLocation(location: Location) {
  locationForm.id = location.id
  locationForm.name = location.name
  locationForm.description = location.description
  locationModalOpen.value = true
}

function closeLocationModal(force = false) {
  if (locationSubmitting.value && !force) return
  locationModalOpen.value = false
  locationSaveConfirmOpen.value = false
  resetLocationForm()
}

function requestSaveLocation() {
  if (!locationForm.name.trim()) {
    locationError.value = app.t('inventory.locations.nameRequired')
    return
  }
  locationError.value = ''
  locationSaveConfirmOpen.value = true
}

function closeLocationSaveConfirm() {
  if (locationSubmitting.value) return
  locationSaveConfirmOpen.value = false
}

async function saveLocation() {
  locationSubmitting.value = true
  locationError.value = ''
  const toastMessage = locationForm.id ? app.t('inventory.locations.updateSuccess') : app.t('inventory.locations.createSuccess')
  try {
    const payload = { name: locationForm.name.trim(), description: locationForm.description.trim() }
    if (locationForm.id) await patchJSON<Location>(`/v1/locations/${locationForm.id}`, payload)
    else await postJSON<Location>('/v1/locations', payload)
    locationSaveConfirmOpen.value = false
    closeLocationModal(true)
    app.pushToast({ type: 'success', message: toastMessage })
    await loadLocations()
  } catch (err) {
    locationError.value = friendlyError(err, 'inventory.locations.saveFailed')
    app.pushToast({ type: 'error', message: app.t('inventory.locations.saveFailed'), description: locationError.value })
  } finally {
    locationSubmitting.value = false
  }
}

function confirmLocationStatus(location: Location, active: boolean) {
  pendingLocation.value = location
  pendingLocationActive.value = active
}

async function applyLocationStatus() {
  if (!pendingLocation.value) return
  confirmSubmitting.value = true
  locationError.value = ''
  try {
    await patchJSON<Location>(`/v1/locations/${pendingLocation.value.id}/status`, { is_active: pendingLocationActive.value })
    app.pushToast({
      type: 'success',
      message: pendingLocationActive.value ? app.t('inventory.locations.activateSuccess') : app.t('inventory.locations.deactivateSuccess'),
      description: pendingLocation.value.name,
    })
    pendingLocation.value = null
    await loadLocations()
  } catch (err) {
    locationError.value = friendlyError(err, 'inventory.locations.statusFailed')
    app.pushToast({ type: 'error', message: app.t('inventory.locations.statusFailed'), description: locationError.value })
  } finally {
    confirmSubmitting.value = false
  }
}

function resetTransferForm() {
  transferForm.product_id = activeProducts.value[0] ? String(activeProducts.value[0].id) : ''
  transferForm.from_location_id = activeLocations.value[0] ? String(activeLocations.value[0].id) : ''
  transferForm.to_location_id = activeLocations.value[1] ? String(activeLocations.value[1].id) : ''
  transferForm.quantity = 1
  transferForm.note = ''
}

async function openCreateTransfer() {
  await loadTransferOptions()
  resetTransferForm()
  transferModalOpen.value = true
}

function closeTransferModal(force = false) {
  if (transferSubmitting.value && !force) return
  transferModalOpen.value = false
  transferSaveConfirmOpen.value = false
  resetTransferForm()
}

function validateTransfer() {
  if (!transferForm.from_location_id || !transferForm.to_location_id) return app.t('inventory.transfers.locationsRequired')
  if (transferForm.from_location_id === transferForm.to_location_id) return app.t('inventory.transfers.sameLocation')
  if (!transferForm.product_id) return app.t('inventory.transfers.productRequired')
  if (Number(transferForm.quantity) <= 0) return app.t('inventory.transfers.quantityRequired')
  if (Number(transferForm.quantity) > sourceStock.value) return app.t('inventory.transfers.insufficientStock')
  return ''
}

function requestCreateTransfer() {
  const validation = validateTransfer()
  if (validation) {
    transferError.value = validation
    return
  }
  transferError.value = ''
  transferSaveConfirmOpen.value = true
}

function closeTransferSaveConfirm() {
  if (transferSubmitting.value) return
  transferSaveConfirmOpen.value = false
}

async function createTransfer() {
  transferSubmitting.value = true
  transferError.value = ''
  try {
    const transfer = await postJSON<StockTransfer>('/v1/stock-transfers', {
      from_location_id: Number(transferForm.from_location_id),
      to_location_id: Number(transferForm.to_location_id),
      note: transferForm.note.trim(),
      items: [{ product_id: Number(transferForm.product_id), quantity: Number(transferForm.quantity) }],
    })
    transferSaveConfirmOpen.value = false
    app.pushToast({ type: 'success', message: app.t('inventory.transfers.createSuccess'), description: transfer.transfer_no })
    selectedTransfer.value = transfer
    closeTransferModal(true)
    transferPage.value = 1
    await loadTransfers()
  } catch (err) {
    transferError.value = friendlyError(err, 'inventory.transfers.saveFailed')
    app.pushToast({ type: 'error', message: app.t('inventory.transfers.saveFailed'), description: transferError.value })
  } finally {
    transferSubmitting.value = false
  }
}

function closeTransferTooltip() {
  transferTooltipID.value = null
  transferTooltipAnchor.value = null
}

function positionTransferTooltip(target: HTMLElement, tooltip?: HTMLElement | null) {
  const rect = target.getBoundingClientRect()
  const width = 320
  const height = tooltip?.offsetHeight || 220
  const gap = 8
  transferTooltipPosition.left = Math.max(16, Math.min(window.innerWidth - width - 16, rect.right - width))
  transferTooltipPosition.top = Math.max(16, Math.min(window.innerHeight - height - 16, rect.top))
  if (rect.bottom + height + gap <= window.innerHeight) transferTooltipPosition.top = rect.bottom + gap
}

function handleDocumentPointerDown(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  if (!target || !transferTooltipID.value) return
  if (transferTooltipRef.value?.contains(target) || target.closest('[data-transfer-tooltip-trigger]')) return
  closeTransferTooltip()
}

function handleTooltipViewportChange() {
  if (!transferTooltipID.value || !transferTooltipAnchor.value) return
  positionTransferTooltip(transferTooltipAnchor.value, transferTooltipRef.value)
}

async function openTransferDetail(event: MouseEvent, transfer: StockTransfer) {
  const target = event.currentTarget as HTMLElement
  if (transferTooltipID.value === transfer.id) {
    closeTransferTooltip()
    return
  }
  selectedTransfer.value = transfer
  transferTooltipAnchor.value = target
  positionTransferTooltip(target)
  transferTooltipID.value = transfer.id
  transferTooltipLoading.value = true
  await nextTick()
  positionTransferTooltip(target, transferTooltipRef.value)
  try {
    selectedTransfer.value = await apiClient<StockTransfer>(`/v1/stock-transfers/${transfer.id}`)
    await nextTick()
    if (transferTooltipAnchor.value) positionTransferTooltip(transferTooltipAnchor.value, transferTooltipRef.value)
  } catch (err) {
    transferError.value = friendlyError(err, 'inventory.transfers.loadFailed')
    closeTransferTooltip()
    app.pushToast({ type: 'error', message: app.t('inventory.transfers.loadFailed'), description: transferError.value })
  } finally {
    transferTooltipLoading.value = false
  }
}

function confirmTransferAction(transfer: StockTransfer, action: TransferAction) {
  pendingTransfer.value = transfer
  pendingTransferAction.value = action
}

async function applyTransferAction() {
  if (!pendingTransfer.value) return
  confirmSubmitting.value = true
  transferError.value = ''
  try {
    const action = pendingTransferAction.value
    const transfer = await postJSON<StockTransfer>(`/v1/stock-transfers/${pendingTransfer.value.id}/${action}`, {})
    app.pushToast({
      type: 'success',
      message: action === 'complete' ? app.t('inventory.transfers.completeSuccess') : app.t('inventory.transfers.cancelSuccess'),
      description: transfer.transfer_no,
    })
    pendingTransfer.value = null
    selectedTransfer.value = transfer
    await loadTransferOptions()
    await loadTransfers()
  } catch (err) {
    transferError.value = friendlyError(err, 'inventory.transfers.actionFailed')
    app.pushToast({ type: 'error', message: app.t('inventory.transfers.actionFailed'), description: transferError.value })
  } finally {
    confirmSubmitting.value = false
  }
}

watch(() => route.query.tab, () => {
  syncActiveTabFromRoute()
  loadActiveTab()
})

watch([transferPage, transferPageSize], () => {
  if (activeTab.value === 'transfers') loadTransfers()
})

onMounted(() => {
  document.addEventListener('mousedown', handleDocumentPointerDown)
  window.addEventListener('resize', handleTooltipViewportChange)
  window.addEventListener('scroll', handleTooltipViewportChange, true)
  syncActiveTabFromRoute()
  loadActiveTab()
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleDocumentPointerDown)
  window.removeEventListener('resize', handleTooltipViewportChange)
  window.removeEventListener('scroll', handleTooltipViewportChange, true)
})
</script>

<template>
  <div class="space-y-6">
    <PageHeader :title="app.t('inventory.title')" :eyebrow="app.t('inventory.eyebrow')" :description="app.t('inventory.description')" icon="map-pin" />

    <AppTabs v-if="tabs.length > 1" :tabs="tabs" :model-value="activeTab" @update:model-value="setActiveTab" />

    <div v-if="activeTab === 'locations'" class=" grid gap-4">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-xl font-black text-slate-950 dark:text-slate-50">{{ app.t('inventory.tabs.locations') }}</h2>
        </div>
        <div class="flex flex-wrap gap-2">
          <AppButton v-if="canCreateLocation" icon="plus" @click="openCreateLocation">
            {{ app.t('inventory.locations.add') }}
          </AppButton>
        </div>
      </div>
      <AppCard class="dark:bg-slate-900/80">
        <div v-if="locationError" class="mb-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-400/30 dark:bg-red-500/10 dark:text-red-100">{{ locationError }}</div>
        <AppLoadingState v-if="loadingLocations" :label="app.t('inventory.locations.loading')" />
        <AppEmptyState v-else-if="locations.length === 0" :title="app.t('inventory.locations.empty')" :description="app.t('inventory.locations.emptyDescription')" icon="map-pin">
        </AppEmptyState>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-700">
            <thead class="text-left text-xs uppercase text-slate-500 dark:text-slate-400">
              <tr>
                <th class="px-3 py-3">{{ app.t('inventory.locations.name') }}</th>
                <th class="px-3 py-3">{{ app.t('inventory.locations.description') }}</th>
                <th class="px-3 py-3">{{ app.t('inventory.locations.status') }}</th>
                <th class="px-3 py-3">{{ app.t('inventory.locations.createdAt') }}</th>
                <th class="px-3 py-3 text-right">{{ app.t('inventory.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="location in locations" :key="location.id" class="align-top">
                <td class="px-3 py-3 font-bold text-slate-900 dark:text-slate-50">{{ location.name }}</td>
                <td class="max-w-md px-3 py-3 text-slate-500 dark:text-slate-300">{{ location.description || '-' }}</td>
                <td class="px-3 py-3">
                  <AppBadge :tone="location.is_active ? 'success' : 'neutral'">
                    {{ location.is_active ? app.t('inventory.status.active') : app.t('inventory.status.inactive') }}
                  </AppBadge>
                </td>
                <td class="px-3 py-3 text-slate-500 dark:text-slate-300">{{ formatDate(location.created_at) }}</td>
                <td class="px-3 py-3">
                  <div class="flex justify-end gap-2">
                    <AppButton v-if="canUpdateLocation" class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('inventory.edit')" :aria-label="app.t('inventory.edit')" @click="openEditLocation(location)" />
                    <AppButton
                      v-if="canDeactivateLocation"
                      :variant="location.is_active ? 'danger' : 'secondary'"
                      @click="confirmLocationStatus(location, !location.is_active)"
                    >
                      {{ location.is_active ? app.t('inventory.deactivate') : app.t('inventory.activate') }}
                    </AppButton>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </AppCard>
    </div>

    <div v-else class=" grid gap-4">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-xl font-black text-slate-950 dark:text-slate-50">{{ app.t('inventory.tabs.transfers') }}</h2>
        </div>
        <div class="flex flex-wrap gap-2">
          <AppButton v-if="canCreateTransfer" icon="arrow-left-right" @click="openCreateTransfer">
            {{ app.t('inventory.transfers.add') }}
          </AppButton>
        </div>
      </div>

      <div class="grid gap-3 md:grid-cols-3">
        <StatCard :label="app.t('inventory.transfer.status.draft')" :value="transferSummary.draft" :helper="app.t('inventory.transfers.draftHelper')" icon="history" />
        <StatCard :label="app.t('inventory.transfer.status.completed')" :value="transferSummary.completed" :helper="app.t('inventory.transfers.completedHelper')" icon="check-circle" tone="success" />
        <StatCard :label="app.t('inventory.transfer.status.cancelled')" :value="transferSummary.cancelled" :helper="app.t('inventory.transfers.cancelledHelper')" icon="circle-x" tone="danger" />
      </div>

      <AppCard class="dark:bg-slate-900/80">
        <div v-if="transferError" class="mb-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-400/30 dark:bg-red-500/10 dark:text-red-100">{{ transferError }}</div>
        <AppLoadingState v-if="loadingTransfers || loadingOptions" :label="app.t('inventory.transfers.loading')" />
        <AppEmptyState v-else-if="transfers.length === 0" :title="app.t('inventory.transfers.empty')" :description="app.t('inventory.transfers.emptyDescription')" icon="arrow-left-right">
        </AppEmptyState>
        <div v-else class="space-y-4">
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-700">
              <thead class="text-left text-xs uppercase text-slate-500 dark:text-slate-400">
                <tr>
                  <th class="px-3 py-3">{{ app.t('inventory.transfers.transferNo') }}</th>
                  <th class="px-3 py-3">{{ app.t('inventory.transfers.route') }}</th>
                  <th class="px-3 py-3">{{ app.t('inventory.transfers.product') }}</th>
                  <th class="px-3 py-3">{{ app.t('inventory.transfers.quantity') }}</th>
                  <th class="px-3 py-3">{{ app.t('inventory.transfers.status') }}</th>
                  <th class="px-3 py-3">{{ app.t('inventory.transfers.createdAt') }}</th>
                  <th class="px-3 py-3 text-right">{{ app.t('inventory.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                <tr v-for="transfer in transfers" :key="transfer.id" class="align-top">
                  <td class="px-3 py-3 font-bold text-slate-900 dark:text-slate-50">{{ transfer.transfer_no }}</td>
                  <td class="px-3 py-3 text-slate-600 dark:text-slate-300">
                    <div class="font-semibold">{{ transfer.from_location_name }}</div>
                    <div class="text-xs text-slate-400">{{ app.t('inventory.transfers.to') }} {{ transfer.to_location_name }}</div>
                  </td>
                  <td class="px-3 py-3">
                    <div v-for="item in transfer.items" :key="`${transfer.id}-${item.product_id}`" class="flex items-center gap-3">
                      <ProductAvatar :src="productByID(item.product_id)?.image_url" :updated-at="productByID(item.product_id)?.image_updated_at" :name="item.product_name" size="sm" />
                      <div>
                        <div class="font-bold text-slate-900 dark:text-slate-50">{{ item.product_name }}</div>
                        <div class="text-xs text-slate-500">{{ item.sku }}</div>
                      </div>
                    </div>
                  </td>
                  <td class="px-3 py-3 font-bold">{{ transfer.items.reduce((sum, item) => sum + item.quantity, 0).toLocaleString(locale) }}</td>
                  <td class="px-3 py-3">
                    <AppBadge :tone="transferStatusTone(transfer.status)">{{ transferStatusLabel(transfer.status) }}</AppBadge>
                  </td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-300">{{ formatDate(transfer.created_at) }}</td>
                  <td class="px-3 py-3">
                    <div class="flex justify-end gap-2">
                      <AppButton data-transfer-tooltip-trigger class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="info" :title="app.t('inventory.view')" :aria-label="app.t('inventory.view')" @click="openTransferDetail($event, transfer)" />
                      <AppButton v-if="transfer.status === 'DRAFT' && canCompleteTransfer" icon="check-circle" @click="confirmTransferAction(transfer, 'complete')">{{ app.t('inventory.transfers.complete') }}</AppButton>
                      <AppButton v-if="transfer.status === 'DRAFT' && canCancelTransfer" variant="danger" icon="circle-x" @click="confirmTransferAction(transfer, 'cancel')">{{ app.t('inventory.transfers.cancel') }}</AppButton>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm dark:border-slate-800 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-center gap-2">
              <span class="text-slate-500 dark:text-slate-400">{{ app.t('inventory.show') }}</span>
              <AppPageSizeSelect :model-value="transferPageSize" @update:model-value="setTransferPageSize" />
              <span class="text-slate-500 dark:text-slate-400">{{ app.t('inventory.perPage') }}</span>
            </div>
            <div class="flex items-center justify-end gap-2">
              <span class="font-semibold text-slate-600 dark:text-slate-300">{{ t('inventory.page', { page: transferPage, total: totalTransferPages }) }}</span>
              <AppButton variant="secondary" :disabled="transferPage <= 1" @click="transferPage--">{{ app.t('inventory.previous') }}</AppButton>
              <AppButton variant="secondary" :disabled="transferPage >= totalTransferPages" @click="transferPage++">{{ app.t('inventory.next') }}</AppButton>
            </div>
          </div>
        </div>
      </AppCard>

    </div>

    <AppModal
      :open="locationModalOpen"
      size="lg"
      :title="locationForm.id ? app.t('inventory.locations.editTitle') : app.t('inventory.locations.createTitle')"
      :description="app.t('inventory.locations.modalDescription')"
      @close="closeLocationModal"
    >
      <form class="space-y-4" @submit.prevent="requestSaveLocation">
        <AppInput v-model="locationForm.name" :label="app.t('inventory.locations.name')" :placeholder="app.t('inventory.locations.namePlaceholder')" />
        <AppTextarea v-model="locationForm.description" :label="app.t('inventory.locations.description')" :placeholder="app.t('inventory.locations.descriptionPlaceholder')" />
        <div class="flex justify-end gap-2">
          <AppButton variant="secondary" :disabled="locationSubmitting" @click="closeLocationModal">{{ app.t('inventory.cancel') }}</AppButton>
          <AppButton type="submit" icon="check-circle" :loading="locationSubmitting">{{ app.t('inventory.save') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <AppModal
      :open="transferModalOpen"
      size="xl"
      :title="app.t('inventory.transfers.createTitle')"
      :description="app.t('inventory.transfers.modalDescription')"
      @close="closeTransferModal"
    >
      <form class="space-y-5" @submit.prevent="requestCreateTransfer">
        <div class="grid gap-4 md:grid-cols-2">
          <AppSelect v-model="transferForm.from_location_id" :label="app.t('inventory.transfers.source')">
            <option value="">{{ app.t('inventory.select') }}</option>
            <option v-for="location in activeLocations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <AppSelect v-model="transferForm.to_location_id" :label="app.t('inventory.transfers.destination')">
            <option value="">{{ app.t('inventory.select') }}</option>
            <option v-for="location in activeLocations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <AppSelect v-model="transferForm.product_id" :label="app.t('inventory.transfers.product')">
            <option value="">{{ app.t('inventory.select') }}</option>
            <option v-for="product in activeProducts" :key="product.id" :value="String(product.id)">{{ product.sku }} - {{ product.name }}</option>
          </AppSelect>
          <AppInput v-model="transferForm.quantity" type="number" min="1" step="1" :label="app.t('inventory.transfers.quantity')" :placeholder="app.t('inventory.transfers.quantityPlaceholder')" />
        </div>

        <div class="grid gap-3 md:grid-cols-[1.3fr_1fr_1fr]">
          <div class="rounded-xl border border-slate-200 bg-white/80 p-3 dark:border-slate-700 dark:bg-slate-950/60">
            <div class="flex items-center gap-3">
              <ProductAvatar :src="selectedProduct?.image_url" :updated-at="selectedProduct?.image_updated_at" :name="selectedProduct?.name" size="lg" />
              <div>
                <p class="text-xs font-bold uppercase text-slate-500 dark:text-slate-400">{{ app.t('inventory.transfers.selectedProduct') }}</p>
                <p class="font-black text-slate-950 dark:text-slate-50">{{ selectedProduct?.name ?? '-' }}</p>
                <p class="text-xs text-slate-500">{{ selectedProduct?.sku ?? '-' }}</p>
              </div>
            </div>
          </div>
          <div class="rounded-xl border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-900">
            <p class="text-xs font-bold uppercase text-slate-500 dark:text-slate-400">{{ selectedSourceLocation?.name ?? app.t('inventory.transfers.source') }}</p>
            <p class="mt-2 text-2xl font-black text-slate-950 dark:text-slate-50">{{ sourceStock.toLocaleString(locale) }}</p>
            <p class="text-xs text-slate-500">{{ t('inventory.transfers.afterSource', { quantity: sourceAfterTransfer }) }}</p>
          </div>
          <div class="rounded-xl border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-900">
            <p class="text-xs font-bold uppercase text-slate-500 dark:text-slate-400">{{ selectedDestinationLocation?.name ?? app.t('inventory.transfers.destination') }}</p>
            <p class="mt-2 text-2xl font-black text-slate-950 dark:text-slate-50">{{ destinationStock.toLocaleString(locale) }}</p>
            <p class="text-xs text-slate-500">{{ t('inventory.transfers.afterDestination', { quantity: destinationAfterTransfer }) }}</p>
          </div>
        </div>

        <AppTextarea v-model="transferForm.note" :label="app.t('inventory.transfers.note')" :placeholder="app.t('inventory.transfers.notePlaceholder')" />
        <div v-if="transferError" class="rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-400/30 dark:bg-red-500/10 dark:text-red-100">{{ transferError }}</div>
        <div class="flex justify-end gap-2">
          <AppButton variant="secondary" :disabled="transferSubmitting" @click="closeTransferModal">{{ app.t('inventory.cancel') }}</AppButton>
          <AppButton type="submit" icon="check-circle" :loading="transferSubmitting">{{ app.t('inventory.transfers.saveDraft') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="locationSaveConfirmOpen"
      :title="locationSaveConfirmTitle"
      :message="locationSaveConfirmMessage"
      :confirm-label="app.t('inventory.save')"
      :cancel-label="app.t('inventory.cancel')"
      :loading="locationSubmitting"
      @close="closeLocationSaveConfirm"
      @confirm="saveLocation"
    />

    <ConfirmDialog
      :open="transferSaveConfirmOpen"
      :title="app.t('inventory.transfers.confirmCreateTitle')"
      :message="t('inventory.transfers.confirmCreateMessage', { product: selectedProduct?.name ?? app.t('inventory.transfers.product'), quantity: Number(transferForm.quantity || 0).toLocaleString(locale) })"
      :confirm-label="app.t('inventory.transfers.saveDraft')"
      :cancel-label="app.t('inventory.cancel')"
      :loading="transferSubmitting"
      @close="closeTransferSaveConfirm"
      @confirm="createTransfer"
    />

    <ConfirmDialog
      :open="locationConfirmOpen"
      :title="pendingLocationActive ? app.t('inventory.locations.activateTitle') : app.t('inventory.locations.deactivateTitle')"
      :message="pendingLocation ? t('inventory.locations.statusMessage', { name: pendingLocation.name }) : ''"
      :confirm-label="pendingLocationActive ? app.t('inventory.activate') : app.t('inventory.deactivate')"
      :destructive="!pendingLocationActive"
      :loading="confirmSubmitting"
      :cancel-label="app.t('inventory.cancel')"
      @close="pendingLocation = null"
      @confirm="applyLocationStatus"
    />

    <ConfirmDialog
      :open="transferConfirmOpen"
      :title="pendingTransferAction === 'complete' ? app.t('inventory.transfers.completeTitle') : app.t('inventory.transfers.cancelTitle')"
      :message="pendingTransfer ? t('inventory.transfers.actionMessage', { transfer: pendingTransfer.transfer_no }) : ''"
      :confirm-label="pendingTransferAction === 'complete' ? app.t('inventory.transfers.complete') : app.t('inventory.transfers.cancel')"
      :destructive="pendingTransferAction === 'cancel'"
      :loading="confirmSubmitting"
      :cancel-label="app.t('inventory.cancel')"
      @close="pendingTransfer = null"
      @confirm="applyTransferAction"
    />

    <Teleport to="body">
      <div
        v-if="transferTooltipID && selectedTransfer"
        ref="transferTooltipRef"
        class="fixed z-[120] w-80 rounded-xl bg-white p-3 text-left shadow-2xl shadow-slate-950/20 dark:border dark:border-slate-700 dark:bg-slate-900 dark:shadow-black/30"
        :style="transferTooltipStyle"
        role="dialog"
        :aria-label="app.t('inventory.transfers.detail')"
      >
        <p v-if="transferTooltipLoading" class="py-4 text-center text-sm text-slate-500 dark:text-slate-400">{{ app.t('inventory.transfers.loading') }}</p>
        <div v-else class="grid gap-3">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="truncate text-sm font-black text-slate-950 dark:text-slate-50">{{ selectedTransfer.transfer_no }}</p>
              <p class="mt-0.5 text-xs text-slate-500 dark:text-slate-400">{{ selectedTransfer.from_location_name }} -> {{ selectedTransfer.to_location_name }}</p>
            </div>
            <AppBadge :tone="transferStatusTone(selectedTransfer.status)">{{ transferStatusLabel(selectedTransfer.status) }}</AppBadge>
          </div>
          <div class="grid gap-2">
            <div v-for="item in selectedTransfer.items" :key="item.id ?? item.product_id" class="rounded-lg bg-slate-50 p-2 dark:bg-slate-950/60">
              <div class="flex items-center justify-between gap-3">
                <div class="flex min-w-0 items-center gap-2">
                  <ProductAvatar :src="productByID(item.product_id)?.image_url" :updated-at="productByID(item.product_id)?.image_updated_at" :name="item.product_name" size="sm" />
                  <div class="min-w-0">
                    <p class="truncate text-sm font-semibold text-slate-800 dark:text-slate-100">{{ item.product_name }}</p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">{{ item.sku }}</p>
                  </div>
                </div>
                <span class="font-black text-brand-700 dark:text-emerald-200">{{ item.quantity.toLocaleString(locale) }}</span>
              </div>
            </div>
          </div>
          <p v-if="selectedTransfer.note" class="rounded-lg bg-slate-50 p-2 text-xs text-slate-600 dark:bg-slate-950/60 dark:text-slate-300">{{ selectedTransfer.note }}</p>
        </div>
      </div>
    </Teleport>
  </div>
</template>
