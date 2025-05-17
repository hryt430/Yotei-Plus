package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hryt430/task-management/internal/modules/notification/domain"
	"github.com/hryt430/task-management/internal/modules/notification/infrastructure/messaging"
	"github.com/hryt430/task-management/internal/modules/notification/usecase/persistence"
	"github.com/hryt430/task-management/pkg/logger"
)

// AppNotificationGateway はアプリ内通知のゲートウェイ実装
type AppNotificationGateway struct {
	repository    persistence.NotificationRepository
	messageBroker messaging.MessageBroker
	logger        logger.Logger
}

// NewAppNotificationGateway は新しいAppNotificationGatewayを作成する
func NewAppNotificationGateway(
	repository persistence.NotificationRepository,
	messageBroker messaging.MessageBroker,
	logger logger.Logger,
) *AppNotificationGateway {
	return &AppNotificationGateway{
		repository:    repository,
		messageBroker: messageBroker,
		logger:        logger,
	}
}

// AppNotificationPayload はアプリ内通知のペイロード
type AppNotificationPayload struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	IsRead    bool              `json:"is_read"`
}

// SendNotification はアプリ内通知を送信する
func (g *AppNotificationGateway) SendNotification(ctx context.Context, userID, title, message string, metadata map[string]string) error {
	// リアルタイム通知のためのメッセージペイロードを作成
	payload := AppNotificationPayload{
		UserID:    userID,
		Title:     title,
		Message:   message,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	// トピック名の構築（ユーザー固有）
	topic := fmt.Sprintf("user.%s.notifications", userID)

	// メッセージをJSON形式に変換
	jsonData, err := json.Marshal(payload)
	if err != nil {
		g.logger.Error("Failed to marshal app notification payload", "error", err)
		return fmt.Errorf("failed to marshal app notification payload: %w", err)
	}

	// メッセージブローカーにPublish
	if err := g.messageBroker.Publish(ctx, topic, jsonData); err != nil {
		g.logger.Error("Failed to publish app notification", "error", err)
		return fmt.Errorf("failed to publish app notification: %w", err)
	}

	g.logger.Info("Successfully sent app notification", "userID", userID)
	return nil
}

// MarkAsRead は通知を既読としてマークする
func (g *AppNotificationGateway) MarkAsRead(ctx context.Context, notificationID string) error {
	// 通知を取得
	notification, err := g.repository.FindByID(ctx, notificationID)
	if err != nil {
		g.logger.Error("Failed to find notification", "id", notificationID, "error", err)
		return fmt.Errorf("failed to find notification: %w", err)
	}

	if notification == nil {
		return fmt.Errorf("notification not found: %s", notificationID)
	}

	// メタデータに既読フラグを設定
	notification.AddMetadata("read", "true")
	notification.AddMetadata("read_at", time.Now().Format(time.RFC3339))

	// 通知を更新
	if err := g.repository.Save(ctx, notification); err != nil {
		g.logger.Error("Failed to update notification", "id", notificationID, "error", err)
		return fmt.Errorf("failed to update notification: %w", err)
	}

	// 既読イベントをリアルタイムで通知
	payload := map[string]interface{}{
		"notification_id": notificationID,
		"user_id":         notification.UserID,
		"read_at":         notification.Metadata["read_at"],
	}

	// JSONに変換
	jsonData, err := json.Marshal(payload)
	if err != nil {
		g.logger.Error("Failed to marshal read notification payload", "error", err)
		return nil // 続行する - これは重要ではない副作用
	}

	// トピック名の構築（ユーザー固有）
	topic := fmt.Sprintf("user.%s.notifications.read", notification.UserID)

	// メッセージブローカーにPublish
	if err := g.messageBroker.Publish(ctx, topic, jsonData); err != nil {
		g.logger.Error("Failed to publish read notification event", "error", err)
		// 続行する - これは重要ではない副作用
	}

	return nil
}

// GetUnreadCount はユーザーの未読通知数を取得する
func (g *AppNotificationGateway) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	// ここでは簡単にステータスが送信済みで、メタデータにread=trueがない通知をカウント
	notifications, err := g.repository.FindByUserID(ctx, userID, 1000, 0)
	if err != nil {
		g.logger.Error("Failed to find user notifications", "userID", userID, "error", err)
		return 0, fmt.Errorf("failed to find user notifications: %w", err)
	}

	unreadCount := 0
	for _, notification := range notifications {
		// AppNotificationType かつ StatusSent であることを確認
		if notification.Type == domain.AppNotification && notification.Status == domain.StatusSent {
			// メタデータにread=trueがなければカウント
			if readValue, exists := notification.Metadata["read"]; !exists || readValue != "true" {
				unreadCount++
			}
		}
	}

	return unreadCount, nil
}
