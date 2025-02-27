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

	// **等待锁可用**
	fmt.Println("⏳ 正在等待锁释放...")
	err := lock.Lock(ctx, time.Minute) // 这里改为 Lock，确保锁可用时才继续
	if err != nil {
		fmt.Println("❌ 获取锁失败:", err)
		return
	}
	fmt.Println("✅ 获取锁成功！")

	// 模拟业务逻辑
	time.Sleep(15 * time.Second)

	// 释放锁
	lock.Unlock(ctx)
	fmt.Println("🔓 释放锁成功！")
}
