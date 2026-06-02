<script setup lang="ts">
defineProps<{ title: string; eyebrow?: string; description?: string }>()
</script>

<template>
  <header class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
    <div>
      <p v-if="eyebrow" class="text-xs font-bold uppercase text-brand-700">{{ eyebrow }}</p>
      <h1 class="mt-1 text-2xl font-bold tracking-tight">{{ title }}</h1>
      <p v-if="description" class="mt-1 max-w-2xl text-sm text-slate-500">{{ description }}</p>
    </div>
    <slot />
  </header>
</template>
