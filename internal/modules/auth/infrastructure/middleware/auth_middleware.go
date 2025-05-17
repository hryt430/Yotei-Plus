package middleware

import (
	"net/http"
	"strings"

	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	"github.com/hryt430/Yotei+/pkg/token"
	"github.com/hryt430/Yotei+/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenUseCase tokenService.TokenUseCase
}

func NewAuthMiddleware(tokenUseCase tokenService.TokenUseCase) *AuthMiddleware {
	return &AuthMiddleware{
		tokenUseCase: tokenUseCase,
	}
}

// AuthRequired は認証を必要とするエンドポイント用のミドルウェア
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// トークンの取得（ヘッダーまたはCookie）
		tokenString := m.extractToken(ctx)
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse("Authorization token required"))
			return
		}

		// トークンの検証
		claims, err := m.tokenUseCase.ValidateAccessToken(tokenString)
		if err != nil {
			if err == token.ErrExpiredToken {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse("Token has expired"))
				return
			}
			if err == token.ErrTokenBlacklisted {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse("Token has been revoked"))
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse("Invalid token"))
			return
		}

		// ユーザー情報をコンテキストに設定
		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)
		ctx.Set("username", claims.Username)
		ctx.Set("role", claims.Role)

		ctx.Next()
	}
}

// extractToken はリクエストからトークンを抽出
func (m *AuthMiddleware) extractToken(ctx *gin.Context) string {
	// Authorizationヘッダーからトークンを取得
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// ヘッダーがなければCookieから
	token, err := ctx.Cookie("access_token")
	if err == nil {
		return token
	}

	return ""
}

// RoleRequired は特定のロールを持つユーザーのみアクセス可能にするミドルウェア
func (m *AuthMiddleware) RoleRequired(role string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// すでに認証済みであることを前提
		userRole, exists := ctx.Get("role")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
			return
		}

		// ロールチェック
		if userRole != role {
			ctx.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse("Access denied: insufficient privileges"))
			return
		}

		ctx.Next()
	}
}
