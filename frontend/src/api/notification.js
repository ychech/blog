import request from './request'

export const getNotifications = (params) => {
  return request.get('/notifications', { params })
}

export const getUnreadCount = () => {
  return request.get('/notifications/unread-count')
}

export const markAsRead = (id) => {
  return request.put(`/notifications/${id}/read`)
}
