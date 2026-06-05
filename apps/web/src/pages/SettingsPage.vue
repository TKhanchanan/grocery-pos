<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppCheckbox from '../components/AppCheckbox.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTabs from '../components/AppTabs.vue'
import AppTextarea from '../components/AppTextarea.vue'
import PageHeader from '../components/PageHeader.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import type { AppSettings, LineSettings, Location, NotificationLog } from '../types/navigation'
import { formatThaiDateTime } from '../utils/date'

type SettingsTab = 'shop' | 'receipt' | 'line' | 'accessibility' | 'system'

const app = useAppStore()
const locations = ref<Location[]>([])
const logs = ref<NotificationLog[]>([])
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const message = ref('')
const error = ref('')
const activeTab = ref<SettingsTab>('shop')

const tabs = computed(() => [
  { key: 'shop' as const, label: app.t('settings.tab.shop') },
  { key: 'receipt' as const, label: app.t('settings.tab.receipt') },
  { key: 'line' as const, label: app.t('settings.tab.line') },
  { key: 'accessibility' as const, label: app.t('settings.tab.accessibility') },
  { key: 'system' as const, label: app.t('settings.tab.system') },
])

const shopForm = reactive({
  shop_name: '',
  shop_phone: '',
  shop_address: '',
  default_location_id: '',
  receipt_footer: '',
})

const lineForm = reactive({
  line_enabled: false,
  line_token: '',
  line_token_masked: '',
  line_configured: false,
  line_target_id: '',
})

function statusClass(status: NotificationLog['status']) {
  return {
    SENT: 'bg-brand-100 text-brand-700',
    FAILED: 'bg-red-100 text-red-700',
    PENDING: 'bg-yellow-100 text-yellow-700',
    SKIPPED: 'bg-slate-100 text-slate-600',
  }[status]
}

function fallbackError(err: unknown, fallback: TranslationKey) {
  return err instanceof Error ? err.message : app.t(fallback)
}

function applySettings(settings: AppSettings) {
  shopForm.shop_name = settings.shop_name
  shopForm.shop_phone = settings.shop_phone
  shopForm.shop_address = settings.shop_address
  shopForm.default_location_id = settings.default_location_id ? String(settings.default_location_id) : ''
  shopForm.receipt_footer = settings.receipt_footer
}

function applyLineSettings(settings: LineSettings) {
  lineForm.line_enabled = settings.line_enabled
  lineForm.line_token = ''
  lineForm.line_token_masked = settings.line_token_masked
  lineForm.line_configured = settings.line_configured
  lineForm.line_target_id = settings.line_target_id
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [settings, lineSettings, locationRows, notificationRows] = await Promise.all([
      apiClient<AppSettings>('/v1/settings'),
      apiClient<LineSettings>('/v1/settings/line'),
      apiClient<Location[]>('/v1/locations'),
      apiClient<NotificationLog[]>('/v1/notification-logs'),
    ])
    applySettings(settings)
    applyLineSettings(lineSettings)
    locations.value = locationRows
    logs.value = notificationRows
  } catch (err) {
    error.value = fallbackError(err, 'settings.loadFailed')
  } finally {
    loading.value = false
  }
}

async function saveShopSettings() {
  saving.value = true
  message.value = ''
  error.value = ''
  try {
    const settings = await patchJSON<AppSettings>('/v1/settings', {
      shop_name: shopForm.shop_name,
      shop_phone: shopForm.shop_phone,
      shop_address: shopForm.shop_address,
      default_location_id: Number(shopForm.default_location_id || 0),
      receipt_footer: shopForm.receipt_footer,
    })
    applySettings(settings)
    message.value = app.t('settings.saved')
    app.pushToast({ type: 'success', message: app.t('settings.saved') })
  } catch (err) {
    error.value = fallbackError(err, 'settings.saveFailed')
    app.pushToast({ type: 'error', message: app.t('settings.saveFailed'), description: error.value })
  } finally {
    saving.value = false
  }
}

async function saveLineSettings() {
  saving.value = true
  message.value = ''
  error.value = ''
  try {
    const lineSettings = await patchJSON<LineSettings>('/v1/settings/line', {
      line_enabled: lineForm.line_enabled,
      line_token: lineForm.line_token,
      line_target_id: lineForm.line_target_id,
    })
    applyLineSettings(lineSettings)
    message.value = app.t('settings.lineSaved')
    app.pushToast({ type: 'success', message: app.t('settings.lineSaved') })
  } catch (err) {
    error.value = fallbackError(err, 'settings.lineSaveFailed')
    app.pushToast({ type: 'error', message: app.t('settings.lineSaveFailed'), description: error.value })
  } finally {
    saving.value = false
  }
}

