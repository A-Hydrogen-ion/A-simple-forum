package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	ReportStatusPending  = 0 // 待处理
	ReportStatusApproved = 1 // 已批准（帖子已处理）
	ReportStatusRejected = 2 // 已拒绝（举报无效）
)

type Report struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`          // 举报用户ID
	PostID    uint           `gorm:"not null" json:"post_id"`          // 被举报帖子ID
	Reason    string         `gorm:"type:text;not null" json:"reason"` // 举报原因
	Status    int            `gorm:"default:0" json:"status"`          // 举报状态：0-待处理，1-已处理，2-已拒绝
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除字段

	// 关联模型
	User User `gorm:"foreignKey:UserID" json:"-"`
	Post Post `gorm:"foreignKey:PostID" json:"-"`
}
type ReportRequest struct {
	PostID uint   `json:"post_id" binding:"required"`
	Reason string `json:"reason" binding:"required,min=1,max=500"`
}
// 审批记录模型
type Approval struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ReportID     uint      `gorm:"not null" json:"report_id"`     // 举报ID
	AdminID      uint      `gorm:"not null" json:"admin_id"`      // 管理员ID
	Action       int       `gorm:"not null" json:"action"`        // 操作：1-批准，2-拒绝
	Reason       string    `gorm:"type:text" json:"reason"`       // 审批理由
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	
	// 关联模型
	Report       Report    `gorm:"foreignKey:ReportID" json:"-"`
	Admin        User      `gorm:"foreignKey:AdminID" json:"-"`
}
// 举报响应结构体
type ReportResponse struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    PostID    uint      `json:"post_id"`
    Reason    string    `json:"reason"`
    Status    int       `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    Username  string    `json:"username"`  // 举报者用户名
    PostTitle string    `json:"post_title"` // 被举报帖子标题
}

// 举报列表响应结构体
type ReportListResponse struct {
    ReportList []ReportResponse `json:"report_list"`
}
// 审批请求结构体
type ApprovalRequest struct {
    ReportID uint `json:"report_id" binding:"required"`
    Approval int  `json:"approval" binding:"required,oneof=0 1"` // 0-不通过，1-通过
}
// 举报结果响应结构体
type ReportResultResponse struct {
    ID         uint      `json:"id"`
    PostID     uint      `json:"post_id"`
    Reason     string    `json:"reason"`
    Status     int       `json:"status"`
    StatusText string    `json:"status_text"` // 状态文字描述
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    PostTitle  string    `json:"post_title"`  // 被举报帖子标题
    AdminNote  string    `json:"admin_note"`  // 管理员备注
}