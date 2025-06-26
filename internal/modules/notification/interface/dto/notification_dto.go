package dto

import (
	"time"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
)

// === リクエストDTO ===

type CreateNotificationRequest struct {
	UserID   string            `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type     string            `json:"type" binding:"required" example:"TASK_ASSIGNED"`
	Title    string            `json:"title" binding:"required" example:"新しいタスクが割り当てられました"`
	Message  string            `json:"message" binding:"required" example:"「重要なプロジェクト」タスクが割り当てられました"`
	Metadata map[string]string `json:"metadata,omitempty" example:"{\"task_id\":\"task-123\",\"priority\":\"HIGH\"}"`
	Channels []string          `json:"channels" binding:"required" example:"[\"app\",\"line\"]"`
} // @name CreateNotificationRequest

type SendNotificationRequest struct {
	NotificationID string `json:"notification_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
} // @name SendNotificationRequest

type MarkAsReadRequest struct {
	NotificationID string `json:"notification_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
} // @name MarkAsReadRequest

// === レスポンスDTO ===

type NotificationResponse struct {
	ID        string            `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    string            `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type      string            `json:"type" example:"TASK_ASSIGNED"`
	Title     string            `json:"title" example:"新しいタスクが割り当てられました"`
	Message   string            `json:"message" example:"「重要なプロジェクト」タスクが割り当てられました"`
	Status    string            `json:"status" example:"SENT"`
	Metadata  map[string]string `json:"metadata,omitempty" example:"{\"task_id\":\"task-123\",\"priority\":\"HIGH\"}"`
	CreatedAt time.Time         `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time         `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	SentAt    *time.Time        `json:"sent_at,omitempty" example:"2024-01-01T00:00:00Z"`
} // @name NotificationResponse

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Pagination    PaginationInfo         `json:"pagination"`
} // @name NotificationListResponse

type UnreadCountResponse struct {
	Count int `json:"count" example:"5"`
} // @name UnreadCountResponse

type PaginationInfo struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
} // @name PaginationInfo

// === 共通レスポンス ===

type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"通知が正常に作成されました"`
} // @name SuccessResponse

type ErrorResponse struct {
	Error   string `json:"error" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"リクエストパラメータが不正です"`
} // @name ErrorResponse

// === 変換関数 ===

// ToNotificationResponse はdomain.NotificationをNotificationResponseに変換する
func ToNotificationResponse(notification *domain.Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      string(notification.Type),
		Title:     notification.Title,
		Message:   notification.Message,
		Status:    string(notification.Status),
		Metadata:  notification.Metadata,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
		SentAt:    notification.SentAt,
	}
}

// ToNotificationListResponse は通知一覧をNotificationListResponseに変換する
func ToNotificationListResponse(notifications []*domain.Notification, total, page, pageSize int) *NotificationListResponse {
	notificationResponses := make([]NotificationResponse, len(notifications))
	for i, notification := range notifications {
		notificationResponses[i] = *ToNotificationResponse(notification)
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &NotificationListResponse{
		Notifications: notificationResponses,
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// ToCreateNotificationInput はCreateNotificationRequestをinput.CreateNotificationInputに変換する
func ToCreateNotificationInput(req *CreateNotificationRequest) *input.CreateNotificationInput {
	return &input.CreateNotificationInput{
		UserID:   req.UserID,
		Type:     req.Type,
		Title:    req.Title,
		Message:  req.Message,
		Metadata: req.Metadata,
		Channels: req.Channels,
	}
}

// ToGetNotificationsInput はパラメータからinput.GetNotificationsInputを作成する
func ToGetNotificationsInput(userID string, limit, offset int) *input.GetNotificationsInput {
	return &input.GetNotificationsInput{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}
}