<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAppStore } from '../stores/app'
import AppButton from './AppButton.vue'
import AppModal from './AppModal.vue'
import AppToast from './AppToast.vue'

const app = useAppStore()
const dismissedResultID = ref(0)
const resultToast = computed(() => [...app.toasts].reverse().find((toast) => (toast.type === 'success' || toast.type === 'error') && toast.resultModal !== false && toast.id !== dismissedResultID.value) ?? null)

function closeResultModal() {
  if (!resultToast.value) return
  dismissedResultID.value = resultToast.value.id
  app.removeToast(resultToast.value.id)
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-x-3 top-3 z-[70] grid justify-items-end gap-2 sm:inset-x-auto sm:right-4 sm:top-4">
      <AppToast
        v-for="toast in app.toasts"
        :key="toast.id"
        :type="toast.type"
        :message="toast.message"
        :description="toast.description"
        @close="app.removeToast(toast.id)"
      />
    </div>

    <AppModal :open="Boolean(resultToast)" :title="resultToast?.message" centered hide-close @close="closeResultModal">
      <div class="text-center">
        <div
          class="mx-auto grid h-12 w-12 place-items-center rounded-2xl"
          :class="resultToast?.type === 'error' ? 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-100' : 'bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-100'"
        >
          <span class="text-2xl font-black">{{ resultToast?.type === 'error' ? '!' : '✓' }}</span>
        </div>
        <p v-if="resultToast?.description" class="mx-auto mt-3 max-w-sm text-sm text-slate-600 dark:text-slate-300">{{ resultToast.description }}</p>
        <div class="mt-5 flex justify-center">
          <AppButton :variant="resultToast?.type === 'error' ? 'danger' : 'primary'" @click="closeResultModal">ตกลง</AppButton>
        </div>
      </div>
    </AppModal>
  </Teleport>
</template>
