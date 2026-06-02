export interface NavigationItem {
  label: string
  to: string
  roles?: Role[]
}

export type Role = 'ADMIN' | 'MANAGER' | 'CASHIER'

export interface User {
  id: number
  username: string
  fullName: string
  role: Role
  active: boolean
  createdAt: string
}

export interface Category {
  id: number
  name: string
  description: string
  is_active: boolean
  created_at: string
}

export interface Location {
  id: number
  name: string
  description: string
  is_active: boolean
  created_at: string
}

export type StockStatus = 'in_stock' | 'low_stock' | 'out_of_stock' | 'reorder_point'

export interface ProductStock {
  product_id: number
  product_name: string
  sku: string
  location_id: number
  location_name: string
  quantity: number
  stock_status: StockStatus
}

export interface Product {
  id: number
  sku: string
  name: string
  barcode: string | null
  category_id: number | null
  category_name: string | null
  selling_price: number
  unit_cost: number
  unit: string
  threshold: number
  reorder_point: number
  is_active: boolean
  total_stock: number
  stock_status: StockStatus
  stocks?: ProductStock[]
  created_at: string
}
