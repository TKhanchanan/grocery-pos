<script setup lang="ts">
import { useAppStore } from '../stores/app'
import AppToast from './AppToast.vue'

const app = useAppStore()
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-x-3 top-3 z-[70] grid justify-items-end gap-2 sm:inset-x-auto sm:right-4 sm:top-4">
      <AppToast
        v-for="toast in app.toasts"
        :key="toast.id"
        :type="toast.type"
        :message="toast.message"
        :description="toast.description"
        @close="app.removeToast(toast.id)"
      />
    </div>
  </Teleport>
</template>
