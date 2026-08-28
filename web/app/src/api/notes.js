// 随笔相关 API：列表 / 详情 / 增删改
import client from './client'

export const listNotes = params => client.get('/api/notes', { params }).then(r => r.data)
export const getNote = id => client.get(`/api/notes/${id}`).then(r => r.data)
export const createNote = data => client.post('/api/notes', data).then(r => r.data)
export const updateNote = (id, data) => client.put(`/api/notes/${id}`, data).then(r => r.data)
export const removeNote = id => client.delete(`/api/notes/${id}`).then(r => r.data)
