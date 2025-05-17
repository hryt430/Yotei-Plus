package events

import "time"

// EventType はイベントの種類を定義します
type EventType string

const (
	// UserRegistered はユーザー登録イベントを表します
	UserRegistered EventType = "user.registered"
	// UserLoggedIn はユーザーログインイベントを表します
	UserLoggedIn EventType = "user.logged_in"
	// TaskCreated はタスク作成イベントを表します
	TaskCreated EventType = "task.created"
	// TaskAssigned はタスク割り当てイベントを表します
	TaskAssigned EventType = "task.assigned"
	// TaskStatusChanged はタスクステータス変更イベントを表します
	TaskStatusChanged EventType = "task.status_changed"
	// NotificationSent は通知送信イベントを表します
	NotificationSent EventType = "notification.sent"
	// NotificationRead は通知既読イベントを表します
	NotificationRead EventType = "notification.read"
)

// Event はシステム内で発生するイベントを表します
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	CreatedAt time.Time   `json:"created_at"`
}

// TaskCreatedPayload はタスク作成イベントのペイロードを表します
type TaskCreatedPayload struct {
	TaskID      string    `json:"task_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// TaskAssignedPayload はタスク割り当てイベントのペイロードを表します
type TaskAssignedPayload struct {
	TaskID       string    `json:"task_id"`
	Title        string    `json:"title"`
	AssignedToID string    `json:"assigned_to_id"`
	AssignedBy   string    `json:"assigned_by"`
	AssignedAt   time.Time `json:"assigned_at"`
}

// TaskStatusChangedPayload はタスクステータス変更イベントのペイロードを表します
type TaskStatusChangedPayload struct {
	TaskID    string    `json:"task_id"`
	Title     string    `json:"title"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	ChangedBy string    `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
}

// NotificationPayload は通知イベントのペイロードを表します
type NotificationPayload struct {
	NotificationID string    `json:"notification_id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	Type           string    `json:"type"`
	ReferenceID    string    `json:"reference_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserEventPayload はユーザー関連イベントのペイロードを表します
type UserEventPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	EventTime time.Time `json:"event_time"`
}
