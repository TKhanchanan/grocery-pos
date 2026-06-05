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
import AppSelect from '../components/AppSelect.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { Role, RoleRecord, User } from '../types/navigation'

interface UserForm {
  id: number
  username: string
  password: string
  fullName: string
  role: Role
  role_ids: number[]
  active: boolean
}

const app = useAppStore()
const auth = useAuthStore()
const users = ref<User[]>([])
const roles = ref<RoleRecord[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const formOpen = ref(false)
const targetUser = ref<User | null>(null)
const confirmOpen = ref(false)
const form = reactive<UserForm>({
  id: 0,
  username: '',
  password: '',
  fullName: '',
  role: 'CASHIER',
  role_ids: [],
  active: true,
})

const editing = computed(() => form.id > 0)
const canCreate = computed(() => auth.hasPermission('users.create'))
const canUpdate = computed(() => auth.hasPermission('users.update'))
const canDeactivate = computed(() => auth.hasPermission('users.deactivate'))
const canAssignRoles = computed(() => auth.hasPermission('users.assign_roles'))
const activeCount = computed(() => users.value.filter((user) => user.active).length)
const inactiveCount = computed(() => users.value.length - activeCount.value)
const adminCount = computed(() => users.value.filter((user) => user.role === 'ADMIN' || user.roles?.some((role) => role.code === 'ADMIN')).length)

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) text = text.replaceAll(`{${name}}`, String(value))
  return text
}

function friendlyError(err: unknown, fallback: TranslationKey) {
  const message = err instanceof Error ? err.message : app.t(fallback)
  return message.toLowerCase().includes('permission') ? app.t('users.noPermission') : message
}

async function loadUsers() {
  loading.value = true
  error.value = ''
  try {
    const [userRows, roleRows] = await Promise.all([
      apiClient<User[]>('/v1/users'),
      apiClient<RoleRecord[]>('/v1/roles'),
    ])
    users.value = userRows
    roles.value = roleRows.filter((role) => role.is_active)
  } catch (err) {
    error.value = friendlyError(err, 'users.loadFailed')
  } finally {
    loading.value = false
  }
}

function resetForm() {
  const cashier = roles.value.find((role) => role.code === 'CASHIER')
  Object.assign(form, {
    id: 0,
    username: '',
    password: '',
    fullName: '',
    role: 'CASHIER' as Role,
    role_ids: cashier ? [cashier.id] : [],
    active: true,
  })
  error.value = ''
}

function openCreate() {
  resetForm()
  formOpen.value = true
}

function openEdit(user: User) {
  Object.assign(form, {
    id: user.id,
    username: user.username,
    password: '',
    fullName: user.fullName,
    role: user.role,
    role_ids: user.roles?.map((role) => role.id).filter(Boolean) ?? [],
    active: user.active,
  })
  error.value = ''
  formOpen.value = true
}

function closeForm() {
  if (saving.value) return
  formOpen.value = false
  resetForm()
}

function toggleRole(roleID: number, checked: boolean) {
  const next = new Set(form.role_ids)
  if (checked) next.add(roleID)
  else next.delete(roleID)
  form.role_ids = [...next]
}

function validateForm() {
  if (!form.username.trim()) return app.t('users.usernameRequired')
  if (!form.fullName.trim()) return app.t('users.nameRequired')
  if (!editing.value && !form.password.trim()) return app.t('users.passwordRequired')
  return ''
}

async function saveUser() {
  const validation = validateForm()
  if (validation) {
    error.value = validation
    return
  }
  error.value = ''
  saving.value = true
  const payload = {
    username: form.username.trim(),
    password: form.password,
    fullName: form.fullName.trim(),
    role: form.role,
    role_ids: form.role_ids,
    active: form.active,
  }
  try {
    if (editing.value) await patchJSON<User>(`/v1/users/${form.id}`, payload)
    else await postJSON<User>('/v1/users', payload)
    await loadUsers()
    app.pushToast({ type: 'success', message: editing.value ? app.t('users.updated') : app.t('users.created') })
    closeForm()
  } catch (err) {
    error.value = friendlyError(err, 'users.saveFailed')
    app.pushToast({ type: 'error', message: app.t('users.saveFailed'), description: error.value })
  } finally {
    saving.value = false
  }
}

