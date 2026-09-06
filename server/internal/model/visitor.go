package model

import "time"

type Visitor struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	UUID     string `json:"uuid" gorm:"uniqueIndex;size:64"`
	Username string `json:"username" gorm:"size:50;index"`
	Password string `json:"-" gorm:"size:100"`
	// 角色：user 普通用户（前台互动），admin 管理员（可进后台）
	Role        string    `json:"role" gorm:"size:10;default:user"`
	Nickname    string    `json:"nickname" gorm:"size:50"`
	AvatarStyle string    `json:"avatar_style" gorm:"size:20;default:lorelei"`
	AvatarURL   string    `json:"avatar_url" gorm:"size:500"`
	Signature   string    `json:"signature" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at"`
}
