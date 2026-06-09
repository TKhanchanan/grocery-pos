<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { assetURL } from '../api/client'
import AppIcon from './AppIcon.vue'

const props = withDefaults(defineProps<{
  src?: string | null
  name?: string
  updatedAt?: string | null
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  shape?: 'rounded' | 'square'
  muted?: boolean
}>(), {
  src: '',
  name: '',
  updatedAt: '',
  size: 'md',
  shape: 'rounded',
  muted: false,
})

const failed = ref(false)

const sizeClass = computed(() => ({
  sm: 'h-10 w-10',
  md: 'h-12 w-12',
  lg: 'h-16 w-16',
  xl: 'h-24 w-24',
  full: 'h-full w-full',
}[props.size]))

const radiusClass = computed(() => props.shape === 'square' ? 'rounded-lg' : 'rounded-2xl')
const iconSize = computed(() => ({
  sm: 18,
  md: 20,
  lg: 24,
  xl: 34,
  full: 42,
}[props.size]))

const resolvedSrc = computed(() => {
  if (!props.src || failed.value) return ''
  const url = assetURL(props.src)
  if (!props.updatedAt || url.startsWith('blob:') || url.startsWith('data:')) return url
  const joiner = url.includes('?') ? '&' : '?'
  return `${url}${joiner}v=${encodeURIComponent(props.updatedAt)}`
})

watch(() => [props.src, props.updatedAt], () => {
  failed.value = false
})
</script>

<template>
  <div
    class="grid shrink-0 place-items-center overflow-hidden bg-brand-100 text-brand-700 ring-brand-200/70 dark:bg-emerald-500/15 dark:text-emerald-100 dark:ring-emerald-400/20"
    :class="[sizeClass, radiusClass, muted ? 'opacity-60 grayscale' : '']"
  >
    <img
      v-if="resolvedSrc"
      :src="resolvedSrc"
      :alt="name"
      class="h-full w-full object-cover"
      @error="failed = true"
    />
    <AppIcon v-else name="package" :size="iconSize" />
  </div>
</template>
