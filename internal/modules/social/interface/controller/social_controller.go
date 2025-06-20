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
	"go.uber.org/zap/zapcore"
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
// POST /api/v1/social/friends/requests
func (sc *SocialController) SendFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.SendFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logError("bind JSON", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	addresseeID, err := sc.validateUUID(req.AddresseeID, "addressee ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	// 自分自身への申請をチェック
	if user.ID == addresseeID {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "self_request_not_allowed",
			Message: "自分自身に友達申請はできません",
		})
		return
	}

	friendship, err := sc.socialService.SendFriendRequest(c.Request.Context(), user.ID, addresseeID, req.Message)
	if err != nil {
		sc.logError("send friend request", err,
			logger.Any("requesterID", user.ID),
			logger.Any("addresseeID", addresseeID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "send_friend_request_failed",
			Message: "友達申請の送信に失敗しました",
		})
		return
	}

	sc.logger.Info("Friend request sent successfully",
		logger.Any("requesterID", user.ID),
		logger.Any("addresseeID", addresseeID))

	response := dto.ToFriendshipResponse(friendship)
	c.JSON(http.StatusCreated, response)
}

// AcceptFriendRequest は友達申請を承認する
// PUT /api/v1/social/friends/requests/{friendshipId}/accept
func (sc *SocialController) AcceptFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	friendshipID, err := sc.validateUUID(c.Param("friendshipId"), "friendship ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_friendship_id",
			Message: "無効な友達申請IDです",
		})
		return
	}

	friendship, err := sc.socialService.AcceptFriendRequest(c.Request.Context(), friendshipID, user.ID)
	if err != nil {
		sc.logError("accept friend request", err,
			logger.Any("friendshipID", friendshipID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "accept_friend_request_failed",
			Message: "友達申請の承認に失敗しました",
		})
		return
	}

	sc.logger.Info("Friend request accepted successfully",
		logger.Any("friendshipID", friendshipID),
		logger.Any("userID", user.ID))

	response := dto.ToFriendshipResponse(friendship)
	c.JSON(http.StatusOK, response)
}

// DeclineFriendRequest は友達申請を拒否する
// PUT /api/v1/social/friends/requests/{friendshipId}/decline
func (sc *SocialController) DeclineFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	friendshipID, err := sc.validateUUID(c.Param("friendshipId"), "friendship ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_friendship_id",
			Message: "無効な友達申請IDです",
		})
		return
	}

	err = sc.socialService.DeclineFriendRequest(c.Request.Context(), friendshipID, user.ID)
	if err != nil {
		sc.logError("decline friend request", err,
			logger.Any("friendshipID", friendshipID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "decline_friend_request_failed",
			Message: "友達申請の拒否に失敗しました",
		})
		return
	}

	sc.logger.Info("Friend request declined successfully",
		logger.Any("friendshipID", friendshipID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "友達申請を拒否しました",
	})
}

// RemoveFriend は友達を削除する
// DELETE /api/v1/social/friends/{userId}
func (sc *SocialController) RemoveFriend(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	friendID, err := sc.validateUUID(c.Param("userId"), "friend ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	err = sc.socialService.RemoveFriend(c.Request.Context(), user.ID, friendID)
	if err != nil {
		sc.logError("remove friend", err,
			logger.Any("userID", user.ID),
			logger.Any("friendID", friendID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "remove_friend_failed",
			Message: "友達の削除に失敗しました",
		})
		return
	}

	sc.logger.Info("Friend removed successfully",
		logger.Any("userID", user.ID),
		logger.Any("friendID", friendID))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "友達を削除しました",
	})
}

// === ブロック機能 ===

// BlockUser はユーザーをブロックする
// POST /api/v1/social/users/{userId}/block
func (sc *SocialController) BlockUser(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetID, err := sc.validateUUID(c.Param("userId"), "target ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	err = sc.socialService.BlockUser(c.Request.Context(), user.ID, targetID)
	if err != nil {
		sc.logError("block user", err,
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "block_user_failed",
			Message: "ユーザーのブロックに失敗しました",
		})
		return
	}

	sc.logger.Info("User blocked successfully",
		logger.Any("userID", user.ID),
		logger.Any("targetID", targetID))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "ユーザーをブロックしました",
	})
}

