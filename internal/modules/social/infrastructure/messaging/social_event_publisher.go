package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// SocialEventType はソーシャル関連のイベントタイプ
type SocialEventType string

const (
	EventFriendRequestSent     SocialEventType = "social.friend_request.sent"
	EventFriendRequestAccepted SocialEventType = "social.friend_request.accepted"
	EventFriendRequestDeclined SocialEventType = "social.friend_request.declined"
	EventFriendRemoved         SocialEventType = "social.friend.removed"
	EventUserBlocked           SocialEventType = "social.user.blocked"
	EventUserUnblocked         SocialEventType = "social.user.unblocked"
	EventInvitationCreated     SocialEventType = "social.invitation.created"
	EventInvitationAccepted    SocialEventType = "social.invitation.accepted"
	EventInvitationDeclined    SocialEventType = "social.invitation.declined"
	EventInvitationExpired     SocialEventType = "social.invitation.expired"
)

// SocialEvent はソーシャル関連のイベント
type SocialEvent struct {
	ID        string          `json:"id"`
	Type      SocialEventType `json:"type"`
	Payload   interface{}     `json:"payload"`
	UserID    uuid.UUID       `json:"user_id"`
	CreatedAt time.Time       `json:"created_at"`
}

// FriendRequestPayload は友達申請イベントのペイロード
type FriendRequestPayload struct {
	FriendshipID uuid.UUID `json:"friendship_id"`
	RequesterID  uuid.UUID `json:"requester_id"`
	AddresseeID  uuid.UUID `json:"addressee_id"`
	Message      string    `json:"message"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// InvitationPayload は招待イベントのペイロード
type InvitationPayload struct {
	InvitationID uuid.UUID               `json:"invitation_id"`
	Type         domain.InvitationType   `json:"type"`
	Method       domain.InvitationMethod `json:"method"`
	InviterID    uuid.UUID               `json:"inviter_id"`
	InviteeID    *uuid.UUID              `json:"invitee_id,omitempty"`
	TargetID     *uuid.UUID              `json:"target_id,omitempty"`
	Code         string                  `json:"code,omitempty"`
	Message      string                  `json:"message"`
	Status       domain.InvitationStatus `json:"status"`
	CreatedAt    time.Time               `json:"created_at"`
}

// NotificationAdapter は通知サービスとの連携アダプター
type NotificationAdapter interface {
	SendFriendRequestNotification(ctx context.Context, requesterID, addresseeID uuid.UUID, message string) error
	SendFriendAcceptedNotification(ctx context.Context, requesterID, accepterID uuid.UUID) error
	SendInvitationNotification(ctx context.Context, invitation *domain.Invitation) error
}

// SocialEventPublisher はソーシャルイベントを発行する
type SocialEventPublisher struct {
	notificationAdapter NotificationAdapter
	logger              logger.Logger
}

// NewSocialEventPublisher は新しいSocialEventPublisherを作成する
func NewSocialEventPublisher(notificationAdapter NotificationAdapter, logger logger.Logger) *SocialEventPublisher {
	return &SocialEventPublisher{
		notificationAdapter: notificationAdapter,
		logger:              logger,
	}
}

// PublishFriendRequestSent は友達申請送信イベントを発行する
func (p *SocialEventPublisher) PublishFriendRequestSent(ctx context.Context, friendship *domain.Friendship, message string) error {
	payload := FriendRequestPayload{
		FriendshipID: friendship.ID,
		RequesterID:  friendship.RequesterID,
		AddresseeID:  friendship.AddresseeID,
		Message:      message,
		Status:       string(friendship.Status),
		CreatedAt:    friendship.CreatedAt,
	}

	event := &SocialEvent{
		ID:        uuid.New().String(),
		Type:      EventFriendRequestSent,
		Payload:   payload,
		UserID:    friendship.AddresseeID, // 通知対象ユーザー
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing friend request sent event",
		logger.Any("eventID", event.ID),
		logger.Any("requesterID", friendship.RequesterID),
		logger.Any("addresseeID", friendship.AddresseeID))

	// 通知サービスに送信
	if err := p.notificationAdapter.SendFriendRequestNotification(ctx, friendship.RequesterID, friendship.AddresseeID, message); err != nil {
		p.logger.Error("Failed to send friend request notification",
			logger.Any("eventID", event.ID),
			logger.Error(err))
		return fmt.Errorf("failed to send friend request notification: %w", err)
	}

	return nil
}

// PublishFriendRequestAccepted は友達申請承認イベントを発行する
func (p *SocialEventPublisher) PublishFriendRequestAccepted(ctx context.Context, friendship *domain.Friendship) error {
	payload := FriendRequestPayload{
		FriendshipID: friendship.ID,
		RequesterID:  friendship.RequesterID,
		AddresseeID:  friendship.AddresseeID,
		Status:       string(friendship.Status),
		CreatedAt:    friendship.CreatedAt,
	}

	event := &SocialEvent{
		ID:        uuid.New().String(),
		Type:      EventFriendRequestAccepted,
		Payload:   payload,
		UserID:    friendship.RequesterID, // 申請者に通知
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing friend request accepted event",
		logger.Any("eventID", event.ID),
		logger.Any("requesterID", friendship.RequesterID),
		logger.Any("addresseeID", friendship.AddresseeID))

	// 通知サービスに送信
	if err := p.notificationAdapter.SendFriendAcceptedNotification(ctx, friendship.RequesterID, friendship.AddresseeID); err != nil {
		p.logger.Error("Failed to send friend accepted notification",
			logger.Any("eventID", event.ID),
			logger.Error(err))
		return fmt.Errorf("failed to send friend accepted notification: %w", err)
	}

	return nil
}

// PublishFriendRequestDeclined は友達申請拒否イベントを発行する
func (p *SocialEventPublisher) PublishFriendRequestDeclined(ctx context.Context, friendship *domain.Friendship) error {
	payload := FriendRequestPayload{
		FriendshipID: friendship.ID,
		RequesterID:  friendship.RequesterID,
		AddresseeID:  friendship.AddresseeID,
		Status:       string(friendship.Status),
		CreatedAt:    friendship.CreatedAt,
	}

	event := &SocialEvent{
		ID:        uuid.New().String(),
		Type:      EventFriendRequestDeclined,
		Payload:   payload,
		UserID:    friendship.RequesterID,
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing friend request declined event",
		logger.Any("eventID", event.ID),
		logger.Any("requesterID", friendship.RequesterID),
		logger.Any("addresseeID", friendship.AddresseeID))

	// 拒否の場合は通知しない（プライバシー配慮）
	return nil
}

// PublishFriendRemoved は友達削除イベントを発行する
func (p *SocialEventPublisher) PublishFriendRemoved(ctx context.Context, userID, friendID uuid.UUID) error {
	event := &SocialEvent{
		ID:   uuid.New().String(),
		Type: EventFriendRemoved,
		Payload: map[string]interface{}{
			"user_id":   userID,
			"friend_id": friendID,
		},
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing friend removed event",
		logger.Any("eventID", event.ID),
		logger.Any("userID", userID),
		logger.Any("friendID", friendID))

	return nil
}

// PublishUserBlocked はユーザーブロックイベントを発行する
func (p *SocialEventPublisher) PublishUserBlocked(ctx context.Context, userID, targetID uuid.UUID) error {
	event := &SocialEvent{
		ID:   uuid.New().String(),
		Type: EventUserBlocked,
		Payload: map[string]interface{}{
			"user_id":   userID,
			"target_id": targetID,
		},
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing user blocked event",
		logger.Any("eventID", event.ID),
		logger.Any("userID", userID),
		logger.Any("targetID", targetID))

	return nil
}

// PublishInvitationCreated は招待作成イベントを発行する
func (p *SocialEventPublisher) PublishInvitationCreated(ctx context.Context, invitation *domain.Invitation) error {
	payload := InvitationPayload{
		InvitationID: invitation.ID,
		Type:         invitation.Type,
		Method:       invitation.Method,
		InviterID:    invitation.InviterID,
		InviteeID:    invitation.InviteeID,
		TargetID:     invitation.TargetID,
		Code:         invitation.Code,
		Message:      invitation.Message,
		Status:       invitation.Status,
		CreatedAt:    invitation.CreatedAt,
	}

	event := &SocialEvent{
		ID:        uuid.New().String(),
		Type:      EventInvitationCreated,
		Payload:   payload,
		UserID:    invitation.InviterID,
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing invitation created event",
		logger.Any("eventID", event.ID),
		logger.Any("invitationID", invitation.ID),
		logger.Any("inviterID", invitation.InviterID))

	// 通知サービスに送信
	if err := p.notificationAdapter.SendInvitationNotification(ctx, invitation); err != nil {
		p.logger.Error("Failed to send invitation notification",
			logger.Any("eventID", event.ID),
			logger.Error(err))
		return fmt.Errorf("failed to send invitation notification: %w", err)
	}

	return nil
}

// PublishInvitationAccepted は招待受諾イベントを発行する
func (p *SocialEventPublisher) PublishInvitationAccepted(ctx context.Context, invitation *domain.Invitation) error {
	payload := InvitationPayload{
		InvitationID: invitation.ID,
		Type:         invitation.Type,
		Method:       invitation.Method,
		InviterID:    invitation.InviterID,
		InviteeID:    invitation.InviteeID,
		TargetID:     invitation.TargetID,
		Status:       invitation.Status,
		CreatedAt:    invitation.CreatedAt,
	}

	event := &SocialEvent{
		ID:        uuid.New().String(),
		Type:      EventInvitationAccepted,
		Payload:   payload,
		UserID:    invitation.InviterID, // 招待者に通知
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing invitation accepted event",
		logger.Any("eventID", event.ID),
		logger.Any("invitationID", invitation.ID),
		logger.Any("inviterID", invitation.InviterID))

	return nil
}

// PublishInvitationDeclined は招待拒否イベントを発行する
func (p *SocialEventPublisher) PublishInvitationDeclined(ctx context.Context, invitation *domain.Invitation) error {
	payload := InvitationPayload{
		InvitationID: invitation.ID,
		Type:         invitation.Type,
		Method:       invitation.Method,
		InviterID:    invitation.InviterID,
		InviteeID:    invitation.InviteeID,
		Status:       invitation.Status,
		CreatedAt:    invitation.CreatedAt,
	}

	event := &SocialEvent{
		ID:        uuid.New().String(),
		Type:      EventInvitationDeclined,
		Payload:   payload,
		UserID:    invitation.InviterID,
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing invitation declined event",
		logger.Any("eventID", event.ID),
		logger.Any("invitationID", invitation.ID))

	return nil
}

// PublishBulkInvitationsExpired は期限切れ招待の一括処理イベントを発行する
func (p *SocialEventPublisher) PublishBulkInvitationsExpired(ctx context.Context, expiredCount int) error {
	event := &SocialEvent{
		ID:   uuid.New().String(),
		Type: EventInvitationExpired,
		Payload: map[string]interface{}{
			"expired_count": expiredCount,
			"processed_at":  time.Now(),
		},
		UserID:    uuid.Nil, // システムイベント
		CreatedAt: time.Now(),
	}

	p.logger.Info("Publishing bulk invitations expired event",
		logger.Any("eventID", event.ID),
		logger.Any("expiredCount", expiredCount))

	return nil
}
