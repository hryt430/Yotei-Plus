package gateway

import (
	"context"
)

// AppNotificationGateway はアプリ内通知のゲートウェイインターフェース
type AppNotificationGateway interface {
	// SendNotification はアプリ内通知を送信する
	SendNotification(ctx context.Context, userID, title, message string, metadata map[string]string) error

	// MarkAsRead は通知を既読としてマークする
	MarkAsRead(ctx context.Context, notificationID string) error

	// GetUnreadCount はユーザーの未読通知数を取得する
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

// LineNotificationGateway はLINE通知のゲートウェイインターフェース
type LineNotificationGateway interface {
	// SendNotification は通知を送信する
	SendNotification(ctx context.Context, userID, title, message string, metadata map[string]string) error

	// SendLineNotification はLINE通知を送信する
	SendLineNotification(ctx context.Context, lineUserID, message string) error
}

// WebhookGateway はWebhook送信のためのゲートウェイインターフェース
type WebhookGateway interface {
	// SendWebhook はWebhookを送信する
	SendWebhook(ctx context.Context, event string, payload interface{}) error
}
