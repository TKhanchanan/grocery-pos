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
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { Category } from '../types/navigation'

const app = useAppStore()
const auth = useAuthStore()
const categories = ref<Category[]>([])
const loading = ref(false)
const saving = ref(false)
const statusSaving = ref(false)
const error = ref('')
const search = ref('')
const modalOpen = ref(false)
const nameError = ref('')
const pendingStatusCategory = ref<Category | null>(null)
const form = reactive({ id: 0, name: '', description: '' })

const canCreate = computed(() => auth.hasPermission('categories.create'))
const canUpdate = computed(() => auth.hasPermission('categories.update'))
const canDeactivate = computed(() => auth.hasPermission('categories.deactivate'))
const modalTitle = computed(() => form.id ? app.t('categories.editTitle') : app.t('categories.create'))
const filteredCategories = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) return categories.value
  return categories.value.filter((item) => `${item.name} ${item.description}`.toLowerCase().includes(query))
})
const pendingStatusNextActive = computed(() => pendingStatusCategory.value ? !pendingStatusCategory.value.is_active : false)
const pendingStatusMessage = computed(() => pendingStatusNextActive.value ? app.t('categories.confirmActivate') : app.t('categories.confirmDeactivate'))
const pendingStatusLabel = computed(() => pendingStatusNextActive.value ? app.t('categories.activate') : app.t('categories.deactivate'))

async function load() {
  loading.value = true
  error.value = ''
  try {
    categories.value = await apiClient<Category[]>('/v1/categories')
  } catch (err) {
    error.value = friendlyError(err, 'categories.loadFailed')
  } finally {
    loading.value = false
  }
}

function friendlyError(err: unknown, fallback: 'categories.loadFailed' | 'categories.saveFailed' | 'categories.statusFailed') {
  const message = err instanceof Error ? err.message : app.t(fallback)
  if (message.toLowerCase().includes('permission')) return app.t('categories.noPermission')
  return message
}

function reset() {
  Object.assign(form, { id: 0, name: '', description: '' })
  nameError.value = ''
}

function openCreate() {
  reset()
  modalOpen.value = true
}

function openEdit(item: Category) {
  Object.assign(form, { id: item.id, name: item.name, description: item.description })
  nameError.value = ''
  modalOpen.value = true
}

function closeModal() {
  if (saving.value) return
  modalOpen.value = false
  reset()
}

function validate() {
  nameError.value = form.name.trim() ? '' : app.t('categories.nameRequired')
  return !nameError.value
}

async function save() {
  if (!validate()) return
  saving.value = true
  error.value = ''
  try {
    const payload = { name: form.name.trim(), description: form.description.trim() }
    if (form.id) {
      await patchJSON<Category>(`/v1/categories/${form.id}`, payload)
      app.pushToast({ type: 'success', message: app.t('categories.updated') })
    } else {
      await postJSON<Category>('/v1/categories', payload)
      app.pushToast({ type: 'success', message: app.t('categories.created') })
    }
    await load()
    modalOpen.value = false
    reset()
  } catch (err) {
    error.value = friendlyError(err, 'categories.saveFailed')
    app.pushToast({ type: 'error', message: app.t('categories.saveFailed'), description: error.value })
  } finally {
    saving.value = false
  }
}

function requestStatusChange(item: Category) {
  pendingStatusCategory.value = item
}

function closeStatusDialog() {
  if (statusSaving.value) return
  pendingStatusCategory.value = null
}

