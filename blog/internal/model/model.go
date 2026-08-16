package model

import "time"

type Article struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Content     string    `gorm:"not null" json:"content"`
	HTML        string    `json:"html"`
	Excerpt     string    `json:"excerpt"`
	Category    string    `json:"category"`
	Tags        string    `json:"tags"` // comma-separated
	IsPublished bool      `gorm:"default:false" json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Category struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	Slug         string `gorm:"uniqueIndex;not null" json:"slug"`
	ArticleCount int    `gorm:"default:0" json:"article_count"`
}

type Tag struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	Slug         string `gorm:"uniqueIndex;not null" json:"slug"`
	ArticleCount int    `gorm:"default:0" json:"article_count"`
}

type Note struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Content     string    `gorm:"not null" json:"content"`
	HTML        string    `json:"html"`
	Images      string    `json:"images"` // comma-separated URLs, max 9
	IsPublished bool      `gorm:"default:false" json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	// CommentCount 由列表查询的子查询填充，不落库（-> 只读禁写，-:migration 跳过建列）
	CommentCount int64 `json:"comment_count" gorm:"->;-:migration"`
}
