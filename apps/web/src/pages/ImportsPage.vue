<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { API_BASE_URL, apiClient, postJSON } from '../api/client'
import { downloadFile } from '../api/download'
import { authHeaders, handleAuthFailure } from '../api/session'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import type { ImportJob, ImportJobRow } from '../types/navigation'
import { formatThaiDateTime } from '../utils/date'

const jobs = ref<ImportJob[]>([])
const previewJob = ref<ImportJob | null>(null)
const selectedJob = ref<ImportJob | null>(null)
const selectedFile = ref<File | null>(null)
const loading = ref(false)
const uploading = ref(false)
const confirming = ref(false)
const error = ref('')

const validRows = computed(() => previewJob.value?.rows?.filter((row) => row.status === 'PENDING').length ?? 0)
const invalidRows = computed(() => previewJob.value?.rows?.filter((row) => row.status === 'FAILED').length ?? 0)

function rowClass(row: ImportJobRow) {
  return row.status === 'FAILED' ? 'bg-red-50' : row.status === 'IMPORTED' ? 'bg-brand-50' : ''
}

async function loadJobs() {
  loading.value = true
  error.value = ''
  try {
    jobs.value = await apiClient<ImportJob[]>('/v1/imports')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load imports'
  } finally {
    loading.value = false
  }
}

async function downloadTemplate() {
  await downloadFile('/v1/imports/products/template', 'product-import-template.csv')
}

function chooseFile(event: Event) {
  const input = event.target as HTMLInputElement
  selectedFile.value = input.files?.[0] ?? null
  previewJob.value = null
  error.value = ''
}

async function previewImport() {
  if (!selectedFile.value) return
  uploading.value = true
  error.value = ''
  const form = new FormData()
  form.append('file', selectedFile.value)
  try {
    const response = await fetch(`${API_BASE_URL}/v1/imports/products/preview`, {
      method: 'POST',
      headers: authHeaders(),
      body: form,
    })
    const envelope = await response.json()
    if (response.status === 401) throw handleAuthFailure(envelope.error?.message)
    if (!response.ok || !envelope.success) {
      throw new Error(envelope.error?.message ?? 'Preview failed')
    }
    previewJob.value = envelope.data as ImportJob
    selectedJob.value = previewJob.value
    await loadJobs()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Preview failed'
  } finally {
    uploading.value = false
  }
}

async function confirmImport() {
  if (!previewJob.value) return
  confirming.value = true
  error.value = ''
  try {
    previewJob.value = await postJSON<ImportJob>('/v1/imports/products/confirm', { job_id: previewJob.value.id })
    selectedJob.value = previewJob.value
    await loadJobs()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Confirm import failed'
  } finally {
    confirming.value = false
  }
}

async function openJob(job: ImportJob) {
  selectedJob.value = await apiClient<ImportJob>(`/v1/imports/${job.id}`)
}

onMounted(loadJobs)
</script>

