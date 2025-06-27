package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	socialUsecase "github.com/hryt430/Yotei+/internal/modules/social/usecase"
)

// === リクエストDTO ===

type SendFriendRequestRequest struct {
	AddresseeID string `json:"addressee_id" binding:"required"`
	Message     string `json:"message" binding:"max=500"`
}

type AcceptFriendRequestRequest struct {
	RequesterID string `json:"requester_id" binding:"required"`
}

type DeclineFriendRequestRequest struct {
	RequesterID string `json:"requester_id" binding:"required"`
}

type BlockUserRequest struct {
	TargetID string `json:"target_id" binding:"required"`
}

type UnblockUserRequest struct {
	TargetID string `json:"target_id" binding:"required"`
}

type CreateInvitationRequest struct {
	Type         string  `json:"type" binding:"required,oneof=FRIEND GROUP"`
	Method       string  `json:"method" binding:"required,oneof=IN_APP CODE URL"`
	Message      string  `json:"message" binding:"max=500"`
	ExpiresHours int     `json:"expires_hours" binding:"min=1,max=168"` // 1-168時間（1週間）
	InviteeEmail *string `json:"invitee_email,omitempty" binding:"omitempty,email"`
	TargetID     *string `json:"target_id,omitempty"` // Group IDなど
}

type AcceptInvitationRequest struct {
	Code string `json:"code" binding:"required"`
}

type DeclineInvitationRequest struct {
	InvitationID string `json:"invitation_id" binding:"required"`
}

// === レスポンスDTO ===

