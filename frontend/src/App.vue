<script setup>
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'

const { isLoggedIn, user, clearAuth } = useUserStore()
const router = useRouter()

const logout = () => {
  clearAuth()
  router.push('/login')
}
</script>

<template>
  <div class="app">
    <nav class="navbar">
      <div class="container nav-content">
        <router-link to="/" class="logo">📝 个人博客</router-link>
        <div class="nav-links">
          <router-link to="/">首页</router-link>
          <router-link v-if="isLoggedIn" to="/create">写文章</router-link>
          <template v-if="!isLoggedIn">
            <router-link to="/login">登录 / 注册</router-link>
          </template>
          <template v-else>
            <router-link v-if="user?.role === 'admin'" to="/admin">管理后台</router-link>
            <router-link to="/profile" class="nickname">{{ user?.nickname || user?.username }}</router-link>
            <a href="#" @click.prevent="logout">退出</a>
          </template>
        </div>
      </div>
    </nav>

    <main class="container main-content">
      <router-view />
    </main>

    <footer class="footer">
      <div class="container">
        <p>© 2026 个人博客 - Powered by Gin + Vue 3</p>
      </div>
    </footer>
  </div>
</template>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f5f5f5;
  color: #333;
  line-height: 1.6;
}

.container {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 16px;
}

.navbar {
  background-color: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: #2c3e50;
  text-decoration: none;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 20px;
}

.nav-links a {
  color: #555;
  text-decoration: none;
  font-size: 15px;
  transition: color 0.2s;
}

.nav-links a:hover,
.nav-links a.router-link-active {
  color: #3498db;
}

.nickname {
  color: #3498db;
  font-weight: 500;
}

.main-content {
  min-height: calc(100vh - 120px);
  padding-top: 24px;
  padding-bottom: 40px;
}

.footer {
  background-color: #fff;
  border-top: 1px solid #eee;
  padding: 20px 0;
  text-align: center;
  color: #999;
  font-size: 14px;
}

.card {
  background-color: #fff;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn:hover {
  opacity: 0.9;
}

.btn-primary {
  background-color: #3498db;
  color: #fff;
}

.btn-danger {
  background-color: #e74c3c;
  color: #fff;
}

.btn-success {
  background-color: #2ecc71;
  color: #fff;
}

.btn:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

input,
textarea,
select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

input:focus,
textarea:focus,
select:focus {
  border-color: #3498db;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #555;
}

.error-msg {
  color: #e74c3c;
  font-size: 14px;
  margin-top: 4px;
}

.empty-tip {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}
</style>