async function confirmStatusChange() {
  const item = pendingStatusCategory.value
  if (!item) return
  statusSaving.value = true
  error.value = ''
  try {
    await patchJSON<Category>(`/v1/categories/${item.id}/status`, { is_active: !item.is_active })
    app.pushToast({ type: 'success', message: app.t('categories.statusUpdated') })
    pendingStatusCategory.value = null
    await load()
  } catch (err) {
    error.value = friendlyError(err, 'categories.statusFailed')
    app.pushToast({ type: 'error', message: app.t('categories.statusFailed'), description: error.value })
  } finally {
    statusSaving.value = false
  }
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader :title="app.t('categories.title')" :eyebrow="app.t('categories.eyebrow')" :description="app.t('categories.description')" icon="tags">
      <AppButton v-if="canCreate" icon="plus" @click="openCreate">{{ app.t('categories.add') }}</AppButton>
    </PageHeader>

    <div class="grid gap-4">
      <AppCard class="dark:bg-slate-900/80">
        <div class="grid gap-3">
          <AppInput v-model="search" :label="app.t('categories.search')" :placeholder="app.t('categories.searchPlaceholder')" />
        </div>
      </AppCard>

      <AppCard class="dark:bg-slate-900/80">
        <div v-if="error" class="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
        <AppLoadingState v-if="loading" :label="app.t('categories.loading')" />
        <AppEmptyState v-else-if="filteredCategories.length === 0" :title="app.t('categories.empty')" :description="app.t('categories.description')">
          <template v-if="canCreate && categories.length === 0">
            <AppButton icon="plus" @click="openCreate">{{ app.t('categories.addFirst') }}</AppButton>
          </template>
        </AppEmptyState>

        <div v-else>
          <div class="hidden overflow-x-auto md:block">
            <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
              <thead class="bg-slate-50 dark:bg-slate-950/70">
                <tr>
                  <th class="px-3 py-3 text-left font-black">{{ app.t('categories.name') }}</th>
                  <th class="px-3 py-3 text-left font-black">{{ app.t('categories.descriptionField') }}</th>
                  <th class="px-3 py-3 text-left font-black">{{ app.t('categories.status') }}</th>
                  <th class="px-3 py-3 text-right font-black">{{ app.t('categories.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                <tr v-for="item in filteredCategories" :key="item.id" class="hover:bg-slate-50/80 dark:hover:bg-slate-900/60">
                  <td class="px-3 py-3 font-semibold">{{ item.name }}</td>
                  <td class="px-3 py-3 text-slate-600 dark:text-slate-300">{{ item.description || app.t('categories.noDescription') }}</td>
                  <td class="px-3 py-3">
                    <AppBadge :tone="item.is_active ? 'success' : 'neutral'">{{ item.is_active ? app.t('categories.active') : app.t('categories.inactive') }}</AppBadge>
                  </td>
                  <td class="px-3 py-3">
                    <div class="flex flex-wrap justify-end gap-2">
                      <AppButton v-if="canUpdate" variant="secondary" icon="settings" @click="openEdit(item)">{{ app.t('categories.edit') }}</AppButton>
                      <AppButton v-if="canDeactivate" :variant="item.is_active ? 'danger' : 'secondary'" @click="requestStatusChange(item)">
                        {{ item.is_active ? app.t('categories.deactivate') : app.t('categories.activate') }}
                      </AppButton>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="grid gap-3 md:hidden">
            <article v-for="item in filteredCategories" :key="item.id" class="rounded-2xl border border-slate-200 bg-white/65 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-950/60">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <h2 class="truncate font-black">{{ item.name }}</h2>
                  <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ item.description || app.t('categories.noDescription') }}</p>
                </div>
                <AppBadge :tone="item.is_active ? 'success' : 'neutral'">{{ item.is_active ? app.t('categories.active') : app.t('categories.inactive') }}</AppBadge>
              </div>
              <div class="mt-4 flex flex-wrap gap-2">
                <AppButton v-if="canUpdate" variant="secondary" icon="settings" @click="openEdit(item)">{{ app.t('categories.edit') }}</AppButton>
                <AppButton v-if="canDeactivate" :variant="item.is_active ? 'danger' : 'secondary'" @click="requestStatusChange(item)">
                  {{ item.is_active ? app.t('categories.deactivate') : app.t('categories.activate') }}
                </AppButton>
              </div>
            </article>
          </div>
        </div>
      </AppCard>
    </div>

    <AppModal :open="modalOpen" :title="modalTitle" :description="app.t('categories.modalDescription')" :close-label="app.t('categories.cancel')" size="lg" @close="closeModal">
      <form class="grid gap-4" @submit.prevent="save">
        <div class="grid gap-3 rounded-2xl bg-slate-50/80 p-4 dark:bg-slate-950/45">
          <AppInput v-model="form.name" :label="app.t('categories.name')" :error="nameError" />
          <AppInput v-model="form.description" :label="app.t('categories.descriptionField')" />
        </div>
        <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <AppButton class="w-full sm:w-auto" type="button" variant="secondary" :disabled="saving" @click="closeModal">{{ app.t('categories.cancel') }}</AppButton>
          <AppButton class="w-full sm:w-auto" type="submit" :loading="saving" :disabled="saving || (!form.id && !canCreate) || (Boolean(form.id) && !canUpdate)" icon="check-circle">{{ app.t('categories.save') }}</AppButton>
        </div>
      </form>
    </AppModal>

    <ConfirmDialog
      :open="Boolean(pendingStatusCategory)"
      :title="app.t('categories.confirmTitle')"
      :message="pendingStatusMessage"
      :confirm-label="pendingStatusLabel"
      :cancel-label="app.t('categories.cancel')"
      :destructive="!pendingStatusNextActive"
      :loading="statusSaving"
      @close="closeStatusDialog"
      @confirm="confirmStatusChange"
    />
  </section>
</template>
