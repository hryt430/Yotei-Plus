package userService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	"github.com/hryt430/Yotei+/internal/modules/auth/usecase/user/mocks"
	"github.com/hryt430/Yotei+/pkg/utils"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go -package=mocks

func TestUserService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	service := NewUserService(mockRepo)

	tests := []struct {
		name          string
		user          *domain.User
		setupMocks    func()
		expectedError string
	}{
		{
			name: "successful user creation",
			user: &domain.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Password: "plainpassword",
				Role:     domain.RoleUser,
			},
			setupMocks: func() {
				// Check email doesn't exist (return nil user, nil error to indicate no user found)
				mockRepo.EXPECT().
					FindUserByEmail("test@example.com").
					Return(nil, nil)

				// Create user
				mockRepo.EXPECT().
					CreateUser(gomock.Any()).
					Do(func(user *domain.User) {
						// Verify password is hashed
						assert.NotEqual(t, "plainpassword", user.Password)
						assert.True(t, len(user.Password) > 20) // Hashed password should be longer
						assert.Equal(t, "test@example.com", user.Email)
						assert.Equal(t, "testuser", user.Username)
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "email already exists",
			user: &domain.User{
				ID:       uuid.New(),
				Email:    "existing@example.com",
				Username: "testuser",
				Password: "plainpassword",
			},
			setupMocks: func() {
				existingUser := &domain.User{
					ID:       uuid.New(),
					Email:    "existing@example.com",
					Username: "existinguser",
				}

				mockRepo.EXPECT().
					FindUserByEmail("existing@example.com").
					Return(existingUser, nil)
			},
			expectedError: "email already exists",
		},
		{
			name: "repository error during email check",
			user: &domain.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Password: "plainpassword",
			},
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByEmail("test@example.com").
					Return(nil, errors.New("database connection error"))
			},
			expectedError: "database connection error",
		},
		{
			name: "repository error during creation",
			user: &domain.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Password: "plainpassword",
			},
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByEmail("test@example.com").
					Return(nil, nil)

				mockRepo.EXPECT().
					CreateUser(gomock.Any()).
					Return(errors.New("database write error"))
			},
			expectedError: "database write error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.CreateUser(tt.user)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.user.Email, result.Email)
				assert.Equal(t, tt.user.Username, result.Username)
				// Password should be hashed
				assert.NotEqual(t, "plainpassword", result.Password)
			}
		})
	}
}

func TestUserService_GetUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	service := NewUserService(mockRepo)

	tests := []struct {
		name          string
		search        string
		setupMocks    func()
		expectedCount int
		expectedError string
	}{
		{
			name:   "successful get users without search",
			search: "",
			setupMocks: func() {
				users := []*domain.User{
					{ID: uuid.New(), Username: "user1", Email: "user1@example.com"},
					{ID: uuid.New(), Username: "user2", Email: "user2@example.com"},
				}

				mockRepo.EXPECT().
					FindUsers("").
					Return(users, nil)
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name:   "successful get users with search",
			search: "test",
			setupMocks: func() {
				users := []*domain.User{
					{ID: uuid.New(), Username: "testuser", Email: "test@example.com"},
				}

				mockRepo.EXPECT().
					FindUsers("test").
					Return(users, nil)
			},
			expectedCount: 1,
			expectedError: "",
		},
		{
			name:   "repository error",
			search: "",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUsers("").
					Return(nil, errors.New("database error"))
			},
			expectedCount: 0,
			expectedError: "database error",
		},
		{
			name:   "no users found",
			search: "nonexistent",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUsers("nonexistent").
					Return([]*domain.User{}, nil)
			},
			expectedCount: 0,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			users, err := service.GetUsers(context.Background(), tt.search)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.expectedCount)
			}
		})
	}
}

