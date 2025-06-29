package controller

import (
	"net/http"
	"strings"
	"time"

	authService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/auth"
	"github.com/hryt430/Yotei+/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	Interactor authService.AuthService
	logger     logger.Logger
}

func NewAuthController(interactor authService.AuthService, logger logger.Logger) *AuthController {
	return &AuthController{
		Interactor: interactor,
		logger:     logger,
	}
}

// RegisterRequest はユーザー登録のリクエスト構造体
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Username string `json:"username" binding:"required,min=3,max=30" example:"johndoe"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
} // @name RegisterRequest

// LoginRequest はログインのリクエスト構造体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
} // @name LoginRequest

// RefreshTokenRequest はトークン更新のリクエスト構造体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
} // @name RefreshTokenRequest

// LogoutRequest はログアウトのリクエスト構造体
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
} // @name LogoutRequest

// ErrorResponse はエラーレスポンス構造体
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"リクエストが無効です"`
} // @name ErrorResponse

// RegisterResponse はユーザー登録のレスポンス構造体
type RegisterResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"User registered successfully"`
	Data    struct {
		UserID   string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Username string `json:"username" example:"johndoe"`
		Email    string `json:"email" example:"user@example.com"`
	} `json:"data"`
} // @name RegisterResponse

// LoginResponse はログインのレスポンス構造体
type LoginResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Login successful"`
	Data    struct {
		AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
		RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
		TokenType    string `json:"token_type" example:"Bearer"`
	} `json:"data"`
} // @name LoginResponse

// RefreshTokenResponse はトークン更新のレスポンス構造体
type RefreshTokenResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Token refreshed successfully"`
	Data    struct {
		AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
		RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
		TokenType    string `json:"token_type" example:"Bearer"`
	} `json:"data"`
} // @name RefreshTokenResponse

// MeResponse は現在のユーザー情報のレスポンス構造体
type MeResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		UserID   string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Email    string `json:"email" example:"user@example.com"`
		Username string `json:"username" example:"johndoe"`
		Role     string `json:"role" example:"user"`
	} `json:"data"`
} // @name MeResponse

// LogoutResponse はログアウトのレスポンス構造体
type LogoutResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Logged out successfully"`
} // @name LogoutResponse

// Register ユーザー登録
// @Summary      ユーザー登録
// @Description  新しいユーザーアカウントを作成します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "ユーザー登録情報"
// @Success      201 {object} RegisterResponse "ユーザー登録成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      409 {object} ErrorResponse "ユーザーが既に存在"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// 入力値のサニタイズ
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	user, err := c.Interactor.AuthRepository.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
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

// Login ユーザーログイン
// @Summary      ユーザーログイン
// @Description  メールアドレスとパスワードでログインし、アクセストークンとリフレッシュトークンを取得します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "ログイン情報"
// @Success      200 {object} LoginResponse "ログイン成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証情報が無効"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// 入力値のサニタイズ
	req.Email = strings.TrimSpace(req.Email)

	accessToken, refreshToken, err := c.Interactor.AuthRepository.Login(ctx, req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "INVALID_CREDENTIALS",
		Message: "Invalid credentials",
	})
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

// RefreshToken トークン更新
// @Summary      アクセストークン更新
// @Description  リフレッシュトークンを使用して新しいアクセストークンを取得します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest false "リフレッシュトークン（Cookieからも取得可能）"
// @Success      200 {object} RefreshTokenResponse "トークン更新成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "トークンが無効または期限切れ"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /auth/refresh-token [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Cookieからリフレッシュトークンを取得を試行
		refreshToken, err := ctx.Cookie("refresh_token")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "MISSING_REFRESH_TOKEN",
		Message: "Refresh token is required",
	})
			return
		}
		req.RefreshToken = refreshToken
	}

	newAccessToken, newRefreshToken, err := c.Interactor.RefreshToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "INVALID_REFRESH_TOKEN",
		Message: "Invalid or expired refresh token",
	})
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

// Logout ログアウト
// @Summary      ユーザーログアウト
// @Description  現在のセッションを終了し、トークンを無効化します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LogoutRequest false "リフレッシュトークン（Cookieからも取得可能）"
// @Security     BearerAuth
// @Success      200 {object} LogoutResponse "ログアウト成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /auth/logout [post]
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
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "MISSING_REFRESH_TOKEN",
		Message: "Refresh token is required",
	})
			return
		}
		req.RefreshToken = refreshToken
	}

	// ログアウト処理
	if err := c.Interactor.AuthRepository.Logout(ctx, accessToken, req.RefreshToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "LOGOUT_FAILED",
		Message: "Failed to logout",
	})
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

// Me 現在のユーザー情報取得
// @Summary      現在のユーザー情報取得
// @Description  認証済みユーザーの詳細情報を取得します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} MeResponse "ユーザー情報取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /auth/me [get]
func (c *AuthController) Me(ctx *gin.Context) {
	// auth_middlewareで設定されたユーザーIDを取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "UNAUTHORIZED",
		Message: "User not authenticated",
	})
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
