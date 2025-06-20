package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase"
	"github.com/hryt430/Yotei+/pkg/utils"
)

// TaskController はタスク関連のHTTPリクエストを処理するコントローラー
type TaskController struct {
	taskService usecase.TaskService
}

// NewTaskController は新しいTaskControllerを作成する
func NewTaskController(taskService usecase.TaskService) *TaskController {
	return &TaskController{
		taskService: taskService,
	}
}

// TaskRequest はタスク作成/更新リクエスト
type TaskRequest struct {
	Title       string        `json:"title" binding:"omitempty,min=1"`
	Description string        `json:"description"`
	Status      string        `json:"status" binding:"omitempty,oneof=TODO IN_PROGRESS DONE"`
	Priority    string        `json:"priority" binding:"omitempty,oneof=LOW MEDIUM HIGH"`
	AssigneeID  *string       `json:"assignee_id"`
	DueDate     *FlexibleTime `json:"due_date"`
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

// FlexibleTime は複数の日付フォーマットに対応するカスタム型
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON は JSON からの柔軟な日付パース
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	if str == "null" || str == "" {
		return nil
	}

	// 対応フォーマット一覧
	formats := []string{
		time.RFC3339,          // "2024-12-01T15:30:00Z"
		"2006-01-02T15:04:05", // "2024-12-01T15:30:00"
		"2006-01-02 15:04:05", // "2024-12-01 15:30:00"
		"2006-01-02",          // "2024-12-01"
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("cannot parse '%s' as valid date format", str)
}

// MarshalJSON は JSON への出力
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time.Format(time.RFC3339))
}

// AssignTaskRequest はタスク割り当てリクエスト
type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required"`
}

// ChangeStatusRequest はステータス変更リクエスト
type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=TODO IN_PROGRESS DONE"`
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
		IsOverdue:   task.CheckIsOverdue(),
	}
}

// tasksToResponse はタスクリストをレスポンス形式に変換する
func tasksToResponse(tasks []*domain.Task) []TaskResponse {
	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, taskToResponse(task))
	}
	return taskResponses
}

// getUserIDFromContext は認証済みユーザーIDをコンテキストから取得する
func getUserIDFromContext(ctx *gin.Context) (string, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return "", errors.New("authentication required")
	}
	return userID.(string), nil
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

	// ユーザーID取得
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	// 優先度のデフォルト設定
	priority := domain.PriorityMedium
	if req.Priority != "" {
		priority = domain.Priority(req.Priority)
	}

	// タスク作成
	task, err := c.taskService.CreateTaskWithDefaults(
		ctx,
		req.Title,
		req.Description,
		priority,
		userID,
	)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	if req.DueDate != nil && !req.DueDate.Time.IsZero() {
		dueDate := req.DueDate.Time
		_, err = c.taskService.UpdateTask(
			ctx,
			task.ID,
			nil, nil, nil, nil, // title, description, status, priority は nil
			&dueDate,
		)
		if err != nil {
			handleServiceError(ctx, err)
			return
		}
		// レスポンス用にタスクの期限日を更新
		task.DueDate = &dueDate
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

	if req.DueDate != nil && !req.DueDate.Time.IsZero() {
		dueDate = &req.DueDate.Time
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
	// フィルタリング条件の設定
	filter := parseListFilter(ctx)

	// ページネーション設定
	pagination := parsePagination(ctx)

	// ソート設定
	sortOptions := parseSortOptions(ctx)

	// タスク一覧取得
	tasks, total, err := c.taskService.ListTasks(ctx, filter, pagination, sortOptions)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	// レスポンス作成
	taskResponses := tasksToResponse(tasks)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks":       taskResponses,
			"total_count": total,
			"page":        pagination.Page,
			"page_size":   pagination.PageSize,
		},
	})
}

// AssignTask はタスクを割り当てるハンドラー
func (c *TaskController) AssignTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	var req AssignTaskRequest
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

	var req ChangeStatusRequest
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

	taskResponses := tasksToResponse(tasks)

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
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	tasks, err := c.taskService.GetTasksByAssignee(ctx, userID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	taskResponses := tasksToResponse(tasks)

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

	taskResponses := tasksToResponse(tasks)

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
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 && limitNum <= 100 {
			limit = limitNum
		}
	}

	// サービス層の SearchTasks メソッドを呼び出し
	tasks, err := c.taskService.SearchTasks(ctx, query, limit)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	taskResponses := tasksToResponse(tasks)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks": taskResponses,
			"count": len(taskResponses),
		},
	})
}

// parseListFilter はクエリパラメータからフィルタを解析する
func parseListFilter(ctx *gin.Context) domain.ListFilter {
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

	if dueDateFromStr := ctx.Query("due_date_from"); dueDateFromStr != "" {
		ft := &FlexibleTime{}
		if err := ft.UnmarshalJSON([]byte(`"` + dueDateFromStr + `"`)); err == nil {
			filter.DueDateFrom = &ft.Time
		}
	}

	if dueDateToStr := ctx.Query("due_date_to"); dueDateToStr != "" {
		ft := &FlexibleTime{}
		if err := ft.UnmarshalJSON([]byte(`"` + dueDateToStr + `"`)); err == nil {
			filter.DueDateTo = &ft.Time
		}
	}

	return filter
}

// parsePagination はクエリパラメータからページネーション情報を解析する
func parsePagination(ctx *gin.Context) domain.Pagination {
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

	return domain.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// parseSortOptions はクエリパラメータからソートオプションを解析する
func parseSortOptions(ctx *gin.Context) domain.SortOptions {
	sortField := "created_at"
	sortDirection := "DESC"

	if sf := ctx.Query("sort_field"); sf != "" {
		// ソートフィールドのバリデーション
		allowedFields := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
			"priority":   true,
			"status":     true,
			"due_date":   true,
		}
		if allowedFields[sf] {
			sortField = sf
		}
	}

	if sd := ctx.Query("sort_direction"); sd == "ASC" || sd == "DESC" {
		sortDirection = sd
	}

	return domain.SortOptions{
		Field:     sortField,
		Direction: sortDirection,
	}
}

// handleServiceError はサービスレイヤーからのエラーを処理する
func handleServiceError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrTaskNotFound):
		ctx.JSON(http.StatusNotFound, utils.ErrorResponse("Task not found"))
	case errors.Is(err, usecase.ErrInvalidParameter):
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid parameters"))
	default:
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Internal server error"))
	}
}
