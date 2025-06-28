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
	Title       string        `json:"title" binding:"omitempty,min=1" example:"重要なタスク"`
	Description string        `json:"description" example:"タスクの詳細説明"`
	Status      string        `json:"status" binding:"omitempty,oneof=TODO IN_PROGRESS DONE" example:"TODO"`
	Priority    string        `json:"priority" binding:"omitempty,oneof=LOW MEDIUM HIGH" example:"HIGH"`
	Category    string        `json:"category" binding:"omitempty,oneof=WORK PERSONAL STUDY HEALTH SHOPPING OTHER" example:"WORK"`
	AssigneeID  *string       `json:"assignee_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	DueDate     *time.Time `json:"due_date" format:"date-time" example:"2024-12-31T23:59:59Z"`
} // @name TaskRequest

// TaskResponse はタスクレスポンス
type TaskResponse struct {
	ID          string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string     `json:"title" example:"重要なタスク"`
	Description string     `json:"description" example:"タスクの詳細説明"`
	Status      string     `json:"status" example:"TODO"`
	Priority    string     `json:"priority" example:"HIGH"`
	Category    string     `json:"category" example:"WORK"`
	AssigneeID  *string    `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedBy   string     `json:"created_by" example:"123e4567-e89b-12d3-a456-426614174000"`
	DueDate     *time.Time `json:"due_date,omitempty" example:"2024-12-31T23:59:59Z"`
	IsOverdue   bool       `json:"is_overdue" example:"false"`
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
} // @name TaskResponse

// TaskCreateResponse はタスク作成レスポンス
type TaskCreateResponse struct {
	Success bool         `json:"success" example:"true"`
	Message string       `json:"message" example:"Task created successfully"`
	Data    TaskResponse `json:"data"`
} // @name TaskCreateResponse

// TaskUpdateResponse はタスク更新レスポンス
type TaskUpdateResponse struct {
	Success bool         `json:"success" example:"true"`
	Message string       `json:"message" example:"Task updated successfully"`
	Data    TaskResponse `json:"data"`
} // @name TaskUpdateResponse

// TaskGetResponse はタスク取得レスポンス
type TaskGetResponse struct {
	Success bool         `json:"success" example:"true"`
	Data    TaskResponse `json:"data"`
} // @name TaskGetResponse

// TaskListResponse はタスク一覧レスポンス
type TaskListResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		Tasks      []TaskResponse `json:"tasks"`
		TotalCount int            `json:"total_count" example:"50"`
		Page       int            `json:"page" example:"1"`
		PageSize   int            `json:"page_size" example:"10"`
	} `json:"data"`
} // @name TaskListResponse

// TaskDeleteResponse はタスク削除レスポンス
type TaskDeleteResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Task deleted successfully"`
} // @name TaskDeleteResponse

// AssignTaskRequest はタスク割り当てリクエスト
type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
} // @name AssignTaskRequest

// ErrorResponse はエラーレスポンス構造体
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"リクエストが無効です"`
} // @name ErrorResponse

// ChangeStatusRequest はステータス変更リクエスト
type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=TODO IN_PROGRESS DONE" example:"IN_PROGRESS"`
} // @name ChangeStatusRequest

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

