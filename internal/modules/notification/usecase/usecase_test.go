package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/mocks"
	"github.com/hryt430/Yotei+/pkg/logger"
)

//go:generate mockgen -source=persistence/notification_repository.go -destination=mocks/mock_repository.go -package=mocks
//go:generate mockgen -source=output/notification_output.go -destination=mocks/mock_gateway.go -package=mocks
//go:generate mockgen -package=mocks -destination=mocks/mock_user_validator.go github.com/hryt430/Yotei+/internal/common/domain UserValidator


// ===================
// UseCase Tests
// ===================

func TestNotificationUseCase_CreateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockAppGateway := mocks.NewMockAppNotificationGateway(ctrl)
	mockLineGateway := mocks.NewMockLineNotificationGateway(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	useCase := NewNotificationUseCase(
		mockRepo,
		mockAppGateway,
		mockLineGateway,
		mockUserValidator,
		mockLogger,
	)

	tests := []struct {
		name          string
		input         input.CreateNotificationInput
		setupMocks    func()
		expectedError string
	}{
		{
			name: "successful app notification creation",
			input: input.CreateNotificationInput{
				UserID:   "user123",
				Type:     "TASK_ASSIGNED",
				Title:    "Task Assigned",
				Message:  "A new task has been assigned to you",
				Metadata: map[string]string{"task_id": "task123"},
				Channels: []string{"app"},
			},
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user123").
					Return(true, nil)

				mockRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, notification *domain.Notification) {
						assert.Equal(t, "user123", notification.UserID)
						assert.Equal(t, domain.TaskAssigned, notification.Type)
						assert.Equal(t, "Task Assigned", notification.Title)
						assert.Equal(t, "A new task has been assigned to you", notification.Message)
						assert.Equal(t, "task123", notification.Metadata["task_id"])
						assert.Len(t, notification.Channels, 1)
						assert.Equal(t, domain.AppInternal, notification.Channels[0].GetType())
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "multiple channels notification",
			input: input.CreateNotificationInput{
				UserID:  "user123",
				Type:    "TASK_DUE_SOON",
				Title:   "Task Due Soon",
				Message: "Your task is due in 2 hours",
				Metadata: map[string]string{
					"task_id":      "task123",
					"line_user_id": "line_user_456",
				},
				Channels: []string{"app", "line"},
			},
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user123").
					Return(true, nil)

				mockRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, notification *domain.Notification) {
						assert.Len(t, notification.Channels, 2)
						// Check that both app and line channels are added
						channelTypes := make(map[domain.ChannelType]bool)
						for _, channel := range notification.Channels {
							channelTypes[channel.GetType()] = true
						}
						assert.True(t, channelTypes[domain.AppInternal])
						assert.True(t, channelTypes[domain.LineMessage])
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "validation error - empty user ID",
			input: input.CreateNotificationInput{
				UserID:   "",
				Type:     "TASK_ASSIGNED",
				Title:    "Task Assigned",
				Message:  "Message",
				Channels: []string{"app"},
			},
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "user ID is required",
		},
		{
			name: "validation error - empty title",
			input: input.CreateNotificationInput{
				UserID:   "user123",
				Type:     "TASK_ASSIGNED",
				Title:    "",
				Message:  "Message",
				Channels: []string{"app"},
			},
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "title is required",
		},
		{
			name: "validation error - empty channels",
			input: input.CreateNotificationInput{
				UserID:   "user123",
				Type:     "TASK_ASSIGNED",
				Title:    "Title",
				Message:  "Message",
				Channels: []string{},
			},
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "at least one channel is required",
		},
		{
			name: "user not found",
			input: input.CreateNotificationInput{
				UserID:   "nonexistent",
				Type:     "TASK_ASSIGNED",
				Title:    "Task Assigned",
				Message:  "Message",
				Channels: []string{"app"},
			},
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "nonexistent").
					Return(false, nil)
			},
			expectedError: "user not found",
		},
		{
			name: "user validation error",
			input: input.CreateNotificationInput{
				UserID:   "user123",
				Type:     "TASK_ASSIGNED",
				Title:    "Task Assigned",
				Message:  "Message",
				Channels: []string{"app"},
			},
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user123").
					Return(false, errors.New("database error"))
			},
			expectedError: "failed to validate user",
		},
		{
			name: "repository save error",
			input: input.CreateNotificationInput{
				UserID:   "user123",
				Type:     "TASK_ASSIGNED",
				Title:    "Task Assigned",
				Message:  "Message",
				Channels: []string{"app"},
			},
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), "user123").
					Return(true, nil)

				mockRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "failed to save notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			notification, err := useCase.CreateNotification(context.Background(), tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, notification)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, notification)
				assert.Equal(t, tt.input.UserID, notification.UserID)
				assert.Equal(t, tt.input.Title, notification.Title)
			}
		})
	}
}

