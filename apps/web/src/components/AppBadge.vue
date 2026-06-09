<script setup lang="ts">
withDefaults(defineProps<{ tone?: 'brand' | 'success' | 'warning' | 'danger' | 'info' | 'neutral' }>(), {
  tone: 'brand',
})
</script>

<template>
  <span
    class="inline-flex shrink-0 items-center whitespace-nowrap rounded-full px-2.5 py-1 text-xs font-bold"
    :class="{
      'bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-100': tone === 'brand' || tone === 'success',
      'bg-amber-100 text-amber-800 dark:bg-amber-500/20 dark:text-amber-100': tone === 'warning',
      'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-100': tone === 'danger',
      'bg-blue-100 text-blue-700 dark:bg-sky-500/20 dark:text-sky-100': tone === 'info',
      'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-100': tone === 'neutral',
    }"
  >
    <slot />
  </span>
</template>
