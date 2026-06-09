<script setup lang="ts">
import { Comment, Fragment, Text, computed, nextTick, onBeforeUnmount, onMounted, ref, useSlots, type CSSProperties, type VNode } from 'vue'
import AppIcon from './AppIcon.vue'

const props = withDefaults(defineProps<{ label?: string; modelValue?: string | number; helper?: string; error?: string; hideArrow?: boolean; disabled?: boolean }>(), {
  disabled: false,
  hideArrow: false,
})
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const slots = useSlots()
const open = ref(false)
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const menu = ref<HTMLElement | null>(null)
const menuStyle = ref<CSSProperties>({})
const popoverID = `select-${Math.random().toString(36).slice(2)}`

interface SelectOption {
  value: string
  label: string
  disabled: boolean
}

const options = computed<SelectOption[]>(() => extractOptions(slots.default?.() ?? []))
const selectedOption = computed(() => options.value.find((option) => option.value === String(props.modelValue ?? '')) ?? options.value[0])

function textFromChildren(children: unknown): string {
  if (typeof children === 'string' || typeof children === 'number') return String(children)
  if (Array.isArray(children)) return children.map((child) => {
    if (typeof child === 'string' || typeof child === 'number') return String(child)
    return textFromChildren((child as VNode).children)
  }).join('')
  return ''
}

function extractOptions(nodes: VNode[]): SelectOption[] {
  const result: SelectOption[] = []
  for (const node of nodes) {
    if (node.type === Comment || node.type === Text) continue
    if (node.type === Fragment && Array.isArray(node.children)) {
      result.push(...extractOptions(node.children as VNode[]))
      continue
    }
    if (node.type === 'option') {
      const rawValue = (node.props as Record<string, unknown> | null)?.value
      const label = textFromChildren(node.children).trim()
      result.push({
        value: rawValue === undefined ? label : String(rawValue),
        label,
        disabled: Boolean((node.props as Record<string, unknown> | null)?.disabled),
      })
    }
  }
  return result
}

function selectOption(option: SelectOption) {
  if (option.disabled) return
  emit('update:modelValue', option.value)
  open.value = false
}

async function setOpen(value: boolean) {
  if (value) window.dispatchEvent(new CustomEvent('app-popover-open', { detail: popoverID }))
  open.value = value
  if (value) {
    await nextTick()
    updateMenuPosition()
  }
}

function updateMenuPosition() {
  if (!open.value || !trigger.value) return
  const rect = trigger.value.getBoundingClientRect()
  const gap = 8
  const viewportPadding = 12
  const availableBelow = window.innerHeight - rect.bottom - gap - viewportPadding
  const availableAbove = rect.top - gap - viewportPadding
  const openAbove = availableBelow < 160 && availableAbove > availableBelow
  const maxHeight = Math.max(96, Math.min(288, openAbove ? availableAbove : availableBelow))

  menuStyle.value = {
    left: `${rect.left}px`,
    top: openAbove ? 'auto' : `${rect.bottom + gap}px`,
    bottom: openAbove ? `${window.innerHeight - rect.top + gap}px` : 'auto',
    width: `${rect.width}px`,
    maxHeight: `${maxHeight}px`,
  }
}

function onDocumentClick(event: MouseEvent) {
  const target = event.target as Node
  if (!root.value?.contains(target) && !menu.value?.contains(target)) open.value = false
}

function onPopoverOpen(event: Event) {
  if ((event as CustomEvent<string>).detail !== popoverID) open.value = false
}

onMounted(() => {
  document.addEventListener('mousedown', onDocumentClick)
  window.addEventListener('app-popover-open', onPopoverOpen)
  window.addEventListener('resize', updateMenuPosition)
  document.addEventListener('scroll', updateMenuPosition, true)
})
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocumentClick)
  window.removeEventListener('app-popover-open', onPopoverOpen)
  window.removeEventListener('resize', updateMenuPosition)
  document.removeEventListener('scroll', updateMenuPosition, true)
})
</script>

<template>
  <div ref="root" class="relative grid gap-1.5 text-sm">
    <span v-if="label" class="font-semibold text-slate-700 dark:text-slate-200">{{ label }}</span>
    <button
      ref="trigger"
      type="button"
      class="flex min-h-11 w-full items-center justify-between gap-2 rounded-xl border border-slate-300 bg-white px-3.5 py-2.5 text-left text-slate-950 shadow-none transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:opacity-70 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-50 dark:hover:bg-slate-800 dark:disabled:bg-slate-800"
      :disabled="disabled"
      :aria-expanded="open"
      @click="setOpen(!open)"
    >
      <span class="min-w-0 truncate">{{ selectedOption?.label ?? '-' }}</span>
      <AppIcon v-if="!hideArrow" name="chevron-down" :size="16" class="shrink-0 text-slate-500" />
    </button>
    <Teleport to="body">
      <div
        v-if="open"
        ref="menu"
        class="fixed z-[220] overflow-auto rounded-xl bg-white p-1 shadow-2xl shadow-slate-950/20 dark:bg-slate-900 dark:shadow-black/35"
        :style="menuStyle"
      >
        <button
          v-for="option in options"
          :key="`${option.value}-${option.label}`"
          type="button"
          class="flex min-h-10 w-full items-center rounded-lg px-3 text-left text-sm font-bold transition disabled:cursor-not-allowed disabled:opacity-50"
          :class="option.value === String(modelValue ?? '') ? 'bg-brand-600 text-white dark:bg-teal-300 dark:text-slate-950' : 'text-slate-700 hover:bg-slate-100 dark:text-slate-100 dark:hover:bg-slate-800'"
          :disabled="option.disabled"
          @click="selectOption(option)"
        >
          <span class="truncate">{{ option.label }}</span>
        </button>
      </div>
    </Teleport>
    <span v-if="error" class="text-xs font-semibold text-red-600 dark:text-red-300">{{ error }}</span>
    <span v-else-if="helper" class="text-xs text-slate-500 dark:text-slate-400">{{ helper }}</span>
  </div>
</template>
