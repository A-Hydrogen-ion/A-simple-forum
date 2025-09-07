package controllers
//点赞写了不代表能用（（（（
//后面不太会用apifox调试了（悲
import (
	"fmt"
	"net/http"
	models "simple-forum/app/model"
	"simple-forum/config/database"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

// GetPostLikes 获取帖子点赞数
func GetPostLikes(c *gin.Context) {
	// 从查询参数中获取帖子ID
	postIDStr := c.Query("post_id")
	if postIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID格式错误"})
		return
	}

	// 尝试从缓存中获取点赞数
	cacheKey := fmt.Sprintf("post:%d:likes", postID)
	likesCount, err := database.RedisClient.Get(context.Background(), cacheKey).Uint64()

	if err == nil {
		// 缓存命中，直接返回
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"likes": likesCount,
			},
			"msg": "success",
		})
		return
	}

	// 缓存未命中，从数据库查询
	var count int64
	if err := database.DB.Model(&models.Like{}).
		Where("post_id = ?", postID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞数失败"})
		return
	}

	// 将结果存入缓存，设置5分钟过期时间
	database.RedisClient.Set(context.Background(), cacheKey, count, 5*time.Minute)

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"likes": count,
		},
		"msg": "success",
	})
}

// ToggleLike 点赞/取消点赞帖子
func Like(c *gin.Context) {
	var input struct {
		PostID uint `json:"post_id" binding:"required"`
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查帖子是否存在
	var post models.Post
	if err := database.DB.First(&post, input.PostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := database.DB.First(&user, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// Redis 键
	likeSetKey := fmt.Sprintf("post:%d:likes:users", input.PostID)
	likeCountKey := fmt.Sprintf("post:%d:likes:count", input.PostID)
	syncSetKey := "likes:to_sync"

	// 检查用户是否已经点赞
	isLiked, err := database.RedisClient.SIsMember(database.Ctx, likeSetKey, input.UserID).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查点赞状态失败"})
		return
	}

	// 使用 Redis 事务确保原子性
	pipe := database.RedisClient.TxPipeline()

	if isLiked {
		// 取消点赞
		pipe.SRem(database.Ctx, likeSetKey, input.UserID)
		pipe.Decr(database.Ctx, likeCountKey)
		// 添加到同步集合（标记为需要删除）
		pipe.SAdd(database.Ctx, syncSetKey, fmt.Sprintf("remove:%d:%d", input.PostID, input.UserID))
	} else {
		// 点赞
		pipe.SAdd(database.Ctx, likeSetKey, input.UserID)
		pipe.Incr(database.Ctx, likeCountKey)
		// 添加到同步集合（标记为需要添加）
		pipe.SAdd(database.Ctx, syncSetKey, fmt.Sprintf("add:%d:%d", input.PostID, input.UserID))
	}

	// 执行事务
	_, err = pipe.Exec(database.Ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": struct{}{},
		"msg":  "success",
	})
}
