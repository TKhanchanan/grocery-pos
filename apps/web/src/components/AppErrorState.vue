<script setup lang="ts">
import AppIcon from './AppIcon.vue'

defineProps<{ title?: string; message?: string }>()
</script>

<template>
  <div class="rounded-2xl border border-red-200 bg-red-50 p-5 text-sm text-red-700">
    <div class="flex items-start gap-3">
      <AppIcon name="circle-x" class="mt-0.5 shrink-0" />
      <div>
        <h3 class="font-bold">{{ title ?? 'Something went wrong' }}</h3>
        <p class="mt-1">{{ message ?? 'Please try again.' }}</p>
        <div v-if="$slots.default" class="mt-3"><slot /></div>
      </div>
    </div>
  </div>
</template>
