import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiClient } from '../api/client'
import type { Category, Location, ReceiptSettings } from '../types/navigation'
import { defaultReceiptSettings } from '../utils/receiptSettings'

export const useReferenceDataStore = defineStore('referenceData', () => {
  const locations = ref<Location[]>([])
  const categories = ref<Category[]>([])
  const posCategories = ref<Category[]>([])
  const receiptSettings = ref<ReceiptSettings>({ ...defaultReceiptSettings })

  async function loadLocations() {
    locations.value = await apiClient<Location[]>('/v1/locations')
    return locations.value
  }

  async function loadCategories() {
    categories.value = await apiClient<Category[]>('/v1/categories')
    return categories.value
  }

  async function loadPOSCategories() {
    posCategories.value = await apiClient<Category[]>('/v1/pos/categories')
    return posCategories.value
  }

  async function loadReceiptSettings() {
    receiptSettings.value = await apiClient<ReceiptSettings>('/v1/receipt-settings')
    return receiptSettings.value
  }

  return {
    locations,
    categories,
    posCategories,
    receiptSettings,
    loadLocations,
    loadCategories,
    loadPOSCategories,
    loadReceiptSettings,
  }
})
