package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
)

type TaskService struct {
	TaskRepository TaskRepository
}

func NewTaskService(taskRepo TaskRepository) *TaskService {
	return &TaskService{
		TaskRepository: taskRepo,
	}
}

// エラー定義
var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrInvalidParameter = errors.New("invalid parameter")
)

// CreateTask はタスクを作成する
func (s *TaskService) CreateTask(
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

	err := s.TaskRepository.CreateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask はIDに基づいてタスクを取得する
func (s *TaskService) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		return nil, ErrInvalidParameter
	}

	return s.TaskRepository.GetTaskByID(ctx, id)
}

// ListTasks はタスク一覧を取得する
func (s *TaskService) ListTasks(
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

	return s.TaskRepository.ListTasks(ctx, filter, pagination, sortOptions)
}

// UpdateTask はタスクを更新する
func (s *TaskService) UpdateTask(
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

	task, err := s.TaskRepository.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
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

	err = s.TaskRepository.UpdateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask はタスクを削除する
func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidParameter
	}

	// 存在確認のために取得を試行
	_, err := s.TaskRepository.GetTaskByID(ctx, id)
	if err != nil {
		return err
	}

	return s.TaskRepository.DeleteTask(ctx, id)
}

// AssignTask はタスクを指定されたユーザーに割り当てる
func (s *TaskService) AssignTask(ctx context.Context, taskID string, assigneeID string) (*domain.Task, error) {
	if taskID == "" || assigneeID == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.TaskRepository.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.AssignTo(assigneeID)

	err = s.TaskRepository.UpdateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// ChangeTaskStatus はタスクのステータスを変更する
func (s *TaskService) ChangeTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) (*domain.Task, error) {
	if taskID == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.TaskRepository.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.SetStatus(status)

	err = s.TaskRepository.UpdateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetOverdueTasks は期限切れのタスクを取得する
func (s *TaskService) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	return s.TaskRepository.GetOverdueTasks(ctx)
}

// GetTasksByAssignee は特定のユーザーに割り当てられたタスクを取得する
func (s *TaskService) GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error) {
	if userID == "" {
		return nil, ErrInvalidParameter
	}

	return s.TaskRepository.GetTasksByAssignee(ctx, userID)
}

// SearchTasks はタスクを検索する
func (s *TaskService) SearchTasks(ctx context.Context, query string, limit int) ([]*domain.Task, error) {
	if query == "" {
		return nil, ErrInvalidParameter
	}

	if limit <= 0 {
		limit = 20
	}

	return s.TaskRepository.SearchTasks(ctx, query, limit)
}
