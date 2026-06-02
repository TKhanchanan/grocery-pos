<script setup lang="ts">
withDefaults(defineProps<{ variant?: 'primary' | 'secondary' | 'danger'; type?: 'button' | 'submit' | 'reset' }>(), {
  variant: 'primary',
  type: 'button',
})
</script>

<template>
  <button
    :type="type"
    class="inline-flex min-h-10 items-center justify-center rounded-md px-4 py-2 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-50"
    :class="{
      'bg-brand-600 text-white hover:bg-brand-700': variant === 'primary',
      'border border-slate-200 bg-white text-slate-700 hover:bg-slate-50': variant === 'secondary',
      'bg-red-600 text-white hover:bg-red-700': variant === 'danger',
    }"
  >
    <slot />
  </button>
</template>
