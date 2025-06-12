package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go
//go:generate mockgen -source=service.go -destination=mocks/mock_interfaces.go

// テスト用のサイレントロガーを作成
func createTestLogger() *logger.Logger {
	cfg := &logger.Config{
		Level:       "fatal", // 何も出力しない
		Output:      "console",
		Development: false,
	}
	return logger.NewLogger(cfg)
}

// MockTaskRepository はテスト用のTaskRepositoryモック
type MockTaskRepository struct {
	CreateTaskFunc         func(ctx context.Context, task *domain.Task) error
	GetTaskByIDFunc        func(ctx context.Context, id string) (*domain.Task, error)
	ListTasksFunc          func(ctx context.Context, filter domain.ListFilter, pagination domain.Pagination, sortOptions domain.SortOptions) ([]*domain.Task, int, error)
	UpdateTaskFunc         func(ctx context.Context, task *domain.Task) error
	DeleteTaskFunc         func(ctx context.Context, id string) error
	GetOverdueTasksFunc    func(ctx context.Context) ([]*domain.Task, error)
	GetTasksByAssigneeFunc func(ctx context.Context, userID string) ([]*domain.Task, error)
	SearchTasksFunc        func(ctx context.Context, query string, limit int) ([]*domain.Task, error)
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, task *domain.Task) error {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, task)
	}
	return nil
}

func (m *MockTaskRepository) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	if m.GetTaskByIDFunc != nil {
		return m.GetTaskByIDFunc(ctx, id)
	}
	return nil, ErrTaskNotFound
}

func (m *MockTaskRepository) ListTasks(ctx context.Context, filter domain.ListFilter, pagination domain.Pagination, sortOptions domain.SortOptions) ([]*domain.Task, int, error) {
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx, filter, pagination, sortOptions)
	}
	return []*domain.Task{}, 0, nil
}

func (m *MockTaskRepository) UpdateTask(ctx context.Context, task *domain.Task) error {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, task)
	}
	return nil
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, id string) error {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(ctx, id)
	}
	return nil
}

