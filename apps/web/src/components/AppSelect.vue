<script setup lang="ts">
defineProps<{ label?: string; modelValue?: string | number; helper?: string; error?: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <label class="grid gap-1.5 text-sm">
    <span v-if="label" class="font-semibold text-slate-700">{{ label }}</span>
    <select
      class="focus-ring min-h-11 rounded-xl border border-slate-200 bg-white/85 px-3.5 py-2.5 shadow-sm transition"
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <slot />
    </select>
    <span v-if="error" class="text-xs font-semibold text-red-600">{{ error }}</span>
    <span v-else-if="helper" class="text-xs text-slate-500">{{ helper }}</span>
  </label>
</template>
