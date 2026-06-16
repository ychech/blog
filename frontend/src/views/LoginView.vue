<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const { setAuth } = useUserStore()

const isLogin = ref(true)
const form = ref({
  username: '',
  password: '',
  nickname: '',
  email: ''
})
const loading = ref(false)
const error = ref('')

const toggleMode = () => {
  isLogin.value = !isLogin.value
  error.value = ''
}

const submit = async () => {
  if (!form.value.username || !form.value.password) {
    error.value = '请填写用户名和密码'
    return
  }

  loading.value = true
  error.value = ''
  try {
    let res
    if (isLogin.value) {
      res = await authApi.login({
        username: form.value.username,
        password: form.value.password
      })
    } else {
      if (!form.value.nickname) form.value.nickname = form.value.username
      res = await authApi.register(form.value)
    }

    setAuth(res.token, res.user)
    router.push('/')
  } catch (err) {
    error.value = err.message || (isLogin.value ? '登录失败' : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <div class="auth-card card">
      <h2>{{ isLogin ? '登录' : '注册' }}</h2>
      <form @submit.prevent="submit">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="form.username" type="text" placeholder="请输入用户名" />
        </div>
        <div class="form-group">
          <label>密码</label>
          <input v-model="form.password" type="password" placeholder="请输入密码" />
        </div>
        <template v-if="!isLogin">
          <div class="form-group">
            <label>昵称</label>
            <input v-model="form.nickname" type="text" placeholder="请输入昵称（可选）" />
          </div>
          <div class="form-group">
            <label>邮箱</label>
            <input v-model="form.email" type="email" placeholder="请输入邮箱（可选）" />
          </div>
        </template>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <button type="submit" class="btn btn-primary" :disabled="loading">
          {{ loading ? '请稍候...' : (isLogin ? '登录' : '注册') }}
        </button>
      </form>
      <p class="switch-mode">
        {{ isLogin ? '还没有账号？' : '已有账号？' }}
        <a href="#" @click.prevent="toggleMode">{{ isLogin ? '立即注册' : '去登录' }}</a>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-container {
  display: flex;
  justify-content: center;
  padding-top: 40px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
}

.auth-card h2 {
  text-align: center;
  margin-bottom: 24px;
  color: #333;
}

.auth-card .btn {
  width: 100%;
  padding: 12px;
  font-size: 16px;
}

.switch-mode {
  text-align: center;
  margin-top: 16px;
  color: #666;
  font-size: 14px;
}

.switch-mode a {
  color: #3498db;
}
</style>
