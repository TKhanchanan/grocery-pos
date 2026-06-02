import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import type { Role } from '../types/navigation'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', name: 'login', component: () => import('../pages/LoginPage.vue'), meta: { public: true, layout: 'auth' } },
    { path: '/forbidden', name: 'forbidden', component: () => import('../pages/ForbiddenPage.vue') },
    { path: '/dashboard', name: 'dashboard', component: () => import('../pages/DashboardPage.vue') },
    { path: '/pos', name: 'pos', component: () => import('../pages/POSPage.vue'), meta: { roles: ['ADMIN', 'CASHIER'] } },
    { path: '/products', name: 'products', component: () => import('../pages/ProductsPage.vue'), meta: { roles: ['ADMIN', 'MANAGER', 'CASHIER'] } },
    { path: '/categories', name: 'categories', component: () => import('../pages/CategoriesPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/restock', name: 'restock', component: () => import('../pages/RestockPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/stock-movements', name: 'stock-movements', component: () => import('../pages/StockMovementsPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/locations', name: 'locations', component: () => import('../pages/LocationsPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/transfers', name: 'transfers', component: () => import('../pages/TransfersPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/sales-history', name: 'sales-history', component: () => import('../pages/SalesHistoryPage.vue'), meta: { roles: ['ADMIN', 'MANAGER', 'CASHIER'] } },
    { path: '/receipt-detail', name: 'receipt-detail', component: () => import('../pages/ReceiptDetailPage.vue'), meta: { roles: ['ADMIN', 'MANAGER', 'CASHIER'] } },
    { path: '/alerts', name: 'alerts', component: () => import('../pages/AlertsPage.vue'), meta: { roles: ['ADMIN', 'MANAGER', 'CASHIER'] } },
    { path: '/reports', name: 'reports', component: () => import('../pages/ReportsPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/exports', name: 'exports', component: () => import('../pages/ExportsPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/imports', name: 'imports', component: () => import('../pages/ImportsPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/purchase-orders', name: 'purchase-orders', component: () => import('../pages/PurchaseOrdersPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/suppliers', name: 'suppliers', component: () => import('../pages/SuppliersPage.vue'), meta: { roles: ['ADMIN', 'MANAGER'] } },
    { path: '/users', name: 'users', component: () => import('../pages/UsersPage.vue'), meta: { roles: ['ADMIN'] } },
    { path: '/settings', name: 'settings', component: () => import('../pages/SettingsPage.vue'), meta: { roles: ['ADMIN'] } },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    return auth.isAuthenticated && to.path === '/login' ? '/dashboard' : true
  }
  if (!auth.isAuthenticated) return '/login'

  const roles = to.meta.roles as Role[] | undefined
  if (!auth.can(roles)) return '/forbidden'
  return true
})

export default router
