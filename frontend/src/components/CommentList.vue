<script setup>
import { ref, computed, watch } from 'vue'
import { commentApi, commentLikeApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'

const props = defineProps({
  postId: {
    type: [String, Number],
    required: true
  },
  comments: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['refresh'])
const { isLoggedIn, user } = useUserStore()
const router = useRouter()

const newComment = ref('')
const submitting = ref(false)
const error = ref('')
const likeStates = ref({})
const likeLoading = ref({})

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 将评论列表构建为树形结构（一级评论 + 回复）
const commentTree = computed(() => {
  const map = {}
  const roots = []
  props.comments.forEach((item) => {
    map[item.id] = { ...item, children: [] }
  })
  props.comments.forEach((item) => {
    if (item.parent_id && map[item.parent_id]) {
      map[item.parent_id].children.push(map[item.id])
    } else {
      roots.push(map[item.id])
    }
  })
  return roots
})

// 当评论数据变化时，初始化点赞状态
watch(
  () => props.comments,
  (comments) => {
    comments.forEach((c) => {
      if (!likeStates.value[c.id]) {
        likeStates.value[c.id] = { count: c.like_count || 0, liked: false }
      } else {
        likeStates.value[c.id].count = c.like_count || 0
      }
    })
    if (isLoggedIn.value) {
      fetchAllLikeStatuses()
    }
  },
  { immediate: true }
)

const fetchAllLikeStatuses = async () => {
  const ids = props.comments.map((c) => c.id)
  if (ids.length === 0) return
  try {
    // 使用批量接口，一次请求获取所有评论的点赞状态，避免 N+1 请求
    const result = await commentLikeApi.batchStatus(ids)
    ids.forEach((id) => {
      likeStates.value[id] = result[id] || { count: props.comments.find((c) => c.id === id)?.like_count || 0, liked: false }
    })
  } catch (err) {
    console.error('加载评论点赞状态失败', err)
  }
}

const toggleLike = async (comment) => {
  if (!isLoggedIn.value) {
    router.push('/login')
    return
  }
  if (likeLoading.value[comment.id]) return
  likeLoading.value[comment.id] = true
  try {
    const res = await commentLikeApi.toggle(comment.id)
    likeStates.value[comment.id] = res
  } catch (err) {
    alert(err.message || '操作失败')
  } finally {
    likeLoading.value[comment.id] = false
  }
}

const submitComment = async () => {
  if (!newComment.value.trim()) return
  submitting.value = true
  error.value = ''
  try {
    await commentApi.create({
      post_id: Number(props.postId),
      content: newComment.value.trim()
    })
    newComment.value = ''
    emit('refresh')
  } catch (err) {
    error.value = err.message || '评论失败'
  } finally {
    submitting.value = false
  }
}

const deleteComment = async (id) => {
  if (!confirm('确定删除这条评论吗？')) return
  try {
    await commentApi.delete(id)
    emit('refresh')
  } catch (err) {
    alert(err.message || '删除失败')
  }
}

const canDelete = (comment) => {
  return user.value && comment.author_id === user.value.id
}
</script>

<template>
  <div class="comment-section">
    <h3 class="section-title">💬 评论（{{ comments.length }}）</h3>

    <div v-if="isLoggedIn" class="comment-form card">
      <div class="form-group">
        <textarea
          v-model="newComment"
          rows="3"
          placeholder="写下你的评论..."
        ></textarea>
        <p v-if="error" class="error-msg">{{ error }}</p>
      </div>
      <button class="btn btn-primary" :disabled="submitting" @click="submitComment">
        {{ submitting ? '提交中...' : '发表评论' }}
      </button>
    </div>
    <div v-else class="login-tip card">
      <router-link to="/login">登录</router-link> 后即可发表评论
    </div>

    <div v-if="commentTree.length === 0" class="empty-tip card">
      暂无评论，来说两句吧～
    </div>

    <div v-else class="comment-list">
      <div v-for="comment in commentTree" :key="comment.id" class="comment-item card">
        <div class="comment-header">
          <span class="comment-author">{{ comment.author_name }}</span>
          <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
          <button
            v-if="canDelete(comment)"
            class="delete-btn"
            @click="deleteComment(comment.id)"
          >
            删除
          </button>
        </div>
        <p class="comment-content">{{ comment.content }}</p>
        <div class="comment-actions">
          <button
            class="like-btn"
            :class="{ liked: likeStates[comment.id]?.liked }"
            :disabled="likeLoading[comment.id]"
            @click="toggleLike(comment)"
          >
            {{ likeStates[comment.id]?.liked ? '❤️' : '🤍' }} {{ likeStates[comment.id]?.count || 0 }}
          </button>
        </div>

        <div v-if="comment.children.length" class="replies">
          <div
            v-for="reply in comment.children"
            :key="reply.id"
            class="reply-item"
          >
            <div class="comment-header">
              <span class="comment-author">{{ reply.author_name }}</span>
              <span class="comment-date">{{ formatDate(reply.created_at) }}</span>
              <button
                v-if="canDelete(reply)"
                class="delete-btn"
                @click="deleteComment(reply.id)"
              >
                删除
              </button>
            </div>
            <p class="comment-content">{{ reply.content }}</p>
            <div class="comment-actions">
              <button
                class="like-btn"
                :class="{ liked: likeStates[reply.id]?.liked }"
                :disabled="likeLoading[reply.id]"
                @click="toggleLike(reply)"
              >
                {{ likeStates[reply.id]?.liked ? '❤️' : '🤍' }} {{ likeStates[reply.id]?.count || 0 }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.comment-section {
  margin-top: 30px;
}

.section-title {
  font-size: 18px;
  margin-bottom: 16px;
  color: #333;
}

.comment-form {
  margin-bottom: 20px;
}

.login-tip {
  color: #666;
  text-align: center;
}

.login-tip a {
  color: #3498db;
}

.comment-item {
  margin-bottom: 12px;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.comment-author {
  font-weight: 600;
  color: #3498db;
}

.comment-date {
  font-size: 12px;
  color: #999;
}

.comment-content {
  color: #444;
  font-size: 14px;
  line-height: 1.6;
}

.comment-actions {
  margin-top: 8px;
}

.like-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid #ddd;
  background-color: #fff;
  border-radius: 12px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.like-btn:hover {
  border-color: #e74c3c;
  color: #e74c3c;
}

.like-btn.liked {
  background-color: #fff0f0;
  border-color: #e74c3c;
  color: #e74c3c;
}

.like-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.replies {
  margin-top: 12px;
  padding-left: 16px;
  border-left: 3px solid #eee;
}

.reply-item {
  padding: 10px 0;
  border-bottom: 1px dashed #eee;
}

.reply-item:last-child {
  border-bottom: none;
}

.delete-btn {
  margin-left: auto;
  background: none;
  border: none;
  color: #e74c3c;
  font-size: 12px;
  cursor: pointer;
}

.delete-btn:hover {
  text-decoration: underline;
}
</style>
