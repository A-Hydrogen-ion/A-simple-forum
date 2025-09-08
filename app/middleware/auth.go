package middleware

import (
	"log"
	"net/http"
	models "simple-forum/app/model"
	"simple-forum/config/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// SessionStore 初始化会话存储
var SessionStore sessions.Store
var secretkey = "dev-secret-key-for-simple-fourm-2025-9" //开发环境使用固定密钥
func InitSessionStore() {
	// 使用 cookie 存储会话数据，密钥用于加密会话数据
	SessionStore = cookie.NewStore([]byte(secretkey))
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前会话
		session := sessions.Default(c)
		// 从会话中获取用户ID
		userID := session.Get("user_id")
		// 检查用户ID是否存在
		if userID == nil {
			// 用户未登录，返回未授权错误
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			c.Abort()
			return
		}

		var userIDUint uint
		switch v := userID.(type) {
		case uint:
			userIDUint = v
		case float64:
			userIDUint = uint(v)
		default:
			session.Clear()
			session.Save()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "会话格式无效"})
			c.Abort()
			return
		}
		var user models.User
		if err := database.DB.First(&user, userIDUint).Error; err != nil {
			session.Clear()
			session.Save()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在，请重新登录"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中，供后续处理使用
		c.Set("user", user)
		c.Set("userID", user.ID)

		// 继续处理请求
		c.Next()
	}
}

// Login1 处理用户登录
func Login1(c *gin.Context, user *models.User) {
	// 获取会话
	session := sessions.Default(c)
	if session == nil {
		log.Println("会话获取失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建会话"})
		return
	}
	// 在会话中存储用户ID
	session.Set("user_id", user.ID)
	session.Options(sessions.Options{
		MaxAge:   86400 * 7, // 7天有效期
		Path:     "/",
		Secure:   false,                // 开发环境为 false，生产环境应为 true
		HttpOnly: true,                 // 增加安全性
		SameSite: http.SameSiteLaxMode, // 根据前端需求调整
	})
	// 设置会话选项
	session.Options(sessions.Options{
		MaxAge: 86400 * 7, // 7天有效期
		Path:   "/",
	})

	// 保存会话
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建会话"})
		return
	}
}

// Logout 处理用户登出
func Logout(c *gin.Context) {
	// 获取会话
	session := sessions.Default(c)

	// 清除会话中的所有数据
	session.Clear()

	// 保存空会话
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "已成功登出"})
}
