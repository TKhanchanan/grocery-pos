<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppIcon from './AppIcon.vue'

type Picker = 'from' | 'to' | 'month' | ''

const props = withDefaults(defineProps<{
  dateFrom: string
  dateTo: string
  month?: string
  dateFromLabel: string
  dateToLabel: string
  monthLabel?: string
  datePlaceholder?: string
  monthPlaceholder?: string
  todayLabel?: string
  thisMonthLabel?: string
  locale?: string
  disabled?: boolean
  showMonth?: boolean
  showShortcuts?: boolean
}>(), {
  month: '',
  datePlaceholder: 'เลือกวันที่',
  monthPlaceholder: 'เลือกเดือน',
  todayLabel: 'วันนี้',
  thisMonthLabel: 'เดือนนี้',
  locale: 'th-TH-u-ca-buddhist',
  disabled: false,
  showMonth: false,
  showShortcuts: true,
})

const emit = defineEmits<{
  'update:dateFrom': [value: string]
  'update:dateTo': [value: string]
  'update:month': [value: string]
}>()

const openPicker = ref<Picker>('')
const root = ref<HTMLElement | null>(null)
const popoverID = `date-range-${Math.random().toString(36).slice(2)}`
const today = new Date()
const viewYear = ref(today.getFullYear())
const viewMonth = ref(today.getMonth())

const monthNames = computed(() => {
  const formatter = new Intl.DateTimeFormat(props.locale, { month: 'short' })
  return Array.from({ length: 12 }, (_, index) => formatter.format(new Date(2024, index, 1)))
})
const dayNames = computed(() => {
  const formatter = new Intl.DateTimeFormat(props.locale, { weekday: 'short' })
  return Array.from({ length: 7 }, (_, index) => formatter.format(new Date(2024, 0, index + 7)))
})

const rootClass = computed(() => {
  if (props.showMonth && props.showShortcuts) {
    return 'grid w-full gap-3 md:grid-cols-2 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto]'
  }

  if (props.showMonth && !props.showShortcuts) {
    return 'grid w-full gap-3 md:grid-cols-2 xl:grid-cols-3'
  }

  if (!props.showMonth && props.showShortcuts) {
    return 'grid w-full gap-3 md:grid-cols-2 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]'
  }

  return 'grid w-full gap-3 md:grid-cols-2'
})

const calendarTitle = computed(() => new Intl.DateTimeFormat(props.locale, { month: 'short', year: 'numeric' }).format(new Date(viewYear.value, viewMonth.value, 1)))
const monthTitle = computed(() => new Intl.DateTimeFormat(props.locale, { year: 'numeric' }).format(new Date(viewYear.value, 0, 1)))
const calendarDays = computed(() => {
  const firstDay = new Date(viewYear.value, viewMonth.value, 1)
  const start = new Date(firstDay)
  start.setDate(firstDay.getDate() - firstDay.getDay())
  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(start)
    date.setDate(start.getDate() + index)
    return {
      key: dateKey(date),
      label: date.getDate(),
      muted: date.getMonth() !== viewMonth.value,
      today: dateKey(date) === todayKey(),
    }
  })
})

watch(openPicker, (value) => {
  const selected = value === 'from' ? props.dateFrom : value === 'to' ? props.dateTo : props.month
  if (!selected) return
  const date = value === 'month' ? new Date(`${selected}-01T00:00:00`) : new Date(`${selected}T00:00:00`)
  if (Number.isNaN(date.getTime())) return
  viewYear.value = date.getFullYear()
  viewMonth.value = date.getMonth()
})

function setOpenPicker(value: Picker) {
  if (value) window.dispatchEvent(new CustomEvent('app-popover-open', { detail: popoverID }))
  openPicker.value = value
}

function onPopoverOpen(event: Event) {
  if ((event as CustomEvent<string>).detail !== popoverID) openPicker.value = ''
}

function onDocumentClick(event: MouseEvent) {
  if (!root.value?.contains(event.target as Node)) openPicker.value = ''
}

onMounted(() => {
  document.addEventListener('mousedown', onDocumentClick)
  window.addEventListener('app-popover-open', onPopoverOpen)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocumentClick)
  window.removeEventListener('app-popover-open', onPopoverOpen)
})

