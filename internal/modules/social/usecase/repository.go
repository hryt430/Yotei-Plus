package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
)

// === Service Interfaces ===

// SocialService はソーシャル機能のサービスインターフェース
type SocialService interface {
	// 友達関係管理
	SendFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID, message string) (*domain.Friendship, error)
	AcceptFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID) (*domain.Friendship, error)
	DeclineFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID) error
	RemoveFriend(ctx context.Context, userID, friendID uuid.UUID) error
	BlockUser(ctx context.Context, userID, targetID uuid.UUID) error
	UnblockUser(ctx context.Context, userID, targetID uuid.UUID) error

	// 友達一覧・検索
	GetFriends(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*FriendWithUserInfo, error)
	GetPendingRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*FriendshipWithUserInfo, error)
	GetSentRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*FriendshipWithUserInfo, error)
	GetMutualFriends(ctx context.Context, userID, targetID uuid.UUID) ([]*FriendWithUserInfo, error)

	// 招待管理
	CreateInvitation(ctx context.Context, input CreateInvitationInput) (*domain.Invitation, error)
	GetInvitation(ctx context.Context, invitationID uuid.UUID) (*domain.Invitation, error)
	GetInvitationByCode(ctx context.Context, code string) (*domain.Invitation, error)
	AcceptInvitation(ctx context.Context, code string, userID uuid.UUID) (*InvitationResult, error)
	DeclineInvitation(ctx context.Context, invitationID, userID uuid.UUID) error
	CancelInvitation(ctx context.Context, invitationID, inviterID uuid.UUID) error
	GetSentInvitations(ctx context.Context, inviterID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error)
	GetReceivedInvitations(ctx context.Context, inviteeID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error)

	// URL・招待コード
	GenerateInviteURL(ctx context.Context, invitationID uuid.UUID) (string, error)
	ValidateInviteCode(ctx context.Context, code string) (*domain.Invitation, error)

	// 関係性チェック
	GetRelationship(ctx context.Context, userID, targetID uuid.UUID) (*UserRelationship, error)
}

// === Input/Output Types ===

// CreateInvitationInput は招待作成の入力
type CreateInvitationInput struct {
	Type         domain.InvitationType
	Method       domain.InvitationMethod
	InviterID    uuid.UUID
	Message      string
	ExpiresHours int
	InviteeEmail *string
	TargetID     *uuid.UUID // Group IDなど
}

// InvitationResult は招待受諾の結果
type InvitationResult struct {
	Success    bool
	Message    string
	Friendship *domain.Friendship
	GroupID    *uuid.UUID
}

// UserRelationship はユーザー間の関係性
type UserRelationship struct {
	IsFriend        bool
	IsBlocked       bool
	RequestSent     bool
	RequestReceived bool
}

// FriendWithUserInfo は友達とユーザー情報
type FriendWithUserInfo struct {
	Friendship *domain.Friendship
	UserInfo   *commonDomain.UserInfo
}

// FriendshipWithUserInfo は友達関係とユーザー情報
type FriendshipWithUserInfo struct {
	Friendship *domain.Friendship
	UserInfo   *commonDomain.UserInfo
}

// FriendshipRepository は友達関係のリポジトリインターフェース
type FriendshipRepository interface {
	// 友達関係管理
	CreateFriendship(ctx context.Context, friendship *domain.Friendship) error
	GetFriendship(ctx context.Context, requesterID, addresseeID uuid.UUID) (*domain.Friendship, error)
	UpdateFriendship(ctx context.Context, friendship *domain.Friendship) error
	DeleteFriendship(ctx context.Context, requesterID, addresseeID uuid.UUID) error

	// 友達リスト・検索
	GetFriends(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Friendship, error)
	GetPendingRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Friendship, error)
	GetSentRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Friendship, error)

	// 関係チェック
	AreFriends(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error)
	IsBlocked(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error)

	// 統計
	GetFriendCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetMutualFriends(ctx context.Context, userID1, userID2 uuid.UUID) ([]*domain.Friendship, error)
}

// InvitationRepository は招待のリポジトリインターフェース
type InvitationRepository interface {
	// 招待管理
	CreateInvitation(ctx context.Context, invitation *domain.Invitation) error
	GetInvitationByID(ctx context.Context, id uuid.UUID) (*domain.Invitation, error)
	GetInvitationByCode(ctx context.Context, code string) (*domain.Invitation, error)
	UpdateInvitation(ctx context.Context, invitation *domain.Invitation) error
	DeleteInvitation(ctx context.Context, id uuid.UUID) error

	// 招待一覧
	GetSentInvitations(ctx context.Context, inviterID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error)
	GetReceivedInvitations(ctx context.Context, inviteeID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error)

	// 期限切れ招待の処理
	MarkExpiredInvitations(ctx context.Context) error
	DeleteExpiredInvitations(ctx context.Context, beforeDate time.Time) error

	// 招待検証
	IsValidInvitation(ctx context.Context, code string) (bool, error)
}
