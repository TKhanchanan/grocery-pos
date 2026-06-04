<script setup lang="ts">
import { computed, ref } from 'vue'
import AppBadge from './AppBadge.vue'
import AppButton from './AppButton.vue'
import AppCard from './AppCard.vue'
import AppEmptyState from './AppEmptyState.vue'
import AppIcon from './AppIcon.vue'
import AppInput from './AppInput.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import type { TranslationKey } from '../i18n'
import { useAppStore } from '../stores/app'
import { useCartStore } from '../stores/cart'

const props = defineProps<{ submitDisabled?: boolean }>()
defineEmits<{ submitSale: [] }>()

const app = useAppStore()
const cart = useCartStore()
const clearConfirmOpen = ref(false)

const insufficientPayment = computed(() => cart.items.length > 0 && cart.receivedAmount < cart.totalAmount)
const checkoutStatus = computed(() => {
  if (cart.items.length === 0) return { label: app.t('pos.emptyCartStatus'), tone: 'neutral' as const }
  if (insufficientPayment.value) return { label: app.t('pos.waitingPayment'), tone: 'warning' as const }
  return { label: app.t('pos.ready'), tone: 'success' as const }
})

const paymentMethods = computed(() => [
  { value: 'CASH' as const, label: app.t('pos.cash'), icon: 'banknote' as const },
  { value: 'QR' as const, label: app.t('pos.qr'), icon: 'qr-code' as const },
])

function t(key: TranslationKey, params: Record<string, string | number> = {}) {
  let text = String(app.t(key))
  for (const [name, value] of Object.entries(params)) {
    text = text.replaceAll(`{${name}}`, String(value))
  }
  return text
}

