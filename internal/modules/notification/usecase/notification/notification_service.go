package notification

import (
	"context"

	"your-app/notification/domain/entity"
)

// NotificationService は通知処理のロジックを定義するインターフェース
type NotificationService interface {
	// SendNotification は通知を各チャネルに送信する
	SendNotification(ctx context.Context, notification *entity.Notification) error

	// SendToChannel は特定のチャネルに通知を送信する
	SendToChannel(ctx context.Context, notification *entity.Notification, channel entity.Channel) error

	// FormatNotification はチャネルに応じた通知フォーマットを行う
	FormatNotification(notification *entity.Notification, channelType entity.ChannelType) (string, error)
}

// DefaultNotificationService は通知サービスの基本実装
type DefaultNotificationService struct {
	lineService LineService
}

// NewNotificationService は通知サービスのインスタンスを作成する
func NewNotificationService(lineService LineService) NotificationService {
	return &DefaultNotificationService{
		lineService: lineService,
	}
}

// SendNotification は通知を各チャネルに送信する
func (s *DefaultNotificationService) SendNotification(ctx context.Context, notification *entity.Notification) error {
	for _, channel := range notification.Channels {
		if err := s.SendToChannel(ctx, notification, channel); err != nil {
			return err
		}
	}
	return nil
}

// SendToChannel は特定のチャネルに通知を送信する
func (s *DefaultNotificationService) SendToChannel(ctx context.Context, notification *entity.Notification, channel entity.Channel) error {
	// チャネルタイプに応じた送信処理
	switch channel.GetType() {
	case entity.AppInternal:
		// アプリ内通知の場合は何もしない（データベースに保存されるため）
		return nil
	case entity.LineMessage:
		lineChannel, ok := channel.(*entity.LineChannel)
		if !ok {
			return nil
		}

		message, err := s.FormatNotification(notification, entity.LineMessage)
		if err != nil {
			return err
		}

		return s.lineService.SendMessage(ctx, lineChannel, message)
	}

	return nil
}

// FormatNotification はチャネルに応じた通知フォーマットを行う
func (s *DefaultNotificationService) FormatNotification(notification *entity.Notification, channelType entity.ChannelType) (string, error) {
	switch channelType {
	case entity.LineMessage:
		// LINE用のシンプルなテキストフォーマット
		return notification.Title + "\n" + notification.Content, nil
	default:
		// デフォルトはそのままコンテンツを返す
		return notification.Content, nil
	}
}
