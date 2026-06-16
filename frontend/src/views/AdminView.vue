<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { categoryApi, tagApi, badgeApi, authApi } from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const { user } = useUserStore()

const activeTab = ref('dashboard')

const categories = ref([])
const tags = ref([])
const badges = ref([])
const users = ref([])
const userBadges = ref([])
const stats = ref(null)

const newCategory = ref('')
const editingCategory = ref({ id: null, name: '' })

const badgeForm = ref({
  name: '',
  description: '',
  icon_url: '',
  contract_address: '',
  token_id: '',
  metadata_url: ''
})
const editingBadge = ref(null)

const awardForm = ref({
  user_id: '',
  badge_id: '',
  reason: ''
})

const loading = ref(false)
const error = ref('')
const message = ref('')

// 简单权限拦截：非管理员访问此页面跳回首页
if (user.value?.role !== 'admin') {
  router.push('/')
}

const fetchData = async () => {
  try {
    const [catRes, tagRes, badgeRes] = await Promise.all([
      categoryApi.list(),
      tagApi.list(),
      badgeApi.list()
    ])
    categories.value = catRes
    tags.value = tagRes
    badges.value = badgeRes
  } catch (err) {
    error.value = err.message || '加载数据失败'
  }
}

const fetchUsers = async () => {
  try {
    const res = await authApi.listUsers({ page_size: 100 })
    users.value = res.data || []
  } catch (err) {
    console.error('加载用户失败', err)
  }
}

const fetchStats = async () => {
  try {
    stats.value = await authApi.getStats()
  } catch (err) {
    console.error('加载统计失败', err)
  }
}

const showMessage = (msg) => {
  message.value = msg
  setTimeout(() => { message.value = '' }, 3000)
}

// 分类管理
const createCategory = async () => {
  const name = newCategory.value.trim()
  if (!name) return
  try {
    await categoryApi.create(name)
    newCategory.value = ''
    showMessage('分类创建成功')
    await fetchData()
  } catch (err) {
    error.value = err.message || '创建分类失败'
  }
}

const startEditCategory = (category) => {
  editingCategory.value = { id: category.id, name: category.name }
}

const cancelEditCategory = () => {
  editingCategory.value = { id: null, name: '' }
}

const updateCategory = async () => {
  if (!editingCategory.value.id) return
  try {
    await categoryApi.update(editingCategory.value.id, editingCategory.value.name.trim())
    editingCategory.value = { id: null, name: '' }
    showMessage('分类更新成功')
    await fetchData()
  } catch (err) {
    error.value = err.message || '更新分类失败'
  }
}

const deleteCategory = async (id) => {
  if (!confirm('确定删除该分类吗？关联文章分类会被置空。')) return
  try {
    await categoryApi.delete(id)
    showMessage('分类删除成功')
    await fetchData()
  } catch (err) {
    error.value = err.message || '删除分类失败'
  }
}

const deleteTag = async (id) => {
  if (!confirm('确定删除该标签吗？')) return
  try {
    await tagApi.delete(id)
    showMessage('标签删除成功')
    await fetchData()
  } catch (err) {
    error.value = err.message || '删除标签失败'
  }
}

// 勋章管理
const resetBadgeForm = () => {
  badgeForm.value = {
    name: '',
    description: '',
    icon_url: '',
    contract_address: '',
    token_id: '',
    metadata_url: ''
  }
  editingBadge.value = null
}

const createBadge = async () => {
  if (!badgeForm.value.name.trim()) {
    error.value = '请输入勋章名称'
    return
  }
  try {
    if (editingBadge.value) {
      await badgeApi.update(editingBadge.value.id, badgeForm.value)
      showMessage('勋章更新成功')
    } else {
      await badgeApi.create(badgeForm.value)
      showMessage('勋章创建成功')
    }
    resetBadgeForm()
    await fetchData()
  } catch (err) {
    error.value = err.message || '保存勋章失败'
  }
}

const startEditBadge = (badge) => {
  editingBadge.value = badge
  badgeForm.value = {
    name: badge.name,
    description: badge.description || '',
    icon_url: badge.icon_url || '',
    contract_address: badge.contract_address || '',
    token_id: badge.token_id || '',
    metadata_url: badge.metadata_url || ''
  }
  activeTab.value = 'badges'
}

