// 文章相关 API：列表 / 搜索 / 详情 / 增删改
import client from './client'

export const listArticles = params => client.get('/api/articles', { params }).then(r => r.data)
export const getArticleBySlug = slug => client.get('/api/articles', { params: { slug } }).then(r => r.data)
export const searchArticles = q => client.get('/api/articles/search', { params: { q } }).then(r => r.data)
export const getArticle = id => client.get(`/api/articles/${id}`).then(r => r.data)
export const createArticle = data => client.post('/api/articles', data).then(r => r.data)
export const updateArticle = (id, data) => client.put(`/api/articles/${id}`, data).then(r => r.data)
export const removeArticle = id => client.delete(`/api/articles/${id}`).then(r => r.data)
