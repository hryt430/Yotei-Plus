package authService

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	userService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/user"
	"github.com/hryt430/Yotei+/pkg/token"
	"github.com/hryt430/Yotei+/pkg/utils"
)

// MockUserRepository はテスト用のユーザーリポジトリモック
type MockUserRepository struct {
	CreateUserFunc      func(user *domain.User) error
	FindUserByEmailFunc func(email string) (*domain.User, error)
	FindUserByIDFunc    func(id uuid.UUID) (*domain.User, error)
	FindUsersFunc       func(search string) ([]*domain.User, error)
	UpdateUserFunc      func(user *domain.User) error
}

func (m *MockUserRepository) CreateUser(user *domain.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(user)
	}
	return nil
}

func (m *MockUserRepository) FindUserByEmail(email string) (*domain.User, error) {
	if m.FindUserByEmailFunc != nil {
		return m.FindUserByEmailFunc(email)
	}
	return nil, nil
}

func (m *MockUserRepository) FindUserByID(id uuid.UUID) (*domain.User, error) {
	if m.FindUserByIDFunc != nil {
		return m.FindUserByIDFunc(id)
	}
	return nil, nil
}

func (m *MockUserRepository) FindUsers(search string) ([]*domain.User, error) {
	if m.FindUsersFunc != nil {
		return m.FindUsersFunc(search)
	}
	return []*domain.User{}, nil
}

func (m *MockUserRepository) UpdateUser(user *domain.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(user)
	}
	return nil
}

// MockTokenRepository はテスト用のトークンリポジトリモック
type MockTokenRepository struct {
	SaveTokenToBlacklistFunc       func(token string, ttl time.Duration) error
	IsTokenBlacklistedFunc         func(token string) bool
	SaveRefreshTokenFunc           func(token *domain.RefreshToken) error
	FindRefreshTokenFunc           func(token string) (*domain.RefreshToken, error)
	RevokeRefreshTokenFunc         func(token string) error
	DeleteExpiredRefreshTokensFunc func() error
}

func (m *MockTokenRepository) SaveTokenToBlacklist(token string, ttl time.Duration) error {
	if m.SaveTokenToBlacklistFunc != nil {
		return m.SaveTokenToBlacklistFunc(token, ttl)
	}
	return nil
}

func (m *MockTokenRepository) IsTokenBlacklisted(token string) bool {
	if m.IsTokenBlacklistedFunc != nil {
		return m.IsTokenBlacklistedFunc(token)
	}
	return false
}

func (m *MockTokenRepository) SaveRefreshToken(token *domain.RefreshToken) error {
	if m.SaveRefreshTokenFunc != nil {
		return m.SaveRefreshTokenFunc(token)
	}
	return nil
}

func (m *MockTokenRepository) FindRefreshToken(token string) (*domain.RefreshToken, error) {
	if m.FindRefreshTokenFunc != nil {
		return m.FindRefreshTokenFunc(token)
	}
	return nil, nil
}

func (m *MockTokenRepository) RevokeRefreshToken(token string) error {
	if m.RevokeRefreshTokenFunc != nil {
		return m.RevokeRefreshTokenFunc(token)
	}
	return nil
}

func (m *MockTokenRepository) DeleteExpiredRefreshTokens() error {
	if m.DeleteExpiredRefreshTokensFunc != nil {
		return m.DeleteExpiredRefreshTokensFunc()
	}
	return nil
}

// テスト用のJWTManagerを作成
func createTestJWTManager() *token.JWTManager {
	return token.NewJWTManager("test_secret_key", "test_issuer")
}

