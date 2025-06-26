package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/internal/modules/notification/interface/dto"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	"github.com/hryt430/Yotei+/pkg/logger"
	"go.uber.org/zap/zapcore"
)

// NotificationController は通知コントローラー
type NotificationController struct {
	notificationUseCase input.NotificationUseCase
	logger              logger.Logger
}

// NewNotificationController は新しいNotificationControllerを作成する
func NewNotificationController(useCase input.NotificationUseCase, logger logger.Logger) *NotificationController {
	return &NotificationController{
		notificationUseCase: useCase,
		logger:              logger,
	}
}

// CreateNotificationRequest は通知作成のリクエスト構造体
type CreateNotificationRequest struct {
	UserID   string            `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type     string            `json:"type" binding:"required" example:"TASK_ASSIGNED"`
	Title    string            `json:"title" binding:"required" example:"新しいタスクが割り当てられました"`
	Message  string            `json:"message" binding:"required" example:"「重要なプロジェクト」タスクが割り当てられました"`
	Metadata map[string]string `json:"metadata,omitempty" example:"{\"task_id\":\"task-123\",\"priority\":\"HIGH\"}"`
	Channels []string          `json:"channels" binding:"required" example:"[\"app\",\"line\"]"`
} // @name CreateNotificationRequest

// NotificationResponse は通知のレスポンス構造体
type NotificationResponse struct {
	ID        string            `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    string            `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type      string            `json:"type" example:"TASK_ASSIGNED"`
	Title     string            `json:"title" example:"新しいタスクが割り当てられました"`
	Message   string            `json:"message" example:"「重要なプロジェクト」タスクが割り当てられました"`
	Status    string            `json:"status" example:"SENT"`
	Metadata  map[string]string `json:"metadata,omitempty" example:"{\"task_id\":\"task-123\",\"priority\":\"HIGH\"}"`
	CreatedAt string            `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string            `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	SentAt    *string           `json:"sent_at,omitempty" example:"2024-01-01T00:05:00Z"`
} // @name NotificationResponse

// CreateNotificationResponse は通知作成のレスポンス構造体
type CreateNotificationResponse struct {
	Success bool                 `json:"success" example:"true"`
	Data    NotificationResponse `json:"data"`
} // @name CreateNotificationResponse

// GetNotificationResponse は通知取得のレスポンス構造体
type GetNotificationResponse struct {
	Success bool                 `json:"success" example:"true"`
	Data    NotificationResponse `json:"data"`
} // @name GetNotificationResponse

// GetUserNotificationsResponse はユーザー通知一覧のレスポンス構造体
type GetUserNotificationsResponse struct {
	Success bool                   `json:"success" example:"true"`
	Data    []NotificationResponse `json:"data"`
} // @name GetUserNotificationsResponse

// UnreadCountResponse は未読通知数のレスポンス構造体
type UnreadCountResponse struct {
	Success bool `json:"success" example:"true"`
	Count   int  `json:"count" example:"5"`
} // @name UnreadCountResponse

// MessageResponse は基本メッセージレスポンス構造体
type MessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"操作が正常に完了しました"`
} // @name MessageResponse

// ErrorResponse はエラーレスポンス構造体
type ErrorResponse struct {
	Error   string `json:"error" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"リクエストが無効です"`
} // @name ErrorResponse

// CreateNotification 通知作成
// @Summary      通知作成
// @Description  新しい通知を作成します
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request body CreateNotificationRequest true "通知作成情報"
// @Security     BearerAuth
// @Success      201 {object} CreateNotificationResponse "通知作成成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications [post]
func (c *NotificationController) CreateNotification(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		c.logError("get user from context", err)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var createInput input.CreateNotificationInput
	if err := ctx.ShouldBindJSON(&createInput); err != nil {
		c.logError("bind JSON", err)
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	notification, err := c.notificationUseCase.CreateNotification(ctx, createInput)
	if err != nil {
		c.logError("create notification", err, logger.Any("userID", user.ID))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "create_notification_failed",
			Message: "通知の作成に失敗しました",
		})
		return
	}

	c.logger.Info("Notification created successfully",
		logger.Any("userID", user.ID),
		logger.Any("notificationID", notification.ID))

	ctx.JSON(http.StatusCreated, notification)
}

