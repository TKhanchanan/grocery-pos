<script setup lang="ts">
import type { IconName } from '../types/icons'

withDefaults(defineProps<{ name: IconName; size?: number; strokeWidth?: number }>(), {
  size: 20,
  strokeWidth: 2,
})
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    :stroke-width="strokeWidth"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <template v-if="name === 'layout-dashboard'"><rect x="3" y="3" width="7" height="9" rx="1" /><rect x="14" y="3" width="7" height="5" rx="1" /><rect x="14" y="12" width="7" height="9" rx="1" /><rect x="3" y="16" width="7" height="5" rx="1" /></template>
    <template v-else-if="name === 'shopping-cart'"><circle cx="8" cy="21" r="1" /><circle cx="19" cy="21" r="1" /><path d="M2 3h3l3 13h11l2-9H7" /></template>
    <template v-else-if="name === 'package'">
      <path d="M4 7.5 12 3l8 4.5-8 4.5L4 7.5Z" />
      <path d="M4 7.5v9L12 21l8-4.5v-9" />
      <path d="M12 12v9" />
      <path d="m8 5.25 8 4.5" />
    </template>
    <template v-else-if="name === 'tags'">
      <path d="M20.6 13.4 13.4 20.6a2 2 0 0 1-2.8 0L3 13.2V3h10.2l7.4 7.6a2 2 0 0 1 0 2.8Z" />
      <circle cx="7.5" cy="7.5" r=".5" />
    </template>
    <template v-else-if="name === 'package-plus'">
      <path d="M4 7.5 12 3l8 4.5-8 4.5L4 7.5Z" />
      <path d="M4 7.5v9L12 21l4-2.25" />
      <path d="M12 12v9" />
      <path d="m8 5.25 8 4.5" />
      <path d="M19 14v6" />
      <path d="M16 17h6" />
    </template>
    <template v-else-if="name === 'history'"><path d="M3 12a9 9 0 1 0 3-6.7" /><path d="M3 4v5h5" /><path d="M12 7v5l3 2" /></template>
    <template v-else-if="name === 'refresh'"><path d="M20 11a8 8 0 0 0-14.2-4.9" /><path d="M5 3v4h4" /><path d="M4 13a8 8 0 0 0 14.2 4.9" /><path d="M19 21v-4h-4" /></template>
    <template v-else-if="name === 'map-pin'"><path d="M20 10c0 5-8 12-8 12S4 15 4 10a8 8 0 1 1 16 0Z" /><circle cx="12" cy="10" r="3" /></template>
    <template v-else-if="name === 'arrow-left-right'"><path d="M8 7h13" /><path d="m18 4 3 3-3 3" /><path d="M16 17H3" /><path d="m6 14-3 3 3 3" /></template>
    <template v-else-if="name === 'receipt-text'">
      <path d="M6 3h12v18l-2-1.25L14 21l-2-1.25L10 21l-2-1.25L6 21V3Z" />
      <path d="M9 8h6" />
      <path d="M9 12h6" />
      <path d="M9 16h3.5" />
    </template>
    <template v-else-if="name === 'purchase-order'">
      <path d="M9 4h6" />
      <path d="M9 4a3 3 0 0 1 6 0" />
      <path d="M8 5H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2" />
      <path d="M8 11h.01" />
      <path d="M11 11h5" />
      <path d="M8 15h.01" />
      <path d="M11 15h5" />
    </template>
    <template v-else-if="name === 'bell'"><path d="M18 8a6 6 0 1 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9" /><path d="M10 21h4" /></template>
    <template v-else-if="name === 'chart-column'"><path d="M4 20h16" /><path d="M6 16V8" /><path d="M12 16V4" /><path d="M18 16v-6" /></template>
    <template v-else-if="name === 'download'"><path d="M12 3v12" /><path d="m7 10 5 5 5-5" /><path d="M4 21h16" /></template>
    <template v-else-if="name === 'upload'"><path d="M12 15V3" /><path d="m7 8 5-5 5 5" /><path d="M4 21h16" /></template>
    <template v-else-if="name === 'clipboard-list'"><rect x="6" y="4" width="12" height="18" rx="2" /><path d="M9 4a3 3 0 0 1 6 0" /><path d="M9 11h6" /><path d="M9 15h6" /></template>
    <template v-else-if="name === 'truck'"><path d="M14 18V6H3v12h11Z" /><path d="M14 9h4l3 4v5h-7" /><circle cx="7" cy="18" r="2" /><circle cx="18" cy="18" r="2" /></template>
    <template v-else-if="name === 'users'"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.9" /><path d="M16 3.1a4 4 0 0 1 0 7.8" /></template>
    <template v-else-if="name === 'settings'"><path d="M12 15.5A3.5 3.5 0 1 0 12 8a3.5 3.5 0 0 0 0 7.5Z" /><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 .6 1.7 1.7 0 0 0-.4 1.1V21a2 2 0 0 1-4 0v-.09A1.7 1.7 0 0 0 8.6 19.4a1.7 1.7 0 0 0-1.88.34l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-.6-1 1.7 1.7 0 0 0-1.1-.4H3a2 2 0 0 1 0-4h.09A1.7 1.7 0 0 0 4.6 8.6a1.7 1.7 0 0 0-.34-1.88l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-.6 1.7 1.7 0 0 0 .4-1.1V3a2 2 0 0 1 4 0v.09A1.7 1.7 0 0 0 15.4 4.6a1.7 1.7 0 0 0 1.88-.34l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.7 1.7 0 0 0 19.4 9c.2.38.52.7.9.9.32.17.68.25 1.1.25H21a2 2 0 0 1 0 4h-.09a1.7 1.7 0 0 0-1.51.85Z" /></template>
    <template v-else-if="name === 'check-circle'"><circle cx="12" cy="12" r="10" /><path d="m8 12 3 3 5-6" /></template>
    <template v-else-if="name === 'triangle-alert'"><path d="m12 3 10 18H2L12 3Z" /><path d="M12 9v5" /><path d="M12 18h.01" /></template>
    <template v-else-if="name === 'circle-x'"><circle cx="12" cy="12" r="10" /><path d="m15 9-6 6" /><path d="m9 9 6 6" /></template>
    <template v-else-if="name === 'info'"><circle cx="12" cy="12" r="10" /><path d="M12 16v-4" /><path d="M12 8h.01" /></template>
    <template v-else-if="name === 'banknote'"><rect x="3" y="6" width="18" height="12" rx="2" /><circle cx="12" cy="12" r="3" /><path d="M6 12h.01M18 12h.01" /></template>
    <template v-else-if="name === 'qr-code'"><path d="M4 4h6v6H4zM14 4h6v6h-6zM4 14h6v6H4z" /><path d="M14 14h2v2h-2zM18 14h2v4h-2zM14 18h4v2h-4z" /></template>
    <template v-else-if="name === 'scan-barcode'"><path d="M7 7v10M10 7v10M14 7v10M17 7v10" /><path d="M4 4h4M16 4h4M4 20h4M16 20h4" /></template>
    <template v-else-if="name === 'menu'"><path d="M4 6h16M4 12h16M4 18h16" /></template>
    <template v-else-if="name === 'panel-left'"><rect x="3" y="4" width="18" height="16" rx="2" /><path d="M9 4v16" /></template>
    <template v-else-if="name === 'log-out'"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><path d="M16 17l5-5-5-5" /><path d="M21 12H9" /></template>
    <template v-else-if="name === 'x'"><path d="M18 6 6 18M6 6l12 12" /></template>
    <template v-else-if="name === 'plus'"><path d="M12 5v14M5 12h14" /></template>
    <template v-else-if="name === 'search'"><circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" /></template>
    <template v-else-if="name === 'sun'"><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></template>
    <template v-else-if="name === 'moon'"><path d="M21 12.8A8.5 8.5 0 1 1 11.2 3a6.7 6.7 0 0 0 9.8 9.8Z" /></template>
    <template v-else-if="name === 'text-size'"><path d="M4 7V4h10v3" /><path d="M9 4v16" /><path d="M6 20h6" /><path d="M15 13v-2h5v2" /><path d="M17.5 11v9" /><path d="M16 20h3" /></template>
    <template v-else-if="name === 'chevron-down'"><path d="m6 9 6 6 6-6" /></template>
    <template v-else-if="name === 'chevron-left'"><path d="m15 18-6-6 6-6" /></template>
    <template v-else-if="name === 'chevron-right'"><path d="m9 18 6-6-6-6" /></template>
    <template v-else-if="name === 'calendar'"><path d="M8 2v4M16 2v4" /><rect x="3" y="5" width="18" height="16" rx="2" /><path d="M3 10h18" /></template>
    <template v-else-if="name === 'role'">
      <path d="M12 3 5 6v5c0 4.5 3 8.5 7 10 4-1.5 7-5.5 7-10V6l-7-3Z" />
      <path d="M12 11a2.25 2.25 0 1 0 0-4.5 2.25 2.25 0 0 0 0 4.5Z" />
      <path d="M8.5 16c.7-1.7 2-2.5 3.5-2.5s2.8.8 3.5 2.5" />
    </template>
    <template v-else><path d="M12 3 3 8l9 5 9-5-9-5Z" /><path d="m3 15 9 5 9-5" /><path d="m3 11 9 5 9-5" /></template>
  </svg>
</template>
