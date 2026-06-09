<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { assetURL } from '../api/client'

const props = withDefaults(
  defineProps<{
    src?: string | null
    name?: string
    size?: 'sm' | 'md' | 'lg' | 'xl'
  }>(),
  {
    size: 'md',
  },
)

const imageFailed = ref(false)

const initials = computed(() => {
  const value = props.name?.trim() || 'User'

  return value
    .split(/\s+/)
    .map((part) => part[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
})

const avatarSrc = computed(() => {
  if (!props.src || imageFailed.value) return ''

  return assetURL(props.src)
})

const sizeClass = computed(
  () =>
    ({
      sm: 'h-9 w-9 text-xs',
      md: 'h-11 w-11 text-sm',
      lg: 'h-16 w-16 text-lg',
      xl: 'h-24 w-24 text-2xl',
    })[props.size],
)

watch(
  () => props.src,
  () => {
    imageFailed.value = false
  },
)
</script>

<template>
  <span
    :class="[
      sizeClass,
      'relative inline-flex shrink-0 overflow-hidden rounded-full border border-brand-100 bg-brand-100 align-middle font-black text-brand-700 shadow-sm',
    ]"
  >
    <img
      v-if="avatarSrc"
      :src="avatarSrc"
      :alt="name || 'Profile avatar'"
      class="absolute inset-0 block h-full w-full rounded-full object-cover object-center"
      loading="lazy"
      decoding="async"
      @error="imageFailed = true"
    />

    <span
      v-else
      class="absolute inset-0 flex h-full w-full items-center justify-center rounded-full"
    >
      {{ initials }}
    </span>
  </span>
</template>