func (m *MockTaskRepository) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	if m.GetOverdueTasksFunc != nil {
		return m.GetOverdueTasksFunc(ctx)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskRepository) GetTasksByAssignee(ctx context.Context, userID string) ([]*domain.Task, error) {
	if m.GetTasksByAssigneeFunc != nil {
		return m.GetTasksByAssigneeFunc(ctx, userID)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskRepository) SearchTasks(ctx context.Context, query string, limit int) ([]*domain.Task, error) {
	if m.SearchTasksFunc != nil {
		return m.SearchTasksFunc(ctx, query, limit)
	}
	return []*domain.Task{}, nil
}

// MockUserValidator はテスト用のUserValidatorモック
type MockUserValidator struct {
	UserExistsFunc        func(ctx context.Context, userID string) (bool, error)
	GetUserInfoFunc       func(ctx context.Context, userID string) (*commonDomain.UserInfo, error)
	GetUsersInfoBatchFunc func(ctx context.Context, userIDs []string) (map[string]*commonDomain.UserInfo, error)
}

func (m *MockUserValidator) UserExists(ctx context.Context, userID string) (bool, error) {
	if m.UserExistsFunc != nil {
		return m.UserExistsFunc(ctx, userID)
	}
	return true, nil
}

func (m *MockUserValidator) GetUserInfo(ctx context.Context, userID string) (*commonDomain.UserInfo, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(ctx, userID)
	}
	return &commonDomain.UserInfo{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}

func (m *MockUserValidator) GetUsersInfoBatch(ctx context.Context, userIDs []string) (map[string]*commonDomain.UserInfo, error) {
	if m.GetUsersInfoBatchFunc != nil {
		return m.GetUsersInfoBatchFunc(ctx, userIDs)
	}
	result := make(map[string]*commonDomain.UserInfo)
	for _, userID := range userIDs {
		result[userID] = &commonDomain.UserInfo{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
		}
	}
	return result, nil
}

// MockEventPublisher はテスト用のEventPublisherモック
type MockEventPublisher struct {
	PublishTaskCreatedFunc   func(ctx context.Context, task *domain.Task) error
	PublishTaskUpdatedFunc   func(ctx context.Context, task *domain.Task) error
	PublishTaskDeletedFunc   func(ctx context.Context, taskID string) error
	PublishTaskAssignedFunc  func(ctx context.Context, task *domain.Task) error
	PublishTaskCompletedFunc func(ctx context.Context, task *domain.Task) error
}

func (m *MockEventPublisher) PublishTaskCreated(ctx context.Context, task *domain.Task) error {
	if m.PublishTaskCreatedFunc != nil {
		return m.PublishTaskCreatedFunc(ctx, task)
	}
	return nil
}

func (m *MockEventPublisher) PublishTaskUpdated(ctx context.Context, task *domain.Task) error {
	if m.PublishTaskUpdatedFunc != nil {
		return m.PublishTaskUpdatedFunc(ctx, task)
	}
	return nil
}

func (m *MockEventPublisher) PublishTaskDeleted(ctx context.Context, taskID string) error {
	if m.PublishTaskDeletedFunc != nil {
		return m.PublishTaskDeletedFunc(ctx, taskID)
	}
	return nil
}

func (m *MockEventPublisher) PublishTaskAssigned(ctx context.Context, task *domain.Task) error {
	if m.PublishTaskAssignedFunc != nil {
		return m.PublishTaskAssignedFunc(ctx, task)
	}
	return nil
}

func (m *MockEventPublisher) PublishTaskCompleted(ctx context.Context, task *domain.Task) error {
	if m.PublishTaskCompletedFunc != nil {
		return m.PublishTaskCompletedFunc(ctx, task)
	}
	return nil
}

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		description   string
		priority      domain.Priority
		category      domain.Category
		createdBy     string
		setupMocks    func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher)
		expectedError error
	}{
		{
			name:        "successful task creation",
			title:       "Test Task",
			description: "Test Description",
			priority:    domain.PriorityHigh,
			category:    domain.CategoryWork,
			createdBy:   "user123",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{
					CreateTaskFunc: func(ctx context.Context, task *domain.Task) error {
						// Verify task properties
						assert.Equal(t, "Test Task", task.Title)
						assert.Equal(t, "Test Description", task.Description)
						assert.Equal(t, domain.PriorityHigh, task.Priority)
						assert.Equal(t, domain.CategoryWork, task.Category)
						assert.Equal(t, "user123", task.CreatedBy)
						assert.Equal(t, domain.TaskStatusTodo, task.Status)
						assert.NotEmpty(t, task.ID)
						return nil
					},
				}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return true, nil
					},
				}

				mockEventPublisher := &MockEventPublisher{
					PublishTaskCreatedFunc: func(ctx context.Context, task *domain.Task) error {
						return nil
					},
				}

				return mockRepo, mockUserValidator, mockEventPublisher
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
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				return &MockTaskRepository{}, &MockUserValidator{}, &MockEventPublisher{}
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
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				return &MockTaskRepository{}, &MockUserValidator{}, &MockEventPublisher{}
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
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return false, nil
					},
				}

				mockEventPublisher := &MockEventPublisher{}

				return mockRepo, mockUserValidator, mockEventPublisher
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
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return false, errors.New("validation error")
					},
				}

				mockEventPublisher := &MockEventPublisher{}

				return mockRepo, mockUserValidator, mockEventPublisher
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
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{
					CreateTaskFunc: func(ctx context.Context, task *domain.Task) error {
						return errors.New("database error")
					},
				}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return true, nil
					},
				}

				mockEventPublisher := &MockEventPublisher{}

				return mockRepo, mockUserValidator, mockEventPublisher
			},
			expectedError: errors.New("failed to create task"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, mockUserValidator, mockEventPublisher := tt.setupMocks()
			mockLogger := createTestLogger()

			service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, *mockLogger)

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
	tests := []struct {
		name          string
		taskID        string
		setupMocks    func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher)
		expectedTask  *domain.Task
		expectedError error
	}{
		{
			name:   "successful get task",
			taskID: "task123",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
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

				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return task, nil
					},
				}

				return mockRepo, &MockUserValidator{}, &MockEventPublisher{}
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
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				return &MockTaskRepository{}, &MockUserValidator{}, &MockEventPublisher{}
			},
			expectedTask:  nil,
			expectedError: ErrInvalidParameter,
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return nil, ErrTaskNotFound
					},
				}

				return mockRepo, &MockUserValidator{}, &MockEventPublisher{}
			},
			expectedTask:  nil,
			expectedError: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, mockUserValidator, mockEventPublisher := tt.setupMocks()
			mockLogger := createTestLogger()

			service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, *mockLogger)

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
	tests := []struct {
		name          string
		taskID        string
		assigneeID    string
		setupMocks    func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher)
		expectedError error
	}{
		{
			name:       "successful assignment",
			taskID:     "task123",
			assigneeID: "user456",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
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

				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return task, nil
					},
					UpdateTaskFunc: func(ctx context.Context, updatedTask *domain.Task) error {
						assert.NotNil(t, updatedTask.AssigneeID)
						assert.Equal(t, "user456", *updatedTask.AssigneeID)
						return nil
					},
				}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return true, nil
					},
				}

				mockEventPublisher := &MockEventPublisher{
					PublishTaskAssignedFunc: func(ctx context.Context, task *domain.Task) error {
						return nil
					},
				}

				return mockRepo, mockUserValidator, mockEventPublisher
			},
			expectedError: nil,
		},
		{
			name:       "empty task ID",
			taskID:     "",
			assigneeID: "user456",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				return &MockTaskRepository{}, &MockUserValidator{}, &MockEventPublisher{}
			},
			expectedError: ErrInvalidParameter,
		},
		{
			name:       "empty assignee ID",
			taskID:     "task123",
			assigneeID: "",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				return &MockTaskRepository{}, &MockUserValidator{}, &MockEventPublisher{}
			},
			expectedError: ErrInvalidParameter,
		},
		{
			name:       "assignee not found",
			taskID:     "task123",
			assigneeID: "nonexistent",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return false, nil
					},
				}

				mockEventPublisher := &MockEventPublisher{}

				return mockRepo, mockUserValidator, mockEventPublisher
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:       "task not found",
			taskID:     "nonexistent",
			assigneeID: "user456",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return nil, ErrTaskNotFound
					},
				}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return true, nil
					},
				}

				mockEventPublisher := &MockEventPublisher{}

				return mockRepo, mockUserValidator, mockEventPublisher
			},
			expectedError: ErrTaskNotFound,
		},
		{
			name:       "duplicate assignment",
			taskID:     "task123",
			assigneeID: "user456",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
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

				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return task, nil
					},
				}

				mockUserValidator := &MockUserValidator{
					UserExistsFunc: func(ctx context.Context, userID string) (bool, error) {
						return true, nil
					},
				}

				mockEventPublisher := &MockEventPublisher{}

				return mockRepo, mockUserValidator, mockEventPublisher
			},
			expectedError: ErrDuplicateAssignment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, mockUserValidator, mockEventPublisher := tt.setupMocks()
			mockLogger := createTestLogger()

			service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, *mockLogger)

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

