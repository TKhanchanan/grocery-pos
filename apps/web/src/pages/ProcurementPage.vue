<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiClient, patchJSON, postJSON } from '../api/client'
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
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { Location, Product, PurchaseOrder, PurchaseOrderItem, Supplier } from '../types/navigation'
import { formatAppDateTime } from '../utils/date'

type ProcurementTab = 'purchase-orders' | 'suppliers'
type POAction = 'send' | 'receive' | 'cancel'

interface ConfirmState {
  type: 'po-action' | 'supplier-status' | ''
  po: PurchaseOrder | null
  supplier: Supplier | null
  action: POAction | ''
}

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const purchaseOrders = ref<PurchaseOrder[]>([])
const suppliers = ref<Supplier[]>([])
const supplierOptions = ref<Supplier[]>([])
const locations = ref<Location[]>([])
const products = ref<Product[]>([])
const selectedPO = ref<PurchaseOrder | null>(null)
const loadingPO = ref(false)
const loadingSuppliers = ref(false)
const savingPO = ref(false)
const savingSupplier = ref(false)
const confirming = ref(false)
const poError = ref('')
const supplierError = ref('')
const poModalOpen = ref(false)
const poDetailOpen = ref(false)
const supplierModalOpen = ref(false)
const poSaveConfirmOpen = ref(false)
const supplierSaveConfirmOpen = ref(false)
const editingPOID = ref<number | null>(null)
const editingSupplierID = ref<number | null>(null)
const poPage = ref(1)
const poPageSize = ref(20)
const supplierPage = ref(1)
const supplierPageSize = ref(20)
const confirmState = reactive<ConfirmState>({ type: '', po: null, supplier: null, action: '' })

const poFilters = reactive({
  search: '',
  supplier_id: '',
  location_id: '',
  status: '',
  date_from: '',
  date_to: '',
})

const supplierFilters = reactive({
  search: '',
  status: '',
})

const poForm = reactive({
  supplier_id: '',
  location_id: '',
  note: '',
  items: [] as PurchaseOrderItem[],
})

const supplierForm = reactive({
  name: '',
  phone: '',
  email: '',
  address: '',
})

const canViewPO = computed(() => auth.hasPermission('purchase_orders.view'))
const canCreatePO = computed(() => auth.hasPermission('purchase_orders.create'))
const canUpdatePO = computed(() => auth.hasPermission('purchase_orders.update'))
const canSendPO = computed(() => auth.hasPermission('purchase_orders.send'))
const canReceivePO = computed(() => auth.hasPermission('purchase_orders.receive'))
const canCancelPO = computed(() => auth.hasPermission('purchase_orders.cancel'))
const canViewSuppliers = computed(() => auth.hasPermission('suppliers.view'))
const canCreateSupplier = computed(() => auth.hasPermission('suppliers.create'))
const canUpdateSupplier = computed(() => auth.hasPermission('suppliers.update'))
const canDeactivateSupplier = computed(() => auth.hasPermission('suppliers.deactivate'))
const activeSuppliers = computed(() => supplierOptions.value.filter((supplier) => supplier.is_active))
const activeSupplierRows = computed(() => suppliers.value.filter((supplier) => supplier.is_active))
const locale = computed(() => app.language === 'th' ? 'th-TH' : 'en-US')
const activeTab = ref<ProcurementTab>(canViewPO.value ? 'purchase-orders' : 'suppliers')

const availableTabs = computed(() => [
  ...(canViewPO.value ? [{ key: 'purchase-orders' as const, label: app.t('procurement.tabs.purchaseOrders') }] : []),
  ...(canViewSuppliers.value ? [{ key: 'suppliers' as const, label: app.t('procurement.tabs.suppliers') }] : []),
])

const poTotalPages = computed(() => Math.max(1, Math.ceil(purchaseOrders.value.length / poPageSize.value)))
const supplierTotalPages = computed(() => Math.max(1, Math.ceil(suppliers.value.length / supplierPageSize.value)))
const visiblePurchaseOrders = computed(() => {
  const start = (poPage.value - 1) * poPageSize.value
  return purchaseOrders.value.slice(start, start + poPageSize.value)
})
const visibleSuppliers = computed(() => {
  const start = (supplierPage.value - 1) * supplierPageSize.value
  return suppliers.value.slice(start, start + supplierPageSize.value)
})
const poSummary = computed(() => ({
  draft: purchaseOrders.value.filter((po) => po.status === 'DRAFT').length,
  sent: purchaseOrders.value.filter((po) => po.status === 'SENT').length,
  received: purchaseOrders.value.filter((po) => po.status === 'RECEIVED').length,
  cancelled: purchaseOrders.value.filter((po) => po.status === 'CANCELLED').length,
}))
const poTotalCost = computed(() => poForm.items.reduce((sum, item) => sum + Number(item.quantity || 0) * Number(item.unit_cost || 0), 0))
const poSaveConfirmTitle = computed(() => editingPOID.value ? app.t('procurement.confirmUpdatePO') : app.t('procurement.confirmCreatePO'))
const poSaveConfirmMessage = computed(() => editingPOID.value ? app.t('procurement.confirmUpdatePOMessage') : app.t('procurement.confirmCreatePOMessage'))
const supplierSaveConfirmTitle = computed(() => editingSupplierID.value ? app.t('procurement.confirmUpdateSupplier') : app.t('procurement.confirmCreateSupplier'))
const supplierSaveConfirmMessage = computed(() => editingSupplierID.value ? app.t('procurement.confirmUpdateSupplierMessage') : app.t('procurement.confirmCreateSupplierMessage'))

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function money(value: number) {
  return t('procurement.currency', { amount: value.toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) })
}

function formatDate(value: string | null) {
  return formatAppDateTime(value, app.language)
}

function friendlyError(err: unknown, fallback: TranslationKey) {
  const message = err instanceof Error ? err.message : app.t(fallback)
  return message.toLowerCase().includes('permission') ? app.t('procurement.noPermission') : message
}

function statusLabel(status: PurchaseOrder['status']) {
  return app.t(`procurement.status.${status}` as TranslationKey)
}

function statusTone(status: PurchaseOrder['status']) {
  if (status === 'CANCELLED') return 'danger'
  if (status === 'RECEIVED') return 'success'
  if (status === 'SENT') return 'info'
  return 'neutral'
}

