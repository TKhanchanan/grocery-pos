<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppModal from '../components/AppModal.vue'
import AppTextarea from '../components/AppTextarea.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
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

const app = useAppStore()
const auth = useAuthStore()
const roles = ref<RoleRecord[]>([])
const permissions = ref<PermissionRecord[]>([])
const selectedRole = ref<RoleRecord | null>(null)
const selectedPermissionCodes = ref<string[]>([])
const permissionSearch = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const formOpen = ref(false)
const confirmOpen = ref(false)
const targetRole = ref<RoleRecord | null>(null)
const form = reactive<RoleForm>({ id: 0, code: '', name: '', description: '', is_active: true })

const canCreate = computed(() => auth.hasPermission('roles.create'))
const canUpdate = computed(() => auth.hasPermission('roles.update'))
const canDeactivate = computed(() => auth.hasPermission('roles.deactivate'))
const canAssign = computed(() => auth.hasPermission('roles.assign_permissions'))

const groupedPermissions = computed(() => {
  const query = permissionSearch.value.trim().toLowerCase()
  const groups: Record<string, PermissionRecord[]> = {}
  for (const permission of permissions.value) {
    if (query && !`${permission.code} ${permission.name} ${permission.module}`.toLowerCase().includes(query)) continue
    if (!groups[permission.module]) groups[permission.module] = []
    groups[permission.module].push(permission)
  }
  return Object.entries(groups).map(([module, items]) => ({ module, items }))
})

const editing = computed(() => form.id > 0)

function resetForm() {
  Object.assign(form, { id: 0, code: '', name: '', description: '', is_active: true })
}

function openCreate() {
  resetForm()
  formOpen.value = true
}

function openEdit(role: RoleRecord) {
  Object.assign(form, {
    id: role.id,
    code: role.code,
    name: role.name,
    description: role.description,
    is_active: role.is_active,
  })
  formOpen.value = true
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
    if (!selectedRole.value && roleRows[0]) await selectRole(roleRows[0])
    else if (selectedRole.value) {
      const fresh = roleRows.find((role) => role.id === selectedRole.value?.id)
      if (fresh) selectedRole.value = fresh
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load roles'
  } finally {
    loading.value = false
  }
}

async function selectRole(role: RoleRecord) {
  selectedRole.value = role
  selectedPermissionCodes.value = await apiClient<string[]>(`/v1/roles/${role.id}/permissions`)
}

async function saveRole() {
  saving.value = true
  error.value = ''
  try {
    const payload = {
      code: form.code,
      name: form.name,
      description: form.description,
      is_active: form.is_active,
    }
    const role = editing.value
      ? await patchJSON<RoleRecord>(`/v1/roles/${form.id}`, payload)
      : await postJSON<RoleRecord>('/v1/roles', payload)
    app.pushToast({ type: 'success', message: editing.value ? 'Role updated' : 'Role created' })
    formOpen.value = false
    await load()
    await selectRole(role)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save role'
    app.pushToast({ type: 'error', message: 'Could not save role', description: error.value })
  } finally {
    saving.value = false
  }
}

function askDeactivate(role: RoleRecord) {
  targetRole.value = role
  confirmOpen.value = true
}

async function deactivateRole() {
  if (!targetRole.value) return
  saving.value = true
  try {
    await patchJSON<RoleRecord>(`/v1/roles/${targetRole.value.id}/status`, { is_active: !targetRole.value.is_active })
    app.pushToast({ type: 'success', message: targetRole.value.is_active ? 'Role deactivated' : 'Role activated' })
    confirmOpen.value = false
    await load()
  } catch (err) {
    app.pushToast({ type: 'error', message: 'Could not update role', description: err instanceof Error ? err.message : '' })
  } finally {
    saving.value = false
  }
}

function togglePermission(code: string, checked: boolean) {
  const next = new Set(selectedPermissionCodes.value)
  if (checked) next.add(code)
  else next.delete(code)
  selectedPermissionCodes.value = [...next].sort()
}