func TestTaskService_DeleteTask(t *testing.T) {
	tests := []struct {
		name          string
		taskID        string
		setupMocks    func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher)
		expectedError error
	}{
		{
			name:   "successful deletion",
			taskID: "task123",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
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

				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return task, nil
					},
					DeleteTaskFunc: func(ctx context.Context, id string) error {
						return nil
					},
				}

				mockEventPublisher := &MockEventPublisher{
					PublishTaskDeletedFunc: func(ctx context.Context, taskID string) error {
						return nil
					},
				}

				return mockRepo, &MockUserValidator{}, mockEventPublisher
			},
			expectedError: nil,
		},
		{
			name:   "empty task ID",
			taskID: "",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				return &MockTaskRepository{}, &MockUserValidator{}, &MockEventPublisher{}
			},
			expectedError: ErrInvalidParameter,
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return nil, ErrTaskNotFound
					},
				}

				return mockRepo, &MockUserValidator{}, &MockEventPublisher{}
			},
			expectedError: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, mockUserValidator, mockEventPublisher := tt.setupMocks()
			mockLogger := createTestLogger()

			service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, *mockLogger)

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
	tests := []struct {
		name          string
		taskID        string
		newStatus     domain.TaskStatus
		setupMocks    func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher)
		expectedError error
	}{
		{
			name:      "change to in progress",
			taskID:    "task123",
			newStatus: domain.TaskStatusInProgress,
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
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

				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return task, nil
					},
					UpdateTaskFunc: func(ctx context.Context, updatedTask *domain.Task) error {
						assert.Equal(t, domain.TaskStatusInProgress, updatedTask.Status)
						return nil
					},
				}

				mockEventPublisher := &MockEventPublisher{
					PublishTaskUpdatedFunc: func(ctx context.Context, task *domain.Task) error {
						return nil
					},
				}

				return mockRepo, &MockUserValidator{}, mockEventPublisher
			},
			expectedError: nil,
		},
		{
			name:      "change to done triggers completion event",
			taskID:    "task123",
			newStatus: domain.TaskStatusDone,
			setupMocks: func() (*MockTaskRepository, *MockUserValidator, *MockEventPublisher) {
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

				mockRepo := &MockTaskRepository{
					GetTaskByIDFunc: func(ctx context.Context, id string) (*domain.Task, error) {
						return task, nil
					},
					UpdateTaskFunc: func(ctx context.Context, updatedTask *domain.Task) error {
						return nil
					},
				}

				mockEventPublisher := &MockEventPublisher{
					PublishTaskUpdatedFunc: func(ctx context.Context, task *domain.Task) error {
						return nil
					},
					PublishTaskCompletedFunc: func(ctx context.Context, task *domain.Task) error {
						return nil
					},
				}

				return mockRepo, &MockUserValidator{}, mockEventPublisher
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, mockUserValidator, mockEventPublisher := tt.setupMocks()
			mockLogger := createTestLogger()

			service := NewTaskService(mockRepo, mockUserValidator, mockEventPublisher, *mockLogger)

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
