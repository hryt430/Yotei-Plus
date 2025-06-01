package messaging

import (
	"context"

	notificationInput "github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
)

// NotificationAdapter は循環参照を避けるためのアダプター
type NotificationAdapter struct {
	notificationUseCase notificationInput.NotificationUseCase
}

// NewNotificationAdapter は新しいNotificationAdapterを作成
func NewNotificationAdapter(notificationUseCase notificationInput.NotificationUseCase) *NotificationAdapter {
	return &NotificationAdapter{
		notificationUseCase: notificationUseCase,
	}
}

// CreateNotification は通知を作成する
func (a *NotificationAdapter) CreateNotification(ctx context.Context, input notificationInput.CreateNotificationInput) (NotificationDomain, error) {
	notification, err := a.notificationUseCase.CreateNotification(ctx, input)
	if err != nil {
		return nil, err
	}

	// NotificationDomainインターフェースを満たすラッパーを返す
	return &NotificationWrapper{notification: notification}, nil
}

// NotificationWrapper はNotificationDomainインターフェースを実装
type NotificationWrapper struct {
	notification interface {
		GetID() string
		GetUserID() string
		GetTitle() string
	}
}

func (w *NotificationWrapper) GetID() string {
	return w.notification.GetID()
}

func (w *NotificationWrapper) GetUserID() string {
	return w.notification.GetUserID()
}

func (w *NotificationWrapper) GetTitle() string {
	return w.notification.GetTitle()
}
