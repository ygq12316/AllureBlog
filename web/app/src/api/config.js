// 博客配置 API：公开读，后台写
import client from './client'

export const getConfig = () => client.get('/api/config').then(r => r.data.config)
export const updateConfig = payload => client.put('/api/admin/config', payload).then(r => r.data.config)