// UnblockUser はブロックを解除する
// DELETE /api/v1/social/users/{userId}/block
func (sc *SocialController) UnblockUser(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetID, err := sc.validateUUID(c.Param("userId"), "target ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	err = sc.socialService.UnblockUser(c.Request.Context(), user.ID, targetID)
	if err != nil {
		sc.logError("unblock user", err,
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "unblock_user_failed",
			Message: "ブロック解除に失敗しました",
		})
		return
	}

	sc.logger.Info("User unblocked successfully",
		logger.Any("userID", user.ID),
		logger.Any("targetID", targetID))

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
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	friends, err := sc.socialService.GetFriends(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logError("get friends", err, logger.Any("userID", user.ID))
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

// GetMutualFriends は共通の友達を取得する
// GET /api/v1/social/friends/{userId}/mutual
func (sc *SocialController) GetMutualFriends(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetID, err := sc.validateUUID(c.Param("userId"), "target ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	mutualFriends, err := sc.socialService.GetMutualFriends(c.Request.Context(), user.ID, targetID)
	if err != nil {
		sc.logError("get mutual friends", err,
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_mutual_friends_failed",
			Message: "共通の友達の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": mutualFriends,
	})
}

// GetPendingRequests は受信した友達申請を取得する
// GET /api/v1/social/friends/requests/received
func (sc *SocialController) GetPendingRequests(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	requests, err := sc.socialService.GetPendingRequests(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logError("get pending requests", err, logger.Any("userID", user.ID))
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
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	requests, err := sc.socialService.GetSentRequests(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logError("get sent requests", err, logger.Any("userID", user.ID))
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
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logError("bind JSON", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
		return
	}

	// 招待タイプのバリデーション
	invitationType := domain.InvitationType(req.Type)
	if invitationType != domain.InvitationTypeFriend && invitationType != domain.InvitationTypeGroup {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_invitation_type",
			Message: "無効な招待タイプです",
		})
		return
	}

	// 招待方法のバリデーション
	invitationMethod := domain.InvitationMethod(req.Method)
	if invitationMethod != domain.MethodInApp && invitationMethod != domain.MethodCode && invitationMethod != domain.MethodURL {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_invitation_method",
			Message: "無効な招待方法です",
		})
		return
	}

	input := usecase.CreateInvitationInput{
		Type:         invitationType,
		Method:       invitationMethod,
		InviterID:    user.ID,
		Message:      req.Message,
		ExpiresHours: req.ExpiresHours,
		InviteeEmail: req.InviteeEmail,
	}

	if req.TargetID != nil {
		targetID, err := sc.validateUUID(*req.TargetID, "target ID")
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_target_id",
				Message: "無効なターゲットIDです",
			})
			return
		}
		input.TargetID = &targetID
	}

	// グループ招待の場合、TargetIDが必要
	if invitationType == domain.InvitationTypeGroup && input.TargetID == nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "target_id_required",
			Message: "グループ招待にはグループIDが必要です",
		})
		return
	}

	// デフォルト値設定
	if input.ExpiresHours == 0 {
		input.ExpiresHours = 168 // 1週間
	}

	invitation, err := sc.socialService.CreateInvitation(c.Request.Context(), input)
	if err != nil {
		sc.logError("create invitation", err, logger.Any("inviterID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "create_invitation_failed",
			Message: "招待の作成に失敗しました",
		})
		return
	}

	sc.logger.Info("Invitation created successfully",
		logger.Any("inviterID", user.ID),
		logger.Any("invitationID", invitation.ID))

	response := dto.ToInvitationResponse(invitation)
	c.JSON(http.StatusCreated, response)
}

