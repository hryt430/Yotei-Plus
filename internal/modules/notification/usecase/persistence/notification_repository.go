package persistence

import (
	"context"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
)

// NotificationRepository は通知のリポジトリインターフェース
type NotificationRepository interface {
	// Save は通知を保存する
	Save(ctx context.Context, notification *domain.Notification) error

	// FindByID はIDから通知を取得する
	FindByID(ctx context.Context, id string) (*domain.Notification, error)

	// FindByUserID はユーザーIDから通知のリストを取得する
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Notification, error)

	// UpdateStatus は通知のステータスを更新する
	UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error

	// CountByUserIDAndStatus はユーザーIDとステータスに基づいて通知数を取得する
	CountByUserIDAndStatus(ctx context.Context, userID string, status domain.NotificationStatus) (int, error)

	// FindPendingNotifications は保留中の通知を取得する
	FindPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)
}