function syncQuery(tab: ProcurementTab) {
  router.replace({ path: '/procurement', query: { ...route.query, tab } })
}

function resetTabState() {
  poError.value = ''
  supplierError.value = ''
  poPage.value = 1
  supplierPage.value = 1
  poFilters.search = ''
  poFilters.supplier_id = ''
  poFilters.location_id = ''
  poFilters.status = ''
  poFilters.date_from = ''
  poFilters.date_to = ''
  supplierFilters.search = ''
  supplierFilters.status = ''
  selectedPO.value = null
  poDetailOpen.value = false
  poModalOpen.value = false
  supplierModalOpen.value = false
  poSaveConfirmOpen.value = false
  supplierSaveConfirmOpen.value = false
  closeConfirm(true)
}

function setTab(tab: ProcurementTab) {
  if (activeTab.value !== tab) resetTabState()
  syncQuery(tab)
}

function usedProductIDs(exceptIndex = -1) {
  return new Set(poForm.items.map((item, index) => index === exceptIndex ? 0 : Number(item.product_id)).filter(Boolean))
}

function isProductSelected(productID: number, currentIndex: number) {
  return usedProductIDs(currentIndex).has(productID)
}

function defaultItem(excludeSelected = false): PurchaseOrderItem {
  const selectedIDs = excludeSelected ? usedProductIDs() : new Set<number>()
  const routeProduct = products.value.find((item) => item.id === Number(route.query.product_id) && !selectedIDs.has(item.id))
  const product = routeProduct ?? products.value.find((item) => !selectedIDs.has(item.id)) ?? products.value[0]
  return {
    product_id: product?.id ?? 0,
    quantity: 1,
    received_quantity: 0,
    unit_cost: product?.unit_cost ?? 0,
    line_cost: product?.unit_cost ?? 0,
  }
}

function resetPOForm() {
  editingPOID.value = null
  poForm.supplier_id = activeSuppliers.value[0] ? String(activeSuppliers.value[0].id) : ''
  poForm.location_id = route.query.location_id ? String(route.query.location_id) : locations.value[0] ? String(locations.value[0].id) : ''
  poForm.note = route.query.product_id ? app.t('procurement.createdFromAlert') : ''
  poForm.items = [defaultItem()]
  poError.value = ''
}

function openCreatePO() {
  resetPOForm()
  poModalOpen.value = true
}

function openEditPO(po: PurchaseOrder) {
  editingPOID.value = po.id
  poForm.supplier_id = String(po.supplier_id)
  poForm.location_id = String(po.location_id)
  poForm.note = po.note
  poForm.items = po.items.map((item) => ({ ...item }))
  poError.value = ''
  poModalOpen.value = true
}

function closePOModal(force = false) {
  if (savingPO.value && !force) return
  poModalOpen.value = false
  poSaveConfirmOpen.value = false
  resetPOForm()
}

function addPOItem() {
  const nextItem = defaultItem(true)
  if (Number(nextItem.product_id) && usedProductIDs().has(Number(nextItem.product_id))) {
    poError.value = app.t('procurement.allProductsSelected')
    app.pushToast({ type: 'warning', message: app.t('procurement.allProductsSelected') })
    return
  }
  poError.value = ''
  poForm.items.push(nextItem)
}

function removePOItem(index: number) {
  poForm.items.splice(index, 1)
  if (poForm.items.length === 0) addPOItem()
}

function syncProductCost(item: PurchaseOrderItem) {
  const product = products.value.find((candidate) => candidate.id === Number(item.product_id))
  if (product && !Number(item.unit_cost)) item.unit_cost = product.unit_cost
}

function validatePO() {
  if (!poForm.supplier_id) return app.t('procurement.supplierRequired')
  if (!poForm.location_id) return app.t('procurement.locationRequired')
  if (poForm.items.length === 0) return app.t('procurement.itemsRequired')
  const selected = new Set<number>()
  for (const item of poForm.items) {
    const productID = Number(item.product_id)
    if (!productID) return app.t('procurement.productRequired')
    if (selected.has(productID)) return app.t('procurement.duplicateProduct')
    selected.add(productID)
    if (Number(item.quantity) <= 0) return app.t('procurement.quantityRequired')
    if (Number(item.unit_cost) < 0) return app.t('procurement.costRequired')
  }
  return ''
}

function requestSavePO() {
  const validation = validatePO()
  if (validation) {
    poError.value = validation
    return
  }
  poError.value = ''
  poSaveConfirmOpen.value = true
}

function closePOSaveConfirm() {
  if (savingPO.value) return
  poSaveConfirmOpen.value = false
}

async function savePO() {
  savingPO.value = true
  poError.value = ''
  const successMessage = editingPOID.value ? app.t('procurement.poUpdated') : app.t('procurement.poCreated')
  try {
    const payload = {
      supplier_id: Number(poForm.supplier_id),
      location_id: Number(poForm.location_id),
      note: poForm.note,
      items: poForm.items.map((item) => ({
        product_id: Number(item.product_id),
        quantity: Number(item.quantity),
        unit_cost: Number(item.unit_cost),
      })),
    }
    selectedPO.value = editingPOID.value
      ? await patchJSON<PurchaseOrder>(`/v1/purchase-orders/${editingPOID.value}`, payload)
      : await postJSON<PurchaseOrder>('/v1/purchase-orders', payload)
    await loadPurchaseOrders()
    poSaveConfirmOpen.value = false
    closePOModal(true)
    app.pushToast({ type: 'success', message: successMessage })
  } catch (err) {
    poError.value = friendlyError(err, 'procurement.poSaveFailed')
    app.pushToast({ type: 'error', message: app.t('procurement.poSaveFailed'), description: poError.value })
  } finally {
    savingPO.value = false
  }
}

async function showPO(po: PurchaseOrder) {
  try {
    selectedPO.value = await apiClient<PurchaseOrder>(`/v1/purchase-orders/${po.id}`)
    poDetailOpen.value = true
  } catch (err) {
    poError.value = friendlyError(err, 'procurement.poLoadFailed')
    app.pushToast({ type: 'error', message: app.t('procurement.poLoadFailed'), description: poError.value })
  }
}

