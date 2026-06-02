<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, postJSON } from '../api'
import { useAuthStore } from '../stores/auth'
import type { Role, User } from '../types'

interface Setting { key: string; value: string }

const auth = useAuthStore()
const users = ref<User[]>([])
const settings = ref<Setting[]>([])
const userForm = reactive({ id: 0, username: '', password: '', fullName: '', role: 'CASHIER' as Role, active: true })
const settingForm = reactive({ key: 'line_notifications_enabled', value: 'false' })
const error = ref('')

async function load() {
  error.value = ''
  try {
    settings.value = await api<Setting[]>('/settings')
    users.value = await api<User[]>('/users')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Admin access required for settings and users'
  }
}

async function saveUser() {
  await postJSON('/users', userForm)
  Object.assign(userForm, { id: 0, username: '', password: '', fullName: '', role: 'CASHIER' as Role, active: true })
  await load()
}

async function saveSetting() {
  await postJSON('/settings', settingForm)
  await load()
}

onMounted(load)
</script>

<template>
  <section class="space-y-5">
    <div><p class="label">Security and shop config</p><h2 class="text-2xl font-bold">Settings & Users</h2></div>
    <p v-if="error" class="panel border-amber-200 bg-amber-50 text-sm text-amber-800">{{ error }}</p>
    <div v-if="auth.user?.role === 'ADMIN'" class="grid gap-5 lg:grid-cols-2">
      <form class="panel space-y-3" @submit.prevent="saveUser">
        <h3 class="font-bold">Users and Roles</h3>
        <input v-model="userForm.username" class="input" placeholder="Username" />
        <input v-model="userForm.fullName" class="input" placeholder="Full name" />
        <input v-model="userForm.password" class="input" type="password" placeholder="Password" />
        <select v-model="userForm.role" class="input"><option>ADMIN</option><option>MANAGER</option><option>CASHIER</option></select>
        <label class="flex items-center gap-2 text-sm"><input v-model="userForm.active" type="checkbox" /> Active</label>
        <button class="btn-primary w-full">Save User</button>
      </form>
      <form class="panel space-y-3" @submit.prevent="saveSetting">
        <h3 class="font-bold">Settings</h3>
        <input v-model="settingForm.key" class="input" />
        <input v-model="settingForm.value" class="input" />
        <button class="btn-primary w-full">Save Setting</button>
        <p class="text-sm text-slate-600">LINE notification calls are enabled when `LINE_CHANNEL_ACCESS_TOKEN` is set on the backend.</p>
      </form>
    </div>
    <div v-if="auth.user?.role === 'ADMIN'" class="grid gap-5 lg:grid-cols-2">
      <div class="table-wrap"><table class="table"><thead><tr><th>Username</th><th>Name</th><th>Role</th><th>Active</th></tr></thead><tbody><tr v-for="u in users" :key="u.id"><td>{{ u.username }}</td><td>{{ u.fullName }}</td><td>{{ u.role }}</td><td>{{ u.active }}</td></tr></tbody></table></div>
      <div class="table-wrap"><table class="table"><thead><tr><th>Key</th><th>Value</th></tr></thead><tbody><tr v-for="s in settings" :key="s.key"><td>{{ s.key }}</td><td>{{ s.value }}</td></tr></tbody></table></div>
    </div>
  </section>
</template>