// テスト用のサービスを作成する関数
func createTestServices(userRepo userService.IUserRepository, tokenRepo tokenService.ITokenRepository) (*userService.UserService, *tokenService.TokenService) {
	// UserServiceを作成
	userSvc := userService.NewUserService(userRepo)

	// TokenServiceを作成
	jwtManager := createTestJWTManager()
	tokenSvc := tokenService.NewTokenService(
		tokenRepo,
		jwtManager,
		1*time.Hour,    // アクセストークンの有効期限
		7*24*time.Hour, // リフレッシュトークンの有効期限
	)

	return userSvc, tokenSvc
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		username      string
		password      string
		setupMocks    func() (*MockUserRepository, *MockTokenRepository)
		expectedError string
	}{
		{
			name:     "successful registration",
			email:    "test@example.com",
			username: "testuser",
			password: "password123",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						return nil, nil // ユーザーが存在しない場合は(nil, nil)を返す
					},
					CreateUserFunc: func(user *domain.User) error {
						// Verify the user properties
						assert.Equal(t, "test@example.com", user.Email)
						assert.Equal(t, "testuser", user.Username)
						assert.Equal(t, "user", user.Role)
						// UserService.CreateUserでハッシュ化されているので、元のパスワードとは異なるはず
						assert.NotEqual(t, "password123", user.Password)
						// ハッシュ化されたパスワードは空でないはず
						assert.NotEmpty(t, user.Password)
						return nil
					},
				}

				mockTokenRepo := &MockTokenRepository{}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "",
		},
		{
			name:     "email already exists",
			email:    "existing@example.com",
			username: "testuser",
			password: "password123",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				existingUser := &domain.User{
					ID:       uuid.New(),
					Email:    "existing@example.com",
					Username: "existinguser",
				}

				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						return existingUser, nil // ユーザーが存在する場合は(user, nil)を返す
					},
				}

				mockTokenRepo := &MockTokenRepository{}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "email already exists",
		},
		{
			name:     "user creation error",
			email:    "test@example.com",
			username: "testuser",
			password: "password123",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						return nil, nil // ユーザーが存在しない
					},
					CreateUserFunc: func(user *domain.User) error {
						return errors.New("database error")
					},
				}

				mockTokenRepo := &MockTokenRepository{}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo, mockTokenRepo := tt.setupMocks()

			// 実際のサービスを作成（依存関係はモックを注入）
			userSvc, tokenSvc := createTestServices(mockUserRepo, mockTokenRepo)

			// AuthServiceを作成
			service := NewAuthService(nil, *userSvc, *tokenSvc)

			user, err := service.Register(tt.email, tt.username, tt.password)

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
	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func() (*MockUserRepository, *MockTokenRepository)
		expectedError string
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				// 実際にパスワードをハッシュ化
				hashedPassword, err := utils.HashPassword("password123")
				assert.NoError(t, err)

				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Password: hashedPassword,
				}

				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						if email == "test@example.com" {
							return user, nil
						}
						return nil, nil
					},
					FindUserByIDFunc: func(id uuid.UUID) (*domain.User, error) {
						if id == user.ID {
							return user, nil
						}
						return nil, errors.New("user not found")
					},
					UpdateUserFunc: func(user *domain.User) error {
						return nil
					},
				}

				mockTokenRepo := &MockTokenRepository{
					SaveRefreshTokenFunc: func(token *domain.RefreshToken) error {
						return nil
					},
				}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "",
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						return nil, nil // ユーザーが存在しない場合
					},
				}

				mockTokenRepo := &MockTokenRepository{}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "invalid email or password",
		},
		{
			name:     "incorrect password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				// 正しいパスワードをハッシュ化
				hashedPassword, err := utils.HashPassword("password123")
				assert.NoError(t, err)

				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Password: hashedPassword,
				}

				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						if email == "test@example.com" {
							return user, nil
						}
						return nil, nil
					},
				}

				mockTokenRepo := &MockTokenRepository{}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "invalid email or password",
		},
		{
			name:     "user service find error",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						return nil, errors.New("database connection error")
					},
				}

				mockTokenRepo := &MockTokenRepository{}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "database connection error",
		},
		{
			name:     "token generation failure",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				// 実際にパスワードをハッシュ化
				hashedPassword, err := utils.HashPassword("password123")
				assert.NoError(t, err)

				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Password: hashedPassword,
				}

				mockUserRepo := &MockUserRepository{
					FindUserByEmailFunc: func(email string) (*domain.User, error) {
						if email == "test@example.com" {
							return user, nil
						}
						return nil, nil
					},
					FindUserByIDFunc: func(id uuid.UUID) (*domain.User, error) {
						if id == user.ID {
							return user, nil
						}
						return nil, errors.New("user not found")
					},
					UpdateUserFunc: func(user *domain.User) error {
						return nil
					},
				}

				mockTokenRepo := &MockTokenRepository{
					SaveRefreshTokenFunc: func(token *domain.RefreshToken) error {
						return errors.New("token save failed")
					},
				}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "token save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo, mockTokenRepo := tt.setupMocks()

			// 実際のサービスを作成（依存関係はモックを注入）
			userSvc, tokenSvc := createTestServices(mockUserRepo, mockTokenRepo)

			// AuthServiceを作成
			service := NewAuthService(nil, *userSvc, *tokenSvc)

			accessToken, refreshToken, err := service.Login(tt.email, tt.password)

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
	tests := []struct {
		name          string
		refreshToken  string
		setupMocks    func() (*MockUserRepository, *MockTokenRepository)
		expectedError string
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid_refresh_token",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
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

				mockUserRepo := &MockUserRepository{
					FindUserByIDFunc: func(id uuid.UUID) (*domain.User, error) {
						if id == userID {
							return user, nil
						}
						return nil, nil
					},
				}

				mockTokenRepo := &MockTokenRepository{
					FindRefreshTokenFunc: func(token string) (*domain.RefreshToken, error) {
						if token == "valid_refresh_token" {
							return refreshTokenEntity, nil
						}
						return nil, nil
					},
					RevokeRefreshTokenFunc: func(token string) error {
						return nil
					},
					SaveRefreshTokenFunc: func(token *domain.RefreshToken) error {
						return nil
					},
				}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "",
		},
		{
			name:         "refresh token not found",
			refreshToken: "nonexistent_token",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				mockUserRepo := &MockUserRepository{}

				mockTokenRepo := &MockTokenRepository{
					FindRefreshTokenFunc: func(token string) (*domain.RefreshToken, error) {
						return nil, errors.New("token not found")
					},
				}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "token not found",
		},
		{
			name:         "refresh token is revoked",
			refreshToken: "revoked_token",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				revokedTime := time.Now().Add(-time.Hour)
				refreshTokenEntity := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     "revoked_token",
					UserID:    uuid.New(),
					ExpiresAt: time.Now().Add(24 * time.Hour),
					RevokedAt: &revokedTime,
				}

				mockUserRepo := &MockUserRepository{}

				mockTokenRepo := &MockTokenRepository{
					FindRefreshTokenFunc: func(token string) (*domain.RefreshToken, error) {
						if token == "revoked_token" {
							return refreshTokenEntity, nil
						}
						return nil, nil
					},
				}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "invalid refresh token",
		},
		{
			name:         "refresh token is expired",
			refreshToken: "expired_token",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository) {
				refreshTokenEntity := &domain.RefreshToken{
					ID:        uuid.New(),
					Token:     "expired_token",
					UserID:    uuid.New(),
					ExpiresAt: time.Now().Add(-time.Hour), // 期限切れ
					RevokedAt: nil,
				}

				mockUserRepo := &MockUserRepository{}

				mockTokenRepo := &MockTokenRepository{
					FindRefreshTokenFunc: func(token string) (*domain.RefreshToken, error) {
						if token == "expired_token" {
							return refreshTokenEntity, nil
						}
						return nil, nil
					},
				}

				return mockUserRepo, mockTokenRepo
			},
			expectedError: "refresh token expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo, mockTokenRepo := tt.setupMocks()

			// 実際のサービスを作成（依存関係はモックを注入）
			userSvc, tokenSvc := createTestServices(mockUserRepo, mockTokenRepo)

			// AuthServiceを作成
			service := NewAuthService(nil, *userSvc, *tokenSvc)

			newAccessToken, newRefreshToken, err := service.RefreshToken(tt.refreshToken)

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
	tests := []struct {
		name          string
		setupMocks    func() (*MockUserRepository, *MockTokenRepository, string, string)
		expectedError string
	}{
		{
			name: "successful logout",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository, string, string) {
				mockUserRepo := &MockUserRepository{}

				mockTokenRepo := &MockTokenRepository{
					SaveTokenToBlacklistFunc: func(token string, ttl time.Duration) error {
						return nil
					},
					RevokeRefreshTokenFunc: func(token string) error {
						return nil
					},
				}

				// 実際のJWTトークンを生成
				jwtManager := createTestJWTManager()
				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Role:     "user",
				}

				claims := &token.Claims{
					UserID:   user.ID.String(),
					Email:    user.Email,
					Username: user.Username,
					Role:     user.Role,
				}

				accessToken, err := jwtManager.Generate(claims, 1*time.Hour)
				assert.NoError(t, err)

				refreshToken := "valid_refresh_token"

				return mockUserRepo, mockTokenRepo, accessToken, refreshToken
			},
			expectedError: "",
		},
		{
			name: "blacklist token failure",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository, string, string) {
				mockUserRepo := &MockUserRepository{}

				mockTokenRepo := &MockTokenRepository{
					SaveTokenToBlacklistFunc: func(token string, ttl time.Duration) error {
						return errors.New("blacklist save failed")
					},
				}

				// 実際のJWTトークンを生成
				jwtManager := createTestJWTManager()
				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Role:     "user",
				}

				claims := &token.Claims{
					UserID:   user.ID.String(),
					Email:    user.Email,
					Username: user.Username,
					Role:     user.Role,
				}

				accessToken, err := jwtManager.Generate(claims, 1*time.Hour)
				assert.NoError(t, err)

				refreshToken := "valid_refresh_token"

				return mockUserRepo, mockTokenRepo, accessToken, refreshToken
			},
			expectedError: "blacklist save failed",
		},
		{
			name: "refresh token revocation failure",
			setupMocks: func() (*MockUserRepository, *MockTokenRepository, string, string) {
				mockUserRepo := &MockUserRepository{}

				mockTokenRepo := &MockTokenRepository{
					SaveTokenToBlacklistFunc: func(token string, ttl time.Duration) error {
						return nil
					},
					RevokeRefreshTokenFunc: func(token string) error {
						return errors.New("refresh token revocation failed")
					},
				}

				// 実際のJWTトークンを生成
				jwtManager := createTestJWTManager()
				user := &domain.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
					Role:     "user",
				}

				claims := &token.Claims{
					UserID:   user.ID.String(),
					Email:    user.Email,
					Username: user.Username,
					Role:     user.Role,
				}

				accessToken, err := jwtManager.Generate(claims, 1*time.Hour)
				assert.NoError(t, err)

				refreshToken := "valid_refresh_token"

				return mockUserRepo, mockTokenRepo, accessToken, refreshToken
			},
			expectedError: "refresh token revocation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo, mockTokenRepo, accessToken, refreshToken := tt.setupMocks()

			// 実際のサービスを作成（依存関係はモックを注入）
			userSvc, tokenSvc := createTestServices(mockUserRepo, mockTokenRepo)

			// AuthServiceを作成
			service := NewAuthService(nil, *userSvc, *tokenSvc)

			err := service.Logout(accessToken, refreshToken)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 統合テスト：フル認証フロー
