package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type InvitationController struct {
	socialService usecase.SocialService
	logger        logger.Logger
}

func NewInvitationController(socialService usecase.SocialService, logger logger.Logger) *InvitationController {
	return &InvitationController{
		socialService: socialService,
		logger:        logger,
	}
}

// CreateInvitation は招待を作成する
// POST /api/v1/invitations
func (ic *InvitationController) CreateInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	var req struct {
		Type         string     `json:"type" binding:"required"`               // "FRIEND" or "GROUP"
		Method       string     `json:"method" binding:"required"`             // "IN_APP", "CODE", "URL"
		TargetID     *uuid.UUID `json:"target_id,omitempty"`                   // グループ招待の場合
		InviteeEmail *string    `json:"invitee_email,omitempty"`               // 未登録ユーザー招待の場合
		Message      string     `json:"message" binding:"max=500"`             // 招待メッセージ
		ExpiresHours int        `json:"expires_hours" binding:"min=1,max=168"` // 1時間〜1週間
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ic.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストの形式が正しくありません"})
		return
	}

	// 招待タイプのバリデーション
	invitationType := domain.InvitationType(req.Type)
	if invitationType != domain.InvitationTypeFriend && invitationType != domain.InvitationTypeGroup {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な招待タイプです"})
		return
	}

	// 招待方法のバリデーション
	invitationMethod := domain.InvitationMethod(req.Method)
	if invitationMethod != domain.MethodInApp && invitationMethod != domain.MethodCode && invitationMethod != domain.MethodURL {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な招待方法です"})
		return
	}

	// グループ招待の場合、TargetIDが必要
	if invitationType == domain.InvitationTypeGroup && req.TargetID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "グループ招待にはグループIDが必要です"})
		return
	}

	// デフォルト値設定
	if req.ExpiresHours == 0 {
		req.ExpiresHours = 168 // 1週間
	}

	input := usecase.CreateInvitationInput{
		Type:         invitationType,
		Method:       invitationMethod,
		TargetID:     req.TargetID,
		InviteeEmail: req.InviteeEmail,
		Message:      req.Message,
		ExpiresHours: req.ExpiresHours,
	}

	invitation, err := ic.socialService.CreateInvitation(c.Request.Context(), input)
	if err != nil {
		ic.logger.Error("Failed to create invitation",
			logger.Any("inviterID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "招待の作成に失敗しました"})
		return
	}

	ic.logger.Info("Invitation created successfully",
		logger.Any("inviterID", user.ID),
		logger.Any("invitationID", invitation.ID))

	c.JSON(http.StatusCreated, gin.H{
		"data":    invitation,
		"message": "招待を作成しました",
	})
}

// GetInvitation は招待詳細を取得する
// GET /api/v1/invitations/{invitationId}
func (ic *InvitationController) GetInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	invitationIDStr := c.Param("invitationId")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		ic.logger.Error("Invalid invitation ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な招待IDです"})
		return
	}

	invitation, err := ic.socialService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		ic.logger.Error("Failed to get invitation",
			logger.Any("invitationID", invitationID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "招待情報の取得に失敗しました"})
		return
	}

	if invitation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "招待が見つかりません"})
		return
	}

	// 権限チェック（招待者または被招待者のみ閲覧可能）
	if invitation.InviterID != user.ID &&
		(invitation.InviteeID == nil || *invitation.InviteeID != user.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "この招待を閲覧する権限がありません"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": invitation,
	})
}

// GetInvitationByCode は招待コードから招待情報を取得する
// GET /api/v1/invitations/code/{code}
func (ic *InvitationController) GetInvitationByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "招待コードが必要です"})
		return
	}

	invitation, err := ic.socialService.GetInvitationByCode(c.Request.Context(), code)
	if err != nil {
		ic.logger.Error("Failed to get invitation by code",
			logger.Any("code", code),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "招待情報の取得に失敗しました"})
		return
	}

	if invitation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "有効な招待が見つかりません"})
		return
	}

	// 期限切れチェック
	if invitation.IsExpired() {
		c.JSON(http.StatusGone, gin.H{"error": "招待の有効期限が切れています"})
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
// POST /api/v1/invitations/{code}/accept
func (ic *InvitationController) AcceptInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "招待コードが必要です"})
		return
	}

	result, err := ic.socialService.AcceptInvitation(c.Request.Context(), code, user.ID)
	if err != nil {
		ic.logger.Error("Failed to accept invitation",
			logger.Any("code", code),
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "招待の受諾に失敗しました"})
		return
	}

	ic.logger.Info("Invitation accepted successfully",
		logger.Any("code", code),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "招待を受諾しました",
	})
}

// DeclineInvitation は招待を拒否する
// PUT /api/v1/invitations/{invitationId}/decline
func (ic *InvitationController) DeclineInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	invitationIDStr := c.Param("invitationId")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		ic.logger.Error("Invalid invitation ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な招待IDです"})
		return
	}

	err = ic.socialService.DeclineInvitation(c.Request.Context(), invitationID, user.ID)
	if err != nil {
		ic.logger.Error("Failed to decline invitation",
			logger.Any("invitationID", invitationID),
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "招待の拒否に失敗しました"})
		return
	}

	ic.logger.Info("Invitation declined successfully",
		logger.Any("invitationID", invitationID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "招待を拒否しました",
	})
}

// CancelInvitation は招待をキャンセルする
// DELETE /api/v1/invitations/{invitationId}
func (ic *InvitationController) CancelInvitation(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	invitationIDStr := c.Param("invitationId")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		ic.logger.Error("Invalid invitation ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な招待IDです"})
		return
	}

	err = ic.socialService.CancelInvitation(c.Request.Context(), invitationID, user.ID)
	if err != nil {
		ic.logger.Error("Failed to cancel invitation",
			logger.Any("invitationID", invitationID),
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "招待のキャンセルに失敗しました"})
		return
	}

	ic.logger.Info("Invitation cancelled successfully",
		logger.Any("invitationID", invitationID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "招待をキャンセルしました",
	})
}

// GetSentInvitations は送信した招待一覧を取得する
// GET /api/v1/invitations/sent
func (ic *InvitationController) GetSentInvitations(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	// クエリパラメータ解析
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := commonDomain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	invitations, err := ic.socialService.GetSentInvitations(c.Request.Context(), user.ID, pagination)
	if err != nil {
		ic.logger.Error("Failed to get sent invitations",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "送信済み招待の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       invitations,
		"pagination": pagination,
	})
}

// GetReceivedInvitations は受信した招待一覧を取得する
// GET /api/v1/invitations/received
func (ic *InvitationController) GetReceivedInvitations(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	// クエリパラメータ解析
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := commonDomain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	invitations, err := ic.socialService.GetReceivedInvitations(c.Request.Context(), user.ID, pagination)
	if err != nil {
		ic.logger.Error("Failed to get received invitations",
			logger.Any("userID", user.ID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "受信済み招待の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       invitations,
		"pagination": pagination,
	})
}

// GenerateURL は招待URLを生成する
// GET /api/v1/invitations/{invitationId}/url
func (ic *InvitationController) GenerateURL(c *gin.Context) {
	_, err := middleware.GetUserFromContext(c)
	if err != nil {
		ic.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}

	invitationIDStr := c.Param("invitationId")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		ic.logger.Error("Invalid invitation ID format", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な招待IDです"})
		return
	}

	url, err := ic.socialService.GenerateInviteURL(c.Request.Context(), invitationID)
	if err != nil {
		ic.logger.Error("Failed to generate invite URL",
			logger.Any("invitationID", invitationID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "招待URLの生成に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
		"message": "招待URLを生成しました",
	})
}
