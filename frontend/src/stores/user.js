import { reactive, computed } from 'vue'

// 简单的全局用户状态（无需 pinia，减少依赖）
const state = reactive({
  token: localStorage.getItem('token') || '',
  user: JSON.parse(localStorage.getItem('user') || 'null')
})

export const useUserStore = () => {
  const isLoggedIn = computed(() => !!state.token)
  const user = computed(() => state.user)

  const setAuth = (token, userInfo) => {
    state.token = token
    state.user = userInfo
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(userInfo))
  }

  const clearAuth = () => {
    state.token = ''
    state.user = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    isLoggedIn,
    user,
    setAuth,
    clearAuth
  }
}
