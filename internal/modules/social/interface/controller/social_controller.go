package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/interface/dto"
	"github.com/hryt430/Yotei+/internal/modules/social/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type SocialController struct {
	socialService usecase.SocialService
	logger        logger.Logger
}

func NewSocialController(socialService usecase.SocialService, logger logger.Logger) *SocialController {
	return &SocialController{
		socialService: socialService,
		logger:        logger,
	}
}

// === 友達関係管理 ===

// SendFriendRequest は友達申請を送信する
// POST /api/v1/social/friends/request
func (sc *SocialController) SendFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.SendFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	addresseeID, err := uuid.Parse(req.AddresseeID)
	if err != nil {
		sc.logger.Error("Invalid addressee ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	friendship, err := sc.socialService.SendFriendRequest(c.Request.Context(), user.ID, addresseeID, req.Message)
	if err != nil {
		sc.logger.Error("Failed to send friend request",
			logger.Any("requesterID", user.ID),
			logger.Any("addresseeID", addresseeID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "send_friend_request_failed",
			Message: "友達申請の送信に失敗しました",
		})
		return
	}

	response := dto.ToFriendshipResponse(friendship)
	c.JSON(http.StatusCreated, response)
}

// AcceptFriendRequest は友達申請を承認する
// POST /api/v1/social/friends/accept
func (sc *SocialController) AcceptFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.AcceptFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	requesterID, err := uuid.Parse(req.RequesterID)
	if err != nil {
		sc.logger.Error("Invalid requester ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	friendship, err := sc.socialService.AcceptFriendRequest(c.Request.Context(), requesterID, user.ID)
	if err != nil {
		sc.logger.Error("Failed to accept friend request",
			logger.Any("requesterID", requesterID),
			logger.Any("addresseeID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "accept_friend_request_failed",
			Message: "友達申請の承認に失敗しました",
		})
		return
	}

	response := dto.ToFriendshipResponse(friendship)
	c.JSON(http.StatusOK, response)
}

// DeclineFriendRequest は友達申請を拒否する
// POST /api/v1/social/friends/decline
func (sc *SocialController) DeclineFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.DeclineFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	requesterID, err := uuid.Parse(req.RequesterID)
	if err != nil {
		sc.logger.Error("Invalid requester ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	err = sc.socialService.DeclineFriendRequest(c.Request.Context(), requesterID, user.ID)
	if err != nil {
		sc.logger.Error("Failed to decline friend request",
			logger.Any("requesterID", requesterID),
			logger.Any("addresseeID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "decline_friend_request_failed",
			Message: "友達申請の拒否に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "友達申請を拒否しました",
	})
}

// RemoveFriend は友達を削除する
// DELETE /api/v1/social/friends/{friendId}
func (sc *SocialController) RemoveFriend(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	friendIDStr := c.Param("friendId")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		sc.logger.Error("Invalid friend ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	err = sc.socialService.RemoveFriend(c.Request.Context(), user.ID, friendID)
	if err != nil {
		sc.logger.Error("Failed to remove friend",
			logger.Any("userID", user.ID),
			logger.Any("friendID", friendID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "remove_friend_failed",
			Message: "友達の削除に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "友達を削除しました",
	})
}

// BlockUser はユーザーをブロックする
// POST /api/v1/social/block
func (sc *SocialController) BlockUser(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		sc.logger.Error("Invalid target ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	err = sc.socialService.BlockUser(c.Request.Context(), user.ID, targetID)
	if err != nil {
		sc.logger.Error("Failed to block user",
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "block_user_failed",
			Message: "ユーザーのブロックに失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "ユーザーをブロックしました",
	})
}

// UnblockUser はブロックを解除する
// POST /api/v1/social/unblock
func (sc *SocialController) UnblockUser(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.UnblockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		sc.logger.Error("Invalid target ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	err = sc.socialService.UnblockUser(c.Request.Context(), user.ID, targetID)
	if err != nil {
		sc.logger.Error("Failed to unblock user",
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "unblock_user_failed",
			Message: "ブロック解除に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "ブロックを解除しました",
	})
}

// === 友達一覧・検索 ===

// GetFriends は友達一覧を取得する
// GET /api/v1/social/friends
func (sc *SocialController) GetFriends(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	friends, err := sc.socialService.GetFriends(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logger.Error("Failed to get friends",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_friends_failed",
			Message: "友達一覧の取得に失敗しました",
		})
		return
	}

	// TODO: 総数を取得する実装が必要
	total := len(friends)
	response := dto.ToFriendsListResponse(friends, total, pagination.Page, pagination.PageSize)
	c.JSON(http.StatusOK, response)
}

// GetPendingRequests は受信した友達申請を取得する
// GET /api/v1/social/friends/requests/received
func (sc *SocialController) GetPendingRequests(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	requests, err := sc.socialService.GetPendingRequests(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logger.Error("Failed to get pending requests",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_pending_requests_failed",
			Message: "友達申請一覧の取得に失敗しました",
		})
		return
	}

	// TODO: 総数を取得する実装が必要
	total := len(requests)
	response := dto.ToPendingRequestsResponse(requests, total, pagination.Page, pagination.PageSize)
	c.JSON(http.StatusOK, response)
}

// GetSentRequests は送信した友達申請を取得する
// GET /api/v1/social/friends/requests/sent
func (sc *SocialController) GetSentRequests(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	requests, err := sc.socialService.GetSentRequests(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logger.Error("Failed to get sent requests",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_sent_requests_failed",
			Message: "送信済み申請一覧の取得に失敗しました",
		})
		return
	}

	// TODO: 総数を取得する実装が必要
	total := len(requests)
	response := dto.ToPendingRequestsResponse(requests, total, pagination.Page, pagination.PageSize)
	c.JSON(http.StatusOK, response)
}

// === 招待管理 ===

// CreateInvitation は招待を作成する
// POST /api/v1/social/invitations
func (sc *SocialController) CreateInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	input := usecase.CreateInvitationInput{
		Type:         domain.InvitationType(req.Type),
		Method:       domain.InvitationMethod(req.Method),
		InviterID:    user.ID,
		Message:      req.Message,
		ExpiresHours: req.ExpiresHours,
		InviteeEmail: req.InviteeEmail,
	}

	if req.TargetID != nil {
		targetID, err := uuid.Parse(*req.TargetID)
		if err != nil {
			sc.logger.Error("Invalid target ID format", logger.Error(err))
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_target_id",
				Message: "無効なターゲットIDです",
			})
			return
		}
		input.TargetID = &targetID
	}

	invitation, err := sc.socialService.CreateInvitation(c.Request.Context(), input)
	if err != nil {
		sc.logger.Error("Failed to create invitation",
			logger.Any("inviterID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "create_invitation_failed",
			Message: "招待の作成に失敗しました",
		})
		return
	}

	response := dto.ToInvitationResponse(invitation)
	c.JSON(http.StatusCreated, response)
}

// AcceptInvitation は招待を受諾する
// POST /api/v1/social/invitations/accept
func (sc *SocialController) AcceptInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	result, err := sc.socialService.AcceptInvitation(c.Request.Context(), req.Code, user.ID)
	if err != nil {
		sc.logger.Error("Failed to accept invitation",
			logger.Any("userID", user.ID),
			logger.Any("code", req.Code),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "accept_invitation_failed",
			Message: "招待の受諾に失敗しました",
		})
		return
	}

	response := dto.ToInvitationResultResponse(result)
	c.JSON(http.StatusOK, response)
}

