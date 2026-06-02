<script setup lang="ts">
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const links = [
  ['/', 'Dashboard'],
  ['/products', 'Products'],
  ['/inventory', 'Inventory'],
  ['/pos', 'POS'],
  ['/sales', 'Sales'],
  ['/reports', 'Reports'],
  ['/suppliers', 'Suppliers & PO'],
  ['/settings', 'Settings'],
] as const

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <RouterView v-if="route.path === '/login'" />
  <div v-else class="min-h-screen">
    <header class="sticky top-0 z-20 border-b border-emerald-100 bg-white/95 backdrop-blur">
      <div class="mx-auto flex max-w-7xl flex-col gap-3 px-4 py-3 md:flex-row md:items-center md:justify-between">
        <div>
          <p class="label">Small grocery operations</p>
          <h1 class="text-xl font-bold text-leaf">Grocery POS & Inventory</h1>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <span class="rounded-full bg-mint px-3 py-1 text-xs font-semibold text-leaf">
            {{ auth.user?.username }} · {{ auth.user?.role }}
          </span>
          <button class="btn-soft" @click="logout">Logout</button>
        </div>
      </div>
      <nav class="mx-auto flex max-w-7xl gap-2 overflow-x-auto px-4 pb-3">
        <RouterLink
          v-for="[href, label] in links"
          :key="href"
          :to="href"
          class="whitespace-nowrap rounded-md px-3 py-2 text-sm font-semibold text-slate-600 hover:bg-mint hover:text-leaf"
          :class="{ 'bg-leaf text-white hover:bg-leaf hover:text-white': route.path === href }"
        >
          {{ label }}
        </RouterLink>
      </nav>
    </header>
    <main class="mx-auto max-w-7xl px-4 py-6">
      <RouterView />
    </main>
  </div>
</template>
