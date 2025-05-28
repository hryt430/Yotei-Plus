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
)

// TaskRepository はタスクのデータベースリポジトリ実装
type TaskRepository struct {
	SqlHandler
}

// NewTaskRepository は新しいTaskRepositoryを作成する
func NewTaskRepository(sqlHandler SqlHandler) usecase.TaskRepository {
	return &TaskRepository{
		SqlHandler: sqlHandler,
	}
}

// 許可されたソートフィールド（SQLインジェクション対策）
var allowedSortFields = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"title":      "title",
	"priority":   "priority",
	"status":     "status",
	"due_date":   "due_date",
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
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// UpdateTask はタスクを更新する
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
		return fmt.Errorf("failed to update task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return usecase.ErrTaskNotFound
	}

	return nil
}

// GetTaskByID はIDによりタスクを取得する
func (r *TaskRepository) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.tasks 
		WHERE id = ?
	`

	row, err := r.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}
	defer row.Close()

	if !row.Next() {
		return nil, usecase.ErrTaskNotFound
	}

	task, err := r.scanTaskFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	return task, nil
}

// DeleteTask はタスクを削除する
func (r *TaskRepository) DeleteTask(ctx context.Context, id string) error {
	query := `DELETE FROM ` + "`Yotei-Plus`" + `.tasks WHERE id = ?`

	result, err := r.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return usecase.ErrTaskNotFound
	}

	return nil
}

// ListTasks はタスク一覧を取得する（フィルタ・ソート・ページネーション対応）
func (r *TaskRepository) ListTasks(
	ctx context.Context,
	filter domain.ListFilter,
	pagination domain.Pagination,
	sort domain.SortOptions,
) ([]*domain.Task, int, error) {
	var conds []string
	var args []interface{}

	// フィルタ条件の構築
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

	// カウント取得
	total, err := r.getTaskCount(whereClause, args)
	if err != nil {
		return nil, 0, err
	}

	// ソートフィールドのバリデーション
	sortField, ok := allowedSortFields[sort.Field]
	if !ok {
		sortField = "created_at"
	}

	// ソート方向のバリデーション
	sortDirection := "DESC"
	if sort.Direction == "ASC" || sort.Direction == "DESC" {
		sortDirection = sort.Direction
	}

	// メインクエリ
	query := fmt.Sprintf(`
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM `+"`Yotei-Plus`"+`.tasks
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortField, sortDirection)

	// ページネーション用のパラメータを追加
	args = append(args, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)

	rows, err := r.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

// SearchTasks はタイトルまたは説明に対する検索
func (r *TaskRepository) SearchTasks(ctx context.Context, query string, limit int) ([]*domain.Task, error) {
	sqlQuery := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE title LIKE ? OR description LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	pattern := "%" + query + "%"
	rows, err := r.Query(sqlQuery, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByAssignee は特定のユーザーに割り当てられたタスクを取得
func (r *TaskRepository) GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE assignee_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by assignee: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetOverdueTasks は期限切れのタスクを取得
func (r *TaskRepository) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE due_date < ? AND status != ?
		ORDER BY due_date ASC
	`

	now := time.Now()
	doneStatus := string(domain.TaskStatusDone)

	rows, err := r.Query(query, now, doneStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTaskFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// scanTaskFromRow はRowからTaskをスキャンする共通処理
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
		return nil, err
	}

	// NULL値の処理
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

// getTaskCount はタスクの総数を取得する
func (r *TaskRepository) getTaskCount(whereClause string, args []interface{}) (int, error) {
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM "+"`Yotei-Plus`"+".tasks %s", whereClause)

	row, err := r.Query(countQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	defer row.Close()

	var count int
	if row.Next() {
		if err := row.Scan(&count); err != nil {
			return 0, fmt.Errorf("failed to scan count: %w", err)
		}
	}

	return count, nil
}
