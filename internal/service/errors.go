package service

import "errors"

// ValidationError 业务校验失败；HTTP 层统一映射 400，Message 原样给用户。
// 与存储层错误区分开：500 只留给真正的意外。
type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

// 领域错误哨兵：HTTP 层用 errors.Is 映射状态码，不再对存储层错误做字符串匹配
var (
	ErrUsernameTaken    = errors.New("用户名已存在")              // → 409
	ErrProtectedVisitor = errors.New("不能删除管理员账号")          // → 403
	ErrVisitorNotFound  = errors.New("访客不存在，请先注册")          // → 404
	ErrNotFound         = errors.New("不存在")                  // → 404
)
