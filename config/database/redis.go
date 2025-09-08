package database

import (
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 服务器地址
		Password: "",               // 密码
		DB:       0,                // 使用默认DB
	})

	// 测试连接
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("连接Redis失败", err)
	}

	fmt.Println("Redis连接成功")
}