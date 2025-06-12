package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
)

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
