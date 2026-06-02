<script setup lang="ts">
import AppCard from './AppCard.vue'

defineProps<{ label: string; value: string | number; helper?: string }>()
</script>

<template>
  <AppCard>
    <p class="text-sm font-medium text-slate-500">{{ label }}</p>
    <p class="mt-2 text-2xl font-bold text-brand-700">{{ value }}</p>
    <p v-if="helper" class="mt-1 text-xs text-slate-500">{{ helper }}</p>
  </AppCard>
</template>
