package authService

import (
	"auth-service/internal/domain/entity"
	"context"
	"errors"
	"time"

	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	userService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/user"

	"github.com/hryt430/Yotei+/pkg/utils"

	"github.com/google/uuid"
)

type authUseCase struct {
	userRepo             userService.UserRepository
	tokenUseCase         tokenService.TokenUseCase
	tokenDuration        time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthUseCase(
	userRepo userService.UserRepository,
	tokenUseCase tokenService.TokenUseCase,
	tokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) AuthUseCase {
	return &authUseCase{
		userRepo:             userRepo,
		tokenUseCase:         tokenUseCase,
		tokenDuration:        tokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (a *authUseCase) Register(ctx context.Context, email, username, password string) (*entity.User, error) {
	// メールアドレスの重複チェック
	existingUser, err := a.userRepo.FindUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// パスワードのハッシュ化
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:        uuid.New(),
		Email:     email,
		Username:  username,
		Password:  hashedPassword,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := a.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (a *authUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := a.userRepo.FindUserByEmail(ctx, email)
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
	user.LastLogin = time.Now()
	if err := a.userRepo.UpdateUser(ctx, user); err != nil {
		return "", "", err
	}

	// アクセストークン生成
	accessToken, err := a.tokenUseCase.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// リフレッシュトークン生成
	refreshTokenString, err := a.tokenUseCase.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenString, nil
}

func (a *authUseCase) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	// リフレッシュトークンの検証
	refreshTokenEntity, err := a.userRepo.FindRefreshToken(ctx, refreshTokenStr)
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
	user, err := a.userRepo.FindUserByID(ctx, refreshTokenEntity.UserID)
	if err != nil {
		return "", "", err
	}

	// 新しいアクセストークン生成
	newAccessToken, err := a.tokenUseCase.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// 古いリフレッシュトークンを無効化
	if err := a.userRepo.RevokeRefreshToken(ctx, refreshTokenStr); err != nil {
		return "", "", err
	}

	// 新しいリフレッシュトークン生成
	newRefreshToken, err := a.tokenUseCase.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (a *authUseCase) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// アクセストークンをブラックリストに追加
	if err := a.tokenUseCase.RevokeAccessToken(accessToken); err != nil {
		return err
	}

	// リフレッシュトークンを無効化
	if err := a.userRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return err
	}

	return nil
}
