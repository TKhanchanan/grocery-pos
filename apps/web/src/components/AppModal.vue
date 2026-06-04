<script setup lang="ts">
import { computed } from 'vue'
import AppIcon from './AppIcon.vue'

const props = withDefaults(defineProps<{ open: boolean; title?: string; description?: string; closeLabel?: string; size?: 'md' | 'lg' | 'xl' }>(), {
  size: 'md',
})
defineEmits<{ close: [] }>()

const sizeClass = computed(() => ({
  md: 'max-w-lg',
  lg: 'max-w-2xl',
  xl: 'max-w-3xl',
}[props.size]))
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 grid place-items-center bg-slate-950/45 p-4 backdrop-blur-sm">
      <section class="premium-surface max-h-[92vh] w-full overflow-y-auto rounded-2xl p-5 shadow-2xl sm:p-6" :class="sizeClass">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-bold">{{ title }}</h2>
            <p v-if="description" class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ description }}</p>
          </div>
          <button class="focus-ring rounded-xl p-2 text-slate-500 hover:bg-brand-50 dark:text-slate-300 dark:hover:bg-teal-400/10" :aria-label="closeLabel ?? 'Close'" @click="$emit('close')">
            <AppIcon name="x" />
          </button>
        </div>
        <div class="mt-4"><slot /></div>
      </section>
    </div>
  </Teleport>
</template>
