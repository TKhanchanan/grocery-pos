<script setup lang="ts">
defineProps<{ columns: string[]; rows?: Record<string, string | number>[] }>()
</script>

<template>
  <div class="overflow-x-auto rounded-2xl border border-slate-200 bg-white/80 shadow-sm">
    <table class="min-w-full divide-y divide-slate-200 text-sm">
      <thead class="bg-slate-50">
        <tr>
          <th v-for="column in columns" :key="column" class="px-4 py-3 text-left font-bold text-slate-600">{{ column }}</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        <tr v-for="(row, index) in rows ?? []" :key="index">
          <td v-for="column in columns" :key="column" class="px-4 py-3">{{ row[column] ?? '-' }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