async function testLine() {
  testing.value = true
  message.value = ''
  error.value = ''
  try {
    await postJSON<NotificationLog>('/v1/settings/line/test', {})
    message.value = app.t('settings.testSent')
    app.pushToast({ type: 'success', message: app.t('settings.testSent') })
    logs.value = await apiClient<NotificationLog[]>('/v1/notification-logs')
  } catch (err) {
    error.value = fallbackError(err, 'settings.testFailed')
    app.pushToast({ type: 'error', message: app.t('settings.testFailed'), description: error.value })
    logs.value = await apiClient<NotificationLog[]>('/v1/notification-logs').catch(() => logs.value)
  } finally {
    testing.value = false
  }
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader :title="app.t('settings.title')" :description="app.t('settings.description')" icon="settings" />

    <div v-if="error" class="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/40 dark:bg-red-950/40 dark:text-red-200">{{ error }}</div>
    <div v-if="message" class="mb-4 rounded-md border border-brand-100 bg-brand-50 p-3 text-sm text-brand-700 dark:border-emerald-400/30 dark:bg-emerald-500/10 dark:text-emerald-100">{{ message }}</div>
    <AppLoadingState v-if="loading" class="mb-4" :label="app.t('settings.loading')" />

    <div class="grid gap-4">
      <AppTabs :tabs="tabs" v-model="activeTab" />

      <AppCard v-if="activeTab === 'shop'" hover class="dark:bg-slate-900/80">
        <form class="grid gap-4" @submit.prevent="saveShopSettings">
          <div>
            <h2 class="font-bold">{{ app.t('settings.shopProfile') }}</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ app.t('settings.shopProfileDescription') }}</p>
          </div>
          <AppInput v-model="shopForm.shop_name" :label="app.t('settings.shopName')" />
          <AppInput v-model="shopForm.shop_phone" :label="app.t('settings.shopPhone')" />
          <AppTextarea v-model="shopForm.shop_address" :label="app.t('settings.shopAddress')" />
          <div class="flex justify-end">
            <AppButton type="submit" :loading="saving" :disabled="saving" icon="check-circle">{{ app.t('settings.saveSettings') }}</AppButton>
          </div>
        </form>
      </AppCard>

      <AppCard v-if="activeTab === 'receipt'" hover class="dark:bg-slate-900/80">
        <form class="grid gap-4" @submit.prevent="saveShopSettings">
          <div>
            <h2 class="font-bold">{{ app.t('settings.receipt') }}</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ app.t('settings.receiptDescription') }}</p>
          </div>
          <AppTextarea v-model="shopForm.receipt_footer" :label="app.t('settings.receiptFooter')" />
          <div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 text-sm dark:border-slate-700 dark:bg-slate-950/45">
            <p class="font-bold">{{ app.t('settings.preview') }}</p>
            <p class="mt-2 whitespace-pre-line text-slate-600 dark:text-slate-300">{{ shopForm.receipt_footer || app.t('settings.receiptFooter') }}</p>
          </div>
          <div class="flex justify-end">
            <AppButton type="submit" :loading="saving" :disabled="saving" icon="check-circle">{{ app.t('settings.saveSettings') }}</AppButton>
          </div>
        </form>
      </AppCard>

      <AppCard v-if="activeTab === 'line'" hover class="dark:bg-slate-900/80">
        <form class="grid gap-4" @submit.prevent="saveLineSettings">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h2 class="font-bold">{{ app.t('settings.line') }}</h2>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ app.t('settings.lineTokenHelp') }}</p>
            </div>
            <span class="rounded-full px-2 py-1 text-xs font-bold" :class="lineForm.line_configured ? 'bg-brand-100 text-brand-700 dark:bg-emerald-500/15 dark:text-emerald-100' : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-200'">
              {{ lineForm.line_configured ? app.t('settings.lineConfigured') : app.t('settings.lineNotConfigured') }}
            </span>
          </div>
          <AppCheckbox v-model="lineForm.line_enabled" :label="app.t('settings.lineEnabled')" :description="app.t('settings.lineTokenHelp')" />
          <AppInput v-model="lineForm.line_target_id" :label="app.t('settings.lineTarget')" />
          <AppInput v-model="lineForm.line_token" :label="app.t('settings.lineToken')" :placeholder="lineForm.line_token_masked || app.t('settings.lineTokenPlaceholder')" type="password" />
          <p class="text-xs text-slate-500 dark:text-slate-400">{{ app.t('settings.lineTokenHelp') }}</p>
          <div class="flex flex-wrap gap-2">
            <AppButton type="submit" :loading="saving" :disabled="saving" icon="check-circle">{{ app.t('settings.saveLine') }}</AppButton>
            <AppButton variant="secondary" :loading="testing" :disabled="testing" icon="bell" @click="testLine">{{ app.t('settings.testLine') }}</AppButton>
          </div>
        </form>
      </AppCard>

      <AppCard v-if="activeTab === 'accessibility'" class="dark:bg-slate-900/80">
        <div class="grid gap-5">
          <div>
            <h2 class="font-bold">{{ app.t('settings.appearance') }}</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ app.t('settings.appearanceDescription') }}</p>
          </div>
          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ app.t('settings.theme') }}</p>
            <div class="grid gap-2 sm:grid-cols-2">
              <AppButton :variant="app.theme === 'light' ? 'primary' : 'secondary'" @click="app.setTheme('light')">{{ app.t('settings.light') }}</AppButton>
              <AppButton :variant="app.theme === 'dark' ? 'primary' : 'secondary'" @click="app.setTheme('dark')">{{ app.t('settings.dark') }}</AppButton>
            </div>
          </div>
          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ app.t('settings.textSize') }}</p>
            <div class="grid gap-2 sm:grid-cols-4">
              <AppButton :variant="app.textSize === 'sm' ? 'primary' : 'secondary'" @click="app.setTextSize('sm')">{{ app.t('settings.small') }}</AppButton>
              <AppButton :variant="app.textSize === 'base' ? 'primary' : 'secondary'" @click="app.setTextSize('base')">{{ app.t('settings.default') }}</AppButton>
              <AppButton :variant="app.textSize === 'lg' ? 'primary' : 'secondary'" @click="app.setTextSize('lg')">{{ app.t('settings.large') }}</AppButton>
              <AppButton :variant="app.textSize === 'xl' ? 'primary' : 'secondary'" @click="app.setTextSize('xl')">{{ app.t('settings.extraLarge') }}</AppButton>
            </div>
          </div>
          <div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-slate-700 dark:bg-slate-950/45">
            <p class="font-bold">{{ app.t('settings.preview') }}</p>
            <p class="mt-2 text-slate-600 dark:text-slate-300">{{ app.t('settings.previewText') }}</p>
          </div>
        </div>
      </AppCard>

      <AppCard v-if="activeTab === 'system'" class="dark:bg-slate-900/80">
        <div class="grid gap-5">
          <div>
            <h2 class="font-bold">{{ app.t('settings.system') }}</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ app.t('settings.systemDescription') }}</p>
          </div>
          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ app.t('settings.language') }}</p>
            <div class="grid gap-2 sm:grid-cols-2">
              <AppButton :variant="app.language === 'th' ? 'primary' : 'secondary'" @click="app.setLanguage('th')">{{ app.t('settings.thai') }}</AppButton>
              <AppButton :variant="app.language === 'en' ? 'primary' : 'secondary'" @click="app.setLanguage('en')">{{ app.t('settings.english') }}</AppButton>
            </div>
          </div>
          <form class="grid gap-4" @submit.prevent="saveShopSettings">
            <AppSelect v-model="shopForm.default_location_id" :label="app.t('settings.defaultLocation')">
              <option value="">-</option>
              <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
            </AppSelect>
            <div class="flex justify-end">
              <AppButton type="submit" :loading="saving" :disabled="saving" icon="check-circle">{{ app.t('settings.saveSettings') }}</AppButton>
            </div>
          </form>
        </div>
      </AppCard>

      <AppCard v-if="activeTab === 'line'" class="dark:bg-slate-900/80">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="font-bold">{{ app.t('settings.notificationLogs') }}</h2>
        </div>
        <p v-if="logs.length === 0" class="mt-4 text-sm text-slate-500 dark:text-slate-400">{{ app.t('settings.noLogs') }}</p>
        <div v-else class="mt-4 hidden overflow-x-auto md:block">
          <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
            <thead class="bg-slate-50 dark:bg-slate-950/70">
              <tr>
                <th class="px-3 py-2 text-left">{{ app.t('settings.logEvent') }}</th>
                <th class="px-3 py-2 text-left">{{ app.t('settings.logChannel') }}</th>
                <th class="px-3 py-2 text-left">{{ app.t('settings.logRecipient') }}</th>
                <th class="px-3 py-2 text-left">{{ app.t('settings.logStatus') }}</th>
                <th class="px-3 py-2 text-left">{{ app.t('settings.logError') }}</th>
                <th class="px-3 py-2 text-left">{{ app.t('settings.logCreated') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="log in logs" :key="log.id">
                <td class="px-3 py-2 font-semibold">{{ log.event_type }}</td>
                <td class="px-3 py-2">{{ log.channel }}</td>
                <td class="px-3 py-2">{{ log.recipient || '-' }}</td>
                <td class="px-3 py-2"><span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(log.status)">{{ log.status }}</span></td>
                <td class="max-w-xs truncate px-3 py-2 text-slate-500 dark:text-slate-400">{{ log.error_message || '-' }}</td>
                <td class="px-3 py-2 text-slate-500 dark:text-slate-400">{{ formatThaiDateTime(log.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="mt-4 grid gap-3 md:hidden">
          <article v-for="log in logs" :key="log.id" class="rounded-lg border border-slate-200 p-3 dark:border-slate-700">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="font-bold">{{ log.event_type }}</p>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ log.channel }} · {{ log.recipient || '-' }}</p>
              </div>
              <span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(log.status)">{{ log.status }}</span>
            </div>
            <p v-if="log.error_message" class="mt-2 text-sm text-red-700">{{ log.error_message }}</p>
            <p class="mt-2 text-xs text-slate-500 dark:text-slate-400">{{ formatThaiDateTime(log.created_at) }}</p>
          </article>
        </div>
      </AppCard>
    </div>
  </section>
</template>
