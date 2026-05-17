import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/RegisterView.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/position',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'strategy',
        name: 'Strategy',
        component: () => import('@/views/StrategyView.vue'),
        meta: { title: '策略管理' },
      },
      {
        path: 'strategy/editor',
        name: 'StrategyEditor',
        component: () => import('@/views/StrategyEditorView.vue'),
        meta: { title: '策略编辑器' },
      },
      {
        path: 'backtest',
        name: 'Backtest',
        component: () => import('@/views/BacktestView.vue'),
        meta: { title: '回测' },
      },
      {
        path: 'position',
        name: 'Position',
        component: () => import('@/views/PositionView.vue'),
        meta: { title: '持仓' },
      },
      {
        path: 'market',
        name: 'Market',
        component: () => import('@/views/MarketView.vue'),
        meta: { title: '行情' },
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/OrdersView.vue'),
        meta: { title: '订单' },
      },
    ],
  },
  // 移动端路由
  {
    path: '/mobile',
    component: () => import('@/layouts/MobileLayout.vue'),
    redirect: '/mobile/position',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'position',
        name: 'MobilePosition',
        component: () => import('@/views/mobile/Position.vue'),
        meta: { title: '持仓' },
      },
      {
        path: 'orders',
        name: 'MobileOrders',
        component: () => import('@/views/mobile/Orders.vue'),
        meta: { title: '订单' },
      },
      {
        path: 'backtest',
        name: 'MobileBacktest',
        component: () => import('@/views/mobile/Backtest.vue'),
        meta: { title: '回测' },
      },
      {
        path: 'strategy',
        name: 'MobileStrategy',
        component: () => import('@/views/mobile/Strategy.vue'),
        meta: { title: '策略' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 权限守卫：未登录跳转 /login
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('zq_token')
  if (to.meta.requiresAuth !== false && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
