package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hryt430/task-management/pkg/logger"

	"github.com/hryt430/task-management/internal/modules/notification/usecase/input"
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

// CreateNotification は新しい通知を作成する
func (c *NotificationController) CreateNotification(ctx *gin.Context) {
	var createInput input.CreateNotificationInput
	if err := ctx.ShouldBindJSON(&createInput); err != nil {
		c.logger.Error("Invalid request body", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	notification, err := c.notificationUseCase.CreateNotification(ctx, createInput)
	if err != nil {
		c.logger.Error("Failed to create notification", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	ctx.JSON(http.StatusCreated, notification)
}

// GetNotification は指定されたIDの通知を取得する
func (c *NotificationController) GetNotification(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	notification, err := c.notificationUseCase.GetNotification(ctx, id)
	if err != nil {
		c.logger.Error("Failed to get notification", "id", id, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification"})
		return
	}

	if notification == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	ctx.JSON(http.StatusOK, notification)
}

// GetUserNotifications はユーザーの通知を取得する
func (c *NotificationController) GetUserNotifications(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// クエリパラメータからlimitとoffsetを取得
	limitStr := ctx.DefaultQuery("limit", "10")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	input := input.GetNotificationsInput{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	notifications, err := c.notificationUseCase.GetUserNotifications(ctx, input)
	if err != nil {
		c.logger.Error("Failed to get user notifications", "userID", userID, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user notifications"})
		return
	}

	ctx.JSON(http.StatusOK, notifications)
}

// SendNotification は通知を即座に送信する
func (c *NotificationController) SendNotification(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	err := c.notificationUseCase.SendNotification(ctx, id)
	if err != nil {
		c.logger.Error("Failed to send notification", "id", id, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Notification sent successfully"})
}

// MarkNotificationAsRead は通知を既読としてマークする
func (c *NotificationController) MarkNotificationAsRead(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	err := c.notificationUseCase.MarkNotificationAsRead(ctx, id)
	if err != nil {
		c.logger.Error("Failed to mark notification as read", "id", id, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// GetUnreadNotificationCount はユーザーの未読通知数を取得する
func (c *NotificationController) GetUnreadNotificationCount(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	count, err := c.notificationUseCase.GetUnreadNotificationCount(ctx, userID)
	if err != nil {
		c.logger.Error("Failed to get unread notification count", "userID", userID, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread notification count"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// WebhookHandler はWebhookリクエストを処理する
func (c *NotificationController) WebhookHandler(ctx *gin.Context) {
	// Webhookの検証処理などを実装
	// ...

	ctx.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}
