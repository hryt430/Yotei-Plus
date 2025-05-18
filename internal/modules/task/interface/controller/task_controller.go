package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase/management"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase/persistence"
	"github.com/hryt430/Yotei+/pkg/utils"
)

// TaskController はタスク関連のHTTPリクエストを処理するコントローラー
type TaskController struct {
	taskService persistence.TaskService
}

// NewTaskController は新しいTaskControllerを作成する
func NewTaskController(taskService persistence.TaskService) *TaskController {
	return &TaskController{
		taskService: taskService,
	}
}

// TaskRequest はタスク作成/更新リクエスト
type TaskRequest struct {
	Title       string     `json:"title" binding:"omitempty,min=1"`
	Description string     `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=TODO IN_PROGRESS DONE"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=LOW MEDIUM HIGH"`
	AssigneeID  *string    `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

// TaskResponse はタスクレスポンス
type TaskResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	CreatedBy   string     `json:"created_by"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	IsOverdue   bool       `json:"is_overdue"`
}

// ListResponse はタスク一覧レスポンス
type ListResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

// taskToResponse はドメインモデルからレスポンスモデルに変換する
func taskToResponse(task *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		AssigneeID:  task.AssigneeID,
		CreatedBy:   task.CreatedBy,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		IsOverdue:   task.IsOverdue(),
	}
}

// CreateTask はタスクを作成するハンドラー
func (c *TaskController) CreateTask(ctx *gin.Context) {
	var req TaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	// リクエストの検証
	if req.Title == "" {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Title is required"))
		return
	}

	// ユーザーID取得（認証済みユーザーのIDをコンテキストから取得）
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse("Authentication required"))
		return
	}

	// 優先度のデフォルト設定
	priority := domain.PriorityMedium
	if req.Priority != "" {
		priority = domain.Priority(req.Priority)
	}

	task, err := c.taskService.CreateTask(
		ctx,
		req.Title,
		req.Description,
		priority,
		userID.(string),
	)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Task created successfully",
		"data":    taskToResponse(task),
	})
}

// GetTask はタスクを取得するハンドラー
func (c *TaskController) GetTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	task, err := c.taskService.GetTask(ctx, taskID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    taskToResponse(task),
	})
}

// UpdateTask はタスクを更新するハンドラー
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	var req TaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	// 更新対象のフィールドを設定
	var title, description *string
	var status *domain.TaskStatus
	var priority *domain.Priority
	var dueDate *time.Time

	if req.Title != "" {
		title = &req.Title
	}
	if req.Description != "" {
		description = &req.Description
	}
	if req.Status != "" {
		s := domain.TaskStatus(req.Status)
		status = &s
	}
	if req.Priority != "" {
		p := domain.Priority(req.Priority)
		priority = &p
	}
	if req.DueDate != nil {
		dueDate = req.DueDate
	}

	task, err := c.taskService.UpdateTask(
		ctx,
		taskID,
		title,
		description,
		status,
		priority,
		dueDate,
	)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task updated successfully",
		"data":    taskToResponse(task),
	})
}

// DeleteTask はタスクを削除するハンドラー
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	err := c.taskService.DeleteTask(ctx, taskID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task deleted successfully",
	})
}

