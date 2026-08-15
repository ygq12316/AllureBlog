package model

import "time"

type Comment struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	NoteID      uint      `json:"note_id" gorm:"index"`
	VisitorUUID string    `json:"visitor_uuid" gorm:"size:64;index"`
	Content     string    `json:"content" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
}

type CommentWithVisitor struct {
	Comment
	Nickname    string `json:"nickname"`
	AvatarStyle string `json:"avatar_style"`
	AvatarURL   string `json:"avatar_url"`
	Signature   string `json:"signature"`
}