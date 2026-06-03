<script setup lang="ts">
defineProps<{ open: boolean; title?: string; closeLabel?: string }>()
defineEmits<{ close: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-40">
      <button class="absolute inset-0 bg-slate-950/45 backdrop-blur-sm" :aria-label="closeLabel ?? 'Close drawer'" @click="$emit('close')" />
      <aside class="premium-surface absolute left-0 top-0 h-full w-80 max-w-[88vw] overflow-y-auto border-r p-4 shadow-2xl">
        <!-- <div class="mb-4 flex items-center justify-between">
          <h2 class="font-bold">{{ title }}</h2>
          <button class="focus-ring rounded-xl px-3 py-2 text-sm font-bold hover:bg-slate-100" @click="$emit('close')">{{ closeLabel ?? 'Close' }}</button>
        </div> -->
        <slot />
      </aside>
    </div>
  </Teleport>
</template>
