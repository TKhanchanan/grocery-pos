<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import PageHeader from '../components/PageHeader.vue'
import type { Role, User } from '../types/navigation'

interface UserForm {
  id: number
  username: string
  password: string
  fullName: string
  role: Role
  active: boolean
}

const users = ref<User[]>([])
const loading = ref(false)
const error = ref('')
const form = reactive<UserForm>({
  id: 0,
  username: '',
  password: '',
  fullName: '',
  role: 'CASHIER',
  active: true,
})

const editing = computed(() => form.id > 0)

async function loadUsers() {
  loading.value = true
  error.value = ''
  try {
    users.value = await apiClient<User[]>('/v1/users')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load users'
  } finally {
    loading.value = false
  }
}

function resetForm() {
  Object.assign(form, { id: 0, username: '', password: '', fullName: '', role: 'CASHIER' as Role, active: true })
}

function editUser(user: User) {
  Object.assign(form, { id: user.id, username: user.username, password: '', fullName: user.fullName, role: user.role, active: user.active })
}

async function saveUser() {
  error.value = ''
  const payload = {
    username: form.username,
    password: form.password,
    fullName: form.fullName,
    role: form.role,
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
    <PageHeader title="Users" eyebrow="Admin only" description="Create, edit, and disable users without exposing password hashes." />
    <div class="grid gap-4 lg:grid-cols-[360px_1fr]">
      <AppCard>
        <form class="grid gap-3" @submit.prevent="saveUser">
          <h2 class="font-bold">{{ editing ? 'Edit user' : 'Create user' }}</h2>
          <AppInput v-model="form.username" label="Username" />
          <AppInput v-model="form.fullName" label="Full name" />
          <AppInput v-model="form.password" label="Password" type="password" :placeholder="editing ? 'Leave blank to keep current password' : ''" />
          <AppSelect v-model="form.role" label="Role">
            <option value="ADMIN">ADMIN</option>
            <option value="MANAGER">MANAGER</option>
            <option value="CASHIER">CASHIER</option>
          </AppSelect>
          <label class="flex items-center gap-2 text-sm font-semibold text-slate-700">
            <input v-model="form.active" type="checkbox" />
            Active
          </label>
          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <div class="flex gap-2">
            <AppButton type="submit">{{ editing ? 'Save changes' : 'Create user' }}</AppButton>
            <AppButton v-if="editing" variant="secondary" @click="resetForm">Cancel</AppButton>
          </div>
        </form>
      </AppCard>

      <AppCard>
        <div class="flex items-center justify-between">
          <h2 class="font-bold">User list</h2>
          <span class="text-sm text-slate-500">{{ users.length }} users</span>
        </div>
        <div v-if="loading" class="mt-4 text-sm text-slate-500">Loading users...</div>
        <AppEmptyState v-else-if="users.length === 0" class="mt-4" title="No users" description="Create the first user from the form." />
        <div v-else class="mt-4 overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50">
              <tr>
                <th class="px-3 py-2 text-left">Username</th>
                <th class="px-3 py-2 text-left">Name</th>
                <th class="px-3 py-2 text-left">Role</th>
                <th class="px-3 py-2 text-left">Status</th>
                <th class="px-3 py-2 text-left">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="user in users" :key="user.id">
                <td class="px-3 py-2 font-semibold">{{ user.username }}</td>
                <td class="px-3 py-2">{{ user.fullName }}</td>
                <td class="px-3 py-2">{{ user.role }}</td>
                <td class="px-3 py-2">{{ user.active ? 'Active' : 'Disabled' }}</td>
                <td class="px-3 py-2">
                  <div class="flex flex-wrap gap-2">
                    <AppButton variant="secondary" @click="editUser(user)">Edit</AppButton>
                    <AppButton :variant="user.active ? 'danger' : 'secondary'" @click="setActive(user, !user.active)">
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
