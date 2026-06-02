<script setup lang="ts">
defineProps<{
  loading?: boolean
  error?: string
  empty?: boolean
  emptyText?: string
}>()
</script>

<template>
  <div v-if="loading" class="panel text-sm text-slate-500">Loading...</div>
  <div v-else-if="error" class="panel border-red-200 bg-red-50 text-sm text-red-700">{{ error }}</div>
  <div v-else-if="empty" class="panel text-sm text-slate-500">{{ emptyText ?? 'No data yet.' }}</div>
  <slot v-else />
</template>