func TestNotificationUseCase_SendNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockAppGateway := mocks.NewMockAppNotificationGateway(ctrl)
	mockLineGateway := mocks.NewMockLineNotificationGateway(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	useCase := NewNotificationUseCase(
		mockRepo,
		mockAppGateway,
		mockLineGateway,
		mockUserValidator,
		mockLogger,
	)

	tests := []struct {
		name           string
		notificationID string
		setupMocks     func()
		expectedError  string
	}{
		{
			name:           "successful app notification send",
			notificationID: "notification123",
			setupMocks: func() {
				notification := &domain.Notification{
					ID:       "notification123",
					UserID:   "user123",
					Title:    "Test Notification",
					Message:  "Test Message",
					Status:   domain.StatusPending,
					Type:     domain.AppNotification,
					Metadata: map[string]string{"key": "value"},
				}
				notification.AddChannel(domain.NewAppChannel("user123"))

				mockRepo.EXPECT().
					FindByID(gomock.Any(), "notification123").
					Return(notification, nil)

				mockAppGateway.EXPECT().
					SendNotification(
						gomock.Any(),
						"user123",
						"Test Notification",
						"Test Message",
						notification.Metadata,
					).
					Return(nil)

				mockRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updatedNotification *domain.Notification) {
						assert.Equal(t, domain.StatusSent, updatedNotification.Status)
						assert.NotNil(t, updatedNotification.SentAt)
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:           "notification already sent",
			notificationID: "sent_notification",
			setupMocks: func() {
				sentTime := time.Now()
				notification := &domain.Notification{
					ID:      "sent_notification",
					UserID:  "user123",
					Title:   "Test Notification",
					Message: "Test Message",
					Status:  domain.StatusSent,
					SentAt:  &sentTime,
				}

				mockRepo.EXPECT().
					FindByID(gomock.Any(), "sent_notification").
					Return(notification, nil)

				// No gateway calls expected
			},
			expectedError: "",
		},
		{
			name:           "notification not found",
			notificationID: "nonexistent",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindByID(gomock.Any(), "nonexistent").
					Return(nil, nil)
			},
			expectedError: "notification not found",
		},
		{
			name:           "repository find error",
			notificationID: "notification123",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindByID(gomock.Any(), "notification123").
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to find notification",
		},
		{
			name:           "gateway send error",
			notificationID: "notification123",
			setupMocks: func() {
				notification := &domain.Notification{
					ID:      "notification123",
					UserID:  "user123",
					Title:   "Test Notification",
					Message: "Test Message",
					Status:  domain.StatusPending,
				}
				notification.AddChannel(domain.NewAppChannel("user123"))

				mockRepo.EXPECT().
					FindByID(gomock.Any(), "notification123").
					Return(notification, nil)

				mockAppGateway.EXPECT().
					SendNotification(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("gateway error"))

				mockRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updatedNotification *domain.Notification) {
						assert.Equal(t, domain.StatusFailed, updatedNotification.Status)
					}).
					Return(nil)
			},
			expectedError: "failed to send to 1 channels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := useCase.SendNotification(context.Background(), tt.notificationID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotificationUseCase_MarkNotificationAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockAppGateway := mocks.NewMockAppNotificationGateway(ctrl)
	mockLineGateway := mocks.NewMockLineNotificationGateway(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	useCase := NewNotificationUseCase(
		mockRepo,
		mockAppGateway,
		mockLineGateway,
		mockUserValidator,
		mockLogger,
	)

	tests := []struct {
		name           string
		notificationID string
		setupMocks     func()
		expectedError  string
	}{
		{
			name:           "successful mark as read",
			notificationID: "notification123",
			setupMocks: func() {
				mockRepo.EXPECT().
					UpdateStatus(gomock.Any(), "notification123", domain.StatusRead).
					Return(nil)

				mockAppGateway.EXPECT().
					MarkAsRead(gomock.Any(), "notification123").
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:           "repository update error",
			notificationID: "notification123",
			setupMocks: func() {
				mockRepo.EXPECT().
					UpdateStatus(gomock.Any(), "notification123", domain.StatusRead).
					Return(errors.New("database error"))
			},
			expectedError: "failed to mark notification as read",
		},
		{
			name:           "gateway mark as read error (non-fatal)",
			notificationID: "notification123",
			setupMocks: func() {
				mockRepo.EXPECT().
					UpdateStatus(gomock.Any(), "notification123", domain.StatusRead).
					Return(nil)

				mockAppGateway.EXPECT().
					MarkAsRead(gomock.Any(), "notification123").
					Return(errors.New("gateway error"))

				// Should not fail the operation
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := useCase.MarkNotificationAsRead(context.Background(), tt.notificationID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotificationUseCase_GetNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockAppGateway := mocks.NewMockAppNotificationGateway(ctrl)
	mockLineGateway := mocks.NewMockLineNotificationGateway(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	useCase := NewNotificationUseCase(
		mockRepo,
		mockAppGateway,
		mockLineGateway,
		mockUserValidator,
		mockLogger,
	)

	tests := []struct {
		name           string
		notificationID string
		setupMocks     func()
		expectedError  string
	}{
		{
			name:           "successful get notification",
			notificationID: "notification123",
			setupMocks: func() {
				notification := &domain.Notification{
					ID:      "notification123",
					UserID:  "user123",
					Title:   "Test Notification",
					Message: "Test Message",
				}

				mockRepo.EXPECT().
					FindByID(gomock.Any(), "notification123").
					Return(notification, nil)
			},
			expectedError: "",
		},
		{
			name:           "repository error",
			notificationID: "notification123",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindByID(gomock.Any(), "notification123").
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to find notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			notification, err := useCase.GetNotification(context.Background(), tt.notificationID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, notification)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, notification)
			}
		})
	}
}
