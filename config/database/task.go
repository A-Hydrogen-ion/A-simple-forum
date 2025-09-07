package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// SyncLikes 同步点赞数据到数据库
func SyncLikes() {
	ctx := context.Background()
	syncSetKey := "likes:to_sync"

	for {
		// 每30秒同步一次
		time.Sleep(30 * time.Second)

		// 获取所有需要同步的点赞操作
		operations, err := RedisClient.SMembers(ctx, syncSetKey).Result()
		if err != nil {
			log.Printf("获取同步集合失败: %v", err)
			continue
		}

		if len(operations) == 0 {
			continue
		}

		// 处理每个操作
		for _, op := range operations {
			parts := strings.Split(op, ":")
			if len(parts) != 3 {
				continue
			}

			action := parts[0]
			postID := parts[1]
			userID := parts[2]

			// 转换为整数
			var postIDUint, userIDUint uint
			fmt.Sscanf(postID, "%d", &postIDUint)
			fmt.Sscanf(userID, "%d", &userIDUint)

			// 根据操作类型执行不同的数据库操作
			if action == "add" {
				// 使用创建或忽略的方式，避免重复记录
				if err := DB.Exec(
					"INSERT IGNORE INTO likes (post_id, user_id, created_at) VALUES (?, ?, NOW())",
					postIDUint, userIDUint,
				).Error; err != nil {
					log.Printf("添加点赞记录失败: %v", err)
					continue
				}
			} else if action == "remove" {
				// 删除点赞记录
				if err := DB.Exec(
					"DELETE FROM likes WHERE post_id = ? AND user_id = ?",
					postIDUint, userIDUint,
				).Error; err != nil {
					log.Printf("删除点赞记录失败: %v", err)
					continue
				}
			}

			// 从同步集合中移除已处理的操作
			RedisClient.SRem(Ctx, syncSetKey, op)
		}

		log.Printf("同步了 %d 个点赞操作", len(operations))
	}
}
