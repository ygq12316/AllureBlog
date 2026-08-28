// 时间与文本格式化:此前 rel() 复制 4 份且输出风格已分叉,统一为一种中文风格。
export function rel(d) {
  if (!d) return ''
  const s = Math.floor((Date.now() - new Date(d).getTime()) / 1000)
  if (s < 60) return '刚刚'
  if (s < 3600) return Math.floor(s / 60) + '分钟前'
  if (s < 86400) return Math.floor(s / 3600) + '小时前'
  if (s < 2592000) return Math.floor(s / 86400) + '天前'
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

export function fmtDate(d, opts = { month: 'short', day: 'numeric' }) {
  return d ? new Date(d).toLocaleDateString('zh-CN', opts) : ''
}

// 去 HTML 标签后截断
export function trunc(html, n) {
  const t = String(html ?? '').replace(/<[^>]*>/g, '')
  return t.length > n ? t.slice(0, n) + '…' : t
}
