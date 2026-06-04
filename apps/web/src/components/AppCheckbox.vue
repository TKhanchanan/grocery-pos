<script setup lang="ts">
defineProps<{ label: string; description?: string; modelValue?: boolean; disabled?: boolean }>()
defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>

<template>
  <label class="flex items-start gap-3 rounded-xl bg-white/75 p-3 text-sm shadow-sm transition hover:bg-brand-50 dark:bg-slate-950/60 dark:hover:bg-teal-400/10">
    <input
      class="focus-ring mt-1 h-4 w-4 rounded border-slate-300 text-brand-600"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled"
      @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
    />
    <span>
      <span class="block font-bold text-slate-700 dark:text-slate-100">{{ label }}</span>
      <span v-if="description" class="mt-0.5 block text-xs text-slate-500 dark:text-slate-400">{{ description }}</span>
    </span>
  </label>
</template>
