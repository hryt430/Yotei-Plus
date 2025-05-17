package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT関連のエラー
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrTokenBlacklisted = errors.New("token has been revoked")
)

// ClaimsはJWTのペイロード部分
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenID  string `json:"jti,omitempty"` // JWT ID for blacklisting
}

// JWTManagerはトークンの生成と検証を担当
type JWTManager struct {
	secretKey []byte
	issuer    string
}

// NewJWTManagerは新しいJWTマネージャーを作成
func NewJWTManager(secretKey string, issuer string) *JWTManager {
	return &JWTManager{secretKey: []byte(secretKey), issuer: issuer}
}

// GenerateはJWTトークンを生成
func (m *JWTManager) Generate(claims *Claims, duration time.Duration) (string, error) {
	tokenID := uuid.New().String()

	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    m.issuer,
		ID:        tokenID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshTokenはリフレッシュトークン用のランダム文字列を生成
func (m *JWTManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// VerifyはJWTトークンを検証し、クレームを返す
func (m *JWTManager) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// 署名アルゴリズムの確認
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return m.secretKey, nil
		},
	)

	if err != nil {
		// 期限切れエラーの特定
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ExtractWithoutValidationはトークンを検証せずにクレームを抽出（失効処理用）
func (m *JWTManager) ExtractWithoutValidation(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
