import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/auth/callback',
      name: 'auth-callback',
      component: () => import('@/views/AuthCallbackView.vue'),
      meta: { public: true },
    },
    {
      path: '/join/:token',
      name: 'join',
      component: () => import('@/views/JoinView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      redirect: '/boloes',
    },
    {
      path: '/boloes',
      name: 'bolao-list',
      component: () => import('@/views/BolaoListView.vue'),
    },
    {
      path: '/boloes/novo',
      name: 'bolao-create',
      component: () => import('@/views/BolaoCreateView.vue'),
    },
    {
      path: '/boloes/:id',
      name: 'bolao',
      component: () => import('@/views/BolaoView.vue'),
    },
    {
      path: '/boloes/:id/ranking',
      name: 'ranking',
      component: () => import('@/views/RankingView.vue'),
    },
    {
      path: '/como-funciona',
      name: 'como-funciona',
      component: () => import('@/views/ComoFuncionaView.vue'),
      meta: { public: true },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'login' }
  }
  if (to.name === 'login' && auth.isAuthenticated) {
    return { path: '/' }
  }
})

export default router
