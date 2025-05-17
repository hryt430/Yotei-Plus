package notification

import (
	"context"

	"your-app/notification/domain/entity"
	"your-app/notification/domain/repository"
)

// ReadNotificationUseCase は通知既読ユースケース
type ReadNotificationUseCase struct {
	repo repository.NotificationRepository
}

// NewReadNotificationUseCase は通知既読ユースケースのインスタンスを作成する
func NewReadNotificationUseCase(repo repository.NotificationRepository) *ReadNotificationUseCase {
	return &ReadNotificationUseCase{
		repo: repo,
	}
}

// Execute は通知既読ユースケースを実行する
func (uc *ReadNotificationUseCase) Execute(ctx context.Context, notificationID uint) error {
	// 通知を既読にする
	return uc.repo.MarkAsRead(ctx, notificationID)
}

// GetNotifications はユーザーの通知一覧を取得する
func (uc *ReadNotificationUseCase) GetNotifications(ctx context.Context, userID uint, limit, offset int) ([]*entity.Notification, error) {
	return uc.repo.GetByUserID(ctx, userID, limit, offset)
}

// GetUnreadNotifications はユーザーの未読通知一覧を取得する
func (uc *ReadNotificationUseCase) GetUnreadNotifications(ctx context.Context, userID uint) ([]*entity.Notification, error) {
	return uc.repo.GetUnreadByUserID(ctx, userID)
}

// GetNotification は通知を取得する
func (uc *ReadNotificationUseCase) GetNotification(ctx context.Context, id uint) (*entity.Notification, error) {
	return uc.repo.GetByID(ctx, id)
}
