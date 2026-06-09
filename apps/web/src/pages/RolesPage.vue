<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppModal from '../components/AppModal.vue'
import AppPageSizeSelect from '../components/AppPageSizeSelect.vue'
import AppTextarea from '../components/AppTextarea.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { PermissionRecord, RoleRecord } from '../types/navigation'

interface RoleForm {
  id: number
  code: string
  name: string
  description: string
  is_active: boolean
}

const moduleLabelKeys: Record<string, TranslationKey> = {
  dashboard: 'roles.module.dashboard',
  pos: 'roles.module.pos',
  products: 'roles.module.products',
  categories: 'roles.module.categories',
  stock: 'roles.module.stock',
  locations: 'roles.module.locations',
  transfers: 'roles.module.transfers',
  sales: 'roles.module.sales',
  alerts: 'roles.module.alerts',
  reports: 'roles.module.reports',
  exports: 'roles.module.exports',
  imports: 'roles.module.imports',
  suppliers: 'roles.module.suppliers',
  purchase_orders: 'roles.module.purchaseOrders',
  users: 'roles.module.users',
  roles: 'roles.module.roles',
  permissions: 'roles.module.permissions',
  settings: 'roles.module.settings',
  notifications: 'roles.module.notifications',
}

const actionLabelKeys: Record<string, TranslationKey> = {
  view: 'roles.action.view',
  create: 'roles.action.create',
  update: 'roles.action.update',
  deactivate: 'roles.action.deactivate',
  sell: 'roles.action.sell',
  clear_cart: 'roles.action.clearCart',
  apply_discount: 'roles.action.applyDiscount',
  import: 'roles.action.import',
  export: 'roles.action.export',
  restock: 'roles.action.restock',
  adjust: 'roles.action.adjust',
  'movements.view': 'roles.action.viewMovements',
  complete: 'roles.action.complete',
  cancel: 'roles.action.cancel',
  'receipt.view': 'roles.action.viewReceipts',
  mark_read: 'roles.action.markRead',
  create_po: 'roles.action.createPO',
  daily_sales: 'roles.action.dailySales',
  monthly_sales: 'roles.action.monthlySales',
  best_selling: 'roles.action.bestSelling',
  profit: 'roles.action.profit',
  stock: 'roles.action.stockReport',
  inventory_valuation: 'roles.action.inventoryValuation',
  payment_summary: 'roles.action.paymentSummary',
  low_stock: 'roles.action.lowStock',
  reorder: 'roles.action.reorder',
  'template.download': 'roles.action.downloadTemplate',
  'products.preview': 'roles.action.previewProducts',
  'products.confirm': 'roles.action.confirmProducts',
  'history.view': 'roles.action.viewHistory',
  send: 'roles.action.send',
  receive: 'roles.action.receive',
  create_from_alert: 'roles.action.createFromAlert',
  assign_roles: 'roles.action.assignRoles',
  assign_permissions: 'roles.action.assignPermissions',
  'line.view': 'roles.action.viewLine',
  'line.update': 'roles.action.updateLine',
  'line.test': 'roles.action.testLine',
}

const app = useAppStore()
const auth = useAuthStore()
const roles = ref<RoleRecord[]>([])
const permissions = ref<PermissionRecord[]>([])
const formPermissionCodes = ref<string[]>([])
const roleSearch = ref('')
const permissionSearch = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const roleModalOpen = ref(false)
const saveConfirmOpen = ref(false)
const confirmOpen = ref(false)
const targetRole = ref<RoleRecord | null>(null)
const page = ref(1)
const pageSize = ref(10)
const form = reactive<RoleForm>({ id: 0, code: '', name: '', description: '', is_active: true })

const canCreate = computed(() => auth.hasPermission('roles.create'))
const canUpdate = computed(() => auth.hasPermission('roles.update'))
const canDeactivate = computed(() => auth.hasPermission('roles.deactivate'))
const canAssign = computed(() => auth.hasPermission('roles.assign_permissions'))
const editing = computed(() => form.id > 0)
const activeRoles = computed(() => roles.value.filter((role) => role.is_active).length)
const inactiveRoles = computed(() => roles.value.length - activeRoles.value)
const totalPermissions = computed(() => permissions.value.length)
const allPermissionCodes = computed(() => permissions.value.map((permission) => permission.code))
const saveConfirmTitle = computed(() => editing.value ? app.t('roles.confirmUpdateTitle') : app.t('roles.confirmCreateTitle'))
const saveConfirmMessage = computed(() => t('roles.confirmSaveMessage', { name: form.name.trim() || form.code.trim() || app.t('roles.role') }))
const editingRole = computed(() => roles.value.find((role) => role.id === form.id))

