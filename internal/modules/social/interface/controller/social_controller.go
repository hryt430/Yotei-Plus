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

// SendFriendRequestRequest は友達申請送信のリクエスト構造体
type SendFriendRequestRequest struct {
	AddresseeID string `json:"addressee_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Message     string `json:"message" binding:"max=500" example:"友達になりませんか？"`
} // @name SendFriendRequestRequest

// CreateInvitationRequest は招待作成のリクエスト構造体
type CreateInvitationRequest struct {
	Type         string  `json:"type" binding:"required" enums:"FRIEND,GROUP" example:"FRIEND"`
	Method       string  `json:"method" binding:"required" enums:"IN_APP,CODE,URL" example:"CODE"`
	Message      string  `json:"message" binding:"max=500" example:"一緒にYotei+を使いませんか？"`
	ExpiresHours int     `json:"expires_hours" binding:"min=1,max=168" example:"168"`
	InviteeEmail *string `json:"invitee_email,omitempty" binding:"omitempty,email" example:"friend@example.com"`
	TargetID     *string `json:"target_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
} // @name CreateInvitationRequest

// FriendshipResponse は友達関係のレスポンス構造体
type FriendshipResponse struct {
	ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	RequesterID uuid.UUID `json:"requester_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AddresseeID uuid.UUID `json:"addressee_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status      string    `json:"status" example:"PENDING"`
	CreatedAt   string    `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   string    `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	AcceptedAt  *string   `json:"accepted_at,omitempty" example:"2024-01-01T01:00:00Z"`
	BlockedAt   *string   `json:"blocked_at,omitempty"`
} // @name FriendshipResponse

// UserInfo はユーザー基本情報
type UserInfo struct {
	ID       string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username string `json:"username" example:"user123"`
	Email    string `json:"email" example:"user@example.com"`
} // @name UserInfo

// FriendWithUserInfoResponse はユーザー情報付き友達レスポンス
type FriendWithUserInfoResponse struct {
	Friendship FriendshipResponse `json:"friendship"`
	UserInfo   *UserInfo          `json:"user_info,omitempty"`
} // @name FriendWithUserInfoResponse

// InvitationResponse は招待のレスポンス構造体
type InvitationResponse struct {
	ID          uuid.UUID           `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type        string              `json:"type" example:"FRIEND"`
	Method      string              `json:"method" example:"CODE"`
	Status      string              `json:"status" example:"PENDING"`
	InviterID   uuid.UUID           `json:"inviter_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	InviteeID   *uuid.UUID          `json:"invitee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	InviteeInfo *domain.InviteeInfo `json:"invitee_info,omitempty"`
	TargetID    *uuid.UUID          `json:"target_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Code        string              `json:"code,omitempty" example:"abc123def456"`
	URL         string              `json:"url,omitempty" example:"https://yotei-plus.com/invite/abc123def456"`
	Message     string              `json:"message" example:"一緒にYotei+を使いませんか？"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	ExpiresAt   string              `json:"expires_at" example:"2024-01-08T00:00:00Z"`
	CreatedAt   string              `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   string              `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	AcceptedAt  *string             `json:"accepted_at,omitempty" example:"2024-01-01T01:00:00Z"`
} // @name InvitationResponse

// InvitationResultResponse は招待受諾結果のレスポンス構造体
type InvitationResultResponse struct {
	Success    bool                `json:"success" example:"true"`
	Message    string              `json:"message" example:"招待を受諾しました"`
	Friendship *FriendshipResponse `json:"friendship,omitempty"`
	GroupID    *uuid.UUID          `json:"group_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
} // @name InvitationResultResponse

// UserRelationshipResponse はユーザー間関係のレスポンス構造体
type UserRelationshipResponse struct {
	IsFriend        bool `json:"is_friend" example:"true"`
	IsBlocked       bool `json:"is_blocked" example:"false"`
	RequestSent     bool `json:"request_sent" example:"false"`
	RequestReceived bool `json:"request_received" example:"false"`
} // @name UserRelationshipResponse

// FriendsListResponse は友達一覧のレスポンス構造体
type FriendsListResponse struct {
	Friends    []FriendWithUserInfoResponse `json:"friends"`
	Pagination PaginationInfo               `json:"pagination"`
} // @name FriendsListResponse

// InviteURLResponse は招待URL生成のレスポンス構造体
type InviteURLResponse struct {
	URL       string `json:"url" example:"https://yotei-plus.com/invite/abc123def456"`
	Code      string `json:"code" example:"abc123def456"`
	ExpiresAt string `json:"expires_at" example:"2024-01-08T00:00:00Z"`
} // @name InviteURLResponse

// PaginationInfo はページング情報
type PaginationInfo struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"20"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"5"`
} // @name PaginationInfo

