package usecase

import (
	"context"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
)

// === Service Interfaces ===

// GroupService はグループ機能のサービスインターフェース
type GroupService interface {
	// グループ管理
	CreateGroup(ctx context.Context, input CreateGroupInput) (*domain.Group, error)
	GetGroup(ctx context.Context, groupID, requesterID uuid.UUID) (*GroupWithMembers, error)
	UpdateGroup(ctx context.Context, groupID uuid.UUID, input UpdateGroupInput, requesterID uuid.UUID) (*domain.Group, error)
	DeleteGroup(ctx context.Context, groupID, requesterID uuid.UUID) error

	// グループ一覧・検索
	GetMyGroups(ctx context.Context, userID uuid.UUID, groupType *domain.GroupType, pagination commonDomain.Pagination) ([]*domain.Group, int, error)
	SearchGroups(ctx context.Context, query string, groupType *domain.GroupType, pagination commonDomain.Pagination) ([]*domain.Group, int, error)

	// メンバー管理
	AddMember(ctx context.Context, groupID, userID, inviterID uuid.UUID, role domain.MemberRole) error
	RemoveMember(ctx context.Context, groupID, userID, requesterID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, groupID, userID, requesterID uuid.UUID, newRole domain.MemberRole) error
	GetMembers(ctx context.Context, groupID uuid.UUID, pagination commonDomain.Pagination) ([]*MemberWithUserInfo, error)

	// 友達招待（Social連携）
	InviteFriendsToGroup(ctx context.Context, groupID, inviterID uuid.UUID, friendIDs []uuid.UUID, message string) ([]*GroupInviteResult, error)
	GetAvailableFriends(ctx context.Context, groupID, userID uuid.UUID) ([]*AvailableFriend, error)

	// 権限・統計
	CheckPermission(ctx context.Context, groupID, userID uuid.UUID, action GroupAction) (bool, error)
	GetUserRole(ctx context.Context, groupID, userID uuid.UUID) (domain.MemberRole, error)
	GetGroupStats(ctx context.Context, groupID, requesterID uuid.UUID) (*domain.GroupStats, error)
	GetGroupActivity(ctx context.Context, groupID uuid.UUID, days int) (*GroupActivity, error)
}

// === Input/Output Types ===

// CreateGroupInput はグループ作成の入力
type CreateGroupInput struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Type        domain.GroupType      `json:"type"`
	OwnerID     uuid.UUID             `json:"owner_id"`
	Settings    domain.GroupSettings  `json:"settings"`
}

// UpdateGroupInput はグループ更新の入力
type UpdateGroupInput struct {
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	Settings    *domain.GroupSettings `json:"settings,omitempty"`
}

// GroupWithMembers はグループとメンバー情報
type GroupWithMembers struct {
	Group   *domain.Group
	Members []*MemberWithUserInfo
	MyRole  domain.MemberRole
}

// MemberWithUserInfo はメンバーとユーザー情報
type MemberWithUserInfo struct {
	Member   *domain.GroupMember
	UserInfo *commonDomain.UserInfo
}

// GroupInviteResult はグループ招待の結果
type GroupInviteResult struct {
	FriendID uuid.UUID `json:"friend_id"`
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Error    string    `json:"error,omitempty"`
}

// AvailableFriend は招待可能な友達
type AvailableFriend struct {
	UserID     uuid.UUID              `json:"user_id"`
	UserInfo   *commonDomain.UserInfo `json:"user_info"`
	IsMember   bool                   `json:"is_member"`
	IsInvited  bool                   `json:"is_invited"`
}

// GroupAction はグループでのアクション
type GroupAction string

const (
	ActionViewGroup      GroupAction = "VIEW_GROUP"
	ActionEditGroup      GroupAction = "EDIT_GROUP"
	ActionDeleteGroup    GroupAction = "DELETE_GROUP"
	ActionInviteMembers  GroupAction = "INVITE_MEMBERS"
	ActionRemoveMembers  GroupAction = "REMOVE_MEMBERS"
	ActionManageRoles    GroupAction = "MANAGE_ROLES"
	ActionCreateTasks    GroupAction = "CREATE_TASKS"
	ActionEditTasks      GroupAction = "EDIT_TASKS"
	ActionDeleteTasks    GroupAction = "DELETE_TASKS"
	ActionViewTasks      GroupAction = "VIEW_TASKS"
	ActionViewSchedules  GroupAction = "VIEW_SCHEDULES"
)

// GroupActivity はグループ活動情報
type GroupActivity struct {
	TasksCreated    int `json:"tasks_created"`
	TasksCompleted  int `json:"tasks_completed"`
	SchedulesShared int `json:"schedules_shared"`
	ActiveMembers   int `json:"active_members"`
}

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
	GetGroupStats(ctx context.Context, groupID uuid.UUID) (*domain.GroupStats, error)
}

