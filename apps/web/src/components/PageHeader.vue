<script setup lang="ts">
import type { IconName } from '../types/icons'
import AppIcon from './AppIcon.vue'

defineProps<{ title: string; eyebrow?: string; description?: string; icon?: IconName }>()
</script>

<template>
  <header class="mb-7 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
    <div class="flex min-w-0 items-start gap-4">
      <div v-if="icon" class="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-brand-100 text-brand-700 shadow-sm">
        <AppIcon :name="icon" />
      </div>
      <div class="min-w-0">
        <p v-if="eyebrow" class="text-xs font-bold uppercase tracking-wide text-brand-700">{{ eyebrow }}</p>
        <h1 class="mt-1 text-3xl font-black tracking-tight text-slate-900 md:text-4xl">{{ title }}</h1>
        <p v-if="description" class="mt-2 max-w-3xl text-sm leading-6 text-slate-500 md:text-base">{{ description }}</p>
      </div>
    </div>
    <div v-if="$slots.default" class="flex flex-wrap gap-2">
      <slot />
    </div>
  </header>
</template>
