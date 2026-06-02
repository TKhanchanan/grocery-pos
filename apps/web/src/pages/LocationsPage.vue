<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAuthStore } from '../stores/auth'
import type { Location } from '../types/navigation'

const auth = useAuthStore()
const locations = ref<Location[]>([])
const loading = ref(false)
const error = ref('')
const form = reactive({ id: 0, name: '', description: '' })
const canManage = computed(() => auth.can(['ADMIN', 'MANAGER']))

async function load() {
  loading.value = true
  error.value = ''
  try {
    locations.value = await apiClient<Location[]>('/v1/locations')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load locations'
  } finally {
    loading.value = false
  }
}

function edit(item: Location) {
  Object.assign(form, { id: item.id, name: item.name, description: item.description })
}

function reset() {
  Object.assign(form, { id: 0, name: '', description: '' })
}

async function save() {
  const payload = { name: form.name, description: form.description }
  if (form.id) await patchJSON<Location>(`/v1/locations/${form.id}`, payload)
  else await postJSON<Location>('/v1/locations', payload)
  reset()
  await load()
}

async function setActive(item: Location, active: boolean) {
  await patchJSON<Location>(`/v1/locations/${item.id}/status`, { is_active: active })
  await load()
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="Locations" eyebrow="Inventory" description="Manage shop front, warehouse, and other inventory locations." />
    <div class="grid gap-4 lg:grid-cols-[340px_1fr]">
      <AppCard v-if="canManage">
        <form class="grid gap-3" @submit.prevent="save">
          <h2 class="font-bold">{{ form.id ? 'Edit location' : 'Create location' }}</h2>
          <AppInput v-model="form.name" label="Name" />
          <AppInput v-model="form.description" label="Description" />
          <div class="flex gap-2">
            <AppButton type="submit">{{ form.id ? 'Save' : 'Create' }}</AppButton>
            <AppButton v-if="form.id" variant="secondary" @click="reset">Cancel</AppButton>
          </div>
        </form>
      </AppCard>
      <AppCard>
        <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
        <div v-else-if="loading" class="text-sm text-slate-500">Loading locations...</div>
        <AppEmptyState v-else-if="locations.length === 0" title="No locations" />
        <div v-else class="grid gap-3">
          <article v-for="item in locations" :key="item.id" class="rounded-lg border border-slate-200 p-3">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h3 class="font-bold">{{ item.name }}</h3>
                <p class="text-sm text-slate-500">{{ item.description || 'No description' }}</p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full px-2 py-1 text-xs font-bold" :class="item.is_active ? 'bg-brand-100 text-brand-700' : 'bg-slate-100 text-slate-600'">{{ item.is_active ? 'Active' : 'Inactive' }}</span>
                <AppButton v-if="canManage" variant="secondary" @click="edit(item)">Edit</AppButton>
                <AppButton v-if="canManage" :variant="item.is_active ? 'danger' : 'secondary'" @click="setActive(item, !item.is_active)">{{ item.is_active ? 'Deactivate' : 'Activate' }}</AppButton>
              </div>
            </div>
          </article>
        </div>
      </AppCard>
    </div>
  </section>
</template>