const filteredRoles = computed(() => {
  const query = roleSearch.value.trim().toLowerCase()
  if (!query) return roles.value
  return roles.value.filter((role) => `${role.code} ${role.name} ${role.description}`.toLowerCase().includes(query))
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredRoles.value.length / pageSize.value)))
const visibleRoles = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filteredRoles.value.slice(start, start + pageSize.value)
})

const filteredPermissions = computed(() => {
  const query = permissionSearch.value.trim().toLowerCase()
  if (!query) return permissions.value
  return permissions.value.filter((permission) => `${permission.code} ${permission.name} ${permission.module} ${permission.action}`.toLowerCase().includes(query))
})

const groupedPermissions = computed(() => {
  const groups: Record<string, PermissionRecord[]> = {}
  for (const permission of filteredPermissions.value) {
    if (!groups[permission.module]) groups[permission.module] = []
    groups[permission.module].push(permission)
  }
  return Object.entries(groups)
    .sort(([a], [b]) => moduleLabel(a).localeCompare(moduleLabel(b)))
    .map(([module, items]) => ({ module, items }))
})

function moduleLabel(module: string) {
  const key = moduleLabelKeys[module]
  return key ? app.t(key) : module.replaceAll('_', ' ')
}

function actionLabel(action: string) {
  const key = actionLabelKeys[action]
  return key ? app.t(key) : action.replaceAll('_', ' ')
}

function permissionTitle(permission: PermissionRecord) {
  return `${actionLabel(permission.action)} - ${moduleLabel(permission.module)}`
}

function hasPermission(code: string) {
  return formPermissionCodes.value.includes(code)
}

function roleStatusTone(role: RoleRecord) {
  if (!role.is_active) return 'neutral'
  return role.is_system ? 'info' : 'success'
}

function resetForm() {
  Object.assign(form, { id: 0, code: '', name: '', description: '', is_active: true })
  formPermissionCodes.value = []
  permissionSearch.value = ''
  error.value = ''
}

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [roleRows, permissionRows] = await Promise.all([
      apiClient<RoleRecord[]>('/v1/roles'),
      apiClient<PermissionRecord[]>('/v1/permissions'),
    ])
    roles.value = roleRows
    permissions.value = permissionRows
    page.value = Math.min(page.value, totalPages.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('roles.loadFailed')
  } finally {
    loading.value = false
  }
}

function openCreateRole() {
  resetForm()
  roleModalOpen.value = true
}

function closeRoleModal(force = false) {
  if (saving.value && !force) return
  roleModalOpen.value = false
  saveConfirmOpen.value = false
  resetForm()
}

async function openEditRole(role: RoleRecord) {
  Object.assign(form, {
    id: role.id,
    code: role.code,
    name: role.name,
    description: role.description,
    is_active: role.is_active,
  })
  permissionSearch.value = ''
  error.value = ''
  roleModalOpen.value = true
  try {
    const codes = await apiClient<string[]>(`/v1/roles/${role.id}/permissions`)
    formPermissionCodes.value = withPermissionDependencies(codes)
  } catch (err) {
    app.pushToast({ type: 'error', message: app.t('roles.permissionsLoadFailed'), description: err instanceof Error ? err.message : '' })
  }
}

function togglePermission(code: string, checked: boolean) {
  const next = new Set(formPermissionCodes.value)
  if (checked) next.add(code)
  else next.delete(code)
  formPermissionCodes.value = withPermissionDependencies([...next])
}

function selectPermissions(items: PermissionRecord[]) {
  const next = new Set(formPermissionCodes.value)
  items.forEach((item) => next.add(item.code))
  formPermissionCodes.value = withPermissionDependencies([...next])
}

function clearPermissions(items: PermissionRecord[]) {
  const remove = new Set(items.map((item) => item.code))
  formPermissionCodes.value = withPermissionDependencies(formPermissionCodes.value.filter((code) => !remove.has(code)))
}

function selectAllPermissions() {
  formPermissionCodes.value = withPermissionDependencies(allPermissionCodes.value)
}

function clearAllPermissions() {
  formPermissionCodes.value = []
}

