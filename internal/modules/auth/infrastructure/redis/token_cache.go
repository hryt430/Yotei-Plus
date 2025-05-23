package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisTokenCache はRedisを使用したトークンキャッシュの実装
type RedisTokenCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisTokenCache(client *redis.Client) *RedisTokenCache {
	return &RedisTokenCache{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisTokenCache) SetWithTTL(key string, value string, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisTokenCache) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *RedisTokenCache) Exists(key string) bool {
	val, err := r.client.Exists(r.ctx, key).Result()
	return err == nil && val > 0
}

func (r *RedisTokenCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}
