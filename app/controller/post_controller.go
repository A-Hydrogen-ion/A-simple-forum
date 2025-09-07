package controllers

import (
	"net/http"
	models "simple-forum/app/model"
	"simple-forum/config/database"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePost 发布帖子
func CreatePost(c *gin.Context) {
	var input models.CreatePostRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取当前用户 (通过中间件设置)
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)

	// 创建帖子
	post := models.Post{
		Title:    input.Title,
		Content:  input.Content,
		UserID:   user.ID,
		Username: user.Username, // 设置作者ID
	}

	if err := database.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布帖子失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": nil,
		"msg":  "帖子发布成功",
	})
}

// 可以添加其他帖子相关功能，如获取帖子列表、更新帖子、删除帖子等

// GetPosts 获取所有帖子（基础版本，点赞数暂设为0）
func GetPosts(c *gin.Context) {
	var posts []models.Post

	// 从数据库获取所有帖子，按创建时间降序排列
	if err := database.DB.Order("created_at desc").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取帖子列表失败"})
		return
	}

	// 转换为响应格式
	var postList []models.PostResponse
	for _, post := range posts {
		postList = append(postList, models.PostResponse{
			ID:      post.ID,
			Content: post.Content,
			UserID:  post.UserID,
			Time:    post.CreatedAt.Format("2006-01-02T15:04:05.999Z07:00"),
			Likes:   0, // 暂时设为0，等点赞功能完成后再改为真实数据
		})
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": models.PostListResponse{
			PostList: postList,
		},
		"msg": "success",
	})
}

// DeletePost 软删除帖子
func DeletePost(c *gin.Context) {
	// 从查询参数中获取 post_id
	postID := c.Query("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	// 从上下文中获取当前用户 (通过中间件设置)
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)

	// 查找帖子
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询帖子失败"})
		}
		return
	}

	// 检查用户是否有权限删除该帖子
	if post.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能删除自己的帖子"})
		return
	}

	// 软删除帖子（使用Delete方法而不是Unscoped().Delete）
	if err := database.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除帖子失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": nil,
		"msg":  "success",
	})
}

// RestorePost 恢复已删除的帖子
func RestorePost(c *gin.Context) {
	// 从查询参数中获取 post_id
	postID := c.Query("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
		return
	}

	// 从上下文中获取当前用户 (通过中间件设置)
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)

	// 查找已删除的帖子（使用Unscoped()来查询包括已删除的记录）
	var post models.Post
	if err := database.DB.Unscoped().First(&post, postID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询帖子失败"})
		}
		return
	}

	// 检查用户是否有权限恢复该帖子
	if post.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能恢复自己的帖子"})
		return
	}

	// 恢复帖子
	if err := database.DB.Unscoped().Model(&post).Update("deleted_at", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复帖子失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// UpdatePost 修改帖子内容
func UpdatePost(c *gin.Context) {
	var input models.UpdatePostRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取当前用户 (通过中间件设置)
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)

	// 查找帖子
	var post models.Post
	if err := database.DB.First(&post, input.PostID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询帖子失败"})
		}
		return
	}

	// 检查用户是否有权限修改该帖子
	if post.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能修改自己的帖子"})
		return
	}

	// 更新帖子内容
	if err := database.DB.Model(&post).Update("content", input.Content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新帖子失败"})
		return
	}

	// 同时更新 updated_at 时间戳
	database.DB.Model(&post).Update("updated_at", time.Now())

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": nil,
		"msg":  "success",
	})
}
