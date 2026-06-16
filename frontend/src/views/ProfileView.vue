<script setup>
import { ref, onMounted } from 'vue'
import { authApi, badgeApi } from '@/api'
import { useUserStore } from '@/stores/user'

const { user, setAuth } = useUserStore()

const form = ref({
  nickname: user.value?.nickname || '',
  email: user.value?.email || '',
  avatar: user.value?.avatar || ''
})

const loading = ref(false)
const uploading = ref(false)
const message = ref('')
const error = ref('')
const badges = ref([])

const fetchProfile = async () => {
  try {
    const res = await authApi.me()
    form.value = {
      nickname: res.nickname || '',
      email: res.email || '',
      avatar: res.avatar || ''
    }
    // 同步更新本地存储的用户信息
    const token = localStorage.getItem('token')
    if (token) setAuth(token, res)
  } catch (err) {
    error.value = err.message || '获取资料失败'
  }
}

const fetchBadges = async () => {
  try {
    badges.value = await badgeApi.getMyBadges()
  } catch (err) {
    console.error('加载勋章失败', err)
  }
}

const uploadAvatar = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  uploading.value = true
  error.value = ''
  message.value = ''
  try {
    const formData = new FormData()
    formData.append('file', file)

    const res = await fetch('/api/uploads', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    })
    const data = await res.json()
    if (data.code !== 0) {
      throw new Error(data.message || '上传失败')
    }
    form.value.avatar = data.data.url
    message.value = '头像上传成功'
  } catch (err) {
    error.value = err.message || '头像上传失败'
  } finally {
    uploading.value = false
    // 清空 input，允许重复选择同一文件
    event.target.value = ''
  }
}

const submit = async () => {
  loading.value = true
  error.value = ''
  message.value = ''
  try {
    const res = await authApi.updateProfile(form.value)
    const token = localStorage.getItem('token')
    if (token) setAuth(token, res)
    message.value = '资料更新成功'
  } catch (err) {
    error.value = err.message || '更新失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchProfile()
  fetchBadges()
})
</script>

<template>
  <div class="profile">
    <div class="card">
      <h2>👤 个人资料</h2>

      <div class="avatar-section">
        <img
          :src="form.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=' + (user?.username || 'user')"
          alt="avatar"
          class="avatar"
        />
        <div class="avatar-actions">
          <label class="btn btn-primary" :class="{ disabled: uploading }">
            {{ uploading ? '上传中...' : '更换头像' }}
            <input type="file" accept="image/*" hidden @change="uploadAvatar" />
          </label>
          <p class="avatar-tip">支持 jpg/png/gif/webp，最大 10MB</p>
        </div>
      </div>

      <form @submit.prevent="submit">
        <div class="form-group">
          <label>用户名</label>
          <input :value="user?.username" type="text" disabled />
        </div>
        <div class="form-group">
          <label>昵称</label>
          <input v-model="form.nickname" type="text" placeholder="请输入昵称" />
        </div>
        <div class="form-group">
          <label>邮箱</label>
          <input v-model="form.email" type="email" placeholder="请输入邮箱" />
        </div>

        <p v-if="message" class="success-msg">{{ message }}</p>
        <p v-if="error" class="error-msg">{{ error }}</p>

        <button type="submit" class="btn btn-success" :disabled="loading">
          {{ loading ? '保存中...' : '保存资料' }}
        </button>
      </form>

      <div class="badges-section">
        <h3>🏅 我的勋章</h3>
        <div v-if="badges.length === 0" class="empty-badges">
          暂无勋章，多发优质文章或参与互动可获得勋章～
        </div>
        <div v-else class="badges-grid">
          <div v-for="item in badges" :key="item.id" class="badge-card" :title="item.reason || item.badge.description">
            <img v-if="item.badge.icon_url" :src="item.badge.icon_url" alt="badge" />
            <span v-else class="badge-emoji">🏅</span>
            <span class="badge-name">{{ item.badge.name }}</span>
            <span v-if="item.reason" class="badge-reason">{{ item.reason }}</span>
            <span v-if="item.badge.contract_address" class="nft-badge">NFT</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile {
  max-width: 560px;
  margin: 0 auto;
}

.profile h2 {
  margin-bottom: 24px;
  color: #333;
}

.avatar-section {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #eee;
}

.avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #eee;
}

.avatar-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.avatar-actions label {
  cursor: pointer;
  width: fit-content;
}

.avatar-actions label.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.avatar-tip {
  font-size: 12px;
  color: #999;
}

.success-msg {
  color: #2ecc71;
  font-size: 14px;
  margin-bottom: 12px;
}

.profile .btn {
  width: 100%;
  padding: 12px;
  font-size: 16px;
}

.badges-section {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #eee;
}

.badges-section h3 {
  margin-bottom: 16px;
  color: #444;
}

.empty-badges {
  color: #999;
  text-align: center;
  padding: 20px 0;
}

.badges-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 16px;
}

.badge-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 12px;
  background-color: #f9f9f9;
  border-radius: 12px;
  text-align: center;
  position: relative;
}

.badge-card img {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  object-fit: cover;
}

.badge-emoji {
  font-size: 40px;
}

.badge-name {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.badge-reason {
  font-size: 12px;
  color: #888;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.nft-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  background-color: #8e44ad;
  color: #fff;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
