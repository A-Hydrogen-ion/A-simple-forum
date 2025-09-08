package main

import (
	"context"
	"log"
	"net/http"
	models "simple-forum/app/model"
	"simple-forum/config/database"
	routes "simple-forum/config/router"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// 连接数据库
	database.ConnectDB()
	database.ConnectRedis()
	// 检查数据库连接是否成功
	if database.DB == nil {
		log.Println("数据库连接失败，程序退出")
	}

	// 检查 Redis 连接是否成功
	if database.RedisClient == nil {
		log.Println("警告: Redis 连接失败，点赞功能将不可用")
	} else {
		// 测试 Redis 连接
		_, err := database.RedisClient.Ping(context.Background()).Result()
		if err != nil {
			log.Printf("Redis 连接测试失败: %v", err)
		} else {
			log.Println("Redis 连接成功")
		}
	}
	// 自动迁移数据库
	err := database.DB.AutoMigrate(&models.User{}, &models.Post{}, &models.Report{}, &models.Like{})
	if err != nil {
		log.Printf("数据库迁移失败: %v", err)
	}

	// 只有在 Redis 连接成功时才启动点赞同步
	if database.RedisClient == nil { //此处似乎有问题,点赞相关的未实现
		go database.SyncLikes()
	}

	// 创建 Gin 引擎
	r := gin.Default()
	// go database.SyncLikes()
	// 配置 Session 中间件，使用 cookie 作为存储方式
	store := cookie.NewStore([]byte("your-secret-key-for-session"))
	store.Options(sessions.Options{
		MaxAge:   86400 * 7, // 7天有效期
		Path:     "/",
		Secure:   false, // 开发环境为 false，生产环境应为 true
		HttpOnly: true,  // 增加安全性
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("mysession", store))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"}, // 明确指定前端起源
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,                       // 设置为true，允许浏览器发送Cookie等凭证信息
		MaxAge:           12 * time.Hour,             // 预检请求缓存时间
		ExposeHeaders:    []string{"Content-Length"}, // 允许前端获取的响应头
	}))

	// 设置路由
	r = routes.SetupRouter(r)
	log.Println("服务器启动在 :8080")
	// 启动服务器
	r.Run(":8080")
}
