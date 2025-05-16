package userService

import (
	"context"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
}
