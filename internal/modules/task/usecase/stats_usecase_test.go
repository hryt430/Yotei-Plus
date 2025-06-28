package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase/mocks"
	"github.com/hryt430/Yotei+/pkg/logger"
)

//go:generate mockgen -source=stats_service.go -destination=mocks/mock_stats_repository.go -package=mocks
//go:generate mockgen -source=repository.go -destination=mocks/mock_task_repository.go -package=mocks


func TestTaskStatsService_GetDashboardStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name          string
		userID        string
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "successful dashboard stats retrieval",
			userID: "user123",
			setupMocks: func() {
				// Today's stats
				todayTasks := []*domain.Task{
					{ID: "1", Status: domain.TaskStatusDone},
					{ID: "2", Status: domain.TaskStatusTodo},
				}
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(todayTasks, nil)

				// Today's GetTasksByDateRange call (from GetDailyStats)
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return([]*domain.Task{}, nil)

				// Weekly overview (7 days of GetTasksByDueDate calls)
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return(todayTasks, nil)
				}
				// Weekly overview (7 days of GetTasksByDateRange calls from GetDailyStats)
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
						Return([]*domain.Task{}, nil)
				}

				// Upcoming week tasks (7 more calls)
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{}, nil)
				}

				// Category breakdown
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(todayTasks, nil)

				// Priority breakdown
				mockTaskRepo.EXPECT().
					ListTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(todayTasks, 2, nil)

				// Recent completions
				mockStatsRepo.EXPECT().
					GetRecentCompletedTasks(gomock.Any(), "user123", 5).
					Return([]*domain.Task{{ID: "completed1", Status: domain.TaskStatusDone}}, nil)

				// Overdue count
				mockStatsRepo.EXPECT().
					GetOverdueTasksCount(gomock.Any(), "user123").
					Return(3, nil)
			},
			expectedError: "",
		},
		{
			name:   "today stats error should return error",
			userID: "user123",
			setupMocks: func() {
				// Today's stats error - GetTasksByDueDate fails, so GetDailyStats fails early
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(nil, errors.New("database error"))
				// GetDailyStats fails before calling GetTasksByDateRange, so no call expected
			},
			expectedError: "failed to get today stats",
		},
		{
			name:   "graceful degradation on partial errors",
			userID: "user123",
			setupMocks: func() {
				// Successful today stats - GetTasksByDueDate call
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return([]*domain.Task{}, nil)
				// Successful today stats - GetTasksByDateRange call
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return([]*domain.Task{}, nil)

				// Weekly overview success (7 GetTasksByDueDate calls)
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{}, nil)
				}
				// Weekly overview success (7 GetTasksByDateRange calls)
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
						Return([]*domain.Task{}, nil)
				}

				// Upcoming week success (7 GetTasksByDueDate calls only)
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{}, nil)
				}

				// Category breakdown error - should not fail overall
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(nil, errors.New("category error"))

				// Priority breakdown error - should not fail overall
				mockTaskRepo.EXPECT().
					ListTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, errors.New("priority error"))

				// Recent completions error - should not fail overall
				mockStatsRepo.EXPECT().
					GetRecentCompletedTasks(gomock.Any(), "user123", 5).
					Return(nil, errors.New("recent error"))

				// Overdue count error - should not fail overall
				mockStatsRepo.EXPECT().
					GetOverdueTasksCount(gomock.Any(), "user123").
					Return(0, errors.New("overdue error"))
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			stats, err := service.GetDashboardStats(context.Background(), tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				assert.NotNil(t, stats.TodayStats)
				assert.NotNil(t, stats.WeeklyOverview)
				assert.NotNil(t, stats.UpcomingWeekTasks)
				assert.NotNil(t, stats.CategoryBreakdown)
				assert.NotNil(t, stats.PriorityBreakdown)
				assert.NotNil(t, stats.RecentCompletions)
			}
		})
	}
}

