package controllers

import (
	"simple-forum/app/middleware"
	// "fmt"
	// "golang.org/x/crypto/bcrypt"
	"net/http"
	models "simple-forum/app/model"
	"simple-forum/config/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// func hashPassword(password string) (string, error) {
// 	// 将明文密码转换为字节切片并哈希，cost 设为 DefaultCost(10)
// 	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return "", fmt.Errorf("哈希密码失败: %w", err)
// 	}
// 	return string(hashedBytes), nil
// }
func Register(c *gin.Context) {
	var input models.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户已存在"})
		return
	}
	// 创建用户
	user := models.User{
		Username: input.Username,
		Password: input.Password, // 这里会映射到数据库的 password_hash 列
		UserType: input.Usertype, // 直接使用输入的用户类型
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": nil, "msg": "success"})

} //注册接口已实现

func Login(c *gin.Context) {
	var input models.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	var user models.User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "1Invalid credentials"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}
	// 验证密码
	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "2Invalid credentials"})
		return
	}
	//创建会话
	middleware.Login1(c, user)   

	response := models.AuthResponse{
		UserID: user.ID,
		IsAdmin:  user.UserType,
	}

	c.JSON(http.StatusOK,gin.H{"code": 200, "data":response ,"msg": "success"})
}//登录接口已实现
