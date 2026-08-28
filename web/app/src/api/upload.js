// 图片上传 API：接受 File 或已装好的 FormData，返回 { url }
import client from './client'

export function uploadFile(file) {
  let fd
  if (file instanceof FormData) {
    fd = file
  } else {
    fd = new FormData()
    fd.append('file', file)
  }
  return client.post('/api/upload', fd).then(r => r.data)
}