const deleteBadge = async (id) => {
  if (!confirm('确定删除该勋章吗？已颁发的勋章也会被删除。')) return
  try {
    await badgeApi.delete(id)
    showMessage('勋章删除成功')
    await fetchData()
  } catch (err) {
    error.value = err.message || '删除勋章失败'
  }
}

// 颁发勋章
const awardBadge = async () => {
  if (!awardForm.value.user_id || !awardForm.value.badge_id) {
    error.value = '请选择用户和勋章'
    return
  }
  try {
    await badgeApi.award({
      user_id: Number(awardForm.value.user_id),
      badge_id: Number(awardForm.value.badge_id),
      reason: awardForm.value.reason
    })
    awardForm.value = { user_id: '', badge_id: '', reason: '' }
    showMessage('勋章颁发成功')
  } catch (err) {
    error.value = err.message || '颁发勋章失败'
  }
}

const viewUserBadges = async (userId) => {
  try {
    const res = await badgeApi.getUserBadges(userId)
    userBadges.value = res
    activeTab.value = 'userBadges'
  } catch (err) {
    error.value = err.message || '加载用户勋章失败'
  }
}

const switchTab = (tab) => {
  activeTab.value = tab
  if (tab === 'users') fetchUsers()
  if (tab === 'dashboard') fetchStats()
}

onMounted(() => {
  fetchData()
  fetchUsers()
  fetchStats()
})
</script>

