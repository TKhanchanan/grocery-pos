import type { TranslationKey } from '../i18n'
import type { IconName } from './icons'

export interface NavigationItem {
  labelKey: TranslationKey
  to: string
  roles?: Role[]
  permission?: PermissionCode
  permissions?: PermissionCode[]
  icon?: IconName
}

export type Role = 'ADMIN' | 'MANAGER' | 'CASHIER'
export type PermissionCode = string

export interface AssignedRole {
  id: number
  code: string
  name: string
}

export interface User {
  id: number
  username: string
  fullName: string
  role: Role
  roles?: AssignedRole[]
  active: boolean
  avatar_url: string
  avatar_updated_at?: string
  createdAt: string
}

export interface AuthMeResponse {
  user: User
  roles: AssignedRole[]
  permissions: PermissionCode[]
}

export interface RoleRecord {
  id: number
  code: string
  name: string
  description: string
  is_system: boolean
  is_active: boolean
  permission_count: number
  user_count: number
  created_at: string
  updated_at: string
}

export interface PermissionRecord {
  id: number
  code: string
  module: string
  action: string
  name: string
  description: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface Category {
  id: number
  name: string
  description: string
  is_active: boolean
  count?: number
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
  image_url: string | null
  image_updated_at: string | null
  total_stock: number
  stock_status: StockStatus
  stocks?: ProductStock[]
  created_at: string
}

export interface StockMovement {
  id: number
  product_id: number
  product_name: string
  sku: string
  image_url: string | null
  image_updated_at: string | null
  location_id: number
  location_name: string
  reference_type: 'RESTOCK' | 'ADJUSTMENT' | string
  reference_id: number | null
  quantity_change: number
  before_stock: number
  after_stock: number
  unit_cost: number | null
  note: string
  created_by: number | null
  created_at: string
}

export interface StockTransferItem {
  id?: number
  transfer_id?: number
  product_id: number
  product_name?: string
  sku?: string
  quantity: number
}

export interface StockTransfer {
  id: number
  transfer_no: string
  from_location_id: number
  from_location_name: string
  to_location_id: number
  to_location_name: string
  status: 'DRAFT' | 'COMPLETED' | 'CANCELLED'
  note: string
  created_by: number | null
  completed_at: string | null
  cancelled_at: string | null
  created_at: string
  items: StockTransferItem[]
}

export type PaymentMethod = 'CASH' | 'QR'

export interface POSProduct {
  id: number
  sku: string
  name: string
  barcode: string | null
  category_id: number | null
  category_name: string | null
  image_url: string | null
  image_updated_at: string | null
  selling_price: number
  unit_cost: number
  unit: string
  threshold: number
  reorder_point: number
  location_id: number
  stock: number
  stock_status: StockStatus
}

export interface POSProductPage {
  items: POSProduct[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ReceiptItem {
  id: number
  product_id: number
  product_name: string
  sku: string
  barcode: string | null
  price: number
  cost: number
  quantity: number
  line_total: number
  line_cost: number
}

export interface Receipt {
  id: number
  receipt_no: string
  location_id: number
  location_name: string
  cashier_id: number
  cashier_name: string
  subtotal: number
  total_amount: number
  total_cost: number
  profit: number
  payment_method: PaymentMethod
  paid_amount: number
  change_amount: number
  status: 'COMPLETED' | 'CANCELLED'
  cancelled_by: number | null
  cancelled_at: string | null
  cancel_reason: string
  created_at: string
  items: ReceiptItem[]
}

export type AlertType = 'LOW_STOCK' | 'OUT_OF_STOCK' | 'REORDER_POINT'

export interface InventoryAlert {
  id: number
  product_id: number
  product_name: string
  sku: string
  location_id: number
  location_name: string
  type: AlertType
  message: string
  read_by: number | null
  read_at: string | null
  resolved_at: string | null
  created_at: string
  links: {
    product: string
    restock: string
    purchase_order: string
  }
}

export interface SalesPeriodReport {
  period: string
  receipt_count: number
  revenue: number
  cost: number
  profit: number
}

export interface ProductSalesReport {
  product_id: number
  product_name: string
  sku: string
  image_url: string | null
  image_updated_at: string | null
  quantity: number
  revenue: number
  cost: number
  profit: number
}

export interface StockReport {
  product_id: number
  product_name: string
  sku: string
  image_url: string | null
  image_updated_at: string | null
  location_id: number
  location_name: string
  quantity: number
  unit_cost: number
  total_value: number
  threshold: number
  reorder_point: number
  stock_status: StockStatus
}

export interface PaymentSummaryReport {
  payment_method: PaymentMethod
  receipt_count: number
  revenue: number
}

export interface DashboardSummary {
  today_sales: number
  today_receipts: number
  gross_profit_this_month: number
  top_product_this_month: ProductSalesReport | null
  low_stock_count: number
  out_of_stock_count: number
  reorder_count: number
  payment_method_summary: PaymentSummaryReport[]
  recent_sales: Receipt[]
  sales_trend: SalesPeriodReport[]
  low_stock_items: StockReport[]
  top_products: ProductSalesReport[]
}

export interface ProductImportRawData {
  sku: string
  name: string
  barcode: string
  category: string
  selling_price: number
  unit_cost: number
  threshold: number
  reorder_point: number
  location: string
  initial_stock: number | null
}

export interface ImportJobRow {
  id: number
  import_job_id: number
  row_index: number
  raw_data: ProductImportRawData
  status: 'PENDING' | 'IMPORTED' | 'FAILED'
  error_message: string
  created_at: string
}

export interface ImportJob {
  id: number
  job_type: string
  file_name: string
  status: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'
  total_rows: number
  success_rows: number
  failed_rows: number
  created_by: number | null
  started_at: string | null
  completed_at: string | null
  created_at: string
  rows?: ImportJobRow[]
}

export interface Supplier {
  id: number
  name: string
  phone: string
  email: string
  address: string
  is_active: boolean
  created_at: string
}

export interface PurchaseOrderItem {
  id?: number
  po_id?: number
  product_id: number
  product_name?: string
  sku?: string
  image_url?: string | null
  image_updated_at?: string | null
  quantity: number
  received_quantity: number
  unit_cost: number
  line_cost: number
}

export interface PurchaseOrder {
  id: number
  po_number: string
  supplier_id: number
  supplier_name: string
  location_id: number
  location_name: string
  status: 'DRAFT' | 'SENT' | 'RECEIVED' | 'CANCELLED'
  total_cost: number
  note: string
  created_by: number | null
  received_by: number | null
  cancelled_by: number | null
  received_at: string | null
  cancelled_at: string | null
  created_at: string
  items: PurchaseOrderItem[]
}

export interface AppSettings {
  shop_name: string
  shop_phone: string
  shop_address: string
  default_location_id: number
  receipt_footer: string
  line_enabled: boolean
  line_token_masked: string
  line_configured: boolean
  line_target_id: string
}

export interface ReceiptSettings {
  shop_name: string
  shop_phone: string
  shop_address: string
  receipt_footer: string
}

export interface LineSettings {
  line_enabled: boolean
  line_token?: string
  line_token_masked: string
  line_configured: boolean
  line_target_id: string
}

export interface NotificationLog {
  id: number
  channel: string
  recipient: string
  event_type: string
  payload: string
  status: 'PENDING' | 'SENT' | 'FAILED' | 'SKIPPED'
  error_message: string
  sent_at: string | null
  created_at: string
}
