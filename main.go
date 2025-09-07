package main

import (
	models "simple-forum/model"
	"simple-forum/pkg/database"
	routes "simple-forum/router"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// 连接数据库
	database.ConnectDB()
	// 自动迁移数据库 schema
	database.DB.AutoMigrate(&models.User{}, &models.Post{}, &models.Report{}, &models.Like{})
	go database.SyncLikes()
	// 创建 Gin 引擎
	r := gin.Default()

	// 配置 Session 中间件，使用 cookie 作为存储方式
	store := cookie.NewStore([]byte("your-secret-key")) // 替换为一个安全的密钥
	r.Use(sessions.Sessions("mysession", store))        // "mysession" 是 cookie 的名称
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"}, // 明确指定你的前端起源，不能用通配符"*"
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,                       // 设置为true，允许浏览器发送Cookie等凭证信息
		MaxAge:           12 * time.Hour,             // 预检请求缓存时间
		ExposeHeaders:    []string{"Content-Length"}, // 允许前端获取的响应头
	}))

	// 设置路由
	r = routes.SetupRouter(r)

	// 启动服务器
	r.Run(":8080")
}
