<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  value: number
  decimals?: number
  locale?: string
  prefix?: string
  suffix?: string
  duration?: number
}>(), {
  decimals: 0,
  locale: 'th-TH',
  prefix: '',
  suffix: '',
  duration: 850,
})

const display = ref(format(props.value))
let frame = 0

function format(value: number) {
  return `${props.prefix}${value.toLocaleString(props.locale, {
    minimumFractionDigits: props.decimals,
    maximumFractionDigits: props.decimals,
  })}${props.suffix}`
}

watch(() => props.value, (next, previous = 0) => {
  window.cancelAnimationFrame(frame)
  const start = Number(previous) || 0
  const delta = next - start
  const startedAt = performance.now()
  const tick = (now: number) => {
    const progress = Math.min((now - startedAt) / props.duration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    display.value = format(start + delta * eased)
    if (progress < 1) frame = window.requestAnimationFrame(tick)
  }
  frame = window.requestAnimationFrame(tick)
}, { immediate: true })

onBeforeUnmount(() => window.cancelAnimationFrame(frame))
</script>

<template>
  <span>{{ display }}</span>
</template>
