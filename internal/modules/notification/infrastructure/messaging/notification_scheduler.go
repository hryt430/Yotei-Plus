package messaging

import (
	"context"
	"fmt"
	"time"

	notiDomain "github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	taskDomain "github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// NotificationService は通知サービスのインターフェース
type NotificationService interface {
	CreateScheduledNotification(
		ctx context.Context,
		userID, title, message string,
		notificationType notiDomain.NotificationType,
		scheduledTime time.Time,
		metadata map[string]string,
	) error
	CreateNotification(ctx context.Context, input input.CreateNotificationInput) (*notiDomain.Notification, error)
}

// TaskDueNotificationScheduler はタスク期限通知のスケジューラー
type TaskDueNotificationScheduler struct {
	taskService         usecase.TaskService
	notificationService NotificationService
	logger              logger.Logger
	ticker              *time.Ticker
	stopCh              chan struct{}
	isRunning           bool
}

// NewTaskDueNotificationScheduler は新しいスケジューラーを作成
func NewTaskDueNotificationScheduler(
	taskService usecase.TaskService,
	notificationService NotificationService,
	logger logger.Logger,
) *TaskDueNotificationScheduler {
	return &TaskDueNotificationScheduler{
		taskService:         taskService,
		notificationService: notificationService,
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

	go func() {
		defer func() {
			s.ticker.Stop()
			s.isRunning = false
		}()

		for {
			select {
			case <-s.ticker.C:
				s.checkAndNotifyDueTasks(ctx)
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

// Stop はスケジューラーを停止
func (s *TaskDueNotificationScheduler) Stop() {
	if !s.isRunning {
		return
	}

	close(s.stopCh)
	s.logger.Info("Stopping task due notification scheduler")
}

// checkAndNotifyDueTasks は12時間以内に期限を迎えるタスクをチェックして通知
func (s *TaskDueNotificationScheduler) checkAndNotifyDueTasks(ctx context.Context) {
	s.logger.Info("Checking tasks due within 12 hours")

	// 現在時刻から12時間後までのタスクを取得
	now := time.Now()
	twelveHoursLater := now.Add(12 * time.Hour)

	// カスタムフィルターでタスクを取得（期限が12時間以内）
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

// getTasksDueWithin12Hours は12時間以内に期限を迎えるタスクを取得
func (s *TaskDueNotificationScheduler) getTasksDueWithin12Hours(ctx context.Context, from, to time.Time) ([]*taskDomain.Task, error) {
	// TaskServiceにカスタムフィルターメソッドを追加する必要がある
	// ここでは仮の実装として、すべてのタスクを取得してフィルタリング
	filter := taskDomain.ListFilter{
		DueDateFrom: &from,
		DueDateTo:   &to,
	}

	pagination := taskDomain.Pagination{
		Page:     1,
		PageSize: 1000, // 大きめの値を設定（実際のプロダクションではより適切な値に調整）
	}

	sortOptions := taskDomain.SortOptions{
		Field:     "due_date",
		Direction: "ASC",
	}

	// まだ完了していないタスクのみを対象
	incompleteTasks := []*taskDomain.Task{}

	tasks, _, err := s.taskService.ListTasks(ctx, filter, pagination, sortOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// 完了していないタスクのみフィルタリング
	for _, task := range tasks {
		if task.Status != taskDomain.TaskStatusDone && task.DueDate != nil {
			// 12時間以内かつ、まだ通知していないタスクをフィルタ
			if s.shouldNotifyForTask(task, from, to) {
				incompleteTasks = append(incompleteTasks, task)
			}
		}
	}

	return incompleteTasks, nil
}

// shouldNotifyForTask はタスクに対して通知すべきかを判断
func (s *TaskDueNotificationScheduler) shouldNotifyForTask(task *taskDomain.Task, from, to time.Time) bool {
	if task.DueDate == nil {
		return false
	}

	// 12時間以内に期限を迎える
	if task.DueDate.After(from) && task.DueDate.Before(to) {
		return true
	}

	return false
}

// createDueNotification は期限通知を作成
func (s *TaskDueNotificationScheduler) createDueNotification(ctx context.Context, task *taskDomain.Task, now time.Time) error {
	// アサインされていないタスクは通知しない
	if task.AssigneeID == nil {
		return nil
	}

	// 期限までの時間を計算
	timeUntilDue := task.DueDate.Sub(now)
	hoursUntilDue := int(timeUntilDue.Hours())

	title := fmt.Sprintf("タスク期限通知: %s", task.Title)
	message := fmt.Sprintf(
		"タスク「%s」の期限まであと%d時間です。\n期限: %s\n優先度: %s",
		task.Title,
		hoursUntilDue,
		task.DueDate.Format("2006-01-02 15:04"),
		task.Priority,
	)

	// メタデータを設定
	metadata := map[string]string{
		"task_id":           task.ID,
		"task_title":        task.Title,
		"due_date":          task.DueDate.Format(time.RFC3339),
		"hours_until":       fmt.Sprintf("%d", hoursUntilDue),
		"priority":          string(task.Priority),
		"notification_type": "task_due_soon",
	}

	// 通知作成の入力データ
	createInput := input.CreateNotificationInput{
		UserID:   *task.AssigneeID,
		Type:     "TASK_DUE_SOON",
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"}, // アプリ内通知のみ
	}

	// 通知を作成
	notification, err := s.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("Created due notification",
		logger.Any("taskID", task.ID),
		logger.Any("notificationID", notification.ID),
		logger.Any("assigneeID", *task.AssigneeID))

	return nil
}

// TaskEventHandler はタスクイベントを処理
type TaskEventHandler struct {
	notificationService NotificationService
	logger              logger.Logger
}

// NewTaskEventHandler は新しいイベントハンドラーを作成
func NewTaskEventHandler(
	notificationService NotificationService,
	logger logger.Logger,
) *TaskEventHandler {
	return &TaskEventHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// HandleTaskAssigned はタスク割り当てイベントを処理
func (h *TaskEventHandler) HandleTaskAssigned(ctx context.Context, task *taskDomain.Task) error {
	if task.AssigneeID == nil {
		return nil
	}

	title := fmt.Sprintf("新しいタスクが割り当てられました: %s", task.Title)
	message := fmt.Sprintf(
		"タスク「%s」が割り当てられました。\n説明: %s\n優先度: %s",
		task.Title,
		task.Description,
		task.Priority,
	)

	if task.DueDate != nil {
		message += fmt.Sprintf("\n期限: %s", task.DueDate.Format("2006-01-02 15:04"))
	}

	metadata := map[string]string{
		"task_id":    task.ID,
		"task_title": task.Title,
		"priority":   string(task.Priority),
		"created_by": task.CreatedBy,
	}

	createInput := input.CreateNotificationInput{
		UserID:   *task.AssigneeID,
		Type:     "TASK_ASSIGNED",
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"},
	}

	_, err := h.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create task assigned notification: %w", err)
	}

	h.logger.Info("Created task assigned notification",
		logger.Any("taskID", task.ID),
		logger.Any("assigneeID", *task.AssigneeID))

	return nil
}

// HandleTaskCompleted はタスク完了イベントを処理
func (h *TaskEventHandler) HandleTaskCompleted(ctx context.Context, task *taskDomain.Task) error {
	// タスク作成者に完了通知を送信
	title := fmt.Sprintf("タスクが完了されました: %s", task.Title)
	message := fmt.Sprintf(
		"タスク「%s」が完了されました。",
		task.Title,
	)

	if task.AssigneeID != nil {
		message += fmt.Sprintf("\n担当者: %s", *task.AssigneeID)
	}

	metadata := map[string]string{
		"task_id":      task.ID,
		"task_title":   task.Title,
		"completed_at": time.Now().Format(time.RFC3339),
	}

	if task.AssigneeID != nil {
		metadata["assignee_id"] = *task.AssigneeID
	}

	createInput := input.CreateNotificationInput{
		UserID:   task.CreatedBy,
		Type:     "TASK_COMPLETED",
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"},
	}

	_, err := h.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create task completed notification: %w", err)
	}

	h.logger.Info("Created task completed notification",
		logger.Any("taskID", task.ID),
		logger.Any("createdBy", task.CreatedBy))

	return nil
}
