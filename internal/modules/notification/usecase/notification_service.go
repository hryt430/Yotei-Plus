package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/output"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/persistence"
)

type notificationUseCase struct {
	repository     persistence.NotificationRepository
	appGateway     output.AppNotificationGateway
	lineGateway    output.LineNotificationGateway
	webhookGateway output.WebhookOutput
}

// NewNotificationUseCase は通知ユースケースのインスタンスを作成する
func NewNotificationUseCase(
	repository persistence.NotificationRepository,
	appGateway output.AppNotificationGateway,
	lineGateway output.LineNotificationGateway,
	webhookGateway output.WebhookOutput,
) input.NotificationUseCase {
	return &notificationUseCase{
		repository:     repository,
		appGateway:     appGateway,
		lineGateway:    lineGateway,
		webhookGateway: webhookGateway,
	}
}

// CreateNotification は新しい通知を作成する
func (uc *notificationUseCase) CreateNotification(ctx context.Context, input input.CreateNotificationInput) (*domain.Notification, error) {
	// 入力バリデーション
	if input.UserID == "" {
		return nil, errors.New("user ID is required")
	}
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	if input.Message == "" {
		return nil, errors.New("message is required")
	}

	// 通知タイプの変換
	var notificationType domain.NotificationType
	switch input.Type {
	case "APP_NOTIFICATION":
		notificationType = domain.AppNotification
	case "TASK_ASSIGNED":
		notificationType = domain.TaskAssigned
	case "TASK_COMPLETED":
		notificationType = domain.TaskCompleted
	case "TASK_DUE_SOON":
		notificationType = domain.TaskDueSoon
	case "SYSTEM_NOTICE":
		notificationType = domain.SystemNotice
	default:
		notificationType = domain.SystemNotice
	}

	// 通知エンティティの作成
	notification := domain.NewNotification(
		input.UserID,
		notificationType,
		input.Title,
		input.Message,
		input.Metadata,
	)

	// チャネルの追加
	for _, channelName := range input.Channels {
		switch channelName {
		case "app":
			notification.AddChannel(domain.NewAppChannel(input.UserID))
		case "line":
			// LINE通知用のチャネル追加
			// 実際の実装ではLINEユーザーIDを取得する必要がある
			lineUserID := input.UserID // 簡易化のため同じIDを使用
			if lineID, ok := input.Metadata["line_user_id"]; ok {
				lineUserID = lineID
			}
			notification.AddChannel(domain.NewLineChannel(input.UserID, lineUserID, ""))
		}
	}

	// 通知をデータベースに保存
	if err := uc.repository.Save(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	return notification, nil
}

// GetNotification は通知を取得する
func (uc *notificationUseCase) GetNotification(ctx context.Context, id string) (*domain.Notification, error) {
	notification, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}
	return notification, nil
}

// GetUserNotifications はユーザーの通知一覧を取得する
func (uc *notificationUseCase) GetUserNotifications(ctx context.Context, input input.GetNotificationsInput) ([]*domain.Notification, error) {
	notifications, err := uc.repository.FindByUserID(ctx, input.UserID, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find user notifications: %w", err)
	}
	return notifications, nil
}

// SendNotification は通知を送信する
func (uc *notificationUseCase) SendNotification(ctx context.Context, id string) error {
	// 通知の取得
	notification, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find notification: %w", err)
	}

	if notification == nil {
		return errors.New("notification not found")
	}

	// 各チャネルに通知を送信
	var sendError error
	for _, channel := range notification.Channels {
		switch channel.GetType() {
		case domain.AppInternal:
			appChannel := channel.(*domain.AppChannel)
			sendError = uc.appGateway.SendNotification(
				ctx,
				appChannel.UserID,
				notification.Title,
				notification.Message,
				notification.Metadata,
			)
		case domain.LineMessage:
			lineChannel := channel.(*domain.LineChannel)
			sendError = uc.lineGateway.SendLineNotification(
				ctx,
				lineChannel.LineUserID,
				notification.Title+"\n"+notification.Message,
			)
		}

		if sendError != nil {
			// 送信エラーを記録
			notification.MarkAsFailed()
			if err := uc.repository.Save(ctx, notification); err != nil {
				return fmt.Errorf("failed to update notification status: %w", err)
			}

			// Webhookイベント送信
			_ = uc.webhookGateway.SendWebhook(ctx, output.EventNotificationFailed, map[string]interface{}{
				"notification_id": notification.ID,
				"error":           sendError.Error(),
			})

			return fmt.Errorf("failed to send notification: %w", sendError)
		}
	}

	// 送信成功を記録
	notification.MarkAsSent()
	if err := uc.repository.Save(ctx, notification); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	// Webhookイベント送信
	_ = uc.webhookGateway.SendWebhook(ctx, output.EventNotificationSent, map[string]interface{}{
		"notification_id": notification.ID,
		"sent_at":         time.Now(),
	})

	return nil
}

// MarkNotificationAsRead は通知を既読としてマークする
func (uc *notificationUseCase) MarkNotificationAsRead(ctx context.Context, id string) error {
	// 通知を既読にする
	if err := uc.repository.UpdateStatus(ctx, id, domain.StatusRead); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	// アプリ内通知を既読としてマーク
	if err := uc.appGateway.MarkAsRead(ctx, id); err != nil {
		return fmt.Errorf("failed to mark app notification as read: %w", err)
	}

	// Webhookイベント送信
	_ = uc.webhookGateway.SendWebhook(ctx, output.EventNotificationRead, map[string]interface{}{
		"notification_id": id,
		"read_at":         time.Now(),
	})

	return nil
}

// GetUnreadNotificationCount はユーザーの未読通知数を取得する
func (uc *notificationUseCase) GetUnreadNotificationCount(ctx context.Context, userID string) (int, error) {
	count, err := uc.appGateway.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}
	return count, nil
}
