package domain

import (
	"time"
)

// TaskStatus はタスクのステータスを表す型
type TaskStatus string

// タスクステータスの定数
const (
	TaskStatusTodo       TaskStatus = "TODO"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusDone       TaskStatus = "DONE"
)

// Priority はタスクの優先度を表す型
type Priority string

// タスク優先度の定数
const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
)

// Task はタスクのドメインモデルを表す
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Priority    Priority   `json:"priority"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	CreatedBy   string     `json:"created_by"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListFilter はタスク一覧取得時のフィルタを表す
type ListFilter struct {
	Status      *TaskStatus `json:"status,omitempty"`
	Priority    *Priority   `json:"priority,omitempty"`
	AssigneeID  *string     `json:"assignee_id,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	DueDateFrom *time.Time  `json:"due_date_from,omitempty"`
	DueDateTo   *time.Time  `json:"due_date_to,omitempty"`
}

// Pagination はページング情報を表す
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// SortOptions はソートオプションを表す
type SortOptions struct {
	Field     string `json:"field"`     // ソートするフィールド
	Direction string `json:"direction"` // ASC または DESC
}

// NewTask は新しいタスクを作成する
func NewTask(title, description string, priority Priority, createdBy string) *Task {
	now := time.Now()
	return &Task{
		Title:       title,
		Description: description,
		Status:      TaskStatusTodo,
		Priority:    priority,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AssignTo はタスクを特定のユーザーに割り当てる
func (t *Task) AssignTo(userID string) {
	t.AssigneeID = &userID
	t.UpdatedAt = time.Now()
}

// SetStatus はタスクのステータスを設定する
func (t *Task) SetStatus(status TaskStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
}

// SetDueDate はタスクの期限を設定する
func (t *Task) SetDueDate(date time.Time) {
	t.DueDate = &date
	t.UpdatedAt = time.Now()
}

// IsOverdue はタスクが期限切れかどうかを判定する
func (t *Task) IsOverdue() bool {
	return t.DueDate != nil && t.Status != TaskStatusDone && time.Now().After(*t.DueDate)
}
