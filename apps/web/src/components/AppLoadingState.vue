<script setup lang="ts">
withDefaults(defineProps<{ label?: string; rows?: number }>(), {
  label: 'Loading...',
  rows: 3,
})
</script>

<template>
  <div class="premium-surface rounded-2xl border p-5">
    <p class="text-sm font-semibold text-slate-500">{{ label }}</p>
    <div class="mt-4 grid gap-3">
      <div v-for="row in rows" :key="row" class="h-12 animate-pulse rounded-xl bg-slate-100" />
    </div>
  </div>
</template>
