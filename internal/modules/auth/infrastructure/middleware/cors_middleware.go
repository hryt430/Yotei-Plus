package middleware

import (
	"net/http"
	"time"

	"github.com/hryt430/Yotei+/pkg/utils"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware はCORSとセキュリティヘッダーを設定するミドルウェア
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 許可するオリジン
		c.Header("Access-Control-Allow-Origin", "https://yourdomain.com")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// XSS対策ヘッダー
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// プリフライトリクエスト処理
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CSRFProtection はCSRF対策を行うミドルウェア
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET以外のリクエストにCSRFトークンを要求
		if c.Request.Method != "GET" && c.Request.Method != "OPTIONS" {
			token := c.GetHeader("X-CSRF-Token")
			cookie, err := c.Cookie("csrf_token")

			if err != nil || token == "" || token != cookie {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "CSRF token verification failed",
				})
				return
			}
		}

		c.Next()
	}
}

// SetCSRFToken はCSRFトークンを設定するミドルウェア
func SetCSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// すでにCSRFトークンがあるか確認
		_, err := c.Cookie("csrf_token")
		if err != nil {
			// ランダムなトークンを生成
			token := utils.GenerateRandomString(32)

			// トークンをCookieに設定
			c.SetCookie(
				"csrf_token",
				token,
				int(24*time.Hour.Seconds()), // 24時間
				"/",
				"",
				true,  // Secure
				false, // HTTPOnly=false to allow JS to read it
			)
		}

		c.Next()
	}
}
