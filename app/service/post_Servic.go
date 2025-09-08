// app/service/post_service.go
package service

import (
	"fmt"
	"log"
	models "simple-forum/app/model"
	"simple-forum/config/database"
	"time"

	"gorm.io/gorm"
)

type PostService struct {
	db *gorm.DB
}

func NewPostService() *PostService {
	if database.DB == nil {
		log.Println("警告: database.DB 为 nil，PostService 将无法工作")
	}
	return &PostService{db: database.DB}
}

// CreatePost 创建帖子
func (s *PostService) CreatePost(post *models.Post) error {
	if s.db == nil {
		return fmt.Errorf("数据库连接未初始化 (s.db is nil)")
	}
	
	err := s.db.Create(post).Error
	if err != nil {
		log.Printf("创建帖子数据库错误: %v", err)
		return err
	}
	
	return nil
}

// GetPosts 获取所有帖子（按创建时间降序排列）
func (s *PostService) GetPosts() ([]models.Post, error) {
	var posts []models.Post
	err := s.db.Order("created_at desc").Find(&posts).Error
	return posts, err
}

// GetPostByID 根据ID获取帖子
func (s *PostService) GetPostByID(id string) (models.Post, error) {
	var post models.Post
	err := s.db.First(&post, id).Error
	return post, err
}

// GetDeletedPostByID 获取已删除的帖子（使用Unscoped）
func (s *PostService) GetDeletedPostByID(id string) (models.Post, error) {
	var post models.Post
	err := s.db.Unscoped().First(&post, id).Error
	return post, err
}

// DeletePost 软删除帖子
func (s *PostService) DeletePost(post *models.Post) error {
	return s.db.Delete(post).Error
}

// RestorePost 恢复已删除的帖子
func (s *PostService) RestorePost(post *models.Post) error {
	return s.db.Unscoped().Model(post).Update("deleted_at", nil).Error
}

// UpdatePostContent 更新帖子内容
func (s *PostService) UpdatePostContent(post *models.Post, content string) error {
	return s.db.Model(post).Updates(map[string]interface{}{
		"content":    content,
		"updated_at": time.Now(),
	}).Error
}
