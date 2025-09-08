package controllers

import (
	"fmt"
	"net/http"
	"simple-forum/app/middleware"
	models "simple-forum/app/model"
	"simple-forum/app/service"

	"github.com/gin-gonic/gin"
)

func getUserService() *service.UserService {
	return service.NewUserService()
}
func Register(c *gin.Context) {
	var input models.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	exists, err := service.NewUserService().CheckUsernameExists(input.Username)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询错误"})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户已存在"})
		return
	}
	// 创建用户
	user := models.User{
		Username: input.Username,
		Password: input.Password, // 这里会映射到数据库的 password_hash 列
		UserType: input.Usertype, // 直接使用输入的用户类型
	}

	if err := service.NewUserService().CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": nil, "msg": "success"})

}

func Login(c *gin.Context) {
	var input models.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userService := getUserService()
	// 查找用户
	user, err := userService.GetUserByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	//创建会话
	middleware.Login1(c, user)

	response := models.AuthResponse{
		UserID:  user.ID,
		IsAdmin: user.UserType,
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": response, "msg": "success"})
}
