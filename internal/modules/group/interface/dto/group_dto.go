package dto

import (
	"time"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
	groupUsecase "github.com/hryt430/Yotei+/internal/modules/group/usecase"
)

// === リクエストDTO ===

type CreateGroupRequest struct {
	Name        string               `json:"name" binding:"required,max=100"`
	Description string               `json:"description" binding:"max=500"`
	Type        string               `json:"type" binding:"required"`
	Settings    domain.GroupSettings `json:"settings"`
}

type UpdateGroupRequest struct {
	Name        *string               `json:"name,omitempty" binding:"omitempty,max=100"`
	Description *string               `json:"description,omitempty" binding:"omitempty,max=500"`
	Settings    *domain.GroupSettings `json:"settings,omitempty"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// === レスポンスDTO ===

type GroupResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Type        string               `json:"type"`
	OwnerID     uuid.UUID            `json:"owner_id"`
	Settings    domain.GroupSettings `json:"settings"`
	MemberCount int                  `json:"member_count"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Version     int                  `json:"version"`
}

type GroupWithMembersResponse struct {
	Group   GroupResponse            `json:"group"`
	Members []MemberWithUserResponse `json:"members"`
	MyRole  string                   `json:"my_role"`
}

type MemberWithUserResponse struct {
	ID       uuid.UUID              `json:"id"`
	GroupID  uuid.UUID              `json:"group_id"`
	UserID   uuid.UUID              `json:"user_id"`
	Role     string                 `json:"role"`
	JoinedAt time.Time              `json:"joined_at"`
	UserInfo *commonDomain.UserInfo `json:"user_info,omitempty"`
}

type GroupListResponse struct {
	Groups     []GroupResponse `json:"groups"`
	Pagination PaginationInfo  `json:"pagination"`
}

type MemberListResponse struct {
	Members []MemberWithUserResponse `json:"members"`
}

type GroupStatsResponse struct {
	MemberCount   int `json:"member_count"`
	TaskCount     int `json:"task_count,omitempty"`
	ScheduleCount int `json:"schedule_count,omitempty"`
	ActiveMembers int `json:"active_members"`
}

type PaginationInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// === 共通レスポンス ===

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

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
		members[i] = MemberWithUserResponse{
			ID:       member.Member.ID,
			GroupID:  member.Member.GroupID,
			UserID:   member.Member.UserID,
			Role:     string(member.Member.Role),
			JoinedAt: member.Member.JoinedAt,
			UserInfo: member.UserInfo,
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
		memberResponses[i] = MemberWithUserResponse{
			ID:       member.Member.ID,
			GroupID:  member.Member.GroupID,
			UserID:   member.Member.UserID,
			Role:     string(member.Member.Role),
			JoinedAt: member.Member.JoinedAt,
			UserInfo: member.UserInfo,
		}
	}

	return &MemberListResponse{
		Members: memberResponses,
	}
}

func ToGroupStatsResponse(stats *groupUsecase.GroupStats) *GroupStatsResponse {
	return &GroupStatsResponse{
		MemberCount:   stats.MemberCount,
		TaskCount:     stats.TaskCount,
		ScheduleCount: stats.ScheduleCount,
		ActiveMembers: stats.ActiveMembers,
	}
}
