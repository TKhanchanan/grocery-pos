<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { apiClient, patchJSON, postJSON } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppCard from '../components/AppCard.vue'
import AppCheckbox from '../components/AppCheckbox.vue'
import AppInput from '../components/AppInput.vue'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppSelect from '../components/AppSelect.vue'
import AppTextarea from '../components/AppTextarea.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAppStore } from '../stores/app'
import type { AppSettings, LineSettings, Location, NotificationLog } from '../types/navigation'

const app = useAppStore()
const locations = ref<Location[]>([])
const logs = ref<NotificationLog[]>([])
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const message = ref('')
const error = ref('')

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
    error.value = err instanceof Error ? err.message : 'Could not load settings'
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
    message.value = 'Settings saved'
    app.pushToast({ type: 'success', message: 'Settings saved' })
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save settings'
    app.pushToast({ type: 'error', message: 'Could not save settings', description: error.value })
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
    message.value = 'LINE settings saved'
    app.pushToast({ type: 'success', message: 'LINE settings saved' })
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save LINE settings'
    app.pushToast({ type: 'error', message: 'Could not save LINE settings', description: error.value })
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
    message.value = 'Test notification sent'
    app.pushToast({ type: 'success', message: 'Test notification sent' })
    logs.value = await apiClient<NotificationLog[]>('/v1/notification-logs')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not send LINE test'
    app.pushToast({ type: 'error', message: 'Could not send LINE test', description: error.value })
    logs.value = await apiClient<NotificationLog[]>('/v1/notification-logs').catch(() => logs.value)
  } finally {
    testing.value = false
  }
}

onMounted(load)
</script>