// GetInvitation は招待詳細を取得する
// GET /api/v1/social/invitations/{invitationId}
func (sc *SocialController) GetInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	invitationID, err := sc.validateUUID(c.Param("invitationId"), "invitation ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_invitation_id",
			Message: "無効な招待IDです",
		})
		return
	}

	invitation, err := sc.socialService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		sc.logError("get invitation", err, logger.Any("invitationID", invitationID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_invitation_failed",
			Message: "招待情報の取得に失敗しました",
		})
		return
	}

	if invitation == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "invitation_not_found",
			Message: "招待が見つかりません",
		})
		return
	}

	// 権限チェック（招待者または被招待者のみ閲覧可能）
	if invitation.InviterID != user.ID &&
		(invitation.InviteeID == nil || *invitation.InviteeID != user.ID) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "access_denied",
			Message: "この招待を閲覧する権限がありません",
		})
		return
	}

	response := dto.ToInvitationResponse(invitation)
	c.JSON(http.StatusOK, response)
}

// GetInvitationByCode は招待コードから招待情報を取得する
// GET /api/v1/social/invitations/code/{code}
func (sc *SocialController) GetInvitationByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "code_required",
			Message: "招待コードが必要です",
		})
		return
	}

	invitation, err := sc.socialService.GetInvitationByCode(c.Request.Context(), code)
	if err != nil {
		sc.logError("get invitation by code", err, logger.Any("code", code))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_invitation_failed",
			Message: "招待情報の取得に失敗しました",
		})
		return
	}

	if invitation == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "invitation_not_found",
			Message: "有効な招待が見つかりません",
		})
		return
	}

	// 期限切れチェック
	if invitation.IsExpired() {
		c.JSON(http.StatusGone, dto.ErrorResponse{
			Error:   "invitation_expired",
			Message: "招待の有効期限が切れています",
		})
		return
	}

	// プライベート情報を除外した公開情報のみ返す
	publicInvitation := struct {
		ID        uuid.UUID               `json:"id"`
		Type      domain.InvitationType   `json:"type"`
		Method    domain.InvitationMethod `json:"method"`
		Status    domain.InvitationStatus `json:"status"`
		Message   string                  `json:"message"`
		ExpiresAt string                  `json:"expires_at"`
		CreatedAt string                  `json:"created_at"`
	}{
		ID:        invitation.ID,
		Type:      invitation.Type,
		Method:    invitation.Method,
		Status:    invitation.Status,
		Message:   invitation.Message,
		ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt: invitation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, gin.H{
		"data": publicInvitation,
	})
}

// AcceptInvitation は招待を受諾する
// POST /api/v1/social/invitations/{code}/accept
func (sc *SocialController) AcceptInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "code_required",
			Message: "招待コードが必要です",
		})
		return
	}

	result, err := sc.socialService.AcceptInvitation(c.Request.Context(), code, user.ID)
	if err != nil {
		sc.logError("accept invitation", err,
			logger.Any("code", code),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "accept_invitation_failed",
			Message: "招待の受諾に失敗しました",
		})
		return
	}

	sc.logger.Info("Invitation accepted successfully",
		logger.Any("code", code),
		logger.Any("userID", user.ID))

	response := dto.ToInvitationResultResponse(result)
	c.JSON(http.StatusOK, response)
}

// DeclineInvitation は招待を拒否する
// PUT /api/v1/social/invitations/{invitationId}/decline
func (sc *SocialController) DeclineInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	invitationID, err := sc.validateUUID(c.Param("invitationId"), "invitation ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_invitation_id",
			Message: "無効な招待IDです",
		})
		return
	}

	err = sc.socialService.DeclineInvitation(c.Request.Context(), invitationID, user.ID)
	if err != nil {
		sc.logError("decline invitation", err,
			logger.Any("invitationID", invitationID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "decline_invitation_failed",
			Message: "招待の拒否に失敗しました",
		})
		return
	}

	sc.logger.Info("Invitation declined successfully",
		logger.Any("invitationID", invitationID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "招待を拒否しました",
	})
}

