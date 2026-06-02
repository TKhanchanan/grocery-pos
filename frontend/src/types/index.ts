export type Role = 'ADMIN' | 'MANAGER' | 'CASHIER'

export interface User {
  id: number
  username: string
  fullName: string
  role: Role
  active: boolean
  createdAt: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface Category {
  id: number
  name: string
}

export interface Product {
  id: number
  categoryId: number | null
  sku: string
  barcode: string | null
  name: string
  unit: string
  price: number
  cost: number
  threshold: number
  reorderPoint: number
  active: boolean
  totalStock: number
}

export interface Location {
  id: number
  name: string
  active: boolean
}

export interface ProductStock {
  productId: number
  locationId: number
  productName: string
  sku: string
  locationName: string
  quantity: number
}

export interface Alert {
  id: number
  productId: number
  locationId: number
  type: 'LOW_STOCK' | 'OUT_OF_STOCK' | 'REORDER_POINT'
  message: string
  productName: string
  locationName: string
  currentStock: number
}

export interface StockMovement {
  id: number
  productName: string
  locationName: string
  referenceType: string
  quantityChange: number
  note: string
  createdAt: string
}

export interface SaleItem {
  id: number
  productId: number
  productNameSnapshot: string
  skuSnapshot: string
  priceSnapshot: number
  costSnapshot: number
  quantity: number
  lineTotal: number
}

export interface Sale {
  id: number
  receiptNo: string
  locationId: number
  locationName: string
  totalAmount: number
  profit: number
  paymentMethod: string
  paidAmount: number
  changeAmount: number
  status: 'COMPLETED' | 'CANCELLED'
  createdAt: string
  items: SaleItem[]
}

export interface Supplier {
  id: number
  name: string
  phone: string
  email: string
  address: string
}

export interface PurchaseOrderItem {
  productId: number
  product: string
  quantity: number
  unitCost: number
  lineCost: number
}

export interface PurchaseOrder {
  id: number
  poNumber: string
  supplierId: number
  supplier: string
  locationId: number
  location: string
  status: 'OPEN' | 'RECEIVED' | 'CANCELLED'
  totalCost: number
  createdAt: string
  items: PurchaseOrderItem[]
}

export interface DashboardSummary {
  revenue: number
  profit: number
  salesCount: number
  itemsSold: number
  inventoryValue: number
  lowAlerts: number
  outAlerts: number
}

export type ReportRow = Record<string, string | number | boolean | null>