function closePODetail() {
  poDetailOpen.value = false
}

function openPOConfirm(po: PurchaseOrder, action: POAction) {
  confirmState.type = 'po-action'
  confirmState.po = po
  confirmState.supplier = null
  confirmState.action = action
}

function closeConfirm(force = false) {
  if (confirming.value && !force) return
  confirmState.type = ''
  confirmState.po = null
  confirmState.supplier = null
  confirmState.action = ''
}

async function confirmPOAction() {
  if (!confirmState.po || !confirmState.action) return
  confirming.value = true
  poError.value = ''
  try {
    selectedPO.value = await postJSON<PurchaseOrder>(`/v1/purchase-orders/${confirmState.po.id}/${confirmState.action}`, {})
    await loadPurchaseOrders()
    app.pushToast({ type: 'success', message: app.t(`procurement.${confirmState.action}Success` as TranslationKey) })
    closeConfirm(true)
  } catch (err) {
    poError.value = friendlyError(err, 'procurement.poActionFailed')
    app.pushToast({ type: 'error', message: app.t('procurement.poActionFailed'), description: poError.value })
  } finally {
    confirming.value = false
  }
}

function resetSupplierForm() {
  editingSupplierID.value = null
  supplierForm.name = ''
  supplierForm.phone = ''
  supplierForm.email = ''
  supplierForm.address = ''
  supplierError.value = ''
}

function openCreateSupplier() {
  resetSupplierForm()
  supplierModalOpen.value = true
}

function openEditSupplier(supplier: Supplier) {
  editingSupplierID.value = supplier.id
  supplierForm.name = supplier.name
  supplierForm.phone = supplier.phone
  supplierForm.email = supplier.email
  supplierForm.address = supplier.address
  supplierError.value = ''
  supplierModalOpen.value = true
}

function closeSupplierModal(force = false) {
  if (savingSupplier.value && !force) return
  supplierModalOpen.value = false
  supplierSaveConfirmOpen.value = false
  resetSupplierForm()
}

function validateSupplier() {
  if (!supplierForm.name.trim()) return app.t('procurement.supplierNameRequired')
  if (supplierForm.email.trim() && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(supplierForm.email.trim())) return app.t('procurement.emailInvalid')
  return ''
}

function requestSaveSupplier() {
  const validation = validateSupplier()
  if (validation) {
    supplierError.value = validation
    return
  }
  supplierError.value = ''
  supplierSaveConfirmOpen.value = true
}

function closeSupplierSaveConfirm() {
  if (savingSupplier.value) return
  supplierSaveConfirmOpen.value = false
}

async function saveSupplier() {
  savingSupplier.value = true
  supplierError.value = ''
  const successMessage = editingSupplierID.value ? app.t('procurement.supplierUpdated') : app.t('procurement.supplierCreated')
  try {
    const payload = {
      name: supplierForm.name.trim(),
      phone: supplierForm.phone.trim(),
      email: supplierForm.email.trim(),
      address: supplierForm.address.trim(),
    }
    if (editingSupplierID.value) {
      await patchJSON<Supplier>(`/v1/suppliers/${editingSupplierID.value}`, payload)
    } else {
      await postJSON<Supplier>('/v1/suppliers', payload)
    }
    await Promise.all([loadSuppliers(), loadSupplierOptions()])
    supplierSaveConfirmOpen.value = false
    closeSupplierModal(true)
    app.pushToast({ type: 'success', message: successMessage })
  } catch (err) {
    supplierError.value = friendlyError(err, 'procurement.supplierSaveFailed')
    app.pushToast({ type: 'error', message: app.t('procurement.supplierSaveFailed'), description: supplierError.value })
  } finally {
    savingSupplier.value = false
  }
}

function openSupplierConfirm(supplier: Supplier) {
  confirmState.type = 'supplier-status'
  confirmState.supplier = supplier
  confirmState.po = null
  confirmState.action = ''
}

async function confirmSupplierStatus() {
  if (!confirmState.supplier) return
  confirming.value = true
  supplierError.value = ''
  try {
    await patchJSON<Supplier>(`/v1/suppliers/${confirmState.supplier.id}/status`, { is_active: !confirmState.supplier.is_active })
    await Promise.all([loadSuppliers(), loadSupplierOptions()])
    app.pushToast({ type: 'success', message: app.t('procurement.supplierStatusUpdated') })
    closeConfirm(true)
  } catch (err) {
    supplierError.value = friendlyError(err, 'procurement.supplierStatusFailed')
    app.pushToast({ type: 'error', message: app.t('procurement.supplierStatusFailed'), description: supplierError.value })
  } finally {
    confirming.value = false
  }
}

async function loadPurchaseOrders() {
  if (!canViewPO.value) return
  loadingPO.value = true
  poError.value = ''
  try {
    const params = new URLSearchParams()
    for (const [key, value] of Object.entries(poFilters)) {
      if (value.trim()) params.set(key, value.trim())
    }
    const query = params.toString()
    purchaseOrders.value = await apiClient<PurchaseOrder[]>(`/v1/purchase-orders${query ? `?${query}` : ''}`)
    poPage.value = Math.min(poPage.value, poTotalPages.value)
  } catch (err) {
    poError.value = friendlyError(err, 'procurement.poLoadFailed')
  } finally {
    loadingPO.value = false
  }
}

async function loadSuppliers() {
  if (!canViewSuppliers.value) return
  loadingSuppliers.value = true
  supplierError.value = ''
  try {
    const params = new URLSearchParams()
    for (const [key, value] of Object.entries(supplierFilters)) {
      if (value.trim()) params.set(key, value.trim())
    }
    const query = params.toString()
    suppliers.value = await apiClient<Supplier[]>(`/v1/suppliers${query ? `?${query}` : ''}`)
    supplierPage.value = Math.min(supplierPage.value, supplierTotalPages.value)
  } catch (err) {
    supplierError.value = friendlyError(err, 'procurement.supplierLoadFailed')
  } finally {
    loadingSuppliers.value = false
  }
}

async function loadSupplierOptions() {
  if (!canViewPO.value && !canViewSuppliers.value) return
  supplierOptions.value = await apiClient<Supplier[]>('/v1/suppliers')
}

