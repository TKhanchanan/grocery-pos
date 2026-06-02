<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import AppCard from '../components/AppCard.vue'
import AppEmptyState from '../components/AppEmptyState.vue'
</script>

<template>
  <section>
    <PageHeader title="Dashboard" eyebrow="Overview" description="Daily sales, stock alerts, profit, and operational summaries will live here." />
    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard label="Today sales" value="0.00" helper="Waiting for POS data" />
      <StatCard label="Open alerts" value="0" helper="Low stock and reorder alerts" />
      <StatCard label="Products" value="0" helper="Catalog count placeholder" />
      <StatCard label="Inventory value" value="0.00" helper="Cost valuation placeholder" />
    </div>
    <AppCard class="mt-4">
      <h2 class="font-bold">Activity feed</h2>
      <AppEmptyState class="mt-4" title="No activity yet" description="Sales, transfers, restocks, and purchase order events will appear here." />
    </AppCard>
  </section>
</template>
