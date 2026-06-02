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
    class="focus-ring inline-flex min-h-11 items-center justify-center gap-2 rounded-xl px-4 py-2.5 text-sm font-bold shadow-sm transition active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-55"
    :class="{
      'bg-brand-600 text-white shadow-brand-600/20 hover:bg-brand-700': variant === 'primary',
      'border border-slate-200 bg-white/80 text-slate-700 hover:bg-slate-50': variant === 'secondary',
      'bg-red-600 text-white shadow-red-600/20 hover:bg-red-700': variant === 'danger',
      'text-slate-600 shadow-none hover:bg-slate-100': variant === 'ghost',
    }"
  >
    <span v-if="loading" class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
    <AppIcon v-else-if="icon" :name="icon" :size="18" />
    <slot />
  </button>
</template>
