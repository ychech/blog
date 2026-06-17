<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { postApi, commentApi, likeApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { renderMarkdown, generateTOC } from '@/utils/markdown'
import CommentList from '@/components/CommentList.vue'

const props = defineProps({
  id: {
    type: String,
    required: true
  }
})

const route = useRoute()
const router = useRouter()
const { isLoggedIn, user } = useUserStore()

const post = ref(null)
const comments = ref([])
const loading = ref(false)
const like = ref({ count: 0, liked: false })
const likeLoading = ref(false)

const fetchPost = async () => {
  loading.value = true
  try {
    const res = await postApi.get(props.id)
    post.value = res
  } catch (err) {
    alert(err.message || '文章加载失败')
  } finally {
    loading.value = false
  }
}

const fetchComments = async () => {
  try {
    const res = await commentApi.listByPost(props.id)
    comments.value = res
  } catch (err) {
    console.error('加载评论失败', err)
  }
}

const fetchLikeStatus = async () => {
  try {
    const res = await likeApi.status(props.id)
    like.value = res
  } catch (err) {
    console.error('加载点赞状态失败', err)
  }
}

const toggleLike = async () => {
  if (!isLoggedIn.value) {
    router.push('/login')
    return
  }
  if (likeLoading.value) return
  likeLoading.value = true
  try {
    const res = await likeApi.toggle(props.id)
    like.value = res
  } catch (err) {
    alert(err.message || '操作失败')
  } finally {
    likeLoading.value = false
  }
}

const deletePost = async () => {
  if (!confirm('确定删除这篇文章吗？')) return
  try {
    await postApi.delete(props.id)
    router.push('/')
  } catch (err) {
    alert(err.message || '删除失败')
  }
}

const renderedContent = computed(() => {
  return renderMarkdown(post.value?.content || '')
})

const toc = computed(() => {
  return generateTOC(post.value?.content || '')
})

const scrollToHeading = (id) => {
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchPost()
  fetchComments()
  fetchLikeStatus()
})
</script>

<template>
  <div class="post-detail">
    <div v-if="loading" class="empty-tip card">加载中...</div>
    <template v-else-if="post">
      <aside v-if="toc.length > 0" class="toc-sidebar card">
        <h3 class="toc-title">目录</h3>
        <ul class="toc-list">
          <li
            v-for="item in toc"
            :key="item.id"
            :class="['toc-item', `toc-level-${item.level}`]"
            @click="scrollToHeading(item.id)"
          >
            {{ item.title }}
          </li>
        </ul>
      </aside>

      <article class="card">
        <h1 class="post-title">{{ post.title }}</h1>
        <div class="post-meta">
          <span>👤 {{ post.author?.nickname || post.author?.username }}</span>
          <span v-if="post.category?.name">📁 {{ post.category.name }}</span>
          <span>📅 {{ formatDate(post.created_at) }}</span>
          <span>👁 {{ post.view_count || 0 }}</span>
        </div>
        <div v-if="post.tags?.length" class="post-tags">
          <span v-for="tag in post.tags" :key="tag.id" class="tag"># {{ tag.name }}</span>
        </div>
        <div class="post-content markdown-body" v-html="renderedContent"></div>

        <div class="like-section">
          <button
            class="like-btn"
            :class="{ liked: like.liked }"
            :disabled="likeLoading"
            @click="toggleLike"
          >
            {{ like.liked ? '❤️' : '🤍' }} 点赞 {{ like.count }}
          </button>
        </div>

        <div v-if="isLoggedIn && user?.id === post.author_id" class="post-actions">
          <router-link :to="`/posts/${post.id}/edit`" class="btn btn-primary">编辑文章</router-link>
          <button class="btn btn-danger" @click="deletePost">删除文章</button>
        </div>
      </article>

      <CommentList
        :post-id="post.id"
        :comments="comments"
        @refresh="fetchComments"
      />
    </template>
    <div v-else class="empty-tip card">文章不存在或已被删除</div>
  </div>
</template>

<style scoped>
.post-detail {
  position: relative;
}

.toc-sidebar {
  position: fixed;
  right: 20px;
  top: 100px;
  width: 220px;
  max-height: calc(100vh - 140px);
  overflow-y: auto;
  padding: 16px;
  z-index: 10;
}

.toc-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #2c3e50;
}

.toc-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.toc-item {
  padding: 6px 0;
  color: #555;
  font-size: 14px;
  cursor: pointer;
  line-height: 1.4;
  border-bottom: none;
}

.toc-item:hover {
  color: #3498db;
}

.toc-level-1 {
  font-weight: 600;
}

.toc-level-2 {
  padding-left: 12px;
}

.toc-level-3,
.toc-level-4,
.toc-level-5,
.toc-level-6 {
  padding-left: 24px;
  font-size: 13px;
}

@media (max-width: 1200px) {
  .toc-sidebar {
    display: none;
  }
}

.post-title {
  font-size: 28px;
  color: #2c3e50;
  margin-bottom: 16px;
  line-height: 1.3;
}

.post-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  color: #888;
  font-size: 14px;
  margin-bottom: 12px;
}

.post-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 20px;
}

.tag {
  background-color: #f0f7ff;
  color: #3498db;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 13px;
}

.post-content {
  color: #444;
  font-size: 16px;
  line-height: 1.8;
  margin-bottom: 24px;
}

/* Markdown 渲染的基础样式 */
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4) {
  color: #2c3e50;
  margin: 20px 0 12px;
}

.markdown-body :deep(p) {
  margin-bottom: 12px;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 12px 0;
  padding-left: 24px;
}

.markdown-body :deep(code) {
  background-color: #f4f4f4;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

.markdown-body :deep(pre) {
  background-color: #f8f8f8;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 12px 0;
}

.markdown-body :deep(pre code) {
  background: none;
  padding: 0;
}

.markdown-body :deep(a) {
  color: #3498db;
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid #3498db;
  padding-left: 16px;
  margin: 12px 0;
  color: #666;
}

.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid #eee;
  margin: 20px 0;
}

.post-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #eee;
  gap: 12px;
}

.like-section {
  display: flex;
  justify-content: center;
  margin: 24px 0;
}

.like-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 24px;
  border: 1px solid #ddd;
  background-color: #fff;
  border-radius: 24px;
  font-size: 16px;
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
</style>
