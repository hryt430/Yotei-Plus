package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	authService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/auth"
)

// NewJWTAuthMiddleware Gin 用 JWT 認証ミドルウェアを返します
func NewJWTAuthMiddleware(authUC *authService.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization ヘッダー取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証トークンがありません"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証トークンのフォーマットが不正です"})
			return
		}
		token := parts[1]

		// トークン検証
		user, err := authUC.TokenService.ValidateAccessToken(token)
		if err != nil {
			log.Printf("トークン検証エラー: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンです"})
			return
		}

		// Gin コンテキストにユーザー情報をセット
		c.Set("user", user)

		c.Next()
	}
}

// RequireRole は指定のロールを持つユーザーのみ許可するミドルウェア生成器です
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証されていません"})
			return
		}

		user, ok := val.(*domain.User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報の取得に失敗"})
			return
		}

		if user.Role == role {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "アクセス権限がありません"})
	}
}

// GetUserFromContext は Gin のコンテキストからドメインユーザーを取得します
func GetUserFromContext(c *gin.Context) (*domain.User, error) {
	val, exists := c.Get("user")
	if !exists {
		return nil, errors.New("コンテキストにユーザー情報がありません")
	}
	user, ok := val.(*domain.User)
	if !ok {
		return nil, errors.New("ユーザー情報の型が不正です")
	}
	return user, nil
}
