package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// TaskStatsRepository はタスク統計情報のデータベースリポジトリ実装
type TaskStatsRepository struct {
	SqlHandler
	logger logger.Logger
}

// NewTaskStatsRepository は新しいTaskStatsRepositoryを作成する
func NewTaskStatsRepository(sqlHandler SqlHandler, logger logger.Logger) usecase.StatsRepository {
	return &TaskStatsRepository{
		SqlHandler: sqlHandler,
		logger:     logger,
	}
}

// GetTasksByDateRange は指定された日付範囲のタスクを取得する
func (r *TaskStatsRepository) GetTasksByDateRange(ctx context.Context, userID string, start, end time.Time) ([]*domain.Task, error) {
	if userID == "" {
		return nil, usecase.ErrInvalidParameter
	}

	query := `
		SELECT id, title, description, status, priority, category, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (assignee_id = ? OR created_by = ?)
		  AND (
		    (created_at BETWEEN ? AND ?) OR
		    (due_date BETWEEN ? AND ?) OR
		    (updated_at BETWEEN ? AND ?)
		  )
		ORDER BY created_at DESC
	`

	rows, err := r.Query(query, userID, userID, start, end, start, end, start, end)
	if err != nil {
		r.logger.Error("Failed to get tasks by date range",
			logger.Any("userID", userID),
			logger.Any("start", start),
			logger.Any("end", end),
			logger.Error(err))
		return nil, fmt.Errorf("failed to query tasks by date range: %w", err)
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
			r.logger.Error("Failed to scan task row in date range query", logger.Error(err))
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Tasks retrieved by date range",
		logger.Any("userID", userID),
		logger.Any("count", len(tasks)),
		logger.Any("start", start),
		logger.Any("end", end))

	return tasks, nil
}

// GetTasksByDueDate は指定された期限日のタスクを取得する
func (r *TaskStatsRepository) GetTasksByDueDate(ctx context.Context, userID string, dueDate time.Time) ([]*domain.Task, error) {
	if userID == "" {
		return nil, usecase.ErrInvalidParameter
	}

	// 指定日の00:00:00から23:59:59までのタスクを取得
	dayStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	dayEnd := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 23, 59, 59, 999999999, dueDate.Location())

	query := `
		SELECT id, title, description, status, priority, category, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (assignee_id = ? OR created_by = ?)
		  AND due_date BETWEEN ? AND ?
		ORDER BY due_date ASC, priority DESC
	`

	rows, err := r.Query(query, userID, userID, dayStart, dayEnd)
	if err != nil {
		r.logger.Error("Failed to get tasks by due date",
			logger.Any("userID", userID),
			logger.Any("dueDate", dueDate),
			logger.Error(err))
		return nil, fmt.Errorf("failed to query tasks by due date: %w", err)
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
			r.logger.Error("Failed to scan task row in due date query", logger.Error(err))
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Tasks retrieved by due date",
		logger.Any("userID", userID),
		logger.Any("count", len(tasks)),
		logger.Any("dueDate", dueDate))

	return tasks, nil
}

