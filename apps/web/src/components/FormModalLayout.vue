<script setup lang="ts">
import AppButton from './AppButton.vue'

withDefaults(defineProps<{
  title: string
  description?: string
  submitLabel?: string
  cancelLabel?: string
  loading?: boolean
}>(), {
  submitLabel: 'Save',
  cancelLabel: 'Cancel',
  loading: false,
})

defineEmits<{ cancel: []; submit: [] }>()
</script>

<template>
  <form class="grid gap-5" @submit.prevent="$emit('submit')">
    <div>
      <h2 class="text-lg font-bold">{{ title }}</h2>
      <p v-if="description" class="mt-1 text-sm text-slate-500">{{ description }}</p>
    </div>
    <div class="grid gap-4">
      <slot />
    </div>
    <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
      <AppButton variant="secondary" :disabled="loading" @click="$emit('cancel')">{{ cancelLabel }}</AppButton>
      <AppButton type="submit" :loading="loading">{{ submitLabel }}</AppButton>
    </div>
  </form>
</template>
