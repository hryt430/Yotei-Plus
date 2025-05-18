package input

import (
	"context"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
)

// CreateNotificationInput は通知作成の入力データ
type CreateNotificationInput struct {
	UserID   string            `json:"user_id" binding:"required"`
	Type     string            `json:"type" binding:"required"`
	Title    string            `json:"title" binding:"required"`
	Message  string            `json:"message" binding:"required"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Channels []string          `json:"channels" binding:"required"` // "app", "line" などのチャネル指定
}

// GetNotificationsInput はユーザー通知一覧取得の入力データ
type GetNotificationsInput struct {
	UserID string `json:"user_id"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// NotificationUseCase は通知のユースケースインターフェース
type NotificationUseCase interface {
	// CreateNotification は新しい通知を作成する
	CreateNotification(ctx context.Context, input CreateNotificationInput) (*domain.Notification, error)

	// GetNotification は通知を取得する
	GetNotification(ctx context.Context, id string) (*domain.Notification, error)

	// GetUserNotifications はユーザーの通知一覧を取得する
	GetUserNotifications(ctx context.Context, input GetNotificationsInput) ([]*domain.Notification, error)

	// SendNotification は通知を送信する
	SendNotification(ctx context.Context, id string) error

	// MarkNotificationAsRead は通知を既読としてマークする
	MarkNotificationAsRead(ctx context.Context, id string) error

	// GetUnreadNotificationCount はユーザーの未読通知数を取得する
	GetUnreadNotificationCount(ctx context.Context, userID string) (int, error)
}