func TestUserService_FindUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	service := NewUserService(mockRepo)

	tests := []struct {
		name          string
		email         string
		setupMocks    func()
		expectedUser  *domain.User
		expectedError string
	}{
		{
			name:  "user found",
			email: "test@example.com",
			setupMocks: func() {
				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
				}

				mockRepo.EXPECT().
					FindUserByEmail("test@example.com").
					Return(user, nil)
			},
			expectedUser: &domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			expectedError: "",
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByEmail("nonexistent@example.com").
					Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: "",
		},
		{
			name:  "repository error",
			email: "test@example.com",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByEmail("test@example.com").
					Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := service.FindUserByEmail(tt.email)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tt.expectedUser != nil {
					assert.NotNil(t, user)
					assert.Equal(t, tt.expectedUser.Email, user.Email)
					assert.Equal(t, tt.expectedUser.Username, user.Username)
				} else {
					assert.Nil(t, user)
				}
			}
		})
	}
}

func TestUserService_FindUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	service := NewUserService(mockRepo)

	userID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		setupMocks    func()
		expectedUser  *domain.User
		expectedError string
	}{
		{
			name: "user found",
			id:   userID,
			setupMocks: func() {
				user := &domain.User{
					ID:       userID,
					Email:    "test@example.com",
					Username: "testuser",
				}

				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(user, nil)
			},
			expectedUser: &domain.User{
				ID:       userID,
				Email:    "test@example.com",
				Username: "testuser",
			},
			expectedError: "",
		},
		{
			name: "user not found",
			id:   userID,
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: "",
		},
		{
			name: "repository error",
			id:   userID,
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := service.FindUserByID(tt.id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tt.expectedUser != nil {
					assert.NotNil(t, user)
					assert.Equal(t, tt.expectedUser.ID, user.ID)
					assert.Equal(t, tt.expectedUser.Email, user.Email)
				} else {
					assert.Nil(t, user)
				}
			}
		})
	}
}