// GetNotification 通知取得
// @Summary      通知取得
// @Description  指定されたIDの通知を取得します
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id path string true "通知ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} GetNotificationResponse "通知取得成功"
// @Failure      400 {object} ErrorResponse "通知IDが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "通知が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications/{id} [get]
func (c *NotificationController) GetNotification(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		c.logError("get user from context", err)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	notificationID, err := c.validateUUID(ctx.Param("id"), "notification ID")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_notification_id",
			Message: "無効な通知IDです",
		})
		return
	}

	notification, err := c.notificationUseCase.GetNotification(ctx, notificationID.String())
	if err != nil {
		c.logError("get notification", err,
			logger.Any("userID", user.ID),
			logger.Any("notificationID", notificationID))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_notification_failed",
			Message: "通知の取得に失敗しました",
		})
		return
	}

	if notification == nil {
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "notification_not_found",
			Message: "通知が見つかりません",
		})
		return
	}

	ctx.JSON(http.StatusOK, notification)
}

// GetUserNotifications ユーザーの通知一覧取得
// @Summary      ユーザーの通知一覧取得
// @Description  指定されたユーザーの通知一覧を取得します（ページング対応）
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        user_id path string true "ユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        limit query int false "取得数の上限" default(10) minimum(1) maximum(100)
// @Param        offset query int false "取得開始位置" default(0) minimum(0)
// @Security     BearerAuth
// @Success      200 {object} GetUserNotificationsResponse "通知一覧取得成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications/user/{user_id} [get]
func (c *NotificationController) GetUserNotifications(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		c.logError("get user from context", err)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetUserID, err := c.validateUUID(ctx.Param("user_id"), "user ID")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	// 権限チェック（自分の通知のみ閲覧可能）
	if user.ID != targetUserID {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "access_denied",
			Message: "他のユーザーの通知を閲覧する権限がありません",
		})
		return
	}

	// クエリパラメータからlimitとoffsetを取得
	limitStr := ctx.DefaultQuery("limit", "10")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	inputData := input.GetNotificationsInput{
		UserID: targetUserID.String(),
		Limit:  limit,
		Offset: offset,
	}

	notifications, err := c.notificationUseCase.GetUserNotifications(ctx, inputData)
	if err != nil {
		c.logError("get user notifications", err,
			logger.Any("userID", user.ID),
			logger.Any("targetUserID", targetUserID))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_user_notifications_failed",
			Message: "ユーザー通知一覧の取得に失敗しました",
		})
		return
	}

	ctx.JSON(http.StatusOK, notifications)
}

// SendNotification 通知送信
// @Summary      通知送信
// @Description  指定された通知を即座に送信します
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id path string true "通知ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} MessageResponse "通知送信成功"
// @Failure      400 {object} ErrorResponse "通知IDが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "通知が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications/{id}/send [post]
func (c *NotificationController) SendNotification(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		c.logError("get user from context", err)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	notificationID, err := c.validateUUID(ctx.Param("id"), "notification ID")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_notification_id",
			Message: "無効な通知IDです",
		})
		return
	}

	err = c.notificationUseCase.SendNotification(ctx, notificationID.String())
	if err != nil {
		c.logError("send notification", err,
			logger.Any("userID", user.ID),
			logger.Any("notificationID", notificationID))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "send_notification_failed",
			Message: "通知の送信に失敗しました",
		})
		return
	}

	c.logger.Info("Notification sent successfully",
		logger.Any("userID", user.ID),
		logger.Any("notificationID", notificationID))

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "通知を送信しました",
	})
}

// MarkNotificationAsRead 通知既読マーク
// @Summary      通知既読マーク
// @Description  指定された通知を既読としてマークします
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id path string true "通知ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} MessageResponse "既読マーク成功"
// @Failure      400 {object} ErrorResponse "通知IDが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "通知が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications/{id}/read [put]
func (c *NotificationController) MarkNotificationAsRead(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		c.logError("get user from context", err)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	notificationID, err := c.validateUUID(ctx.Param("id"), "notification ID")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_notification_id",
			Message: "無効な通知IDです",
		})
		return
	}

	err = c.notificationUseCase.MarkNotificationAsRead(ctx, notificationID.String())
	if err != nil {
		c.logError("mark notification as read", err,
			logger.Any("userID", user.ID),
			logger.Any("notificationID", notificationID))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "mark_as_read_failed",
			Message: "通知の既読マークに失敗しました",
		})
		return
	}

	c.logger.Info("Notification marked as read successfully",
		logger.Any("userID", user.ID),
		logger.Any("notificationID", notificationID))

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "通知を既読にしました",
	})
}

