package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		priority    Priority
		category    Category
		createdBy   string
		want        func(*Task) bool
	}{
		{
			name:        "valid task creation",
			title:       "Test Task",
			description: "Test Description",
			priority:    PriorityHigh,
			category:    CategoryWork,
			createdBy:   "user123",
			want: func(task *Task) bool {
				return task.Title == "Test Task" &&
					task.Description == "Test Description" &&
					task.Priority == PriorityHigh &&
					task.Category == CategoryWork &&
					task.CreatedBy == "user123" &&
					task.Status == TaskStatusTodo &&
					task.AssigneeID == nil &&
					!task.CreatedAt.IsZero() &&
					!task.UpdatedAt.IsZero()
			},
		},
		{
			name:        "task with default category",
			title:       "Default Category Task",
			description: "Test",
			priority:    PriorityLow,
			category:    CategoryOther,
			createdBy:   "user456",
			want: func(task *Task) bool {
				return task.Category == CategoryOther
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask(tt.title, tt.description, tt.priority, tt.category, tt.createdBy)
			require.NotNil(t, task)
			assert.True(t, tt.want(task))
		})
	}
}

func TestTask_AssignTo(t *testing.T) {
	task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
	originalUpdatedAt := task.UpdatedAt

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(1 * time.Millisecond)

	userID := "assignee123"
	task.AssignTo(userID)

	assert.Equal(t, &userID, task.AssigneeID)
	assert.True(t, task.UpdatedAt.After(originalUpdatedAt))
}

func TestTask_SetStatus(t *testing.T) {
	tests := []struct {
		name      string
		oldStatus TaskStatus
		newStatus TaskStatus
	}{
		{
			name:      "todo to in progress",
			oldStatus: TaskStatusTodo,
			newStatus: TaskStatusInProgress,
		},
		{
			name:      "in progress to done",
			oldStatus: TaskStatusInProgress,
			newStatus: TaskStatusDone,
		},
		{
			name:      "done to todo",
			oldStatus: TaskStatusDone,
			newStatus: TaskStatusTodo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
			task.Status = tt.oldStatus
			originalUpdatedAt := task.UpdatedAt

			time.Sleep(1 * time.Millisecond)

			task.SetStatus(tt.newStatus)

			assert.Equal(t, tt.newStatus, task.Status)
			assert.True(t, task.UpdatedAt.After(originalUpdatedAt))
		})
	}
}

func TestTask_SetDueDate(t *testing.T) {
	task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
	originalUpdatedAt := task.UpdatedAt

	time.Sleep(1 * time.Millisecond)

	dueDate := time.Now().Add(24 * time.Hour)
	task.SetDueDate(dueDate)

	require.NotNil(t, task.DueDate)
	assert.Equal(t, dueDate, *task.DueDate)
	assert.True(t, task.UpdatedAt.After(originalUpdatedAt))
}

