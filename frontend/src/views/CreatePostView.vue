<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { postApi, categoryApi, tagApi } from '@/api'
import { renderMarkdown } from '@/utils/markdown'

const router = useRouter()

const form = ref({
  title: '',
  summary: '',
  content: '',
  cover_url: '',
  status: 'published',
  category_id: '',
  tag_ids: []
})

const categories = ref([])
const tags = ref([])
const loading = ref(false)
const error = ref('')
const newTagName = ref('')
const addingTag = ref(false)

const showPreview = ref(false)

const renderedContent = computed(() => renderMarkdown(form.value.content))

const addTag = async () => {
  const name = newTagName.value.trim()
  if (!name) return

  // 如果标签已存在，直接选中
  const existing = tags.value.find((t) => t.name === name)
  if (existing) {
    if (!form.value.tag_ids.includes(existing.id)) {
      form.value.tag_ids.push(existing.id)
    }
    newTagName.value = ''
    return
  }

  addingTag.value = true
  try {
    const res = await tagApi.create(name)
    tags.value.push(res)
    form.value.tag_ids.push(res.id)
    newTagName.value = ''
  } catch (err) {
    error.value = err.message || '创建标签失败'
  } finally {
    addingTag.value = false
  }
}

const handleTagKeydown = (e) => {
  if (e.key === 'Enter') {
    e.preventDefault()
    addTag()
  }
}

const fetchFilters = async () => {
  try {
    const [catRes, tagRes] = await Promise.all([
      categoryApi.list(),
      tagApi.list()
    ])
    categories.value = catRes
    tags.value = tagRes
  } catch (err) {
    console.error('加载分类标签失败', err)
  }
}

const submit = async () => {
  if (!form.value.title || !form.value.content || !form.value.category_id) {
    error.value = '请填写标题、正文并选择分类'
    return
  }

  loading.value = true
  error.value = ''
  try {
    const payload = {
      ...form.value,
      category_id: Number(form.value.category_id),
      tag_ids: form.value.tag_ids.map(Number)
    }
    const res = await postApi.create(payload)
    router.push(`/posts/${res.id}`)
  } catch (err) {
    error.value = err.message || '发布失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFilters()
})
</script>

<template>
  <div class="create-post">
    <div class="card">
      <h2>✍️ 写文章</h2>
      <form @submit.prevent="submit">
        <div class="form-group">
          <label>标题 *</label>
          <input v-model="form.title" type="text" placeholder="请输入文章标题" />
        </div>

        <div class="form-group">
          <label>摘要</label>
          <input v-model="form.summary" type="text" placeholder="一句话概括文章内容（可选）" />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>分类 *</label>
            <select v-model="form.category_id">
              <option value="">请选择分类</option>
              <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>状态</label>
            <select v-model="form.status">
              <option value="published">立即发布</option>
              <option value="draft">保存为草稿</option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>封面 URL</label>
          <input v-model="form.cover_url" type="text" placeholder="https://example.com/cover.png" />
        </div>

        <div class="form-group">
          <label>标签</label>
          <div class="tag-creator">
            <input
              v-model="newTagName"
              type="text"
              placeholder="输入新标签，按回车或点击添加"
              @keydown="handleTagKeydown"
            />
            <button type="button" class="btn btn-primary" :disabled="addingTag" @click="addTag">
              {{ addingTag ? '添加中...' : '添加' }}
            </button>
          </div>
          <div class="tag-select">
            <label v-for="tag in tags" :key="tag.id" class="tag-checkbox">
              <input
                v-model="form.tag_ids"
                type="checkbox"
                :value="tag.id"
              />
              {{ tag.name }}
            </label>
          </div>
        </div>

        <div class="form-group">
          <div class="content-label">
            <label>正文 *</label>
            <button type="button" class="preview-toggle" @click="showPreview = !showPreview">
              {{ showPreview ? '隐藏预览' : '预览 Markdown' }}
            </button>
          </div>
          <div class="editor-wrapper">
            <textarea
              v-model="form.content"
              class="markdown-editor"
              rows="16"
              placeholder="支持 Markdown 语法，例如 # 标题、**粗体**、`代码` 等"
            ></textarea>
            <div
              v-if="showPreview"
              class="markdown-preview markdown-body"
              v-html="renderedContent"
            ></div>
          </div>
        </div>

        <p v-if="error" class="error-msg">{{ error }}</p>
        <button type="submit" class="btn btn-success" :disabled="loading">
          {{ loading ? '发布中...' : '立即发布' }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.create-post h2 {
  margin-bottom: 20px;
  color: #333;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.tag-creator {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.tag-creator input {
  flex: 1;
}

.tag-creator .btn {
  width: auto;
  padding: 8px 16px;
}

.tag-select {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.tag-checkbox {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #555;
  cursor: pointer;
}

.tag-checkbox input {
  width: auto;
}

.content-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.preview-toggle {
  background: none;
  border: none;
  color: #3498db;
  cursor: pointer;
  font-size: 14px;
}

.editor-wrapper {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.markdown-editor,
.markdown-preview {
  min-height: 360px;
}

.markdown-preview {
  border: 1px solid #eee;
  border-radius: 4px;
  padding: 12px;
  background-color: #fafafa;
  overflow-y: auto;
}

.markdown-preview :deep(h1),
.markdown-preview :deep(h2),
.markdown-preview :deep(h3) {
  color: #2c3e50;
  margin: 16px 0 10px;
}

.markdown-preview :deep(p) {
  margin-bottom: 10px;
}

.markdown-preview :deep(code) {
  background-color: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.markdown-preview :deep(pre) {
  background-color: #f4f4f4;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
}

@media (max-width: 768px) {
  .form-row,
  .editor-wrapper {
    grid-template-columns: 1fr;
  }

  .markdown-preview {
    min-height: 240px;
  }
}
</style>
