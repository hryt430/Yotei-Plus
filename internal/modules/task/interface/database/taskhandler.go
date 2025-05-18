package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/interface/dto"
)

var (
	// ErrTaskNotFound はタスクが見つからない場合のエラー
	ErrTaskNotFound = errors.New("task not found")
)

type UserServiceRepository struct {
	SqlHandler
}

// Save はタスクを保存する（作成または更新）
func (r *UserServiceRepository) Save(ctx context.Context, task *domain.Task) error {
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
	return err
}

// GetByID は ID によりタスクを取得する
func (r *UserServiceRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at FROM tasks WHERE id = ?`

	row, err := r.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, ErrTaskNotFound
	}

	var m dto.TaskModel
	var assigneeID sql.NullString
	var dueDate sql.NullTime

	err = row.Scan(
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
		return nil, fmt.Errorf("Scan失敗: %w", err)
	}

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

// List はタスク一覧を取得する（フィルタ・ソート・ページネーション対応）
func (r *UserServiceRepository) List(
	ctx context.Context,
	filter domain.ListFilter,
	pagination domain.Pagination,
	sort domain.SortOptions,
) ([]*domain.Task, int, error) {
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

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	// カウント
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", where)
	cr, err := r.Query(countQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer cr.Close()

	var total int
	if cr.Next() {
		if err := cr.Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	order := fmt.Sprintf("ORDER BY %s %s", sort.Field, sort.Direction)
	limit := fmt.Sprintf("LIMIT %d OFFSET %d", pagination.PageSize, (pagination.Page-1)*pagination.PageSize)

	q := fmt.Sprintf(
		"SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at FROM tasks %s %s %s",
		where, order, limit,
	)
	rows, err := r.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*domain.Task
	for rows.Next() {
		var m dto.TaskModel
		var asn sql.NullString
		var due sql.NullTime

		err = rows.Scan(
			&m.ID, &m.Title, &m.Description, &m.Status,
			&m.Priority, &asn, &m.CreatedBy, &due,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if asn.Valid {
			s := asn.String
			m.AssigneeID = &s
		}
		if due.Valid {
			dt := due.Time
			m.DueDate = &dt
		}
		list = append(list, m.ToDomain())
	}
	return list, total, nil
}

// Delete はタスクを削除する
func (r *UserServiceRepository) Delete(ctx context.Context, id string) error {
	_, err := r.Execute("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// Search はタイトル or 説明に対する全文検索
func (r *UserServiceRepository) Search(ctx context.Context, q string, limit int) ([]*domain.Task, error) {
	sqlQ := `
		SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks
		WHERE title LIKE ? OR description LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	pattern := "%" + q + "%"
	rows, err := r.Query(sqlQ, pattern, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*domain.Task
	for rows.Next() {
		var m dto.TaskModel
		var asn sql.NullString
		var due sql.NullTime

		err = rows.Scan(
			&m.ID, &m.Title, &m.Description, &m.Status,
			&m.Priority, &asn, &m.CreatedBy, &due,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if asn.Valid {
			s := asn.String
			m.AssigneeID = &s
		}
		if due.Valid {
			dt := due.Time
			m.DueDate = &dt
		}
		res = append(res, m.ToDomain())
	}
	return res, nil
}

// GetTasksByAssignee はユーザー単位の一覧取得
func (r *UserServiceRepository) GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error) {
	rows, err := r.Query(`SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at FROM tasks WHERE assignee_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*domain.Task
	for rows.Next() {
		var m dto.TaskModel
		var asn sql.NullString
		var due sql.NullTime

		err = rows.Scan(
			&m.ID, &m.Title, &m.Description, &m.Status,
			&m.Priority, &asn, &m.CreatedBy, &due,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if asn.Valid {
			s := asn.String
			m.AssigneeID = &s
		}
		if due.Valid {
			dt := due.Time
			m.DueDate = &dt
		}
		res = append(res, m.ToDomain())
	}
	return res, nil
}

// GetOverdueTasks は期限切れタスクを取得
func (r *UserServiceRepository) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	now := time.Now()
	done := string(domain.TaskStatusDone)
	rows, err := r.Query(`SELECT id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at FROM tasks WHERE due_date < ? AND status != ? ORDER BY due_date ASC`, now, done)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*domain.Task
	for rows.Next() {
		var m dto.TaskModel
		var asn sql.NullString
		var due sql.NullTime

		err = rows.Scan(
			&m.ID, &m.Title, &m.Description, &m.Status,
			&m.Priority, &asn, &m.CreatedBy, &due,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if asn.Valid {
			s := asn.String
			m.AssigneeID = &s
		}
		if due.Valid {
			dt := due.Time
			m.DueDate = &dt
		}
		res = append(res, m.ToDomain())
	}
	return res, nil
}