// SuccessResponse は成功レスポンス
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"操作が正常に完了しました"`
} // @name SuccessResponse

// ErrorResponse はエラーレスポンス
type ErrorResponse struct {
	Error   string `json:"error" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"リクエストが無効です"`
} // @name ErrorResponse

// === 友達関係管理 ===

// SendFriendRequest 友達申請送信
// @Summary      友達申請送信
// @Description  指定されたユーザーに友達申請を送信します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        request body SendFriendRequestRequest true "友達申請情報"
// @Security     BearerAuth
// @Success      201 {object} FriendshipResponse "友達申請送信成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効（自分自身への申請など）"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "ユーザーが見つからない"
// @Failure      409 {object} ErrorResponse "既に友達または申請済み"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends/requests [post]
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

// AcceptFriendRequest 友達申請承認
// @Summary      友達申請承認
// @Description  受信した友達申請を承認します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        friendshipId path string true "友達申請ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} FriendshipResponse "友達申請承認成功"
// @Failure      400 {object} ErrorResponse "友達申請IDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "この申請を承認する権限がない"
// @Failure      404 {object} ErrorResponse "友達申請が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends/requests/{friendshipId}/accept [put]
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

// DeclineFriendRequest 友達申請拒否
// @Summary      友達申請拒否
// @Description  受信した友達申請を拒否します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        friendshipId path string true "友達申請ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "友達申請拒否成功"
// @Failure      400 {object} ErrorResponse "友達申請IDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "この申請を拒否する権限がない"
// @Failure      404 {object} ErrorResponse "友達申請が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends/requests/{friendshipId}/decline [put]
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

// RemoveFriend 友達削除
// @Summary      友達削除
// @Description  友達関係を解除します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        userId path string true "削除する友達のユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "友達削除成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "友達関係が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends/{userId} [delete]
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

// BlockUser ユーザーブロック
// @Summary      ユーザーブロック
// @Description  指定されたユーザーをブロックします
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        userId path string true "ブロックするユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "ユーザーブロック成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "ユーザーが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/users/{userId}/block [post]
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

// UnblockUser ブロック解除
// @Summary      ブロック解除
// @Description  指定されたユーザーのブロックを解除します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        userId path string true "ブロック解除するユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "ブロック解除成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "ブロック関係が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/users/{userId}/block [delete]
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

// GetFriends 友達一覧取得
// @Summary      友達一覧取得
// @Description  自分の友達一覧を取得します（ページング対応）
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(20) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} FriendsListResponse "友達一覧取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends [get]
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

// GetPendingRequests 受信した友達申請取得
// @Summary      受信した友達申請取得
// @Description  自分宛の友達申請一覧を取得します（ページング対応）
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(20) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} dto.PendingRequestsResponse "友達申請一覧取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends/requests/received [get]
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

// GetSentRequests 送信した友達申請取得
// @Summary      送信した友達申請取得
// @Description  自分が送信した友達申請一覧を取得します（ページング対応）
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(20) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} dto.SentRequestsResponse "送信済み申請一覧取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends/requests/sent [get]
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

// GetMutualFriends 共通の友達取得
// @Summary      共通の友達取得
// @Description  指定されたユーザーとの共通の友達を取得します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        userId path string true "対象ユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} object{data=[]FriendWithUserInfoResponse} "共通の友達取得成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "ユーザーが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/friends/{userId}/mutual [get]
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

// === 招待管理 ===

// CreateInvitation 招待作成
// @Summary      招待作成
// @Description  友達招待またはグループ招待を作成します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        request body CreateInvitationRequest true "招待作成情報"
// @Security     BearerAuth
// @Success      201 {object} InvitationResponse "招待作成成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations [post]
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

// GetInvitation 招待詳細取得
// @Summary      招待詳細取得
// @Description  招待の詳細情報を取得します（権限チェック付き）
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        invitationId path string true "招待ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} InvitationResponse "招待詳細取得成功"
// @Failure      400 {object} ErrorResponse "招待IDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "この招待を閲覧する権限がない"
// @Failure      404 {object} ErrorResponse "招待が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/{invitationId} [get]
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

// GetInvitationByCode 招待コードから招待取得
// @Summary      招待コードから招待取得
// @Description  招待コードを使用して招待情報を取得します（パブリック情報のみ）
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        code path string true "招待コード" example:"abc123def456"
// @Success      200 {object} object{data=object{id=string,type=string,method=string,status=string,message=string,expires_at=string,created_at=string}} "招待情報取得成功"
// @Failure      400 {object} ErrorResponse "招待コードが必要"
// @Failure      404 {object} ErrorResponse "有効な招待が見つからない"
// @Failure      410 {object} ErrorResponse "招待の有効期限が切れている"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/code/{code} [get]
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

// AcceptInvitation 招待受諾
// @Summary      招待受諾
// @Description  招待コードを使用して招待を受諾します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        code path string true "招待コード" example:"abc123def456"
// @Security     BearerAuth
// @Success      200 {object} InvitationResultResponse "招待受諾成功"
// @Failure      400 {object} ErrorResponse "招待コードが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "有効な招待が見つからない"
// @Failure      410 {object} ErrorResponse "招待の有効期限が切れている"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/{code}/accept [post]
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

// DeclineInvitation 招待拒否
// @Summary      招待拒否
// @Description  受信した招待を拒否します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        invitationId path string true "招待ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "招待拒否成功"
// @Failure      400 {object} ErrorResponse "招待IDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "この招待を拒否する権限がない"
// @Failure      404 {object} ErrorResponse "招待が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/{invitationId}/decline [put]
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

// CancelInvitation 招待キャンセル
// @Summary      招待キャンセル
// @Description  送信した招待をキャンセルします
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        invitationId path string true "招待ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "招待キャンセル成功"
// @Failure      400 {object} ErrorResponse "招待IDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "この招待をキャンセルする権限がない"
// @Failure      404 {object} ErrorResponse "招待が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/{invitationId} [delete]
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

// GetSentInvitations 送信した招待一覧取得
// @Summary      送信した招待一覧取得
// @Description  自分が送信した招待一覧を取得します（ページング対応）
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(20) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} dto.InvitationsListResponse "送信済み招待一覧取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/sent [get]
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

// GetReceivedInvitations 受信した招待一覧取得
// @Summary      受信した招待一覧取得
// @Description  自分が受信した招待一覧を取得します（ページング対応）
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(20) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} dto.InvitationsListResponse "受信済み招待一覧取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/received [get]
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

// GenerateInviteURL 招待URL生成
// @Summary      招待URL生成
// @Description  指定された招待のURLを生成します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        invitationId path string true "招待ID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} InviteURLResponse "招待URL生成成功"
// @Failure      400 {object} ErrorResponse "招待IDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "この招待のURLを生成する権限がない"
// @Failure      404 {object} ErrorResponse "招待が見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/invitations/{invitationId}/url [get]
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

// GetRelationship ユーザー間関係取得
// @Summary      ユーザー間関係取得
// @Description  指定されたユーザーとの関係性（友達、ブロック、申請状況）を取得します
// @Tags         social
// @Accept       json
// @Produce      json
// @Param        userId path string true "対象ユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} UserRelationshipResponse "関係性取得成功"
// @Failure      400 {object} ErrorResponse "ユーザーIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "ユーザーが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /social/relationships/{userId} [get]
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

// RegisterSocialRoutes はソーシャル関連のルートを登録する
func RegisterSocialRoutes(router *gin.RouterGroup, controller *SocialController) {
	social := router.Group("/social")
	{
		// 友達関係管理
		social.POST("/friends/request", controller.SendFriendRequest)
		social.PUT("/friends/requests/:friendshipId/accept", controller.AcceptFriendRequest)
		social.PUT("/friends/requests/:friendshipId/decline", controller.DeclineFriendRequest)
		social.DELETE("/friends/:userId", controller.RemoveFriend)

		// ブロック機能
		social.POST("/users/:userId/block", controller.BlockUser)
		social.DELETE("/users/:userId/block", controller.UnblockUser)

		// 友達一覧・検索
		social.GET("/friends", controller.GetFriends)
		social.GET("/friends/:userId/mutual", controller.GetMutualFriends)
		social.GET("/friends/requests/received", controller.GetPendingRequests)
		social.GET("/friends/requests/sent", controller.GetSentRequests)

		// 招待管理
		social.POST("/invitations", controller.CreateInvitation)
		social.GET("/invitations/:invitationId", controller.GetInvitation)
		social.GET("/invitations/code/:code", controller.GetInvitationByCode)
		social.POST("/invitations/:code/accept", controller.AcceptInvitation)
		social.PUT("/invitations/:invitationId/decline", controller.DeclineInvitation)
		social.DELETE("/invitations/:invitationId", controller.CancelInvitation)
		social.GET("/invitations/sent", controller.GetSentInvitations)
		social.GET("/invitations/received", controller.GetReceivedInvitations)
		social.GET("/invitations/:invitationId/url", controller.GenerateInviteURL)

		// 関係性チェック
		social.GET("/relationships/:userId", controller.GetRelationship)
	}
}
