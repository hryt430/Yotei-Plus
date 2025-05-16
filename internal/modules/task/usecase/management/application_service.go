package management

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase/persistence"
)

// エラー定義
var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrInvalidParameter = errors.New("invalid parameter")
)

// taskService はタスク管理のビジネスロジックを実装
type taskService struct {
	taskRepo TaskRepository
}

// NewTaskService は新しいTaskServiceのインスタンスを作成
func NewTaskService(taskRepo TaskRepository) persistence.TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

// CreateTask はタスクを作成する
func (s *taskService) CreateTask(
	ctx context.Context,
	title,
	description string,
	priority domain.Priority,
	createdBy string,
) (*domain.Task, error) {
	if title == "" {
		return nil, ErrInvalidParameter
	}

	task := domain.NewTask(title, description, priority, createdBy)
	task.ID = uuid.New().String()

	err := s.taskRepo.Save(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask はIDに基づいてタスクを取得する
func (s *taskService) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// ListTasks はタスク一覧を取得する
func (s *taskService) ListTasks(
	ctx context.Context,
	filter domain.ListFilter,
	pagination domain.Pagination,
	sortOptions domain.SortOptions,
) ([]*domain.Task, int, error) {
	// デフォルト値の設定
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}

	// ソートオプションのデフォルト値
	if sortOptions.Field == "" {
		sortOptions.Field = "created_at"
	}
	if sortOptions.Direction == "" {
		sortOptions.Direction = "DESC"
	}

	return s.taskRepo.List(ctx, filter, pagination, sortOptions)
}

// UpdateTask はタスクを更新する
func (s *taskService) UpdateTask(
	ctx context.Context,
	id string,
	title, description *string,
	status *domain.TaskStatus,
	priority *domain.Priority,
	dueDate *time.Time,
) (*domain.Task, error) {
	if id == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, ErrTaskNotFound
	}

	// 各フィールドの更新（指定されている場合のみ）
	if title != nil {
		task.Title = *title
	}
	if description != nil {
		task.Description = *description
	}
	if status != nil {
		task.Status = *status
	}
	if priority != nil {
		task.Priority = *priority
	}
	if dueDate != nil {
		task.DueDate = dueDate
	}

	task.UpdatedAt = time.Now()

	err = s.taskRepo.Save(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask はタスクを削除する
func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidParameter
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if task == nil {
		return ErrTaskNotFound
	}

	return s.taskRepo.Delete(ctx, id)
}

// AssignTask はタスクを指定されたユーザーに割り当てる
func (s *taskService) AssignTask(ctx context.Context, taskID string, assigneeID string) (*domain.Task, error) {
	if taskID == "" || assigneeID == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, ErrTaskNotFound
	}

	task.AssignTo(assigneeID)

	err = s.taskRepo.Save(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// ChangeTaskStatus はタスクのステータスを変更する
func (s *taskService) ChangeTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) (*domain.Task, error) {
	if taskID == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, ErrTaskNotFound
	}

	task.SetStatus(status)

	err = s.taskRepo.Save(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetOverdueTasks は期限切れのタスクを取得する
func (s *taskService) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	return s.taskRepo.GetOverdueTasks(ctx)
}

// GetTasksByAssignee は特定のユーザーに割り当てられたタスクを取得する
func (s *taskService) GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error) {
	if userID == "" {
		return nil, ErrInvalidParameter
	}

	return s.taskRepo.GetTasksByAssignee(ctx, userID)
}
