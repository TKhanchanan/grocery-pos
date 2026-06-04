<script setup lang="ts" generic="T extends string">
defineProps<{ tabs: Array<{ key: T; label: string }>; modelValue: T }>()
defineEmits<{ 'update:modelValue': [value: T] }>()
</script>

<template>
  <div class="max-w-full gap-1 overflow-x-auto rounded-xl">
    <button
      v-for="tab in tabs"
      :key="tab.key"
      class="focus-ring whitespace-nowrap rounded-xl px-4 py-2 text-sm font-bold transition"
      :class="modelValue === tab.key ? 'bg-brand-600 text-white shadow-sm dark:bg-teal-300 dark:text-slate-950' : 'text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800'"
      @click="$emit('update:modelValue', tab.key)"
    >
      {{ tab.label }}
    </button>
  </div>
</template>
