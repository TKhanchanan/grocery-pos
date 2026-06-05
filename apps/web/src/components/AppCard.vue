<script setup lang="ts">
withDefaults(defineProps<{ padded?: boolean; hover?: boolean }>(), {
  padded: true,
  hover: false,
})
</script>

<template>
  <section class="premium-surface rounded-[var(--radius-card)] shadow-card" :class="[padded ? 'p-5 md:p-6' : '', hover ? 'premium-card-hover' : '']">
    <slot />
  </section>
</template>
