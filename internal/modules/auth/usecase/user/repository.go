package userService

import (
	"github.com/hryt430/Yotei+/internal/modules/auth/domain"

	"github.com/google/uuid"
)

type IUserRepository interface {
	CreateUser(user *domain.User) error
	FindUserByEmail(email string) (*domain.User, error)
	FindUserByID(id uuid.UUID) (*domain.User, error)
	FindUsers(search string) ([]*domain.User, error)
	UpdateUser(user *domain.User) error
}
