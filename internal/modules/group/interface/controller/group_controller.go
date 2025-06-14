// internal/modules/group/interface/controller/group_controller.go
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

// CreateGroup はグループを作成する
func (gc *GroupController) CreateGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
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
		gc.logger.Error("Failed to create group", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの作成に失敗しました",
		})
		return
	}

	response := dto.ToGroupResponse(group)
	c.JSON(http.StatusCreated, response)
}

// GetGroup はグループ詳細を取得する
func (gc *GroupController) GetGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	groupWithMembers, err := gc.groupService.GetGroup(c.Request.Context(), groupID, user.ID)
	if err != nil {
		gc.logger.Error("Failed to get group", logger.Error(err))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "GROUP_NOT_FOUND",
			Message: "グループが見つかりません",
		})
		return
	}

	response := dto.ToGroupWithMembersResponse(groupWithMembers)
	c.JSON(http.StatusOK, response)
}

// UpdateGroup はグループ情報を更新する
func (gc *GroupController) UpdateGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
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
		gc.logger.Error("Failed to update group", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの更新に失敗しました",
		})
		return
	}

	response := dto.ToGroupResponse(group)
	c.JSON(http.StatusOK, response)
}

// DeleteGroup はグループを削除する
func (gc *GroupController) DeleteGroup(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	err = gc.groupService.DeleteGroup(c.Request.Context(), groupID, user.ID)
	if err != nil {
		gc.logger.Error("Failed to delete group", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの削除に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "グループを削除しました",
	})
}

// ListMyGroups は自分のグループ一覧を取得する
func (gc *GroupController) ListMyGroups(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
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
		gc.logger.Error("Failed to get my groups", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループ一覧の取得に失敗しました",
		})
		return
	}

	response := dto.ToGroupListResponse(groups, total, page, pageSize)
	c.JSON(http.StatusOK, response)
}

// SearchGroups はグループを検索する
func (gc *GroupController) SearchGroups(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
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
		gc.logger.Error("Failed to search groups", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループの検索に失敗しました",
		})
		return
	}

	response := dto.ToGroupListResponse(groups, total, page, pageSize)
	c.JSON(http.StatusOK, response)
}

// AddMember はメンバーを追加する
func (gc *GroupController) AddMember(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "リクエストボディが不正です",
		})
		return
	}

	userIDToAdd, err := uuid.Parse(req.UserID)
	if err != nil {
		gc.logger.Error("Invalid user ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
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
		gc.logger.Error("Failed to add member", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバーの追加に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "メンバーを追加しました",
	})
}

// RemoveMember はメンバーを削除する
func (gc *GroupController) RemoveMember(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	userIDStr := c.Param("userId")
	userIDToRemove, err := uuid.Parse(userIDStr)
	if err != nil {
		gc.logger.Error("Invalid user ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "ユーザーIDが不正です",
		})
		return
	}

	err = gc.groupService.RemoveMember(c.Request.Context(), groupID, userIDToRemove, user.ID)
	if err != nil {
		gc.logger.Error("Failed to remove member", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバーの削除に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "メンバーを削除しました",
	})
}

// UpdateMemberRole はメンバーの権限を変更する
func (gc *GroupController) UpdateMemberRole(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	userIDStr := c.Param("userId")
	userIDToUpdate, err := uuid.Parse(userIDStr)
	if err != nil {
		gc.logger.Error("Invalid user ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "ユーザーIDが不正です",
		})
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gc.logger.Error("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "リクエストボディが不正です",
		})
		return
	}

	newRole := domain.MemberRole(req.Role)

	err = gc.groupService.UpdateMemberRole(c.Request.Context(), groupID, userIDToUpdate, user.ID, newRole)
	if err != nil {
		gc.logger.Error("Failed to update member role", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバー権限の更新に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "メンバー権限を更新しました",
	})
}

// ListMembers はメンバー一覧を取得する
func (gc *GroupController) ListMembers(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
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
		gc.logger.Error("Failed to get members", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "メンバー一覧の取得に失敗しました",
		})
		return
	}

	response := dto.ToMemberListResponse(members)
	c.JSON(http.StatusOK, response)
}

// GetGroupStats はグループ統計情報を取得する
func (gc *GroupController) GetGroupStats(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		gc.logger.Error("Failed to get user from context", logger.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "認証が必要です",
		})
		return
	}

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		gc.logger.Error("Invalid group ID", logger.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "INVALID_GROUP_ID",
			Message: "グループIDが不正です",
		})
		return
	}

	stats, err := gc.groupService.GetGroupStats(c.Request.Context(), groupID, user.ID)
	if err != nil {
		gc.logger.Error("Failed to get group stats", logger.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "グループ統計の取得に失敗しました",
		})
		return
	}

	response := dto.ToGroupStatsResponse(stats)
	c.JSON(http.StatusOK, response)
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
