package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// TaskDueNotificationScheduler はタスク期限通知のスケジューラー（改良版）
type TaskDueNotificationScheduler struct {
	taskService         usecase.TaskService
	notificationService NotificationService
	eventPublisher      *TaskEventPublisher
	logger              logger.Logger
	ticker              *time.Ticker
	stopCh              chan struct{}
	isRunning           bool
}

// NewTaskDueNotificationScheduler は新しいスケジューラーを作成
func NewTaskDueNotificationScheduler(
	taskService usecase.TaskService,
	notificationService NotificationService,
	eventPublisher *TaskEventPublisher,
	logger logger.Logger,
) *TaskDueNotificationScheduler {
	return &TaskDueNotificationScheduler{
		taskService:         taskService,
		notificationService: notificationService,
		eventPublisher:      eventPublisher,
		logger:              logger,
		stopCh:              make(chan struct{}),
	}
}

// Start はスケジューラーを開始（1時間ごとにチェック）
func (s *TaskDueNotificationScheduler) Start(ctx context.Context) {
	if s.isRunning {
		s.logger.Warn("Task due notification scheduler already running")
		return
	}

	s.isRunning = true
	s.ticker = time.NewTicker(1 * time.Hour) // 1時間ごとにチェック

	s.logger.Info("Starting task due notification scheduler")

	// 初回実行
	go s.checkAndNotifyDueTasks(ctx)
	go s.checkAndNotifyOverdueTasks(ctx)

	go func() {
		defer func() {
			s.ticker.Stop()
			s.isRunning = false
		}()

		for {
			select {
			case <-s.ticker.C:
				s.checkAndNotifyDueTasks(ctx)
				s.checkAndNotifyOverdueTasks(ctx)
			case <-s.stopCh:
				s.logger.Info("Task due notification scheduler stopped")
				return
			case <-ctx.Done():
				s.logger.Info("Task due notification scheduler stopped due to context cancellation")
				return
			}
		}
	}()
}

// checkAndNotifyDueTasks は12時間以内に期限を迎えるタスクをチェックして通知
func (s *TaskDueNotificationScheduler) checkAndNotifyDueTasks(ctx context.Context) {
	s.logger.Info("Checking tasks due within 12 hours")

	now := time.Now()
	twelveHoursLater := now.Add(12 * time.Hour)

	// 期限が12時間以内のタスクを取得
	tasks, err := s.getTasksDueWithin12Hours(ctx, now, twelveHoursLater)
	if err != nil {
		s.logger.Error("Failed to get tasks due within 12 hours", logger.Error(err))
		return
	}

	s.logger.Info("Found tasks due within 12 hours", logger.Any("count", len(tasks)))

	// 各タスクについて通知を作成
	for _, task := range tasks {
		if err := s.createDueNotification(ctx, task, now); err != nil {
			s.logger.Error("Failed to create due notification",
				logger.Any("taskID", task.ID),
				logger.Error(err))
			continue
		}
	}
}

// checkAndNotifyOverdueTasks は期限切れタスクをチェックして通知
func (s *TaskDueNotificationScheduler) checkAndNotifyOverdueTasks(ctx context.Context) {
	s.logger.Info("Checking overdue tasks")

	// 期限切れタスクを取得
	tasks, err := s.taskService.GetOverdueTasks(ctx)
	if err != nil {
		s.logger.Error("Failed to get overdue tasks", logger.Error(err))
		return
	}

	s.logger.Info("Found overdue tasks", logger.Any("count", len(tasks)))

	// 各期限切れタスクについて通知を作成
	for _, task := range tasks {
		if task.AssigneeID != nil {
			if err := s.eventPublisher.PublishTaskOverdue(ctx, task); err != nil {
				s.logger.Error("Failed to publish task overdue event",
					logger.Any("taskID", task.ID),
					logger.Error(err))
				continue
			}
		}
	}
}

// getTasksDueWithin12Hours は12時間以内に期限を迎えるタスクを取得
func (s *TaskDueNotificationScheduler) getTasksDueWithin12Hours(ctx context.Context, from, to time.Time) ([]*domain.Task, error) {
	// 期限でフィルタリング
	filter := domain.ListFilter{
		DueDateFrom: &from,
		DueDateTo:   &to,
	}

	pagination := domain.Pagination{
		Page:     1,
		PageSize: 1000,
	}

	sortOptions := domain.SortOptions{
		Field:     "due_date",
		Direction: "ASC",
	}

	// 完了していないタスクのみを対象
	tasks, _, err := s.taskService.ListTasks(ctx, filter, pagination, sortOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// 完了していない、かつ割り当てられているタスクのみフィルタリング
	var dueTasks []*domain.Task
	for _, task := range tasks {
		if task.Status != domain.TaskStatusDone &&
			task.AssigneeID != nil &&
			task.DueDate != nil &&
			s.shouldNotifyForTask(task, from, to) {
			dueTasks = append(dueTasks, task)
		}
	}

	return dueTasks, nil
}

// shouldNotifyForTask はタスクに対して通知すべきかを判断
func (s *TaskDueNotificationScheduler) shouldNotifyForTask(task *domain.Task, from, to time.Time) bool {
	if task.DueDate == nil {
		return false
	}

	// 12時間以内に期限を迎える
	return task.DueDate.After(from) && task.DueDate.Before(to)
}

// createDueNotification は期限通知を作成
func (s *TaskDueNotificationScheduler) createDueNotification(ctx context.Context, task *domain.Task, now time.Time) error {
	if task.AssigneeID == nil {
		return nil
	}

	// 期限までの時間を計算
	timeUntilDue := task.DueDate.Sub(now)
	hoursUntilDue := int(timeUntilDue.Hours())

	title := fmt.Sprintf("⏰ タスク期限通知")
	message := fmt.Sprintf(
		"タスク「%s」の期限まであと%d時間です。\n\n期限: %s\n優先度: %s",
		task.Title,
		hoursUntilDue,
		task.DueDate.Format("2006-01-02 15:04"),
		task.Priority,
	)

	metadata := map[string]string{
		"task_id":           task.ID,
		"task_title":        task.Title,
		"due_date":          task.DueDate.Format(time.RFC3339),
		"hours_until":       fmt.Sprintf("%d", hoursUntilDue),
		"priority":          string(task.Priority),
		"notification_type": "task_due_soon",
		"action_url":        fmt.Sprintf("/tasks/%s", task.ID),
	}

	createInput := input.CreateNotificationInput{
		UserID:   *task.AssigneeID,
		Type:     "TASK_DUE_SOON",
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"},
	}

	notification, err := s.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("Created due notification",
		logger.Any("taskID", task.ID),
		logger.Any("notificationID", notification.GetID()),
		logger.Any("assigneeID", *task.AssigneeID))

	return nil
}

// Stop はスケジューラーを停止
func (s *TaskDueNotificationScheduler) Stop() {
	if !s.isRunning {
		return
	}

	close(s.stopCh)
	s.logger.Info("Stopping task due notification scheduler")
}
