package redislock

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strings"
	"sync"
	"time"
)

type RedisLock struct {
	client   redis.UniversalClient
	key      string
	value    string
	ttl      time.Duration
	unlockCh chan struct{}
	mu       sync.Mutex
	keep     bool
}

// NewRedisLock 创建一个新的 Redis 分布式锁实例
func NewRedisLock(key, value string, keepAlive bool, ttl time.Duration) *RedisLock {
	if ttl <= 3*time.Second {
		panic("redisLock ttl must be greater than 3 seconds")
	}
	if !strings.HasPrefix(key, PreString) {
		key = PreString + key
	}
	return &RedisLock{
		client:   redisClient,
		key:      key,
		value:    value,
		ttl:      ttl,
		unlockCh: make(chan struct{}),
		keep:     keepAlive,
	}
}

// TryLock 尝试获取锁，不阻塞
func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	// 拼接 key:value 形式
	lockKey := fmt.Sprintf("%s:%s", l.key, l.value)

	// 使用 SET NX 获取锁
	ok, err := l.client.SetNX(ctx, lockKey, "", l.ttl).Result()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	if l.keep {
		// 启动续期协程
		go l.keepAlive(ctx)
	}
	return true, nil
}

// Lock 阻塞直到获取锁，支持超时
func (l *RedisLock) Lock(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		log.Println("failed to acquire lock, retrying...")
		time.Sleep(100 * time.Millisecond) // 100ms 轮询
	}

	return errors.New("failed to acquire lock")
}

// Unlock 释放锁
func (l *RedisLock) Unlock(ctx context.Context) error {
	// 拼接 key:value 形式
	lockKey := fmt.Sprintf("%s:%s", l.key, l.value)

	// 直接删除 Key
	_, err := l.client.Del(ctx, lockKey).Result()
	if err != nil {
		return err
	}

	// 关闭续期协程（避免 panic）
	select {
	case <-l.unlockCh:
	default:
		close(l.unlockCh)
	}

	return nil
}

// 续期锁，防止因 TTL 到期释放
func (l *RedisLock) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(l.ttl / 2) // 每 50% TTL 续期
	defer ticker.Stop()
	lockKey := fmt.Sprintf("%s:%s", l.key, l.value)

	for {
		select {
		case <-ticker.C:
			l.client.Expire(ctx, lockKey, l.ttl)
		case <-l.unlockCh:
			return
		}
	}
}
