package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		username    string
		password    string
		wantAsserts func(*User)
	}{
		{
			name:     "valid user creation",
			email:    "test@example.com",
			username: "testuser",
			password: "hashedpassword123",
			wantAsserts: func(user *User) {
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "testuser", user.Username)
				assert.Equal(t, "hashedpassword123", user.Password)
				assert.Equal(t, RoleUser, user.Role)
				assert.False(t, user.EmailVerified)
				assert.Nil(t, user.LastLogin)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.False(t, user.CreatedAt.IsZero())
				assert.False(t, user.UpdatedAt.IsZero())
				assert.Equal(t, user.CreatedAt, user.UpdatedAt)
			},
		},
		{
			name:     "user with empty fields",
			email:    "",
			username: "",
			password: "",
			wantAsserts: func(user *User) {
				assert.Equal(t, "", user.Email)
				assert.Equal(t, "", user.Username)
				assert.Equal(t, "", user.Password)
				assert.Equal(t, RoleUser, user.Role)
				assert.False(t, user.EmailVerified)
				assert.Nil(t, user.LastLogin)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.email, tt.username, tt.password)
			require.NotNil(t, user)
			tt.wantAsserts(user)
		})
	}
}

func TestUser_SetRole(t *testing.T) {
	user := NewUser("test@example.com", "testuser", "password")
	originalUpdatedAt := user.UpdatedAt

	tests := []struct {
		name          string
		role          string
		expectedError bool
		expectedRole  string
	}{
		{
			name:          "set admin role",
			role:          RoleAdmin,
			expectedError: false,
			expectedRole:  RoleAdmin,
		},
		{
			name:          "set user role",
			role:          RoleUser,
			expectedError: false,
			expectedRole:  RoleUser,
		},
		{
			name:          "invalid role",
			role:          "invalid_role",
			expectedError: true,
			expectedRole:  RoleUser, // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sleep to ensure UpdatedAt changes
			time.Sleep(1 * time.Millisecond)

			err := user.SetRole(tt.role)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid role")
				assert.Equal(t, originalUpdatedAt, user.UpdatedAt) // Should not be updated on error
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRole, user.Role)
				assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
				originalUpdatedAt = user.UpdatedAt // Update for next test
			}
		})
	}
}

