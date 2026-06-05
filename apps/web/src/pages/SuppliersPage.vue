<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import AppInput from '../components/AppInput.vue'
import AppTextarea from '../components/AppTextarea.vue'
import PageHeader from '../components/PageHeader.vue'
import type { Supplier } from '../types/navigation'

const suppliers = ref<Supplier[]>([])
const editingID = ref<number | null>(null)
const loading = ref(false)
const error = ref('')

const form = reactive({
  name: '',
  phone: '',
  email: '',
  address: '',
})

function resetForm() {
  editingID.value = null
  form.name = ''
  form.phone = ''
  form.email = ''
  form.address = ''
}

async function loadSuppliers() {
  loading.value = true
  error.value = ''
  try {
    suppliers.value = await apiClient<Supplier[]>('/v1/suppliers')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load suppliers'
  } finally {
    loading.value = false
  }
}

function editSupplier(supplier: Supplier) {
  editingID.value = supplier.id
  form.name = supplier.name
  form.phone = supplier.phone
  form.email = supplier.email
  form.address = supplier.address
}

async function saveSupplier() {
  error.value = ''
  try {
    const payload = { name: form.name, phone: form.phone, email: form.email, address: form.address }
    if (editingID.value) {
      await patchJSON<Supplier>(`/v1/suppliers/${editingID.value}`, payload)
    } else {
      await postJSON<Supplier>('/v1/suppliers', payload)
    }
    resetForm()
    await loadSuppliers()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save supplier'
  }
}

async function toggleSupplier(supplier: Supplier) {
  await patchJSON<Supplier>(`/v1/suppliers/${supplier.id}/status`, { is_active: !supplier.is_active })
  await loadSuppliers()
}

onMounted(loadSuppliers)
</script>

<template>
  <section>
    <PageHeader title="Suppliers" eyebrow="Purchasing" description="Supplier contact records used for purchase order workflows." />
    <div class="grid gap-4 xl:grid-cols-[380px_1fr]">
      <AppCard>
        <form class="grid gap-3" @submit.prevent="saveSupplier">
          <h2 class="font-bold">{{ editingID ? 'Edit supplier' : 'Create supplier' }}</h2>
          <AppInput v-model="form.name" label="Name" />
          <AppInput v-model="form.phone" label="Phone" />
          <AppInput v-model="form.email" label="Email" />
          <AppTextarea v-model="form.address" label="Address" />
          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <div class="flex gap-2">
            <AppButton type="submit">{{ editingID ? 'Update' : 'Create' }}</AppButton>
            <AppButton variant="secondary" @click="resetForm">Clear</AppButton>
          </div>
        </form>
      </AppCard>

      <AppCard>
        <div class="flex items-center justify-between gap-3">
          <h2 class="font-bold">Supplier list</h2>
        </div>
        <div v-if="loading" class="mt-4 text-sm text-slate-500">Loading suppliers...</div>
        <AppEmptyState v-else-if="suppliers.length === 0" class="mt-4" title="No suppliers" description="Create a supplier to start purchase orders." />
        <div v-else class="mt-4 grid gap-3 md:grid-cols-2">
          <article v-for="supplier in suppliers" :key="supplier.id" class="rounded-lg border border-slate-200 p-3">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="truncate font-bold">{{ supplier.name }}</h3>
                <p class="text-sm text-slate-500">{{ supplier.phone || '-' }} · {{ supplier.email || '-' }}</p>
              </div>
              <span class="rounded-full px-2 py-1 text-xs font-bold" :class="supplier.is_active ? 'bg-brand-100 text-brand-700' : 'bg-slate-100 text-slate-600'">
                {{ supplier.is_active ? 'ACTIVE' : 'INACTIVE' }}
              </span>
            </div>
            <p class="mt-2 text-sm text-slate-600">{{ supplier.address || 'No address' }}</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <AppButton variant="secondary" @click="editSupplier(supplier)">Edit</AppButton>
              <AppButton :variant="supplier.is_active ? 'danger' : 'secondary'" @click="toggleSupplier(supplier)">
                {{ supplier.is_active ? 'Disable' : 'Enable' }}
              </AppButton>
            </div>
          </article>
        </div>
      </AppCard>
    </div>
  </section>
</template>
