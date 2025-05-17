package domain

import (
	"context"
)

// NotificationRepository は通知のデータアクセスのインターフェース
type NotificationRepository interface {
	// Create は新しい通知を作成する
	Create(ctx context.Context, notification *Notification) error

	// GetByID はIDから通知を取得する
	GetByID(ctx context.Context, id uint) (*Notification, error)

	// GetByUserID はユーザーIDから通知のリストを取得する
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*Notification, error)

	// GetUnreadByUserID はユーザーIDから未読通知のリストを取得する
	GetUnreadByUserID(ctx context.Context, userID uint) ([]*Notification, error)

	// Update は通知を更新する
	Update(ctx context.Context, notification *Notification) error

	// MarkAsRead は通知を既読にする
	MarkAsRead(ctx context.Context, id uint) error

	// Delete は通知を削除する
	Delete(ctx context.Context, id uint) error

	// GetUnreadCount はユーザーの未読通知数を取得する
	GetUnreadCount(ctx context.Context, userID uint) (int, error)

	// GetLineChannelByUserID はユーザーIDからLINEチャネル情報を取得する
	GetLineChannelByUserID(ctx context.Context, userID uint) (*LineChannel, error)

	// SaveUserLineID はユーザーIDとLINE IDのマッピングを保存する
	SaveUserLineID(ctx context.Context, userID uint, lineUserID string) error
}
