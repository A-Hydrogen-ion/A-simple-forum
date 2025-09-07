package models

import (
	"time"
)

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null;index" json:"post_id"` // 帖子ID
	UserID    uint      `gorm:"not null;index" json:"user_id"` // 用户ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联模型
	Post Post `gorm:"foreignKey:PostID" json:"-"`
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// 点赞请求结构体
type LikeRequest struct {
	PostID uint `json:"post_id" binding:"required"`
}
