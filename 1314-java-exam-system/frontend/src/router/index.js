import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/exam-list'
  },
  {
    path: '/exam-list',
    name: 'ExamList',
    component: () => import('../views/ExamList.vue')
  },
  {
    path: '/exam/:examRecordId',
    name: 'Exam',
    component: () => import('../views/Exam.vue')
  },
  {
    path: '/report/:examRecordId',
    name: 'Report',
    component: () => import('../views/Report.vue')
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('../views/Admin.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
