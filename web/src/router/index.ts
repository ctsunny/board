import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getBoardBaseHref } from '@/utils/runtime'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/components/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/Dashboard.vue') },
      { path: 'customers', name: 'Customers', component: () => import('@/views/Customers.vue') },
      { path: 'live-resources', name: 'LiveResources', component: () => import('@/views/LiveResources.vue') },
      { path: 'regions', name: 'Regions', redirect: { name: 'LiveResources', query: { tab: 'regions' } } },
      { path: 'routes', name: 'Routes', redirect: { name: 'LiveResources', query: { tab: 'routes' } } },
      { path: 'servers', name: 'Servers', redirect: { name: 'LiveResources', query: { tab: 'servers' } } },
      { path: 'nodes', name: 'Nodes', redirect: { name: 'LiveResources', query: { tab: 'nodes' } } },
      { path: 'tokens', name: 'APITokens', component: () => import('@/views/APITokens.vue') },
      { path: 'audit-logs', name: 'AuditLogs', component: () => import('@/views/AuditLogs.vue') },
      { path: 'settings', name: 'Settings', component: () => import('@/views/Settings.vue') },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHashHistory(getBoardBaseHref()),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth !== false && !auth.isLoggedIn) {
    return { name: 'Login' }
  }
  if (to.name === 'Login' && auth.isLoggedIn) {
    return { name: 'Dashboard' }
  }
})

export default router
