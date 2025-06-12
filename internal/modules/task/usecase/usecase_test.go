package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	mock_usecase "github.com/hryt430/Yotei+/internal/modules/task/usecase/mocks"
	"github.com/hryt430/Yotei+/pkg/logger"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go
//go:generate mockgen -source=service.go -destination=mocks/mock_interfaces.go

func TestTaskService_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockTaskRepository(ctrl)
	mockUserValidator := mock_usecase.NewMockUserValidator(ctrl)
	mockEventPublisher := mock_usecase.NewMockEventPublisher(ctrl)
	mockLogger := logger.NewMockLogger()

	service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, mockLogger)

	tests := []struct {
		name          string
		title         string
		description   string
		priority      domain.Priority
		category      domain.Category
		createdBy     string
		setupMocks    func()
		expectedError error
	}{
		{
			name:        "successful task creation",
			title:       "Test Task",
			description: "Test Description",
			priority:    domain.PriorityHigh,
			category:    domain.CategoryWork,
			createdBy:   "user123",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user123").
					Return(true, nil)

				mockRepo.EXPECT().
					CreateTask(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, task *domain.Task) {
						// Verify task properties
						assert.Equal(t, "Test Task", task.Title)
						assert.Equal(t, "Test Description", task.Description)
						assert.Equal(t, domain.PriorityHigh, task.Priority)
						assert.Equal(t, domain.CategoryWork, task.Category)
						assert.Equal(t, "user123", task.CreatedBy)
						assert.Equal(t, domain.TaskStatusTodo, task.Status)
						assert.NotEmpty(t, task.ID)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskCreated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "empty title",
			title:       "",
			description: "Test Description",
			priority:    domain.PriorityHigh,
			category:    domain.CategoryWork,
			createdBy:   "user123",
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: ErrInvalidParameter,
		},
		{
			name:        "empty created by",
			title:       "Test Task",
			description: "Test Description",
			priority:    domain.PriorityHigh,
			category:    domain.CategoryWork,
			createdBy:   "",
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: ErrInvalidParameter,
		},
		{
			name:        "user not found",
			title:       "Test Task",
			description: "Test Description",
			priority:    domain.PriorityHigh,
			category:    domain.CategoryWork,
			createdBy:   "nonexistent",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "nonexistent").
					Return(false, nil)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:        "user validation error",
			title:       "Test Task",
			description: "Test Description",
			priority:    domain.PriorityHigh,
			category:    domain.CategoryWork,
			createdBy:   "user123",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user123").
					Return(false, errors.New("validation error"))
			},
			expectedError: errors.New("failed to validate user"),
		},
		{
			name:        "repository error",
			title:       "Test Task",
			description: "Test Description",
			priority:    domain.PriorityHigh,
			category:    domain.CategoryWork,
			createdBy:   "user123",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user123").
					Return(true, nil)

				mockRepo.EXPECT().
					CreateTask(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: errors.New("failed to create task"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			task, err := service.CreateTask(
				context.Background(),
				tt.title,
				tt.description,
				tt.priority,
				tt.category,
				tt.createdBy,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, tt.title, task.Title)
				assert.Equal(t, tt.createdBy, task.CreatedBy)
			}
		})
	}
}

func TestTaskService_GetTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockTaskRepository(ctrl)
	mockUserValidator := mock_usecase.NewMockUserValidator(ctrl)
	mockEventPublisher := mock_usecase.NewMockEventPublisher(ctrl)
	mockLogger := logger.NewMockLogger()

	service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, mockLogger)

	tests := []struct {
		name          string
		taskID        string
		setupMocks    func()
		expectedTask  *domain.Task
		expectedError error
	}{
		{
			name:   "successful get task",
			taskID: "task123",
			setupMocks: func() {
				task := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Test Description",
					Status:      domain.TaskStatusTodo,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(task, nil)
			},
			expectedTask: &domain.Task{
				ID:          "task123",
				Title:       "Test Task",
				Description: "Test Description",
			},
			expectedError: nil,
		},
		{
			name:   "empty task ID",
			taskID: "",
			setupMocks: func() {
				// No mocks needed
			},
			expectedTask:  nil,
			expectedError: ErrInvalidParameter,
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			setupMocks: func() {
				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "nonexistent").
					Return(nil, ErrTaskNotFound)
			},
			expectedTask:  nil,
			expectedError: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			task, err := service.GetTask(context.Background(), tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, tt.expectedTask.ID, task.ID)
				assert.Equal(t, tt.expectedTask.Title, task.Title)
			}
		})
	}
}

