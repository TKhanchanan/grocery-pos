<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAuthStore } from '../stores/auth'
import type { Category } from '../types/navigation'

const auth = useAuthStore()
const categories = ref<Category[]>([])
const loading = ref(false)
const error = ref('')
const form = reactive({ id: 0, name: '', description: '' })
const canCreate = computed(() => auth.hasPermission('categories.create'))
const canUpdate = computed(() => auth.hasPermission('categories.update'))

async function load() {
  loading.value = true
  error.value = ''
  try {
    categories.value = await apiClient<Category[]>('/v1/categories')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load categories'
  } finally {
    loading.value = false
  }
}

function edit(item: Category) {
  Object.assign(form, { id: item.id, name: item.name, description: item.description })
}

function reset() {
  Object.assign(form, { id: 0, name: '', description: '' })
}

async function save() {
  const payload = { name: form.name, description: form.description }
  if (form.id) await patchJSON<Category>(`/v1/categories/${form.id}`, payload)
  else await postJSON<Category>('/v1/categories', payload)
  reset()
  await load()
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader title="Categories" eyebrow="Catalog" description="Manage product category names and grouping." />
    <div class="grid gap-4 lg:grid-cols-[340px_1fr]">
      <AppCard v-if="canCreate || canUpdate">
        <form class="grid gap-3" @submit.prevent="save">
          <h2 class="font-bold">{{ form.id ? 'Edit category' : 'Create category' }}</h2>
          <AppInput v-model="form.name" label="Name" />
          <AppInput v-model="form.description" label="Description" />
          <div class="flex gap-2">
            <AppButton type="submit" :disabled="(!form.id && !canCreate) || (Boolean(form.id) && !canUpdate)">{{ form.id ? 'Save' : 'Create' }}</AppButton>
            <AppButton v-if="form.id" variant="secondary" @click="reset">Cancel</AppButton>
          </div>
        </form>
      </AppCard>
      <AppCard>
        <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
        <div v-else-if="loading" class="text-sm text-slate-500">Loading categories...</div>
        <AppEmptyState v-else-if="categories.length === 0" title="No categories" />
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50"><tr><th class="px-3 py-2 text-left">Name</th><th class="px-3 py-2 text-left">Description</th><th class="px-3 py-2 text-left">Status</th><th class="px-3 py-2 text-left"></th></tr></thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="item in categories" :key="item.id">
                <td class="px-3 py-2 font-semibold">{{ item.name }}</td>
                <td class="px-3 py-2">{{ item.description || '-' }}</td>
                <td class="px-3 py-2">{{ item.is_active ? 'Active' : 'Inactive' }}</td>
                <td class="px-3 py-2"><AppButton v-if="canUpdate" variant="secondary" @click="edit(item)">Edit</AppButton></td>
              </tr>
            </tbody>
          </table>
        </div>
      </AppCard>
    </div>
  </section>
</template>
