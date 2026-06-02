<script setup lang="ts">
import AppIcon from './AppIcon.vue'

defineProps<{ open: boolean; title?: string; description?: string; closeLabel?: string }>()
defineEmits<{ close: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 grid place-items-center bg-slate-950/45 p-4 backdrop-blur-sm">
      <section class="premium-surface max-h-[92vh] w-full max-w-lg overflow-y-auto rounded-2xl border p-5 shadow-2xl">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-bold">{{ title }}</h2>
            <p v-if="description" class="mt-1 text-sm text-slate-500">{{ description }}</p>
          </div>
          <button class="focus-ring rounded-xl p-2 text-slate-500 hover:bg-slate-100" :aria-label="closeLabel ?? 'Close'" @click="$emit('close')">
            <AppIcon name="x" />
          </button>
        </div>
        <div class="mt-4"><slot /></div>
      </section>
    </div>
  </Teleport>
</template>
