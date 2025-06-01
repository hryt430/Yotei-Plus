package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/interface/dto"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// TaskRepository はタスクのデータベースリポジトリ実装（改良版）
type TaskRepository struct {
	SqlHandler
	logger logger.Logger
}

// NewTaskRepository は新しいTaskRepositoryを作成する
func NewTaskRepository(sqlHandler SqlHandler, logger logger.Logger) usecase.TaskRepository {
	return &TaskRepository{
		SqlHandler: sqlHandler,
		logger:     logger,
	}
}

// SQLインジェクション対策：許可されたソートフィールドの定義
var allowedSortFields = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"title":      "title",
	"priority":   "priority",
	"status":     "status",
	"due_date":   "due_date",
}

// SQLインジェクション対策：許可されたフィルタフィールドの定義
var allowedFilterFields = map[string]bool{
	"status":      true,
	"priority":    true,
	"assignee_id": true,
	"created_by":  true,
	"due_date":    true,
}

// CreateTask はタスクを作成する
func (r *TaskRepository) CreateTask(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO ` + "`Yotei-Plus`" + `.tasks (
			id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	model := dto.FromDomain(task)
	_, err := r.Execute(query,
		model.ID,
		model.Title,
		model.Description,
		model.Status,
		model.Priority,
		model.AssigneeID,
		model.CreatedBy,
		model.DueDate,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to create task", logger.Any("taskID", task.ID), logger.Error(err))
		return fmt.Errorf("failed to create task: %w", err)
	}

	r.logger.Debug("Task created successfully", logger.Any("taskID", task.ID))
	return nil
}

// GetTaskByID はIDによりタスクを取得する（コネクション管理改善）
func (r *TaskRepository) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		return nil, usecase.ErrInvalidParameter
	}

	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.tasks 
		WHERE id = ?
		LIMIT 1
	`

	row, err := r.Query(query, id)
	if err != nil {
		r.logger.Error("Failed to query task by ID", logger.Any("id", id), logger.Error(err))
		return nil, fmt.Errorf("failed to query task: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			r.logger.Error("Failed to close row", logger.Error(closeErr))
		}
	}()

	if !row.Next() {
		return nil, usecase.ErrTaskNotFound
	}

	task, err := r.scanTaskFromRow(row)
	if err != nil {
		r.logger.Error("Failed to scan task", logger.Any("id", id), logger.Error(err))
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	return task, nil
}

// ListTasks はタスク一覧を取得する（SQLインジェクション対策、パフォーマンス改善）
func (r *TaskRepository) ListTasks(
	ctx context.Context,
	filter domain.ListFilter,
	pagination domain.Pagination,
	sort domain.SortOptions,
) ([]*domain.Task, int, error) {
	// 入力値の検証とサニタイズ
	if err := r.validateListParams(filter, pagination, sort); err != nil {
		return nil, 0, err
	}

	// WHERE句とパラメータの構築（SQLインジェクション対策）
	whereClause, args := r.buildWhereClause(filter)

	// カウント取得（パフォーマンス改善：インデックス使用）
	total, err := r.getTaskCount(ctx, whereClause, args)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*domain.Task{}, 0, nil
	}

	// ソートフィールドとディレクションの検証
	sortField := r.validateSortField(sort.Field)
	sortDirection := r.validateSortDirection(sort.Direction)

	// メインクエリ（パフォーマンス改善：必要なカラムのみ選択）
	query := fmt.Sprintf(`
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM `+"`Yotei-Plus`"+`.tasks
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortField, sortDirection)

	// ページネーション用のパラメータを追加
	offset := (pagination.Page - 1) * pagination.PageSize
	args = append(args, pagination.PageSize, offset)

	rows, err := r.Query(query, args...)
	if err != nil {
		r.logger.Error("Failed to list tasks", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			r.logger.Error("Failed to scan task row", logger.Error(err))
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Tasks listed successfully",
		logger.Any("count", len(tasks)),
		logger.Any("total", total))

	return tasks, total, nil
}