// GenerateInviteURL は招待URLを生成する
// GET /api/v1/social/invitations/{invitationId}/url
func (sc *SocialController) GenerateInviteURL(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	invitationIDStr := c.Param("invitationId")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		sc.logger.Error("Invalid invitation ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_invitation_id",
			Message: "無効な招待IDです",
		})
		return
	}

	url, err := sc.socialService.GenerateInviteURL(c.Request.Context(), invitationID)
	if err != nil {
		sc.logger.Error("Failed to generate invite URL",
			logger.Any("userID", user.ID),
			logger.Any("invitationID", invitationID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "generate_url_failed",
			Message: "招待URLの生成に失敗しました",
		})
		return
	}

	// 招待情報も取得してレスポンスに含める
	invitation, err := sc.socialService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		sc.logger.Error("Failed to get invitation details", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_invitation_failed",
			Message: "招待情報の取得に失敗しました",
		})
		return
	}

	response := dto.InviteURLResponse{
		URL:       url,
		Code:      invitation.Code,
		ExpiresAt: invitation.ExpiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetRelationship はユーザー間の関係を取得する
// GET /api/v1/social/relationship/{userId}
func (sc *SocialController) GetRelationship(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		sc.logger.Error("Invalid user ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	relationship, err := sc.socialService.GetRelationship(c.Request.Context(), user.ID, targetUserID)
	if err != nil {
		sc.logger.Error("Failed to get relationship",
			logger.Any("userID", user.ID),
			logger.Any("targetUserID", targetUserID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_relationship_failed",
			Message: "関係性の取得に失敗しました",
		})
		return
	}

	response := dto.ToUserRelationshipResponse(relationship)
	c.JSON(http.StatusOK, response)
}

// === ヘルパーメソッド ===

func (sc *SocialController) getPaginationFromQuery(c *gin.Context) commonDomain.Pagination {
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return commonDomain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}