<script setup lang="ts">
import { ref } from 'vue'
import AppButton from './AppButton.vue'

withDefaults(defineProps<{ label?: string }>(), {
  label: 'Actions',
})

const open = ref(false)
</script>

<template>
  <div class="relative inline-block">
    <AppButton variant="secondary" @click="open = !open">{{ label }}</AppButton>
    <div v-if="open" class="premium-surface absolute right-0 z-30 mt-2 min-w-44 rounded-2xl border p-2 shadow-xl" @click="open = false">
      <slot />
    </div>
  </div>
</template>
