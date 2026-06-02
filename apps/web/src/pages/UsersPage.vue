<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
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

const auth = useAuthStore()
const users = ref<User[]>([])
const roles = ref<RoleRecord[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
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
    if (!editing.value && form.role_ids.length === 0) resetForm()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load users'
  } finally {
    loading.value = false
  }
}

function resetForm() {
  const cashier = roles.value.find((role) => role.code === 'CASHIER')
  Object.assign(form, { id: 0, username: '', password: '', fullName: '', role: 'CASHIER' as Role, role_ids: cashier ? [cashier.id] : [], active: true })
}

function editUser(user: User) {
  Object.assign(form, {
    id: user.id,
    username: user.username,
    password: '',
    fullName: user.fullName,
    role: user.role,
    role_ids: user.roles?.map((role) => role.id).filter(Boolean) ?? [],
    active: user.active,
  })
}

function toggleRole(roleID: number, checked: boolean) {
  const next = new Set(form.role_ids)
  if (checked) next.add(roleID)
  else next.delete(roleID)
  form.role_ids = [...next]
}

async function saveUser() {
  error.value = ''
  saving.value = true
  const payload = {
    username: form.username,
    password: form.password,
    fullName: form.fullName,
    role: form.role,
    role_ids: form.role_ids,
    active: form.active,
  }
  try {
    if (editing.value) {
      await patchJSON<User>(`/v1/users/${form.id}`, payload)
    } else {
      await postJSON<User>('/v1/users', payload)
    }
    resetForm()
    await loadUsers()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save user'
  } finally {
    saving.value = false
  }
}

async function setActive(user: User, active: boolean) {
  await patchJSON<User>(`/v1/users/${user.id}/status`, { active })
  await loadUsers()
}

onMounted(loadUsers)
</script>

<template>
  <section>
    <PageHeader title="Users" eyebrow="Admin only" description="Create, edit, disable users, and assign dynamic roles." icon="users" />
    <div class="grid gap-4 lg:grid-cols-[380px_1fr]">
      <AppCard v-if="canCreate || canUpdate">
        <form class="grid gap-3" @submit.prevent="saveUser">
          <h2 class="font-bold">{{ editing ? 'Edit user' : 'Create user' }}</h2>
          <AppInput v-model="form.username" label="Username" />
          <AppInput v-model="form.fullName" label="Full name" />
          <AppInput v-model="form.password" label="Password" type="password" :placeholder="editing ? 'Leave blank to keep current password' : ''" />
          <AppSelect v-model="form.role" label="Legacy role">
            <option value="ADMIN">ADMIN</option>
            <option value="MANAGER">MANAGER</option>
            <option value="CASHIER">CASHIER</option>
          </AppSelect>

          <div class="grid gap-2">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm font-semibold text-slate-700">Dynamic roles</p>
              <AppBadge tone="info">{{ form.role_ids.length }} selected</AppBadge>
            </div>
            <div class="grid gap-2">
              <label v-for="role in roles" :key="role.id" class="flex items-start gap-3 rounded-xl border border-slate-200 bg-white/70 p-3 text-sm">
                <input
                  class="mt-1 h-4 w-4 rounded border-slate-300 text-brand-600"
                  type="checkbox"
                  :checked="form.role_ids.includes(role.id)"
                  :disabled="!canAssignRoles"
                  @change="toggleRole(role.id, ($event.target as HTMLInputElement).checked)"
                />
                <span>
                  <span class="block font-bold">{{ role.name }}</span>
                  <span class="block text-xs text-slate-500">{{ role.code }}</span>
                </span>
              </label>
            </div>
          </div>

          <label class="flex items-center gap-2 text-sm font-semibold text-slate-700">
            <input v-model="form.active" type="checkbox" />
            Active
          </label>
          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <div class="flex gap-2">
            <AppButton type="submit" :loading="saving" :disabled="saving || (editing ? !canUpdate : !canCreate)">{{ editing ? 'Save changes' : 'Create user' }}</AppButton>
            <AppButton v-if="editing" variant="secondary" @click="resetForm">Cancel</AppButton>
          </div>
        </form>
      </AppCard>

      <AppCard>
        <div class="flex items-center justify-between">
          <h2 class="font-bold">User list</h2>
          <span class="text-sm text-slate-500">{{ users.length }} users</span>
        </div>
        <AppLoadingState v-if="loading" class="mt-4" label="Loading users..." />
        <AppEmptyState v-else-if="users.length === 0" class="mt-4" title="No users" description="Create the first user from the form." />
        <div v-else class="mt-4 overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50">
              <tr>
                <th class="px-3 py-2 text-left">Username</th>
                <th class="px-3 py-2 text-left">Name</th>
                <th class="px-3 py-2 text-left">Roles</th>
                <th class="px-3 py-2 text-left">Status</th>
                <th class="px-3 py-2 text-left">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="user in users" :key="user.id">
                <td class="px-3 py-2 font-semibold">{{ user.username }}</td>
                <td class="px-3 py-2">{{ user.fullName }}</td>
                <td class="px-3 py-2">
                  <div class="flex flex-wrap gap-1">
                    <AppBadge v-for="role in user.roles?.length ? user.roles : [{ code: user.role, name: user.role }]" :key="role.code">{{ role.name }}</AppBadge>
                  </div>
                </td>
                <td class="px-3 py-2">{{ user.active ? 'Active' : 'Disabled' }}</td>
                <td class="px-3 py-2">
                  <div class="flex flex-wrap gap-2">
                    <AppButton v-if="canUpdate" variant="secondary" @click="editUser(user)">Edit</AppButton>
                    <AppButton v-if="canDeactivate" :variant="user.active ? 'danger' : 'secondary'" @click="setActive(user, !user.active)">
                      {{ user.active ? 'Disable' : 'Enable' }}
                    </AppButton>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </AppCard>
    </div>
  </section>
</template>
