package notification

import (
	"context"
	"errors"

	"your-app/notification/domain/entity"
	"your-app/notification/domain/repository"
	"your-app/notification/domain/service"
)

// CreateNotificationInput は通知作成の入力パラメータ
type CreateNotificationInput struct {
	UserID    uint
	Type      entity.NotificationType
	Title     string
	Content   string
	RelatedID *uint
	Metadata  map[string]interface{}
	Channels  []string // "app", "line" などのチャネル指定
}

// CreateNotificationUseCase は通知作成ユースケース
type CreateNotificationUseCase struct {
	repo      repository.NotificationRepository
	notifySvc service.NotificationService
}

// NewCreateNotificationUseCase は通知作成ユースケースのインスタンスを作成する
func NewCreateNotificationUseCase(
	repo repository.NotificationRepository,
	notifySvc service.NotificationService,
) *CreateNotificationUseCase {
	return &CreateNotificationUseCase{
		repo:      repo,
		notifySvc: notifySvc,
	}
}

// Execute は通知作成ユースケースを実行する
func (uc *CreateNotificationUseCase) Execute(ctx context.Context, input CreateNotificationInput) (*entity.Notification, error) {
	// 入力検証
	if input.UserID == 0 {
		return nil, errors.New("user ID is required")
	}
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	if input.Content == "" {
		return nil, errors.New("content is required")
	}

	// 通知エンティティ作成
	notification := entity.NewNotification(
		input.UserID,
		input.Type,
		input.Title,
		input.Content,
		input.RelatedID,
		input.Metadata,
	)

	// チャネル指定の処理
	for _, channelName := range input.Channels {
		switch channelName {
		case "app":
			notification.AddChannel(entity.NewAppChannel(input.UserID))
		case "line":
			// LINEチャネル情報を取得
			lineChannel, err := uc.repo.GetLineChannelByUserID(ctx, input.UserID)
			if err != nil {
				// LINEチャネル情報が取得できない場合は、そのチャネルはスキップ
				continue
			}
			notification.AddChannel(lineChannel)
		}
	}

	// 通知の永続化
	if err := uc.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	return notification, nil
}
