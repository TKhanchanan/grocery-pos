<script setup lang="ts">
defineProps<{ label?: string; placeholder?: string; modelValue?: string; helper?: string; error?: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <label class="grid gap-1.5 text-sm">
    <span v-if="label" class="font-semibold text-slate-700 dark:text-slate-200">{{ label }}</span>
    <textarea
      class="focus-ring min-h-28 rounded-xl bg-white/90 px-3.5 py-2.5 text-slate-950 shadow-sm transition placeholder:text-slate-400 dark:bg-slate-950/80 dark:text-slate-50 dark:placeholder:text-slate-500"
      :placeholder="placeholder"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
    <span v-if="error" class="text-xs font-semibold text-red-600 dark:text-red-300">{{ error }}</span>
    <span v-else-if="helper" class="text-xs text-slate-500 dark:text-slate-400">{{ helper }}</span>
  </label>
</template>
