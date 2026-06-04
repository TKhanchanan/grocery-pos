<script setup lang="ts">
defineProps<{ label?: string; modelValue?: string | number; helper?: string; error?: string; hideArrow?: boolean }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <label class="grid gap-1.5 text-sm">
    <span v-if="label" class="font-semibold text-slate-700 dark:text-slate-200">{{ label }}</span>
    <select
      class="focus-ring min-h-11 rounded-xl bg-white/90 px-3.5 py-2.5 text-slate-950 shadow-sm transition dark:bg-slate-950/80 dark:text-slate-50"
      :class="hideArrow ? 'no-native-select-arrow appearance-none bg-none pr-4' : ''"
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <slot />
    </select>
    <span v-if="error" class="text-xs font-semibold text-red-600 dark:text-red-300">{{ error }}</span>
    <span v-else-if="helper" class="text-xs text-slate-500 dark:text-slate-400">{{ helper }}</span>
  </label>
</template>

<style scoped>
.no-native-select-arrow {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image: none;
}
</style>
