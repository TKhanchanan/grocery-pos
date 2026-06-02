<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import AppToastHost from './components/AppToastHost.vue'
import AppLayout from './layouts/AppLayout.vue'

const route = useRoute()
</script>

<template>
  <RouterView v-if="route.meta.layout === 'auth'" />
  <AppLayout v-else />
  <AppToastHost />
</template>
