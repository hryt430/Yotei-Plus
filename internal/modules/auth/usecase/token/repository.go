package tokenService

import (
	"auth-service/internal/domain/entity"
	"auth-service/pkg/token"
)

type TokenUseCase interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateAccessToken(tokenString string) (*token.Claims, error)
	RevokeAccessToken(tokenString string) error
	IsTokenRevoked(tokenString string) bool
}