// CancelInvitation は招待をキャンセルする
// DELETE /api/v1/social/invitations/{invitationId}
func (sc *SocialController) CancelInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	invitationID, err := sc.validateUUID(c.Param("invitationId"), "invitation ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_invitation_id",
			Message: "無効な招待IDです",
		})
		return
	}

	err = sc.socialService.CancelInvitation(c.Request.Context(), invitationID, user.ID)
	if err != nil {
		sc.logError("cancel invitation", err,
			logger.Any("invitationID", invitationID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "cancel_invitation_failed",
			Message: "招待のキャンセルに失敗しました",
		})
		return
	}

	sc.logger.Info("Invitation cancelled successfully",
		logger.Any("invitationID", invitationID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "招待をキャンセルしました",
	})
}

// GetSentInvitations は送信した招待一覧を取得する
// GET /api/v1/social/invitations/sent
func (sc *SocialController) GetSentInvitations(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	invitations, err := sc.socialService.GetSentInvitations(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logError("get sent invitations", err, logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_sent_invitations_failed",
			Message: "送信済み招待の取得に失敗しました",
		})
		return
	}

	// TODO: 総数を取得する実装が必要
	total := len(invitations)
	response := dto.ToInvitationsListResponse(invitations, total, pagination.Page, pagination.PageSize)
	c.JSON(http.StatusOK, response)
}

// GetReceivedInvitations は受信した招待一覧を取得する
// GET /api/v1/social/invitations/received
func (sc *SocialController) GetReceivedInvitations(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	pagination := sc.getPaginationFromQuery(c)
	invitations, err := sc.socialService.GetReceivedInvitations(c.Request.Context(), user.ID, pagination)
	if err != nil {
		sc.logError("get received invitations", err, logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "get_received_invitations_failed",
			Message: "受信済み招待の取得に失敗しました",
		})
		return
	}

	// TODO: 総数を取得する実装が必要
	total := len(invitations)
	response := dto.ToInvitationsListResponse(invitations, total, pagination.Page, pagination.PageSize)
	c.JSON(http.StatusOK, response)
}

// GenerateInviteURL は招待URLを生成する
// GET /api/v1/social/invitations/{invitationId}/url
func (sc *SocialController) GenerateInviteURL(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	invitationID, err := sc.validateUUID(c.Param("invitationId"), "invitation ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_invitation_id",
			Message: "無効な招待IDです",
		})
		return
	}

	url, err := sc.socialService.GenerateInviteURL(c.Request.Context(), invitationID)
	if err != nil {
		sc.logError("generate invite URL", err,
			logger.Any("userID", user.ID),
			logger.Any("invitationID", invitationID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "generate_url_failed",
			Message: "招待URLの生成に失敗しました",
		})
		return
	}

	// 招待情報も取得してレスポンスに含める
	invitation, err := sc.socialService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		sc.logError("get invitation details", err)
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
// GET /api/v1/social/relationships/{userId}
func (sc *SocialController) GetRelationship(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		sc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
		return
	}

	targetUserID, err := sc.validateUUID(c.Param("userId"), "user ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "無効なユーザーIDです",
		})
		return
	}

	relationship, err := sc.socialService.GetRelationship(c.Request.Context(), user.ID, targetUserID)
	if err != nil {
		sc.logError("get relationship", err,
			logger.Any("userID", user.ID),
			logger.Any("targetUserID", targetUserID))
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

func (sc *SocialController) validateUUID(id string, fieldName string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		sc.logger.Error("Invalid UUID format",
			logger.String("field", fieldName),
			logger.String("value", id),
			logger.Error(err))
		return uuid.Nil, err
	}
	return parsedID, nil
}

func (sc *SocialController) logError(operation string, err error, fields ...zapcore.Field) {
	sc.logger.Error("Operation failed",
		append([]zapcore.Field{
			logger.String("operation", operation),
			logger.Error(err),
		}, fields...)...)
}

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
