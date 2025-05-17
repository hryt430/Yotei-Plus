package notification

import (
	"context"

	"your-app/notification/domain/entity"
)

// NotificationUseCase は通知に関するユースケースを定義するインターフェース
type NotificationUseCase interface {
	// CreateNotification は通知を作成する
	CreateNotification(ctx context.Context, input CreateNotificationInput) (*entity.Notification, error)

	// SendNotification は通知を送信する
	SendNotification(ctx context.Context, notification *entity.Notification) error

	// GetNotifications はユーザーの通知一覧を取得する
	GetNotifications(ctx context.Context, userID uint, limit, offset int) ([]*entity.Notification, error)

	// GetUnreadNotifications はユーザーの未読通知一覧を取得する
	GetUnreadNotifications(ctx context.Context, userID uint) ([]*entity.Notification, error)

	// MarkAsRead は通知を既読にする
	MarkAsRead(ctx context.Context, id uint) error

	// GetNotification は通知を取得する
	GetNotification(ctx context.Context, id uint) (*entity.Notification, error)
}