func TestTask_CheckIsOverdue(t *testing.T) {
	tests := []struct {
		name      string
		setupTask func() *Task
		want      bool
	}{
		{
			name: "no due date",
			setupTask: func() *Task {
				return NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
			},
			want: false,
		},
		{
			name: "due date in future",
			setupTask: func() *Task {
				task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
				futureDate := time.Now().Add(24 * time.Hour)
				task.SetDueDate(futureDate)
				return task
			},
			want: false,
		},
		{
			name: "due date in past - todo status",
			setupTask: func() *Task {
				task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
				pastDate := time.Now().Add(-24 * time.Hour)
				task.SetDueDate(pastDate)
				return task
			},
			want: true,
		},
		{
			name: "due date in past - done status",
			setupTask: func() *Task {
				task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
				pastDate := time.Now().Add(-24 * time.Hour)
				task.SetDueDate(pastDate)
				task.SetStatus(TaskStatusDone)
				return task
			},
			want: false,
		},
		{
			name: "due date in past - in progress status",
			setupTask: func() *Task {
				task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
				pastDate := time.Now().Add(-24 * time.Hour)
				task.SetDueDate(pastDate)
				task.SetStatus(TaskStatusInProgress)
				return task
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := tt.setupTask()
			result := task.CheckIsOverdue()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestTask_UpdateIsOverdue(t *testing.T) {
	task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")

	// Initially should not be overdue
	task.UpdateIsOverdue()
	assert.False(t, task.IsOverdue)

	// Set past due date
	pastDate := time.Now().Add(-24 * time.Hour)
	task.SetDueDate(pastDate)
	task.UpdateIsOverdue()
	assert.True(t, task.IsOverdue)

	// Complete the task
	task.SetStatus(TaskStatusDone)
	task.UpdateIsOverdue()
	assert.False(t, task.IsOverdue)
}

func TestTask_PrepareForResponse(t *testing.T) {
	task := NewTask("Test", "Description", PriorityMedium, CategoryWork, "creator")
	pastDate := time.Now().Add(-24 * time.Hour)
	task.SetDueDate(pastDate)

	// IsOverdue might be stale
	task.IsOverdue = false

	task.PrepareForResponse()

	// Should update IsOverdue field
	assert.True(t, task.IsOverdue)
}

func TestTaskSliceHelper_UpdateAllIsOverdue(t *testing.T) {
	tasks := []*Task{
		NewTask("Task1", "Description", PriorityMedium, CategoryWork, "creator"),
		NewTask("Task2", "Description", PriorityHigh, CategoryPersonal, "creator"),
	}

	// Set one task as overdue
	pastDate := time.Now().Add(-24 * time.Hour)
	tasks[0].SetDueDate(pastDate)
	tasks[0].IsOverdue = false // Set stale value

	helper := TaskSliceHelper(tasks)
	helper.UpdateAllIsOverdue()

	assert.True(t, tasks[0].IsOverdue)
	assert.False(t, tasks[1].IsOverdue)
}

func TestPrepareTasksForResponse(t *testing.T) {
	tasks := []*Task{
		NewTask("Task1", "Description", PriorityMedium, CategoryWork, "creator"),
		NewTask("Task2", "Description", PriorityHigh, CategoryPersonal, "creator"),
	}

	// Set one task as overdue
	pastDate := time.Now().Add(-24 * time.Hour)
	tasks[0].SetDueDate(pastDate)
	tasks[0].IsOverdue = false // Set stale value

	PrepareTasksForResponse(tasks)

	assert.True(t, tasks[0].IsOverdue)
	assert.False(t, tasks[1].IsOverdue)
}

func TestPrepareTaskForResponse(t *testing.T) {
	t.Run("valid task", func(t *testing.T) {
		task := NewTask("Task", "Description", PriorityMedium, CategoryWork, "creator")
		pastDate := time.Now().Add(-24 * time.Hour)
		task.SetDueDate(pastDate)
		task.IsOverdue = false // Set stale value

		PrepareTaskForResponse(task)

		assert.True(t, task.IsOverdue)
	})

	t.Run("nil task", func(t *testing.T) {
		// Should not panic
		PrepareTaskForResponse(nil)
	})
}

func TestCategory_GetDisplayName(t *testing.T) {
	tests := []struct {
		category Category
		want     string
	}{
		{CategoryWork, "仕事"},
		{CategoryPersonal, "個人"},
		{CategoryStudy, "学習"},
		{CategoryHealth, "健康"},
		{CategoryShopping, "買い物"},
		{CategoryOther, "その他"},
		{Category("UNKNOWN"), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.category.GetDisplayName())
		})
	}
}

func TestPriority_GetDisplayName(t *testing.T) {
	tests := []struct {
		priority Priority
		want     string
	}{
		{PriorityHigh, "高"},
		{PriorityMedium, "中"},
		{PriorityLow, "低"},
		{Priority("UNKNOWN"), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.priority.GetDisplayName())
		})
	}
}

func TestTaskStatus_GetDisplayName(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   string
	}{
		{TaskStatusTodo, "未着手"},
		{TaskStatusInProgress, "進行中"},
		{TaskStatusDone, "完了"},
		{TaskStatus("UNKNOWN"), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.GetDisplayName())
		})
	}
}

func TestGetAllCategories(t *testing.T) {
	categories := GetAllCategories()

	expected := []Category{
		CategoryWork,
		CategoryPersonal,
		CategoryStudy,
		CategoryHealth,
		CategoryShopping,
		CategoryOther,
	}

	assert.Equal(t, expected, categories)
}

func TestGetAllPriorities(t *testing.T) {
	priorities := GetAllPriorities()

	expected := []Priority{
		PriorityHigh,
		PriorityMedium,
		PriorityLow,
	}

	assert.Equal(t, expected, priorities)
}

func TestGetAllStatuses(t *testing.T) {
	statuses := GetAllStatuses()

	expected := []TaskStatus{
		TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusDone,
	}

	assert.Equal(t, expected, statuses)
}
