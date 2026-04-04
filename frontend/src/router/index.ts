import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/survey/:language',
      name: 'survey',
      component: () => import('@/views/SurveyView.vue'),
      props: true,
    },
    {
      path: '/contribute/:language',
      name: 'contribute',
      component: () => import('@/views/ContributeView.vue'),
      props: true,
    },
  ],
})

export default router
