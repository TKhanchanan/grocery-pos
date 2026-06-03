<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { useAppStore } from '../stores/app'
import { useAuthStore } from '../stores/auth'
import type { AuthMeResponse, User } from '../types/navigation'
import AppAvatar from './AppAvatar.vue'
import AppButton from './AppButton.vue'
import AppIcon from './AppIcon.vue'
import AppModal from './AppModal.vue'

const app = useAppStore()
const auth = useAuthStore()
const router = useRouter()
const open = ref(false)
const profileOpen = ref(false)
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const root = ref<HTMLElement | null>(null)

const displayName = computed(() => auth.user?.fullName || auth.user?.username || 'User')
const roleLabel = computed(() => auth.roles.map((role) => role.name).join(', ') || auth.user?.role || '-')
const permissionsSummary = computed(() => `${auth.permissions.length} permissions`)

function onDocumentClick(event: MouseEvent) {
  if (root.value && !root.value.contains(event.target as Node)) open.value = false
}

function showProfile() {
  open.value = false
  profileOpen.value = true
}

function chooseAvatar() {
  open.value = false
  profileOpen.value = true
  window.setTimeout(() => fileInput.value?.click(), 50)
}

async function uploadAvatar(file?: File) {
  if (!file) return
  const allowed = ['image/jpeg', 'image/png', 'image/webp']
  if (!allowed.includes(file.type)) {
    app.pushToast({ type: 'error', message: 'Invalid image type', description: 'รองรับ JPG, PNG, WEBP เท่านั้น' })
    return
  }
  if (file.size > 2 * 1024 * 1024) {
    app.pushToast({ type: 'error', message: 'Image is too large', description: 'ขนาดรูปต้องไม่เกิน 2MB' })
    return
  }
  uploading.value = true
  try {
    const body = new FormData()
    body.append('avatar', file)
    await apiClient<{ user: User }>('/v1/profile/avatar', { method: 'POST', body })
    await auth.loadMe()
    app.pushToast({ type: 'success', message: 'Profile image updated' })
  } catch (err) {
    app.pushToast({ type: 'error', message: 'Could not upload profile image', description: err instanceof Error ? err.message : '' })
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function removeAvatar() {
  uploading.value = true
  try {
    await apiClient<User>('/v1/profile/avatar', { method: 'DELETE' })
    await auth.loadMe()
    app.pushToast({ type: 'success', message: 'Profile image removed' })
  } catch (err) {
    app.pushToast({ type: 'error', message: 'Could not remove profile image', description: err instanceof Error ? err.message : '' })
  } finally {
    uploading.value = false
  }
}

async function logout() {
  open.value = false
  await auth.logout()
  app.pushToast({ type: 'info', message: 'Logged out' })
  router.push('/login')
}

onMounted(() => document.addEventListener('mousedown', onDocumentClick))
onBeforeUnmount(() => document.removeEventListener('mousedown', onDocumentClick))
</script>

<template>
  <div ref="root" class="relative">
    <button
      aria-label="เปิดเมนูโปรไฟล์" type="button" @click="open = !open">
      <AppAvatar :src="auth.user?.avatar_url" :name="displayName" size="md" />
    </button>
    <div v-if="open" class="premium-surface absolute right-0 z-50 mt-3 w-[min(320px,calc(100vw-24px))] rounded-2xl border p-2 shadow-2xl">
      <div class="flex items-center gap-3 border-b border-slate-100 p-3 dark:border-slate-800">
        <AppAvatar :src="auth.user?.avatar_url" :name="displayName" size="md" />
        <div class="min-w-0">
          <p class="truncate text-base font-black">{{ displayName }}</p>
          <p class="truncate text-sm text-slate-500 dark:text-slate-400">{{ roleLabel }}</p>
        </div>
      </div>
      <div class="grid gap-1 p-2">
        <button class="flex min-h-11 items-center gap-3 rounded-xl px-3 text-left text-sm font-bold text-slate-700 hover:bg-brand-50 dark:text-slate-100 dark:hover:bg-slate-800" @click="showProfile">
          <AppIcon name="users" :size="18" />ดูโปรไฟล์
        </button>
        <button class="flex min-h-11 items-center gap-3 rounded-xl px-3 text-left text-sm font-bold text-slate-700 hover:bg-brand-50 dark:text-slate-100 dark:hover:bg-slate-800" @click="chooseAvatar">
          <AppIcon name="upload" :size="18" />อัปโหลดรูปโปรไฟล์
        </button>
        <RouterLink v-if="auth.hasPermission('settings.view')" to="/settings" class="flex min-h-11 items-center gap-3 rounded-xl px-3 text-sm font-bold text-slate-700 hover:bg-brand-50 dark:text-slate-100 dark:hover:bg-slate-800" @click="open = false">
          <AppIcon name="settings" :size="18" />ตั้งค่าบัญชี
        </RouterLink>
        <button class="mt-1 flex min-h-11 items-center gap-3 rounded-xl px-3 text-left text-sm font-bold text-red-700 hover:bg-red-50 dark:text-red-200 dark:hover:bg-red-500/15" @click="logout">
          <AppIcon name="log-out" :size="18" />ออกจากระบบ
        </button>
      </div>
    </div>

    <AppModal :open="profileOpen" title="ดูโปรไฟล์" @close="profileOpen = false">
      <div class="grid gap-5">
        <div class="flex flex-col items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50/80 p-5 text-center dark:border-slate-700 dark:bg-slate-950/60">
          <AppAvatar :src="auth.user?.avatar_url" :name="displayName" size="xl" />
          <div>
            <h2 class="text-xl font-black">{{ displayName }}</h2>
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ auth.user?.username }}</p>
          </div>
          <div class="flex flex-wrap justify-center gap-2">
            <span v-for="role in auth.roles" :key="role.code" class="rounded-full bg-brand-100 px-2.5 py-1 text-xs font-bold text-brand-700 dark:bg-emerald-500/20 dark:text-emerald-100">{{ role.name }}</span>
          </div>
        </div>

        <dl class="grid gap-3 text-sm sm:grid-cols-2">
          <div class="rounded-xl bg-white/70 p-3 dark:bg-slate-950/60">
            <dt class="text-slate-500 dark:text-slate-400">Account status</dt>
            <dd class="font-bold">{{ auth.user?.active ? 'Active' : 'Disabled' }}</dd>
          </div>
          <div class="rounded-xl bg-white/70 p-3 dark:bg-slate-950/60">
            <dt class="text-slate-500 dark:text-slate-400">Permissions</dt>
            <dd class="font-bold">{{ permissionsSummary }}</dd>
          </div>
          <div class="rounded-xl bg-white/70 p-3 dark:bg-slate-950/60 sm:col-span-2">
            <dt class="text-slate-500 dark:text-slate-400">Avatar updated</dt>
            <dd class="font-bold">{{ auth.user?.avatar_updated_at ? new Date(auth.user.avatar_updated_at).toLocaleString() : '-' }}</dd>
          </div>
        </dl>

        <div class="rounded-2xl border border-slate-200 p-4 dark:border-slate-700">
          <p class="font-bold">อัปโหลดรูปโปรไฟล์</p>
          <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">รองรับ JPG, PNG, WEBP ขนาดไม่เกิน 2MB</p>
          <input ref="fileInput" class="hidden" type="file" accept="image/jpeg,image/png,image/webp" @change="uploadAvatar(($event.target as HTMLInputElement).files?.[0])" />
          <div class="mt-4 flex flex-wrap gap-2">
            <AppButton variant="secondary" icon="upload" :loading="uploading" @click="fileInput?.click()">เลือกรูปโปรไฟล์</AppButton>
            <AppButton v-if="auth.user?.avatar_url" variant="danger" :loading="uploading" @click="removeAvatar">ลบรูปโปรไฟล์</AppButton>
          </div>
        </div>
      </div>
    </AppModal>
  </div>
</template>
