package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/output"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/persistence"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// UserInfo は通知モジュール用のユーザー情報（共通定義を使用）
type UserInfo = commonDomain.UserInfo

// UserValidator は通知モジュール用のユーザーバリデーター（共通定義を使用）
type UserValidator = commonDomain.UserValidator

type notificationUseCase struct {
	repository    persistence.NotificationRepository
	appGateway    output.AppNotificationGateway
	lineGateway   output.LineNotificationGateway
	userValidator UserValidator
	logger        logger.Logger
}

// NewNotificationUseCase は通知ユースケースのインスタンスを作成する
func NewNotificationUseCase(
	repository persistence.NotificationRepository,
	appGateway output.AppNotificationGateway,
	lineGateway output.LineNotificationGateway,
	userValidator UserValidator,
	logger logger.Logger,
) input.NotificationUseCase {
	return &notificationUseCase{
		repository:    repository,
		appGateway:    appGateway,
		lineGateway:   lineGateway,
		userValidator: userValidator,
		logger:        logger,
	}
}

// CreateNotification は新しい通知を作成する
func (uc *notificationUseCase) CreateNotification(ctx context.Context, input input.CreateNotificationInput) (*domain.Notification, error) {
	// 入力バリデーション
	if err := uc.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// ユーザー存在確認（統一インターフェース使用）
	exists, err := uc.userValidator.UserExists(ctx, input.UserID)
	if err != nil {
		uc.logger.Error("Failed to validate user existence", logger.Any("userID", input.UserID), logger.Error(err))
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return nil, errors.New("user not found")
	}

	// 通知タイプの変換
	notificationType := uc.convertNotificationType(input.Type)

	// 通知エンティティの作成
	notification := domain.NewNotification(
		input.UserID,
		notificationType,
		input.Title,
		input.Message,
		input.Metadata,
	)

	// チャネルの追加
	if err := uc.addChannelsToNotification(ctx, notification, input); err != nil {
		return nil, fmt.Errorf("failed to add channels: %w", err)
	}

	// 通知をデータベースに保存
	if err := uc.repository.Save(ctx, notification); err != nil {
		uc.logger.Error("Failed to save notification", logger.Any("notificationID", notification.ID), logger.Error(err))
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	uc.logger.Info("Notification created successfully", logger.Any("notificationID", notification.ID), logger.Any("userID", input.UserID))
	return notification, nil
}

// CreateScheduledNotification はスケジュール通知を作成する
func (uc *notificationUseCase) CreateScheduledNotification(
	ctx context.Context,
	userID, title, message string,
	notificationType domain.NotificationType,
	scheduledTime time.Time,
	metadata map[string]string,
) error {
	// ユーザー存在確認
	exists, err := uc.userValidator.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return errors.New("user not found")
	}

	// スケジュール時刻が未来であることを確認
	if scheduledTime.Before(time.Now()) {
		return errors.New("scheduled time must be in the future")
	}

	notification := domain.NewNotification(userID, notificationType, title, message, metadata)
	notification.AddMetadata("scheduled_time", scheduledTime.Format(time.RFC3339))
	notification.AddChannel(domain.NewAppChannel(userID))

	if err := uc.repository.Save(ctx, notification); err != nil {
		return fmt.Errorf("failed to save scheduled notification: %w", err)
	}

	uc.logger.Info("Scheduled notification created",
		logger.Any("notificationID", notification.ID),
		logger.Any("scheduledTime", scheduledTime))

	return nil
}

// SendNotification は通知を送信する（エラーハンドリング強化）
func (uc *notificationUseCase) SendNotification(ctx context.Context, id string) error {
	notification, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find notification: %w", err)
	}

	if notification == nil {
		return errors.New("notification not found")
	}

	// 送信済みの場合はスキップ
	if notification.Status == domain.StatusSent {
		uc.logger.Warn("Notification already sent", logger.Any("notificationID", id))
		return nil
	}

	// 各チャネルに通知を送信（並行処理で高速化）
	return uc.sendToAllChannels(ctx, notification)
}

