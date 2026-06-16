<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { postApi, categoryApi, tagApi } from '@/api'
import PostCard from '@/components/PostCard.vue'

const route = useRoute()
const router = useRouter()

const posts = ref([])
const total = ref(0)
const loading = ref(false)
const categories = ref([])
const tags = ref([])
const hotPosts = ref([])
const searchKeyword = ref(route.query.keyword || '')
let keywordDebounceTimer = null

const query = ref({
  page: Number(route.query.page) || 1,
  page_size: 10,
  keyword: route.query.keyword || '',
  category_id: route.query.category_id || '',
  tag_id: route.query.tag_id || '',
  order_by: route.query.order_by || 'created_at'
})

// 防抖：用户输入关键词后 300ms 再触发搜索，减少无效请求
const debounceKeywordSearch = (value) => {
  clearTimeout(keywordDebounceTimer)
  keywordDebounceTimer = setTimeout(() => {
    query.value.keyword = value
    applyFilter()
  }, 300)
}

const fetchPosts = async () => {
  loading.value = true
  try {
    const res = await postApi.list(query.value)
    posts.value = res.data
    total.value = res.total
  } catch (err) {
    alert(err.message || '加载文章失败')
  } finally {
    loading.value = false
  }
}

const fetchFilters = async () => {
  try {
    const [catRes, tagRes, hotRes] = await Promise.all([
      categoryApi.list(),
      tagApi.list(),
      postApi.hot({ limit: 5 })
    ])
    categories.value = catRes
    tags.value = tagRes
    hotPosts.value = hotRes
  } catch (err) {
    console.error('加载筛选条件失败', err)
  }
}

const applyFilter = () => {
  query.value.page = 1
  updateRoute()
}

const updateRoute = () => {
  const q = {}
  if (query.value.page > 1) q.page = query.value.page
  if (query.value.keyword) q.keyword = query.value.keyword
  if (query.value.category_id) q.category_id = query.value.category_id
  if (query.value.tag_id) q.tag_id = query.value.tag_id
  if (query.value.order_by !== 'created_at') q.order_by = query.value.order_by
  router.push({ path: '/', query: q })
}

const changePage = (page) => {
  query.value.page = page
  updateRoute()
}

const totalPages = () => Math.ceil(total.value / query.value.page_size)

watch(() => route.query, () => {
  query.value.page = Number(route.query.page) || 1
  query.value.keyword = route.query.keyword || ''
  query.value.category_id = route.query.category_id || ''
  query.value.tag_id = route.query.tag_id || ''
  query.value.order_by = route.query.order_by || 'created_at'
  searchKeyword.value = query.value.keyword
  fetchPosts()
}, { immediate: true })

watch(searchKeyword, (value) => {
  debounceKeywordSearch(value)
})

onMounted(() => {
  fetchFilters()
})
</script>

<template>
  <div class="home">
    <div class="main">
      <!-- 搜索与筛选 -->
      <div class="filter-bar card">
        <input
          v-model="searchKeyword"
          type="text"
          placeholder="搜索文章标题或内容..."
          @keyup.enter="applyFilter"
        />
        <select v-model="query.category_id" @change="applyFilter">
          <option value="">全部分类</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select v-model="query.tag_id" @change="applyFilter">
          <option value="">全部标签</option>
          <option v-for="t in tags" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
        <select v-model="query.order_by" @change="applyFilter">
          <option value="created_at">最新发布</option>
          <option value="view_count">最多阅读</option>
        </select>
        <button class="btn btn-primary" @click="applyFilter">搜索</button>
      </div>

      <!-- 文章列表 -->
      <div v-if="loading" class="empty-tip card">加载中...</div>
      <template v-else>
        <PostCard v-for="post in posts" :key="post.id" :post="post" />
        <div v-if="posts.length === 0" class="empty-tip card">没有找到相关文章</div>
      </template>

      <!-- 分页 -->
      <div v-if="totalPages() > 1" class="pagination">
        <button
          v-for="page in totalPages()"
          :key="page"
          class="page-btn"
          :class="{ active: page === query.page }"
          @click="changePage(page)"
        >
          {{ page }}
        </button>
      </div>
    </div>

    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="card">
        <h4>🔥 热门文章</h4>
        <ul class="hot-list">
          <li v-for="post in hotPosts" :key="post.id">
            <router-link :to="`/posts/${post.id}`">{{ post.title }}</router-link>
            <span class="hot-views">{{ post.view_count }}</span>
          </li>
        </ul>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.home {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 20px;
}

.filter-bar {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr auto;
  gap: 12px;
  align-items: center;
}

.pagination {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 20px;
}

.page-btn {
  padding: 6px 12px;
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 4px;
  cursor: pointer;
}

.page-btn.active {
  background-color: #3498db;
  color: #fff;
  border-color: #3498db;
}

.sidebar h4 {
  margin-bottom: 12px;
  color: #333;
}

.hot-list {
  list-style: none;
}

.hot-list li {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.hot-list li:last-child {
  border-bottom: none;
}

.hot-list a {
  color: #555;
  text-decoration: none;
  font-size: 14px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hot-list a:hover {
  color: #3498db;
}

.hot-views {
  color: #999;
  font-size: 12px;
  margin-left: 8px;
}

@media (max-width: 768px) {
  .home {
    grid-template-columns: 1fr;
  }

  .filter-bar {
    grid-template-columns: 1fr;
  }
}
</style>
