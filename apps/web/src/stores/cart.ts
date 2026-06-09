import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { PaymentMethod, POSProduct } from '../types/navigation'

export interface CartItem {
  productId: number
  sku: string
  name: string
  barcode: string | null
  unit: string
  price: number
  cost: number
  stock: number
  quantity: number
}

function roundMoney(value: number) {
  return Math.round(value * 100) / 100
}

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])
  const receivedAmount = ref(0)
  const paymentMethod = ref<PaymentMethod>('CASH')
  const isSubmitting = ref(false)

  const totalAmount = computed(() => roundMoney(items.value.reduce((sum, item) => sum + item.price * item.quantity, 0)))
  const totalItems = computed(() => items.value.reduce((sum, item) => sum + item.quantity, 0))
  const changeAmount = computed(() => roundMoney(Math.max(0, receivedAmount.value - totalAmount.value)))

  function addItem(product: POSProduct) {
    if (product.stock <= 0) return
    const existing = items.value.find((item) => item.productId === product.id)
    if (existing) {
      updateQuantity(product.id, existing.quantity + 1)
      return
    }
    items.value.push({
      productId: product.id,
      sku: product.sku,
      name: product.name,
      barcode: product.barcode,
      unit: product.unit,
      price: product.selling_price,
      cost: product.unit_cost,
      stock: product.stock,
      quantity: 1,
    })
  }

  function updateQuantity(productId: number, quantity: number) {
    const item = items.value.find((line) => line.productId === productId)
    if (!item) return
    const nextQuantity = Math.trunc(Number(quantity || 0))
    if (nextQuantity <= 0) {
      removeItem(productId)
      return
    }
    item.quantity = Math.min(nextQuantity, item.stock)
  }

  function removeItem(productId: number) {
    items.value = items.value.filter((item) => item.productId !== productId)
  }

  function clearCart() {
    items.value = []
    receivedAmount.value = 0
    paymentMethod.value = 'CASH'
    isSubmitting.value = false
  }

  function setReceivedAmount(value: number) {
    receivedAmount.value = Math.max(0, Number(value || 0))
  }

  function setPaymentMethod(value: PaymentMethod) {
    paymentMethod.value = value
    if (value === 'QR') {
      receivedAmount.value = totalAmount.value
    }
  }

  function refreshStock(products: POSProduct[]) {
    for (const item of items.value) {
      const product = products.find((candidate) => candidate.id === item.productId)
      if (!product) continue
      item.stock = product.stock
      item.price = product.selling_price
      if (item.quantity > product.stock) item.quantity = product.stock
    }
    items.value = items.value.filter((item) => item.stock > 0 && item.quantity > 0)
  }

  return {
    items,
    receivedAmount,
    paymentMethod,
    isSubmitting,
    totalAmount,
    totalItems,
    changeAmount,
    addItem,
    updateQuantity,
    removeItem,
    clearCart,
    setReceivedAmount,
    setPaymentMethod,
    refreshStock,
  }
})