func TestTaskStatsService_GetDailyStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name          string
		userID        string
		date          time.Time
		setupMocks    func()
		expectedStats *domain.DailyStats
		expectedError string
	}{
		{
			name:   "successful daily stats",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			setupMocks: func() {
				dueTasks := []*domain.Task{
					{ID: "1", Status: domain.TaskStatusDone},
					{ID: "2", Status: domain.TaskStatusTodo},
				}
				createdTasks := []*domain.Task{
					{ID: "3", Status: domain.TaskStatusInProgress},
				}

				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(dueTasks, nil)

				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(createdTasks, nil)
			},
			expectedStats: &domain.DailyStats{
				TotalTasks:      3, // Merged unique tasks
				CompletedTasks:  1,
				TodoTasks:       1,
				InProgressTasks: 1,
			},
			expectedError: "",
		},
		{
			name:   "due date query error",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			setupMocks: func() {
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedStats: nil,
			expectedError: "failed to get tasks by due date",
		},
		{
			name:   "date range query error",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			setupMocks: func() {
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return([]*domain.Task{}, nil)

				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(nil, errors.New("range query error"))
			},
			expectedStats: nil,
			expectedError: "failed to get tasks by date range",
		},
		{
			name:   "duplicate task handling",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			setupMocks: func() {
				// Same task appears in both due and created lists
				duplicateTask := &domain.Task{ID: "duplicate", Status: domain.TaskStatusDone}
				dueTasks := []*domain.Task{duplicateTask}
				createdTasks := []*domain.Task{duplicateTask, {ID: "unique", Status: domain.TaskStatusTodo}}

				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(dueTasks, nil)

				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(createdTasks, nil)
			},
			expectedStats: &domain.DailyStats{
				TotalTasks:     2, // Should deduplicate
				CompletedTasks: 1,
				TodoTasks:      1,
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			stats, err := service.GetDailyStats(context.Background(), tt.userID, tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				if tt.expectedStats != nil {
					assert.Equal(t, tt.expectedStats.TotalTasks, stats.TotalTasks)
					assert.Equal(t, tt.expectedStats.CompletedTasks, stats.CompletedTasks)
					assert.Equal(t, tt.expectedStats.TodoTasks, stats.TodoTasks)
					assert.Equal(t, tt.expectedStats.InProgressTasks, stats.InProgressTasks)
				}
			}
		})
	}
}

func TestTaskStatsService_GetWeeklyStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name          string
		userID        string
		date          time.Time
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "successful weekly stats",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // Monday
			setupMocks: func() {
				// Mock 7 days of daily stats (Monday to Sunday)
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{{ID: "task1", Status: domain.TaskStatusDone}}, nil)

					mockStatsRepo.EXPECT().
						GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
						Return([]*domain.Task{}, nil)
				}
			},
			expectedError: "",
		},
		{
			name:   "partial failure in weekly stats",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			setupMocks: func() {
				// First day fails - GetTasksByDueDate fails, so GetDailyStats fails before GetTasksByDateRange
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(nil, errors.New("first day error"))
				// No GetTasksByDateRange call for first day since GetTasksByDueDate failed

				// Remaining 6 days succeed
				for i := 0; i < 6; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{}, nil)

					mockStatsRepo.EXPECT().
						GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
						Return([]*domain.Task{}, nil)
				}
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			stats, err := service.GetWeeklyStats(context.Background(), tt.userID, tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				assert.NotNil(t, stats.DailyStats)
				assert.Len(t, stats.DailyStats, 7) // Should have 7 days
				assert.False(t, stats.WeekStart.IsZero())
				assert.False(t, stats.WeekEnd.IsZero())
			}
		})
	}
}

