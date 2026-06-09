<script setup lang="ts">
import type { IconName } from '../types/icons'
import AppIcon from './AppIcon.vue'

withDefaults(defineProps<{ title?: string; description?: string; icon?: IconName }>(), {
  icon: 'sparkles',
})
</script>

<template>
  <div class="rounded-2xl border border-dashed border-slate-300 bg-white/65 p-8 text-center dark:border-slate-600 dark:bg-slate-950/55">
    <div class="mx-auto grid h-12 w-12 place-items-center rounded-2xl bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-200">
      <AppIcon :name="icon" />
    </div>
    <h3 class="mt-4 text-lg font-bold text-slate-950 dark:text-slate-50">{{ title ?? 'No data yet' }}</h3>
    <p class="mx-auto mt-1 max-w-md text-sm leading-6 text-slate-500 dark:text-slate-300">{{ description ?? 'New records will appear here.' }}</p>
    <div v-if="$slots.default" class="mt-4"><slot /></div>
  </div>
</template>
