<script setup lang="ts">
defineProps<{ message?: string; type?: 'success' | 'error' | 'info' }>()
</script>

<template>
  <div v-if="message" class="rounded-md border p-3 text-sm shadow-sm" :class="type === 'error' ? 'border-red-200 bg-red-50 text-red-700' : 'border-brand-100 bg-brand-50 text-brand-700'">
    {{ message }}
  </div>
</template>