func TestTaskService_AssignTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockTaskRepository(ctrl)
	mockUserValidator := mock_usecase.NewMockUserValidator(ctrl)
	mockEventPublisher := mock_usecase.NewMockEventPublisher(ctrl)
	mockLogger := logger.NewMockLogger()

	service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, mockLogger)

	tests := []struct {
		name          string
		taskID        string
		assigneeID    string
		setupMocks    func()
		expectedError error
	}{
		{
			name:       "successful assignment",
			taskID:     "task123",
			assigneeID: "user456",
			setupMocks: func() {
				task := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Test Description",
					Status:      domain.TaskStatusTodo,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					AssigneeID:  nil,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user456").
					Return(true, nil)

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(task, nil)

				mockRepo.EXPECT().
					UpdateTask(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updatedTask *domain.Task) {
						assert.NotNil(t, updatedTask.AssigneeID)
						assert.Equal(t, "user456", *updatedTask.AssigneeID)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskAssigned(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "empty task ID",
			taskID:     "",
			assigneeID: "user456",
			setupMocks: func() {
				// No mocks needed
			},
			expectedError: ErrInvalidParameter,
		},
		{
			name:       "empty assignee ID",
			taskID:     "task123",
			assigneeID: "",
			setupMocks: func() {
				// No mocks needed
			},
			expectedError: ErrInvalidParameter,
		},
		{
			name:       "assignee not found",
			taskID:     "task123",
			assigneeID: "nonexistent",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "nonexistent").
					Return(false, nil)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:       "task not found",
			taskID:     "nonexistent",
			assigneeID: "user456",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user456").
					Return(true, nil)

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "nonexistent").
					Return(nil, ErrTaskNotFound)
			},
			expectedError: ErrTaskNotFound,
		},
		{
			name:       "duplicate assignment",
			taskID:     "task123",
			assigneeID: "user456",
			setupMocks: func() {
				assigneeID := "user456"
				task := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Test Description",
					Status:      domain.TaskStatusTodo,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					AssigneeID:  &assigneeID,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user456").
					Return(true, nil)

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(task, nil)
			},
			expectedError: ErrDuplicateAssignment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			task, err := service.AssignTask(context.Background(), tt.taskID, tt.assigneeID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.NotNil(t, task.AssigneeID)
				assert.Equal(t, tt.assigneeID, *task.AssigneeID)
			}
		})
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockTaskRepository(ctrl)
	mockUserValidator := mock_usecase.NewMockUserValidator(ctrl)
	mockEventPublisher := mock_usecase.NewMockEventPublisher(ctrl)
	mockLogger := logger.NewMockLogger()

	service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, mockLogger)

	tests := []struct {
		name          string
		taskID        string
		title         *string
		description   *string
		status        *domain.TaskStatus
		priority      *domain.Priority
		dueDate       *time.Time
		setupMocks    func()
		expectedError error
	}{
		{
			name:   "successful update with title change",
			taskID: "task123",
			title:  stringPtr("Updated Title"),
			setupMocks: func() {
				originalTask := &domain.Task{
					ID:          "task123",
					Title:       "Original Title",
					Description: "Description",
					Status:      domain.TaskStatusTodo,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(originalTask, nil)

				mockRepo.EXPECT().
					UpdateTask(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, task *domain.Task) {
						assert.Equal(t, "Updated Title", task.Title)
						assert.Equal(t, "Description", task.Description) // Unchanged
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskUpdated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "update status to done triggers completion event",
			taskID: "task123",
			status: (*domain.TaskStatus)(&[]domain.TaskStatus{domain.TaskStatusDone}[0]),
			setupMocks: func() {
				originalTask := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Description",
					Status:      domain.TaskStatusInProgress,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(originalTask, nil)

				mockRepo.EXPECT().
					UpdateTask(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, task *domain.Task) {
						assert.Equal(t, domain.TaskStatusDone, task.Status)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskUpdated(gomock.Any(), gomock.Any()).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskCompleted(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "no changes should return early",
			taskID: "task123",
			// No fields to update
			setupMocks: func() {
				originalTask := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Description",
					Status:      domain.TaskStatusTodo,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(originalTask, nil)

				// No update or event calls expected
			},
			expectedError: nil,
		},
		{
			name:          "empty task ID",
			taskID:        "",
			title:         stringPtr("Updated Title"),
			setupMocks:    func() {},
			expectedError: ErrInvalidParameter,
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			title:  stringPtr("Updated Title"),
			setupMocks: func() {
				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "nonexistent").
					Return(nil, ErrTaskNotFound)
			},
			expectedError: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			task, err := service.UpdateTask(
				context.Background(),
				tt.taskID,
				tt.title,
				tt.description,
				tt.status,
				tt.priority,
				tt.dueDate,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
			}
		})
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockTaskRepository(ctrl)
	mockUserValidator := mock_usecase.NewMockUserValidator(ctrl)
	mockEventPublisher := mock_usecase.NewMockEventPublisher(ctrl)
	mockLogger := logger.NewMockLogger()

	service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, mockLogger)

	tests := []struct {
		name          string
		taskID        string
		setupMocks    func()
		expectedError error
	}{
		{
			name:   "successful deletion",
			taskID: "task123",
			setupMocks: func() {
				task := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Description",
					Status:      domain.TaskStatusTodo,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(task, nil)

				mockRepo.EXPECT().
					DeleteTask(gomock.Any(), "task123").
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskDeleted(gomock.Any(), "task123").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "empty task ID",
			taskID:        "",
			setupMocks:    func() {},
			expectedError: ErrInvalidParameter,
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			setupMocks: func() {
				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "nonexistent").
					Return(nil, ErrTaskNotFound)
			},
			expectedError: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.DeleteTask(context.Background(), tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskService_ChangeTaskStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockTaskRepository(ctrl)
	mockUserValidator := mock_usecase.NewMockUserValidator(ctrl)
	mockEventPublisher := mock_usecase.NewMockEventPublisher(ctrl)
	mockLogger := logger.NewMockLogger()

	service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, mockLogger)

	tests := []struct {
		name          string
		taskID        string
		newStatus     domain.TaskStatus
		setupMocks    func()
		expectedError error
	}{
		{
			name:      "change to in progress",
			taskID:    "task123",
			newStatus: domain.TaskStatusInProgress,
			setupMocks: func() {
				task := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Description",
					Status:      domain.TaskStatusTodo,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(task, nil)

				mockRepo.EXPECT().
					UpdateTask(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updatedTask *domain.Task) {
						assert.Equal(t, domain.TaskStatusInProgress, updatedTask.Status)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskUpdated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "change to done triggers completion event",
			taskID:    "task123",
			newStatus: domain.TaskStatusDone,
			setupMocks: func() {
				task := &domain.Task{
					ID:          "task123",
					Title:       "Test Task",
					Description: "Description",
					Status:      domain.TaskStatusInProgress,
					Priority:    domain.PriorityMedium,
					Category:    domain.CategoryWork,
					CreatedBy:   "user123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockRepo.EXPECT().
					GetTaskByID(gomock.Any(), "task123").
					Return(task, nil)

				mockRepo.EXPECT().
					UpdateTask(gomock.Any(), gomock.Any()).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskUpdated(gomock.Any(), gomock.Any()).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishTaskCompleted(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			task, err := service.ChangeTaskStatus(context.Background(), tt.taskID, tt.newStatus)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, tt.newStatus, task.Status)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
