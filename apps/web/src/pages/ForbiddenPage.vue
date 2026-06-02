<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import PageHeader from '../components/PageHeader.vue'
</script>

<template>
  <section>
    <PageHeader title="Forbidden" eyebrow="403" description="Your role does not have permission to access this page." />
    <AppCard>
      <p class="text-sm text-slate-600">Ask an administrator if you need additional access.</p>
      <RouterLink to="/dashboard" class="mt-4 inline-flex">
        <AppButton variant="secondary">Back to dashboard</AppButton>
      </RouterLink>
    </AppCard>
  </section>
</template>
