import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import type { PermissionCode, Role } from '../types/navigation'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', name: 'login', component: () => import('../pages/LoginPage.vue'), meta: { public: true, layout: 'auth' } },
    { path: '/forbidden', name: 'forbidden', component: () => import('../pages/ForbiddenPage.vue') },
    { path: '/dashboard', name: 'dashboard', component: () => import('../pages/DashboardPage.vue'), meta: { permission: 'dashboard.view' } },
    { path: '/pos', name: 'pos', component: () => import('../pages/POSPage.vue'), meta: { permission: 'pos.view' } },
    { path: '/products', name: 'products', component: () => import('../pages/ProductsPage.vue'), meta: { permission: 'products.view' } },
    { path: '/categories', name: 'categories', component: () => import('../pages/CategoriesPage.vue'), meta: { permission: 'categories.view' } },
    { path: '/stock-operations', name: 'stock-operations', component: () => import('../pages/StockOperationsPage.vue'), meta: { permissions: ['stock.restock', 'stock.adjust', 'stock.movements.view'] } },
    { path: '/restock', redirect: '/stock-operations?tab=restock' },
    { path: '/stock-movements', redirect: '/stock-operations?tab=movements' },
    { path: '/inventory-management', name: 'inventory-management', component: () => import('../pages/InventoryManagementPage.vue'), meta: { permissions: ['locations.view', 'transfers.view'] } },
    { path: '/locations', redirect: '/inventory-management?tab=locations' },
    { path: '/transfers', redirect: '/inventory-management?tab=transfers' },
    { path: '/sales-history', name: 'sales-history', component: () => import('../pages/SalesHistoryPage.vue'), meta: { permission: 'sales.view' } },
    { path: '/receipt-detail', name: 'receipt-detail', component: () => import('../pages/ReceiptDetailPage.vue'), meta: { permission: 'sales.receipt.view' } },
    { path: '/sales/:id/receipt', name: 'sale-receipt', component: () => import('../pages/ReceiptDetailPage.vue'), meta: { permission: 'sales.receipt.view' } },
    { path: '/alerts', name: 'alerts', component: () => import('../pages/AlertsPage.vue'), meta: { permission: 'alerts.view' } },
    { path: '/reports', name: 'reports', component: () => import('../pages/ReportsPage.vue'), meta: { permission: 'reports.view' } },
    { path: '/exports', name: 'exports', component: () => import('../pages/ExportsPage.vue'), meta: { permission: 'exports.view' } },
    { path: '/imports', name: 'imports', component: () => import('../pages/ImportsPage.vue'), meta: { permission: 'imports.view' } },
    { path: '/purchase-orders', name: 'purchase-orders', component: () => import('../pages/PurchaseOrdersPage.vue'), meta: { permission: 'purchase_orders.view' } },
    { path: '/suppliers', name: 'suppliers', component: () => import('../pages/SuppliersPage.vue'), meta: { permission: 'suppliers.view' } },
    { path: '/users', name: 'users', component: () => import('../pages/UsersPage.vue'), meta: { permission: 'users.view' } },
    { path: '/roles', name: 'roles', component: () => import('../pages/RolesPage.vue'), meta: { permission: 'roles.view' } },
    { path: '/settings', name: 'settings', component: () => import('../pages/SettingsPage.vue'), meta: { permission: 'settings.view' } },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    return auth.isAuthenticated && to.path === '/login' ? '/dashboard' : true
  }
  if (!auth.isAuthenticated) return '/login'
  if (auth.permissions.length === 0) {
    await auth.loadMe().catch(() => undefined)
  }

  const permission = to.meta.permission as PermissionCode | undefined
  const permissions = to.meta.permissions as PermissionCode[] | undefined
  const roles = to.meta.roles as Role[] | undefined
  if (permissions && !auth.hasAnyPermission(permissions)) return '/forbidden'
  if (permission ? !auth.hasPermission(permission) : !auth.can(roles)) return '/forbidden'
  return true
})

export default router