function withPermissionDependencies(codes: string[]) {
  const selected = new Set(codes)
  for (const permission of permissions.value) {
    if (!selected.has(permission.code) || !permissionRequiresRead(permission.action)) continue
    const parts = permission.action.split('.')
    const prefix = parts.length > 1 ? parts.slice(0, -1).join('.') : ''
    for (const candidate of permissions.value) {
      if (candidate.module !== permission.module) continue
      if (candidate.action === 'view' || candidate.action === 'read' || (prefix && (candidate.action === `${prefix}.view` || candidate.action === `${prefix}.read`))) {
        selected.add(candidate.code)
      }
    }
  }
  return [...selected].sort()
}

function permissionRequiresRead(action: string) {
  return ['create', 'update', 'delete', 'deactivate'].includes(action.split('.').at(-1) ?? '')
}

function validateRole() {
  if (!form.code.trim()) return app.t('roles.codeRequired')
  if (!form.name.trim()) return app.t('roles.nameRequired')
  return ''
}

function requestSaveRole() {
  const validation = validateRole()
  if (validation) {
    error.value = validation
    return
  }
  error.value = ''
  saveConfirmOpen.value = true
}

function closeSaveConfirm() {
  if (saving.value) return
  saveConfirmOpen.value = false
}

async function saveRole() {
  saving.value = true
  error.value = ''
  const toastMessage = editing.value ? app.t('roles.updated') : app.t('roles.created')
  try {
    const payload = {
      code: form.code.trim(),
      name: form.name.trim(),
      description: form.description.trim(),
      is_active: form.is_active,
    }
    const role = editing.value
      ? await patchJSON<RoleRecord>(`/v1/roles/${form.id}`, payload)
      : await postJSON<RoleRecord>('/v1/roles', payload)
    if (canAssign.value) {
      await apiClient<string[]>(`/v1/roles/${role.id}/permissions`, {
        method: 'PUT',
        body: JSON.stringify({ permission_codes: formPermissionCodes.value }),
      })
    }
    saveConfirmOpen.value = false
    closeRoleModal(true)
    app.pushToast({ type: 'success', message: toastMessage })
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : app.t('roles.saveFailed')
    app.pushToast({ type: 'error', message: app.t('roles.saveFailed'), description: error.value })
  } finally {
    saving.value = false
  }
}

function askDeactivate(role: RoleRecord) {
  if (role.is_active && role.user_count > 0) {
    app.pushToast({ type: 'error', message: app.t('roles.assignedRoleCannotDeactivate') })
    return
  }
  targetRole.value = role
  confirmOpen.value = true
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
}

function previousPage() {
  if (page.value > 1) page.value -= 1
}

function nextPage() {
  if (page.value < totalPages.value) page.value += 1
}

async function deactivateRole() {
  if (!targetRole.value) return
  saving.value = true
  try {
    await patchJSON<RoleRecord>(`/v1/roles/${targetRole.value.id}/status`, { is_active: !targetRole.value.is_active })
    app.pushToast({ type: 'success', message: targetRole.value.is_active ? app.t('roles.deactivated') : app.t('roles.activated') })
    confirmOpen.value = false
    targetRole.value = null
    await load()
  } catch (err) {
    app.pushToast({ type: 'error', message: app.t('roles.statusFailed'), description: err instanceof Error ? err.message : '' })
  } finally {
    saving.value = false
  }
}

watch(roleSearch, () => {
  page.value = 1
})

onMounted(load)
</script>

