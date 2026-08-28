// 访客 API：身份/账号为公开接口，管理列表走后台鉴权路由
import client from './client'

export const saveVisitor = visitor => client.post('/api/visitor', visitor).then(r => r.data)
export const registerAccount = payload => client.post('/api/visitor/register', payload).then(r => r.data)
export const loginAccount = payload => client.post('/api/visitor/login', payload).then(r => r.data)
export const getVisitor = uuid => client.get(`/api/visitor/${uuid}`).then(r => r.data.visitor)

export const listVisitors = () => client.get('/api/admin/visitors').then(r => r.data.visitors || [])
export const updateVisitor = (uuid, payload) => client.put(`/api/admin/visitors/${uuid}`, payload).then(r => r.data)
export const removeVisitor = uuid => client.delete(`/api/admin/visitors/${uuid}`).then(r => r.data)
