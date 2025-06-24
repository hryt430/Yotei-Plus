package tokenService

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	"github.com/hryt430/Yotei+/internal/modules/auth/usecase/token/mocks"
	"github.com/hryt430/Yotei+/pkg/token"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go -package=mocks

func TestTokenService_GenerateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,      // access token duration
		7*24*time.Hour, // refresh token duration
	)

	tests := []struct {
		name          string
		user          *domain.User
		expectedError string
	}{
		{
			name: "successful access token generation",
			user: &domain.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Role:     domain.RoleUser,
			},
			expectedError: "",
		},
		{
			name: "generate token for admin user",
			user: &domain.User{
				ID:       uuid.New(),
				Email:    "admin@example.com",
				Username: "admin",
				Role:     domain.RoleAdmin,
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken, err := service.GenerateAccessToken(tt.user)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, accessToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, accessToken)

				// Verify token can be parsed and contains correct claims
				claims, err := jwtManager.Verify(accessToken)
				assert.NoError(t, err)
				assert.Equal(t, tt.user.ID.String(), claims.UserID)
				assert.Equal(t, tt.user.Email, claims.Email)
				assert.Equal(t, tt.user.Username, claims.Username)
				assert.Equal(t, tt.user.Role, claims.Role)
			}
		})
	}
}

func TestTokenService_GenerateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Role:     domain.RoleUser,
	}

	tests := []struct {
		name          string
		setupMocks    func()
		expectedError string
	}{
		{
			name: "successful refresh token generation",
			setupMocks: func() {
				mockRepo.EXPECT().
					SaveRefreshToken(gomock.Any()).
					Do(func(refreshToken *domain.RefreshToken) {
						assert.Equal(t, user.ID, refreshToken.UserID)
						assert.NotEmpty(t, refreshToken.Token)
						assert.True(t, refreshToken.ExpiresAt.After(time.Now()))
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "repository save error",
			setupMocks: func() {
				mockRepo.EXPECT().
					SaveRefreshToken(gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			refreshToken, err := service.GenerateRefreshToken(user)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, refreshToken)
			}
		})
	}
}

func TestTokenService_ValidateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Role:     domain.RoleUser,
	}

	// Generate a valid token for testing
	validToken, err := service.GenerateAccessToken(user)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		tokenString   string
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			setupMocks: func() {
				mockRepo.EXPECT().
					IsTokenBlacklisted(validToken).
					Return(false)
			},
			expectedError: "",
		},
		{
			name:        "blacklisted token",
			tokenString: validToken,
			setupMocks: func() {
				mockRepo.EXPECT().
					IsTokenBlacklisted(validToken).
					Return(true)
			},
			expectedError: "token has been revoked",
		},
		{
			name:        "invalid token format",
			tokenString: "invalid-token",
			setupMocks: func() {
				mockRepo.EXPECT().
					IsTokenBlacklisted("invalid-token").
					Return(false)
			},
			expectedError: "invalid token",
		},
		{
			name:        "expired token",
			tokenString: generateExpiredToken(t, jwtManager, user),
			setupMocks: func() {
				expiredToken := generateExpiredToken(t, jwtManager, user)
				mockRepo.EXPECT().
					IsTokenBlacklisted(expiredToken).
					Return(false)
			},
			expectedError: "token has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			claims, err := service.ValidateAccessToken(tt.tokenString)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, user.Email, claims.Email)
				assert.Equal(t, user.Username, claims.Username)
			}
		})
	}
}

func TestTokenService_ValidateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	userID := uuid.New()
	tokenString := "test-refresh-token"

	tests := []struct {
		name          string
		tokenString   string
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "valid refresh token",
			tokenString: tokenString,
			setupMocks: func() {
				validToken := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     tokenString,
					UserID:    userID,
					ExpiresAt: time.Now().Add(time.Hour),
					IssuedAt:  time.Now().Add(-time.Hour),
					RevokedAt: nil,
				}

				mockRepo.EXPECT().
					FindRefreshToken(tokenString).
					Return(validToken, nil)
			},
			expectedError: "",
		},
		{
			name:        "token not found",
			tokenString: "nonexistent-token",
			setupMocks: func() {
				mockRepo.EXPECT().
					FindRefreshToken("nonexistent-token").
					Return(nil, nil)
			},
			expectedError: "refresh token not found",
		},
		{
			name:        "repository error",
			tokenString: tokenString,
			setupMocks: func() {
				mockRepo.EXPECT().
					FindRefreshToken(tokenString).
					Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:        "revoked token",
			tokenString: tokenString,
			setupMocks: func() {
				revokedTime := time.Now().Add(-time.Minute)
				revokedToken := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     tokenString,
					UserID:    userID,
					ExpiresAt: time.Now().Add(time.Hour),
					IssuedAt:  time.Now().Add(-time.Hour),
					RevokedAt: &revokedTime,
				}

				mockRepo.EXPECT().
					FindRefreshToken(tokenString).
					Return(revokedToken, nil)
			},
			expectedError: "refresh token has been revoked",
		},
		{
			name:        "expired token",
			tokenString: tokenString,
			setupMocks: func() {
				expiredToken := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     tokenString,
					UserID:    userID,
					ExpiresAt: time.Now().Add(-time.Hour), // Expired
					IssuedAt:  time.Now().Add(-2 * time.Hour),
					RevokedAt: nil,
				}

				mockRepo.EXPECT().
					FindRefreshToken(tokenString).
					Return(expiredToken, nil)
			},
			expectedError: "refresh token has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			refreshToken, err := service.ValidateRefreshToken(tt.tokenString)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, refreshToken)
				assert.Equal(t, userID, refreshToken.UserID)
			}
		})
	}
}