function dateKey(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function todayKey() {
  return dateKey(new Date())
}

function monthKey() {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
}

function formatDate(value: string) {
  if (!value) return props.datePlaceholder
  const date = new Date(`${value}T00:00:00`)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(props.locale, { day: '2-digit', month: 'short', year: 'numeric' }).format(date)
}

function formatMonth(value: string) {
  if (!value) return props.monthPlaceholder
  const date = new Date(`${value}-01T00:00:00`)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(props.locale, { month: 'long', year: 'numeric' }).format(date)
}

function setToday() {
  const value = todayKey()
  emit('update:dateFrom', value)
  emit('update:dateTo', value)
  emit('update:month', '')
  setOpenPicker('')
}

function setThisMonth() {
  emit('update:dateFrom', '')
  emit('update:dateTo', '')
  emit('update:month', monthKey())
  setOpenPicker('')
}

function previousMonth() {
  const date = new Date(viewYear.value, viewMonth.value - 1, 1)
  viewYear.value = date.getFullYear()
  viewMonth.value = date.getMonth()
}

function nextMonth() {
  const date = new Date(viewYear.value, viewMonth.value + 1, 1)
  viewYear.value = date.getFullYear()
  viewMonth.value = date.getMonth()
}

function isDateDisabled(value: string) {
  if (openPicker.value === 'from' && props.dateTo) return value > props.dateTo
  if (openPicker.value === 'to' && props.dateFrom) return value < props.dateFrom
  return false
}

function calendarDayClass(value: string, selected: string, muted: boolean, isToday: boolean) {
  const unavailable = isDateDisabled(value)
  return [
    value === selected ? 'bg-brand-600 text-white dark:bg-teal-300 dark:text-slate-950' : 'hover:bg-slate-100 dark:hover:bg-slate-800',
    muted ? 'text-slate-400 dark:text-slate-500' : 'text-slate-700 dark:text-slate-100',
    isToday && value !== selected ? 'text-brand-700 dark:text-teal-200' : '',
    unavailable ? 'cursor-not-allowed opacity-35 hover:bg-transparent dark:hover:bg-transparent' : '',
  ]
}

function selectDate(value: string) {
  if (isDateDisabled(value)) return
  if (openPicker.value === 'from') emit('update:dateFrom', value)
  if (openPicker.value === 'to') emit('update:dateTo', value)
  emit('update:month', '')
  setOpenPicker('')
}

function selectMonth(monthIndex: number) {
  emit('update:dateFrom', '')
  emit('update:dateTo', '')
  emit('update:month', `${viewYear.value}-${String(monthIndex + 1).padStart(2, '0')}`)
  setOpenPicker('')
}

function fieldClass(active: boolean) {
  return [
    'flex min-h-11 w-full items-center justify-between gap-2 rounded-xl border border-slate-300 bg-white px-3.5 py-2.5 text-left text-sm font-semibold text-slate-800 shadow-none transition disabled:cursor-not-allowed disabled:bg-slate-100 disabled:opacity-70 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:disabled:bg-slate-800',
    active ? 'border-brand-500 dark:border-teal-300' : '',
  ]
}
</script>

<template>
  <div ref="root" :class="rootClass">
    <div class="relative grid min-w-0 gap-1.5 text-sm">
      <span class="font-semibold text-slate-700 dark:text-slate-200">{{ dateFromLabel }}</span>
      <button type="button" :class="fieldClass(openPicker === 'from')" :disabled="disabled" @click="setOpenPicker(openPicker === 'from' ? '' : 'from')">
        <span>{{ formatDate(dateFrom) }}</span>
        <AppIcon name="calendar" :size="17" />
      </button>
      <div v-if="openPicker === 'from'" class="absolute left-0 top-full z-[110] mt-2 w-[min(340px,calc(100vw-2rem))] rounded-2xl bg-white p-3 shadow-2xl shadow-slate-950/20 dark:bg-slate-900 dark:shadow-black/35">
        <div class="flex items-center justify-between gap-2">
          <button type="button" class="grid h-9 w-9 place-items-center rounded-xl bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200" @click="previousMonth"><AppIcon name="chevron-left" :size="17" /></button>
          <p class="font-black text-slate-900 dark:text-slate-50">{{ calendarTitle }}</p>
          <button type="button" class="grid h-9 w-9 place-items-center rounded-xl bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200" @click="nextMonth"><AppIcon name="chevron-right" :size="17" /></button>
        </div>
        <div class="mt-3 grid grid-cols-7 gap-1 text-center text-xs font-black text-slate-500 dark:text-slate-400">
          <span v-for="day in dayNames" :key="day" class="py-1">{{ day }}</span>
        </div>
        <div class="mt-1 grid grid-cols-7 gap-1">
          <button v-for="day in calendarDays" :key="day.key" type="button" class="grid h-9 place-items-center rounded-lg text-sm font-bold transition disabled:pointer-events-none" :class="calendarDayClass(day.key, dateFrom, day.muted, day.today)" :disabled="isDateDisabled(day.key)" @click="selectDate(day.key)">
            {{ day.label }}
          </button>
        </div>
      </div>
    </div>

    <div class="relative grid min-w-0 gap-1.5 text-sm">
      <span class="font-semibold text-slate-700 dark:text-slate-200">{{ dateToLabel }}</span>
      <button type="button" :class="fieldClass(openPicker === 'to')" :disabled="disabled" @click="setOpenPicker(openPicker === 'to' ? '' : 'to')">
        <span>{{ formatDate(dateTo) }}</span>
        <AppIcon name="calendar" :size="17" />
      </button>
      <div v-if="openPicker === 'to'" class="absolute left-0 top-full z-[110] mt-2 w-[min(340px,calc(100vw-2rem))] rounded-2xl bg-white p-3 shadow-2xl shadow-slate-950/20 dark:bg-slate-900 dark:shadow-black/35">
        <div class="flex items-center justify-between gap-2">
          <button type="button" class="grid h-9 w-9 place-items-center rounded-xl bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200" @click="previousMonth"><AppIcon name="chevron-left" :size="17" /></button>
          <p class="font-black text-slate-900 dark:text-slate-50">{{ calendarTitle }}</p>
          <button type="button" class="grid h-9 w-9 place-items-center rounded-xl bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200" @click="nextMonth"><AppIcon name="chevron-right" :size="17" /></button>
        </div>
        <div class="mt-3 grid grid-cols-7 gap-1 text-center text-xs font-black text-slate-500 dark:text-slate-400">
          <span v-for="day in dayNames" :key="day" class="py-1">{{ day }}</span>
        </div>
        <div class="mt-1 grid grid-cols-7 gap-1">
          <button v-for="day in calendarDays" :key="day.key" type="button" class="grid h-9 place-items-center rounded-lg text-sm font-bold transition disabled:pointer-events-none" :class="calendarDayClass(day.key, dateTo, day.muted, day.today)" :disabled="isDateDisabled(day.key)" @click="selectDate(day.key)">
            {{ day.label }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showMonth" class="relative grid min-w-0 gap-1.5 text-sm">
      <span class="font-semibold text-slate-700 dark:text-slate-200">{{ monthLabel }}</span>
      <button type="button" :class="fieldClass(openPicker === 'month')" :disabled="disabled" @click="setOpenPicker(openPicker === 'month' ? '' : 'month')">
        <span class="min-w-0 truncate">{{ formatMonth(month) }}</span>
        <AppIcon name="calendar" :size="17" class="shrink-0" />
      </button>
      <div v-if="openPicker === 'month'" class="absolute left-0 top-full z-[110] mt-2 w-[min(340px,calc(100vw-2rem))] rounded-2xl bg-white p-3 shadow-2xl shadow-slate-950/20 dark:bg-slate-900 dark:shadow-black/35">
        <div class="flex items-center justify-between gap-2">
          <button type="button" class="grid h-9 w-9 place-items-center rounded-xl bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200" @click="viewYear -= 1"><AppIcon name="chevron-left" :size="17" /></button>
          <p class="font-black text-slate-900 dark:text-slate-50">{{ monthTitle }}</p>
          <button type="button" class="grid h-9 w-9 place-items-center rounded-xl bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200" @click="viewYear += 1"><AppIcon name="chevron-right" :size="17" /></button>
        </div>
        <div class="mt-3 grid grid-cols-3 gap-2">
          <button v-for="(label, index) in monthNames" :key="label" type="button" class="min-h-10 rounded-xl text-sm font-black transition hover:bg-slate-100 dark:hover:bg-slate-800" :class="month === `${viewYear}-${String(index + 1).padStart(2, '0')}` ? 'bg-brand-600 text-white dark:bg-teal-300 dark:text-slate-950' : 'text-slate-700 dark:text-slate-100'" @click="selectMonth(index)">
            {{ label }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showShortcuts" class="flex items-end gap-2">
      <button type="button"
        class="min-h-11 rounded-xl bg-slate-100 px-3 text-sm font-black text-slate-700 hover:bg-slate-200 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700"
        :disabled="disabled" @click="setToday">
        {{ todayLabel }}
      </button>

      <button v-if="showMonth" type="button"
        class="min-h-11 rounded-xl bg-slate-100 px-3 text-sm font-black text-slate-700 hover:bg-slate-200 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700"
        :disabled="disabled" @click="setThisMonth">
        {{ thisMonthLabel }}
      </button>
    </div>
  </div>
</template>
