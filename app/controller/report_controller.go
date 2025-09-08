package controllers

import (
	"net/http"
	models "simple-forum/app/model"
	"simple-forum/config/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReportPost 举报帖子
func ReportPost(c *gin.Context) {
	var input models.ReportRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取当前用户
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)

	// 检查帖子是否存在
	var post models.Post
	if err := database.DB.First(&post, input.PostID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询帖子失败"})
		}
		return
	}

	// 检查用户是否已经举报过该帖子
	var existingReport models.Report
	if err := database.DB.Where("user_id = ? AND post_id = ?", user.ID, input.PostID).First(&existingReport).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您已经举报过该帖子"})
		return
	}

	// 创建举报记录
	report := models.Report{
		UserID: user.ID,
		PostID: input.PostID,
		Reason: input.Reason,
		Status: models.ReportStatusPending, // 初始状态为待处理
	}

	if err := database.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "举报失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": nil,
		"msg":  "success",
	})
}

//ViewReportApproval 获取所有未审批的举报
func ViewReportApproval(c *gin.Context) {
	// 从上下文中获取当前用户
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)
	// 检查用户类型是否为管理员 (管理员usertyp为2)
	if user.UserType != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
		return
	}

	// 获取所有待处理的举报
	var reports []models.Report
	if err := database.DB.
		Where("status = ?", models.ReportStatusPending).
		Order("created_at desc").
		Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取举报列表失败"})
		return
	}

	// 转换为响应格式
	var reportList []models.ReportResponse
	for _, report := range reports {
		// 获取举报者用户名
		var reporter models.User
		database.DB.First(&reporter, report.UserID)

		// 获取被举报帖子标题
		var post models.Post
		database.DB.First(&post, report.PostID)

		reportList = append(reportList, models.ReportResponse{
			ID:        report.ID,
			UserID:    report.UserID,
			PostID:    report.PostID,
			Reason:    report.Reason,
			Status:    report.Status,
			CreatedAt: report.CreatedAt,
			Username:  reporter.Username,
			PostTitle: post.Title,
		})
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": models.ReportListResponse{
			ReportList: reportList,
		},
		"msg": "success",
	})
}

// ApproveReport 管理员审批举报
func ApproveReport(c *gin.Context) {
	var input models.ApprovalRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取当前用户
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)

	// 检查用户是否是管理员
	if user.UserType != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 获取举报信息
	var report models.Report
	if err := database.DB.First(&report, input.ReportID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "举报记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询举报记录失败"})
		}
		return
	}

	// 检查举报是否已被处理
	if report.Status != models.ReportStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该举报已被处理"})
		return
	}

	// 开始数据库事务
	tx := database.DB.Begin()

	// 根据审批结果进行处理
	if input.Approval == 1 {
		// 审批通过，删除被举报的帖子
		var post models.Post
		if err := tx.First(&post, report.PostID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "被举报的帖子不存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查询帖子失败"})
			}
			return
		}

		// 软删除帖子
		if err := tx.Delete(&post).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除帖子失败"})
			return
		}

		// 更新举报状态为已批准
		report.Status = models.ReportStatusApproved
	} else {
		// 审批不通过，更新举报状态为已拒绝
		report.Status = models.ReportStatusRejected
	}

	// 更新举报记录
	if err := tx.Save(&report).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新举报状态失败"})
		return
	}

	// 提交事务
	tx.Commit()

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": nil,
		"msg":  "审批成功",
	})
}

// GetReportResults 用户查看举报审核结果
func GetReportResults(c *gin.Context) {
	// 从上下文中获取当前用户
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	user := currentUser.(models.User)

	// 获取该用户的所有举报记录
	var reports []models.Report
	if err := database.DB.
		Where("user_id = ?", user.ID).
		Order("created_at desc").
		Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取举报记录失败"})
		return
	}

	// 转换为响应格式
	var reportResults []models.ReportResultResponse
	for _, report := range reports {
		// 获取被举报帖子标题
		var post models.Post
		database.DB.Unscoped().First(&post, report.PostID) // 使用Unscoped获取已删除的帖子

		// 确定状态文字描述
		statusText := getStatusText(report.Status)

		reportResults = append(reportResults, models.ReportResultResponse{
			ID:         report.ID,
			PostID:     report.PostID,
			Reason:     report.Reason,
			Status:     report.Status,
			StatusText: statusText,
			CreatedAt:  report.CreatedAt,
			UpdatedAt:  report.UpdatedAt,
			PostTitle:  post.Title,
			AdminNote:  "", // 如果有管理员备注字段，可以在这里添加
		})
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"report_results": reportResults,
		},
		"msg": "success",
	})
}

// 辅助函数：获取状态文字描述
func getStatusText(status int) string {
	switch status {
	case models.ReportStatusPending:
		return "待处理"
	case models.ReportStatusApproved:
		return "已通过"
	case models.ReportStatusRejected:
		return "已拒绝"
	default:
		return "未知状态"
	}
}
