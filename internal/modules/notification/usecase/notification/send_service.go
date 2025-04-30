package notification

import (
	"context"

	"your-app/notification/domain/entity"
	"your-app/notification/domain/service"
)

// SendNotificationUseCase は通知送信ユースケース
type SendNotificationUseCase struct {
	notifySvc service.NotificationService
}

// NewSendNotificationUseCase は通知送信ユースケースのインスタンスを作成する
func NewSendNotificationUseCase(notifySvc service.NotificationService) *SendNotificationUseCase {
	return &SendNotificationUseCase{
		notifySvc: notifySvc,
	}
}

// Execute は通知送信ユースケースを実行する
func (uc *SendNotificationUseCase) Execute(ctx context.Context, notification *entity.Notification) error {
	// 通知を各チャネルに送信
	return uc.notifySvc.SendNotification(ctx, notification)
}