// GetUnreadNotificationCount 未読通知数取得
// @Summary      未読通知数取得
// @Description  指定されたユーザーの未読通知数を取得します
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        user_id path string true "ユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} UnreadCountResponse "未読通知数取得成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications/user/{user_id}/unread/count [get]
func (c *NotificationController) GetUnreadNotificationCount(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		c.logError("get user from context", err)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetUserID, err := c.validateUUID(ctx.Param("user_id"), "user ID")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	// 権限チェック（自分の通知数のみ取得可能）
	if user.ID != targetUserID {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "access_denied",
			Message: "他のユーザーの通知数を取得する権限がありません",
		})
		return
	}

	count, err := c.notificationUseCase.GetUnreadNotificationCount(ctx, targetUserID.String())
	if err != nil {
		c.logError("get unread notification count", err,
			logger.Any("userID", user.ID),
			logger.Any("targetUserID", targetUserID))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_unread_count_failed",
			Message: "未読通知数の取得に失敗しました",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   count,
	})
}

// MarkAllNotificationsAsRead 全通知既読マーク
// @Summary      全通知既読マーク
// @Description  指定されたユーザーの全ての未読通知を既読としてマークします
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        user_id path string true "ユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} MessageResponse "全既読マーク成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "権限がない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications/user/{user_id}/read-all [put]
func (c *NotificationController) MarkAllNotificationsAsRead(ctx *gin.Context) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		c.logError("get user from context", err)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetUserID, err := c.validateUUID(ctx.Param("user_id"), "user ID")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	// 権限チェック（自分の通知のみ操作可能）
	if user.ID != targetUserID {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "access_denied",
			Message: "他のユーザーの通知を操作する権限がありません",
		})
		return
	}

	err = c.notificationUseCase.MarkNotificationAsRead(ctx, targetUserID.String())
	if err != nil {
		c.logError("mark all notifications as read", err,
			logger.Any("userID", user.ID),
			logger.Any("targetUserID", targetUserID))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "mark_all_as_read_failed",
			Message: "全通知の既読マークに失敗しました",
		})
		return
	}

	c.logger.Info("All notifications marked as read successfully",
		logger.Any("userID", user.ID),
		logger.Any("targetUserID", targetUserID))

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "全ての通知を既読にしました",
	})
}

// WebhookHandler Webhook処理
// @Summary      Webhook処理
// @Description  外部サービスからのWebhookリクエストを処理します
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        payload body object true "Webhookペイロード"
// @Success      200 {object} MessageResponse "Webhook受信成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /notifications/webhook [post]
func (c *NotificationController) WebhookHandler(ctx *gin.Context) {
	var payload map[string]interface{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		c.logError("bind webhook payload", err)
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_payload",
			Message: "Webhookペイロードが無効です",
		})
		return
	}

	// Webhookの検証処理やビジネスロジックを実装
	c.logger.Info("Webhook received", logger.Any("payload", payload))

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Webhookを受信しました",
	})
}

// === ヘルパーメソッド ===

func (c *NotificationController) validateUUID(id string, fieldName string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.logger.Error("Invalid UUID format",
			logger.String("field", fieldName),
			logger.String("value", id),
			logger.Error(err))
		return uuid.Nil, err
	}
	return parsedID, nil
}

func (c *NotificationController) logError(operation string, err error, fields ...zapcore.Field) {
	c.logger.Error("Operation failed",
		append([]zapcore.Field{
			logger.String("operation", operation),
			logger.Error(err),
		}, fields...)...)
}

// RegisterNotificationRoutes は通知コントローラーのルートを登録する
func RegisterNotificationRoutes(router *gin.RouterGroup, controller *NotificationController) {
	notifications := router.Group("/notifications")
	{
		notifications.POST("", controller.CreateNotification)
		notifications.GET("/:id", controller.GetNotification)
		notifications.GET("/user/:user_id", controller.GetUserNotifications)
		notifications.POST("/:id/send", controller.SendNotification)
		notifications.PUT("/:id/read", controller.MarkNotificationAsRead)
		notifications.GET("/user/:user_id/unread/count", controller.GetUnreadNotificationCount)
		notifications.PUT("/user/:user_id/read-all", controller.MarkAllNotificationsAsRead)
		notifications.POST("/webhook", controller.WebhookHandler)
	}
}
