package controllers

import (
	"net/http"
	"simple-forum/app/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var likeService = service.NewLikeService()

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

	// 使用Service获取点赞数
	count, err := likeService.GetPostLikes(uint(postID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞数失败"})
		return
	}
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"likes": count,
		},
		"msg": "success",
	})
}

// Like 点赞/取消点赞帖子
func Like(c *gin.Context) {
	var input struct {
		PostID uint `json:"post_id" binding:"required"`
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用Service检查帖子是否存在
	postExists, err := likeService.CheckPostExists(input.PostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查帖子失败"})
		return
	}
	if !postExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}
	// 使用Service检查用户是否存在
	userExists, err := likeService.CheckUserExists(input.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查用户失败"})
		return
	}
	if !userExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 使用Service进行点赞/取消点赞操作
	err = likeService.ToggleLike(input.PostID, input.UserID)
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
