package authService

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
)

//go:generate mockgen -source=user/repository.go -destination=mocks/mock_user_repository.go
//go:generate mockgen -source=token/repository.go -destination=mocks/mock_token_repository.go
//go:generate mockgen -package=authService -destination=mocks/mock_services.go github.com/hryt430/Yotei+/internal/modules/auth/usecase/user UserService
//go:generate mockgen -package=authService -destination=mocks/mock_token_service.go github.com/hryt430/Yotei+/internal/modules/auth/usecase/token TokenService

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockAuthRepo := mock.NewMockIAuthRepository(ctrl)

	authService := mock.NewAuthService(mockAuthRepo, mockUserService, mockTokenService)

	tests := []struct {
		name          string
		email         string
		username      string
		password      string
		setupMocks    func()
		expectedError string
	}{
		{
			name:     "successful registration",
			email:    "test@example.com",
			username: "testuser",
			password: "password123",
			setupMocks: func() {
				// FindUserByEmail should return nil (user doesn't exist)
				mockUserService.EXPECT().
					FindUserByEmail("test@example.com").
					Return(nil, errors.New("user not found"))

				// CreateUser should be called and return the new user
				expectedUser := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Role:     "user",
				}

				mockUserService.EXPECT().
					CreateUser(gomock.Any()).
					Do(func(user *domain.User) {
						assert.Equal(t, "test@example.com", user.Email)
						assert.Equal(t, "testuser", user.Username)
						assert.Equal(t, "user", user.Role)
						// Password should be hashed, so it won't equal the original
						assert.NotEqual(t, "password123", user.Password)
					}).
					Return(expectedUser, nil)
			},
			expectedError: "",
		},
		{
			name:     "email already exists",
			email:    "existing@example.com",
			username: "testuser",
			password: "password123",
			setupMocks: func() {
				existingUser := &domain.User{
					ID:       uuid.New(),
					Email:    "existing@example.com",
					Username: "existinguser",
				}

				mockUserService.EXPECT().
					FindUserByEmail("existing@example.com").
					Return(existingUser, nil)
			},
			expectedError: "email already exists",
		},
		{
			name:     "user service error during creation",
			email:    "test@example.com",
			username: "testuser",
			password: "password123",
			setupMocks: func() {
				mockUserService.EXPECT().
					FindUserByEmail("test@example.com").
					Return(nil, errors.New("user not found"))

				mockUserService.EXPECT().
					CreateUser(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := authService.Register(tt.email, tt.username, tt.password)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.username, user.Username)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock.NewMockUserService(ctrl)
	mockTokenService := mock.NewMockTokenService(ctrl)
	mockAuthRepo := mock.NewMockIAuthRepository(ctrl)

	authService := mock.NewAuthService(mockAuthRepo, mockUserService, mockTokenService)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func()
		expectedError string
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() {
				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Password: "$2a$10$hashedpassword", // Mock hashed password
				}

				mockUserService.EXPECT().
					FindUserByEmail("test@example.com").
					Return(user, nil)

				// Mock password verification (we'll assume it passes)
				// In real implementation, this would use bcrypt.CompareHashAndPassword

				mockUserService.EXPECT().
					UpdateLastLogin(user.ID).
					Return(nil)

				mockTokenService.EXPECT().
					GenerateAccessToken(user).
					Return("access_token_string", nil)

				mockTokenService.EXPECT().
					GenerateRefreshToken(user).
					Return("refresh_token_string", nil)
			},
			expectedError: "",
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func() {
				mockUserService.EXPECT().
					FindUserByEmail("nonexistent@example.com").
					Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
		{
			name:     "user service returns nil user",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() {
				mockUserService.EXPECT().
					FindUserByEmail("test@example.com").
					Return(nil, nil)
			},
			expectedError: "invalid email or password",
		},
		{
			name:     "token generation fails",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() {
				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Password: "$2a$10$hashedpassword",
				}

				mockUserService.EXPECT().
					FindUserByEmail("test@example.com").
					Return(user, nil)

				mockUserService.EXPECT().
					UpdateLastLogin(user.ID).
					Return(nil)

				mockTokenService.EXPECT().
					GenerateAccessToken(user).
					Return("", errors.New("token generation failed"))
			},
			expectedError: "token generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			accessToken, refreshToken, err := authService.Login(tt.email, tt.password)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock.NewMockUserService(ctrl)
	mockTokenService := mock.NewMockTokenService(ctrl)
	mockAuthRepo := mock.NewMockIAuthRepository(ctrl)

	authService := mock.NewAuthService(mockAuthRepo, mockUserService, mockTokenService)

	tests := []struct {
		name          string
		refreshToken  string
		setupMocks    func()
		expectedError string
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid_refresh_token",
			setupMocks: func() {
				userID := uuid.New()
				refreshTokenEntity := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     "valid_refresh_token",
					UserID:    userID,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					RevokedAt: nil,
				}

				user := &domain.User{
					ID:       userID,
					Email:    "test@example.com",
					Username: "testuser",
				}

				mockTokenService.EXPECT().
					TokenRepository.
					FindRefreshToken("valid_refresh_token").
					Return(refreshTokenEntity, nil)

				mockUserService.EXPECT().
					FindUserByID(userID).
					Return(user, nil)

				mockTokenService.EXPECT().
					GenerateAccessToken(user).
					Return("new_access_token", nil)

				mockTokenService.EXPECT().
					RevokeToken("valid_refresh_token").
					Return(nil)

				mockTokenService.EXPECT().
					GenerateRefreshToken(user).
					Return("new_refresh_token", nil)
			},
			expectedError: "",
		},
		{
			name:         "refresh token not found",
			refreshToken: "nonexistent_token",
			setupMocks: func() {
				mockTokenService.EXPECT().
					TokenRepository.
					FindRefreshToken("nonexistent_token").
					Return(nil, errors.New("token not found"))
			},
			expectedError: "token not found",
		},
		{
			name:         "refresh token is revoked",
			refreshToken: "revoked_token",
			setupMocks: func() {
				revokedTime := time.Now().Add(-time.Hour)
				refreshTokenEntity := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     "revoked_token",
					UserID:    uuid.New(),
					ExpiresAt: time.Now().Add(24 * time.Hour),
					RevokedAt: &revokedTime,
				}

				mockTokenService.EXPECT().
					TokenRepository.
					FindRefreshToken("revoked_token").
					Return(refreshTokenEntity, nil)
			},
			expectedError: "invalid refresh token",
		},
		{
			name:         "refresh token is expired",
			refreshToken: "expired_token",
			setupMocks: func() {
				refreshTokenEntity := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     "expired_token",
					UserID:    uuid.New(),
					ExpiresAt: time.Now().Add(-time.Hour), // Expired
					RevokedAt: nil,
				}

				mockTokenService.EXPECT().
					TokenRepository.
					FindRefreshToken("expired_token").
					Return(refreshTokenEntity, nil)
			},
			expectedError: "refresh token expired",
		},
		{
			name:         "user not found for refresh token",
			refreshToken: "valid_token_but_no_user",
			setupMocks: func() {
				userID := uuid.New()
				refreshTokenEntity := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     "valid_token_but_no_user",
					UserID:    userID,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					RevokedAt: nil,
				}

				mockTokenService.EXPECT().
					TokenRepository.
					FindRefreshToken("valid_token_but_no_user").
					Return(refreshTokenEntity, nil)

				mockUserService.EXPECT().
					FindUserByID(userID).
					Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			newAccessToken, newRefreshToken, err := authService.RefreshToken(tt.refreshToken)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, newAccessToken)
				assert.Empty(t, newRefreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, newAccessToken)
				assert.NotEmpty(t, newRefreshToken)
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock.NewMockUserService(ctrl)
	mockTokenService := mock.NewMockTokenService(ctrl)
	mockAuthRepo := mock.NewMockIAuthRepository(ctrl)

	authService := mock.NewAuthService(mockAuthRepo, mockUserService, mockTokenService)

	tests := []struct {
		name          string
		accessToken   string
		refreshToken  string
		setupMocks    func()
		expectedError string
	}{
		{
			name:         "successful logout",
			accessToken:  "valid_access_token",
			refreshToken: "valid_refresh_token",
			setupMocks: func() {
				mockTokenService.EXPECT().
					RevokeAccessToken("valid_access_token").
					Return(nil)

				mockTokenService.EXPECT().
					RevokeToken("valid_refresh_token").
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:         "access token revocation fails",
			accessToken:  "invalid_access_token",
			refreshToken: "valid_refresh_token",
			setupMocks: func() {
				mockTokenService.EXPECT().
					RevokeAccessToken("invalid_access_token").
					Return(errors.New("failed to revoke access token"))
			},
			expectedError: "failed to revoke access token",
		},
		{
			name:         "refresh token revocation fails",
			accessToken:  "valid_access_token",
			refreshToken: "invalid_refresh_token",
			setupMocks: func() {
				mockTokenService.EXPECT().
					RevokeAccessToken("valid_access_token").
					Return(nil)

				mockTokenService.EXPECT().
					RevokeToken("invalid_refresh_token").
					Return(errors.New("failed to revoke refresh token"))
			},
			expectedError: "failed to revoke refresh token",
		},
		{
			name:         "empty tokens",
			accessToken:  "",
			refreshToken: "",
			setupMocks: func() {
				mockTokenService.EXPECT().
					RevokeAccessToken("").
					Return(nil)

				mockTokenService.EXPECT().
					RevokeToken("").
					Return(nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := authService.Logout(tt.accessToken, tt.refreshToken)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Integration-style test that exercises multiple methods
func TestAuthService_FullAuthFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock.NewMockUserService(ctrl)
	mockTokenService := mock.NewMockTokenService(ctrl)
	mockAuthRepo := mock.NewMockIAuthRepository(ctrl)

	authService := mock.NewAuthService(mockAuthRepo, mockUserService, mockTokenService)

	email := "integration@example.com"
	username := "integrationuser"
	password := "password123"
	userID := uuid.New()

	// Step 1: Register
	t.Run("register", func(t *testing.T) {
		mockUserService.EXPECT().
			FindUserByEmail(email).
			Return(nil, errors.New("user not found"))

		expectedUser := &domain.User{
			ID:       userID,
			Email:    email,
			Username: username,
			Role:     "user",
		}

		mockUserService.EXPECT().
			CreateUser(gomock.Any()).
			Return(expectedUser, nil)

		user, err := authService.Register(email, username, password)
		require.NoError(t, err)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, username, user.Username)
	})

	// Step 2: Login
	var accessToken, refreshToken string
	t.Run("login", func(t *testing.T) {
		user := &domain.User{
			ID:       userID,
			Email:    email,
			Username: username,
			Password: "$2a$10$hashedpassword",
		}

		mockUserService.EXPECT().
			FindUserByEmail(email).
			Return(user, nil)

		mockUserService.EXPECT().
			UpdateLastLogin(userID).
			Return(nil)

		mockTokenService.EXPECT().
			GenerateAccessToken(user).
			Return("test_access_token", nil)

		mockTokenService.EXPECT().
			GenerateRefreshToken(user).
			Return("test_refresh_token", nil)

		at, rt, err := authService.Login(email, password)
		require.NoError(t, err)
		accessToken = at
		refreshToken = rt
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	// Step 3: Logout
	t.Run("logout", func(t *testing.T) {
		mockTokenService.EXPECT().
			RevokeAccessToken(accessToken).
			Return(nil)

		mockTokenService.EXPECT().
			RevokeToken(refreshToken).
			Return(nil)

		err := authService.Logout(accessToken, refreshToken)
		assert.NoError(t, err)
	})
}
