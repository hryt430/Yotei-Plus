package domain

import (
	"context"
)

// NotificationService は通知処理のロジックを定義するインターフェース
type NotificationService interface {
	// SendNotification は通知を各チャネルに送信する
	SendNotification(ctx context.Context, notification *Notification) error

	// SendToChannel は特定のチャネルに通知を送信する
	SendToChannel(ctx context.Context, notification *Notification, channel Channel) error

	// FormatNotification はチャネルに応じた通知フォーマットを行う
	FormatNotification(notification *Notification, channelType ChannelType) (string, error)
}

// LineService はLINE通知に関するロジックを定義するインターフェース
type LineService interface {
	// SendMessage はLINEにメッセージを送信する
	SendMessage(ctx context.Context, channel *LineChannel, message string) error
}
