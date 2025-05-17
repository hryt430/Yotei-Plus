package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/hryt430/task-management/internal/modules/auth/domain/entity"
	authService "github.com/hryt430/task-management/internal/modules/auth/usecase/auth"
)

// JWTAuthMiddleware は認証用のミドルウェアです
type JWTAuthMiddleware struct {
	authUsecase authService.AuthUseCase
}

// NewJWTAuthMiddleware は新しいJWT認証ミドルウェアを作成します
func NewJWTAuthMiddleware(authUsecase authService.AuthUseCase) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		authUsecase: authUsecase,
	}
}

// Middleware はHTTPハンドラーを認証ミドルウェアでラップします
func (m *JWTAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorizationヘッダーからトークンを取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "認証トークンがありません", http.StatusUnauthorized)
			return
		}

		// Bearer tokenの形式を想定
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "認証トークンのフォーマットが不正です", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]

		// トークンの検証
		user, err := m.authUsecase.ValidateToken(r.Context(), token)
		if err != nil {
			log.Printf("トークン検証エラー: %v", err)
			http.Error(w, "無効なトークンです", http.StatusUnauthorized)
			return
		}

		// コンテキストにユーザー情報を追加
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole は特定のロールを持つユーザーのみアクセスを許可します
func (m *JWTAuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// コンテキストからユーザー情報を取得
			user, ok := r.Context().Value("user").(*entity.User)
			if !ok {
				http.Error(w, "認証されていません", http.StatusUnauthorized)
				return
			}

			// ユーザーが必要なロールを持っているか確認
			hasRole := false
			for _, userRole := range user.Roles {
				if userRole.Name == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "アクセス権限がありません", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext はコンテキストからユーザー情報を取得するヘルパー関数です
func GetUserFromContext(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value("user").(*entity.User)
	if !ok {
		return nil, errors.New("コンテキストからユーザー情報を取得できません")
	}
	return user, nil
}
