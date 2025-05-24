// internal/modules/notification/infrastructure/gateway/app_notification_gateway.go
package gateway

import (
	"context"
	"fmt"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/internal/modules/notification/interface/websocket"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/output"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/persistence"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// AppNotificationGateway はアプリ内通知のゲートウェイ実装
type AppNotificationGateway struct {
	config     *config.Config
	repository persistence.NotificationRepository
	wsHub      *websocket.Hub // WebSocketハブ
	logger     logger.Logger
}

// NewAppNotificationGateway は新しいAppNotificationGatewayを作成する
func NewAppNotificationGateway(
	config *config.Config,
	repository persistence.NotificationRepository,
	wsHub *websocket.Hub,
	logger logger.Logger,
) output.AppNotificationGateway {
	return &AppNotificationGateway{
		config:     config,
		repository: repository,
		wsHub:      wsHub,
		logger:     logger,
	}
}

// SendNotification はアプリ内通知を送信する
func (g *AppNotificationGateway) SendNotification(ctx context.Context, userID, title, message string, metadata map[string]string) error {
	// 通知オブジェクトを作成
	notification := domain.NewNotification(
		userID,
		domain.AppNotification,
		title,
		message,
		metadata,
	)

	// データベースに保存
	if err := g.repository.Save(ctx, notification); err != nil {
		g.logger.Error("Failed to save app notification", logger.Error(err))
		return fmt.Errorf("failed to save app notification: %w", err)
	}

	// WebSocketでリアルタイム送信
	if g.wsHub != nil {
		g.wsHub.SendNotification(notification)
		g.logger.Info("Sent real-time notification", logger.Any("userID", userID), logger.Any("notificationID", notification.ID))
	}

	return nil
}

// MarkAsRead は通知を既読としてマークする
func (g *AppNotificationGateway) MarkAsRead(ctx context.Context, notificationID string) error {
	if err := g.repository.UpdateStatus(ctx, notificationID, domain.StatusRead); err != nil {
		g.logger.Error("Failed to mark notification as read", logger.Any("notificationID", notificationID), logger.Error(err))
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	g.logger.Info("Marked notification as read", logger.Any("notificationID", notificationID))
	return nil
}

// GetUnreadCount はユーザーの未読通知数を取得する
func (g *AppNotificationGateway) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	count, err := g.repository.CountByUserIDAndStatus(ctx, userID, domain.StatusSent)
	if err != nil {
		g.logger.Error("Failed to get unread count", logger.Any("userID", userID), logger.Error(err))
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}