func TestTaskStatsService_GetCategoryBreakdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name              string
		userID            string
		setupMocks        func()
		expectedBreakdown map[domain.Category]int
		expectedError     string
	}{
		{
			name:   "successful category breakdown",
			userID: "user123",
			setupMocks: func() {
				tasks := []*domain.Task{
					{ID: "1", Category: domain.CategoryWork},
					{ID: "2", Category: domain.CategoryWork},
					{ID: "3", Category: domain.CategoryPersonal},
					{ID: "4", Category: domain.CategoryStudy},
				}

				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(tasks, nil)
			},
			expectedBreakdown: map[domain.Category]int{
				domain.CategoryWork:     2,
				domain.CategoryPersonal: 1,
				domain.CategoryStudy:    1,
			},
			expectedError: "",
		},
		{
			name:   "repository error",
			userID: "user123",
			setupMocks: func() {
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedBreakdown: nil,
			expectedError:     "failed to get tasks for category breakdown",
		},
		{
			name:   "empty tasks",
			userID: "user123",
			setupMocks: func() {
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return([]*domain.Task{}, nil)
			},
			expectedBreakdown: map[domain.Category]int{},
			expectedError:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			breakdown, err := service.GetCategoryBreakdown(context.Background(), tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, breakdown)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBreakdown, breakdown)
			}
		})
	}
}

func TestTaskStatsService_GetPriorityBreakdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name              string
		userID            string
		setupMocks        func()
		expectedBreakdown map[domain.Priority]int
		expectedError     string
	}{
		{
			name:   "successful priority breakdown",
			userID: "user123",
			setupMocks: func() {
				tasks := []*domain.Task{
					{ID: "1", Priority: domain.PriorityHigh, Status: domain.TaskStatusTodo},
					{ID: "2", Priority: domain.PriorityHigh, Status: domain.TaskStatusInProgress},
					{ID: "3", Priority: domain.PriorityMedium, Status: domain.TaskStatusTodo},
					{ID: "4", Priority: domain.PriorityLow, Status: domain.TaskStatusDone}, // Should be excluded
				}

				mockTaskRepo.EXPECT().
					ListTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, filter domain.ListFilter, pagination domain.Pagination, sortOptions domain.SortOptions) {
						// Verify filter contains assignee ID
						assert.Equal(t, "user123", *filter.AssigneeID)
						// Verify pagination
						assert.Equal(t, 1, pagination.Page)
						assert.Equal(t, 1000, pagination.PageSize)
					}).
					Return(tasks, 4, nil)
			},
			expectedBreakdown: map[domain.Priority]int{
				domain.PriorityHigh:   2, // Only non-DONE tasks
				domain.PriorityMedium: 1,
				// PriorityLow task is DONE, so excluded
			},
			expectedError: "",
		},
		{
			name:   "repository error",
			userID: "user123",
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					ListTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, errors.New("database error"))
			},
			expectedBreakdown: nil,
			expectedError:     "failed to get tasks for priority breakdown",
		},
		{
			name:   "all tasks completed",
			userID: "user123",
			setupMocks: func() {
				tasks := []*domain.Task{
					{ID: "1", Priority: domain.PriorityHigh, Status: domain.TaskStatusDone},
					{ID: "2", Priority: domain.PriorityMedium, Status: domain.TaskStatusDone},
				}

				mockTaskRepo.EXPECT().
					ListTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(tasks, 2, nil)
			},
			expectedBreakdown: map[domain.Priority]int{},
			expectedError:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			breakdown, err := service.GetPriorityBreakdown(context.Background(), tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, breakdown)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBreakdown, breakdown)
			}
		})
	}
}

func TestTaskStatsService_GetProgressSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name          string
		userID        string
		days          int
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "successful progress summary",
			userID: "user123",
			days:   7,
			setupMocks: func() {
				// Mock 7 days of daily stats
				for i := 0; i < 7; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{{ID: "task1", Status: domain.TaskStatusDone}}, nil)

					mockStatsRepo.EXPECT().
						GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
						Return([]*domain.Task{}, nil)
				}
			},
			expectedError: "",
		},
		{
			name:   "partial failures in progress summary",
			userID: "user123",
			days:   3,
			setupMocks: func() {
				// First day fails - GetTasksByDueDate fails, so GetDailyStats fails before GetTasksByDateRange
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(nil, errors.New("day error"))
				// No GetTasksByDateRange call for first day since GetTasksByDueDate failed

				// Other 2 days succeed
				for i := 0; i < 2; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{}, nil)

					mockStatsRepo.EXPECT().
						GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
						Return([]*domain.Task{}, nil)
				}
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			summary, err := service.GetProgressSummary(context.Background(), tt.userID, tt.days)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, summary)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, summary)
				assert.Len(t, summary, tt.days)

				// Verify chronological order (oldest to newest)
				for i := 1; i < len(summary); i++ {
					assert.True(t, summary[i].Date.After(summary[i-1].Date) || summary[i].Date.Equal(summary[i-1].Date))
				}
			}
		})
	}
}

