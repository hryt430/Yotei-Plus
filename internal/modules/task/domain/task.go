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

// Category はタスクのカテゴリを表す型
type Category string

// タスクカテゴリの定数
const (
	CategoryWork     Category = "WORK"     // 仕事
	CategoryPersonal Category = "PERSONAL" // 個人
	CategoryStudy    Category = "STUDY"    // 学習
	CategoryHealth   Category = "HEALTH"   // 健康
	CategoryShopping Category = "SHOPPING" // 買い物
	CategoryOther    Category = "OTHER"    // その他
)

// Task はタスクのドメインモデルを表す
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Priority    Priority   `json:"priority"`
	Category    Category   `json:"category"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	CreatedBy   string     `json:"created_by"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	IsOverdue   bool       `json:"is_overdue"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListFilter はタスク一覧取得時のフィルタを表す
type ListFilter struct {
	Status      *TaskStatus `json:"status,omitempty"`
	Priority    *Priority   `json:"priority,omitempty"`
	Category    *Category   `json:"category,omitempty"`
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

// NewTask は新しいタスクを作成する（Category引数を追加）
func NewTask(title, description string, priority Priority, category Category, createdBy string) *Task {
	now := time.Now()
	task := &Task{
		Title:       title,
		Description: description,
		Status:      TaskStatusTodo,
		Priority:    priority,
		Category:    category,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	task.UpdateIsOverdue()

	return task
}

// NewTaskWithDefaults はデフォルト値でタスクを作成する（下位互換性）
func NewTaskWithDefaults(title, description string, priority Priority, createdBy string) *Task {
	return NewTask(title, description, priority, CategoryOther, createdBy)
}

// AssignTo はタスクを特定のユーザーに割り当てる
func (t *Task) AssignTo(userID string) {
	t.AssigneeID = &userID
	t.UpdatedAt = time.Now()
	t.UpdateIsOverdue()
}

// SetStatus はタスクのステータスを設定する
func (t *Task) SetStatus(status TaskStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
	t.UpdateIsOverdue()
}

// SetDueDate はタスクの期限を設定する
func (t *Task) SetDueDate(date time.Time) {
	t.DueDate = &date
	t.UpdatedAt = time.Now()
	t.UpdateIsOverdue()
}

// SetCategory はタスクのカテゴリを設定する
func (t *Task) SetCategory(category Category) {
	t.Category = category
	t.UpdatedAt = time.Now()
	t.UpdateIsOverdue()
}

// IsOverdue はタスクが期限切れかどうかを判定する（メソッド版も維持）
func (t *Task) CheckIsOverdue() bool {
	return t.DueDate != nil && t.Status != TaskStatusDone && time.Now().After(*t.DueDate)
}

// UpdateIsOverdue はIsOverdueフィールドを最新の状態に更新する
func (t *Task) UpdateIsOverdue() {
	t.IsOverdue = t.CheckIsOverdue()
}

// PrepareForResponse はレスポンス送信前にフィールドを最新状態に更新する
func (t *Task) PrepareForResponse() {
	t.UpdateIsOverdue()
}

// GetCategoryDisplayName はカテゴリの表示名を取得する
func (c Category) GetDisplayName() string {
	switch c {
	case CategoryWork:
		return "仕事"
	case CategoryPersonal:
		return "個人"
	case CategoryStudy:
		return "学習"
	case CategoryHealth:
		return "健康"
	case CategoryShopping:
		return "買い物"
	case CategoryOther:
		return "その他"
	default:
		return string(c)
	}
}

// GetPriorityDisplayName は優先度の表示名を取得する
func (p Priority) GetDisplayName() string {
	switch p {
	case PriorityHigh:
		return "高"
	case PriorityMedium:
		return "中"
	case PriorityLow:
		return "低"
	default:
		return string(p)
	}
}

// GetStatusDisplayName はステータスの表示名を取得する
func (s TaskStatus) GetDisplayName() string {
	switch s {
	case TaskStatusTodo:
		return "未着手"
	case TaskStatusInProgress:
		return "進行中"
	case TaskStatusDone:
		return "完了"
	default:
		return string(s)
	}
}

// GetAllCategories は利用可能な全カテゴリを取得する
func GetAllCategories() []Category {
	return []Category{
		CategoryWork,
		CategoryPersonal,
		CategoryStudy,
		CategoryHealth,
		CategoryShopping,
		CategoryOther,
	}
}

// GetAllPriorities は利用可能な全優先度を取得する
func GetAllPriorities() []Priority {
	return []Priority{
		PriorityHigh,
		PriorityMedium,
		PriorityLow,
	}
}

// GetAllStatuses は利用可能な全ステータスを取得する
func GetAllStatuses() []TaskStatus {
	return []TaskStatus{
		TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusDone,
	}
}

// TaskSliceHelper はタスクスライス用のヘルパーメソッド
type TaskSliceHelper []*Task

// UpdateAllIsOverdue はスライス内の全タスクのIsOverdueフィールドを更新
func (tasks TaskSliceHelper) UpdateAllIsOverdue() {
	for _, task := range tasks {
		task.UpdateIsOverdue()
	}
}

// PrepareAllForResponse はスライス内の全タスクをレスポンス用に準備
func (tasks TaskSliceHelper) PrepareAllForResponse() {
	for _, task := range tasks {
		task.PrepareForResponse()
	}
}

// タスクリストをレスポンス用に準備するヘルパー関数
func PrepareTasksForResponse(tasks []*Task) {
	helper := TaskSliceHelper(tasks)
	helper.PrepareAllForResponse()
}

// 単一タスクをレスポンス用に準備するヘルパー関数
func PrepareTaskForResponse(task *Task) {
	if task != nil {
		task.PrepareForResponse()
	}
}