// SearchTasks はタスクを検索する（SQLインジェクション対策、パフォーマンス改善）
func (r *TaskRepository) SearchTasks(ctx context.Context, query string, limit int) ([]*domain.Task, error) {
	// 入力値のサニタイズ
	query = strings.TrimSpace(query)
	if query == "" {
		return []*domain.Task{}, nil
	}

	// リミットの検証
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// FULLTEXT検索またはLIKE検索（パフォーマンス改善）
	// 本来はFULLTEXTのインデックスを使用するのが理想
	sqlQuery := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (title LIKE ? OR description LIKE ?)
		ORDER BY 
			CASE 
				WHEN title LIKE ? THEN 1 
				WHEN description LIKE ? THEN 2 
				ELSE 3 
			END,
			created_at DESC
		LIMIT ?
	`

	// ワイルドカードパターンの構築（SQLインジェクション対策）
	pattern := "%" + r.escapeLikePattern(query) + "%"
	exactPattern := r.escapeLikePattern(query) + "%"

	rows, err := r.Query(sqlQuery, pattern, pattern, exactPattern, exactPattern, limit)
	if err != nil {
		r.logger.Error("Failed to search tasks", logger.Any("query", query), logger.Error(err))
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			r.logger.Error("Failed to scan task in search", logger.Error(err))
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Task search completed",
		logger.Any("query", query),
		logger.Any("resultCount", len(tasks)))

	return tasks, nil
}

// GetOverdueTasks は期限切れのタスクを取得（大量データ対策）
func (r *TaskRepository) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	// 大量データ対策：期限を制限（例：過去30日以内）
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	now := time.Now()
	doneStatus := string(domain.TaskStatusDone)

	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE due_date < ? 
		  AND due_date >= ?
		  AND status != ?
		ORDER BY due_date ASC
		LIMIT 1000
	`

	rows, err := r.Query(query, now, thirtyDaysAgo, doneStatus)
	if err != nil {
		r.logger.Error("Failed to get overdue tasks", logger.Error(err))
		return nil, fmt.Errorf("failed to get overdue tasks: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			r.logger.Error("Failed to scan overdue task", logger.Error(err))
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Overdue tasks retrieved", logger.Any("count", len(tasks)))
	return tasks, nil
}

// GetTasksByAssignee は特定のユーザーに割り当てられたタスクを取得（パフォーマンス改善）
func (r *TaskRepository) GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error) {
	if userID == "" {
		return nil, usecase.ErrInvalidParameter
	}

	// パフォーマンス改善：インデックス利用、大量データ対策
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE assignee_id = ?
		ORDER BY 
			CASE status 
				WHEN 'TODO' THEN 1 
				WHEN 'IN_PROGRESS' THEN 2 
				WHEN 'DONE' THEN 3 
				ELSE 4 
			END,
			due_date ASC NULLS LAST,
			created_at DESC
		LIMIT 500
	`

	rows, err := r.Query(query, userID)
	if err != nil {
		r.logger.Error("Failed to get tasks by assignee", logger.Any("userID", userID), logger.Error(err))
		return nil, fmt.Errorf("failed to get tasks by assignee: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			r.logger.Error("Failed to scan assignee task", logger.Error(err))
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Tasks by assignee retrieved",
		logger.Any("userID", userID),
		logger.Any("count", len(tasks)))

	return tasks, nil
}

// UpdateTask はタスクを更新する（コネクション管理改善）
func (r *TaskRepository) UpdateTask(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE ` + "`Yotei-Plus`" + `.tasks SET
			title = ?,
			description = ?,
			status = ?,
			priority = ?,
			assignee_id = ?,
			due_date = ?,
			updated_at = ?
		WHERE id = ?
	`

	model := dto.FromDomain(task)
	result, err := r.Execute(query,
		model.Title,
		model.Description,
		model.Status,
		model.Priority,
		model.AssigneeID,
		model.DueDate,
		model.UpdatedAt,
		model.ID,
	)
	if err != nil {
		r.logger.Error("Failed to update task", logger.Any("taskID", task.ID), logger.Error(err))
		return fmt.Errorf("failed to update task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", logger.Error(err))
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return usecase.ErrTaskNotFound
	}

	r.logger.Debug("Task updated successfully", logger.Any("taskID", task.ID))
	return nil
}

// DeleteTask はタスクを削除する（物理削除）
func (r *TaskRepository) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return usecase.ErrInvalidParameter
	}

	query := `DELETE FROM ` + "`Yotei-Plus`" + `.tasks WHERE id = ?`

	result, err := r.Execute(query, id)
	if err != nil {
		r.logger.Error("Failed to delete task", logger.Any("taskID", id), logger.Error(err))
		return fmt.Errorf("failed to delete task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", logger.Error(err))
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return usecase.ErrTaskNotFound
	}

	r.logger.Debug("Task deleted successfully", logger.Any("taskID", id))
	return nil
}

// バリデーションと補助メソッド

