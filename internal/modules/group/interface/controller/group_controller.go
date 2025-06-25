package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/interface/dto"
	groupUsecase "github.com/hryt430/Yotei+/internal/modules/group/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
	"go.uber.org/zap/zapcore"
)

type GroupController struct {
	groupService groupUsecase.GroupService
	logger       logger.Logger
}

func NewGroupController(groupService groupUsecase.GroupService, logger logger.Logger) *GroupController {
	return &GroupController{
		groupService: groupService,
		logger:       logger,
	}
}

// Swagger用のリクエスト/レスポンス構造体定義
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

// ErrorResponse はエラーレスポンス構造体
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"リクエストが無効です"`
} // @name ErrorResponse

// CreateGroup グループ作成
// @Summary      グループ作成
// @Description  新しいグループを作成します（プロジェクト管理用または予定共有用）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateGroupRequest true "グループ作成情報"
// @Security     BearerAuth
// @Success      201 {object} dto.GroupResponse "グループ作成成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups [post]
func (gc *GroupController) CreateGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logError("bind JSON", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "リクエストボディが不正です",
		})
		return
	}

	input := groupUsecase.CreateGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Type:        domain.GroupType(req.Type),
		OwnerID:     user.ID,
		Settings:    req.Settings,
	}

	group, err := gc.groupService.CreateGroup(c.Request.Context(), input)
	if err != nil {
		gc.logError("create group", err,
			logger.Any("userID", user.ID),
			logger.Any("groupName", req.Name))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの作成に失敗しました",
		})
		return
	}

	gc.logger.Info("Group created successfully",
		logger.Any("groupID", group.ID),
		logger.Any("ownerID", user.ID))

	response := dto.ToGroupResponse(group)
	c.JSON(http.StatusCreated, response)
}

// GetGroup グループ詳細取得
// @Summary      グループ詳細取得
// @Description  指定されたIDのグループの詳細情報とメンバー一覧を取得します
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} GroupWithMembersResponse "グループ詳細取得成功"
// @Failure      400 {object} ErrorResponse "グループIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "グループが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId} [get]
func (gc *GroupController) GetGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	groupWithMembers, err := gc.groupService.GetGroup(c.Request.Context(), groupID, user.ID)
	if err != nil {
		gc.logError("get group", err,
			logger.Any("groupID", groupID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "GROUP_NOT_FOUND",
			Message: "グループが見つかりません",
		})
		return
	}

	response := dto.ToGroupWithMembersResponse(groupWithMembers)
	c.JSON(http.StatusOK, response)
}

// UpdateGroup グループ更新
// @Summary      グループ更新
// @Description  指定されたIDのグループ情報を更新します（管理者のみ）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        request body UpdateGroupRequest true "グループ更新情報"
// @Security     BearerAuth
// @Success      200 {object} GroupResponse "グループ更新成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "権限不足"
// @Failure      404 {object} ErrorResponse "グループが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId} [put]
func (gc *GroupController) UpdateGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logError("bind JSON", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "リクエストボディが不正です",
		})
		return
	}

	input := groupUsecase.UpdateGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Settings:    req.Settings,
	}

	group, err := gc.groupService.UpdateGroup(c.Request.Context(), groupID, input, user.ID)
	if err != nil {
		gc.logError("update group", err,
			logger.Any("groupID", groupID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの更新に失敗しました",
		})
		return
	}

	gc.logger.Info("Group updated successfully",
		logger.Any("groupID", groupID),
		logger.Any("userID", user.ID))

	response := dto.ToGroupResponse(group)
	c.JSON(http.StatusOK, response)
}

// DeleteGroup グループ削除
// @Summary      グループ削除
// @Description  指定されたIDのグループを削除します（オーナーのみ）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "グループ削除成功"
// @Failure      400 {object} ErrorResponse "グループIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "権限不足（オーナーのみ削除可能）"
// @Failure      404 {object} ErrorResponse "グループが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId} [delete]
func (gc *GroupController) DeleteGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	err = gc.groupService.DeleteGroup(c.Request.Context(), groupID, user.ID)
	if err != nil {
		gc.logError("delete group", err,
			logger.Any("groupID", groupID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの削除に失敗しました",
		})
		return
	}

	gc.logger.Info("Group deleted successfully",
		logger.Any("groupID", groupID),
		logger.Any("userID", user.ID))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "グループを削除しました",
	})
}

// ListMyGroups 自分のグループ一覧取得
// @Summary      自分のグループ一覧取得
// @Description  自分が所属しているグループの一覧を取得します（ページング対応）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        type query string false "グループタイプでフィルタ" enums:"PROJECT,SCHEDULE"
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(10) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} GroupListResponse "グループ一覧取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/my [get]
func (gc *GroupController) ListMyGroups(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	// クエリパラメータの解析
	var groupType *domain.GroupType
	if typeStr := c.Query("type"); typeStr != "" {
		gt := domain.GroupType(typeStr)
		groupType = &gt
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	pagination := commonDomain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	groups, total, err := gc.groupService.GetMyGroups(c.Request.Context(), user.ID, groupType, pagination)
	if err != nil {
		gc.logError("get my groups", err, logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループ一覧の取得に失敗しました",
		})
		return
	}

	response := dto.ToGroupListResponse(groups, total, page, pageSize)
	c.JSON(http.StatusOK, response)
}

// SearchGroups グループ検索
// @Summary      グループ検索
// @Description  キーワードでグループを検索します（ページング対応）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        q query string true "検索クエリ" example:"プロジェクト"
// @Param        type query string false "グループタイプでフィルタ" enums:"PROJECT,SCHEDULE"
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(10) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} GroupListResponse "グループ検索成功"
// @Failure      400 {object} ErrorResponse "検索クエリが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/search [get]
func (gc *GroupController) SearchGroups(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "MISSING_QUERY",
			Message: "検索クエリが必要です",
		})
		return
	}

	var groupType *domain.GroupType
	if typeStr := c.Query("type"); typeStr != "" {
		gt := domain.GroupType(typeStr)
		groupType = &gt
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	pagination := commonDomain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	groups, total, err := gc.groupService.SearchGroups(c.Request.Context(), query, groupType, pagination)
	if err != nil {
		gc.logError("search groups", err,
			logger.Any("query", query),
			logger.Any("groupType", groupType))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの検索に失敗しました",
		})
		return
	}

	response := dto.ToGroupListResponse(groups, total, page, pageSize)
	c.JSON(http.StatusOK, response)
}

// AddMember メンバー追加
// @Summary      メンバー追加
// @Description  指定されたグループにメンバーを追加します（管理者のみ）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        request body AddMemberRequest true "メンバー追加情報"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "メンバー追加成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "権限不足"
// @Failure      404 {object} ErrorResponse "グループまたはユーザーが見つからない"
// @Failure      409 {object} ErrorResponse "既にメンバー"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId}/members [post]
func (gc *GroupController) AddMember(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logError("bind JSON", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "リクエストボディが不正です",
		})
		return
	}

	userIDToAdd, err := gc.validateUUID(req.UserID, "user ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "ユーザーIDが不正です",
		})
		return
	}

	role := domain.MemberRole(req.Role)
	if role == "" {
		role = domain.RoleMember
	}

	err = gc.groupService.AddMember(c.Request.Context(), groupID, userIDToAdd, user.ID, role)
	if err != nil {
		gc.logError("add member", err,
			logger.Any("groupID", groupID),
			logger.Any("userIDToAdd", userIDToAdd),
			logger.Any("requesterID", user.ID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバーの追加に失敗しました",
		})
		return
	}

	gc.logger.Info("Member added successfully",
		logger.Any("groupID", groupID),
		logger.Any("userIDToAdd", userIDToAdd),
		logger.Any("role", role))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "メンバーを追加しました",
	})
}

// RemoveMember メンバー削除
// @Summary      メンバー削除
// @Description  指定されたグループからメンバーを削除します（管理者のみ、または自分自身の脱退）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        userId path string true "削除するユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "メンバー削除成功"
// @Failure      400 {object} ErrorResponse "IDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "権限不足"
// @Failure      404 {object} ErrorResponse "グループまたはユーザーが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId}/members/{userId} [delete]
func (gc *GroupController) RemoveMember(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	userIDToRemove, err := gc.validateUUID(c.Param("userId"), "user ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "ユーザーIDが不正です",
		})
		return
	}

	err = gc.groupService.RemoveMember(c.Request.Context(), groupID, userIDToRemove, user.ID)
	if err != nil {
		gc.logError("remove member", err,
			logger.Any("groupID", groupID),
			logger.Any("userIDToRemove", userIDToRemove),
			logger.Any("requesterID", user.ID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバーの削除に失敗しました",
		})
		return
	}

	gc.logger.Info("Member removed successfully",
		logger.Any("groupID", groupID),
		logger.Any("userIDToRemove", userIDToRemove))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "メンバーを削除しました",
	})
}

// UpdateMemberRole メンバー権限変更
// @Summary      メンバー権限変更
// @Description  指定されたメンバーの権限を変更します（管理者のみ）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        userId path string true "権限変更するユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        request body UpdateMemberRoleRequest true "権限変更情報"
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse "権限変更成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "権限不足"
// @Failure      404 {object} ErrorResponse "グループまたはユーザーが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId}/members/{userId}/role [put]
func (gc *GroupController) UpdateMemberRole(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	userIDToUpdate, err := gc.validateUUID(c.Param("userId"), "user ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "ユーザーIDが不正です",
		})
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logError("bind JSON", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "リクエストボディが不正です",
		})
		return
	}

	newRole := domain.MemberRole(req.Role)

	err = gc.groupService.UpdateMemberRole(c.Request.Context(), groupID, userIDToUpdate, user.ID, newRole)
	if err != nil {
		gc.logError("update member role", err,
			logger.Any("groupID", groupID),
			logger.Any("userIDToUpdate", userIDToUpdate),
			logger.Any("newRole", newRole),
			logger.Any("requesterID", user.ID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバー権限の更新に失敗しました",
		})
		return
	}

	gc.logger.Info("Member role updated successfully",
		logger.Any("groupID", groupID),
		logger.Any("userIDToUpdate", userIDToUpdate),
		logger.Any("newRole", newRole))

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "メンバー権限を更新しました",
	})
}

// ListMembers メンバー一覧取得
// @Summary      メンバー一覧取得
// @Description  指定されたグループのメンバー一覧を取得します（ページング対応）
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(20) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} MemberListResponse "メンバー一覧取得成功"
// @Failure      400 {object} ErrorResponse "グループIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "グループへのアクセス権限なし"
// @Failure      404 {object} ErrorResponse "グループが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId}/members [get]
func (gc *GroupController) ListMembers(c *gin.Context) {
	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := commonDomain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	members, err := gc.groupService.GetMembers(c.Request.Context(), groupID, pagination)
	if err != nil {
		gc.logError("get members", err, logger.Any("groupID", groupID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバー一覧の取得に失敗しました",
		})
		return
	}

	response := dto.ToMemberListResponse(members)
	c.JSON(http.StatusOK, response)
}

// GetGroupStats グループ統計取得
// @Summary      グループ統計取得
// @Description  指定されたグループの統計情報（メンバー数、タスク数など）を取得します
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        groupId path string true "グループID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} GroupStatsResponse "統計情報取得成功"
// @Failure      400 {object} ErrorResponse "グループIDが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      403 {object} ErrorResponse "グループへのアクセス権限なし"
// @Failure      404 {object} ErrorResponse "グループが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /groups/{groupId}/stats [get]
func (gc *GroupController) GetGroupStats(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logError("get user from context", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupID, err := gc.validateUUID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	stats, err := gc.groupService.GetGroupStats(c.Request.Context(), groupID, user.ID)
	if err != nil {
		gc.logError("get group stats", err,
			logger.Any("groupID", groupID),
			logger.Any("userID", user.ID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループ統計の取得に失敗しました",
		})
		return
	}

	response := dto.ToGroupStatsResponse(stats)
	c.JSON(http.StatusOK, response)
}

// === ヘルパーメソッド ===

func (gc *GroupController) validateUUID(id string, fieldName string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		gc.logger.Error("Invalid UUID format",
			logger.String("field", fieldName),
			logger.String("value", id),
			logger.Error(err))
		return uuid.Nil, err
	}
	return parsedID, nil
}

func (gc *GroupController) logError(operation string, err error, fields ...zapcore.Field) {
	allFields := append([]zapcore.Field{
		logger.String("operation", operation),
		logger.Error(err),
	}, fields...)
	gc.logger.Error("Operation failed", allFields...)
}

// RegisterGroupRoutes はグループ関連のルートを登録する
func RegisterGroupRoutes(router *gin.RouterGroup, controller *GroupController) {
	groups := router.Group("/groups")
	{
		// グループ基本操作
		groups.POST("", controller.CreateGroup)
		groups.GET("/my", controller.ListMyGroups)
		groups.GET("/search", controller.SearchGroups)
		groups.GET("/:groupId", controller.GetGroup)
		groups.PUT("/:groupId", controller.UpdateGroup)
		groups.DELETE("/:groupId", controller.DeleteGroup)

		// メンバー管理
		groups.POST("/:groupId/members", controller.AddMember)
		groups.DELETE("/:groupId/members/:userId", controller.RemoveMember)
		groups.PUT("/:groupId/members/:userId/role", controller.UpdateMemberRole)
		groups.GET("/:groupId/members", controller.ListMembers)

		// 統計情報
		groups.GET("/:groupId/stats", controller.GetGroupStats)
	}
}
