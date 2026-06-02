import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
    { path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue'), meta: { auth: true } },
    { path: '/products', name: 'products', component: () => import('../views/ProductsView.vue'), meta: { auth: true } },
    { path: '/inventory', name: 'inventory', component: () => import('../views/InventoryView.vue'), meta: { auth: true } },
    { path: '/pos', name: 'pos', component: () => import('../views/POSView.vue'), meta: { auth: true } },
    { path: '/sales', name: 'sales', component: () => import('../views/SalesView.vue'), meta: { auth: true } },
    { path: '/reports', name: 'reports', component: () => import('../views/ReportsView.vue'), meta: { auth: true } },
    { path: '/suppliers', name: 'suppliers', component: () => import('../views/SuppliersPOView.vue'), meta: { auth: true } },
    { path: '/settings', name: 'settings', component: () => import('../views/SettingsUsersView.vue'), meta: { auth: true } },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.isLoggedIn) return '/login'
  if (to.path === '/login' && auth.isLoggedIn) return '/'
  return true
})

export default router
