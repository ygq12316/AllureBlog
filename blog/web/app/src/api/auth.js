// 管理员登录 API
import client from './client'

export const login = (username, password) => client.post('/api/login', { username, password }).then(r => r.data)
