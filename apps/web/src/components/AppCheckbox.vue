<script setup lang="ts">
defineProps<{ label: string; description?: string; modelValue?: boolean; disabled?: boolean }>()
defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>

<template>
  <label class="flex items-start gap-3 rounded-xl border border-slate-200 bg-white/70 p-3 text-sm shadow-sm transition hover:bg-slate-50">
    <input
      class="focus-ring mt-1 h-4 w-4 rounded border-slate-300 text-brand-600"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled"
      @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
    />
    <span>
      <span class="block font-bold text-slate-700">{{ label }}</span>
      <span v-if="description" class="mt-0.5 block text-xs text-slate-500">{{ description }}</span>
    </span>
  </label>
</template>
