<script setup>
import { stripMarkdown } from '@/utils/markdown'

defineProps({
  post: {
    type: Object,
    required: true
  }
})

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}
</script>

<template>
  <article class="post-card card">
    <router-link :to="`/posts/${post.id}`" class="post-title">
      {{ post.title }}
      <span v-if="post.status === 'draft'" class="draft-badge">草稿</span>
    </router-link>
    <p class="post-summary">{{ post.summary || stripMarkdown(post.content, 120) }}</p>
    <div class="post-meta">
      <span class="author">👤 {{ post.author?.nickname || post.author?.username }}</span>
      <span v-if="post.category?.name" class="category">📁 {{ post.category.name }}</span>
      <span class="date">📅 {{ formatDate(post.created_at) }}</span>
      <span class="views">👁 {{ post.view_count || 0 }}</span>
      <span class="likes">❤️ {{ post.like_count || 0 }}</span>
    </div>
    <div v-if="post.tags?.length" class="post-tags">
      <span v-for="tag in post.tags" :key="tag.id" class="tag"># {{ tag.name }}</span>
    </div>
  </article>
</template>

<style scoped>
.post-card {
  transition: transform 0.2s, box-shadow 0.2s;
}

.post-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.post-title {
  font-size: 20px;
  font-weight: 600;
  color: #2c3e50;
  text-decoration: none;
  display: block;
  margin-bottom: 10px;
}

.post-title:hover {
  color: #3498db;
}

.draft-badge {
  display: inline-block;
  background-color: #f39c12;
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  margin-left: 8px;
  font-weight: normal;
}

.post-summary {
  color: #666;
  font-size: 14px;
  line-height: 1.6;
  margin-bottom: 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.post-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 13px;
  color: #888;
  margin-bottom: 10px;
}

.post-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  background-color: #f0f7ff;
  color: #3498db;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
</style>
