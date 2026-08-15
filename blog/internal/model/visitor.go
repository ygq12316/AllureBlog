package model

import "time"

type Visitor struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UUID        string    `json:"uuid" gorm:"uniqueIndex;size:64"`
	Username    string    `json:"username" gorm:"size:50;index"`
	Password    string    `json:"-" gorm:"size:100"`
	Nickname    string    `json:"nickname" gorm:"size:50"`
	AvatarStyle string    `json:"avatar_style" gorm:"size:20;default:lorelei"`
	AvatarURL   string    `json:"avatar_url" gorm:"size:500"`
	Signature   string    `json:"signature" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at"`
}
