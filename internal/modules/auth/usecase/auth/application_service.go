package authService

import (
	"errors"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"

	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	userService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/user"

	"github.com/hryt430/Yotei+/pkg/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	AuthRepository IAuthRepository
	UserService    userService.UserService
	TokenService   tokenService.TokenService
}

func NewAuthService(authRepository IAuthRepository, userService userService.UserService, tokenService tokenService.TokenService) *AuthService {
	return &AuthService{AuthRepository: authRepository, UserService: userService, TokenService: tokenService}
}

func (a *AuthService) Register(email, username, password string) (*domain.User, error) {
	// メールアドレスの重複チェック
	existingUser, err := a.UserService.FindUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// パスワードのハッシュ化
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Username:  username,
		Password:  hashedPassword,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newUser, err := a.UserService.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (a *AuthService) Login(email, password string) (string, string, error) {
	user, err := a.UserService.FindUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", errors.New("invalid email or password")
	}

	// パスワード検証
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", "", errors.New("invalid email or password")
	}

	// 最終ログイン時間を更新
	if err := a.UserService.UpdateLastLogin(user.ID); err != nil {
		return "", "", err
	}

	// アクセストークン生成
	accessToken, err := a.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// リフレッシュトークン生成
	refreshTokenString, err := a.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenString, nil
}

func (a *AuthService) RefreshToken(refreshTokenStr string) (string, string, error) {
	// リフレッシュトークンの検証
	refreshTokenEntity, err := a.TokenService.TokenRepository.FindRefreshToken(refreshTokenStr)
	if err != nil {
		return "", "", err
	}

	if refreshTokenEntity == nil || refreshTokenEntity.RevokedAt != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// 有効期限切れ確認
	if time.Now().After(refreshTokenEntity.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	// ユーザー取得
	user, err := a.UserService.FindUserByID(refreshTokenEntity.UserID)
	if err != nil {
		return "", "", err
	}

	// 新しいアクセストークン生成
	newAccessToken, err := a.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// 古いリフレッシュトークンを無効化
	if err := a.TokenService.RevokeToken(refreshTokenStr); err != nil {
		return "", "", err
	}

	// 新しいリフレッシュトークン生成
	newRefreshToken, err := a.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (a *AuthService) Logout(accessToken, refreshToken string) error {
	// アクセストークンをブラックリストに追加
	if err := a.TokenService.RevokeAccessToken(accessToken); err != nil {
		return err
	}

	// リフレッシュトークンを無効化
	if err := a.TokenService.RevokeToken(refreshToken); err != nil {
		return err
	}

	return nil
}
