// dicebear 头像 URL 的唯一构造点:换版本/换域名只改这里(此前复制 8 份)。
export function dicebearUrl(style = 'lorelei', seed = 'demo') {
  return `https://api.dicebear.com/9.x/${style}/svg?seed=${encodeURIComponent(seed)}`
}

// 实体头像:有上传头像用上传的,否则按风格+种子生成。
// entity 需含 avatar_url/avatar_style/uuid 中的任意子集(访客、评论、博主皆可)。
export function entityAvatar(entity, seedFallback = 'demo') {
  if (entity?.avatar_url) return entity.avatar_url
  return dicebearUrl(entity?.avatar_style || 'lorelei', entity?.uuid || seedFallback)
}
