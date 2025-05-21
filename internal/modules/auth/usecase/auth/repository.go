package authService

import (
	"context"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
)

type AuthServiceRepository interface {
	Register(ctx context.Context, email, username, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}
