package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mosesyu95/redislock"
	"time"
)

func main() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	redislock.Init("prefix:", client)
	// 创建锁（TTL=10s，自动续期）
	lock := redislock.NewRedisLock("my_key", "my_lock", true, 10*time.Second)

	// 尝试获取锁
	success, err := lock.TryLock(ctx)
	if err != nil {
		fmt.Println("❌ 获取锁失败:", err)
		return
	}
	if success {
		fmt.Println("✅ 获取锁成功！")
	} else {
		fmt.Println("⛔️ 锁已被占用")
		return
	}

	// 模拟业务逻辑
	time.Sleep(15 * time.Second)

	// 释放锁
	lock.Unlock(ctx)

	fmt.Println("🔓 释放锁成功！")

}