// sendToAllChannels は全チャネルに並行して通知を送信
func (uc *notificationUseCase) sendToAllChannels(ctx context.Context, notification *domain.Notification) error {
	if len(notification.Channels) == 0 {
		// デフォルトでアプリ内通知チャネルを追加
		notification.AddChannel(domain.NewAppChannel(notification.UserID))
	}

	errorCh := make(chan error, len(notification.Channels))

	for _, channel := range notification.Channels {
		go func(ch domain.Channel) {
			defer func() {
				if r := recover(); r != nil {
					uc.logger.Error("Panic in notification sending", logger.Any("panic", r))
					errorCh <- fmt.Errorf("panic occurred: %v", r)
				}
			}()

			err := uc.sendToChannel(ctx, notification, ch)
			errorCh <- err
		}(channel)
	}

	// 全チャネルの送信結果を待機
	var errors []error
	for i := 0; i < len(notification.Channels); i++ {
		if err := <-errorCh; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		// 一部でも失敗した場合は失敗とする
		notification.MarkAsFailed()
		uc.repository.Save(ctx, notification)
		return fmt.Errorf("failed to send to %d channels: %v", len(errors), errors[0])
	}

	// 全て成功
	notification.MarkAsSent()
	if err := uc.repository.Save(ctx, notification); err != nil {
		uc.logger.Error("Failed to update notification status to sent", logger.Error(err))
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

// sendToChannel は個別チャネルに送信
func (uc *notificationUseCase) sendToChannel(ctx context.Context, notification *domain.Notification, channel domain.Channel) error {
	switch channel.GetType() {
	case domain.AppInternal:
		appChannel := channel.(*domain.AppChannel)
		return uc.appGateway.SendNotification(
			ctx,
			appChannel.UserID,
			notification.Title,
			notification.Message,
			notification.Metadata,
		)
	case domain.LineMessage:
		lineChannel := channel.(*domain.LineChannel)
		return uc.lineGateway.SendLineNotification(
			ctx,
			lineChannel.LineUserID,
			notification.Title+"\n"+notification.Message,
		)
	default:
		return fmt.Errorf("unsupported channel type: %s", channel.GetType())
	}
}

// ProcessPendingNotifications は保留中の通知を処理する
func (uc *notificationUseCase) ProcessPendingNotifications(ctx context.Context, batchSize int) error {
	notifications, err := uc.repository.FindPendingNotifications(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to find pending notifications: %w", err)
	}

	uc.logger.Info("Processing pending notifications", logger.Any("count", len(notifications)))

	for _, notification := range notifications {
		if err := uc.SendNotification(ctx, notification.ID); err != nil {
			uc.logger.Error("Failed to send pending notification",
				logger.Any("notificationID", notification.ID),
				logger.Error(err))
			continue
		}
	}

	return nil
}

// validateCreateInput は作成入力をバリデーション
func (uc *notificationUseCase) validateCreateInput(input input.CreateNotificationInput) error {
	if input.UserID == "" {
		return errors.New("user ID is required")
	}
	if input.Title == "" {
		return errors.New("title is required")
	}
	if input.Message == "" {
		return errors.New("message is required")
	}
	if len(input.Channels) == 0 {
		return errors.New("at least one channel is required")
	}
	return nil
}

// convertNotificationType は文字列を通知タイプに変換
func (uc *notificationUseCase) convertNotificationType(typeStr string) domain.NotificationType {
	switch typeStr {
	case "APP_NOTIFICATION":
		return domain.AppNotification
	case "TASK_ASSIGNED":
		return domain.TaskAssigned
	case "TASK_COMPLETED":
		return domain.TaskCompleted
	case "TASK_DUE_SOON":
		return domain.TaskDueSoon
	case "SYSTEM_NOTICE":
		return domain.SystemNotice
	default:
		return domain.SystemNotice
	}
}

// addChannelsToNotification は通知にチャネルを追加
func (uc *notificationUseCase) addChannelsToNotification(ctx context.Context, notification *domain.Notification, input input.CreateNotificationInput) error {
	for _, channelName := range input.Channels {
		switch channelName {
		case "app":
			notification.AddChannel(domain.NewAppChannel(input.UserID))
		case "line":
			lineUserID := input.UserID
			if lineID, ok := input.Metadata["line_user_id"]; ok {
				lineUserID = lineID
			}
			notification.AddChannel(domain.NewLineChannel(input.UserID, lineUserID, ""))
		default:
			uc.logger.Warn("Unknown channel type", logger.Any("channel", channelName))
		}
	}
	return nil
}

// MarkNotificationAsRead は通知を既読としてマークする
func (uc *notificationUseCase) MarkNotificationAsRead(ctx context.Context, id string) error {
	if err := uc.repository.UpdateStatus(ctx, id, domain.StatusRead); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if err := uc.appGateway.MarkAsRead(ctx, id); err != nil {
		uc.logger.Error("Failed to mark app notification as read", logger.Error(err))
		// アプリ内通知の既読更新失敗は致命的ではないので続行
	}

	return nil
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

// GetUnreadNotificationCount はユーザーの未読通知数を取得する
func (uc *notificationUseCase) GetUnreadNotificationCount(ctx context.Context, userID string) (int, error) {
	count, err := uc.appGateway.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}
	return count, nil
}