func (r *TaskRepository) validateListParams(filter domain.ListFilter, pagination domain.Pagination, sort domain.SortOptions) error {
	// ページネーションの検証
	if pagination.Page <= 0 {
		return fmt.Errorf("invalid page number: %d", pagination.Page)
	}
	if pagination.PageSize <= 0 || pagination.PageSize > 100 {
		return fmt.Errorf("invalid page size: %d (must be 1-100)", pagination.PageSize)
	}

	// ソートフィールドの検証
	if sort.Field != "" && allowedSortFields[sort.Field] == "" {
		return fmt.Errorf("invalid sort field: %s", sort.Field)
	}

	return nil
}

func (r *TaskRepository) buildWhereClause(filter domain.ListFilter) (string, []interface{}) {
	var conds []string
	var args []interface{}

	if filter.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, string(*filter.Status))
	}
	if filter.Priority != nil {
		conds = append(conds, "priority = ?")
		args = append(args, string(*filter.Priority))
	}
	if filter.AssigneeID != nil {
		conds = append(conds, "assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}
	if filter.CreatedBy != nil {
		conds = append(conds, "created_by = ?")
		args = append(args, *filter.CreatedBy)
	}
	if filter.DueDateFrom != nil {
		conds = append(conds, "due_date >= ?")
		args = append(args, *filter.DueDateFrom)
	}
	if filter.DueDateTo != nil {
		conds = append(conds, "due_date <= ?")
		args = append(args, *filter.DueDateTo)
	}

	whereClause := ""
	if len(conds) > 0 {
		whereClause = "WHERE " + strings.Join(conds, " AND ")
	}

	return whereClause, args
}

func (r *TaskRepository) validateSortField(field string) string {
	if validField, exists := allowedSortFields[field]; exists {
		return validField
	}
	return "created_at" // デフォルト
}

func (r *TaskRepository) validateSortDirection(direction string) string {
	if direction == "ASC" || direction == "DESC" {
		return direction
	}
	return "DESC" // デフォルト
}

// escapeLikePattern はLIKE演算子のワイルドカードをエスケープ（SQLインジェクション対策）
func (r *TaskRepository) escapeLikePattern(pattern string) string {
	// MySQL/MariaDBの場合のエスケープ
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	return pattern
}

// scanTaskFromRow はRowからTaskをスキャンする共通処理（改善版）
func (r *TaskRepository) scanTaskFromRow(row Row) (*domain.Task, error) {
	var m dto.TaskModel
	var assigneeID sql.NullString
	var dueDate sql.NullTime

	err := row.Scan(
		&m.ID,
		&m.Title,
		&m.Description,
		&m.Status,
		&m.Priority,
		&assigneeID,
		&m.CreatedBy,
		&dueDate,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// NULL値の安全な処理
	if assigneeID.Valid {
		id := assigneeID.String
		m.AssigneeID = &id
	}
	if dueDate.Valid {
		d := dueDate.Time
		m.DueDate = &d
	}

	return m.ToDomain(), nil
}

// getTaskCount はタスクの総数を取得する（パフォーマンス改善）
func (r *TaskRepository) getTaskCount(ctx context.Context, whereClause string, args []interface{}) (int, error) {
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM "+"`Yotei-Plus`"+".tasks %s", whereClause)

	row, err := r.Query(countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count tasks", logger.Error(err))
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			r.logger.Error("Failed to close count row", logger.Error(closeErr))
		}
	}()

	var count int
	if row.Next() {
		if err := row.Scan(&count); err != nil {
			r.logger.Error("Failed to scan count", logger.Error(err))
			return 0, fmt.Errorf("failed to scan count: %w", err)
		}
	}

	return count, nil
}

// GetTasksForNotification は通知用のタスク取得（効率的なクエリ）
func (r *TaskRepository) GetTasksForNotification(ctx context.Context, from, to time.Time) ([]*domain.Task, error) {
	// 期限が近いアサイン済みタスクのみを効率的に取得
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE due_date BETWEEN ? AND ?
		  AND assignee_id IS NOT NULL
		  AND status IN ('TODO', 'IN_PROGRESS')
		ORDER BY due_date ASC
		LIMIT 1000
	`

	rows, err := r.Query(query, from, to)
	if err != nil {
		r.logger.Error("Failed to get tasks for notification", logger.Error(err))
		return nil, fmt.Errorf("failed to get tasks for notification: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			r.logger.Error("Failed to scan notification task", logger.Error(err))
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Tasks for notification retrieved", logger.Any("count", len(tasks)))
	return tasks, nil
}
