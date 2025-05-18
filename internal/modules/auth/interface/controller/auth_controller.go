package controller

import (
	"net/http"
	"strings"
	"time"

	authService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/auth"
	"github.com/hryt430/Yotei+/pkg/logger"
	"github.com/hryt430/Yotei+/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUseCase authService.AuthUseCase
	logger      logger.Logger
}

func NewAuthController(authUseCase authService.AuthUseCase) *AuthController {
	return &AuthController{
		authUseCase: authUseCase,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	// 入力値のサニタイズ
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	user, err := c.authUseCase.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	// 入力値のサニタイズ
	req.Email = strings.TrimSpace(req.Email)

	accessToken, refreshToken, err := c.authUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse("Invalid credentials"))
		return
	}

	// HTTPOnly cookieにトークンを設定
	ctx.SetCookie(
		"access_token",
		accessToken,
		int(time.Hour.Seconds()), // 1時間
		"/",
		"",
		true, // Secure
		true, // HTTPOnly
	)

	// リフレッシュトークンもCookieとして設定
	ctx.SetCookie(
		"refresh_token",
		refreshToken,
		int((7 * 24 * time.Hour).Seconds()), // 7日間
		"/",
		"",
		true, // Secure
		true, // HTTPOnly
	)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
		},
	})
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Cookieからリフレッシュトークンを取得を試行
		refreshToken, err := ctx.Cookie("refresh_token")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Refresh token is required"))
			return
		}
		req.RefreshToken = refreshToken
	}

	newAccessToken, newRefreshToken, err := c.authUseCase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse("Invalid or expired refresh token"))
		return
	}

	// 新しいアクセストークンとリフレッシュトークンをCookieに設定
	ctx.SetCookie(
		"access_token",
		newAccessToken,
		int(time.Hour.Seconds()), // 1時間
		"/",
		"",
		true, // Secure
		true, // HTTPOnly
	)

	ctx.SetCookie(
		"refresh_token",
		newRefreshToken,
		int((7 * 24 * time.Hour).Seconds()), // 7日間
		"/",
		"",
		true, // Secure
		true, // HTTPOnly
	)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token refreshed successfully",
		"data": gin.H{
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
			"token_type":    "Bearer",
		},
	})
}

func (c *AuthController) Logout(ctx *gin.Context) {
	// アクセストークンをヘッダーから取得
	authHeader := ctx.GetHeader("Authorization")
	accessToken := ""

	// ヘッダーからトークン取得
	if authHeader != "" {
		// Bearer トークンから抽出
		if strings.HasPrefix(authHeader, "Bearer ") {
			accessToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	} else {
		// ヘッダーにない場合はCookieから取得
		var err error
		accessToken, err = ctx.Cookie("access_token")
		if err != nil {
			// アクセストークンが見つからない場合でも処理を続行
			c.logger.Warn("Access token not found in cookie", logger.Error(err))
		}
	}

	// リフレッシュトークンをリクエストボディまたはCookieから取得
	var req LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Cookieからリフレッシュトークンを取得
		refreshToken, err := ctx.Cookie("refresh_token")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Refresh token is required"))
			return
		}
		req.RefreshToken = refreshToken
	}

	// ログアウト処理
	if err := c.authUseCase.Logout(ctx, accessToken, req.RefreshToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to logout"))
		return
	}

	// Cookieを削除
	ctx.SetCookie("access_token", "", -1, "/", "", true, true)
	ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// ユーザー情報取得API (認証済みユーザー用)
func (c *AuthController) Me(ctx *gin.Context) {
	// auth_middlewareで設定されたユーザーIDを取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse("User not authenticated"))
		return
	}

	// ユーザー情報を返す
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":  userID,
			"email":    ctx.GetString("email"),
			"username": ctx.GetString("username"),
			"role":     ctx.GetString("role"),
		},
	})
}
