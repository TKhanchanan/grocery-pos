<script setup lang="ts">
import { computed } from 'vue'
import AppIcon from './AppIcon.vue'

const props = withDefaults(defineProps<{ open: boolean; title?: string; description?: string; closeLabel?: string; size?: 'md' | 'lg' | 'xl'; hideClose?: boolean; centered?: boolean }>(), {
  size: 'md',
  hideClose: false,
  centered: false,
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
    <div v-if="open" class="fixed inset-0 z-[100] grid place-items-center bg-slate-950/45 p-4 backdrop-blur-sm dark:bg-slate-900/80">
      <div class="relative w-full" :class="sizeClass">
        <button
          v-if="!hideClose"
          class="absolute -right-3 -top-3 z-[120] grid h-8 w-8 place-items-center rounded-lg bg-white text-slate-500 shadow-lg shadow-slate-950/20 transition hover:bg-slate-50 hover:text-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:shadow-black/35 dark:hover:bg-slate-700"
          :aria-label="closeLabel ?? 'Close'"
          @click="$emit('close')"
        >
          <AppIcon name="x" :size="18" />
        </button>
        <section class="premium-surface max-h-[92vh] w-full overflow-y-auto rounded-2xl bg-white p-5 shadow-2xl dark:bg-slate-900 sm:p-6">
        <div class="flex items-start gap-3 pr-10" :class="centered ? 'justify-center text-center' : 'justify-between'">
          <div :class="centered ? 'mx-auto' : ''">
            <h2 class="text-lg font-bold">{{ title }}</h2>
            <p v-if="description" class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ description }}</p>
          </div>
        </div>
        <div class="mt-4"><slot /></div>
      </section>
      </div>
    </div>
  </Teleport>
</template>
