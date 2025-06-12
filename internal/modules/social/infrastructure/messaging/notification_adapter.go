// internal/modules/social/infrastructure/messaging/notification_adapter.go
package messaging

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	notificationInput "github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// SocialNotificationAdapter はSocialモジュールから通知サービスへの連携を行う
type SocialNotificationAdapter struct {
	notificationUseCase notificationInput.NotificationUseCase
	logger              logger.Logger
}

// NewSocialNotificationAdapter は新しいSocialNotificationAdapterを作成する
func NewSocialNotificationAdapter(
	notificationUseCase notificationInput.NotificationUseCase,
	logger logger.Logger,
) *SocialNotificationAdapter {
	return &SocialNotificationAdapter{
		notificationUseCase: notificationUseCase,
		logger:              logger,
	}
}

// SendFriendRequestNotification は友達申請通知を送信する
func (a *SocialNotificationAdapter) SendFriendRequestNotification(ctx context.Context, requesterID, addresseeID uuid.UUID, message string) error {
	// 通知タイトルとメッセージを構築
	title := "新しい友達申請"
	notificationMessage := "友達申請が届きました"
	if message != "" {
		notificationMessage = fmt.Sprintf("友達申請が届きました: %s", message)
	}

	// 通知作成
	input := notificationInput.CreateNotificationInput{
		UserID:  addresseeID.String(),
		Type:    "FRIEND_REQUEST",
		Title:   title,
		Message: notificationMessage,
		Metadata: map[string]string{
			"requester_id": requesterID.String(),
			"request_type": "friend_request",
			"action_type":  "received",
		},
		Channels: []string{"app"}, // アプリ内通知
	}

	notification, err := a.notificationUseCase.CreateNotification(ctx, input)
	if err != nil {
		a.logger.Error("Failed to create friend request notification",
			logger.Any("requesterID", requesterID),
			logger.Any("addresseeID", addresseeID),
			logger.Error(err))
		return fmt.Errorf("failed to create friend request notification: %w", err)
	}

	// 通知送信
	if err := a.notificationUseCase.SendNotification(ctx, notification.GetID()); err != nil {
		a.logger.Error("Failed to send friend request notification",
			logger.Any("notificationID", notification.GetID()),
			logger.Error(err))
		return fmt.Errorf("failed to send friend request notification: %w", err)
	}

	a.logger.Info("Friend request notification sent successfully",
		logger.Any("requesterID", requesterID),
		logger.Any("addresseeID", addresseeID),
		logger.Any("notificationID", notification.GetID()))

	return nil
}

// SendFriendAcceptedNotification は友達申請承認通知を送信する
func (a *SocialNotificationAdapter) SendFriendAcceptedNotification(ctx context.Context, requesterID, accepterID uuid.UUID) error {
	title := "友達申請が承認されました"
	notificationMessage := "友達申請が承認されました！"

	input := notificationInput.CreateNotificationInput{
		UserID:  requesterID.String(),
		Type:    "FRIEND_ACCEPTED",
		Title:   title,
		Message: notificationMessage,
		Metadata: map[string]string{
			"accepter_id":  accepterID.String(),
			"request_type": "friend_request",
			"action_type":  "accepted",
		},
		Channels: []string{"app"},
	}

	notification, err := a.notificationUseCase.CreateNotification(ctx, input)
	if err != nil {
		a.logger.Error("Failed to create friend accepted notification",
			logger.Any("requesterID", requesterID),
			logger.Any("accepterID", accepterID),
			logger.Error(err))
		return fmt.Errorf("failed to create friend accepted notification: %w", err)
	}

	if err := a.notificationUseCase.SendNotification(ctx, notification.GetID()); err != nil {
		a.logger.Error("Failed to send friend accepted notification",
			logger.Any("notificationID", notification.GetID()),
			logger.Error(err))
		return fmt.Errorf("failed to send friend accepted notification: %w", err)
	}

	a.logger.Info("Friend accepted notification sent successfully",
		logger.Any("requesterID", requesterID),
		logger.Any("accepterID", accepterID))

	return nil
}

