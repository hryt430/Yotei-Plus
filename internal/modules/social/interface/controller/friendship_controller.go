package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type FriendshipController struct {
	socialService usecase.SocialService
	logger        logger.Logger
}

func NewFriendshipController(socialService usecase.SocialService, logger logger.Logger) *FriendshipController {
	return &FriendshipController{
		socialService: socialService,
		logger:        logger,
	}
}

// SendFriendRequest は友達申請を送信する
// POST /api/v1/friends/request
func (fc *FriendshipController) SendFriendRequest(c *gin.Context) {
	// 認証ユーザー取得
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストの形式が正しくありません"})
		return
	}

	// UUIDパース
	addresseeID, err := uuid.Parse(req.UserID)
	if err != nil {
		fc.logger.Error("Invalid user ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なユーザーIDです"})
		return
	}

	// 自分自身への申請をチェック
	if user.ID == addresseeID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "自分自身に友達申請はできません"})
		return
	}

	friendship, err := fc.socialService.SendFriendRequest(c.Request.Context(), user.ID, addresseeID, req.Message)
	if err != nil {
		fc.logger.Error("Failed to send friend request",
			logger.Any("requesterID", user.ID),
			logger.Any("addresseeID", addresseeID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "友達申請の送信に失敗しました"})
		return
	}

	fc.logger.Info("Friend request sent successfully",
		logger.Any("requesterID", user.ID),
		logger.Any("addresseeID", addresseeID))

	c.JSON(http.StatusCreated, gin.H{
		"data":    friendship,
		"message": "友達申請を送信しました",
	})
}

// AcceptFriendRequest は友達申請を承認する
// PUT /api/v1/friends/{friendshipId}/accept
func (fc *FriendshipController) AcceptFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	friendshipIDStr := c.Param("friendshipId")
	friendshipID, err := uuid.Parse(friendshipIDStr)
	if err != nil {
		fc.logger.Error("Invalid friendship ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な友達申請IDです"})
		return
	}

	// TODO: friendshipIDから実際のrequesterIDとaddresseeIDを取得する必要があります
	// 現在のAPIエンドポイント設計では、friendshipIDからrequesterIDを特定する必要があります
	// この部分は実装時に調整が必要です

	friendship, err := fc.socialService.AcceptFriendRequest(c.Request.Context(), friendshipID, user.ID)
	if err != nil {
		fc.logger.Error("Failed to accept friend request",
			logger.Any("friendshipID", friendshipID),
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "友達申請の承認に失敗しました"})
		return
	}

	fc.logger.Info("Friend request accepted successfully",
		logger.Any("friendshipID", friendshipID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"data":    friendship,
		"message": "友達申請を承認しました",
	})
}

// DeclineFriendRequest は友達申請を拒否する
// PUT /api/v1/friends/{friendshipId}/decline
func (fc *FriendshipController) DeclineFriendRequest(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	friendshipIDStr := c.Param("friendshipId")
	friendshipID, err := uuid.Parse(friendshipIDStr)
	if err != nil {
		fc.logger.Error("Invalid friendship ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な友達申請IDです"})
		return
	}

	err = fc.socialService.DeclineFriendRequest(c.Request.Context(), friendshipID, user.ID)
	if err != nil {
		fc.logger.Error("Failed to decline friend request",
			logger.Any("friendshipID", friendshipID),
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "友達申請の拒否に失敗しました"})
		return
	}

	fc.logger.Info("Friend request declined successfully",
		logger.Any("friendshipID", friendshipID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "友達申請を拒否しました",
	})
}

// RemoveFriend は友達を削除する
// DELETE /api/v1/friends/{userId}
func (fc *FriendshipController) RemoveFriend(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	friendIDStr := c.Param("userId")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		fc.logger.Error("Invalid friend ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なユーザーIDです"})
		return
	}

	err = fc.socialService.RemoveFriend(c.Request.Context(), user.ID, friendID)
	if err != nil {
		fc.logger.Error("Failed to remove friend",
			logger.Any("userID", user.ID),
			logger.Any("friendID", friendID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "友達の削除に失敗しました"})
		return
	}

	fc.logger.Info("Friend removed successfully",
		logger.Any("userID", user.ID),
		logger.Any("friendID", friendID))

	c.JSON(http.StatusOK, gin.H{
		"message": "友達を削除しました",
	})
}

// BlockUser はユーザーをブロックする
// POST /api/v1/friends/{userId}/block
func (fc *FriendshipController) BlockUser(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	targetIDStr := c.Param("userId")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		fc.logger.Error("Invalid target ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なユーザーIDです"})
		return
	}

	err = fc.socialService.BlockUser(c.Request.Context(), user.ID, targetID)
	if err != nil {
		fc.logger.Error("Failed to block user",
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザーのブロックに失敗しました"})
		return
	}

	fc.logger.Info("User blocked successfully",
		logger.Any("userID", user.ID),
		logger.Any("targetID", targetID))

	c.JSON(http.StatusOK, gin.H{
		"message": "ユーザーをブロックしました",
	})
}

// UnblockUser はブロックを解除する
// DELETE /api/v1/friends/{userId}/block
func (fc *FriendshipController) UnblockUser(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	targetIDStr := c.Param("userId")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		fc.logger.Error("Invalid target ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なユーザーIDです"})
		return
	}

	err = fc.socialService.UnblockUser(c.Request.Context(), user.ID, targetID)
	if err != nil {
		fc.logger.Error("Failed to unblock user",
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ブロック解除に失敗しました"})
		return
	}

	fc.logger.Info("User unblocked successfully",
		logger.Any("userID", user.ID),
		logger.Any("targetID", targetID))

	c.JSON(http.StatusOK, gin.H{
		"message": "ブロックを解除しました",
	})
}

// GetFriends は友達一覧を取得する
// GET /api/v1/friends
func (fc *FriendshipController) GetFriends(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	// クエリパラメータ解析
	status := c.Query("status")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := domain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	friends, err := fc.socialService.GetFriends(c.Request.Context(), user.ID, pagination)
	if err != nil {
		fc.logger.Error("Failed to get friends",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "友達一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       friends,
		"pagination": pagination,
	})
}

// GetPendingRequests は受信した友達申請を取得する
// GET /api/v1/friends/pending
func (fc *FriendshipController) GetPendingRequests(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := domain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	requests, err := fc.socialService.GetPendingRequests(c.Request.Context(), user.ID, pagination)
	if err != nil {
		fc.logger.Error("Failed to get pending requests",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "友達申請の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       requests,
		"pagination": pagination,
	})
}

// GetSentRequests は送信した友達申請を取得する
// GET /api/v1/friends/sent
func (fc *FriendshipController) GetSentRequests(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := domain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	requests, err := fc.socialService.GetSentRequests(c.Request.Context(), user.ID, pagination)
	if err != nil {
		fc.logger.Error("Failed to get sent requests",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "送信済み友達申請の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       requests,
		"pagination": pagination,
	})
}

// GetMutualFriends は共通の友達を取得する
// GET /api/v1/friends/{userId}/mutual
func (fc *FriendshipController) GetMutualFriends(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		fc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	targetIDStr := c.Param("userId")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		fc.logger.Error("Invalid target ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なユーザーIDです"})
		return
	}

	mutualFriends, err := fc.socialService.GetMutualFriends(c.Request.Context(), user.ID, targetID)
	if err != nil {
		fc.logger.Error("Failed to get mutual friends",
			logger.Any("userID", user.ID),
			logger.Any("targetID", targetID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "共通の友達の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": mutualFriends,
	})
}
