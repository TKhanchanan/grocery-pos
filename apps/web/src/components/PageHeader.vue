<script setup lang="ts">
import type { IconName } from '../types/icons'

defineProps<{ title: string; eyebrow?: string; description?: string; icon?: IconName }>()
</script>

<template>
  <header class="mb-7 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
    <div class="min-w-0">
      <h1 class="text-3xl font-black tracking-tight text-slate-900 dark:text-slate-50 md:text-4xl">{{ title }}</h1>
      <p v-if="description" class="mt-2 max-w-3xl text-sm leading-6 text-slate-500 dark:text-slate-400 md:text-base">{{ description }}</p>
    </div>
    <div v-if="$slots.default" class="flex flex-wrap gap-2">
      <slot />
    </div>
  </header>
</template>
