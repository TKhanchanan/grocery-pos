<script setup lang="ts">
import type { IconName } from '../types/icons'
import AppIcon from './AppIcon.vue'

withDefaults(defineProps<{ variant?: 'primary' | 'secondary' | 'danger' | 'ghost'; type?: 'button' | 'submit' | 'reset'; disabled?: boolean; loading?: boolean; icon?: IconName }>(), {
  variant: 'primary',
  type: 'button',
  disabled: false,
  loading: false,
})
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    class="focus-ring inline-flex min-h-11 items-center justify-center gap-2 rounded-xl px-4 py-2.5 text-sm font-bold shadow-none transition active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-75"
    :class="{
      'bg-brand-600 text-white hover:bg-brand-700 dark:bg-teal-300 dark:text-slate-950 dark:hover:bg-teal-200': variant === 'primary',
      'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700': variant === 'secondary',
      'bg-red-600 text-white hover:bg-red-700 dark:bg-red-500 dark:text-white dark:hover:bg-red-400': variant === 'danger',
      'text-slate-600 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800': variant === 'ghost',
    }"
  >
    <span v-if="loading" class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
    <AppIcon v-else-if="icon" :name="icon" :size="18" />
    <slot />
  </button>
</template>
