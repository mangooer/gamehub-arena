package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisService struct {
	client *RedisClient
}

func NewRedisService(client *RedisClient) CacheService {
	return &redisService{client: client}
}

func (s *redisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data = string(jsonData)
	}
	return s.client.client.Set(ctx, key, data, expiration).Err()
}

func (r *redisService) Get(ctx context.Context, key string) (string, error) {
	return r.client.client.Get(ctx, key).Result()
}

func (r *redisService) Del(ctx context.Context, keys ...string) error {
	return r.client.client.Del(ctx, keys...).Err()
}

func (r *redisService) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.client.Exists(ctx, keys...).Result()
}

func (r *redisService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.client.Expire(ctx, key, expiration).Err()
}

// Hash操作实现
func (r *redisService) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.client.HSet(ctx, key, values...).Err()
}

func (r *redisService) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.client.HGet(ctx, key, field).Result()
}

func (r *redisService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.client.HGetAll(ctx, key).Result()
}

func (r *redisService) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.client.HDel(ctx, key, fields...).Err()
}

// Set操作实现
func (r *redisService) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.client.SAdd(ctx, key, members...).Err()
}

func (r *redisService) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.client.SRem(ctx, key, members...).Err()
}

func (r *redisService) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.client.SMembers(ctx, key).Result()
}

func (r *redisService) SCard(ctx context.Context, key string) (int64, error) {
	return r.client.client.SCard(ctx, key).Result()
}

// Sorted Set操作实现
func (r *redisService) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.client.client.ZAdd(ctx, key, members...).Err()
}

func (r *redisService) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.client.ZRem(ctx, key, members...).Err()
}

func (r *redisService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.client.ZRange(ctx, key, start, stop).Result()
}

func (r *redisService) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.client.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

func (r *redisService) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.client.ZRevRange(ctx, key, start, stop).Result()
}

func (r *redisService) ZScore(ctx context.Context, key string, member string) (float64, error) {
	return r.client.client.ZScore(ctx, key, member).Result()
}

func (r *redisService) ZRevRank(ctx context.Context, key string, member string) (int64, error) {
	return r.client.client.ZRevRank(ctx, key, member).Result()
}

// List操作实现
func (r *redisService) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.client.LPush(ctx, key, values...).Err()
}

func (r *redisService) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.client.RPush(ctx, key, values...).Err()
}

func (r *redisService) LPop(ctx context.Context, key string) (string, error) {
	return r.client.client.LPop(ctx, key).Result()
}

func (r *redisService) RPop(ctx context.Context, key string) (string, error) {
	return r.client.client.RPop(ctx, key).Result()
}

func (r *redisService) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.client.LLen(ctx, key).Result()
}

// 分布式锁实现
func (r *redisService) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result, err := r.client.client.SetNX(ctx, key, "locked", expiration).Result()
	return result, err
}

func (r *redisService) Unlock(ctx context.Context, key string) error {
	return r.client.client.Del(ctx, key).Err()
}
