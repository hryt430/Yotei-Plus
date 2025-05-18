package output

import (
	"context"
)

// NotificationGateway は通知送信のためのゲートウェイインターフェース
type NotificationGateway interface {
	// SendNotification はユーザーに通知を送信する
	SendNotification(ctx context.Context, userID, title, message string, metadata map[string]string) error
}

// AppNotificationGateway はアプリ内通知のゲートウェイインターフェース
type AppNotificationGateway interface {
	NotificationGateway
	// MarkAsRead は通知を既読としてマークする
	MarkAsRead(ctx context.Context, notificationID string) error
	// GetUnreadCount はユーザーの未読通知数を取得する
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

// LineNotificationGateway はLINE通知のゲートウェイインターフェース
type LineNotificationGateway interface {
	NotificationGateway
	// SendLineNotification はLINE通知を送信する
	SendLineNotification(ctx context.Context, lineUserID, message string) error
}

// WebhookEvent はWebhookイベントの種類
type WebhookEvent string

const (
	EventNotificationSent   WebhookEvent = "notification.sent"
	EventNotificationRead   WebhookEvent = "notification.read"
	EventNotificationFailed WebhookEvent = "notification.failed"
)

// WebhookOutput はWebhook送信のためのインターフェース
type WebhookOutput interface {
	// SendWebhook はWebhookを送信する
	SendWebhook(ctx context.Context, event WebhookEvent, payload interface{}) error
}
