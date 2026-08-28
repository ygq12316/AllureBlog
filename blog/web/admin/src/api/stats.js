// 统计 API
import client from './client'

export const getStats = () => client.get('/api/stats').then(r => r.data)