<template>
  <section>
    <PageHeader :title="app.t('settings.title')" :eyebrow="app.t('settings.eyebrow')" :description="app.t('settings.description')" icon="settings">
      <AppButton variant="secondary" icon="history" @click="load">{{ app.t('settings.refresh') }}</AppButton>
    </PageHeader>

    <div v-if="error" class="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
    <div v-if="message" class="mb-4 rounded-md border border-brand-100 bg-brand-50 p-3 text-sm text-brand-700">{{ message }}</div>
    <AppLoadingState v-if="loading" class="mb-4" :label="app.t('settings.loading')" />

    <div class="grid gap-4 xl:grid-cols-2">
      <AppCard hover>
        <form class="grid gap-4" @submit.prevent="saveShopSettings">
          <div>
            <h2 class="font-bold">{{ app.t('settings.shopProfile') }}</h2>
            <p class="mt-1 text-sm text-slate-500">{{ app.t('settings.appearanceDescription') }}</p>
          </div>
          <AppInput v-model="shopForm.shop_name" :label="app.t('settings.shopName')" />
          <AppInput v-model="shopForm.shop_phone" :label="app.t('settings.shopPhone')" />
          <AppTextarea v-model="shopForm.shop_address" :label="app.t('settings.shopAddress')" />
          <AppSelect v-model="shopForm.default_location_id" :label="app.t('settings.defaultLocation')">
            <option value="">-</option>
            <option v-for="location in locations" :key="location.id" :value="String(location.id)">{{ location.name }}</option>
          </AppSelect>
          <div class="grid gap-2">
            <h3 class="font-bold">{{ app.t('settings.receipt') }}</h3>
            <AppTextarea v-model="shopForm.receipt_footer" :label="app.t('settings.receiptFooter')" />
          </div>
          <AppButton type="submit" :loading="saving" :disabled="saving" icon="check-circle">{{ app.t('settings.saveSettings') }}</AppButton>
        </form>
      </AppCard>

      <AppCard hover>
        <form class="grid gap-4" @submit.prevent="saveLineSettings">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h2 class="font-bold">{{ app.t('settings.line') }}</h2>
              <p class="mt-1 text-sm text-slate-500">{{ app.t('settings.lineTokenHelp') }}</p>
            </div>
            <span class="rounded-full px-2 py-1 text-xs font-bold" :class="lineForm.line_configured ? 'bg-brand-100 text-brand-700' : 'bg-slate-100 text-slate-600'">
              {{ lineForm.line_configured ? app.t('settings.lineConfigured') : app.t('settings.lineNotConfigured') }}
            </span>
          </div>
          <AppCheckbox v-model="lineForm.line_enabled" :label="app.t('settings.lineEnabled')" :description="app.t('settings.lineTokenHelp')" />
          <AppInput v-model="lineForm.line_target_id" :label="app.t('settings.lineTarget')" />
          <AppInput v-model="lineForm.line_token" :label="app.t('settings.lineToken')" :placeholder="lineForm.line_token_masked || 'Enter token'" type="password" />
          <p class="text-xs text-slate-500">{{ app.t('settings.lineTokenHelp') }}</p>
          <div class="flex flex-wrap gap-2">
            <AppButton type="submit" :loading="saving" :disabled="saving" icon="check-circle">{{ app.t('settings.saveLine') }}</AppButton>
            <AppButton variant="secondary" :loading="testing" :disabled="testing" icon="bell" @click="testLine">{{ app.t('settings.testLine') }}</AppButton>
          </div>
        </form>
      </AppCard>

      <AppCard>
        <div class="grid gap-5">
          <div>
            <h2 class="font-bold">{{ app.t('settings.appearance') }}</h2>
            <p class="mt-1 text-sm text-slate-500">{{ app.t('settings.appearanceDescription') }}</p>
          </div>
          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700">{{ app.t('settings.language') }}</p>
            <div class="grid gap-2 sm:grid-cols-2">
              <AppButton :variant="app.language === 'th' ? 'primary' : 'secondary'" @click="app.setLanguage('th')">{{ app.t('settings.thai') }}</AppButton>
              <AppButton :variant="app.language === 'en' ? 'primary' : 'secondary'" @click="app.setLanguage('en')">{{ app.t('settings.english') }}</AppButton>
            </div>
          </div>
          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700">{{ app.t('settings.theme') }}</p>
            <div class="grid gap-2 sm:grid-cols-2">
              <AppButton :variant="app.theme === 'light' ? 'primary' : 'secondary'" @click="app.setTheme('light')">{{ app.t('settings.light') }}</AppButton>
              <AppButton :variant="app.theme === 'dark' ? 'primary' : 'secondary'" @click="app.setTheme('dark')">{{ app.t('settings.dark') }}</AppButton>
            </div>
          </div>
          <div class="grid gap-2">
            <p class="text-sm font-semibold text-slate-700">{{ app.t('settings.textSize') }}</p>
            <div class="grid gap-2 sm:grid-cols-4">
              <AppButton :variant="app.textSize === 'sm' ? 'primary' : 'secondary'" @click="app.setTextSize('sm')">{{ app.t('settings.small') }}</AppButton>
              <AppButton :variant="app.textSize === 'base' ? 'primary' : 'secondary'" @click="app.setTextSize('base')">{{ app.t('settings.default') }}</AppButton>
              <AppButton :variant="app.textSize === 'lg' ? 'primary' : 'secondary'" @click="app.setTextSize('lg')">{{ app.t('settings.large') }}</AppButton>
              <AppButton :variant="app.textSize === 'xl' ? 'primary' : 'secondary'" @click="app.setTextSize('xl')">{{ app.t('settings.extraLarge') }}</AppButton>
            </div>
          </div>
        </div>
      </AppCard>

      <AppCard class="xl:col-span-2">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="font-bold">{{ app.t('settings.notificationLogs') }}</h2>
          <AppButton variant="secondary" icon="history" @click="load">{{ app.t('settings.refresh') }}</AppButton>
        </div>
        <p v-if="logs.length === 0" class="mt-4 text-sm text-slate-500">{{ app.t('settings.noLogs') }}</p>
        <div v-else class="mt-4 hidden overflow-x-auto md:block">
          <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50">
              <tr>
                <th class="px-3 py-2 text-left">Event</th>
                <th class="px-3 py-2 text-left">Channel</th>
                <th class="px-3 py-2 text-left">Recipient</th>
                <th class="px-3 py-2 text-left">Status</th>
                <th class="px-3 py-2 text-left">Error</th>
                <th class="px-3 py-2 text-left">Created</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="log in logs" :key="log.id">
                <td class="px-3 py-2 font-semibold">{{ log.event_type }}</td>
                <td class="px-3 py-2">{{ log.channel }}</td>
                <td class="px-3 py-2">{{ log.recipient || '-' }}</td>
                <td class="px-3 py-2"><span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(log.status)">{{ log.status }}</span></td>
                <td class="max-w-xs truncate px-3 py-2 text-slate-500">{{ log.error_message || '-' }}</td>
                <td class="px-3 py-2 text-slate-500">{{ new Date(log.created_at).toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="mt-4 grid gap-3 md:hidden">
          <article v-for="log in logs" :key="log.id" class="rounded-lg border border-slate-200 p-3">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="font-bold">{{ log.event_type }}</p>
                <p class="text-sm text-slate-500">{{ log.channel }} · {{ log.recipient || '-' }}</p>
              </div>
              <span class="rounded-full px-2 py-1 text-xs font-bold" :class="statusClass(log.status)">{{ log.status }}</span>
            </div>
            <p v-if="log.error_message" class="mt-2 text-sm text-red-700">{{ log.error_message }}</p>
            <p class="mt-2 text-xs text-slate-500">{{ new Date(log.created_at).toLocaleString() }}</p>
          </article>
        </div>
      </AppCard>
    </div>
  </section>
</template>
