<script setup lang="ts">
import AppButton from './AppButton.vue'
import AppModal from './AppModal.vue'

withDefaults(defineProps<{
  open: boolean
  title?: string
  message?: string
  consequence?: string
  confirmLabel?: string
  cancelLabel?: string
  destructive?: boolean
  loading?: boolean
}>(), {
  destructive: true,
  loading: false,
})
defineEmits<{ close: []; confirm: [] }>()
</script>

<template>
  <AppModal :open="open" :title="title ?? 'Please confirm'" @close="$emit('close')">
    <p class="text-sm text-slate-600">{{ message ?? 'Are you sure you want to continue?' }}</p>
    <p v-if="consequence" class="mt-3 rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm font-semibold text-amber-800">{{ consequence }}</p>
    <div class="mt-3">
      <slot />
    </div>
    <div class="mt-5 flex justify-end gap-2">
      <AppButton variant="secondary" :disabled="loading" @click="$emit('close')">{{ cancelLabel ?? 'Cancel' }}</AppButton>
      <AppButton :variant="destructive ? 'danger' : 'primary'" :loading="loading" @click="$emit('confirm')">{{ confirmLabel ?? 'Confirm' }}</AppButton>
    </div>
  </AppModal>
</template>
