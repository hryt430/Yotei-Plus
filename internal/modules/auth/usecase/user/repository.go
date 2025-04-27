package userService

import (
	"context"

	"auth-service/internal/domain/entity"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
}
