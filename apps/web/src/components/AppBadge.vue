<script setup lang="ts">
withDefaults(defineProps<{ tone?: 'brand' | 'success' | 'warning' | 'danger' | 'info' | 'neutral' }>(), {
  tone: 'brand',
})
</script>

<template>
  <span
    class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-bold"
    :class="{
      'bg-brand-100 text-brand-700': tone === 'brand' || tone === 'success',
      'bg-amber-100 text-amber-800': tone === 'warning',
      'bg-red-100 text-red-700': tone === 'danger',
      'bg-blue-100 text-blue-700': tone === 'info',
      'bg-slate-100 text-slate-600': tone === 'neutral',
    }"
  >
    <slot />
  </span>
</template>
