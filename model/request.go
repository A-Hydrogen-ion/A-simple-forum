package models

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=10"`
	Name     string `json:"name"     binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=8,max=16"`
	Usertype int    `json:"user_type"  binding:"required,oneof=1 2"` //1代表学生，2代表管理员
} //RegisterRequest结构体将用于处理用户注册请求的数据绑定和验证
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
} //LoginRequest结构体将用于处理用户登录请求的数据绑定和验证
type AuthResponse struct {
	UserID   uint   `json:"user_id"` 
	IsAdmin  int    `json:"user_type"`
} //？
type Response struct {
	Data string `json:"data"`
	Msg  string `json:"msg"`
}
