package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// GroupType はグループの種類を表す
type GroupType string

const (
	GroupTypeProject  GroupType = "PROJECT"  // プロジェクト管理用
	GroupTypeSchedule GroupType = "SCHEDULE" // 予定共有用
)

// MemberRole はグループ内の権限を表す
type MemberRole string

const (
	RoleOwner  MemberRole = "OWNER"  // 所有者
	RoleAdmin  MemberRole = "ADMIN"  // 管理者
	RoleMember MemberRole = "MEMBER" // メンバー
)

// Group はグループ情報を表すドメインエンティティ
type Group struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        GroupType     `json:"type"`
	OwnerID     uuid.UUID     `json:"owner_id"`
	Settings    GroupSettings `json:"settings"`
	MemberCount int           `json:"member_count"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Version     int           `json:"version"` // 楽観的ロック用
}

// GroupSettings はグループの設定を表す
type GroupSettings struct {
	IsPublic            bool `json:"is_public"`            // 公開/非公開
	AllowMemberInvite   bool `json:"allow_member_invite"`  // メンバーの招待許可
	RequireApproval     bool `json:"require_approval"`     // 参加承認制
	EnableNotifications bool `json:"enable_notifications"` // 通知有効

	// 予定共有グループ用
	DefaultPrivacyLevel  PrivacyLevel `json:"default_privacy_level,omitempty"`
	AllowScheduleDetails bool         `json:"allow_schedule_details,omitempty"`

	// プロジェクトグループ用
	EnableGanttChart     bool `json:"enable_gantt_chart,omitempty"`
	EnableTaskDependency bool `json:"enable_task_dependency,omitempty"`
}

// PrivacyLevel は予定の公開レベル
type PrivacyLevel string

const (
	PrivacyLevelNone    PrivacyLevel = "NONE"    // 非表示
	PrivacyLevelBusy    PrivacyLevel = "BUSY"    // 予定ありのみ
	PrivacyLevelTitle   PrivacyLevel = "TITLE"   // タイトルまで
	PrivacyLevelDetails PrivacyLevel = "DETAILS" // 詳細まで
)

// NewGroup は新しいグループを作成する
func NewGroup(name, description string, groupType GroupType, ownerID uuid.UUID) *Group {
	now := time.Now()
	return &Group{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Type:        groupType,
		OwnerID:     ownerID,
		Settings:    getDefaultSettings(groupType),
		MemberCount: 1, // 作成者
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
}

// getDefaultSettings はグループタイプに応じたデフォルト設定を返す
func getDefaultSettings(groupType GroupType) GroupSettings {
	base := GroupSettings{
		IsPublic:            false,
		AllowMemberInvite:   true,
		RequireApproval:     true,
		EnableNotifications: true,
	}

	switch groupType {
	case GroupTypeSchedule:
		base.DefaultPrivacyLevel = PrivacyLevelBusy
		base.AllowScheduleDetails = false
	case GroupTypeProject:
		base.EnableGanttChart = true
		base.EnableTaskDependency = false
	}

	return base
}

// UpdateSettings はグループ設定を更新する
func (g *Group) UpdateSettings(settings GroupSettings) {
	g.Settings = settings
	g.UpdatedAt = time.Now()
	g.Version++
}

// AddMember はメンバー数を増加させる
func (g *Group) AddMember() {
	g.MemberCount++
	g.UpdatedAt = time.Now()
	g.Version++
}

// RemoveMember はメンバー数を減少させる
func (g *Group) RemoveMember() error {
	if g.MemberCount <= 1 {
		return errors.New("cannot remove the last member")
	}
	g.MemberCount--
	g.UpdatedAt = time.Now()
	g.Version++
	return nil
}

// GroupMember はグループメンバーシップを表す
type GroupMember struct {
	ID        uuid.UUID  `json:"id"`
	GroupID   uuid.UUID  `json:"group_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Role      MemberRole `json:"role"`
	JoinedAt  time.Time  `json:"joined_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NewGroupMember は新しいグループメンバーを作成する
func NewGroupMember(groupID, userID uuid.UUID, role MemberRole) *GroupMember {
	now := time.Now()
	return &GroupMember{
		ID:        uuid.New(),
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  now,
		UpdatedAt: now,
	}
}

// CanManageGroup はグループ管理権限があるかチェック
func (gm *GroupMember) CanManageGroup() bool {
	return gm.Role == RoleOwner || gm.Role == RoleAdmin
}

// CanInviteMembers はメンバー招待権限があるかチェック
func (gm *GroupMember) CanInviteMembers(groupSettings GroupSettings) bool {
	if gm.Role == RoleOwner || gm.Role == RoleAdmin {
		return true
	}
	return gm.Role == RoleMember && groupSettings.AllowMemberInvite
}

// PromoteToAdmin は管理者に昇格させる
func (gm *GroupMember) PromoteToAdmin() error {
	if gm.Role == RoleOwner {
		return errors.New("owner cannot be promoted")
	}
	gm.Role = RoleAdmin
	gm.UpdatedAt = time.Now()
	return nil
}

// DemoteToMember は一般メンバーに降格させる
func (gm *GroupMember) DemoteToMember() error {
	if gm.Role == RoleOwner {
		return errors.New("owner cannot be demoted")
	}
	gm.Role = RoleMember
	gm.UpdatedAt = time.Now()
	return nil
}

// TransferOwnership は所有権を移譲する
func (gm *GroupMember) TransferOwnership() {
	gm.Role = RoleOwner
	gm.UpdatedAt = time.Now()
}

// GroupStats はグループ統計情報
type GroupStats struct {
	MemberCount   int `json:"member_count"`
	TaskCount     int `json:"task_count,omitempty"`     // プロジェクトグループの場合
	ScheduleCount int `json:"schedule_count,omitempty"` // 予定共有グループの場合
	ActiveMembers int `json:"active_members"`           // 最近活動したメンバー数
}
