package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
)

// EventType はイベントの種類を表す型
type EventType string

const (
	// イベントタイプの定義
	EventTaskCreated   EventType = "task.created"
	EventTaskUpdated   EventType = "task.updated"
	EventTaskDeleted   EventType = "task.deleted"
	EventTaskAssigned  EventType = "task.assigned"
	EventTaskCompleted EventType = "task.completed"
	EventTaskOverdue   EventType = "task.overdue"
)

// TaskEvent はタスク関連のイベントを表す構造体
type TaskEvent struct {
	Type      EventType   `json:"type"`
	TaskID    string      `json:"task_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// EventPublisher はイベントを発行するインターフェース
type EventPublisher interface {
	PublishTaskCreated(ctx context.Context, task *domain.Task) error
	PublishTaskUpdated(ctx context.Context, task *domain.Task) error
	PublishTaskDeleted(ctx context.Context, taskID string) error
	PublishTaskAssigned(ctx context.Context, task *domain.Task) error
	PublishTaskCompleted(ctx context.Context, task *domain.Task) error
	PublishTaskOverdue(ctx context.Context, task *domain.Task) error
}

// LogEventPublisher はログにイベントを出力するシンプルな実装
type LogEventPublisher struct{}

// NewLogEventPublisher は新しいLogEventPublisherを作成する
func NewLogEventPublisher() EventPublisher {
	return &LogEventPublisher{}
}

// PublishTaskCreated はタスク作成イベントを発行する
func (p *LogEventPublisher) PublishTaskCreated(ctx context.Context, task *domain.Task) error {
	return p.publishEvent(ctx, EventTaskCreated, task.ID, task)
}

// PublishTaskUpdated はタスク更新イベントを発行する
func (p *LogEventPublisher) PublishTaskUpdated(ctx context.Context, task *domain.Task) error {
	return p.publishEvent(ctx, EventTaskUpdated, task.ID, task)
}

// PublishTaskDeleted はタスク削除イベントを発行する
func (p *LogEventPublisher) PublishTaskDeleted(ctx context.Context, taskID string) error {
	return p.publishEvent(ctx, EventTaskDeleted, taskID, nil)
}

// PublishTaskAssigned はタスク割り当てイベントを発行する
func (p *LogEventPublisher) PublishTaskAssigned(ctx context.Context, task *domain.Task) error {
	data := map[string]interface{}{
		"task_id":     task.ID,
		"assignee_id": *task.AssigneeID,
	}
	return p.publishEvent(ctx, EventTaskAssigned, task.ID, data)
}

// PublishTaskCompleted はタスク完了イベントを発行する
func (p *LogEventPublisher) PublishTaskCompleted(ctx context.Context, task *domain.Task) error {
	return p.publishEvent(ctx, EventTaskCompleted, task.ID, task)
}

// PublishTaskOverdue はタスク期限切れイベントを発行する
func (p *LogEventPublisher) PublishTaskOverdue(ctx context.Context, task *domain.Task) error {
	return p.publishEvent(ctx, EventTaskOverdue, task.ID, task)
}

// publishEvent はイベントをシリアライズしてログに出力する
func (p *LogEventPublisher) publishEvent(ctx context.Context, eventType EventType, taskID string, data interface{}) error {
	event := TaskEvent{
		Type:      eventType,
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data:      data,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 実際のメッセージキューやイベントバスへの送信は将来的に実装
	// 現在はログに出力するだけ
	log.Printf("[EVENT] %s", string(eventJSON))
	return nil
}
