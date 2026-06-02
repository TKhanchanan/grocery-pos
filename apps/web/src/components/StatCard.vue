<script setup lang="ts">
import AppCard from './AppCard.vue'
import type { IconName } from '../types/icons'
import AppIcon from './AppIcon.vue'

defineProps<{ label: string; value: string | number; helper?: string; icon?: IconName; tone?: 'brand' | 'success' | 'warning' | 'danger' | 'info' }>()
</script>

<template>
  <AppCard hover>
    <div class="flex items-start justify-between gap-3">
      <p class="text-sm font-bold text-slate-500">{{ label }}</p>
      <div v-if="icon" class="grid h-10 w-10 place-items-center rounded-2xl bg-brand-100 text-brand-700">
        <AppIcon :name="icon" />
      </div>
    </div>
    <p class="mt-3 text-3xl font-black tracking-tight text-brand-700">{{ value }}</p>
    <p v-if="helper" class="mt-1 text-xs text-slate-500">{{ helper }}</p>
  </AppCard>
</template>