function askStatus(user: User) {
  targetUser.value = user
  confirmOpen.value = true
}

async function confirmStatus() {
  if (!targetUser.value) return
  saving.value = true
  try {
    await patchJSON<User>(`/v1/users/${targetUser.value.id}/status`, { active: !targetUser.value.active })
    await loadUsers()
    app.pushToast({ type: 'success', message: app.t('users.statusUpdated') })
    confirmOpen.value = false
    targetUser.value = null
  } catch (err) {
    const message = friendlyError(err, 'users.statusFailed')
    app.pushToast({ type: 'error', message: app.t('users.statusFailed'), description: message })
  } finally {
    saving.value = false
  }
}

function userEmail(user: User) {
  return (user as User & { email?: string }).email || '-'
}

onMounted(loadUsers)
</script>

<template>
  <section>
    <PageHeader :title="app.t('users.title')" :eyebrow="app.t('users.eyebrow')" :description="app.t('users.description')" icon="users">
      <div class="flex flex-wrap gap-2">
        <AppButton v-if="canCreate" icon="plus" @click="openCreate">{{ app.t('users.add') }}</AppButton>
      </div>
    </PageHeader>

    <div class="mb-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard :label="app.t('users.total')" :value="users.length" :helper="app.t('users.totalHelper')" icon="users" />
      <StatCard :label="app.t('users.active')" :value="activeCount" :helper="app.t('users.activeHelper')" icon="check-circle" tone="success" />
      <StatCard :label="app.t('users.inactive')" :value="inactiveCount" :helper="app.t('users.inactiveHelper')" icon="triangle-alert" tone="warning" />
      <StatCard :label="app.t('users.admins')" :value="adminCount" :helper="app.t('users.adminsHelper')" icon="role" tone="info" />
    </div>

    <AppCard class="dark:bg-slate-900/80">
      <div v-if="error && !formOpen" class="mb-3 rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
      <AppLoadingState v-if="loading" :label="app.t('users.loading')" />
      <AppEmptyState v-else-if="users.length === 0" :title="app.t('users.empty')" :description="app.t('users.emptyDescription')" icon="users">
        <template v-if="canCreate">
          <AppButton icon="plus" @click="openCreate">{{ app.t('users.add') }}</AppButton>
        </template>
      </AppEmptyState>

      <div v-else>
        <div class="hidden overflow-x-auto lg:block">
          <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
            <thead class="bg-slate-50 dark:bg-slate-950/70">
              <tr>
                <th class="px-3 py-3 text-left">{{ app.t('users.name') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('users.username') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('users.email') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('users.roles') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('users.status') }}</th>
                <th class="px-3 py-3 text-left">{{ app.t('users.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                <td class="px-3 py-3 font-semibold">{{ user.fullName }}</td>
                <td class="px-3 py-3">{{ user.username }}</td>
                <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ userEmail(user) }}</td>
                <td class="px-3 py-3">
                  <div class="flex flex-wrap gap-1">
                    <AppBadge v-for="role in user.roles?.length ? user.roles : [{ code: user.role, name: user.role }]" :key="role.code">{{ role.name }}</AppBadge>
                  </div>
                </td>
                <td class="px-3 py-3"><AppBadge :tone="user.active ? 'success' : 'neutral'">{{ user.active ? app.t('users.active') : app.t('users.inactive') }}</AppBadge></td>
                <td class="px-3 py-3">
                  <div class="flex flex-wrap gap-2">
                    <AppButton v-if="canUpdate" variant="secondary" icon="settings" @click="openEdit(user)">{{ app.t('users.edit') }}</AppButton>
                    <AppButton v-if="canDeactivate" :variant="user.active ? 'danger' : 'secondary'" @click="askStatus(user)">
                      {{ user.active ? app.t('users.deactivate') : app.t('users.activate') }}
                    </AppButton>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-3 lg:hidden">
          <article v-for="user in users" :key="user.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 dark:border-slate-700 dark:bg-slate-950/60">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="truncate font-black">{{ user.fullName }}</h3>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ user.username }} · {{ userEmail(user) }}</p>
              </div>
              <AppBadge :tone="user.active ? 'success' : 'neutral'">{{ user.active ? app.t('users.active') : app.t('users.inactive') }}</AppBadge>
            </div>
            <div class="mt-3 flex flex-wrap gap-1">
              <AppBadge v-for="role in user.roles?.length ? user.roles : [{ code: user.role, name: user.role }]" :key="role.code">{{ role.name }}</AppBadge>
            </div>
            <div class="mt-3 flex flex-wrap gap-2">
              <AppButton v-if="canUpdate" variant="secondary" icon="settings" @click="openEdit(user)">{{ app.t('users.edit') }}</AppButton>
              <AppButton v-if="canDeactivate" :variant="user.active ? 'danger' : 'secondary'" @click="askStatus(user)">
                {{ user.active ? app.t('users.deactivate') : app.t('users.activate') }}
              </AppButton>
            </div>
          </article>
        </div>
      </div>
    </AppCard>

    <AppModal :open="formOpen" :title="editing ? app.t('users.editTitle') : app.t('users.createTitle')" :description="app.t('users.modalDescription')" :close-label="app.t('users.cancel')" size="lg" @close="closeForm">
      <form class="grid gap-4" @submit.prevent="saveUser">
        <div class="grid gap-3 md:grid-cols-2">
          <AppInput v-model="form.fullName" :label="app.t('users.name')" />
          <AppInput v-model="form.username" :label="app.t('users.username')" />
          <AppInput v-model="form.password" :label="app.t('users.password')" type="password" :placeholder="editing ? app.t('users.passwordOptional') : ''" />
          <AppSelect v-model="form.role" :label="app.t('users.legacyRole')">
            <option value="ADMIN">ADMIN</option>
            <option value="MANAGER">MANAGER</option>
            <option value="CASHIER">CASHIER</option>
          </AppSelect>
        </div>

        <section class="grid gap-2 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/45">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('users.dynamicRoles') }}</p>
            <AppBadge tone="info">{{ t('users.selectedCount', { count: form.role_ids.length }) }}</AppBadge>
          </div>
          <div class="grid gap-2 md:grid-cols-2">
            <label v-for="role in roles" :key="role.id" class="flex items-start gap-3 rounded-xl border border-slate-200 bg-white/70 p-3 text-sm dark:border-slate-700 dark:bg-slate-900/60">
              <input
                class="mt-1 h-4 w-4 rounded border-slate-300 text-brand-600"
                type="checkbox"
                :checked="form.role_ids.includes(role.id)"
                :disabled="!canAssignRoles"
                @change="toggleRole(role.id, ($event.target as HTMLInputElement).checked)"
              />
              <span>
                <span class="block font-bold">{{ role.name }}</span>
                <span class="block text-xs text-slate-500 dark:text-slate-400">{{ role.code }}</span>
              </span>
            </label>
          </div>
        </section>

        <label class="flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-200">
          <input v-model="form.active" type="checkbox" />
          {{ app.t('users.active') }}
        </label>
        <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
        <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <AppButton type="button" variant="secondary" :disabled="saving" @click="closeForm">{{ app.t('users.cancel') }}</AppButton>
          <AppButton type="submit" :loading="saving" :disabled="saving || (editing ? !canUpdate : !canCreate)" icon="check-circle">
            {{ editing ? app.t('users.save') : app.t('users.create') }}
          </AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="confirmOpen"
      :title="targetUser?.active ? app.t('users.confirmDeactivate') : app.t('users.confirmActivate')"
      :message="targetUser ? t('users.confirmStatusMessage', { name: targetUser.fullName || targetUser.username }) : ''"
      :consequence="targetUser?.active ? app.t('users.deactivateConsequence') : undefined"
      :confirm-label="targetUser?.active ? app.t('users.deactivate') : app.t('users.activate')"
      :cancel-label="app.t('users.cancel')"
      :destructive="Boolean(targetUser?.active)"
      :loading="saving"
      @close="confirmOpen = false"
      @confirm="confirmStatus"
    />
  </section>
</template>