func TestAuthService_FullAuthFlow(t *testing.T) {
	email := "integration@example.com"
	username := "integrationuser"
	password := "password123"

	// 状態を保持するための変数
	var storedUser *domain.User
	var accessToken, refreshToken string

	// フロー全体を通して使用するモックを作成
	mockUserRepo := &MockUserRepository{}
	mockTokenRepo := &MockTokenRepository{}

	// 実際のサービスを作成
	userSvc, tokenSvc := createTestServices(mockUserRepo, mockTokenRepo)
	service := NewAuthService(nil, *userSvc, *tokenSvc)

	// ステップ1: ユーザー登録
	t.Run("register user", func(t *testing.T) {
		// 登録時の振る舞いを設定
		mockUserRepo.FindUserByEmailFunc = func(email string) (*domain.User, error) {
			if storedUser != nil && storedUser.Email == email {
				return storedUser, nil
			}
			return nil, nil
		}

		mockUserRepo.CreateUserFunc = func(user *domain.User) error {
			storedUser = user
			return nil
		}

		user, err := service.Register(email, username, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, username, user.Username)

		storedUser = user
	})

	// ステップ2: ログイン
	t.Run("login user", func(t *testing.T) {
		// ログイン時の振る舞いを設定
		mockUserRepo.FindUserByEmailFunc = func(email string) (*domain.User, error) {
			if storedUser != nil && storedUser.Email == email {
				return storedUser, nil
			}
			return nil, nil
		}

		mockUserRepo.FindUserByIDFunc = func(id uuid.UUID) (*domain.User, error) {
			if storedUser != nil && storedUser.ID == id {
				return storedUser, nil
			}
			return nil, errors.New("user not found")
		}

		mockUserRepo.UpdateUserFunc = func(user *domain.User) error {
			storedUser = user
			return nil
		}

		mockTokenRepo.SaveRefreshTokenFunc = func(token *domain.RefreshToken) error {
			return nil
		}

		at, rt, err := service.Login(email, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, at)
		assert.NotEmpty(t, rt)

		accessToken = at
		refreshToken = rt
	})

	// ステップ3: ログアウト
	t.Run("logout user", func(t *testing.T) {
		mockTokenRepo.SaveTokenToBlacklistFunc = func(token string, ttl time.Duration) error {
			return nil
		}

		mockTokenRepo.RevokeRefreshTokenFunc = func(token string) error {
			return nil
		}

		err := service.Logout(accessToken, refreshToken)
		assert.NoError(t, err)
	})
}
