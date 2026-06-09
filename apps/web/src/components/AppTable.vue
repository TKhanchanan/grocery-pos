<script setup lang="ts">
defineProps<{ columns: string[]; rows?: Record<string, string | number>[] }>()
</script>

<template>
  <div class="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-none dark:border-slate-700 dark:bg-slate-900/80">
    <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-700">
      <thead class="bg-slate-50 dark:bg-slate-900">
        <tr>
          <th v-for="column in columns" :key="column" class="px-4 py-3 text-left font-bold text-slate-600 dark:text-slate-200">{{ column }}</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
        <tr v-for="(row, index) in rows ?? []" :key="index">
          <td v-for="column in columns" :key="column" class="px-4 py-3">{{ row[column] ?? '-' }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
