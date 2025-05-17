package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"your-app/notification/domain/entity"
	"your-app/notification/usecase"
)

// NotificationHandler は通知に関するHTTPハンドラー
type NotificationHandler struct {
	createUseCase *usecase.CreateNotificationUseCase
	sendUseCase   *usecase.SendNotificationUseCase
	readUseCase   *usecase.ReadNotificationUseCase
}

// NewNotificationHandler は通知ハンドラーのインスタンスを作成する
func NewNotificationHandler(
	createUseCase *usecase.CreateNotificationUseCase,
	sendUseCase *usecase.SendNotificationUseCase,
	readUseCase *usecase.ReadNotificationUseCase,
) *NotificationHandler {
	return &NotificationHandler{
		createUseCase: createUseCase,
		sendUseCase:   sendUseCase,
		readUseCase:   readUseCase,
	}
}

// CreateNotificationRequest は通知作成リクエスト
type CreateNotificationRequest struct {
	UserID    uint                   `json:"user_id" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	Title     string                 `json:"title" binding:"required"`
	Content   string                 `json:"content" binding:"required"`
	RelatedID *uint                  `json:"related_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Channels  []string               `json:"channels" binding:"required"` // "app", "line" など
}

// Create は通知を作成するハンドラー
func (h *NotificationHandler) Create(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// リクエストをユースケースの入力に変換
	input := usecase.CreateNotificationInput{
		UserID:    req.UserID,
		Type:      entity.NotificationType(req.Type),
		Title:     req.Title,
		Content:   req.Content,
		RelatedID: req.RelatedID,
		Metadata:  req.Metadata,
		Channels:  req.Channels,
	}

	// 通知作成ユースケースの実行
	notification, err := h.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通知送信ユースケースの実行
	if err := h.sendUseCase.Execute(c.Request.Context(), notification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "notification created but failed to send: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetAll はユーザーの通知一覧を取得するハンドラー
func (h *NotificationHandler) GetAll(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// クエリパラメータからlimitとoffsetを取得
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 通知一覧取得
	notifications, err := h.readUseCase.GetNotifications(c.Request.Context(), uint(userID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// GetUnread はユーザーの未読通知一覧を取得するハンドラー
func (h *NotificationHandler) GetUnread(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// 未読通知一覧取得
	notifications, err := h.readUseCase.GetUnreadNotifications(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkAsRead は通知を既読にするハンドラー
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	// 通知既読ユースケースの実行
	if err := h.readUseCase.Execute(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetByID は通知を取得するハンドラー
func (h *NotificationHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	// 通知取得
	notification, err := h.readUseCase.GetNotification(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if notification == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, notification)
}
