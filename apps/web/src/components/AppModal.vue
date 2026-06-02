<script setup lang="ts">
defineProps<{ open: boolean; title?: string }>()
defineEmits<{ close: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 grid place-items-center bg-slate-900/40 p-4">
      <section class="w-full max-w-lg rounded-lg bg-white p-4 shadow-xl">
        <div class="flex items-center justify-between gap-3">
          <h2 class="font-bold">{{ title }}</h2>
          <button class="rounded-md px-2 py-1 text-sm font-semibold" @click="$emit('close')">Close</button>
        </div>
        <div class="mt-4"><slot /></div>
      </section>
    </div>
  </Teleport>
</template>
