<script setup lang="ts">
defineProps<{ open: boolean; title?: string }>()
defineEmits<{ close: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-40 lg:hidden">
      <button class="absolute inset-0 bg-slate-900/40" aria-label="Close drawer" @click="$emit('close')" />
      <aside class="absolute left-0 top-0 h-full w-80 max-w-[86vw] overflow-y-auto bg-white p-4 shadow-xl">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="font-bold">{{ title }}</h2>
          <button class="rounded-md px-2 py-1 text-sm font-semibold" @click="$emit('close')">Close</button>
        </div>
        <slot />
      </aside>
    </div>
  </Teleport>
</template>
