package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
	groupUsecase "github.com/hryt430/Yotei+/internal/modules/group/usecase"
)

// === リクエストDTO ===

type CreateGroupRequest struct {
	Name        string               `json:"name" binding:"required,max=100" example:"プロジェクトチーム"`
	Description string               `json:"description" binding:"max=500" example:"新製品開発プロジェクトのチーム"`
	Type        string               `json:"type" binding:"required" enums:"PROJECT,SCHEDULE" example:"PROJECT"`
	Settings    domain.GroupSettings `json:"settings"`
} // @name CreateGroupRequest

type UpdateGroupRequest struct {
	Name        *string               `json:"name,omitempty" binding:"omitempty,max=100" example:"プロジェクトチーム"`
	Description *string               `json:"description,omitempty" binding:"omitempty,max=500" example:"新製品開発プロジェクトのチーム"`
	Settings    *domain.GroupSettings `json:"settings,omitempty"`
} // @name UpdateGroupRequest

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Role   string `json:"role" enums:"OWNER,ADMIN,MEMBER" example:"MEMBER"`
} // @name AddMemberRequest

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required" enums:"OWNER,ADMIN,MEMBER" example:"ADMIN"`
} // @name UpdateMemberRoleRequest

// === レスポンスDTO ===

type GroupResponse struct {
	ID          uuid.UUID            `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string               `json:"name" example:"プロジェクトチーム"`
	Description string               `json:"description" example:"新製品開発プロジェクトのチーム"`
	Type        string               `json:"type" example:"PROJECT"`
	OwnerID     uuid.UUID            `json:"owner_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Settings    domain.GroupSettings `json:"settings"`
	MemberCount int                  `json:"member_count" example:"5"`
	CreatedAt   time.Time            `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time            `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Version     int                  `json:"version" example:"1"`
} // @name GroupResponse

type GroupWithMembersResponse struct {
	Group   GroupResponse            `json:"group"`
	Members []MemberWithUserResponse `json:"members"`
	MyRole  string                   `json:"my_role" example:"ADMIN"`
} // @name GroupWithMembersResponse

type MemberWithUserResponse struct {
	ID       uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	GroupID  uuid.UUID `json:"group_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID   uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Role     string    `json:"role" example:"MEMBER"`
	JoinedAt time.Time `json:"joined_at" example:"2024-01-01T00:00:00Z"`
	UserInfo *UserInfo `json:"user_info,omitempty"`
} // @name MemberWithUserResponse

type GroupListResponse struct {
	Groups     []GroupResponse `json:"groups"`
	Pagination PaginationInfo  `json:"pagination"`
} // @name GroupListResponse

type MemberListResponse struct {
	Members []MemberWithUserResponse `json:"members"`
} // @name MemberListResponse

type GroupStatsResponse struct {
	MemberCount   int `json:"member_count" example:"5"`
	TaskCount     int `json:"task_count,omitempty" example:"10"`
	ScheduleCount int `json:"schedule_count,omitempty" example:"3"`
	ActiveMembers int `json:"active_members" example:"4"`
} // @name GroupStatsResponse

type PaginationInfo struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
} // @name PaginationInfo

// UserInfo はユーザー基本情報
type UserInfo struct {
	ID       string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username string `json:"username" example:"user123"`
	Email    string `json:"email" example:"user@example.com"`
} // @name UserInfo

// === 変換関数 ===

func ToGroupResponse(group *domain.Group) *GroupResponse {
	return &GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Type:        string(group.Type),
		OwnerID:     group.OwnerID,
		Settings:    group.Settings,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
		Version:     group.Version,
	}
}

func ToGroupWithMembersResponse(groupWithMembers *groupUsecase.GroupWithMembers) *GroupWithMembersResponse {
	groupResp := ToGroupResponse(groupWithMembers.Group)

	members := make([]MemberWithUserResponse, len(groupWithMembers.Members))
	for i, member := range groupWithMembers.Members {
		var userInfo *UserInfo
		if member.UserInfo != nil {
			userInfo = &UserInfo{
				ID:       member.UserInfo.ID,
				Username: member.UserInfo.Username,
				Email:    member.UserInfo.Email,
			}
		}
		members[i] = MemberWithUserResponse{
			ID:       member.Member.ID,
			GroupID:  member.Member.GroupID,
			UserID:   member.Member.UserID,
			Role:     string(member.Member.Role),
			JoinedAt: member.Member.JoinedAt,
			UserInfo: userInfo,
		}
	}

	return &GroupWithMembersResponse{
		Group:   *groupResp,
		Members: members,
		MyRole:  string(groupWithMembers.MyRole),
	}
}

func ToGroupListResponse(groups []*domain.Group, total, page, pageSize int) *GroupListResponse {
	groupResponses := make([]GroupResponse, len(groups))
	for i, group := range groups {
		groupResponses[i] = *ToGroupResponse(group)
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &GroupListResponse{
		Groups: groupResponses,
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

func ToMemberListResponse(members []*groupUsecase.MemberWithUserInfo) *MemberListResponse {
	memberResponses := make([]MemberWithUserResponse, len(members))
	for i, member := range members {
		var userInfo *UserInfo
		if member.UserInfo != nil {
			userInfo = &UserInfo{
				ID:       member.UserInfo.ID,
				Username: member.UserInfo.Username,
				Email:    member.UserInfo.Email,
			}
		}
		memberResponses[i] = MemberWithUserResponse{
			ID:       member.Member.ID,
			GroupID:  member.Member.GroupID,
			UserID:   member.Member.UserID,
			Role:     string(member.Member.Role),
			JoinedAt: member.Member.JoinedAt,
			UserInfo: userInfo,
		}
	}

	return &MemberListResponse{
		Members: memberResponses,
	}
}

func ToGroupStatsResponse(stats *domain.GroupStats) *GroupStatsResponse {
	return &GroupStatsResponse{
		MemberCount:   stats.MemberCount,
		TaskCount:     stats.TaskCount,
		ScheduleCount: stats.ScheduleCount,
		ActiveMembers: stats.ActiveMembers,
	}
}

// === 共通レスポンス ===

// SuccessResponse は成功レスポンス構造体
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"操作が正常に完了しました"`
} // @name SuccessResponse

// ErrorResponse はエラーレスポンス構造体
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"リクエストが無効です"`
} // @name ErrorResponse
