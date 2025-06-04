package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type UserValidator = commonDomain.UserValidator

// EventPublisher はイベント発行のインターフェース
type EventPublisher interface {
	PublishTaskCreated(ctx context.Context, task *domain.Task) error
	PublishTaskUpdated(ctx context.Context, task *domain.Task) error
	PublishTaskDeleted(ctx context.Context, taskID string) error
	PublishTaskAssigned(ctx context.Context, task *domain.Task) error
	PublishTaskCompleted(ctx context.Context, task *domain.Task) error
}

// === 構造体定義 ===

// / UserInfo はユーザーの基本情報（共通定義を使用）
type UserInfo = commonDomain.UserInfo

// TaskWithUserInfo はタスクとユーザー情報を含む構造体（N+1問題解決用）
type TaskWithUserInfo struct {
	Task         *domain.Task `json:"task"`
	CreatorInfo  *UserInfo    `json:"creator_info,omitempty"`
	AssigneeInfo *UserInfo    `json:"assignee_info,omitempty"`
}

// TaskService は改良されたタスクサービス
type TaskService struct {
	TaskRepository TaskRepository
	UserValidator  UserValidator
	EventPublisher EventPublisher
	Logger         logger.Logger

	// 非同期イベント設定
	AsyncEventTimeout time.Duration
	MaxRetries        int
}

// NewTaskService はTaskServiceのコンストラクタ
func NewTaskService(
	taskRepo TaskRepository,
	userValidator UserValidator,
	eventPublisher EventPublisher,
	logger logger.Logger,
) *TaskService {
	return &TaskService{
		TaskRepository:    taskRepo,
		UserValidator:     userValidator,
		EventPublisher:    eventPublisher,
		Logger:            logger,
		AsyncEventTimeout: 30 * time.Second,
		MaxRetries:        3,
	}
}

// === エラー定義 ===

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrInvalidParameter    = errors.New("invalid parameter")
	ErrUserNotFound        = errors.New("user not found")
	ErrDuplicateAssignment = errors.New("task already assigned to this user")
)

// === メインサービスメソッド ===

// CreateTask はタスクを作成する（統一インターフェース使用）
func (s *TaskService) CreateTask(
	ctx context.Context,
	title,
	description string,
	priority domain.Priority,
	category domain.Category,
	createdBy string,
) (*domain.Task, error) {
	// 入力バリデーション
	if err := s.validateCreateTaskInput(title, description, createdBy); err != nil {
		return nil, err
	}

	// 作成者の存在確認（統一インターフェース使用）
	exists, err := s.UserValidator.UserExists(ctx, createdBy)
	if err != nil {
		s.Logger.Error("Failed to validate user existence",
			logger.Any("userID", createdBy), logger.Error(err))
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// タスク作成
	task := domain.NewTask(title, description, priority, category, createdBy)
	task.ID = uuid.New().String()

	err = s.TaskRepository.CreateTask(ctx, task)
	if err != nil {
		s.Logger.Error("Failed to create task",
			logger.Any("taskID", task.ID), logger.Error(err))
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// イベント発行（非同期）
	s.publishEventAsync(ctx, "task_created", func() error {
		return s.EventPublisher.PublishTaskCreated(ctx, task)
	})

	s.Logger.Info("Task created successfully",
		logger.Any("taskID", task.ID), logger.Any("createdBy", createdBy))

	return task, nil
}

// CreateTaskWithDefaults はデフォルトカテゴリでタスクを作成する（下位互換性）
func (s *TaskService) CreateTaskWithDefaults(
	ctx context.Context,
	title,
	description string,
	priority domain.Priority,
	createdBy string,
) (*domain.Task, error) {
	return s.CreateTask(ctx, title, description, priority, domain.CategoryOther, createdBy)
}

// GetTask はIDに基づいてタスクを取得する
func (s *TaskService) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		return nil, ErrInvalidParameter
	}
	return s.TaskRepository.GetTaskByID(ctx, id)
}

