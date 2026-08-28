// 评论 API：列表与创建是访客公开接口，删除走后台鉴权路由
import client from './client'

export const listComments = noteId => client.get(`/api/notes/${noteId}/comments`).then(r => r.data)
export const createComment = (noteId, payload) => client.post(`/api/notes/${noteId}/comments`, payload).then(r => r.data)
export const removeComment = id => client.delete(`/api/admin/comments/${id}`).then(r => r.data)