// SendInvitationNotification は招待通知を送信する
func (a *SocialNotificationAdapter) SendInvitationNotification(ctx context.Context, invitation *domain.Invitation) error {
	// 招待タイプに応じてメッセージを構築
	var title, message string
	var channels []string

	switch invitation.Type {
	case domain.InvitationTypeFriend:
		title = "友達招待"
		message = "友達になりませんか？"
		channels = []string{"app"}
	case domain.InvitationTypeGroup:
		title = "グループ招待"
		message = "グループに招待されました"
		channels = []string{"app"}
	default:
		title = "招待"
		message = "招待が届きました"
		channels = []string{"app"}
	}

	if invitation.Message != "" {
		message = fmt.Sprintf("%s: %s", message, invitation.Message)
	}

	// 招待方法に応じて通知チャネルを調整
	switch invitation.Method {
	case domain.MethodInApp:
		channels = []string{"app"}
	case domain.MethodCode, domain.MethodQR:
		channels = []string{"app"}
		// QRコードや招待コードの場合、必要に応じて他のチャネルも追加
	}

	metadata := map[string]string{
		"invitation_id":     invitation.ID.String(),
		"invitation_type":   string(invitation.Type),
		"invitation_method": string(invitation.Method),
		"inviter_id":        invitation.InviterID.String(),
	}

	if invitation.Code != "" {
		metadata["invitation_code"] = invitation.Code
	}

	if invitation.TargetID != nil {
		metadata["target_id"] = invitation.TargetID.String()
	}

	// 被招待者が登録済みユーザーの場合
	if invitation.InviteeID != nil {
		input := notificationInput.CreateNotificationInput{
			UserID:   invitation.InviteeID.String(),
			Type:     "INVITATION_RECEIVED",
			Title:    title,
			Message:  message,
			Metadata: metadata,
			Channels: channels,
		}

		notification, err := a.notificationUseCase.CreateNotification(ctx, input)
		if err != nil {
			a.logger.Error("Failed to create invitation notification",
				logger.Any("invitationID", invitation.ID),
				logger.Error(err))
			return fmt.Errorf("failed to create invitation notification: %w", err)
		}

		if err := a.notificationUseCase.SendNotification(ctx, notification.GetID()); err != nil {
			a.logger.Error("Failed to send invitation notification",
				logger.Any("notificationID", notification.GetID()),
				logger.Error(err))
			return fmt.Errorf("failed to send invitation notification: %w", err)
		}

		a.logger.Info("Invitation notification sent successfully",
			logger.Any("invitationID", invitation.ID),
			logger.Any("inviteeID", invitation.InviteeID))
	} else {
		// 未登録ユーザーの場合は、メール通知などの外部通知を検討
		a.logger.Info("Invitation created for unregistered user",
			logger.Any("invitationID", invitation.ID),
			logger.Any("inviteeInfo", invitation.InviteeInfo))
	}

	return nil
}

// SendGroupInvitationNotification はグループ招待通知を送信する
func (a *SocialNotificationAdapter) SendGroupInvitationNotification(ctx context.Context, invitation *domain.Invitation, groupName string) error {
	if invitation.Type != domain.InvitationTypeGroup {
		return fmt.Errorf("invalid invitation type for group invitation: %s", invitation.Type)
	}

	title := fmt.Sprintf("「%s」への招待", groupName)
	message := fmt.Sprintf("グループ「%s」に招待されました", groupName)

	if invitation.Message != "" {
		message = fmt.Sprintf("%s: %s", message, invitation.Message)
	}

	metadata := map[string]string{
		"invitation_id":   invitation.ID.String(),
		"invitation_type": string(invitation.Type),
		"inviter_id":      invitation.InviterID.String(),
		"group_name":      groupName,
	}

	if invitation.TargetID != nil {
		metadata["group_id"] = invitation.TargetID.String()
	}

	if invitation.InviteeID != nil {
		input := notificationInput.CreateNotificationInput{
			UserID:   invitation.InviteeID.String(),
			Type:     "GROUP_INVITATION",
			Title:    title,
			Message:  message,
			Metadata: metadata,
			Channels: []string{"app"},
		}

		notification, err := a.notificationUseCase.CreateNotification(ctx, input)
		if err != nil {
			a.logger.Error("Failed to create group invitation notification",
				logger.Any("invitationID", invitation.ID),
				logger.Error(err))
			return fmt.Errorf("failed to create group invitation notification: %w", err)
		}

		if err := a.notificationUseCase.SendNotification(ctx, notification.GetID()); err != nil {
			a.logger.Error("Failed to send group invitation notification",
				logger.Any("notificationID", notification.GetID()),
				logger.Error(err))
			return fmt.Errorf("failed to send group invitation notification: %w", err)
		}

		a.logger.Info("Group invitation notification sent successfully",
			logger.Any("invitationID", invitation.ID),
			logger.Any("groupName", groupName))
	}

	return nil
}

// SendBulkNotifications は一括通知を送信する（バッチ処理用）
func (a *SocialNotificationAdapter) SendBulkNotifications(ctx context.Context, notifications []notificationInput.CreateNotificationInput) error {
	for _, input := range notifications {
		notification, err := a.notificationUseCase.CreateNotification(ctx, input)
		if err != nil {
			a.logger.Error("Failed to create bulk notification",
				logger.Any("userID", input.UserID),
				logger.Error(err))
			continue
		}

		if err := a.notificationUseCase.SendNotification(ctx, notification.GetID()); err != nil {
			a.logger.Error("Failed to send bulk notification",
				logger.Any("notificationID", notification.GetID()),
				logger.Error(err))
			continue
		}
	}

	a.logger.Info("Bulk notifications processed",
		logger.Any("count", len(notifications)))

	return nil
}
