package model

import "time"

type Danmaku struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	VisitorUUID string    `json:"visitor_uuid" gorm:"size:64;index"`
	Content     string    `json:"content" gorm:"size:100"`
	Color       string    `json:"color" gorm:"size:20;default:#b8944c"`
	CreatedAt   time.Time `json:"created_at"`
}

type DanmakuWithVisitor struct {
	Danmaku
	Nickname    string `json:"nickname"`
	AvatarStyle string `json:"avatar_style"`
}