function moduleChecked(items: PermissionRecord[]) {
  return items.every((item) => selectedPermissionCodes.value.includes(item.code))
}

function selectModule(items: PermissionRecord[]) {
  const next = new Set(selectedPermissionCodes.value)
  items.forEach((item) => next.add(item.code))
  selectedPermissionCodes.value = [...next].sort()
}

function clearModule(items: PermissionRecord[]) {
  const remove = new Set(items.map((item) => item.code))
  selectedPermissionCodes.value = selectedPermissionCodes.value.filter((code) => !remove.has(code))
}

async function savePermissions() {
  if (!selectedRole.value) return
  saving.value = true
  try {
    selectedPermissionCodes.value = await apiClient<string[]>(`/v1/roles/${selectedRole.value.id}/permissions`, {
      method: 'PUT',
      body: JSON.stringify({ permission_codes: selectedPermissionCodes.value }),
    })
    app.pushToast({ type: 'success', message: 'Permissions saved' })
    await load()
  } catch (err) {
    app.pushToast({ type: 'error', message: 'Could not save permissions', description: err instanceof Error ? err.message : '' })
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="จัดการบทบาทและสิทธิ์" eyebrow="Roles & Permissions" description="กำหนดว่าแต่ละบทบาทสามารถเข้าถึงเมนูและทำรายการใดได้บ้าง" icon="settings">
      <AppButton v-if="canCreate" icon="plus" @click="openCreate">Create role</AppButton>
    </PageHeader>

    <div v-if="error" class="mb-4 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-semibold text-red-700">{{ error }}</div>
    <AppLoadingState v-if="loading" class="mb-4" label="Loading roles..." />

    <div class="grid gap-4 xl:grid-cols-[420px_minmax(0,1fr)]">
      <AppCard>
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="font-bold">บทบาท</h2>
            <p class="text-sm text-slate-500">{{ roles.length }} roles</p>
          </div>
          <AppBadge tone="info">RBAC</AppBadge>
        </div>
        <AppEmptyState v-if="!loading && roles.length === 0" class="mt-4" title="No roles" description="Create a role to assign permissions." />
        <div v-else class="mt-4 grid gap-3">
          <article
            v-for="role in roles"
            :key="role.id"
            class="cursor-pointer rounded-2xl border p-4 transition hover:border-brand-300 hover:bg-brand-50/60"
            :class="selectedRole?.id === role.id ? 'border-brand-500 bg-brand-50/80' : 'border-slate-200 bg-white/70'"
            @click="selectRole(role)"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="truncate font-black">{{ role.name }}</h3>
                <p class="text-xs font-bold uppercase text-slate-500">{{ role.code }}</p>
              </div>
              <div class="flex flex-wrap justify-end gap-1">
                <AppBadge v-if="role.is_system">บทบาทระบบ</AppBadge>
                <AppBadge :tone="role.is_active ? 'success' : 'neutral'">{{ role.is_active ? 'Active' : 'Inactive' }}</AppBadge>
              </div>
            </div>
            <p class="mt-2 line-clamp-2 text-sm text-slate-500">{{ role.description || 'No description' }}</p>
            <div class="mt-3 grid grid-cols-2 gap-2 text-sm">
              <div class="rounded-xl bg-slate-50 p-3">
                <p class="text-slate-500">สิทธิ์ในระบบ</p>
                <p class="text-lg font-black">{{ role.permission_count }}</p>
              </div>
              <div class="rounded-xl bg-slate-50 p-3">
                <p class="text-slate-500">ผู้ใช้งาน</p>
                <p class="text-lg font-black">{{ role.user_count }}</p>
              </div>
            </div>
            <div class="mt-3 flex flex-wrap gap-2">
              <AppButton v-if="canUpdate" variant="secondary" icon="settings" @click.stop="openEdit(role)">Edit</AppButton>
              <AppButton v-if="canDeactivate" :variant="role.is_active ? 'danger' : 'secondary'" icon="triangle-alert" @click.stop="askDeactivate(role)">
                {{ role.is_active ? 'Deactivate' : 'Activate' }}
              </AppButton>
            </div>
          </article>
        </div>
      </AppCard>

      <AppCard>
        <div v-if="!selectedRole">
          <AppEmptyState title="Select a role" description="Choose a role to configure permissions." />
        </div>
        <div v-else class="grid gap-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p class="text-xs font-bold uppercase text-brand-700">สิทธิ์การใช้งาน</p>
              <h2 class="text-2xl font-black">{{ selectedRole.name }}</h2>
              <p class="mt-1 text-sm text-slate-500">{{ selectedPermissionCodes.length }} selected permissions</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <AppBadge v-if="selectedRole.is_system">บทบาทระบบ</AppBadge>
              <AppButton :disabled="!canAssign || saving" :loading="saving" icon="check-circle" @click="savePermissions">บันทึกสิทธิ์</AppButton>
            </div>
          </div>

          <div v-if="selectedRole.is_system" class="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm font-semibold text-amber-800">
            This is a system role. Keep admin-capable permissions assigned to avoid lockout.
          </div>

          <AppInput v-model="permissionSearch" label="Search permissions" placeholder="products.view, reports, บทบาท..." />

          <div class="grid gap-3">
            <section v-for="group in groupedPermissions" :key="group.module" class="rounded-2xl border border-slate-200 bg-white/70 p-4">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 class="font-black capitalize">{{ group.module.replaceAll('_', ' ') }}</h3>
                  <p class="text-sm text-slate-500">{{ group.items.length }} permissions</p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <AppButton variant="secondary" :disabled="!canAssign || moduleChecked(group.items)" @click="selectModule(group.items)">เลือกทั้งหมด</AppButton>
                  <AppButton variant="ghost" :disabled="!canAssign" @click="clearModule(group.items)">ล้างทั้งหมด</AppButton>
                </div>
              </div>
              <div class="mt-4 grid gap-2 md:grid-cols-2">
                <label v-for="permission in group.items" :key="permission.id" class="flex items-start gap-3 rounded-xl border border-slate-200 bg-white/75 p-3 text-sm">
                  <input
                    class="mt-1 h-4 w-4 rounded border-slate-300 text-brand-600"
                    type="checkbox"
                    :checked="selectedPermissionCodes.includes(permission.code)"
                    :disabled="!canAssign"
                    @change="togglePermission(permission.code, ($event.target as HTMLInputElement).checked)"
                  />
                  <span class="min-w-0">
                    <span class="block break-words font-bold">{{ permission.code }}</span>
                    <span class="block text-xs text-slate-500">{{ permission.name }}</span>
                  </span>
                </label>
              </div>
            </section>
          </div>
        </div>
      </AppCard>
    </div>

    <AppModal :open="formOpen" :title="editing ? 'Edit role' : 'Create role'" @close="formOpen = false">
      <form class="grid gap-4" @submit.prevent="saveRole">
        <AppInput v-model="form.code" label="Role code" placeholder="INVENTORY_STAFF" :disabled="editing" helper="Code is stable and cannot be changed after creation." />
        <AppInput v-model="form.name" label="Role name" />
        <AppTextarea v-model="form.description" label="Description" />
        <label class="flex items-center gap-2 text-sm font-semibold text-slate-700">
          <input v-model="form.is_active" type="checkbox" />
          Active
        </label>
        <div class="flex justify-end gap-2">
          <AppButton variant="secondary" @click="formOpen = false">Cancel</AppButton>
          <AppButton type="submit" :loading="saving" :disabled="saving" icon="check-circle">Save role</AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="confirmOpen"
      :title="targetRole?.is_active ? 'Deactivate role?' : 'Activate role?'"
      :message="targetRole ? `Role: ${targetRole.name}` : ''"
      :destructive="targetRole?.is_active"
      :loading="saving"
      confirm-label="Confirm"
      @close="confirmOpen = false"
      @confirm="deactivateRole"
    />
  </section>
</template>
