// 分类 API
import client from './client'

export const listCategories = () => client.get('/api/categories').then(r => r.data.categories || [])
export const createCategory = name => client.post('/api/categories', { name }).then(r => r.data)
export const removeCategory = id => client.delete(`/api/categories/${id}`).then(r => r.data)
