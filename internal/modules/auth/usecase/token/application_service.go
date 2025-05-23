package tokenService

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	"github.com/hryt430/Yotei+/pkg/token"
)

type TokenService struct {
	TokenRepository      ITokenRepository
	jwtManager           *token.JWTManager
	tokenDuration        time.Duration
	refreshTokenDuration time.Duration
}

func NewTokenService(
	tokenRepository ITokenRepository,
	jwtManager *token.JWTManager,
	tokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) *TokenService {
	return &TokenService{
		TokenRepository:      tokenRepository,
		jwtManager:           jwtManager,
		tokenDuration:        tokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (t *TokenService) RevokeAccessToken(tokenString string) error {
	// トークンをブラックリストに追加
	claims, err := t.jwtManager.ExtractWithoutValidation(tokenString)
	if err != nil {
		return err
	}

	// 有効期限を計算
	expirationTime := time.Unix(claims.ExpiresAt.Time.Unix(), 0)
	ttl := time.Until(expirationTime)
	if ttl < 0 {
		ttl = 0
	}

	// リポジトリを使用してブラックリストに保存
	return t.TokenRepository.SaveTokenToBlacklist(tokenString, ttl)
}

func (t *TokenService) GenerateAccessToken(user *domain.User) (string, error) {
	// JWTトークン生成
	claims := &token.Claims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}

	return t.jwtManager.Generate(claims, t.tokenDuration)
}

func (t *TokenService) GenerateRefreshToken(user *domain.User) (string, error) {
	// ランダムなリフレッシュトークン生成
	refreshTokenStr, err := t.jwtManager.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	// DBにリフレッシュトークンを保存
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		Token:     refreshTokenStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(t.refreshTokenDuration),
		IssuedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := t.TokenRepository.SaveRefreshToken(refreshToken); err != nil {
		return "", err
	}

	return refreshTokenStr, nil
}

func (t *TokenService) ValidateAccessToken(tokenString string) (*token.Claims, error) {
	// トークンがブラックリストにないか確認
	if t.TokenRepository.IsTokenBlacklisted(tokenString) {
		return nil, token.ErrTokenBlacklisted
	}

	// トークン検証
	return t.jwtManager.Verify(tokenString)
}

// ValidateRefreshToken はリフレッシュトークンを検証する
func (u *TokenService) ValidateRefreshToken(token string) (*domain.RefreshToken, error) {
	refreshToken, err := u.TokenRepository.FindRefreshToken(token)
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
func (u *TokenService) GenerateNewRefreshToken(userID uuid.UUID) (*domain.RefreshToken, error) {
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
	if err := u.TokenRepository.SaveRefreshToken(refreshToken); err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// RevokeToken はトークンを無効化する
func (u *TokenService) RevokeToken(token string) error {
	return u.TokenRepository.RevokeRefreshToken(token)
}

// CleanupExpiredTokens は期限切れのトークンをクリーンアップする
func (u *TokenService) CleanupExpiredTokens() error {
	return u.TokenRepository.DeleteExpiredRefreshTokens()
}
func (t *TokenService) IsTokenRevoked(tokenString string) bool {
	return t.TokenRepository.IsTokenBlacklisted(tokenString)
}