func TestUser_UpdateLastLogin(t *testing.T) {
	user := NewUser("test@example.com", "testuser", "password")
	originalUpdatedAt := user.UpdatedAt

	// Initially LastLogin should be nil
	assert.Nil(t, user.LastLogin)

	// Sleep to ensure timestamps are different
	time.Sleep(1 * time.Millisecond)

	user.UpdateLastLogin()

	// LastLogin should be set to a recent time
	require.NotNil(t, user.LastLogin)
	assert.True(t, user.LastLogin.After(originalUpdatedAt))

	// UpdatedAt should be updated
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// LastLogin and UpdatedAt should be very close in time
	timeDiff := user.UpdatedAt.Sub(*user.LastLogin)
	assert.True(t, timeDiff < time.Millisecond)
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "admin user",
			role:     RoleAdmin,
			expected: true,
		},
		{
			name:     "regular user",
			role:     RoleUser,
			expected: false,
		},
		{
			name:     "invalid role",
			role:     "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser("test@example.com", "testuser", "password")
			user.Role = tt.role

			result := user.IsAdmin()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewRefreshToken(t *testing.T) {
	userID := uuid.New()
	token := "refresh_token_string"
	duration := 7 * 24 * time.Hour

	refreshToken := NewRefreshToken(userID, token, duration)

	require.NotNil(t, refreshToken)
	assert.NotEqual(t, uuid.Nil, refreshToken.ID)
	assert.Equal(t, token, refreshToken.Token)
	assert.Equal(t, userID, refreshToken.UserID)
	assert.False(t, refreshToken.IssuedAt.IsZero())
	assert.False(t, refreshToken.CreatedAt.IsZero())
	assert.False(t, refreshToken.UpdatedAt.IsZero())
	assert.Nil(t, refreshToken.RevokedAt)

	// ExpiresAt should be roughly duration from now
	expectedExpiration := time.Now().Add(duration)
	timeDiff := refreshToken.ExpiresAt.Sub(expectedExpiration)
	assert.True(t, timeDiff < time.Second && timeDiff > -time.Second, "ExpiresAt should be close to expected time")
}

func TestRefreshToken_IsExpired(t *testing.T) {
	userID := uuid.New()
	token := "refresh_token_string"

	tests := []struct {
		name     string
		duration time.Duration
		expected bool
	}{
		{
			name:     "not expired - future expiration",
			duration: time.Hour,
			expected: false,
		},
		{
			name:     "expired - past expiration",
			duration: -time.Hour,
			expected: true,
		},
		{
			name:     "just expired",
			duration: -time.Millisecond,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := NewRefreshToken(userID, token, tt.duration)
			result := refreshToken.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	userID := uuid.New()
	token := "refresh_token_string"
	duration := time.Hour

	refreshToken := NewRefreshToken(userID, token, duration)

	// Initially should not be revoked
	assert.False(t, refreshToken.IsRevoked())

	// After setting RevokedAt, should be revoked
	now := time.Now()
	refreshToken.RevokedAt = &now
	assert.True(t, refreshToken.IsRevoked())
}

func TestRefreshToken_Revoke(t *testing.T) {
	userID := uuid.New()
	token := "refresh_token_string"
	duration := time.Hour

	refreshToken := NewRefreshToken(userID, token, duration)
	originalUpdatedAt := refreshToken.UpdatedAt

	// Initially should not be revoked
	assert.False(t, refreshToken.IsRevoked())
	assert.Nil(t, refreshToken.RevokedAt)

	// Sleep to ensure timestamps are different
	time.Sleep(1 * time.Millisecond)

	refreshToken.Revoke()

	// Should now be revoked
	assert.True(t, refreshToken.IsRevoked())
	require.NotNil(t, refreshToken.RevokedAt)

	// RevokedAt should be recent
	timeSinceRevoked := time.Since(*refreshToken.RevokedAt)
	assert.True(t, timeSinceRevoked < time.Second)

	// UpdatedAt should be updated
	assert.True(t, refreshToken.UpdatedAt.After(originalUpdatedAt))

	// RevokedAt and UpdatedAt should be very close in time
	timeDiff := refreshToken.UpdatedAt.Sub(*refreshToken.RevokedAt)
	assert.True(t, timeDiff < time.Millisecond)
}

func TestRefreshToken_CompleteLifecycle(t *testing.T) {
	userID := uuid.New()
	token := "refresh_token_string"
	duration := time.Hour

	// Create new token
	refreshToken := NewRefreshToken(userID, token, duration)

	// Verify initial state
	assert.False(t, refreshToken.IsExpired())
	assert.False(t, refreshToken.IsRevoked())

	// Revoke the token
	refreshToken.Revoke()

	// Verify revoked state
	assert.True(t, refreshToken.IsRevoked())
	assert.False(t, refreshToken.IsExpired()) // Still not expired, just revoked

	// Create expired token
	expiredToken := NewRefreshToken(userID, "expired_token", -time.Hour)

	// Verify expired state
	assert.True(t, expiredToken.IsExpired())
	assert.False(t, expiredToken.IsRevoked())

	// Revoke expired token
	expiredToken.Revoke()

	// Verify both expired and revoked
	assert.True(t, expiredToken.IsExpired())
	assert.True(t, expiredToken.IsRevoked())
}

// Test edge cases and validation
func TestUser_RoleConstants(t *testing.T) {
	assert.Equal(t, "user", RoleUser)
	assert.Equal(t, "admin", RoleAdmin)
}

func TestRefreshToken_ZeroDuration(t *testing.T) {
	userID := uuid.New()
	token := "test_token"

	refreshToken := NewRefreshToken(userID, token, 0)

	// Should be immediately expired
	assert.True(t, refreshToken.IsExpired())
}

func TestUser_ConcurrentLastLoginUpdate(t *testing.T) {
	user := NewUser("test@example.com", "testuser", "password")

	// Simulate multiple rapid login updates
	var lastLogins []*time.Time

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Millisecond)
		user.UpdateLastLogin()
		lastLogins = append(lastLogins, user.LastLogin)
	}

	// Each update should result in a later timestamp
	for i := 1; i < len(lastLogins); i++ {
		assert.True(t, lastLogins[i].After(*lastLogins[i-1]))
	}
}
