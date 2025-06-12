package usecase

import (
	"context"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
)

// GroupRepository はグループ関連のリポジトリインターフェース
type GroupRepository interface {
	// グループ管理
	CreateGroup(ctx context.Context, group *domain.Group) error
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	UpdateGroup(ctx context.Context, group *domain.Group) error
	DeleteGroup(ctx context.Context, id uuid.UUID) error

	// グループ検索・一覧
	ListGroupsByOwner(ctx context.Context, ownerID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Group, int, error)
	ListGroupsByMember(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Group, int, error)
	SearchGroups(ctx context.Context, query string, groupType *domain.GroupType, pagination commonDomain.Pagination) ([]*domain.Group, int, error)

	// グループメンバー管理
	AddMember(ctx context.Context, member *domain.GroupMember) error
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
	UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role domain.MemberRole) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	ListMembers(ctx context.Context, groupID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.GroupMember, error)

	// メンバーシップチェック
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (domain.MemberRole, error)

	// 統計情報
	GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error)
	GetGroupStats(ctx context.Context, groupID uuid.UUID) (*GroupStats, error)
}

// GroupStats はグループ統計情報
type GroupStats struct {
	MemberCount   int `json:"member_count"`
	TaskCount     int `json:"task_count,omitempty"`     // プロジェクトグループの場合
	ScheduleCount int `json:"schedule_count,omitempty"` // 予定共有グループの場合
	ActiveMembers int `json:"active_members"`           // 最近活動したメンバー数
}