// ListTasks はタスク一覧を取得するハンドラー
func (c *TaskController) ListTasks(ctx *gin.Context) {
	// クエリパラメータの解析
	// フィルタリング条件の設定
	var filter domain.ListFilter

	if status := ctx.Query("status"); status != "" {
		s := domain.TaskStatus(status)
		filter.Status = &s
	}

	if priority := ctx.Query("priority"); priority != "" {
		p := domain.Priority(priority)
		filter.Priority = &p
	}

	if assigneeID := ctx.Query("assignee_id"); assigneeID != "" {
		filter.AssigneeID = &assigneeID
	}

	if createdBy := ctx.Query("created_by"); createdBy != "" {
		filter.CreatedBy = &createdBy
	}

	// ページネーション設定
	page := 1
	pageSize := 10

	if p := ctx.Query("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
			page = pageNum
		}
	}

	if ps := ctx.Query("page_size"); ps != "" {
		if pageSizeNum, err := strconv.Atoi(ps); err == nil && pageSizeNum > 0 && pageSizeNum <= 100 {
			pageSize = pageSizeNum
		}
	}

	pagination := domain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// ソート設定
	sortField := "created_at"
	sortDirection := "DESC"

	if sf := ctx.Query("sort_field"); sf != "" {
		sortField = sf
	}

	if sd := ctx.Query("sort_direction"); sd == "ASC" || sd == "DESC" {
		sortDirection = sd
	}

	sortOptions := domain.SortOptions{
		Field:     sortField,
		Direction: sortDirection,
	}

	// タスク一覧取得
	tasks, total, err := c.taskService.ListTasks(ctx, filter, pagination, sortOptions)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	// レスポンス作成
	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, taskToResponse(task))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks":       taskResponses,
			"total_count": total,
			"page":        page,
			"page_size":   pageSize,
		},
	})
}

// AssignTask はタスクを割り当てるハンドラー
func (c *TaskController) AssignTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	task, err := c.taskService.AssignTask(ctx, taskID, req.AssigneeID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task assigned successfully",
		"data":    taskToResponse(task),
	})
}

// ChangeTaskStatus はタスクのステータスを変更するハンドラー
func (c *TaskController) ChangeTaskStatus(ctx *gin.Context) {
	taskID := ctx.Param("id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=TODO IN_PROGRESS DONE"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	status := domain.TaskStatus(req.Status)
	task, err := c.taskService.ChangeTaskStatus(ctx, taskID, status)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task status changed successfully",
		"data":    taskToResponse(task),
	})
}

// GetOverdueTasks は期限切れのタスクを取得するハンドラー
func (c *TaskController) GetOverdueTasks(ctx *gin.Context) {
	tasks, err := c.taskService.GetOverdueTasks(ctx)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, taskToResponse(task))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks": taskResponses,
			"count": len(taskResponses),
		},
	})
}

// GetMyTasks は現在のユーザーのタスクを取得するハンドラー
func (c *TaskController) GetMyTasks(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse("Authentication required"))
		return
	}

	tasks, err := c.taskService.GetTasksByAssignee(ctx, userID.(string))
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, taskToResponse(task))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks": taskResponses,
			"count": len(taskResponses),
		},
	})
}

// GetUserTasks は特定のユーザーのタスクを取得するハンドラー
func (c *TaskController) GetUserTasks(ctx *gin.Context) {
	userID := ctx.Param("user_id")

	tasks, err := c.taskService.GetTasksByAssignee(ctx, userID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, taskToResponse(task))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks": taskResponses,
			"count": len(taskResponses),
		},
	})
}

// SearchTasks はタスクを検索するハンドラー
func (c *TaskController) SearchTasks(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Search query is required"))
		return
	}

	limit := 20
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = limitNum
		}
	}

	// リポジトリから直接検索を呼び出す
	// インターフェースチェックのために型アサーション
	taskRepo, ok := c.taskService.(interface {
		Search(ctx context.Context, query string, limit int) ([]*domain.Task, error)
	})

	if !ok {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Search functionality not available"))
		return
	}

	tasks, err := taskRepo.(interface {
		Search(ctx context.Context, query string, limit int) ([]*domain.Task, error)
	}).Search(ctx, query, limit)

	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, taskToResponse(task))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks": taskResponses,
			"count": len(taskResponses),
		},
	})
}

// ユーティリティ関数
// handleServiceError はサービスレイヤーからのエラーを処理する
func handleServiceError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, management.ErrTaskNotFound):
		ctx.JSON(http.StatusNotFound, utils.ErrorResponse("Task not found"))
	case errors.Is(err, management.ErrInvalidParameter):
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid parameters"))
	default:
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Internal server error"))
	}
}
