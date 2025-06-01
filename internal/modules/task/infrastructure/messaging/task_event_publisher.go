package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// NotificationService は通知サービスのインターフェース
type NotificationService interface {
	CreateNotification(ctx context.Context, input input.CreateNotificationInput) (*NotificationDomain, error)
}

// NotificationDomain は通知ドメインのインターフェース（循環参照回避のため）
type NotificationDomain interface {
	GetID() string
	GetUserID() string
	GetTitle() string
}

// TaskEventPublisher は実際に通知を作成するEventPublisher
type TaskEventPublisher struct {
	notificationService NotificationService
	logger              logger.Logger
}

// NewTaskEventPublisher は新しいTaskEventPublisherを作成
func NewTaskEventPublisher(
	notificationService NotificationService,
	logger logger.Logger,
) *TaskEventPublisher {
	return &TaskEventPublisher{
		notificationService: notificationService,
		logger:              logger,
	}
}

// PublishTaskCreated はタスク作成イベントを発行する
func (p *TaskEventPublisher) PublishTaskCreated(ctx context.Context, task *domain.Task) error {
	p.logger.Info("Publishing task created event", logger.Any("taskID", task.ID))

	// タスク作成者には通知を送らない（自分で作成したため）
	// 将来的にはチーム通知などに拡張可能

	return nil
}

// PublishTaskUpdated はタスク更新イベントを発行する
func (p *TaskEventPublisher) PublishTaskUpdated(ctx context.Context, task *domain.Task) error {
	p.logger.Info("Publishing task updated event", logger.Any("taskID", task.ID))

	// タスクが割り当てられている場合、担当者に更新通知を送信
	if task.AssigneeID != nil && *task.AssigneeID != task.CreatedBy {
		return p.createTaskUpdateNotification(ctx, task)
	}

	return nil
}

// PublishTaskDeleted はタスク削除イベントを発行する
func (p *TaskEventPublisher) PublishTaskDeleted(ctx context.Context, taskID string) error {
	p.logger.Info("Publishing task deleted event", logger.Any("taskID", taskID))

	// タスク削除の通知は現在は実装しない
	// 必要に応じて将来実装

	return nil
}

// PublishTaskAssigned はタスク割り当てイベントを発行する
func (p *TaskEventPublisher) PublishTaskAssigned(ctx context.Context, task *domain.Task) error {
	p.logger.Info("Publishing task assigned event", logger.Any("taskID", task.ID))

	if task.AssigneeID == nil {
		return nil
	}

	return p.createTaskAssignedNotification(ctx, task)
}

// PublishTaskCompleted はタスク完了イベントを発行する
func (p *TaskEventPublisher) PublishTaskCompleted(ctx context.Context, task *domain.Task) error {
	p.logger.Info("Publishing task completed event", logger.Any("taskID", task.ID))

	// タスク作成者に完了通知を送信（担当者が異なる場合）
	if task.AssigneeID != nil && *task.AssigneeID != task.CreatedBy {
		return p.createTaskCompletedNotification(ctx, task)
	}

	return nil
}

// PublishTaskOverdue はタスク期限切れイベントを発行する
func (p *TaskEventPublisher) PublishTaskOverdue(ctx context.Context, task *domain.Task) error {
	p.logger.Info("Publishing task overdue event", logger.Any("taskID", task.ID))

	if task.AssigneeID == nil {
		return nil
	}

	return p.createTaskOverdueNotification(ctx, task)
}

// createTaskAssignedNotification はタスク割り当て通知を作成
func (p *TaskEventPublisher) createTaskAssignedNotification(ctx context.Context, task *domain.Task) error {
	title := fmt.Sprintf("新しいタスクが割り当てられました")
	message := fmt.Sprintf(
		"タスク「%s」があなたに割り当てられました。\n\n説明: %s\n優先度: %s",
		task.Title,
		task.Description,
		task.Priority,
	)

	if task.DueDate != nil {
		message += fmt.Sprintf("\n期限: %s", task.DueDate.Format("2006-01-02 15:04"))
	}

	metadata := map[string]string{
		"task_id":           task.ID,
		"task_title":        task.Title,
		"priority":          string(task.Priority),
		"created_by":        task.CreatedBy,
		"notification_type": "task_assigned",
		"action_url":        fmt.Sprintf("/tasks/%s", task.ID),
	}

	if task.DueDate != nil {
		metadata["due_date"] = task.DueDate.Format(time.RFC3339)
	}

	createInput := input.CreateNotificationInput{
		UserID:   *task.AssigneeID,
		Type:     "TASK_ASSIGNED",
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"}, // アプリ内通知
	}

	notification, err := p.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		p.logger.Error("Failed to create task assigned notification",
			logger.Any("taskID", task.ID),
			logger.Error(err))
		return fmt.Errorf("failed to create task assigned notification: %w", err)
	}

	p.logger.Info("Task assigned notification created",
		logger.Any("taskID", task.ID),
		logger.Any("notificationID", (*notification).GetID()),
		logger.Any("assigneeID", *task.AssigneeID))

	return nil
}

