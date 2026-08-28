package model

import "time"

type BlogConfig struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AuthorName   string    `json:"author_name" gorm:"size:50;default:'Allure'"`
	AuthorAvatar string    `json:"author_avatar" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}