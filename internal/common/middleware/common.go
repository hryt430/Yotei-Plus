package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// LoggerMiddleware はリクエストのロギングを行うミドルウェアです
func LoggerMiddleware(log logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.Info("HTTP Request",
			logger.Any("method", param.Method),
			logger.Any("path", param.Path),
			logger.Any("status", param.StatusCode),
			logger.Any("latency", param.Latency),
			logger.Any("client_ip", param.ClientIP),
			logger.Any("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

// RecoveryMiddleware はパニックからの回復を処理するミドルウェアです
func RecoveryMiddleware(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Error("Panic recovered", logger.Any("error", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// CORSMiddleware はCross-Origin Resource Sharingを処理するミドルウェアです
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 設定からCORSの許可されたオリジンを取得
		allowedOrigins := cfg.GetAllowedOrigins()

		// 開発環境では全てのオリジンを許可
		if cfg.IsDevelopment() || isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// プリフライトリクエスト（OPTIONS）の処理
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware はレート制限を行うミドルウェアです
func RateLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// レート制限の実装（Redisを使用する場合はここで実装）
		// 現在は簡単な通過のみ
		c.Next()
	}
}

// SecurityHeadersMiddleware はセキュリティヘッダーを設定するミドルウェアです
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 本番環境ではHTTPS必須
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// CSRFProtection はCSRF攻撃を防ぐミドルウェアです
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET, HEAD, OPTIONS は CSRF チェックをスキップ
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// CSRFトークンの取得（ヘッダーまたはフォームから）
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.PostForm("_csrf")
		}

		// セッションからCSRFトークンを取得（実際の実装ではセッションストアを使用）
		sessionToken, exists := c.Get("csrf_token")
		if !exists || token == "" || token != sessionToken.(string) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
			})
			return
		}

		c.Next()
	}
}

// SetCSRFToken はCSRFトークンを生成・設定するミドルウェアです
func SetCSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 既にトークンが設定されている場合はスキップ
		if _, exists := c.Get("csrf_token"); exists {
			c.Next()
			return
		}

		// CSRFトークンの生成
		token, err := generateCSRFToken()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate CSRF token",
			})
			return
		}

		// コンテキストにトークンを設定
		c.Set("csrf_token", token)

		// レスポンスヘッダーにもトークンを設定
		c.Header("X-CSRF-Token", token)

		c.Next()
	}
}

// RequestIDMiddleware はリクエストIDを生成・設定するミドルウェアです
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// TimeoutMiddleware はリクエストタイムアウトを設定するミドルウェアです
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// タイムアウト設定（実際の実装ではcontext.WithTimeoutを使用）
		c.Next()
	}
}

// isOriginAllowed は指定されたオリジンが許可されているかチェックします
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed || allowed == "*" {
			return true
		}
	}
	return false
}

// generateCSRFToken はCSRFトークンを生成します
func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// generateRequestID はリクエストIDを生成します
func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
