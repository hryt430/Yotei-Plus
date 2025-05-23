package redis

import (
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	"github.com/hryt430/Yotei+/internal/modules/auth/infrastructure/redis"
	"github.com/hryt430/Yotei+/internal/modules/auth/interface/database"
	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
)

// TokenRepositoryAdapter はinfrastructure層の実装をusecase層のインターフェースに適合させる
type TokenRepositoryAdapter struct {
	tokenCache   *redis.RedisTokenCache
	tokenStorage *database.TokenStorage
}

func NewTokenRepositoryAdapter(
	tokenCache *redis.RedisTokenCache,
	tokenStorage *database.TokenStorage,
) tokenService.ITokenRepository {
	return &TokenRepositoryAdapter{
		tokenCache:   tokenCache,
		tokenStorage: tokenStorage,
	}
}

// ブラックリスト関連（Redis使用）
func (r *TokenRepositoryAdapter) SaveTokenToBlacklist(token string, ttl time.Duration) error {
	key := "blacklist:" + token
	return r.tokenCache.SetWithTTL(key, "1", ttl)
}

func (r *TokenRepositoryAdapter) IsTokenBlacklisted(token string) bool {
	key := "blacklist:" + token
	return r.tokenCache.Exists(key)
}

// リフレッシュトークン関連（DB使用）
func (r *TokenRepositoryAdapter) SaveRefreshToken(token *domain.RefreshToken) error {
	return r.tokenStorage.SaveRefreshToken(token)
}

func (r *TokenRepositoryAdapter) FindRefreshToken(token string) (*domain.RefreshToken, error) {
	return r.tokenStorage.FindRefreshTokenByToken(token)
}

func (r *TokenRepositoryAdapter) RevokeRefreshToken(token string) error {
	return r.tokenStorage.RevokeRefreshToken(token)
}

func (r *TokenRepositoryAdapter) DeleteExpiredRefreshTokens() error {
	// 実装は要件に応じて
	return r.tokenStorage.DeleteExpiredRefreshTokens()
}
