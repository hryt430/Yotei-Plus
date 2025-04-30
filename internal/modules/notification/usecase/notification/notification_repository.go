package notification

import (
	"context"

	"your-app/notification/domain/entity"
)

// NotificationRepository は通知のデータアクセスのインターフェース
type NotificationRepository interface {
	// Create は新しい通知を作成する
	Create(ctx context.Context, notification *entity.Notification) error

	// GetByID はIDから通知を取得する
	GetByID(ctx context.Context, id uint) (*entity.Notification, error)

	// GetByUserID はユーザーIDから通知のリストを取得する
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entity.Notification, error)

	// GetUnreadByUserID はユーザーIDから未読通知のリストを取得する
	GetUnreadByUserID(ctx context.Context, userID uint) ([]*entity.Notification, error)

	// Update は通知を更新する
	Update(ctx context.Context, notification *entity.Notification) error

	// MarkAsRead は通知を既読にする
	MarkAsRead(ctx context.Context, id uint) error

	// Delete は通知を削除する
	Delete(ctx context.Context, id uint) error

	// GetLineChannelByUserID はユーザーIDからLINEチャネル情報を取得する
	GetLineChannelByUserID(ctx context.Context, userID uint) (*entity.LineChannel, error)
}
