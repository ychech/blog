import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import PostDetailView from '@/views/PostDetailView.vue'
import LoginView from '@/views/LoginView.vue'
import CreatePostView from '@/views/CreatePostView.vue'
import EditPostView from '@/views/EditPostView.vue'
import ProfileView from '@/views/ProfileView.vue'
import AdminView from '@/views/AdminView.vue'

const routes = [
  { path: '/', name: 'Home', component: HomeView },
  { path: '/posts/:id', name: 'PostDetail', component: PostDetailView, props: true },
  { path: '/login', name: 'Login', component: LoginView },
  { path: '/create', name: 'CreatePost', component: CreatePostView, meta: { requiresAuth: true } },
  { path: '/posts/:id/edit', name: 'EditPost', component: EditPostView, meta: { requiresAuth: true }, props: true },
  { path: '/profile', name: 'Profile', component: ProfileView, meta: { requiresAuth: true } },
  { path: '/admin', name: 'Admin', component: AdminView, meta: { requiresAuth: true, requiresAdmin: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫：拦截需要登录/管理员权限的页面
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const user = JSON.parse(localStorage.getItem('user') || 'null')

  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }

  if (to.meta.requiresAdmin && user?.role !== 'admin') {
    next('/')
    return
  }

  next()
})

export default router
