package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	userService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/user"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type UserController struct {
	UserService userService.UserService
	logger      logger.Logger
}

func NewUserController(userService userService.UserService, logger logger.Logger) *UserController {
	return &UserController{
		UserService: userService,
		logger:      logger,
	}
}

// UpdateUserRequest はユーザー更新のリクエスト構造体
type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=30"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// UserResponse はAPIレスポンス用のユーザー情報
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// DetailedUserResponse は詳細なユーザー情報（本人または管理者用）
type DetailedUserResponse struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	LastLogin     string `json:"last_login"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}


// GetUsers はユーザー一覧を取得する（タスク割り当て用）
func (c *UserController) GetUsers(ctx *gin.Context) {
	// 検索クエリの取得
	search := strings.TrimSpace(ctx.Query("search"))

	users, err := c.UserService.GetUsers(ctx, search)
	if err != nil {
		c.logger.Error("Failed to get users", logger.Error(err))
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get users",
	})
		return
	}

	// 基本情報のみを返す（セキュリティ考慮）
	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userResponses,
	})
}

// GetUser は特定のユーザー情報を取得する
func (c *UserController) GetUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "User ID is required",
	})
		return
	}

	// UUIDの検証
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid user ID format",
	})
		return
	}

	// ユーザー取得
	user, err := c.UserService.FindUserByID(parsedID)
	if err != nil {
		c.logger.Error("Failed to get user", logger.Any("userID", userID), logger.Error(err))
		ctx.JSON(http.StatusNotFound, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "User not found",
	})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "User not found",
	})
		return
	}

	// 現在のユーザー情報を取得
	currentUserID := ctx.GetString("user_id")
	currentUserRole := ctx.GetString("role")

	// 権限チェック：自分の情報または管理者は詳細情報を取得
	if userID == currentUserID || currentUserRole == "admin" {
		// 詳細情報を返す
		detailedResponse := DetailedUserResponse{
			ID:            user.ID.String(),
			Username:      user.Username,
			Email:         user.Email,
			Role:          user.Role,
			EmailVerified: user.EmailVerified,
			LastLogin:     user.LastLogin.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    detailedResponse,
		})
	} else {
		// 他人の情報は基本情報のみ
		basicResponse := UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    basicResponse,
		})
	}
}

// UpdateUser はユーザー情報を更新する
func (c *UserController) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "User ID is required",
	})
		return
	}

	// UUIDの検証
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid user ID format",
	})
		return
	}

	// 現在のユーザー情報を取得
	currentUserID := ctx.GetString("user_id")
	currentUserRole := ctx.GetString("role")

	// 権限チェック：自分の情報のみ更新可能（管理者は例外）
	if userID != currentUserID && currentUserRole != "admin" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Access denied: You can only update your own profile",
	})
		return
	}

	// リクエストボディの解析
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// 少なくとも1つのフィールドが更新対象である必要がある
	if req.Username == "" && req.Email == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "At least one field (username or email) must be provided",
	})
		return
	}

	// 入力値のサニタイズ
	if req.Username != "" {
		req.Username = strings.TrimSpace(req.Username)
	}
	if req.Email != "" {
		req.Email = strings.TrimSpace(req.Email)
	}

	// ユーザー更新
	updatedUser, err := c.UserService.UpdateUserProfile(parsedID, req.Username, req.Email)
	if err != nil {
		c.logger.Error("Failed to update user", logger.Any("userID", userID), logger.Error(err))
		if strings.Contains(err.Error(), "email already exists") {
			ctx.JSON(http.StatusConflict, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Email already exists",
	})
			return
		}
		if strings.Contains(err.Error(), "username already exists") {
			ctx.JSON(http.StatusConflict, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Username already exists",
	})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to update user",
	})
		return
	}

	// 更新されたユーザー情報を返す
	response := DetailedUserResponse{
		ID:            updatedUser.ID.String(),
		Username:      updatedUser.Username,
		Email:         updatedUser.Email,
		Role:          updatedUser.Role,
		EmailVerified: updatedUser.EmailVerified,
		LastLogin:     updatedUser.LastLogin.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:     updatedUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User updated successfully",
		"data":    response,
	})
}

// GetCurrentUser は現在のユーザー情報を取得する（互換性維持）
func (c *UserController) GetCurrentUser(ctx *gin.Context) {
	// auth_middlewareで設定されたユーザーIDを取得
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "User not authenticated",
	})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid user ID",
	})
		return
	}

	user, err := c.UserService.FindUserByID(userID)
	if err != nil || user == nil {
		c.logger.Error("Failed to get current user", logger.Any("userID", userID), logger.Error(err))
		ctx.JSON(http.StatusNotFound, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "User not found",
	})
		return
	}

	// 詳細情報を返す
	response := DetailedUserResponse{
		ID:            user.ID.String(),
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		LastLogin:     user.LastLogin.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateCurrentUser は現在のユーザー情報を更新する（互換性維持）
func (c *UserController) UpdateCurrentUser(ctx *gin.Context) {
	// auth_middlewareで設定されたユーザーIDを取得
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "User not authenticated",
	})
		return
	}

	// パラメータを設定して既存のUpdateUserメソッドを呼び出し
	ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: userIDStr.(string)})
	c.UpdateUser(ctx)
}
