package userService

import (
	"context"
	"errors"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	"github.com/hryt430/Yotei+/pkg/utils"

	"github.com/google/uuid"
)

// userUseCase はユーザー関連のユースケースを実装する構造体
type UserService struct {
	UserServiceRepository IUserRepository
}

// NewUserUseCase は新しいUserUseCaseインスタンスを生成する
func NewUserUseCase(userRepo IUserRepository) *UserService {
	return &UserService{
		UserServiceRepository: userRepo,
	}
}

// CreateUser は新しいユーザーを作成する
func (u *UserService) CreateUser(email, username, password string) (*domain.User, error) {
	// メールアドレスの重複チェック
	existingUser, err := u.UserServiceRepository.FindUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// パスワードのハッシュ化
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 新しいユーザーの作成
	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Username:  username,
		Password:  hashedPassword,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ユーザーの保存
	if err := u.UserServiceRepository.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail はメールアドレスでユーザーを検索する
func (u *UserService) GetUserByEmail(email string) (*domain.User, error) {
	return u.UserServiceRepository.FindUserByEmail(email)
}

// GetUserByID はIDでユーザーを検索する
func (u *UserService) GetUserByID(id uuid.UUID) (*domain.User, error) {
	return u.UserServiceRepository.FindUserByID(id)
}

// UpdateUserProfile はユーザープロフィールを更新する
func (u *UserService) UpdateUserProfile(id uuid.UUID, username string) error {
	user, err := u.UserServiceRepository.FindUserByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.Username = username
	user.UpdatedAt = time.Now()

	return u.UserServiceRepository.UpdateUser(user)
}

// ChangePassword はユーザーのパスワードを変更する
func (u *UserService) ChangePassword(id uuid.UUID, oldPassword, newPassword string) error {
	user, err := u.UserServiceRepository.FindUserByID(id)
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

	return u.UserServiceRepository.UpdateUser(user)
}

// ValidateRefreshToken はリフレッシュトークンを検証する
func (u *UserService) ValidateRefreshToken(token string) (*domain.RefreshToken, error) {
	refreshToken, err := u.UserServiceRepository.FindRefreshToken(token)
	if err != nil {
		return nil, err
	}
	if refreshToken == nil {
		return nil, errors.New("refresh token not found")
	}

	// トークンが取り消されていないか確認
	if refreshToken.RevokedAt != nil {
		return nil, errors.New("refresh token has been revoked")
	}

	// 有効期限の確認
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, errors.New("refresh token has expired")
	}

	return refreshToken, nil
}

// GenerateNewRefreshToken は新しいリフレッシュトークンを生成する
func (u *UserService) GenerateNewRefreshToken(userID uuid.UUID) (*domain.RefreshToken, error) {
	// トークン文字列の生成（実際の実装では安全な方法で）
	tokenString := uuid.New().String()

	// 有効期限の設定（例: 7日間）
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// リフレッシュトークンの作成
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: expiresAt,
		IssuedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// リフレッシュトークンの保存
	if err := u.UserServiceRepository.SaveRefreshToken(refreshToken); err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// RevokeToken はトークンを無効化する
func (u *UserService) RevokeToken(ctx context.Context, token string) error {
	return u.UserServiceRepository.RevokeRefreshToken(token)
}

// CleanupExpiredTokens は期限切れのトークンをクリーンアップする
func (u *UserService) CleanupExpiredTokens(ctx context.Context) error {
	return u.UserServiceRepository.DeleteExpiredRefreshTokens()
}
