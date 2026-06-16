import request from './request'

// 认证相关
export const authApi = {
  register: (data) => request.post('/auth/register', data),
  login: (data) => request.post('/auth/login', data),
  me: () => request.get('/auth/me'),
  updateProfile: (data) => request.put('/auth/me', data),
  listUsers: (params) => request.get('/auth/users', { params }),
  getStats: () => request.get('/auth/stats')
}

// 文章相关
export const postApi = {
  list: (params) => request.get('/posts', { params }),
  hot: (params) => request.get('/posts/hot', { params }),
  get: (id) => request.get(`/posts/${id}`),
  create: (data) => request.post('/posts', data),
  update: (id, data) => request.put(`/posts/${id}`, data),
  delete: (id) => request.delete(`/posts/${id}`)
}

// 点赞相关
export const likeApi = {
  status: (postId) => request.get(`/posts/${postId}/like`),
  toggle: (postId) => request.post(`/posts/${postId}/like`)
}

// 分类相关
export const categoryApi = {
  list: () => request.get('/categories'),
  create: (name) => request.post('/categories', { name }),
  update: (id, name) => request.put(`/categories/${id}`, { name }),
  delete: (id) => request.delete(`/categories/${id}`)
}

// 标签相关
export const tagApi = {
  list: () => request.get('/tags'),
  create: (name) => request.post('/tags', { name }),
  delete: (id) => request.delete(`/tags/${id}`)
}

// 评论相关
export const commentApi = {
  listByPost: (postId) => request.get(`/posts/${postId}/comments`),
  create: (data) => request.post('/comments', data),
  delete: (id) => request.delete(`/comments/${id}`)
}

// 评论点赞相关
export const commentLikeApi = {
  status: (commentId) => request.get(`/comments/${commentId}/like`),
  batchStatus: (commentIds) => request.get('/comments/likes', { params: { ids: commentIds.join(',') } }),
  toggle: (commentId) => request.post(`/comments/${commentId}/like`)
}

// 勋章相关
export const badgeApi = {
  list: () => request.get('/badges'),
  create: (data) => request.post('/badges', data),
  update: (id, data) => request.put(`/badges/${id}`, data),
  delete: (id) => request.delete(`/badges/${id}`),
  award: (data) => request.post('/badges/award', data),
  revoke: (id) => request.delete(`/user-badges/${id}`),
  getUserBadges: (userId) => request.get(`/users/${userId}/badges`),
  getMyBadges: () => request.get('/auth/badges')
}