// CreateTask タスク作成
// @Summary      タスク作成
// @Description  新しいタスクを作成します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request body TaskRequest true "タスク作成情報"
// @Security     BearerAuth
// @Success      201 {object} TaskCreateResponse "タスク作成成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks [post]
func (c *TaskController) CreateTask(ctx *gin.Context) {
	var req TaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// リクエストの検証
	if req.Title == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Title is required",
	})
		return
	}

	// ユーザーID取得
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
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

	if req.DueDate != nil && !req.DueDate.IsZero() {
		dueDate := *req.DueDate
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

// GetTask タスク取得
// @Summary      タスク取得
// @Description  指定されたIDのタスクを取得します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id path string true "タスクID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} TaskGetResponse "タスク取得成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "タスクが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/{id} [get]
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

// UpdateTask タスク更新
// @Summary      タスク更新
// @Description  指定されたIDのタスクを更新します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id path string true "タスクID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        request body TaskRequest true "タスク更新情報"
// @Security     BearerAuth
// @Success      200 {object} TaskUpdateResponse "タスク更新成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "タスクが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/{id} [put]
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	var req TaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
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

	if req.DueDate != nil && !req.DueDate.IsZero() {
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

// DeleteTask タスク削除
// @Summary      タスク削除
// @Description  指定されたIDのタスクを削除します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id path string true "タスクID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} TaskDeleteResponse "タスク削除成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "タスクが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/{id} [delete]
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

// ListTasks タスク一覧取得
// @Summary      タスク一覧取得
// @Description  フィルタリング、ページング、ソート機能付きでタスク一覧を取得します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        status query string false "ステータスフィルタ" Enums(TODO,IN_PROGRESS,DONE)
// @Param        priority query string false "優先度フィルタ" Enums(LOW,MEDIUM,HIGH)
// @Param        category query string false "カテゴリフィルタ" Enums(WORK,PERSONAL,STUDY,HEALTH,SHOPPING,OTHER)
// @Param        assignee_id query string false "担当者IDフィルタ" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        created_by query string false "作成者IDフィルタ" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        due_date_from query string false "期限日FROM" example:"2024-01-01"
// @Param        due_date_to query string false "期限日TO" example:"2024-12-31"
// @Param        page query int false "ページ番号" default(1) minimum(1)
// @Param        page_size query int false "ページサイズ" default(10) minimum(1) maximum(100)
// @Param        sort_field query string false "ソートフィールド" Enums(created_at,updated_at,title,priority,status,due_date) default(created_at)
// @Param        sort_direction query string false "ソート方向" Enums(ASC,DESC) default(DESC)
// @Security     BearerAuth
// @Success      200 {object} TaskListResponse "タスク一覧取得成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks [get]
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

// AssignTask タスク割り当て
// @Summary      タスク割り当て
// @Description  指定されたタスクをユーザーに割り当てます
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id path string true "タスクID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        request body AssignTaskRequest true "割り当て情報"
// @Security     BearerAuth
// @Success      200 {object} TaskUpdateResponse "タスク割り当て成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "タスクまたはユーザーが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/{id}/assign [put]
func (c *TaskController) AssignTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	var req AssignTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
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

// ChangeTaskStatus タスクステータス変更
// @Summary      タスクステータス変更
// @Description  指定されたタスクのステータスを変更します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id path string true "タスクID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Param        request body ChangeStatusRequest true "ステータス変更情報"
// @Security     BearerAuth
// @Success      200 {object} TaskUpdateResponse "ステータス変更成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "タスクが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/{id}/status [put]
func (c *TaskController) ChangeTaskStatus(ctx *gin.Context) {
	taskID := ctx.Param("id")

	var req ChangeStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
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

// GetOverdueTasks 期限切れタスク取得
// @Summary      期限切れタスク取得
// @Description  期限が過ぎているタスクの一覧を取得します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} TaskListResponse "期限切れタスク取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/overdue [get]
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

// GetMyTasks 自分のタスク取得
// @Summary      自分のタスク取得
// @Description  現在認証されているユーザーに割り当てられたタスクを取得します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} TaskListResponse "自分のタスク取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/my [get]
func (c *TaskController) GetMyTasks(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
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

// GetUserTasks 特定ユーザーのタスク取得
// @Summary      特定ユーザーのタスク取得
// @Description  指定されたユーザーに割り当てられたタスクを取得します
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        user_id path string true "ユーザーID" example:"123e4567-e89b-12d3-a456-426614174000"
// @Security     BearerAuth
// @Success      200 {object} TaskListResponse "ユーザータスク取得成功"
// @Failure      400 {object} ErrorResponse "リクエストが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      404 {object} ErrorResponse "ユーザーが見つからない"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/user/{user_id} [get]
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

// SearchTasks タスク検索
// @Summary      タスク検索
// @Description  キーワードでタスクを検索します（タイトルと説明が対象）
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        q query string true "検索クエリ" example:"重要"
// @Param        limit query int false "結果の最大数" default(20) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} TaskListResponse "タスク検索成功"
// @Failure      400 {object} ErrorResponse "検索クエリが必要"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/search [get]
func (c *TaskController) SearchTasks(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Search query is required",
	})
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

// 以下、既存のヘルパー関数たち...

// taskToResponse はドメインモデルからレスポンスモデルに変換する
func taskToResponse(task *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		Category:    string(task.Category),
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

	if category := ctx.Query("category"); category != "" {
		c := domain.Category(category)
		filter.Category = &c
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
		ctx.JSON(http.StatusNotFound, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Task not found",
	})
	case errors.Is(err, usecase.ErrInvalidParameter):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid parameters",
	})
	default:
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Internal server error",
	})
	}
}
