<script setup lang="ts">
defineProps<{ title?: string; description?: string }>()
</script>

<template>
  <div class="rounded-lg border border-dashed border-slate-300 bg-white p-8 text-center">
    <h3 class="font-bold">{{ title ?? 'No data yet' }}</h3>
    <p class="mt-1 text-sm text-slate-500">{{ description ?? 'New records will appear here.' }}</p>
  </div>
</template>