<template>
  <section>
    <PageHeader title="Imports" eyebrow="Product CSV" description="Preview product import rows, inspect row errors, and confirm before saving.">
      <AppButton variant="secondary" @click="downloadTemplate">Download template</AppButton>
    </PageHeader>

    <div class="grid gap-4">
      <AppCard>
        <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <label class="grid gap-2 text-sm">
            <span class="font-semibold text-slate-700">Upload CSV file</span>
            <input class="rounded-md border border-dashed border-slate-300 bg-white p-4" type="file" accept=".csv,text/csv" @change="chooseFile" />
            <span class="text-xs text-slate-500">CSV columns: sku, name, barcode, category, selling_price, unit_cost, threshold, reorder_point, location, initial_stock.</span>
          </label>
          <div class="flex gap-2">
            <AppButton variant="secondary" :disabled="!selectedFile || uploading" @click="previewImport">
              {{ uploading ? 'Previewing...' : 'Preview file' }}
            </AppButton>
            <AppButton :disabled="!previewJob || previewJob.status !== 'PENDING' || validRows === 0 || confirming" @click="confirmImport">
              {{ confirming ? 'Importing...' : 'Confirm import' }}
            </AppButton>
          </div>
        </div>
        <div v-if="error" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
      </AppCard>

      <AppCard v-if="previewJob">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="font-bold">Preview: {{ previewJob.file_name }}</h2>
            <p class="text-sm text-slate-500">{{ validRows }} valid · {{ invalidRows }} invalid · not saved until confirm</p>
          </div>
          <span class="rounded-full bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600">{{ previewJob.status }}</span>
        </div>
        <div class="mt-4 overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50">
              <tr>
                <th class="px-3 py-2 text-left">Row</th>
                <th class="px-3 py-2 text-left">SKU</th>
                <th class="px-3 py-2 text-left">Name</th>
                <th class="px-3 py-2 text-left">Category</th>
                <th class="px-3 py-2 text-right">Price</th>
                <th class="px-3 py-2 text-right">Stock</th>
                <th class="px-3 py-2 text-left">Status / Error</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="row in previewJob.rows" :key="row.id" :class="rowClass(row)">
                <td class="px-3 py-2">{{ row.row_index }}</td>
                <td class="px-3 py-2 font-semibold">{{ row.raw_data.sku }}</td>
                <td class="px-3 py-2">{{ row.raw_data.name }}</td>
                <td class="px-3 py-2">{{ row.raw_data.category || '-' }}</td>
                <td class="px-3 py-2 text-right">{{ row.raw_data.selling_price }}</td>
                <td class="px-3 py-2 text-right">{{ row.raw_data.initial_stock ?? '-' }}</td>
                <td class="px-3 py-2">
                  <span class="font-bold" :class="row.status === 'FAILED' ? 'text-red-700' : 'text-brand-700'">{{ row.status }}</span>
                  <p v-if="row.error_message" class="mt-1 text-xs text-red-700">{{ row.error_message }}</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </AppCard>

      <div class="grid gap-4 xl:grid-cols-[360px_1fr]">
        <AppCard>
          <div class="flex items-center justify-between gap-3">
            <h2 class="font-bold">Import history</h2>
          </div>
          <div v-if="loading" class="mt-4 text-sm text-slate-500">Loading imports...</div>
          <AppEmptyState v-else-if="jobs.length === 0" class="mt-4" title="No imports" description="Previewed and confirmed imports appear here." />
          <div v-else class="mt-4 grid gap-2">
            <button v-for="job in jobs" :key="job.id" class="rounded-lg border border-slate-200 p-3 text-left hover:bg-slate-50" @click="openJob(job)">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <p class="truncate font-bold">#{{ job.id }} {{ job.file_name }}</p>
                  <p class="text-xs text-slate-500">{{ formatThaiDateTime(job.created_at) }}</p>
                </div>
                <span class="rounded-full bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600">{{ job.status }}</span>
              </div>
              <p class="mt-2 text-sm text-slate-500">{{ job.success_rows }} ok · {{ job.failed_rows }} failed</p>
            </button>
          </div>
        </AppCard>

        <AppCard v-if="selectedJob">
          <h2 class="font-bold">Import detail #{{ selectedJob.id }}</h2>
          <p class="mt-1 text-sm text-slate-500">{{ selectedJob.file_name }} · {{ selectedJob.status }}</p>
          <div class="mt-4 grid gap-3 sm:grid-cols-3">
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Total</p><p class="font-bold">{{ selectedJob.total_rows }}</p></div>
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Success</p><p class="font-bold">{{ selectedJob.success_rows }}</p></div>
            <div class="rounded-lg border border-slate-200 p-3"><p class="text-sm text-slate-500">Failed</p><p class="font-bold">{{ selectedJob.failed_rows }}</p></div>
          </div>
          <div class="mt-4 grid gap-3 md:grid-cols-2">
            <article v-for="row in selectedJob.rows" :key="row.id" class="rounded-lg border border-slate-200 p-3" :class="rowClass(row)">
              <div class="flex items-start justify-between gap-2">
                <div>
                  <h3 class="font-bold">{{ row.raw_data.name || '-' }}</h3>
                  <p class="text-sm text-slate-500">{{ row.raw_data.sku || '-' }} · row {{ row.row_index }}</p>
                </div>
                <span class="rounded-full bg-white px-2 py-1 text-xs font-bold">{{ row.status }}</span>
              </div>
              <p v-if="row.error_message" class="mt-2 text-sm text-red-700">{{ row.error_message }}</p>
            </article>
          </div>
        </AppCard>
      </div>
    </div>
  </section>
</template>