async function applyPOFilters() {
  poPage.value = 1
  await loadPurchaseOrders()
}

async function resetPOFilters() {
  poFilters.search = ''
  poFilters.supplier_id = ''
  poFilters.location_id = ''
  poFilters.status = ''
  poFilters.date_from = ''
  poFilters.date_to = ''
  poPage.value = 1
  await loadPurchaseOrders()
}

async function applySupplierFilters() {
  supplierPage.value = 1
  await loadSuppliers()
}

async function resetSupplierFilters() {
  supplierFilters.search = ''
  supplierFilters.status = ''
  supplierPage.value = 1
  await loadSuppliers()
}

async function loadOptions() {
  const requests: Promise<unknown>[] = []
  if (canViewPO.value) {
    requests.push(apiClient<Location[]>('/v1/locations').then((rows) => { locations.value = rows.filter((item) => item.is_active) }))
    requests.push(apiClient<Product[]>('/v1/products').then((rows) => { products.value = rows.filter((item) => item.is_active) }))
  }
  if (canViewPO.value || canViewSuppliers.value) requests.push(loadSupplierOptions())
  if (canViewSuppliers.value) requests.push(loadSuppliers())
  await Promise.all(requests)
  if (poForm.items.length === 0) resetPOForm()
}

async function refreshActiveTab() {
  if (activeTab.value === 'purchase-orders') await loadPurchaseOrders()
  if (activeTab.value === 'suppliers') await loadSuppliers()
}

function changePOPageSize(value: number) {
  poPageSize.value = value
  poPage.value = 1
}

function changeSupplierPageSize(value: number) {
  supplierPageSize.value = value
  supplierPage.value = 1
}

function alignTabWithRoute() {
  const requested = route.query.tab === 'suppliers' ? 'suppliers' : 'purchase-orders'
  const fallback = availableTabs.value[0]?.key
  activeTab.value = availableTabs.value.some((tab) => tab.key === requested) ? requested : fallback
}

onMounted(async () => {
  alignTabWithRoute()
  await loadOptions()
  await loadPurchaseOrders()
})

watch(() => route.query.tab, async () => {
  const previousTab = activeTab.value
  alignTabWithRoute()
  if (previousTab !== activeTab.value) resetTabState()
  await refreshActiveTab()
})
</script>