<template>
  <section>
    <PageHeader :title="app.t('roles.title')" :eyebrow="app.t('roles.eyebrow')" :description="app.t('roles.description')" icon="role">
      <div class="flex flex-wrap gap-2">
        <AppButton v-if="canCreate" icon="plus" @click="openCreateRole">{{ app.t('roles.add') }}</AppButton>
      </div>
    </PageHeader>

    <div class="mb-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard :label="app.t('roles.totalRoles')" :value="roles.length" :helper="app.t('roles.totalRolesHelper')" icon="role" />
      <StatCard :label="app.t('roles.activeRoles')" :value="activeRoles" :helper="app.t('roles.activeRolesHelper')" icon="check-circle" tone="success" />
      <StatCard :label="app.t('roles.inactiveRoles')" :value="inactiveRoles" :helper="app.t('roles.inactiveRolesHelper')" icon="triangle-alert" tone="warning" />
      <StatCard :label="app.t('roles.totalPermissions')" :value="totalPermissions" :helper="app.t('roles.totalPermissionsHelper')" icon="sparkles" tone="info" />
    </div>

    <div v-if="error && !roleModalOpen" class="mb-4 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
    <AppLoadingState v-if="loading" class="mb-4" :label="app.t('roles.loading')" />

    <AppCard class="dark:bg-slate-900/80">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="font-black">{{ app.t('roles.roleTable') }}</h2>
          <p class="text-sm text-slate-500 dark:text-slate-400">{{ app.t('roles.roleTableDescription') }}</p>
        </div>
        <AppInput v-model="roleSearch" class="w-full sm:w-80" :placeholder="app.t('roles.searchRolesPlaceholder')" />
      </div>

      <AppEmptyState v-if="!loading && filteredRoles.length === 0" class="mt-5" :title="app.t('roles.empty')" :description="app.t('roles.emptyDescription')" icon="role" />

      <div v-else class="mt-5 hidden w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain lg:block">
        <table class="w-full min-w-[980px] divide-y divide-slate-200 text-sm dark:divide-slate-800">
          <thead class="bg-slate-50 dark:bg-slate-950/70">
            <tr>
              <th class="px-3 py-3 text-left">{{ app.t('roles.name') }}</th>
              <th class="px-3 py-3 text-left">{{ app.t('roles.code') }}</th>
              <th class="px-3 py-3 text-right">{{ app.t('roles.permissions') }}</th>
              <th class="px-3 py-3 text-right">{{ app.t('roles.assignedUsers') }}</th>
              <th class="px-3 py-3 pl-8 text-left">{{ app.t('roles.status') }}</th>
              <th class="px-3 py-3 text-right">{{ app.t('roles.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
            <tr v-for="role in visibleRoles" :key="role.id" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
              <td class="px-3 py-3 pl-8">
                <div class="font-black">{{ role.name }}</div>
                <div class="mt-1 max-w-md truncate text-xs text-slate-500 dark:text-slate-400">{{ role.description || app.t('roles.noDescription') }}</div>
              </td>
              <td class="px-3 py-3"><AppBadge tone="neutral">{{ role.code }}</AppBadge></td>
              <td class="px-3 py-3 text-right font-black">{{ role.permission_count }}</td>
              <td class="px-3 py-3 text-right font-black">{{ role.user_count }}</td>
              <td class="px-3 py-3 pl-8">
                <div class="flex flex-wrap gap-1">
                  <AppBadge v-if="role.is_system" tone="info">{{ app.t('roles.systemRole') }}</AppBadge>
                  <AppBadge :tone="roleStatusTone(role)">{{ role.is_active ? app.t('roles.active') : app.t('roles.inactive') }}</AppBadge>
                </div>
              </td>
              <td class="px-3 py-3">
                <div class="flex justify-end gap-2">
                  <AppButton v-if="canUpdate" class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('roles.edit')" :aria-label="app.t('roles.edit')" @click="openEditRole(role)" />
                  <AppButton v-if="canDeactivate" :variant="role.is_active ? 'danger' : 'secondary'" icon="triangle-alert" :disabled="role.is_active && role.user_count > 0" :title="role.is_active && role.user_count > 0 ? app.t('roles.assignedRoleCannotDeactivate') : undefined" @click="askDeactivate(role)">
                    {{ role.is_active ? app.t('roles.deactivate') : app.t('roles.activate') }}
                  </AppButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mt-5 grid gap-3 lg:hidden">
        <article v-for="role in visibleRoles" :key="role.id" class="rounded-2xl border border-slate-200 bg-white/70 p-4 dark:border-slate-700 dark:bg-slate-950/60">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <h3 class="truncate font-black">{{ role.name }}</h3>
              <p class="text-xs font-bold uppercase text-slate-500 dark:text-slate-400">{{ role.code }}</p>
              <p class="mt-2 line-clamp-2 text-sm text-slate-500 dark:text-slate-400">{{ role.description || app.t('roles.noDescription') }}</p>
            </div>
            <AppBadge :tone="roleStatusTone(role)">{{ role.is_active ? app.t('roles.active') : app.t('roles.inactive') }}</AppBadge>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-2 text-sm">
            <div class="rounded-xl bg-slate-50 p-3 dark:bg-slate-900/70">
              <p class="text-slate-500 dark:text-slate-400">{{ app.t('roles.permissions') }}</p>
              <p class="text-lg font-black">{{ role.permission_count }}</p>
            </div>
            <div class="rounded-xl bg-slate-50 p-3 dark:bg-slate-900/70">
              <p class="text-slate-500 dark:text-slate-400">{{ app.t('roles.assignedUsers') }}</p>
              <p class="text-lg font-black">{{ role.user_count }}</p>
            </div>
          </div>
          <div class="mt-3 flex flex-wrap gap-2">
            <AppBadge v-if="role.is_system" tone="info">{{ app.t('roles.systemRole') }}</AppBadge>
            <AppButton v-if="canUpdate" class="!h-10 !min-h-10 !w-10 !px-0 !py-0" variant="secondary" icon="settings" :title="app.t('roles.edit')" :aria-label="app.t('roles.edit')" @click="openEditRole(role)" />
            <AppButton v-if="canDeactivate" :variant="role.is_active ? 'danger' : 'secondary'" icon="triangle-alert" :disabled="role.is_active && role.user_count > 0" :title="role.is_active && role.user_count > 0 ? app.t('roles.assignedRoleCannotDeactivate') : undefined" @click="askDeactivate(role)">
              {{ role.is_active ? app.t('roles.deactivate') : app.t('roles.activate') }}
            </AppButton>
          </div>
        </article>
      </div>

      <div v-if="filteredRoles.length > 0" class="mt-4 flex flex-col gap-3 border-t border-slate-200 pt-4 text-sm dark:border-slate-800 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex flex-wrap items-center gap-2">
          <span class="text-slate-500 dark:text-slate-400">{{ app.t('roles.show') }}</span>
          <AppPageSizeSelect :model-value="pageSize" @update:model-value="changePageSize" />
          <span class="text-slate-500 dark:text-slate-400">{{ app.t('roles.perPage') }}</span>
          <span class="text-slate-500 dark:text-slate-400">{{ app.t('roles.totalRows') }} {{ filteredRoles.length }}</span>
        </div>
        <div class="flex items-center justify-end gap-2">
          <AppButton variant="secondary" :disabled="page <= 1" @click="previousPage">{{ app.t('roles.previous') }}</AppButton>
          <span class="font-bold text-slate-600 dark:text-slate-300">{{ t('roles.page', { page, total: totalPages }) }}</span>
          <AppButton variant="secondary" :disabled="page >= totalPages" @click="nextPage">{{ app.t('roles.next') }}</AppButton>
        </div>
      </div>
    </AppCard>

    <AppModal :open="roleModalOpen" :title="editing ? app.t('roles.editTitle') : app.t('roles.createTitle')" :description="app.t('roles.roleModalDescription')" :close-label="app.t('roles.cancel')" size="xl" @close="closeRoleModal">
      <form class="grid gap-5" @submit.prevent="requestSaveRole">
        <section class="grid gap-3 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/45">
          <div class="grid gap-3 md:grid-cols-2">
            <AppInput v-model="form.code" :label="app.t('roles.code')" :placeholder="app.t('roles.codePlaceholder')" :disabled="editing" :helper="app.t('roles.codeHelper')" />
            <AppInput v-model="form.name" :label="app.t('roles.name')" :placeholder="app.t('roles.namePlaceholder')" :helper="app.t('roles.nameHelper')" />
          </div>
          <AppTextarea v-model="form.description" :label="app.t('roles.descriptionField')" :placeholder="app.t('roles.descriptionPlaceholder')" />
          <label class="flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-200">
            <input v-model="form.is_active" type="checkbox" :disabled="Boolean(editingRole?.is_active && editingRole.user_count > 0)" />
            {{ app.t('roles.active') }}
          </label>
          <p v-if="editingRole?.is_active && editingRole.user_count > 0" class="text-xs font-semibold text-amber-700 dark:text-amber-200">{{ app.t('roles.assignedRoleCannotDeactivate') }}</p>
        </section>

        <section class="grid gap-4">
          <div class="flex flex-wrap items-end justify-between gap-3">
            <div>
              <h3 class="font-black">{{ app.t('roles.permissions') }}</h3>
              <p class="text-sm text-slate-500 dark:text-slate-400">{{ app.t('roles.permissionModalHint') }}</p>
              <p class="mt-1 text-xs font-semibold text-brand-700 dark:text-emerald-300">{{ app.t('roles.permissionDependencyHint') }}</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <AppButton type="button" variant="secondary" :disabled="!canAssign" @click="selectAllPermissions">{{ app.t('roles.selectAllPermissions') }}</AppButton>
              <AppButton type="button" variant="ghost" :disabled="!canAssign" @click="clearAllPermissions">{{ app.t('roles.clearAllPermissions') }}</AppButton>
            </div>
          </div>

          <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
            <AppInput v-model="permissionSearch" :label="app.t('roles.searchPermissions')" :placeholder="app.t('roles.searchPermissionsPlaceholder')" />
            <AppBadge tone="info">{{ formPermissionCodes.length }} / {{ totalPermissions }}</AppBadge>
          </div>

          <div v-if="form.code === 'ADMIN' || (editing && form.code === 'ADMIN')" class="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm font-semibold text-amber-800 dark:border-amber-400/30 dark:bg-amber-500/10 dark:text-amber-100">
            {{ app.t('roles.systemWarning') }}
          </div>

          <AppEmptyState v-if="filteredPermissions.length === 0" :title="app.t('roles.noPermissions')" :description="app.t('roles.noPermissionsDescription')" icon="search" />
          <div v-else class="max-h-[430px] overflow-auto rounded-2xl border border-slate-200 bg-white/70 p-3 dark:border-slate-700 dark:bg-slate-950/40">
            <section v-for="group in groupedPermissions" :key="group.module" class="border-b border-slate-100 py-3 last:border-b-0 dark:border-slate-800">
              <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
                <div>
                  <h4 class="font-black">{{ moduleLabel(group.module) }}</h4>
                  <p class="text-xs text-slate-500 dark:text-slate-400">{{ group.items.length }} {{ app.t('roles.permissions') }}</p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <AppButton type="button" variant="secondary" :disabled="!canAssign" @click="selectPermissions(group.items)">{{ app.t('roles.selectAll') }}</AppButton>
                  <AppButton type="button" variant="ghost" :disabled="!canAssign" @click="clearPermissions(group.items)">{{ app.t('roles.clearAll') }}</AppButton>
                </div>
              </div>
              <div class="grid gap-2 md:grid-cols-2">
                <label v-for="permission in group.items" :key="permission.id" class="flex items-start gap-3 rounded-xl border p-3 text-sm transition" :class="hasPermission(permission.code) ? 'border-brand-300 bg-brand-50/70 dark:border-emerald-400/50 dark:bg-emerald-500/10' : 'border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-900/70'">
                  <input
                    class="mt-1 h-4 w-4 rounded border-slate-300 text-brand-600"
                    type="checkbox"
                    :checked="hasPermission(permission.code)"
                    :disabled="!canAssign"
                    @change="togglePermission(permission.code, ($event.target as HTMLInputElement).checked)"
                  />
                  <span class="min-w-0">
                    <span class="block font-bold">{{ permissionTitle(permission) }}</span>
                    <span class="mt-1 block break-words text-xs text-slate-500 dark:text-slate-400">{{ permission.code }}</span>
                  </span>
                </label>
              </div>
            </section>
          </div>
        </section>

        <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
        <div class="sticky -bottom-6 z-30 -mx-5 -mb-5 flex flex-col-reverse gap-2 border-t border-slate-200 bg-white px-5 pb-6 pt-4 dark:border-slate-700 dark:bg-slate-900 dark:shadow-[0_-18px_30px_rgba(0,0,0,0.35)] sm:-mx-6 sm:-mb-6 sm:flex-row sm:justify-end sm:px-6">
          <AppButton type="button" variant="secondary" :disabled="saving" @click="closeRoleModal()">{{ app.t('roles.cancel') }}</AppButton>
          <AppButton type="submit" :loading="saving" :disabled="saving || (editing ? !canUpdate : !canCreate)" icon="check-circle">{{ app.t('roles.save') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="saveConfirmOpen"
      :title="saveConfirmTitle"
      :message="saveConfirmMessage"
      :confirm-label="app.t('roles.save')"
      :cancel-label="app.t('roles.cancel')"
      :loading="saving"
      @close="closeSaveConfirm"
      @confirm="saveRole"
    />

    <ConfirmDialog
      :open="confirmOpen"
      :title="targetRole?.is_active ? app.t('roles.confirmDeactivate') : app.t('roles.confirmActivate')"
      :message="targetRole ? `${app.t('roles.role')}: ${targetRole.name}` : ''"
      :destructive="targetRole?.is_active"
      :loading="saving"
      :confirm-label="app.t('roles.confirm')"
      :cancel-label="app.t('roles.cancel')"
      @close="confirmOpen = false"
      @confirm="deactivateRole"
    />
  </section>
</template>
