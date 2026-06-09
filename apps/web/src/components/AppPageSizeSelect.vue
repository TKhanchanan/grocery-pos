<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import AppIcon from './AppIcon.vue'

withDefaults(defineProps<{ modelValue: number; options?: number[] }>(), {
  options: () => [10, 20, 50],
})

defineEmits<{ 'update:modelValue': [value: number] }>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const popoverID = `page-size-${Math.random().toString(36).slice(2)}`

function setOpen(value: boolean) {
  if (value) window.dispatchEvent(new CustomEvent('app-popover-open', { detail: popoverID }))
  open.value = value
}

function onPopoverOpen(event: Event) {
  if ((event as CustomEvent<string>).detail !== popoverID) open.value = false
}

function onDocumentClick(event: MouseEvent) {
  if (!root.value?.contains(event.target as Node)) open.value = false
}

onMounted(() => {
  document.addEventListener('mousedown', onDocumentClick)
  window.addEventListener('app-popover-open', onPopoverOpen)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocumentClick)
  window.removeEventListener('app-popover-open', onPopoverOpen)
})
</script>

<template>
  <div ref="root" class="relative inline-block">
    <button
      type="button"
      class="inline-flex min-h-10 min-w-20 items-center justify-between gap-2 rounded-xl border border-slate-300 bg-white px-3 text-sm font-black text-slate-700 shadow-none hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
      :aria-expanded="open"
      @click="setOpen(!open)"
    >
      {{ modelValue }}
      <AppIcon name="chevron-down" :size="16" />
    </button>
    <div v-if="open" class="absolute left-0 z-[80] w-full overflow-hidden rounded-xl bg-white py-1 shadow-xl shadow-slate-950/15 dark:bg-slate-900 dark:shadow-black/30">
      <button
        v-for="option in options"
        :key="option"
        type="button"
        class="block min-h-10 w-full px-3 text-left text-sm font-black transition hover:bg-slate-100 dark:hover:bg-slate-800"
        :class="modelValue === option ? 'text-brand-700 dark:text-teal-200' : 'text-slate-600 dark:text-slate-300'"
        @click="$emit('update:modelValue', option); open = false"
      >
        {{ option }}
      </button>
    </div>
  </div>
</template>
