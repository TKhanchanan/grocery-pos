<script setup lang="ts">
import type { IconName } from '../types/icons'
import AppIcon from './AppIcon.vue'

const props = withDefaults(defineProps<{ message?: string; description?: string; type?: 'success' | 'error' | 'warning' | 'info' }>(), {
  type: 'success',
})
defineEmits<{ close: [] }>()

const iconMap: Record<string, IconName> = {
  success: 'check-circle',
  error: 'circle-x',
  warning: 'triangle-alert',
  info: 'info',
}
</script>

<template>
  <div
    v-if="message"
    class="premium-surface flex w-full max-w-sm items-start gap-3 rounded-2xl border p-4 text-sm shadow-xl"
    :class="{
      'text-brand-700': type === 'success',
      'text-red-700': type === 'error',
      'text-amber-800': type === 'warning',
      'text-blue-700': type === 'info',
    }"
  >
    <AppIcon :name="iconMap[props.type]" class="mt-0.5 shrink-0" />
    <div class="min-w-0 flex-1">
      <p class="font-bold">{{ message }}</p>
      <p v-if="description" class="mt-1 text-slate-500">{{ description }}</p>
    </div>
    <button class="rounded-lg px-2 text-slate-500 hover:bg-slate-100" aria-label="Close notification" @click="$emit('close')">×</button>
  </div>
</template>
