<script setup lang="ts">
import AppButton from './AppButton.vue'
import AppModal from './AppModal.vue'

defineProps<{ open: boolean; title?: string; message?: string }>()
defineEmits<{ close: []; confirm: [] }>()
</script>

<template>
  <AppModal :open="open" :title="title ?? 'Please confirm'" @close="$emit('close')">
    <p class="text-sm text-slate-600">{{ message ?? 'Are you sure you want to continue?' }}</p>
    <div class="mt-5 flex justify-end gap-2">
      <AppButton variant="secondary" @click="$emit('close')">Cancel</AppButton>
      <AppButton variant="danger" @click="$emit('confirm')">Confirm</AppButton>
    </div>
  </AppModal>
</template>
