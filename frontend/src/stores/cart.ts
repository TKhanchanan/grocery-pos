import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { Product } from '../types'

export interface CartLine {
  product: Product
  quantity: number
}

export const useCartStore = defineStore('cart', () => {
  const lines = ref<CartLine[]>([])
  const total = computed(() => lines.value.reduce((sum, line) => sum + line.product.price * line.quantity, 0))

  function add(product: Product) {
    const existing = lines.value.find((line) => line.product.id === product.id)
    if (existing) existing.quantity += 1
    else lines.value.push({ product, quantity: 1 })
  }

  function remove(productId: number) {
    lines.value = lines.value.filter((line) => line.product.id !== productId)
  }

  function setQty(productId: number, quantity: number) {
    const line = lines.value.find((item) => item.product.id === productId)
    if (!line) return
    line.quantity = Math.max(1, quantity)
  }

  function clear() {
    lines.value = []
  }

  return { lines, total, add, remove, setQty, clear }
})
