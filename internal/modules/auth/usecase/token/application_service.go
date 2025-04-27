package tokenService

import (
	"context"
	"time"

	"auth-service/internal/domain/entity"
	userService "auth-service/internal/usecase/user"

	"auth-service/pkg/token"

	"github.com/google/uuid"
)

type tokenUseCase struct {
	userRepo             userService.UserRepository
	jwtManager           *token.JWTManager
	blacklist            *token.TokenBlacklist
	tokenDuration        time.Duration
	refreshTokenDuration time.Duration
}

func NewTokenUseCase(
	userRepo userService.UserRepository,
	jwtManager *token.JWTManager,
	blacklist *token.TokenBlacklist,
	tokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) TokenUseCase {
	return &tokenUseCase{
		userRepo:             userRepo,
		jwtManager:           jwtManager,
		blacklist:            blacklist,
		tokenDuration:        tokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (t *tokenUseCase) GenerateAccessToken(user *entity.User) (string, error) {
	// JWTトークン生成
	claims := &token.Claims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}

	return t.jwtManager.Generate(claims, t.tokenDuration)
}

func (t *tokenUseCase) GenerateRefreshToken(user *entity.User) (string, error) {
	// ランダムなリフレッシュトークン生成
	refreshTokenStr, err := t.jwtManager.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	// DBにリフレッシュトークンを保存
	refreshToken := &entity.RefreshToken{
		ID:        uuid.New(),
		Token:     refreshTokenStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(t.refreshTokenDuration),
		IssuedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.Background()
	if err := t.userRepo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return "", err
	}

	return refreshTokenStr, nil
}

func (t *tokenUseCase) ValidateAccessToken(tokenString string) (*token.Claims, error) {
	// トークンがブラックリストにないか確認
	if t.blacklist.IsTokenBlacklisted(tokenString) {
		return nil, token.ErrTokenBlacklisted
	}

	// トークン検証
	return t.jwtManager.Verify(tokenString)
}

func (t *tokenUseCase) RevokeAccessToken(tokenString string) error {
	// トークンをブラックリストに追加
	claims, err := t.jwtManager.ExtractWithoutValidation(tokenString)
	if err != nil {
		return err
	}

	// 有効期限を計算
	expirationTime := time.Unix(claims.ExpiresAt.Time.Unix(), 0)
	ttl := time.Until(expirationTime)
	if ttl < 0 {
		ttl = 0 // すでに期限切れの場合は最小値を設定
	}

	return t.blacklist.AddToBlacklist(tokenString, ttl)
}

func (t *tokenUseCase) IsTokenRevoked(tokenString string) bool {
	return t.blacklist.IsTokenBlacklisted(tokenString)
}