func TestUserService_UpdateUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	service := NewUserService(mockRepo)

	userID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		username      string
		email         string
		setupMocks    func()
		expectedError string
	}{
		{
			name:     "successful email update",
			id:       userID,
			username: "",
			email:    "newemail@example.com",
			setupMocks: func() {
				// Create a copy of originalUser to avoid modification
				pastTime := time.Now().Add(-time.Hour)
				userCopy := &domain.User{
					ID:            userID,
					Email:         "original@example.com",
					Username:      "originaluser",
					EmailVerified: true,
					CreatedAt:     time.Now().Add(-24 * time.Hour),
					UpdatedAt:     pastTime,
				}
				
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(userCopy, nil)

				// Check email doesn't exist for other users
				mockRepo.EXPECT().
					FindUserByEmail("newemail@example.com").
					Return(nil, nil)

				mockRepo.EXPECT().
					UpdateUser(gomock.Any()).
					Do(func(user *domain.User) {
						assert.Equal(t, "newemail@example.com", user.Email)
						assert.False(t, user.EmailVerified) // Should be reset
						assert.True(t, user.UpdatedAt.After(pastTime))
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:     "no changes",
			id:       userID,
			username: "",
			email:    "",
			setupMocks: func() {
				// Create a copy of originalUser to avoid modification
				userCopy := &domain.User{
					ID:            userID,
					Email:         "original@example.com",
					Username:      "originaluser",
					EmailVerified: true,
					CreatedAt:     time.Now().Add(-24 * time.Hour),
					UpdatedAt:     time.Now().Add(-time.Hour),
				}
				
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(userCopy, nil)

				// No update call expected
			},
			expectedError: "",
		},
		{
			name:     "email already exists for different user",
			id:       userID,
			username: "",
			email:    "existing@example.com",
			setupMocks: func() {
				// Create a copy of originalUser to avoid modification
				userCopy := &domain.User{
					ID:            userID,
					Email:         "original@example.com",
					Username:      "originaluser",
					EmailVerified: true,
					CreatedAt:     time.Now().Add(-24 * time.Hour),
					UpdatedAt:     time.Now().Add(-time.Hour),
				}
				
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(userCopy, nil)

				existingUser := &domain.User{
					ID:    uuid.New(), // Different ID
					Email: "existing@example.com",
				}

				mockRepo.EXPECT().
					FindUserByEmail("existing@example.com").
					Return(existingUser, nil)
			},
			expectedError: "email already exists",
		},
		{
			name:     "user not found",
			id:       userID,
			username: "",
			email:    "newemail@example.com",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name:     "repository update error",
			id:       userID,
			username: "",
			email:    "newemail@example.com",
			setupMocks: func() {
				// Create a copy of originalUser to avoid modification
				userCopy := &domain.User{
					ID:            userID,
					Email:         "original@example.com",
					Username:      "originaluser",
					EmailVerified: true,
					CreatedAt:     time.Now().Add(-24 * time.Hour),
					UpdatedAt:     time.Now().Add(-time.Hour),
				}
				
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(userCopy, nil)

				mockRepo.EXPECT().
					FindUserByEmail("newemail@example.com").
					Return(nil, nil)

				mockRepo.EXPECT().
					UpdateUser(gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := service.UpdateUserProfile(tt.id, tt.username, tt.email)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.email != "" {
					assert.Equal(t, tt.email, user.Email)
				}
			}
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	service := NewUserService(mockRepo)

	userID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		oldPassword   string
		newPassword   string
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful password change",
			id:          userID,
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMocks: func() {
				// Hash the old password for testing
				hashedOldPassword, _ := utils.HashPassword("oldpassword")
				user := &domain.User{
					ID:       userID,
					Password: hashedOldPassword,
				}

				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(user, nil)

				mockRepo.EXPECT().
					UpdateUser(gomock.Any()).
					Do(func(updatedUser *domain.User) {
						// Password should be hashed and different from old
						assert.NotEqual(t, hashedOldPassword, updatedUser.Password)
						assert.NotEqual(t, "newpassword123", updatedUser.Password)
						assert.True(t, len(updatedUser.Password) > 20) // Hashed
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "user not found",
			id:          userID,
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name:        "repository find error",
			id:          userID,
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "repository update error",
			id:          userID,
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMocks: func() {
				// Hash the old password for testing
				hashedOldPassword, _ := utils.HashPassword("oldpassword")
				user := &domain.User{
					ID:       userID,
					Password: hashedOldPassword,
				}

				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(user, nil)

				mockRepo.EXPECT().
					UpdateUser(gomock.Any()).
					Return(errors.New("database update error"))
			},
			expectedError: "database update error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.ChangePassword(tt.id, tt.oldPassword, tt.newPassword)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_UpdateLastLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	service := NewUserService(mockRepo)

	userID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name: "successful last login update",
			id:   userID,
			setupMocks: func() {
				user := &domain.User{
					ID:        userID,
					Email:     "test@example.com",
					LastLogin: nil, // Initially nil
				}

				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(user, nil)

				mockRepo.EXPECT().
					UpdateUser(gomock.Any()).
					Do(func(updatedUser *domain.User) {
						assert.NotNil(t, updatedUser.LastLogin)
						// Should be very recent
						timeSince := time.Since(*updatedUser.LastLogin)
						assert.True(t, timeSince < time.Second)
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "user not found",
			id:   userID,
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name: "repository find error",
			id:   userID,
			setupMocks: func() {
				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name: "repository update error",
			id:   userID,
			setupMocks: func() {
				user := &domain.User{
					ID:        userID,
					Email:     "test@example.com",
					LastLogin: nil,
				}

				mockRepo.EXPECT().
					FindUserByID(userID).
					Return(user, nil)

				mockRepo.EXPECT().
					UpdateUser(gomock.Any()).
					Return(errors.New("update error"))
			},
			expectedError: "update error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.UpdateLastLogin(tt.id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
