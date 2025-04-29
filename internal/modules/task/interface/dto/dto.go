package dto

import (
	"time"

	"github.com/hryt430/task-management/internal/modules/task/domain"
)

// TaskModel はPostgreSQLのタスクテーブルにマッピングするための構造体
type TaskModel struct {
	ID          string     `db:"id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	Status      string     `db:"status"`
	Priority    string     `db:"priority"`
	AssigneeID  *string    `db:"assignee_id"`
	CreatedBy   string     `db:"created_by"`
	DueDate     *time.Time `db:"due_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// ToDomain はモデルをドメインエンティティに変換する
func (m *TaskModel) ToDomain() *domain.Task {
	return &domain.Task{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Status:      domain.TaskStatus(m.Status),
		Priority:    domain.Priority(m.Priority),
		AssigneeID:  m.AssigneeID,
		CreatedBy:   m.CreatedBy,
		DueDate:     m.DueDate,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromDomain はドメインエンティティからモデルを作成する
func FromDomain(task *domain.Task) *TaskModel {
	return &TaskModel{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		AssigneeID:  task.AssigneeID,
		CreatedBy:   task.CreatedBy,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
