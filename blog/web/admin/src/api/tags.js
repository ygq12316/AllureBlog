// 标签 API（列表为已发布文章的聚合统计）
import client from './client'

export const listTags = () => client.get('/api/tags').then(r => r.data.tags || [])
export const createTag = name => client.post('/api/tags', { name }).then(r => r.data)
export const removeTag = id => client.delete(`/api/tags/${id}`).then(r => r.data)