func TestTokenService_RevokeAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Role:     domain.RoleUser,
	}

	validToken, err := service.GenerateAccessToken(user)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		tokenString   string
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful token revocation",
			tokenString: validToken,
			setupMocks: func() {
				mockRepo.EXPECT().
					SaveTokenToBlacklist(validToken, gomock.Any()).
					Do(func(token string, ttl time.Duration) {
						assert.Equal(t, validToken, token)
						assert.True(t, ttl > 0)
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "invalid token format",
			tokenString: "invalid-token",
			setupMocks:  func() {
				// No mock expectations as the token should fail parsing
			},
			expectedError: "token is malformed",
		},
		{
			name:        "repository error",
			tokenString: validToken,
			setupMocks: func() {
				mockRepo.EXPECT().
					SaveTokenToBlacklist(validToken, gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.RevokeAccessToken(tt.tokenString)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenService_RevokeToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	refreshTokenString := "test-refresh-token"

	tests := []struct {
		name          string
		tokenString   string
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful refresh token revocation",
			tokenString: refreshTokenString,
			setupMocks: func() {
				mockRepo.EXPECT().
					RevokeRefreshToken(refreshTokenString).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "repository error",
			tokenString: refreshTokenString,
			setupMocks: func() {
				mockRepo.EXPECT().
					RevokeRefreshToken(refreshTokenString).
					Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.RevokeToken(tt.tokenString)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenService_GenerateNewRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "successful new refresh token generation",
			userID: userID,
			setupMocks: func() {
				mockRepo.EXPECT().
					SaveRefreshToken(gomock.Any()).
					Do(func(refreshToken *domain.RefreshToken) {
						assert.Equal(t, userID, refreshToken.UserID)
						assert.NotEmpty(t, refreshToken.Token)
						assert.True(t, refreshToken.ExpiresAt.After(time.Now()))
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "repository save error",
			userID: userID,
			setupMocks: func() {
				mockRepo.EXPECT().
					SaveRefreshToken(gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			refreshToken, err := service.GenerateNewRefreshToken(tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, refreshToken)
				assert.Equal(t, tt.userID, refreshToken.UserID)
				assert.NotEmpty(t, refreshToken.Token)
			}
		})
	}
}

func TestTokenService_CleanupExpiredTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	tests := []struct {
		name          string
		setupMocks    func()
		expectedError string
	}{
		{
			name: "successful cleanup",
			setupMocks: func() {
				mockRepo.EXPECT().
					DeleteExpiredRefreshTokens().
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			setupMocks: func() {
				mockRepo.EXPECT().
					DeleteExpiredRefreshTokens().
					Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.CleanupExpiredTokens()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenService_IsTokenRevoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockITokenRepository(ctrl)
	jwtManager := token.NewJWTManager("test-secret-key", "test-issuer")

	service := NewTokenService(
		mockRepo,
		jwtManager,
		time.Hour,
		7*24*time.Hour,
	)

	tokenString := "test-token"

	tests := []struct {
		name           string
		tokenString    string
		setupMocks     func()
		expectedResult bool
	}{
		{
			name:        "token is revoked",
			tokenString: tokenString,
			setupMocks: func() {
				mockRepo.EXPECT().
					IsTokenBlacklisted(tokenString).
					Return(true)
			},
			expectedResult: true,
		},
		{
			name:        "token is not revoked",
			tokenString: tokenString,
			setupMocks: func() {
				mockRepo.EXPECT().
					IsTokenBlacklisted(tokenString).
					Return(false)
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result := service.IsTokenRevoked(tt.tokenString)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// Helper function to generate an expired token for testing
func generateExpiredToken(t *testing.T, jwtManager *token.JWTManager, user *domain.User) string {
	claims := &token.Claims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}

	// Generate a token with negative duration (expired)
	expiredToken, err := jwtManager.Generate(claims, -time.Hour)
	assert.NoError(t, err)
	return expiredToken
}