// GetRecentCompletedTasks は最近完了したタスクを取得する
func (r *TaskStatsRepository) GetRecentCompletedTasks(ctx context.Context, userID string, limit int) ([]*domain.Task, error) {
	if userID == "" {
		return nil, usecase.ErrInvalidParameter
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `
		SELECT id, title, description, status, priority, category, assignee_id, created_by, due_date, created_at, updated_at
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (assignee_id = ? OR created_by = ?)
		  AND status = ?
		ORDER BY updated_at DESC
		LIMIT ?
	`

	rows, err := r.Query(query, userID, userID, string(domain.TaskStatusDone), limit)
	if err != nil {
		r.logger.Error("Failed to get recent completed tasks",
			logger.Any("userID", userID),
			logger.Any("limit", limit),
			logger.Error(err))
		return nil, fmt.Errorf("failed to query recent completed tasks: %w", err)
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
			r.logger.Error("Failed to scan task row in recent completed query", logger.Error(err))
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	r.logger.Debug("Recent completed tasks retrieved",
		logger.Any("userID", userID),
		logger.Any("count", len(tasks)))

	return tasks, nil
}

// GetOverdueTasksCount は期限切れタスク数を取得する
func (r *TaskStatsRepository) GetOverdueTasksCount(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, usecase.ErrInvalidParameter
	}

	now := time.Now()

	query := `
		SELECT COUNT(*)
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (assignee_id = ? OR created_by = ?)
		  AND due_date < ?
		  AND status != ?
	`

	row, err := r.Query(query, userID, userID, now, string(domain.TaskStatusDone))
	if err != nil {
		r.logger.Error("Failed to get overdue tasks count",
			logger.Any("userID", userID),
			logger.Error(err))
		return 0, fmt.Errorf("failed to query overdue tasks count: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			r.logger.Error("Failed to close row", logger.Error(closeErr))
		}
	}()

	var count int
	if row.Next() {
		if err := row.Scan(&count); err != nil {
			r.logger.Error("Failed to scan overdue count", logger.Error(err))
			return 0, fmt.Errorf("failed to scan overdue count: %w", err)
		}
	}

	r.logger.Debug("Overdue tasks count retrieved",
		logger.Any("userID", userID),
		logger.Any("count", count))

	return count, nil
}

// GetTasksCountByStatus はステータス別のタスク数を取得する
func (r *TaskStatsRepository) GetTasksCountByStatus(ctx context.Context, userID string, start, end time.Time) (map[domain.TaskStatus]int, error) {
	if userID == "" {
		return nil, usecase.ErrInvalidParameter
	}

	query := `
		SELECT status, COUNT(*) as count
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (assignee_id = ? OR created_by = ?)
		  AND created_at BETWEEN ? AND ?
		GROUP BY status
	`

	rows, err := r.Query(query, userID, userID, start, end)
	if err != nil {
		r.logger.Error("Failed to get tasks count by status",
			logger.Any("userID", userID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to query tasks count by status: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	counts := make(map[domain.TaskStatus]int)
	for rows.Next() {
		var status string
		var count int

		if err := rows.Scan(&status, &count); err != nil {
			r.logger.Error("Failed to scan status count", logger.Error(err))
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}

		counts[domain.TaskStatus(status)] = count
	}

	return counts, nil
}

// GetTasksCountByCategory はカテゴリ別のタスク数を取得する
func (r *TaskStatsRepository) GetTasksCountByCategory(ctx context.Context, userID string, start, end time.Time) (map[domain.Category]int, error) {
	if userID == "" {
		return nil, usecase.ErrInvalidParameter
	}

	query := `
		SELECT category, COUNT(*) as count
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (assignee_id = ? OR created_by = ?)
		  AND created_at BETWEEN ? AND ?
		GROUP BY category
	`

	rows, err := r.Query(query, userID, userID, start, end)
	if err != nil {
		r.logger.Error("Failed to get tasks count by category",
			logger.Any("userID", userID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to query tasks count by category: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	counts := make(map[domain.Category]int)
	for rows.Next() {
		var category string
		var count int

		if err := rows.Scan(&category, &count); err != nil {
			r.logger.Error("Failed to scan category count", logger.Error(err))
			return nil, fmt.Errorf("failed to scan category count: %w", err)
		}

		counts[domain.Category(category)] = count
	}

	return counts, nil
}

// GetTasksCountByPriority は優先度別のタスク数を取得する
func (r *TaskStatsRepository) GetTasksCountByPriority(ctx context.Context, userID string) (map[domain.Priority]int, error) {
	if userID == "" {
		return nil, usecase.ErrInvalidParameter
	}

	query := `
		SELECT priority, COUNT(*) as count
		FROM ` + "`Yotei-Plus`" + `.tasks
		WHERE (assignee_id = ? OR created_by = ?)
		  AND status != ?
		GROUP BY priority
	`

	rows, err := r.Query(query, userID, userID, string(domain.TaskStatusDone))
	if err != nil {
		r.logger.Error("Failed to get tasks count by priority",
			logger.Any("userID", userID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to query tasks count by priority: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("Failed to close rows", logger.Error(closeErr))
		}
	}()

	counts := make(map[domain.Priority]int)
	for rows.Next() {
		var priority string
		var count int

		if err := rows.Scan(&priority, &count); err != nil {
			r.logger.Error("Failed to scan priority count", logger.Error(err))
			return nil, fmt.Errorf("failed to scan priority count: %w", err)
		}

		counts[domain.Priority(priority)] = count
	}

	return counts, nil
}

// scanTaskFromRow は共通のタスクスキャン処理（TaskRepositoryと重複するが統計用に独立させる）
func (r *TaskStatsRepository) scanTaskFromRow(row Row) (*domain.Task, error) {
	var task domain.Task
	var assigneeID, dueDate, category *string

	err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&category,
		&assigneeID,
		&task.CreatedBy,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan task row: %w", err)
	}

	// NULL値の安全な処理
	if assigneeID != nil {
		task.AssigneeID = assigneeID
	}
	if dueDate != nil {
		if parsedDate, err := time.Parse("2006-01-02 15:04:05", *dueDate); err == nil {
			task.DueDate = &parsedDate
		}
	}
	if category != nil {
		task.Category = domain.Category(*category)
	} else {
		task.Category = domain.CategoryOther // デフォルト値
	}

	return &task, nil
}
