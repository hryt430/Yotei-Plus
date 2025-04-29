package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"task-management-service/internal/core/domain"
	"task-management-service/internal/core/ports"
	"task-management-service/internal/infrastructure/persistence/models"
)

var (
	// ErrTaskNotFound はタスクが見つからない場合のエラー
	ErrTaskNotFound = errors.New("task not found")
)

// PostgresTaskRepository はPostgreSQLを使用したTaskRepositoryの実装
type PostgresTaskRepository struct {
	db *sqlx.DB
}

// NewTaskRepository は新しいPostgresTaskRepositoryを作成する
func NewTaskRepository(db *sqlx.DB) ports.TaskRepository {
	return &PostgresTaskRepository{
		db: db,
	}
}

// Save はタスクを保存する（作成または更新）
func (r *PostgresTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (
			id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		) VALUES (
			:id, :title, :description, :status, :priority, :assignee_id, :created_by, :due_date, :created_at, :updated_at
		) ON CONFLICT (id) DO UPDATE SET
			title = :title,
			description = :description,
			status = :status,
			priority = :priority,
			assignee_id = :assignee_id,
			due_date = :due_date,
			updated_at = :updated_at
	`

	taskModel := models.FromDomain(task)
	_, err := r.db.NamedExecContext(ctx, query, taskModel)
	return err
}

// GetByID はIDによりタスクを取得する
func (r *PostgresTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `SELECT * FROM tasks WHERE id = $1`
	
	var taskModel models.TaskModel
	err := r.db.GetContext(ctx, &taskModel, query, id)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	return taskModel.ToDomain(), nil
}

// List はタスク一覧を取得する（フィルタリング、ソート、ページネーション対応）
func (r *PostgresTaskRepository) List(
	ctx context.Context, 
	filter domain.ListFilter, 
	pagination domain.Pagination, 
	sortOptions domain.SortOptions,
) ([]*domain.Task, int, error) {
	// WHERE句の条件を構築
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}
	
	if filter.Priority != nil {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, string(*filter.Priority))
		argIndex++
	}
	
	if filter.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("assignee_id = $%d", argIndex))
		args = append(args, *filter.AssigneeID)
		argIndex++
	}
	
	if filter.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *filter.CreatedBy)
		argIndex++
	}
	
	if filter.DueDateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("due_date >= $%d", argIndex))
		args = append(args, *filter.DueDateFrom)
		argIndex++
	}
	
	if filter.DueDateTo != nil {
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argIndex))
		args = append(args, *filter.DueDateTo)
		argIndex++
	}
	
	// WHERE句の構築
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	// 総件数のカウント
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", whereClause)
	var totalCount int
	err := r.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	// ソートとページネーションを追加
	orderClause := fmt.Sprintf("ORDER BY %s %s", sortOptions.Field, sortOptions.Direction)
	limitOffset := fmt.Sprintf("LIMIT %d OFFSET %d", pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
	
	// 最終的なクエリ
	query := fmt.Sprintf("SELECT * FROM tasks %s %s %s", whereClause, orderClause, limitOffset)
	
	// タスク一覧の取得
	var taskModels []models.TaskModel
	err = r.db.SelectContext(ctx, &taskModels, query, args...)
	if err != nil {
		return nil, 0, err
	}
	
	// ドメインモデルに変換
	tasks := make([]*domain.Task, len(taskModels))
	for i, model := range taskModels {
		tasks[i] = model.ToDomain()
	}
	
	return tasks, totalCount, nil
}

// Delete はタスクを削除する
func (r *PostgresTaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Search はクエリ文字列によりタスクを検索する
func (r *PostgresTaskRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Task, error) {
	sqlQuery := `
		SELECT * FROM tasks 
		WHERE 
			title ILIKE $1 OR 
			description ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	
	searchPattern := "%" + query + "%"
	
	var taskModels []models.TaskModel