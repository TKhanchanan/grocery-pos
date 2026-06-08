import { apiClient } from '../api/client'
import type { ReceiptSettings } from '../types/navigation'

export const defaultReceiptSettings: ReceiptSettings = {
  shop_name: 'Grocery POS',
  shop_phone: '',
  shop_address: '',
  receipt_footer: '',
}

export function loadReceiptSettings() {
  return apiClient<ReceiptSettings>('/v1/receipt-settings')
}