<template>
  <section class="min-w-0 max-w-full">
    <PageHeader :title="app.t('procurement.title')" :eyebrow="app.t('procurement.eyebrow')" :description="app.t('procurement.description')" icon="clipboard-list" />

    <div class="grid min-w-0 max-w-full gap-4">
      <AppTabs v-if="availableTabs.length > 1" :tabs="availableTabs" :model-value="activeTab" @update:model-value="setTab" />

      <div v-if="activeTab === 'purchase-orders'" class="grid min-w-0 max-w-full gap-4">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-xl font-black text-slate-950 dark:text-slate-50">{{ app.t('procurement.purchaseOrders') }}</h2>
          </div>
          <div class="flex flex-wrap gap-2">
            <AppButton v-if="canCreatePO" icon="plus" @click="openCreatePO">{{ app.t('procurement.createPO') }}</AppButton>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <StatCard :label="app.t('procurement.status.DRAFT')" :value="poSummary.draft" :helper="app.t('procurement.draftHelper')" icon="clipboard-list" />
          <StatCard :label="app.t('procurement.status.SENT')" :value="poSummary.sent" :helper="app.t('procurement.sentHelper')" icon="truck" tone="info" />
          <StatCard :label="app.t('procurement.status.RECEIVED')" :value="poSummary.received" :helper="app.t('procurement.receivedHelper')" icon="package-plus" tone="success" />
          <StatCard :label="app.t('procurement.status.CANCELLED')" :value="poSummary.cancelled" :helper="app.t('procurement.cancelledHelper')" icon="triangle-alert" tone="danger" />
        </div>

        <AppCard class="min-w-0 max-w-full dark:bg-slate-900/80">
          <div class="grid min-w-0 max-w-full gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <AppInput v-model="poFilters.search" :label="app.t('procurement.search')" :placeholder="app.t('procurement.poSearchPlaceholder')" @keyup.enter="applyPOFilters" />
            <AppSelect v-model="poFilters.supplier_id" :label="app.t('procurement.supplier')">
              <option value="">{{ app.t('procurement.allSuppliers') }}</option>
              <option v-for="supplier in supplierOptions" :key="supplier.id" :value="String(supplier.id)">{{ supplier.name }}</option>
            </AppSelect>
            <AppSelect v-model="poFilters.location_id" :label="app.t('procurement.targetLocation')">
              <option value="">{{ app.t('procurement.allLocations') }}</option>
              <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
            </AppSelect>
            <AppSelect v-model="poFilters.status" :label="app.t('procurement.status')">
              <option value="">{{ app.t('procurement.allStatuses') }}</option>
              <option value="DRAFT">{{ app.t('procurement.status.DRAFT') }}</option>
              <option value="SENT">{{ app.t('procurement.status.SENT') }}</option>
              <option value="RECEIVED">{{ app.t('procurement.status.RECEIVED') }}</option>
              <option value="CANCELLED">{{ app.t('procurement.status.CANCELLED') }}</option>
            </AppSelect>
          </div>
          <div class="mt-3 grid min-w-0 max-w-full gap-3 lg:grid-cols-[minmax(0,1fr)_240px] lg:items-end">
            <AppDateRangeFilter
              v-model:date-from="poFilters.date_from"
              v-model:date-to="poFilters.date_to"
              :date-from-label="app.t('procurement.dateFrom')"
              :date-to-label="app.t('procurement.dateTo')"
              :date-placeholder="app.t('procurement.selectDate')"
              :locale="app.language === 'th' ? 'th-TH-u-ca-buddhist' : 'en-US'"
              :show-shortcuts="false"
            />
            <div class="grid grid-cols-2 gap-2">
              <AppButton class="w-full whitespace-nowrap" icon="search" @click="applyPOFilters">{{ app.t('procurement.applyFilters') }}</AppButton>
              <AppButton class="w-full whitespace-nowrap" variant="secondary" @click="resetPOFilters">{{ app.t('procurement.resetFilters') }}</AppButton>
            </div>
          </div>
        </AppCard>

        <AppCard class="min-w-0 max-w-full overflow-hidden dark:bg-slate-900/80">
          <div v-if="poError" class="mb-3 rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ poError }}</div>
          <AppLoadingState v-if="loadingPO" :label="app.t('procurement.loadingPO')" />
          <AppEmptyState v-else-if="purchaseOrders.length === 0" :title="app.t('procurement.noPO')" :description="app.t('procurement.noPODescription')" icon="clipboard-list" />
          <div v-else class="min-w-0 max-w-full">
            <div class="hidden w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain pb-2 [scrollbar-gutter:stable] md:block">
              <table class="w-full min-w-[1480px] divide-y divide-slate-200 whitespace-nowrap text-sm dark:divide-slate-800">
                <thead class="bg-slate-50 dark:bg-slate-950/70">
                  <tr>
                    <th class="px-3 py-3 text-left whitespace-nowrap">{{ app.t('procurement.poNo') }}</th>
                    <th class="px-3 py-3 text-left whitespace-nowrap">{{ app.t('procurement.supplier') }}</th>
                    <th class="px-3 py-3 text-left whitespace-nowrap">{{ app.t('procurement.status') }}</th>
                    <th class="px-3 py-3 text-right whitespace-nowrap">{{ app.t('procurement.items') }}</th>
                    <th class="px-3 py-3 text-right whitespace-nowrap">{{ app.t('procurement.total') }}</th>
                    <th class="px-3 py-3 pl-8 text-left whitespace-nowrap">{{ app.t('procurement.createdDate') }}</th>
                    <th class="px-3 py-3 text-right whitespace-nowrap">{{ app.t('procurement.actions') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr v-for="po in visiblePurchaseOrders" :key="po.id" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                    <td class="px-3 py-3 font-black whitespace-nowrap">{{ po.po_number }}</td>
                    <td class="px-3 py-3">
                      <div class="flex items-center gap-2 whitespace-nowrap">
                        <span class="font-semibold">{{ po.supplier_name }}</span>
                        <span class="text-xs text-slate-400">·</span>
                        <span class="text-xs text-slate-500 dark:text-slate-400">{{ po.location_name }}</span>
                      </div>
                    </td>
                    <td class="px-3 py-3 whitespace-nowrap"><AppBadge :tone="statusTone(po.status)">{{ statusLabel(po.status) }}</AppBadge></td>
                    <td class="px-3 py-3 text-right whitespace-nowrap">{{ po.items.length.toLocaleString(locale) }}</td>
                    <td class="px-3 py-3 text-right font-semibold whitespace-nowrap">{{ money(po.total_cost) }}</td>
                    <td class="px-3 py-3 pl-8 whitespace-nowrap">{{ formatDate(po.created_at) }}</td>
                    <td class="px-3 py-3">
                      <div class="flex flex-nowrap justify-end gap-2 whitespace-nowrap">
                        <AppButton class="!h-10 !min-h-10 whitespace-nowrap !py-0" variant="secondary" @click="showPO(po)">{{ app.t('procurement.detail') }}</AppButton>
                        <AppButton v-if="po.status === 'DRAFT' && canUpdatePO" class="!box-border !h-10 !min-h-10 !w-10 !min-w-10 !shrink-0 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('procurement.edit')" :aria-label="app.t('procurement.edit')" @click="openEditPO(po)" />
                        <AppButton v-if="po.status === 'DRAFT' && canSendPO" class="!h-10 !min-h-10 whitespace-nowrap !py-0" @click="openPOConfirm(po, 'send')">{{ app.t('procurement.sendPO') }}</AppButton>
                        <AppButton v-if="(po.status === 'DRAFT' || po.status === 'SENT') && canReceivePO" class="!h-10 !min-h-10 whitespace-nowrap !py-0" @click="openPOConfirm(po, 'receive')">{{ app.t('procurement.receivePO') }}</AppButton>
                        <AppButton v-if="(po.status === 'DRAFT' || po.status === 'SENT') && canCancelPO" class="!h-10 !min-h-10 whitespace-nowrap !py-0" variant="danger" @click="openPOConfirm(po, 'cancel')">{{ app.t('procurement.cancelPO') }}</AppButton>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="grid gap-3 md:hidden">
              <article v-for="po in visiblePurchaseOrders" :key="po.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 dark:border-slate-700 dark:bg-slate-950/60">
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <h3 class="truncate font-black">{{ po.po_number }}</h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400">{{ po.supplier_name }} · {{ po.location_name }}</p>
                  </div>
                  <AppBadge :tone="statusTone(po.status)">{{ statusLabel(po.status) }}</AppBadge>
                </div>
                <dl class="mt-3 grid grid-cols-2 gap-2 text-sm">
                  <div><dt class="text-slate-500">{{ app.t('procurement.items') }}</dt><dd class="font-bold">{{ po.items.length }}</dd></div>
                  <div><dt class="text-slate-500">{{ app.t('procurement.total') }}</dt><dd class="font-bold">{{ money(po.total_cost) }}</dd></div>
                  <div class="col-span-2"><dt class="text-slate-500">{{ app.t('procurement.createdDate') }}</dt><dd class="font-bold">{{ formatDate(po.created_at) }}</dd></div>
                </dl>
                <div class="mt-3 flex flex-wrap gap-2">
                  <AppButton variant="secondary" @click="showPO(po)">{{ app.t('procurement.detail') }}</AppButton>
                  <AppButton v-if="po.status === 'DRAFT' && canUpdatePO" class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('procurement.edit')" :aria-label="app.t('procurement.edit')" @click="openEditPO(po)" />
                  <AppButton v-if="po.status === 'DRAFT' && canSendPO" @click="openPOConfirm(po, 'send')">{{ app.t('procurement.sendPO') }}</AppButton>
                  <AppButton v-if="(po.status === 'DRAFT' || po.status === 'SENT') && canReceivePO" @click="openPOConfirm(po, 'receive')">{{ app.t('procurement.receivePO') }}</AppButton>
                  <AppButton v-if="(po.status === 'DRAFT' || po.status === 'SENT') && canCancelPO" variant="danger" @click="openPOConfirm(po, 'cancel')">{{ app.t('procurement.cancelPO') }}</AppButton>
                </div>
              </article>
            </div>

            <div class="mt-4 flex flex-col gap-3 border-t border-slate-200 pt-4 text-sm dark:border-slate-800 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-slate-500 dark:text-slate-400">{{ app.t('procurement.show') }}</span>
                <AppPageSizeSelect :model-value="poPageSize" @update:model-value="changePOPageSize" />
                <span class="text-slate-500 dark:text-slate-400">{{ app.t('procurement.perPage') }}</span>
                <span class="text-slate-500 dark:text-slate-400">{{ app.t('procurement.totalRows') }} {{ purchaseOrders.length.toLocaleString(locale) }}</span>
              </div>
              <div class="flex items-center justify-end gap-2">
                <AppButton variant="secondary" :disabled="poPage <= 1" @click="poPage -= 1">{{ app.t('procurement.previous') }}</AppButton>
                <span class="font-bold text-slate-600 dark:text-slate-300">{{ t('procurement.page', { page: poPage, total: poTotalPages }) }}</span>
                <AppButton variant="secondary" :disabled="poPage >= poTotalPages" @click="poPage += 1">{{ app.t('procurement.next') }}</AppButton>
              </div>
            </div>
          </div>
        </AppCard>
      </div>

      <div v-if="activeTab === 'suppliers'" class="grid min-w-0 max-w-full gap-4">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-xl font-black text-slate-950 dark:text-slate-50">{{ app.t('procurement.suppliers') }}</h2>
          </div>
          <div class="flex flex-wrap gap-2">
            <AppButton v-if="canCreateSupplier" icon="plus" @click="openCreateSupplier">{{ app.t('procurement.addSupplier') }}</AppButton>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-3">
          <StatCard :label="app.t('procurement.totalSuppliers')" :value="suppliers.length" :helper="app.t('procurement.totalSuppliersHelper')" icon="truck" />
          <StatCard :label="app.t('procurement.active')" :value="activeSupplierRows.length" :helper="app.t('procurement.activeSuppliersHelper')" icon="check-circle" tone="success" />
          <StatCard :label="app.t('procurement.inactive')" :value="suppliers.length - activeSupplierRows.length" :helper="app.t('procurement.inactiveSuppliersHelper')" icon="triangle-alert" tone="warning" />
        </div>

        <AppCard class="min-w-0 max-w-full dark:bg-slate-900/80">
          <div class="grid min-w-0 max-w-full gap-3 sm:grid-cols-2 xl:grid-cols-[minmax(0,1fr)_220px_240px] xl:items-end">
            <AppInput v-model="supplierFilters.search" :label="app.t('procurement.search')" :placeholder="app.t('procurement.supplierSearchPlaceholder')" @keyup.enter="applySupplierFilters" />
            <AppSelect v-model="supplierFilters.status" :label="app.t('procurement.status')">
              <option value="">{{ app.t('procurement.allStatuses') }}</option>
              <option value="active">{{ app.t('procurement.active') }}</option>
              <option value="inactive">{{ app.t('procurement.inactive') }}</option>
            </AppSelect>
            <div class="grid grid-cols-2 gap-2 sm:col-span-2 xl:col-span-1">
              <AppButton class="w-full whitespace-nowrap" icon="search" @click="applySupplierFilters">{{ app.t('procurement.applyFilters') }}</AppButton>
              <AppButton class="w-full whitespace-nowrap" variant="secondary" @click="resetSupplierFilters">{{ app.t('procurement.resetFilters') }}</AppButton>
            </div>
          </div>
        </AppCard>

        <AppCard class="min-w-0 max-w-full overflow-hidden dark:bg-slate-900/80">
          <div v-if="supplierError" class="mb-3 rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ supplierError }}</div>
          <AppLoadingState v-if="loadingSuppliers" :label="app.t('procurement.loadingSuppliers')" />
          <AppEmptyState v-else-if="suppliers.length === 0" :title="app.t('procurement.noSuppliers')" :description="app.t('procurement.noSuppliersDescription')" icon="truck" />
          <div v-else class="min-w-0 max-w-full">
            <div class="hidden w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain pb-2 [scrollbar-gutter:stable] md:block">
              <table class="w-full min-w-[980px] divide-y divide-slate-200 text-sm dark:divide-slate-800">
                <thead class="bg-slate-50 dark:bg-slate-950/70">
                  <tr>
                    <th class="px-3 py-3 text-left">{{ app.t('procurement.supplierName') }}</th>
                    <th class="px-3 py-3 text-left">{{ app.t('procurement.address') }}</th>
                    <th class="px-3 py-3 text-left">{{ app.t('procurement.phone') }}</th>
                    <th class="px-3 py-3 text-left">{{ app.t('procurement.email') }}</th>
                    <th class="px-3 py-3 text-right">{{ app.t('procurement.status') }}</th>
                    <th class="px-3 py-3 text-right">{{ app.t('procurement.actions') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr v-for="supplier in visibleSuppliers" :key="supplier.id" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                    <td class="px-3 py-3 font-semibold">{{ supplier.name }}</td>
                    <td class="px-3 py-3">{{ supplier.address || '-' }}</td>
                    <td class="px-3 py-3">{{ supplier.phone || '-' }}</td>
                    <td class="px-3 py-3">{{ supplier.email || '-' }}</td>
                    <td class="px-3 py-3 text-right"><AppBadge :tone="supplier.is_active ? 'success' : 'neutral'">{{ supplier.is_active ? app.t('procurement.active') : app.t('procurement.inactive') }}</AppBadge></td>
                    <td class="px-3 py-3">
                      <div class="flex flex-nowrap justify-end gap-2 whitespace-nowrap">
                        <AppButton v-if="canUpdateSupplier" class="!box-border !h-10 !min-h-10 !w-10 !min-w-10 !shrink-0 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('procurement.edit')" :aria-label="app.t('procurement.edit')" @click="openEditSupplier(supplier)" />
                        <AppButton v-if="canDeactivateSupplier" class="!h-10 !min-h-10 !min-w-28 whitespace-nowrap !py-0" :variant="supplier.is_active ? 'danger' : 'secondary'" @click="openSupplierConfirm(supplier)">
                          {{ supplier.is_active ? app.t('procurement.deactivate') : app.t('procurement.activate') }}
                        </AppButton>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="grid gap-3 md:hidden">
              <article v-for="supplier in visibleSuppliers" :key="supplier.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 dark:border-slate-700 dark:bg-slate-950/60">
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <h3 class="truncate font-black">{{ supplier.name }}</h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400">{{ supplier.phone || '-' }} · {{ supplier.email || '-' }}</p>
                  </div>
                  <AppBadge :tone="supplier.is_active ? 'success' : 'neutral'">{{ supplier.is_active ? app.t('procurement.active') : app.t('procurement.inactive') }}</AppBadge>
                </div>
                <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">{{ supplier.address || app.t('procurement.noAddress') }}</p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <AppButton v-if="canUpdateSupplier" class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('procurement.edit')" :aria-label="app.t('procurement.edit')" @click="openEditSupplier(supplier)" />
                  <AppButton v-if="canDeactivateSupplier" :variant="supplier.is_active ? 'danger' : 'secondary'" @click="openSupplierConfirm(supplier)">
                    {{ supplier.is_active ? app.t('procurement.deactivate') : app.t('procurement.activate') }}
                  </AppButton>
                </div>
              </article>
            </div>

            <div class="mt-4 flex flex-col gap-3 border-t border-slate-200 pt-4 text-sm dark:border-slate-800 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-slate-500 dark:text-slate-400">{{ app.t('procurement.show') }}</span>
                <AppPageSizeSelect :model-value="supplierPageSize" @update:model-value="changeSupplierPageSize" />
                <span class="text-slate-500 dark:text-slate-400">{{ app.t('procurement.perPage') }}</span>
                <span class="text-slate-500 dark:text-slate-400">{{ app.t('procurement.totalRows') }} {{ suppliers.length.toLocaleString(locale) }}</span>
              </div>
              <div class="flex items-center justify-end gap-2">
                <AppButton variant="secondary" :disabled="supplierPage <= 1" @click="supplierPage -= 1">{{ app.t('procurement.previous') }}</AppButton>
                <span class="font-bold text-slate-600 dark:text-slate-300">{{ t('procurement.page', { page: supplierPage, total: supplierTotalPages }) }}</span>
                <AppButton variant="secondary" :disabled="supplierPage >= supplierTotalPages" @click="supplierPage += 1">{{ app.t('procurement.next') }}</AppButton>
              </div>
            </div>
          </div>
        </AppCard>
      </div>
    </div>

    <AppModal :open="poModalOpen" :title="editingPOID ? app.t('procurement.editPO') : app.t('procurement.createPO')" :description="app.t('procurement.poModalDescription')" :close-label="app.t('procurement.cancel')" size="xl" @close="closePOModal">
      <form class="grid gap-4" @submit.prevent="requestSavePO">
        <div class="grid gap-3">
          <AppSelect v-model="poForm.supplier_id" :label="app.t('procurement.supplier')">
            <option value="">{{ app.t('procurement.selectSupplier') }}</option>
            <option v-for="supplier in activeSuppliers" :key="supplier.id" :value="String(supplier.id)">{{ supplier.name }}</option>
          </AppSelect>
          <AppSelect v-model="poForm.location_id" :label="app.t('procurement.targetLocation')">
            <option value="">{{ app.t('procurement.selectLocation') }}</option>
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <AppTextarea v-model="poForm.note" :label="app.t('procurement.note')" :placeholder="app.t('procurement.notePlaceholder')" />
        </div>

        <section class="grid max-h-[48vh] gap-3 overflow-y-auto rounded-2xl bg-slate-50/80 p-4 pr-2 dark:bg-slate-950/45">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <h3 class="text-sm font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('procurement.items') }}</h3>
            <AppButton type="button" variant="secondary" icon="plus" @click="addPOItem">{{ app.t('procurement.addItem') }}</AppButton>
          </div>
          <article v-for="(item, index) in poForm.items" :key="index" class="rounded-xl border border-slate-200 bg-white/70 p-3 dark:border-slate-700 dark:bg-slate-900/60">
            <div class="grid gap-3">
              <AppSelect v-model="item.product_id" :label="app.t('procurement.product')" @update:model-value="syncProductCost(item)">
                <option v-for="product in products" :key="product.id" :value="product.id" :disabled="isProductSelected(product.id, index)">{{ product.name }} · {{ product.sku }}</option>
              </AppSelect>
              <div class="grid gap-3 sm:grid-cols-[1fr_1fr_auto] sm:items-end">
                <AppInput v-model="item.quantity" :label="app.t('procurement.quantity')" type="number" :placeholder="app.t('procurement.quantityPlaceholder')" />
                <AppInput v-model="item.unit_cost" :label="app.t('procurement.unitCost')" type="number" :placeholder="app.t('procurement.unitCostPlaceholder')" />
                <AppButton type="button" variant="danger" @click="removePOItem(index)">{{ app.t('procurement.remove') }}</AppButton>
              </div>
            </div>
            <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">{{ app.t('procurement.lineCost') }}: {{ money(Number(item.quantity || 0) * Number(item.unit_cost || 0)) }}</p>
          </article>
          <div class="sticky bottom-0 rounded-xl border border-slate-200 bg-white/95 p-3 text-sm font-bold shadow-sm backdrop-blur dark:border-slate-700 dark:bg-slate-900/95">{{ app.t('procurement.total') }}: {{ money(poTotalCost) }}</div>
        </section>

        <div v-if="poError" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ poError }}</div>
        <div class="sticky -bottom-6 -mx-5 -mb-5 flex flex-col-reverse gap-2 border-t border-slate-200 bg-white/90 p-4 backdrop-blur sm:-mx-6 sm:-mb-6 sm:flex-row sm:justify-end dark:border-slate-800 dark:bg-slate-900/90">
          <AppButton class="w-full sm:w-auto" type="button" variant="secondary" :disabled="savingPO" @click="closePOModal">{{ app.t('procurement.cancel') }}</AppButton>
          <AppButton class="w-full sm:w-auto" type="submit" :loading="savingPO" :disabled="savingPO">{{ editingPOID ? app.t('procurement.save') : app.t('procurement.createDraft') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <AppModal :open="poDetailOpen" :title="selectedPO ? t('procurement.detailTitle', { number: selectedPO.po_number }) : app.t('procurement.detail')" :close-label="app.t('procurement.cancel')" size="xl" @close="closePODetail">
      <div v-if="selectedPO" class="grid gap-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="font-bold">{{ selectedPO.supplier_name }}</p>
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ selectedPO.location_name }} · {{ formatDate(selectedPO.created_at) }}</p>
            <p v-if="selectedPO.note" class="mt-2 text-sm text-slate-600 dark:text-slate-300">{{ selectedPO.note }}</p>
          </div>
          <AppBadge :tone="statusTone(selectedPO.status)">{{ statusLabel(selectedPO.status) }}</AppBadge>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
            <thead class="bg-slate-50 dark:bg-slate-950/70">
              <tr>
                <th class="px-3 py-2 text-left">{{ app.t('procurement.product') }}</th>
                <th class="px-3 py-2 text-right">{{ app.t('procurement.quantity') }}</th>
                <th class="px-3 py-2 text-right">{{ app.t('procurement.receivedQuantity') }}</th>
                <th class="px-3 py-2 text-right">{{ app.t('procurement.unitCost') }}</th>
                <th class="px-3 py-2 text-right">{{ app.t('procurement.lineCost') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="item in selectedPO.items" :key="item.id">
                <td class="px-3 py-2">
                  <div class="flex min-w-0 items-center gap-3">
                    <ProductAvatar :src="item.image_url" :updated-at="item.image_updated_at" :name="item.product_name" size="sm" shape="square" />
                    <div class="min-w-0">
                      <p class="truncate font-semibold">{{ item.product_name }}</p>
                      <p class="text-xs text-slate-500 dark:text-slate-400">{{ item.sku }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-3 py-2 text-right">{{ item.quantity }}</td>
                <td class="px-3 py-2 text-right">{{ item.received_quantity }}</td>
                <td class="px-3 py-2 text-right">{{ money(item.unit_cost) }}</td>
                <td class="px-3 py-2 text-right font-semibold">{{ money(item.line_cost) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="rounded-xl bg-slate-50 p-3 text-right text-sm font-black dark:bg-slate-950/60">{{ app.t('procurement.total') }}: {{ money(selectedPO.total_cost) }}</div>
      </div>
    </AppModal>

    <AppModal :open="supplierModalOpen" :title="editingSupplierID ? app.t('procurement.editSupplier') : app.t('procurement.createSupplier')" :description="app.t('procurement.supplierModalDescription')" :close-label="app.t('procurement.cancel')" size="lg" @close="closeSupplierModal">
      <form class="grid gap-4" @submit.prevent="requestSaveSupplier">
        <div class="grid gap-3">
          <AppInput v-model="supplierForm.name" :label="app.t('procurement.supplierName')" :placeholder="app.t('procurement.supplierNamePlaceholder')" />
          <AppInput v-model="supplierForm.phone" :label="app.t('procurement.phone')" :placeholder="app.t('procurement.phonePlaceholder')" />
          <AppInput v-model="supplierForm.email" :label="app.t('procurement.email')" :placeholder="app.t('procurement.emailPlaceholder')" />
          <AppTextarea v-model="supplierForm.address" :label="app.t('procurement.address')" :placeholder="app.t('procurement.addressPlaceholder')" />
        </div>
        <div v-if="supplierError" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ supplierError }}</div>
        <div class="sticky -bottom-6 -mx-5 -mb-5 flex flex-col-reverse gap-2 border-t border-slate-200 bg-white/90 p-4 backdrop-blur sm:-mx-6 sm:-mb-6 sm:flex-row sm:justify-end dark:border-slate-800 dark:bg-slate-900/90">
          <AppButton class="w-full sm:w-auto" type="button" variant="secondary" :disabled="savingSupplier" @click="closeSupplierModal">{{ app.t('procurement.cancel') }}</AppButton>
          <AppButton class="w-full sm:w-auto" type="submit" :loading="savingSupplier" :disabled="savingSupplier">{{ app.t('procurement.save') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="poSaveConfirmOpen"
      :title="poSaveConfirmTitle"
      :message="poSaveConfirmMessage"
      :confirm-label="editingPOID ? app.t('procurement.save') : app.t('procurement.createDraft')"
      :cancel-label="app.t('procurement.cancel')"
      :loading="savingPO"
      @close="closePOSaveConfirm"
      @confirm="savePO"
    />

    <ConfirmDialog
      :open="supplierSaveConfirmOpen"
      :title="supplierSaveConfirmTitle"
      :message="supplierSaveConfirmMessage"
      :confirm-label="app.t('procurement.save')"
      :cancel-label="app.t('procurement.cancel')"
      :loading="savingSupplier"
      @close="closeSupplierSaveConfirm"
      @confirm="saveSupplier"
    />

    <ConfirmDialog
      :open="confirmState.type === 'po-action'"
      :title="confirmState.action ? app.t(`procurement.confirm.${confirmState.action}` as TranslationKey) : app.t('procurement.confirmAction')"
      :message="confirmState.po ? t('procurement.confirmPOMessage', { number: confirmState.po.po_number }) : ''"
      :confirm-label="confirmState.action ? app.t(`procurement.${confirmState.action}PO` as TranslationKey) : app.t('procurement.confirm')"
      :cancel-label="app.t('procurement.cancel')"
      :destructive="confirmState.action === 'cancel'"
      :loading="confirming"
      @close="closeConfirm"
      @confirm="confirmPOAction"
    />

    <ConfirmDialog
      :open="confirmState.type === 'supplier-status'"
      :title="confirmState.supplier?.is_active ? app.t('procurement.confirmDeactivateSupplier') : app.t('procurement.confirmActivateSupplier')"
      :message="confirmState.supplier ? t('procurement.confirmSupplierMessage', { name: confirmState.supplier.name }) : ''"
      :confirm-label="confirmState.supplier?.is_active ? app.t('procurement.deactivate') : app.t('procurement.activate')"
      :cancel-label="app.t('procurement.cancel')"
      :destructive="Boolean(confirmState.supplier?.is_active)"
      :loading="confirming"
      @close="closeConfirm"
      @confirm="confirmSupplierStatus"
    />
  </section>
</template>
