import axios from 'axios'

// 创建 axios 实例，所有请求都以 /api 为前缀
const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器：自动附加 JWT Token
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器：统一处理 code != 0 的情况
request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code !== 0) {
      const err = new Error(res.message || '请求失败')
      err.code = res.code
      return Promise.reject(err)
    }
    return res.data
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      // Token 失效时清除登录状态并跳转登录页
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