type FriendshipResponse struct {
	ID          uuid.UUID  `json:"id"`
	RequesterID uuid.UUID  `json:"requester_id"`
	AddresseeID uuid.UUID  `json:"addressee_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	BlockedAt   *time.Time `json:"blocked_at,omitempty"`
}

type FriendWithUserInfoResponse struct {
	Friendship FriendshipResponse `json:"friendship"`
	UserInfo   *UserInfo          `json:"user_info,omitempty"`
}

type FriendshipWithUserInfoResponse struct {
	Friendship FriendshipResponse `json:"friendship"`
	UserInfo   *UserInfo          `json:"user_info,omitempty"`
}

type InvitationResponse struct {
	ID          uuid.UUID           `json:"id"`
	Type        string              `json:"type"`
	Method      string              `json:"method"`
	Status      string              `json:"status"`
	InviterID   uuid.UUID           `json:"inviter_id"`
	InviteeID   *uuid.UUID          `json:"invitee_id,omitempty"`
	InviteeInfo *domain.InviteeInfo `json:"invitee_info,omitempty"`
	TargetID    *uuid.UUID          `json:"target_id,omitempty"`
	Code        string              `json:"code,omitempty"`
	URL         string              `json:"url,omitempty"`
	Message     string              `json:"message"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	ExpiresAt   time.Time           `json:"expires_at"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	AcceptedAt  *time.Time          `json:"accepted_at,omitempty"`
}

type InvitationResultResponse struct {
	Success    bool                `json:"success"`
	Message    string              `json:"message"`
	Friendship *FriendshipResponse `json:"friendship,omitempty"`
	GroupID    *uuid.UUID          `json:"group_id,omitempty"`
}

type UserRelationshipResponse struct {
	IsFriend        bool `json:"is_friend"`
	IsBlocked       bool `json:"is_blocked"`
	RequestSent     bool `json:"request_sent"`
	RequestReceived bool `json:"request_received"`
}

type FriendsListResponse struct {
	Friends    []FriendWithUserInfoResponse `json:"friends"`
	Pagination PaginationInfo               `json:"pagination"`
}

type PendingRequestsResponse struct {
	Requests   []FriendshipWithUserInfoResponse `json:"requests"`
	Pagination PaginationInfo                   `json:"pagination"`
}

type SentRequestsResponse struct {
	Requests   []FriendshipWithUserInfoResponse `json:"requests"`
	Pagination PaginationInfo                   `json:"pagination"`
}

type InvitationsListResponse struct {
	Invitations []InvitationResponse `json:"invitations"`
	Pagination  PaginationInfo       `json:"pagination"`
}

type InviteURLResponse struct {
	URL       string    `json:"url"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PaginationInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// UserInfo はユーザー基本情報
type UserInfo struct {
	ID       string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username string `json:"username" example:"user123"`
	Email    string `json:"email" example:"user@example.com"`
} // @name UserInfo

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

func ToFriendshipResponse(friendship *domain.Friendship) *FriendshipResponse {
	return &FriendshipResponse{
		ID:          friendship.ID,
		RequesterID: friendship.RequesterID,
		AddresseeID: friendship.AddresseeID,
		Status:      string(friendship.Status),
		CreatedAt:   friendship.CreatedAt,
		UpdatedAt:   friendship.UpdatedAt,
		AcceptedAt:  friendship.AcceptedAt,
		BlockedAt:   friendship.BlockedAt,
	}
}

func ToFriendWithUserInfoResponse(friend *socialUsecase.FriendWithUserInfo) *FriendWithUserInfoResponse {
	var userInfo *UserInfo
	if friend.UserInfo != nil {
		userInfo = &UserInfo{
			ID:       friend.UserInfo.ID,
			Username: friend.UserInfo.Username,
			Email:    friend.UserInfo.Email,
		}
	}
	return &FriendWithUserInfoResponse{
		Friendship: *ToFriendshipResponse(friend.Friendship),
		UserInfo:   userInfo,
	}
}

func ToFriendshipWithUserInfoResponse(friendship *socialUsecase.FriendshipWithUserInfo) *FriendshipWithUserInfoResponse {
	var userInfo *UserInfo
	if friendship.UserInfo != nil {
		userInfo = &UserInfo{
			ID:       friendship.UserInfo.ID,
			Username: friendship.UserInfo.Username,
			Email:    friendship.UserInfo.Email,
		}
	}
	return &FriendshipWithUserInfoResponse{
		Friendship: *ToFriendshipResponse(friendship.Friendship),
		UserInfo:   userInfo,
	}
}

func ToInvitationResponse(invitation *domain.Invitation) *InvitationResponse {
	return &InvitationResponse{
		ID:          invitation.ID,
		Type:        string(invitation.Type),
		Method:      string(invitation.Method),
		Status:      string(invitation.Status),
		InviterID:   invitation.InviterID,
		InviteeID:   invitation.InviteeID,
		InviteeInfo: invitation.InviteeInfo,
		TargetID:    invitation.TargetID,
		Code:        invitation.Code,
		URL:         invitation.URL,
		Message:     invitation.Message,
		Metadata:    invitation.Metadata,
		ExpiresAt:   invitation.ExpiresAt,
		CreatedAt:   invitation.CreatedAt,
		UpdatedAt:   invitation.UpdatedAt,
		AcceptedAt:  invitation.AcceptedAt,
	}
}

func ToInvitationResultResponse(result *socialUsecase.InvitationResult) *InvitationResultResponse {
	response := &InvitationResultResponse{
		Success: result.Success,
		Message: result.Message,
	}

	if result.Friendship != nil {
		response.Friendship = ToFriendshipResponse(result.Friendship)
	}

	if result.GroupID != nil {
		response.GroupID = result.GroupID
	}

	return response
}

func ToUserRelationshipResponse(relationship *socialUsecase.UserRelationship) *UserRelationshipResponse {
	return &UserRelationshipResponse{
		IsFriend:        relationship.IsFriend,
		IsBlocked:       relationship.IsBlocked,
		RequestSent:     relationship.RequestSent,
		RequestReceived: relationship.RequestReceived,
	}
}

func ToFriendsListResponse(friends []*socialUsecase.FriendWithUserInfo, total, page, pageSize int) *FriendsListResponse {
	friendResponses := make([]FriendWithUserInfoResponse, len(friends))
	for i, friend := range friends {
		friendResponses[i] = *ToFriendWithUserInfoResponse(friend)
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &FriendsListResponse{
		Friends: friendResponses,
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

func ToPendingRequestsResponse(requests []*socialUsecase.FriendshipWithUserInfo, total, page, pageSize int) *PendingRequestsResponse {
	requestResponses := make([]FriendshipWithUserInfoResponse, len(requests))
	for i, request := range requests {
		requestResponses[i] = *ToFriendshipWithUserInfoResponse(request)
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &PendingRequestsResponse{
		Requests: requestResponses,
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

func ToInvitationsListResponse(invitations []*domain.Invitation, total, page, pageSize int) *InvitationsListResponse {
	invitationResponses := make([]InvitationResponse, len(invitations))
	for i, invitation := range invitations {
		invitationResponses[i] = *ToInvitationResponse(invitation)
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &InvitationsListResponse{
		Invitations: invitationResponses,
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