// GetTaskWithUserInfo はタスクとユーザー情報を一緒に取得（統一インターフェース使用）
func (s *TaskService) GetTaskWithUserInfo(ctx context.Context, id string) (*TaskWithUserInfo, error) {
	if id == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.TaskRepository.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &TaskWithUserInfo{
		Task: task,
	}

	// ユーザー情報を一括取得（N+1問題解決）
	userIDs := []string{task.CreatedBy}
	if task.AssigneeID != nil {
		userIDs = append(userIDs, *task.AssigneeID)
	}

	userInfoMap, err := s.UserValidator.GetUsersInfoBatch(ctx, userIDs)
	if err != nil {
		s.Logger.Error("Failed to get user info batch", logger.Error(err))
		// エラーでもタスク情報は返す（ユーザー情報は空）
	} else {
		result.CreatorInfo = userInfoMap[task.CreatedBy]
		if task.AssigneeID != nil {
			result.AssigneeInfo = userInfoMap[*task.AssigneeID]
		}
	}

	return result, nil
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
	if pagination.PageSize > 100 {
		pagination.PageSize = 100 // 大量データ防止
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

// ListTasksWithUserInfo はタスク一覧とユーザー情報を一緒に取得（N+1問題解決）
func (s *TaskService) ListTasksWithUserInfo(
	ctx context.Context,
	filter domain.ListFilter,
	pagination domain.Pagination,
	sortOptions domain.SortOptions,
) ([]*TaskWithUserInfo, int, error) {
	// タスク一覧を取得
	tasks, total, err := s.ListTasks(ctx, filter, pagination, sortOptions)
	if err != nil {
		return nil, 0, err
	}

	if len(tasks) == 0 {
		return []*TaskWithUserInfo{}, total, nil
	}

	// ユーザーIDを収集（重複除去でN+1問題解決）
	userIDSet := make(map[string]bool)
	for _, task := range tasks {
		userIDSet[task.CreatedBy] = true
		if task.AssigneeID != nil {
			userIDSet[*task.AssigneeID] = true
		}
	}

	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	// ユーザー情報を一括取得（修正：UserInfoで統一）
	userInfoMap := make(map[string]*UserInfo)
	if len(userIDs) > 0 {
		if batchInfo, err := s.UserValidator.GetUsersInfoBatch(ctx, userIDs); err == nil {
			userInfoMap = batchInfo
		} else {
			s.Logger.Error("Failed to get user info batch", logger.Error(err))
			// エラーログは出力するが、処理は継続（グレースフルな劣化）
		}
	}

	// 結果を組み立て
	result := make([]*TaskWithUserInfo, len(tasks))
	for i, task := range tasks {
		result[i] = &TaskWithUserInfo{
			Task:        task,
			CreatorInfo: userInfoMap[task.CreatedBy],
		}
		if task.AssigneeID != nil {
			result[i].AssigneeInfo = userInfoMap[*task.AssigneeID]
		}
	}

	return result, total, nil
}

// UpdateTask はタスクを更新する（イベント発行）
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

	// 更新内容のバリデーション
	if err := s.validateUpdateTaskInput(title, description); err != nil {
		return nil, err
	}

	task, err := s.TaskRepository.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 変更追跡
	hasChanges := false
	oldStatus := task.Status

	// 各フィールドの更新（指定されている場合のみ）
	if title != nil && *title != task.Title {
		task.Title = *title
		hasChanges = true
	}
	if description != nil && *description != task.Description {
		task.Description = *description
		hasChanges = true
	}
	if status != nil && *status != task.Status {
		task.Status = *status
		hasChanges = true
	}
	if priority != nil && *priority != task.Priority {
		task.Priority = *priority
		hasChanges = true
	}
	if dueDate != nil {
		if task.DueDate == nil || !dueDate.Equal(*task.DueDate) {
			task.DueDate = dueDate
			hasChanges = true
		}
	}

	// 変更がない場合は早期リターン
	if !hasChanges {
		return task, nil
	}

	task.UpdatedAt = time.Now()

	err = s.TaskRepository.UpdateTask(ctx, task)
	if err != nil {
		s.Logger.Error("Failed to update task",
			logger.Any("taskID", id), logger.Error(err))
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// イベント発行（非同期）
	s.publishEventAsync(ctx, "task_updated", func() error {
		return s.EventPublisher.PublishTaskUpdated(ctx, task)
	})

	// 完了状態になった場合の追加イベント
	if oldStatus != domain.TaskStatusDone && task.Status == domain.TaskStatusDone {
		s.publishEventAsync(ctx, "task_completed", func() error {
			return s.EventPublisher.PublishTaskCompleted(ctx, task)
		})
	}

	s.Logger.Info("Task updated successfully", logger.Any("taskID", id))
	return task, nil
}

// DeleteTask はタスクを削除する（イベント発行）
func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidParameter
	}

	// 存在確認
	_, err := s.TaskRepository.GetTaskByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.TaskRepository.DeleteTask(ctx, id)
	if err != nil {
		s.Logger.Error("Failed to delete task",
			logger.Any("taskID", id), logger.Error(err))
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// イベント発行（非同期）
	s.publishEventAsync(ctx, "task_deleted", func() error {
		return s.EventPublisher.PublishTaskDeleted(ctx, id)
	})

	s.Logger.Info("Task deleted successfully", logger.Any("taskID", id))
	return nil
}

// / AssignTask はタスクを指定されたユーザーに割り当てる（統一インターフェース使用）
func (s *TaskService) AssignTask(ctx context.Context, taskID string, assigneeID string) (*domain.Task, error) {
	if taskID == "" || assigneeID == "" {
		return nil, ErrInvalidParameter
	}

	// アサイン先ユーザーの存在確認（統一インターフェース使用）
	exists, err := s.UserValidator.UserExists(ctx, assigneeID)
	if err != nil {
		s.Logger.Error("Failed to validate assignee existence",
			logger.Any("assigneeID", assigneeID), logger.Error(err))
		return nil, fmt.Errorf("failed to validate assignee: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	task, err := s.TaskRepository.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 既に同じユーザーにアサインされているかチェック
	if task.AssigneeID != nil && *task.AssigneeID == assigneeID {
		return nil, ErrDuplicateAssignment
	}

	task.AssignTo(assigneeID)

	err = s.TaskRepository.UpdateTask(ctx, task)
	if err != nil {
		s.Logger.Error("Failed to update task assignment",
			logger.Any("taskID", taskID), logger.Error(err))
		return nil, fmt.Errorf("failed to update task assignment: %w", err)
	}

	// イベント発行（非同期）
	s.publishEventAsync(ctx, "task_assigned", func() error {
		return s.EventPublisher.PublishTaskAssigned(ctx, task)
	})

	s.Logger.Info("Task assigned successfully",
		logger.Any("taskID", taskID), logger.Any("assigneeID", assigneeID))

	return task, nil
}

// ChangeTaskStatus はタスクのステータスを変更する（イベント発行）
func (s *TaskService) ChangeTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) (*domain.Task, error) {
	if taskID == "" {
		return nil, ErrInvalidParameter
	}

	task, err := s.TaskRepository.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldStatus := task.Status
	task.SetStatus(status)

	err = s.TaskRepository.UpdateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	// イベント発行（非同期）
	s.publishEventAsync(ctx, "task_updated", func() error {
		return s.EventPublisher.PublishTaskUpdated(ctx, task)
	})

	// 完了状態になった場合の追加イベント
	if oldStatus != domain.TaskStatusDone && status == domain.TaskStatusDone {
		s.publishEventAsync(ctx, "task_completed", func() error {
			return s.EventPublisher.PublishTaskCompleted(ctx, task)
		})
	}

	return task, nil
}

// === その他のメソッド ===

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
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.TaskRepository.SearchTasks(ctx, query, limit)
}

// === 非同期イベント発行メソッド ===

// publishEventAsync はイベントを非同期で発行する
func (s *TaskService) publishEventAsync(ctx context.Context, eventType string, publishFunc func() error) {
	go func() {
		// タイムアウト付きコンテキスト
		_, cancel := context.WithTimeout(ctx, s.AsyncEventTimeout)
		defer cancel()

		var lastErr error
		for attempt := 1; attempt <= s.MaxRetries; attempt++ {
			if err := publishFunc(); err != nil {
				lastErr = err
				s.Logger.Warn("Event publish failed, retrying...",
					logger.Any("eventType", eventType),
					logger.Any("attempt", attempt),
					logger.Error(err))

				// 指数バックオフで再試行
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			// 成功
			s.Logger.Debug("Event published successfully",
				logger.Any("eventType", eventType),
				logger.Any("attempt", attempt))
			return
		}

		// 全ての試行が失敗
		s.Logger.Error("Event publish failed after all retries",
			logger.Any("eventType", eventType),
			logger.Any("maxRetries", s.MaxRetries),
			logger.Error(lastErr))
	}()
}

// === バリデーション関数 ===

func (s *TaskService) validateCreateTaskInput(title, description, createdBy string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidParameter)
	}
	if len(title) > 255 {
		return fmt.Errorf("%w: title too long (max 255 characters)", ErrInvalidParameter)
	}
	if len(description) > 2000 {
		return fmt.Errorf("%w: description too long (max 2000 characters)", ErrInvalidParameter)
	}
	if strings.TrimSpace(createdBy) == "" {
		return fmt.Errorf("%w: createdBy is required", ErrInvalidParameter)
	}
	return nil
}

func (s *TaskService) validateUpdateTaskInput(title, description *string) error {
	if title != nil {
		if strings.TrimSpace(*title) == "" {
			return ErrInvalidParameter
		}
		if len(*title) > 255 {
			return errors.New("title too long (max 255 characters)")
		}
	}
	if description != nil && len(*description) > 5000 {
		return errors.New("description too long (max 5000 characters)")
	}
	return nil
}
