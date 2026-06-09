<script setup lang="ts">
import AppBadge from '../components/AppBadge.vue'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
import PageHeader from '../components/PageHeader.vue'

defineProps<{
  title: string
  description: string
  moduleName?: string
}>()
</script>

<template>
  <section>
    <PageHeader :title="title" eyebrow="Module foundation" :description="description">
      <AppButton variant="secondary">Primary action placeholder</AppButton>
    </PageHeader>

    <div class="grid gap-4 lg:grid-cols-[1fr_320px]">
      <AppCard>
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="font-bold">{{ moduleName ?? title }} workspace</h2>
            <p class="mt-1 text-sm text-slate-500">API integration and business workflows will attach here in the next implementation prompts.</p>
          </div>
          <AppBadge>V1 shell</AppBadge>
        </div>
        <AppEmptyState class="mt-4" title="No records loaded" description="This placeholder is ready for forms, tables, filters, and detail panels." />
      </AppCard>

      <AppCard>
        <h3 class="font-bold">Responsive notes</h3>
        <ul class="mt-3 grid gap-2 text-sm text-slate-600">
          <li>Desktop: sidebar navigation stays visible.</li>
          <li>Mobile: drawer navigation opens from the topbar.</li>
          <li>Tables can swap to mobile cards when data is added.</li>
        </ul>
      </AppCard>
    </div>
  </section>
</template>
