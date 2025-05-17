package management

import (
	"context"

	"github.com/hryt430/task-management/internal/modules/task/domain"
)

// TaskRepository はタスク永続化のためのインターフェース
type TaskRepository interface {
	// タスクの保存 (作成または更新)
	Save(ctx context.Context, task *domain.Task) error

	// IDによるタスクの取得
	GetByID(ctx context.Context, id string) (*domain.Task, error)

	// タスク一覧の取得 (フィルタリング、ソート、ページネーション対応)
	List(
		ctx context.Context,
		filter domain.ListFilter,
		pagination domain.Pagination,
		sortOptions domain.SortOptions,
	) ([]*domain.Task, int, error)

	// IDによるタスクの削除
	Delete(ctx context.Context, id string) error

	// 複数の条件によるタスクの検索
	Search(ctx context.Context, query string, limit int) ([]*domain.Task, error)

	// 特定のユーザーに割り当てられたタスクの取得
	GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error)

	// 期限切れのタスクの取得
	GetOverdueTasks(ctx context.Context) ([]*domain.Task, error)
}