// createTaskCompletedNotification はタスク完了通知を作成
func (p *TaskEventPublisher) createTaskCompletedNotification(ctx context.Context, task *domain.Task) error {
	title := fmt.Sprintf("タスクが完了されました")
	message := fmt.Sprintf(
		"タスク「%s」が完了されました。\n\n担当者: %s",
		task.Title,
		*task.AssigneeID, // 実際のプロダクトではユーザー名を取得
	)

	metadata := map[string]string{
		"task_id":           task.ID,
		"task_title":        task.Title,
		"assignee_id":       *task.AssigneeID,
		"completed_at":      time.Now().Format(time.RFC3339),
		"notification_type": "task_completed",
		"action_url":        fmt.Sprintf("/tasks/%s", task.ID),
	}

	createInput := input.CreateNotificationInput{
		UserID:   task.CreatedBy,
		Type:     "TASK_COMPLETED",
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"},
	}

	notification, err := p.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		p.logger.Error("Failed to create task completed notification",
			logger.Any("taskID", task.ID),
			logger.Error(err))
		return fmt.Errorf("failed to create task completed notification: %w", err)
	}

	p.logger.Info("Task completed notification created",
		logger.Any("taskID", task.ID),
		logger.Any("notificationID", (*notification).GetID()),
		logger.Any("createdBy", task.CreatedBy))

	return nil
}

// createTaskUpdateNotification はタスク更新通知を作成
func (p *TaskEventPublisher) createTaskUpdateNotification(ctx context.Context, task *domain.Task) error {
	title := fmt.Sprintf("担当タスクが更新されました")
	message := fmt.Sprintf(
		"あなたが担当するタスク「%s」が更新されました。",
		task.Title,
	)

	metadata := map[string]string{
		"task_id":           task.ID,
		"task_title":        task.Title,
		"updated_by":        task.CreatedBy, // 簡略化、実際は更新者を追跡
		"updated_at":        time.Now().Format(time.RFC3339),
		"notification_type": "task_updated",
		"action_url":        fmt.Sprintf("/tasks/%s", task.ID),
	}

	createInput := input.CreateNotificationInput{
		UserID:   *task.AssigneeID,
		Type:     "TASK_ASSIGNED", // 更新通知も割り当て通知と同じタイプを使用
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"},
	}

	notification, err := p.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		p.logger.Error("Failed to create task update notification",
			logger.Any("taskID", task.ID),
			logger.Error(err))
		return fmt.Errorf("failed to create task update notification: %w", err)
	}

	p.logger.Info("Task update notification created",
		logger.Any("taskID", task.ID),
		logger.Any("notificationID", (*notification).GetID()),
		logger.Any("assigneeID", *task.AssigneeID))

	return nil
}

// createTaskOverdueNotification はタスク期限切れ通知を作成
func (p *TaskEventPublisher) createTaskOverdueNotification(ctx context.Context, task *domain.Task) error {
	title := fmt.Sprintf("⚠️ タスクが期限切れです")
	message := fmt.Sprintf(
		"タスク「%s」の期限が過ぎています。\n\n期限: %s\n優先度: %s",
		task.Title,
		task.DueDate.Format("2006-01-02 15:04"),
		task.Priority,
	)

	metadata := map[string]string{
		"task_id":           task.ID,
		"task_title":        task.Title,
		"due_date":          task.DueDate.Format(time.RFC3339),
		"priority":          string(task.Priority),
		"notification_type": "task_overdue",
		"action_url":        fmt.Sprintf("/tasks/%s", task.ID),
		"urgency":           "high",
	}

	createInput := input.CreateNotificationInput{
		UserID:   *task.AssigneeID,
		Type:     "TASK_DUE_SOON", // 期限切れも期限間近通知と同じタイプ
		Title:    title,
		Message:  message,
		Metadata: metadata,
		Channels: []string{"app"},
	}

	notification, err := p.notificationService.CreateNotification(ctx, createInput)
	if err != nil {
		p.logger.Error("Failed to create task overdue notification",
			logger.Any("taskID", task.ID),
			logger.Error(err))
		return fmt.Errorf("failed to create task overdue notification: %w", err)
	}

	p.logger.Info("Task overdue notification created",
		logger.Any("taskID", task.ID),
		logger.Any("notificationID", (*notification).GetID()),
		logger.Any("assigneeID", *task.AssigneeID))

	return nil
}
