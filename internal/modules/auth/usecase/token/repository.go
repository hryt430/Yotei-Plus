package tokenService

import (
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
)

// ITokenRepository はトークンの永続化に関する操作を定義する
type ITokenRepository interface {
	// ブラックリスト関連
	SaveTokenToBlacklist(token string, ttl time.Duration) error
	IsTokenBlacklisted(token string) bool

	// リフレッシュトークン関連
	SaveRefreshToken(token *domain.RefreshToken) error
	FindRefreshToken(token string) (*domain.RefreshToken, error)
	RevokeRefreshToken(token string) error
	DeleteExpiredRefreshTokens() error
}
