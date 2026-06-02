<script setup lang="ts" generic="T extends string">
defineProps<{ tabs: Array<{ key: T; label: string }>; modelValue: T }>()
defineEmits<{ 'update:modelValue': [value: T] }>()
</script>

<template>
  <div class="inline-flex max-w-full gap-1 overflow-x-auto rounded-2xl border border-slate-200 bg-white/70 p-1 shadow-sm">
    <button
      v-for="tab in tabs"
      :key="tab.key"
      class="focus-ring whitespace-nowrap rounded-xl px-4 py-2 text-sm font-bold transition"
      :class="modelValue === tab.key ? 'bg-brand-600 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-100'"
      @click="$emit('update:modelValue', tab.key)"
    >
      {{ tab.label }}
    </button>
  </div>
</template>
