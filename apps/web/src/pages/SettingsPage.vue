<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAppStore } from '../stores/app'

const app = useAppStore()
</script>

<template>
  <section>
    <PageHeader :title="app.t('settings.title')" :eyebrow="app.t('settings.eyebrow')" :description="app.t('settings.description')" />

    <div class="grid gap-4 lg:grid-cols-[minmax(0,520px)_1fr]">
      <AppCard>
        <div class="grid gap-5">
          <div>
            <h2 class="font-bold">{{ app.t('settings.appearance') }}</h2>
            <p class="mt-1 text-sm text-slate-500">{{ app.t('settings.appearanceDescription') }}</p>
          </div>

          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700">{{ app.t('settings.language') }}</p>
            <div class="grid gap-2 sm:grid-cols-2">
              <AppButton :variant="app.language === 'th' ? 'primary' : 'secondary'" @click="app.setLanguage('th')">
                {{ app.t('settings.thai') }}
              </AppButton>
              <AppButton :variant="app.language === 'en' ? 'primary' : 'secondary'" @click="app.setLanguage('en')">
                {{ app.t('settings.english') }}
              </AppButton>
            </div>
          </div>

          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700">{{ app.t('settings.theme') }}</p>
            <div class="grid gap-2 sm:grid-cols-2">
              <AppButton :variant="app.theme === 'light' ? 'primary' : 'secondary'" @click="app.setTheme('light')">
                {{ app.t('settings.light') }}
              </AppButton>
              <AppButton :variant="app.theme === 'dark' ? 'primary' : 'secondary'" @click="app.setTheme('dark')">
                {{ app.t('settings.dark') }}
              </AppButton>
            </div>
          </div>
        </div>
      </AppCard>

      <AppCard>
        <div class="grid gap-4">
          <div>
            <p class="text-xs font-bold uppercase text-brand-700">{{ app.t('settings.preview') }}</p>
            <h2 class="mt-1 text-xl font-bold">{{ app.t('app.name') }}</h2>
            <p class="mt-1 text-sm text-slate-500">{{ app.t('settings.previewText') }}</p>
          </div>
          <div class="grid gap-3 sm:grid-cols-3">
            <div class="rounded-lg border border-slate-200 bg-white p-3">
              <p class="text-xs text-slate-500">{{ app.t('nav.products') }}</p>
              <p class="mt-1 text-2xl font-bold text-brand-700">128</p>
            </div>
            <div class="rounded-lg border border-slate-200 bg-white p-3">
              <p class="text-xs text-slate-500">{{ app.t('nav.alerts') }}</p>
              <p class="mt-1 text-2xl font-bold text-brand-700">{{ app.alertCount }}</p>
            </div>
            <div class="rounded-lg border border-slate-200 bg-white p-3">
              <p class="text-xs text-slate-500">{{ app.t('settings.theme') }}</p>
              <p class="mt-1 text-2xl font-bold text-brand-700">{{ app.isDark ? app.t('settings.dark') : app.t('settings.light') }}</p>
            </div>
          </div>
        </div>
      </AppCard>
    </div>
  </section>
</template>
