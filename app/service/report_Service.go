// app/service/report_service.go
package service

import (
	models "simple-forum/app/model"
	"simple-forum/config/database"

	"gorm.io/gorm"
)

type ReportService struct {
	db *gorm.DB
}

func NewReportService() *ReportService {
	return &ReportService{db: database.DB}
}

// CheckPostExists 检查帖子是否存在
func (s *ReportService) CheckPostExists(postID uint) (bool, error) {
	var post models.Post
	if err := s.db.First(&post, postID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CheckReportExists 检查用户是否已经举报过该帖子
func (s *ReportService) CheckReportExists(userID, postID uint) (bool, error) {
	var existingReport models.Report
	if err := s.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingReport).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateReport 创建举报记录
func (s *ReportService) CreateReport(report *models.Report) error {
	return s.db.Create(report).Error
}

// GetPendingReports 获取所有待处理的举报
func (s *ReportService) GetPendingReports() ([]models.Report, error) {
	var reports []models.Report
	err := s.db.
		Where("status = ?", models.ReportStatusPending).
		Order("created_at desc").
		Find(&reports).Error
	return reports, err
}

// GetReportByID 根据ID获取举报信息
func (s *ReportService) GetReportByID(reportID uint) (models.Report, error) {
	var report models.Report
	err := s.db.First(&report, reportID).Error
	return report, err
}

// UpdateReportStatus 更新举报状态
func (s *ReportService) UpdateReportStatus(report *models.Report, status int) error {
	return s.db.Model(report).Update("status", status).Error
}

// DeletePost 删除帖子
func (s *ReportService) DeletePost(post *models.Post) error {
	return s.db.Delete(post).Error
}

// GetReportsByUserID 获取用户的所有举报记录
func (s *ReportService) GetReportsByUserID(userID uint) ([]models.Report, error) {
	var reports []models.Report
	err := s.db.
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&reports).Error
	return reports, err
}

// GetPostByID 根据ID获取帖子（包括已删除的）
func (s *ReportService) GetPostByID(postID uint) (models.Post, error) {
	var post models.Post
	err := s.db.Unscoped().First(&post, postID).Error
	return post, err
}

// GetUserByID 根据ID获取用户
func (s *ReportService) GetUserByID(userID uint) (models.User, error) {
	var user models.User
	err := s.db.First(&user, userID).Error
	return user, err
}