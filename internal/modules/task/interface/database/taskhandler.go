package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hryt430/task-management/internal/modules/task/domain"
	"github.com/hryt430/task-management/internal/modules/task/infrastructure/database"
	"github.com/hryt430/task-management/internal/modules/task/infrastructure/database/models"
	"github.com/hryt430/task-management/internal/modules/task/usecase/management"
)

var (
	// ErrTaskNotFound はタスクが見つからない場合のエラー
	ErrTaskNotFound = errors.New("task not found")
)

// MySQLTaskRepository はMySQLを使用したTaskRepositoryの実装
type MySQLTaskRepository struct {
	handler database.SqlHandler
}

// NewTaskRepository は新しいMySQLTaskRepositoryを作成する
func NewTaskRepository(handler database.SqlHandler) management.TaskRepository {
	return &MySQLTaskRepository{
		handler: handler,
	}
}

// Save はタスクを保存する（作成または更新）
func (r *MySQLTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (
			id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			description = VALUES(description),
			status = VALUES(status),
			priority = VALUES(priority),
			assignee_id = VALUES(assignee_id),
			due_date = VALUES(due_date),
			updated_at = VALUES(updated_at)
	`

	model := models.FromDomain(task)
	_, err := r.handler.Execute(query,
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
	return err
}

// GetByID はIDによりタスクを取得する
func (r *MySQLTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at FROM tasks WHERE id = ?`

	row, err := r.handler.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, nil
	}

	var model models.TaskModel
	var assigneeID sql.NullString
	var dueDate sql.NullTime

	err = row.Scan(
		&model.ID,
		&model.Title,
		&model.Description,
		&model.Status,
		&model.Priority,
		&assigneeID,
		&model.CreatedBy,
		&dueDate,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if assigneeID.Valid {
		assigneeIDStr := assigneeID.String
		model.AssigneeID = &assigneeIDStr
	}

	if dueDate.Valid {
		dueDateVal := dueDate.Time
		model.DueDate = &dueDateVal
	}

	return model.ToDomain(), nil
}

// List はタスク一覧を取得する（フィルタリング、ソート、ページネーション対応）
func (r *MySQLTaskRepository) List(
	ctx context.Context,
	filter domain.ListFilter,
	pagination domain.Pagination,
	sortOptions domain.SortOptions,
) ([]*domain.Task, int, error) {
	// WHERE句の条件を構築
	var conditions []string
	var args []interface{}

	if filter.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, string(*filter.Status))
	}

	if filter.Priority != nil {
		conditions = append(conditions, "priority = ?")
		args = append(args, string(*filter.Priority))
	}

	if filter.AssigneeID != nil {
		conditions = append(conditions, "assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}

	if filter.CreatedBy != nil {
		conditions = append(conditions, "created_by = ?")
		args = append(args, *filter.CreatedBy)
	}

	if filter.DueDateFrom != nil {
		conditions = append(conditions, "due_date >= ?")
		args = append(args, *filter.DueDateFrom)
	}

	if filter.DueDateTo != nil {
		conditions = append(conditions, "due_date <= ?")
		args = append(args, *filter.DueDateTo)
	}

	// WHERE句の構築
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 総件数のカウント
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", whereClause)
	countRow, err := r.handler.Query(countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer countRow.Close()

	var totalCount int
	if countRow.Next() {
		if err := countRow.Scan(&totalCount); err != nil {
			return nil, 0, err
		}
	}

	// ソートとページネーションを追加
	orderClause := fmt.Sprintf("ORDER BY %s %s", sortOptions.Field, sortOptions.Direction)
	limitOffset := fmt.Sprintf("LIMIT %d OFFSET %d", pagination.PageSize, (pagination.Page-1)*pagination.PageSize)

	// 最終的なクエリ
	query := fmt.Sprintf("SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at FROM tasks %s %s %s", whereClause, orderClause, limitOffset)

	// タスク一覧の取得
	rows, err := r.handler.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*domain.Task

	for rows.Next() {
		var model models.TaskModel
		var assigneeID sql.NullString
		var dueDate sql.NullTime

		err = rows.Scan(
			&model.ID,
			&model.Title,
			&model.Description,
			&model.Status,
			&model.Priority,
			&assigneeID,
			&model.CreatedBy,
			&dueDate,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		if assigneeID.Valid {
			assigneeIDStr := assigneeID.String
			model.AssigneeID = &assigneeIDStr
		}

		if dueDate.Valid {
			dueDateVal := dueDate.Time
			model.DueDate = &dueDateVal
		}

		tasks = append(tasks, model.ToDomain())
	}

	return tasks, totalCount, nil
}

// Delete はタスクを削除する
func (r *MySQLTaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := r.handler.Execute(query, id)
	return err
}

// Search はクエリ文字列によりタスクを検索する
func (r *MySQLTaskRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Task, error) {
	sqlQuery := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks 
		WHERE 
			title LIKE ? OR 
			description LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	searchPattern := "%" + query + "%"

	rows, err := r.handler.Query(sqlQuery, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task

	for rows.Next() {
		var model models.TaskModel
		var assigneeID sql.NullString
		var dueDate sql.NullTime

		err = rows.Scan(
			&model.ID,
			&model.Title,
			&model.Description,
			&model.Status,
			&model.Priority,
			&assigneeID,
			&model.CreatedBy,
			&dueDate,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			assigneeIDStr := assigneeID.String
			model.AssigneeID = &assigneeIDStr
		}

		if dueDate.Valid {
			dueDateVal := dueDate.Time
			model.DueDate = &dueDateVal
		}

		tasks = append(tasks, model.ToDomain())
	}

	return tasks, nil
}

// GetTasksByAssignee は特定のユーザーに割り当てられたタスクを取得する
func (r *MySQLTaskRepository) GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks 
		WHERE assignee_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.handler.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task

	for rows.Next() {
		var model models.TaskModel
		var assigneeID sql.NullString
		var dueDate sql.NullTime

		err = rows.Scan(
			&model.ID,
			&model.Title,
			&model.Description,
			&model.Status,
			&model.Priority,
			&assigneeID,
			&model.CreatedBy,
			&dueDate,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			assigneeIDStr := assigneeID.String
			model.AssigneeID = &assigneeIDStr
		}

		if dueDate.Valid {
			dueDateVal := dueDate.Time
			model.DueDate = &dueDateVal
		}

		tasks = append(tasks, model.ToDomain())
	}

	return tasks, nil
}

// GetOverdueTasks は期限切れのタスクを取得する
func (r *MySQLTaskRepository) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks 
		WHERE 
			due_date < ? AND 
			status != ?
		ORDER BY due_date ASC
	`

	now := time.Now()
	doneStatus := string(domain.TaskStatusDone)

	rows, err := r.handler.Query(query, now, doneStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task

	for rows.Next() {
		var model models.TaskModel
		var assigneeID sql.NullString
		var dueDate sql.NullTime

		err = rows.Scan(
			&model.ID,
			&model.Title,
			&model.Description,
			&model.Status,
			&model.Priority,
			&assigneeID,
			&model.CreatedBy,
			&dueDate,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			assigneeIDStr := assigneeID.String
			model.AssigneeID = &assigneeIDStr
		}

		if dueDate.Valid {
			dueDateVal := dueDate.Time
			model.DueDate = &dueDateVal
		}

		tasks = append(tasks, model.ToDomain())
	}

	return tasks, nil
}
