package common

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"
	"github.com/go-redis/redis/v8"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
)

var RDB redis.Cmdable

var redisEnabled atomic.Bool

func init() {
	redisEnabled.Store(true)
}

func IsRedisEnabled() bool {
	return redisEnabled.Load()
}

func SetRedisEnabled(enabled bool) {
	redisEnabled.Store(enabled)
}

// InitRedisClient This function is called after init()

func InitRedisClient() (err error) {
	if config.RedisConnString == "" {
		SetRedisEnabled(false)
		logger.Logger.Info("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil
	}
	if config.SyncFrequency == 0 {
		SetRedisEnabled(false)
		logger.Logger.Info("SYNC_FREQUENCY not set, Redis is disabled")
		return nil
	}
	redisConnString := config.RedisConnString
	if config.RedisMasterName == "" {
		logger.Logger.Info("Redis is enabled")
		opt, err := redis.ParseURL(redisConnString)
		if err != nil {
			logger.Logger.Fatal("failed to parse Redis connection string", zap.Error(err))
		}
		RDB = redis.NewClient(opt)
	} else {
		// cluster mode
		logger.Logger.Info("Redis cluster mode enabled")
		RDB = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      strings.Split(redisConnString, ","),
			Password:   config.RedisPassword,
			MasterName: config.RedisMasterName,
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RDB.Ping(ctx).Result()
	if err != nil {
		logger.Logger.Fatal("Redis ping test failed", zap.Error(err))
	}
	SetRedisEnabled(true)
	return nil
}

func ParseRedisOption() *redis.Options {
	opt, err := redis.ParseURL(config.RedisConnString)
	if err != nil {
		logger.Logger.Fatal("failed to parse Redis connection string", zap.Error(err))
	}
	return opt
}

func RedisSet(key string, value string, expiration time.Duration) error {
	ctx := context.Background()
	if RDB == nil {
		return errors.New("redis not initialized")
	}
	err := RDB.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return errors.Wrapf(err, "failed to set redis key: %s", key)
	}
	return nil
}

func RedisGet(key string) (string, error) {
	ctx := context.Background()
	if RDB == nil {
		return "", errors.New("redis not initialized")
	}
	val, err := RDB.Get(ctx, key).Result()
	if err != nil {
		return "", errors.Wrapf(err, "failed to get redis key: %s", key)
	}
	return val, nil
}

func RedisDel(key string) error {
	ctx := context.Background()
	if RDB == nil {
		return errors.New("redis not initialized")
	}
	err := RDB.Del(ctx, key).Err()
	if err != nil {
		return errors.Wrapf(err, "failed to delete redis key: %s", key)
	}
	return nil
}

func RedisDecrease(key string, value int64) error {
	ctx := context.Background()
	if RDB == nil {
		return errors.New("redis not initialized")
	}
	err := RDB.DecrBy(ctx, key, value).Err()
	if err != nil {
		return errors.Wrapf(err, "failed to decrease redis key: %s by %d", key, value)
	}
	return nil
}
