<script setup lang="ts">
import type { IconName } from '../types/icons'
import AnimatedNumber from './AnimatedNumber.vue'
import AppIcon from './AppIcon.vue'

withDefaults(defineProps<{
  label: string
  value?: number
  textValue?: string
  decimals?: number
  locale?: string
  suffix?: string
  helper?: string
  trend?: string
  icon: IconName
  tone?: 'brand' | 'success' | 'warning' | 'danger' | 'info'
}>(), {
  value: 0,
  decimals: 0,
  locale: 'th-TH',
  suffix: '',
  tone: 'brand',
})
</script>

<template>
  <article class="dashboard-kpi group relative overflow-hidden rounded-2xl bg-white/80 p-5 shadow-card transition duration-300 hover:-translate-y-1 hover:shadow-xl hover:shadow-brand-950/10 dark:bg-slate-900/80 dark:hover:shadow-black/30">
    <div class="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-brand-500 via-emerald-400 to-sky-400 opacity-80" />
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <p class="text-sm font-bold text-slate-500 dark:text-slate-400">{{ label }}</p>
        <p class="mt-3 break-words text-3xl font-black text-slate-950 dark:text-slate-50">
          <span v-if="textValue">{{ textValue }}</span>
          <AnimatedNumber v-else :value="value" :decimals="decimals" :locale="locale" :suffix="suffix" />
        </p>
      </div>
      <div
        class="grid h-12 w-12 shrink-0 place-items-center rounded-2xl shadow-sm transition group-hover:scale-105"
        :class="{
          'bg-brand-100 text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-200': tone === 'brand',
          'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-200': tone === 'success',
          'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-200': tone === 'warning',
          'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-200': tone === 'danger',
          'bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-200': tone === 'info',
        }"
      >
        <AppIcon :name="icon" />
      </div>
    </div>
    <div class="mt-4 flex flex-wrap items-center gap-2">
      <span v-if="trend" class="rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-black text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-200">{{ trend }}</span>
      <span v-if="helper" class="text-xs font-semibold text-slate-500 dark:text-slate-400">{{ helper }}</span>
    </div>
  </article>
</template>

<style scoped>
.dashboard-kpi {
  animation: dashboard-card-in 480ms ease both;
}

@keyframes dashboard-card-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
