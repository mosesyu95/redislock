package redislock

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

var (
	redisClient  redis.UniversalClient
	once         sync.Once
	DialTimeout  = 5 * time.Second
	ReadTimeout  = 3 * time.Second
	WriteTimeout = 3 * time.Second
	PoolSize     = 10
	PreString    = ""
)

// Init 初始化 Redis 客户端（支持单实例、哨兵、集群）
// preString: 前缀字符串
func Init(preString string, client redis.UniversalClient) {
	PreString = preString
	redisClient = client
}

// InitPrefixString 初始化前缀字符串
// preString: 前缀字符串
func InitPrefixString(preString string) {
	PreString = preString
}

// InitRedis 初始化 Redis 客户端（支持单实例、哨兵、集群）
// redisType: 连接模式，支持 "single", "sentinel", "cluster"
// addresses: Redis 地址，格式为 "host:port" , 可以是多个地址
// password: Redis 密码
// masterName: 哨兵模式下的主节点名称
// db: Redis 数据库编号
// 返回值：
// redis.UniversalClient: Redis 客户端实例
// error: 初始化错误
func InitRedis(redisType string, addresses []string, password, masterName string, db int, ctx context.Context) (redis.UniversalClient, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses provided")
	}
	var err error
	once.Do(func() {
		switch redisType {
		case "single":
			redisClient = redis.NewClient(&redis.Options{
				Addr:         addresses[0],
				Password:     password,
				DB:           db,
				DialTimeout:  DialTimeout,
				ReadTimeout:  ReadTimeout,
				WriteTimeout: WriteTimeout,
				PoolSize:     PoolSize,
			})

		case "sentinel":
			redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:    masterName,
				SentinelAddrs: addresses,
				Password:      password,
				DB:            db,
				DialTimeout:   DialTimeout,
				ReadTimeout:   ReadTimeout,
				WriteTimeout:  WriteTimeout,
				PoolSize:      PoolSize,
			})

		case "cluster":
			redisClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:        addresses,
				Password:     password,
				DialTimeout:  DialTimeout,
				ReadTimeout:  ReadTimeout,
				WriteTimeout: WriteTimeout,
				PoolSize:     PoolSize,
			})

		default:
			err = fmt.Errorf("unknown redis type %s", redisType)
			return
		}

		// 测试连接
		if err = redisClient.Ping(ctx).Err(); err != nil {
			err = fmt.Errorf("failed to connect to Redis: %v", err)
		}
	})
	return redisClient, err
}

// GetRedisClient 获取全局 Redis 客户端
func GetRedisClient() redis.UniversalClient {
	return redisClient
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() {
	if redisClient != nil {
		redisClient.Close()
	}
}
