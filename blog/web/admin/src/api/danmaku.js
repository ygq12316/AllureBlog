// 弹幕 API：列表与发布是访客公开接口，删除走后台鉴权路由
import client from './client'

export const listDanmaku = () => client.get('/api/danmaku').then(r => r.data.danmaku || [])
export const createDanmaku = payload => client.post('/api/danmaku', payload).then(r => r.data)
export const removeDanmaku = id => client.delete(`/api/admin/danmaku/${id}`).then(r => r.data)
