package models

import (
	"time"

	"gorm.io/gorm"
)
//帖子结构体
type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"size:200;not null" json:"title" binding:"required"`
	Content   string         `gorm:"type:text;not null" json:"content" binding:"required"`
	UserID    uint           `gorm:"not null" json:"user_id"`          // 关联用户ID
	Username  string         `gorm:"size:50;not null" json:"username"` // 关联用户
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 发布帖子请求结构体
type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
}

// 返回帖子请求结构体
type PostResponse struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
	Time    string `json:"time"`
	Likes   int    `json:"likes"`
}
// 返回的帖子列表结构体
type PostListResponse struct {
	PostList []PostResponse `json:"post_list"`
}

// 更新帖子请求结构体
type UpdatePostRequest struct {
    PostID  uint   `json:"post_id" binding:"required"`
    Content string `json:"content" binding:"required,min=1"`
}