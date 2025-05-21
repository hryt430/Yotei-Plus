package tokenService

import (
	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	"github.com/hryt430/Yotei+/pkg/token"
)

type ITokenRepository interface {
	GenerateAccessToken(user *domain.User) (string, error)
	GenerateRefreshToken(user *domain.User) (string, error)
	ValidateAccessToken(tokenString string) (*token.Claims, error)
	RevokeAccessToken(tokenString string) error
	IsTokenRevoked(tokenString string) bool
}
