<script setup lang="ts">
defineProps<{ label?: string; modelValue?: string | number }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <label class="grid gap-1 text-sm">
    <span v-if="label" class="font-semibold text-slate-700">{{ label }}</span>
    <select
      class="focus-ring min-h-10 rounded-md border border-slate-200 bg-white px-3 py-2"
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <slot />
    </select>
  </label>
</template>
