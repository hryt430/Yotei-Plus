package persistence

import (
	"context"
	"time"

	"github.com/hryt430/task-management/internal/modules/task/domain"
)

// TaskService はタスク管理のビジネスロジックを担当するインターフェース
type TaskService interface {
	// タスクの作成
	CreateTask(ctx context.Context, title, description string, priority domain.Priority, createdBy string) (*domain.Task, error)

	// タスクの取得
	GetTask(ctx context.Context, id string) (*domain.Task, error)

	// タスク一覧の取得 (フィルタリング、ソート、ページネーション対応)
	ListTasks(
		ctx context.Context,
		filter domain.ListFilter,
		pagination domain.Pagination,
		sortOptions domain.SortOptions,
	) ([]*domain.Task, int, error)

	// タスクの更新
	UpdateTask(
		ctx context.Context,
		id string,
		title, description *string,
		status *domain.TaskStatus,
		priority *domain.Priority,
		dueDate *time.Time,
	) (*domain.Task, error)

	// タスクの削除
	DeleteTask(ctx context.Context, id string) error

	// タスクの割り当て
	AssignTask(ctx context.Context, taskID string, assigneeID string) (*domain.Task, error)

	// タスクのステータス変更
	ChangeTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) (*domain.Task, error)

	// 期限切れのタスクを取得
	GetOverdueTasks(ctx context.Context) ([]*domain.Task, error)

	// 特定のユーザーに割り当てられたタスクの取得
	GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error)
}
