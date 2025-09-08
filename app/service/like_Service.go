package service

import (
	"context"
	"fmt"
	models "simple-forum/app/model"
	"simple-forum/config/database"
	"time"
	"strconv"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type LikeService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewLikeService() *LikeService {
	return &LikeService{
		db:    database.DB,
		redis: database.RedisClient,
	}
}

// GetPostLikes 获取帖子点赞数
func (s *LikeService) GetPostLikes(postID uint) (int64, error) {
	// 尝试从缓存中获取点赞数
	cacheKey := fmt.Sprintf("post:%d:likes", postID)
	likesCountStr, err := s.redis.Get(context.Background(), cacheKey).Result()
	
	if err == nil {
		// 缓存命中，直接返回
		likesCount, err := strconv.ParseInt(likesCountStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return likesCount, nil
	}

	// 缓存未命中，从数据库查询
	var count int64
	if err := s.db.Model(&models.Like{}).
		Where("post_id = ?", postID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	// 将结果存入缓存，设置5分钟过期时间
	s.redis.Set(context.Background(), cacheKey, count, 5*time.Minute)

	return count, nil
}

// CheckPostExists 检查帖子是否存在
func (s *LikeService) CheckPostExists(postID uint) (bool, error) {
	var post models.Post
	if err := s.db.First(&post, postID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CheckUserExists 检查用户是否存在
func (s *LikeService) CheckUserExists(userID uint) (bool, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ToggleLike 点赞/取消点赞帖子
func (s *LikeService) ToggleLike(postID, userID uint) error {
	// Redis 键
	likeSetKey := fmt.Sprintf("post:%d:likes:users", postID)
	likeCountKey := fmt.Sprintf("post:%d:likes:count", postID)
	syncSetKey := "likes:to_sync"

	// 检查用户是否已经点赞
	isLiked, err := s.redis.SIsMember(context.Background(), likeSetKey, userID).Result()
	if err != nil {
		return err
	}

	// 使用 Redis 事务确保原子性
	pipe := s.redis.TxPipeline()

	if isLiked {
		// 取消点赞
		pipe.SRem(context.Background(), likeSetKey, userID)
		pipe.Decr(context.Background(), likeCountKey)
		// 添加到同步集合（标记为需要删除）
		pipe.SAdd(context.Background(), syncSetKey, fmt.Sprintf("remove:%d:%d", postID, userID))
	} else {
		// 点赞
		pipe.SAdd(context.Background(), likeSetKey, userID)
		pipe.Incr(context.Background(), likeCountKey)
		// 添加到同步集合（标记为需要添加）
		pipe.SAdd(context.Background(), syncSetKey, fmt.Sprintf("add:%d:%d", postID, userID))
	}

	// 执行事务
	_, err = pipe.Exec(context.Background())
	return err
}