function money(value: number) {
  const locale = app.language === 'th' ? 'th-TH' : 'en-US'
  const amount = value.toLocaleString(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return t('pos.currency', { amount })
}

function setExactPayment() {
  cart.setReceivedAmount(cart.totalAmount)
}

function confirmClearCart() {
  cart.clearCart()
  clearConfirmOpen.value = false
}

function setQuantity(productId: number, quantity: number) {
  cart.updateQuantity(productId, quantity)
}
</script>

<template>
  <AppCard class="h-full overflow-hidden">
    <div class="flex items-center justify-between gap-3">
      <div>
        <p class="text-xs font-black uppercase text-brand-700 dark:text-emerald-300">{{ app.t('pos.payment') }}</p>
        <h2 class="text-xl font-black">{{ app.t('pos.cart') }}</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400">{{ cart.totalItems }} {{ app.t('pos.items') }}</p>
      </div>
      <AppButton variant="secondary" :disabled="cart.items.length === 0 || cart.isSubmitting" @click="clearConfirmOpen = true">
        {{ app.t('pos.clear') }}
      </AppButton>
    </div>

    <AppEmptyState
      v-if="cart.items.length === 0"
      class="mt-4"
      :title="app.t('pos.cartEmpty')"
      :description="app.t('pos.cartEmptyDescription')"
      icon="shopping-cart"
    />

    <div v-else class="mt-4 grid max-h-[42vh] gap-3 overflow-y-auto pr-1 xl:max-h-[45vh]">
      <article v-for="item in cart.items" :key="item.productId" class="pos-cart-item rounded-2xl border border-slate-200 bg-white/70 p-3 dark:border-slate-700 dark:bg-slate-950/45">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h3 class="truncate font-bold">{{ item.name }}</h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ item.sku }} · {{ app.t('pos.stock') }} {{ t('pos.stockLine', { stock: item.stock, unit: item.unit }) }}</p>
          </div>
          <button class="grid h-9 w-9 shrink-0 place-items-center rounded-xl text-red-600 hover:bg-red-50 dark:text-red-300 dark:hover:bg-red-500/15" :aria-label="app.t('pos.remove')" @click="cart.removeItem(item.productId)">
            <AppIcon name="x" :size="18" />
          </button>
        </div>

        <div class="mt-3 grid gap-3 sm:grid-cols-[1fr_150px] sm:items-end">
          <div>
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ app.t('pos.unitPrice') }}</p>
            <p class="text-sm font-bold">{{ money(item.price) }}</p>
            <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ app.t('pos.lineTotal') }}</p>
            <p class="font-bold">{{ money(item.price * item.quantity) }}</p>
          </div>
          <div>
            <p class="mb-1.5 text-sm font-semibold text-slate-700 dark:text-slate-200">{{ app.t('pos.qty') }}</p>
            <div class="grid grid-cols-[42px_1fr_42px] overflow-hidden rounded-xl border border-slate-200 bg-white/80 dark:border-slate-700 dark:bg-slate-950/80">
              <button class="min-h-11 text-lg font-black hover:bg-slate-100 dark:hover:bg-slate-800" @click="setQuantity(item.productId, item.quantity - 1)">-</button>
              <input
                class="pos-quantity-input min-h-11 min-w-0 border-x border-slate-200 bg-transparent text-center font-black outline-none dark:border-slate-700"
                type="number"
                :value="item.quantity"
                @input="setQuantity(item.productId, Number(($event.target as HTMLInputElement).value))"
              />
              <button class="min-h-11 text-lg font-black hover:bg-slate-100 dark:hover:bg-slate-800" @click="setQuantity(item.productId, item.quantity + 1)">+</button>
            </div>
          </div>
        </div>
      </article>
    </div>

    <div class="mt-4 grid gap-3 border-t border-slate-200 pt-4 dark:border-slate-700">
      <div class="rounded-2xl bg-brand-50 p-4 dark:bg-emerald-500/10">
        <div class="flex items-center justify-between gap-3 text-sm">
          <span class="font-bold text-brand-700 dark:text-emerald-200">{{ app.t('pos.total') }}</span>
          <span class="text-2xl font-black text-brand-900 dark:text-emerald-100">{{ money(cart.totalAmount) }}</span>
        </div>
      </div>

      <div>
        <p class="mb-2 text-sm font-semibold text-slate-700 dark:text-slate-200">{{ app.t('pos.paymentMethod') }}</p>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="method in paymentMethods"
            :key="method.value"
            class="focus-ring flex min-h-12 items-center justify-center gap-2 rounded-xl border px-3 text-sm font-black transition"
            :class="cart.paymentMethod === method.value ? 'border-brand-600 bg-brand-600 text-white dark:border-emerald-400 dark:bg-emerald-500 dark:text-slate-950' : 'border-slate-200 bg-white/80 text-slate-700 hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-950/70 dark:text-slate-200 dark:hover:bg-slate-800'"
            @click="cart.setPaymentMethod(method.value)"
          >
            <AppIcon :name="method.icon" :size="18" />
            {{ method.label }}
          </button>
        </div>
      </div>

      <div class="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-end">
        <AppInput :label="app.t('pos.receivedAmount')" type="number" :model-value="cart.receivedAmount" @update:model-value="cart.setReceivedAmount(Number($event))" />
        <AppButton variant="secondary" :disabled="cart.items.length === 0 || cart.isSubmitting" @click="setExactPayment">
          {{ app.t('pos.exactAmount') }}
        </AppButton>
      </div>

      <div class="grid grid-cols-2 gap-3 text-sm">
        <div class="rounded-lg bg-slate-50 p-3 dark:bg-slate-950/60">
          <p class="text-slate-500 dark:text-slate-400">{{ app.t('pos.change') }}</p>
          <p class="text-xl font-black">{{ money(cart.changeAmount) }}</p>
        </div>
        <div class="rounded-lg bg-slate-50 p-3 dark:bg-slate-950/60">
          <p class="text-slate-500 dark:text-slate-400">{{ app.t('pos.status') }}</p>
          <AppBadge :tone="checkoutStatus.tone">{{ checkoutStatus.label }}</AppBadge>
        </div>
      </div>

      <p v-if="insufficientPayment" class="text-sm font-semibold text-red-600 dark:text-red-300">{{ app.t('pos.paymentNotEnough') }}</p>
      <AppButton class="w-full" :disabled="props.submitDisabled" :loading="cart.isSubmitting" @click="$emit('submitSale')">
        {{ props.submitDisabled ? checkoutStatus.label : app.t('pos.confirmSale') }}
      </AppButton>
    </div>
  </AppCard>

  <ConfirmDialog
    :open="clearConfirmOpen"
    :title="app.t('pos.clearCartQuestion')"
    :message="app.t('pos.clearCartMessage')"
    :confirm-label="app.t('pos.confirmClearCart')"
    :cancel-label="app.t('pos.cancel')"
    @close="clearConfirmOpen = false"
    @confirm="confirmClearCart"
  />
</template>

<style scoped>
.pos-cart-item {
  animation: pos-cart-in 180ms ease both;
}

.pos-quantity-input::-webkit-outer-spin-button,
.pos-quantity-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.pos-quantity-input {
  -moz-appearance: textfield;
}

@keyframes pos-cart-in {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .pos-cart-item {
    animation: none;
  }
}
</style>
