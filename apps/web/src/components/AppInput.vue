<script setup lang="ts">
defineProps<{ label?: string; placeholder?: string; modelValue?: string | number; type?: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <label class="grid gap-1 text-sm">
    <span v-if="label" class="font-semibold text-slate-700">{{ label }}</span>
    <input
      class="focus-ring min-h-10 rounded-md border border-slate-200 bg-white px-3 py-2"
      :type="type ?? 'text'"
      :placeholder="placeholder"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </label>
</template>
