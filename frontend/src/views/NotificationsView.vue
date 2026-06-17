<script setup>
import { ref, onMounted } from 'vue'
import { getNotifications, markAsRead } from '@/api/notification'

const notifications = ref([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)

const fetchNotifications = async () => {
  loading.value = true
  try {
    const res = await getNotifications({ page: page.value, page_size: 20 })
    notifications.value = res.data?.data?.data || []
    total.value = res.data?.data?.total || 0
  } catch (e) {
    // 静默失败
  } finally {
    loading.value = false
  }
}

const handleMarkAsRead = async (id) => {
  try {
    await markAsRead(id)
    const item = notifications.value.find(n => n.id === id)
    if (item) item.is_read = true
  } catch (e) {
    alert('标记失败')
  }
}

onMounted(fetchNotifications)
</script>

<template>
  <div class="notifications-page">
    <h2 class="page-title">通知中心</h2>
    <div v-if="loading" class="empty-tip card">加载中...</div>
    <div v-else-if="notifications.length === 0" class="empty-tip card">暂无通知</div>
    <div v-else class="notification-list">
      <div
        v-for="item in notifications"
        :key="item.id"
        :class="['notification-item card', { unread: !item.is_read }]"
      >
        <div class="notification-content">
          <h4 class="notification-title">{{ item.title }}</h4>
          <p class="notification-body">{{ item.content }}</p>
          <span class="notification-time">{{ new Date(item.created_at).toLocaleString() }}</span>
        </div>
        <button
          v-if="!item.is_read"
          class="btn btn-primary"
          @click="handleMarkAsRead(item.id)"
        >
          标记已读
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.notifications-page {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  margin-bottom: 20px;
  color: #2c3e50;
}

.notification-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.notification-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  padding: 16px 20px;
  border-left: 4px solid transparent;
}

.notification-item.unread {
  border-left-color: #3498db;
  background-color: #f8fbff;
}

.notification-title {
  font-size: 16px;
  margin-bottom: 6px;
  color: #2c3e50;
}

.notification-body {
  color: #666;
  font-size: 14px;
  margin-bottom: 8px;
}

.notification-time {
  color: #999;
  font-size: 12px;
}

.notification-item .btn {
  flex-shrink: 0;
}
</style>