<template>
  <div class="admin">
    <div class="card">
      <h2>⚙️ 管理后台</h2>
      <p class="admin-tip">只有管理员可以访问后台管理功能。</p>

      <p v-if="message" class="success-msg">{{ message }}</p>
      <p v-if="error" class="error-msg">{{ error }}</p>

      <div class="tabs">
        <button class="tab-btn" :class="{ active: activeTab === 'dashboard' }" @click="switchTab('dashboard')">
          仪表盘
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'categories' }" @click="switchTab('categories')">
          分类管理
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'tags' }" @click="switchTab('tags')">
          标签管理
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'badges' }" @click="switchTab('badges')">
          勋章管理
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'users' }" @click="switchTab('users')">
          用户列表
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'award' }" @click="switchTab('award')">
          颁发勋章
        </button>
      </div>

      <!-- 仪表盘 -->
      <div v-if="activeTab === 'dashboard'" class="admin-section">
        <h3>📊 站点概览</h3>
        <div v-if="stats" class="stats-grid">
          <div class="stat-card">
            <span class="stat-number">{{ stats.user_count }}</span>
            <span class="stat-label">注册用户</span>
          </div>
          <div class="stat-card">
            <span class="stat-number">{{ stats.post_count }}</span>
            <span class="stat-label">文章总数</span>
          </div>
          <div class="stat-card">
            <span class="stat-number">{{ stats.comment_count }}</span>
            <span class="stat-label">评论总数</span>
          </div>
          <div class="stat-card">
            <span class="stat-number">{{ stats.category_count }}</span>
            <span class="stat-label">分类数量</span>
          </div>
          <div class="stat-card">
            <span class="stat-number">{{ stats.tag_count }}</span>
            <span class="stat-label">标签数量</span>
          </div>
          <div class="stat-card">
            <span class="stat-number">{{ stats.badge_count }}</span>
            <span class="stat-label">勋章数量</span>
          </div>
        </div>
      </div>

      <!-- 分类管理 -->
      <div v-if="activeTab === 'categories'" class="admin-section">
        <h3>📁 分类管理</h3>
        <div class="create-row">
          <input v-model="newCategory" type="text" placeholder="输入新分类名称" @keyup.enter="createCategory" />
          <button class="btn btn-primary" @click="createCategory">创建分类</button>
        </div>

        <ul class="item-list">
          <li v-for="category in categories" :key="category.id" class="item">
            <template v-if="editingCategory.id === category.id">
              <input v-model="editingCategory.name" type="text" />
              <div class="item-actions">
                <button class="btn btn-success" @click="updateCategory">保存</button>
                <button class="btn" @click="cancelEditCategory">取消</button>
              </div>
            </template>
            <template v-else>
              <span class="item-name">{{ category.name }}</span>
              <div class="item-actions">
                <button class="btn btn-primary" @click="startEditCategory(category)">编辑</button>
                <button class="btn btn-danger" @click="deleteCategory(category.id)">删除</button>
              </div>
            </template>
          </li>
        </ul>
      </div>

      <!-- 标签管理 -->
      <div v-if="activeTab === 'tags'" class="admin-section">
        <h3>🏷️ 标签管理</h3>
        <p class="section-tip">标签可在写文章时由登录用户创建，此处仅支持删除。</p>
        <ul class="item-list">
          <li v-for="tag in tags" :key="tag.id" class="item">
            <span class="item-name">{{ tag.name }}</span>
            <div class="item-actions">
              <button class="btn btn-danger" @click="deleteTag(tag.id)">删除</button>
            </div>
          </li>
        </ul>
      </div>

      <!-- 勋章管理 -->
      <div v-if="activeTab === 'badges'" class="admin-section">
        <h3>🏅 勋章管理</h3>

        <div class="form-card">
          <h4>{{ editingBadge ? '编辑勋章' : '创建勋章' }}</h4>
          <div class="form-group">
            <label>勋章名称 *</label>
            <input v-model="badgeForm.name" type="text" placeholder="例如：年度作者" />
          </div>
          <div class="form-group">
            <label>描述</label>
            <input v-model="badgeForm.description" type="text" placeholder="勋章描述" />
          </div>
          <div class="form-group">
            <label>图标 URL</label>
            <input v-model="badgeForm.icon_url" type="text" placeholder="https://example.com/badge.png" />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>合约地址（NFT 可选）</label>
              <input v-model="badgeForm.contract_address" type="text" placeholder="0x..." />
            </div>
            <div class="form-group">
              <label>Token ID（NFT 可选）</label>
              <input v-model="badgeForm.token_id" type="text" placeholder="123" />
            </div>
          </div>
          <div class="form-group">
            <label>Metadata URL（NFT 可选）</label>
            <input v-model="badgeForm.metadata_url" type="text" placeholder="https://example.com/metadata.json" />
          </div>
          <div class="form-actions">
            <button v-if="editingBadge" class="btn" @click="resetBadgeForm">取消</button>
            <button class="btn btn-success" @click="createBadge">{{ editingBadge ? '保存修改' : '创建勋章' }}</button>
          </div>
        </div>

        <h4>勋章列表</h4>
        <ul class="badge-list">
          <li v-for="badge in badges" :key="badge.id" class="badge-item">
            <img v-if="badge.icon_url" :src="badge.icon_url" alt="icon" class="badge-icon" />
            <div v-else class="badge-icon placeholder">🏅</div>
            <div class="badge-info">
              <strong>{{ badge.name }}</strong>
              <p>{{ badge.description || '暂无描述' }}</p>
              <span v-if="badge.contract_address" class="nft-tag">NFT</span>
            </div>
            <div class="item-actions">
              <button class="btn btn-primary" @click="startEditBadge(badge)">编辑</button>
              <button class="btn btn-danger" @click="deleteBadge(badge.id)">删除</button>
            </div>
          </li>
        </ul>
      </div>

      <!-- 用户列表 -->
      <div v-if="activeTab === 'users'" class="admin-section">
        <h3>👥 用户列表</h3>
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>用户名</th>
              <th>昵称</th>
              <th>角色</th>
              <th>注册时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td>{{ u.id }}</td>
              <td>{{ u.username }}</td>
              <td>{{ u.nickname || '-' }}</td>
              <td>
                <span class="role-badge" :class="u.role">{{ u.role === 'admin' ? '管理员' : '用户' }}</span>
              </td>
              <td>{{ new Date(u.created_at).toLocaleDateString('zh-CN') }}</td>
              <td>
                <button class="btn btn-primary btn-sm" @click="awardForm.user_id = u.id; activeTab = 'award'">颁奖</button>
                <button class="btn btn-sm" @click="viewUserBadges(u.id)">勋章</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 颁发勋章 -->
      <div v-if="activeTab === 'award'" class="admin-section">
        <h3>🎖️ 颁发勋章</h3>
        <div class="form-card">
          <div class="form-group">
            <label>选择用户 *</label>
            <select v-model="awardForm.user_id">
              <option value="">请选择用户</option>
              <option v-for="u in users" :key="u.id" :value="u.id">{{ u.username }}（ID: {{ u.id }}）</option>
            </select>
          </div>
          <div class="form-group">
            <label>选择勋章 *</label>
            <select v-model="awardForm.badge_id">
              <option value="">请选择勋章</option>
              <option v-for="badge in badges" :key="badge.id" :value="badge.id">{{ badge.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>颁发原因</label>
            <input v-model="awardForm.reason" type="text" placeholder="为什么颁发这个勋章？" />
          </div>
          <button class="btn btn-success" @click="awardBadge">颁发勋章</button>
        </div>
      </div>

      <!-- 用户勋章详情 -->
      <div v-if="activeTab === 'userBadges'" class="admin-section">
        <h3>🏅 用户勋章</h3>
        <button class="btn btn-sm" @click="activeTab = 'users'">返回用户列表</button>
        <div v-if="userBadges.length === 0" class="empty-tip">该用户暂无勋章</div>
        <div v-else class="badges-grid">
          <div v-for="item in userBadges" :key="item.id" class="badge-card">
            <img v-if="item.badge.icon_url" :src="item.badge.icon_url" alt="icon" />
            <span v-else class="badge-emoji">🏅</span>
            <span class="badge-name">{{ item.badge.name }}</span>
            <span v-if="item.reason" class="badge-reason">{{ item.reason }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.admin {
  max-width: 900px;
  margin: 0 auto;
}

.admin h2 {
  margin-bottom: 8px;
  color: #333;
}

.admin-tip {
  color: #888;
  font-size: 14px;
  margin-bottom: 20px;
}

.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  border-bottom: 1px solid #eee;
  padding-bottom: 12px;
  flex-wrap: wrap;
}

.tab-btn {
  padding: 8px 16px;
  border: none;
  background: #f0f0f0;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #555;
}

.tab-btn.active {
  background: #3498db;
  color: #fff;
}

.admin-section {
  margin-top: 16px;
}

.admin-section h3 {
  margin-bottom: 16px;
  color: #444;
}

.section-tip {
  color: #999;
  font-size: 13px;
  margin-bottom: 12px;
}

/* 仪表盘 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 16px;
}

.stat-card {
  background-color: #f9f9f9;
  border-radius: 10px;
  padding: 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-number {
  font-size: 28px;
  font-weight: bold;
  color: #3498db;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

/* 通用 */
.create-row {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.create-row input {
  flex: 1;
}

.create-row .btn {
  width: auto;
}

.item-list {
  list-style: none;
}

.item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: #f9f9f9;
  border-radius: 6px;
  margin-bottom: 8px;
}

.item-name {
  font-size: 15px;
  color: #333;
}

.item-actions {
  display: flex;
  gap: 8px;
}

.item-actions .btn {
  padding: 6px 12px;
  font-size: 13px;
}

.form-card {
  background-color: #f9f9f9;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.form-card h4 {
  margin-bottom: 12px;
  color: #444;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

/* 勋章列表 */
.badge-list {
  list-style: none;
}

.badge-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 10px;
}

.badge-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
}

.badge-icon.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fff;
  border: 1px solid #eee;
  font-size: 24px;
}

.badge-info {
  flex: 1;
}

.badge-info strong {
  color: #333;
}

.badge-info p {
  color: #888;
  font-size: 13px;
  margin-top: 2px;
}

.nft-tag {
  display: inline-block;
  background-color: #8e44ad;
  color: #fff;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  margin-top: 4px;
}

/* 用户表格 */
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.data-table th {
  color: #666;
  font-weight: 600;
  background-color: #f9f9f9;
}

.role-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.role-badge.admin {
  background-color: #e74c3c;
  color: #fff;
}

.role-badge.user {
  background-color: #3498db;
  color: #fff;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
}

/* 用户勋章 */
.badges-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 16px;
  margin-top: 16px;
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
}

.success-msg {
  color: #2ecc71;
  margin-bottom: 12px;
}

.error-msg {
  color: #e74c3c;
  margin-bottom: 12px;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .data-table {
    font-size: 12px;
  }

  .data-table th,
  .data-table td {
    padding: 8px;
  }
}
</style>