func TestTaskStatsService_GetMonthlyStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name          string
		userID        string
		year          int
		month         time.Month
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "successful monthly stats",
			userID: "user123",
			year:   2024,
			month:  time.January,
			setupMocks: func() {
				tasks := []*domain.Task{
					{
						ID:        "1",
						Status:    domain.TaskStatusDone,
						DueDate:   &[]time.Time{time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}[0],
						CreatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        "2",
						Status:    domain.TaskStatusTodo,
						CreatedAt: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
					},
				}

				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, userID string, start, end time.Time) {
						// Verify date range is for January 2024
						assert.Equal(t, 2024, start.Year())
						assert.Equal(t, time.January, start.Month())
						assert.Equal(t, 1, start.Day())

						assert.Equal(t, 2024, end.Year())
						assert.Equal(t, time.January, end.Month())
						assert.Equal(t, 31, end.Day())
					}).
					Return(tasks, nil)
			},
			expectedError: "",
		},
		{
			name:   "repository error",
			userID: "user123",
			year:   2024,
			month:  time.February,
			setupMocks: func() {
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get monthly tasks",
		},
		{
			name:   "leap year february",
			userID: "user123",
			year:   2024, // Leap year
			month:  time.February,
			setupMocks: func() {
				mockStatsRepo.EXPECT().
					GetTasksByDateRange(gomock.Any(), "user123", gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, userID string, start, end time.Time) {
						// Verify February has 29 days in 2024
						assert.Equal(t, 29, end.Day())
					}).
					Return([]*domain.Task{}, nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			stats, err := service.GetMonthlyStats(context.Background(), tt.userID, tt.year, tt.month)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				assert.NotNil(t, stats.DailyStats)
				assert.False(t, stats.WeekStart.IsZero())
				assert.False(t, stats.WeekEnd.IsZero())
			}
		})
	}
}

// Integration test for GetWeeklyPreview
func TestTaskStatsService_GetWeeklyPreview(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockStatsRepo := mocks.NewMockStatsRepository(ctrl)
	// Create a test logger
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	logger.Init(cfg)
	testLogger := logger.Get()
	defer testLogger.Close()

	service := NewTaskStatsService(mockTaskRepo, mockStatsRepo, testLogger)

	tests := []struct {
		name          string
		userID        string
		date          time.Time
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "successful weekly preview",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			setupMocks: func() {
				// Mock 7 days of tasks
				for i := 0; i < 7; i++ {
					tasks := []*domain.Task{}
					if i == 0 { // Monday has tasks
						tasks = []*domain.Task{
							{ID: "1", Status: domain.TaskStatusTodo},
							{ID: "2", Status: domain.TaskStatusDone},
						}
					}

					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return(tasks, nil)
				}
			},
			expectedError: "",
		},
		{
			name:   "preview with errors should continue",
			userID: "user123",
			date:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			setupMocks: func() {
				// First day errors, others succeed
				mockStatsRepo.EXPECT().
					GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
					Return(nil, errors.New("first day error"))

				// Remaining 6 days
				for i := 0; i < 6; i++ {
					mockStatsRepo.EXPECT().
						GetTasksByDueDate(gomock.Any(), "user123", gomock.Any()).
						Return([]*domain.Task{}, nil)
				}
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			preview, err := service.GetWeeklyPreview(context.Background(), tt.userID, tt.date)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, preview)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, preview)
				assert.NotNil(t, preview.DailyPreview)
				assert.Len(t, preview.DailyPreview, 7)
				assert.False(t, preview.WeekStart.IsZero())
				assert.False(t, preview.WeekEnd.IsZero())
			}
		})
	}
}
