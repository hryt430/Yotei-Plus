package token

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklistはJWTの失効管理を担当
type TokenBlacklist struct {
	redisClient *redis.Client
	prefix      string
}

// NewTokenBlacklistは新しいトークンブラックリストを作成
func NewTokenBlacklist(redisClient *redis.Client, prefix string) *TokenBlacklist {
	return &TokenBlacklist{
		redisClient: redisClient,
		prefix:      prefix,
	}
}

// AddToBlacklistはトークンをブラックリストに追加
func (b *TokenBlacklist) AddToBlacklist(token string, ttl time.Duration) error {
	ctx := context.Background()
	key := b.prefix + token
	return b.redisClient.Set(ctx, key, "1", ttl).Err()
}

// IsTokenBlacklistedはトークンがブラックリストに含まれているか確認
func (b *TokenBlacklist) IsTokenBlacklisted(token string) bool {
	ctx := context.Background()
	key := b.prefix + token
	_, err := b.redisClient.Get(ctx, key).Result()
	return err == nil // エラーがなければキーが存在する（ブラックリストに含まれる）
}
