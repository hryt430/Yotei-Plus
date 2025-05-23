package usecase

import (
	"context"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
)

// TaskRepository はタスク永続化のためのインターフェース
type TaskRepository interface {
	// タスクの作成
	CreateTask(ctx context.Context, task *domain.Task) error

	// タスクの取得
	GetTaskByID(ctx context.Context, id string) (*domain.Task, error)

	// タスク一覧の取得 (フィルタリング、ソート、ページネーション対応)
	ListTasks(
		ctx context.Context,
		filter domain.ListFilter,
		pagination domain.Pagination,
		sortOptions domain.SortOptions,
	) ([]*domain.Task, int, error)

	// タスクの更新
	UpdateTask(ctx context.Context, task *domain.Task) error

	// タスクの削除
	DeleteTask(ctx context.Context, id string) error

	// 期限切れのタスクを取得
	GetOverdueTasks(ctx context.Context) ([]*domain.Task, error)

	// 特定のユーザーに割り当てられたタスクの取得
	GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error)

	// タスクの検索
	SearchTasks(ctx context.Context, query string, limit int) ([]*domain.Task, error)
}
