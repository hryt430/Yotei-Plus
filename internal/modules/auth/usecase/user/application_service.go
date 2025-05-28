package userService

import (
	"errors"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	"github.com/hryt430/Yotei+/pkg/utils"

	"github.com/google/uuid"
)

// userUseCase はユーザー関連のユースケースを実装する構造体
type UserService struct {
	UserRepository IUserRepository
}

// NewUserUseCase は新しいUserUseCaseインスタンスを生成する
func NewUserService(userRepo IUserRepository) *UserService {
	return &UserService{
		UserRepository: userRepo,
	}
}

// CreateUser は新しいユーザーを作成する
func (u *UserService) CreateUser(user *domain.User) (*domain.User, error) {
	// メールアドレスの重複チェック
	existingUser, err := u.UserRepository.FindUserByEmail(user.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// パスワードのハッシュ化
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	user.Password = hashedPassword

	// ユーザーの保存
	if err := u.UserRepository.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail はメールアドレスでユーザーを検索する
func (u *UserService) FindUserByEmail(email string) (*domain.User, error) {
	user, err := u.UserRepository.FindUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID はIDでユーザーを検索する
func (u *UserService) FindUserByID(id uuid.UUID) (*domain.User, error) {
	user, err := u.UserRepository.FindUserByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserProfile はユーザープロフィールを更新する
func (u *UserService) UpdateUserProfile(id uuid.UUID, username string) error {
	user, err := u.UserRepository.FindUserByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.Username = username
	user.UpdatedAt = time.Now()

	return u.UserRepository.UpdateUser(user)
}

// ChangePassword はユーザーのパスワードを変更する
func (u *UserService) ChangePassword(id uuid.UUID, oldPassword, newPassword string) error {
	user, err := u.UserRepository.FindUserByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 古いパスワードの検証
	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		return errors.New("incorrect password")
	}

	// 新しいパスワードのハッシュ化
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	return u.UserRepository.UpdateUser(user)
}

// UpdateLastLogin はユーザーの最終ログイン時間を更新する
func (u *UserService) UpdateLastLogin(id uuid.UUID) error {
	user, err := u.UserRepository.FindUserByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	now := time.Now()
	user.LastLogin = &now

	return u.UserRepository.UpdateUser(user)
}
