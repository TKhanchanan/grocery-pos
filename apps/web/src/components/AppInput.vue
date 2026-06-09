<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{ label?: string; placeholder?: string; modelValue?: string | number; type?: string; disabled?: boolean; helper?: string; error?: string; min?: string | number; max?: string | number; step?: string | number }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const inputRef = ref<HTMLInputElement | null>(null)
const inputType = computed(() => props.type ?? 'text')

function focus() {
  inputRef.value?.focus()
}

function updateValue(value: string) {
  if (inputType.value === 'number' && value !== '' && props.min !== undefined && Number(value) < Number(props.min)) {
    emit('update:modelValue', String(props.min))
    return
  }
  emit('update:modelValue', value)
}

defineExpose({ focus })
</script>

<template>
  <label class="grid gap-1.5 text-sm">
    <span v-if="label" class="font-semibold text-slate-700 dark:text-slate-200">{{ label }}</span>
    <input
      ref="inputRef"
      class="focus-ring min-h-11 rounded-xl border border-slate-300 bg-white px-3.5 py-2.5 text-slate-950 shadow-none transition placeholder:text-slate-400 disabled:bg-slate-100 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-50 dark:placeholder:text-slate-500 dark:disabled:bg-slate-800"
      :type="inputType"
      :placeholder="placeholder"
      :value="modelValue"
      :disabled="disabled"
      :min="min"
      :max="max"
      :step="step"
      @input="updateValue(($event.target as HTMLInputElement).value)"
    />
    <span v-if="error" class="text-xs font-semibold text-red-600 dark:text-red-300">{{ error }}</span>
    <span v-else-if="helper" class="text-xs text-slate-500 dark:text-slate-400">{{ helper }}</span>
  </label>
</template>
