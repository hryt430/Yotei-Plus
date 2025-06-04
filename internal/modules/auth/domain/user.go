package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User roles
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID            uuid.UUID      `json:"id"`
	Email         string         `json:"email"`
	Username      string         `json:"username"`
	Password      string         `json:"-"` // パスワードはJSONに含めない
	Role          string         `json:"role"`
	EmailVerified bool           `json:"email_verified"`
	LastLogin     *time.Time     `json:"last_login"`
	RefreshTokens []RefreshToken `json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// NewUser は新しいUserを作成する
func NewUser(email, username, password string) *User {
	now := time.Now()
	return &User{
		ID:            uuid.New(),
		Email:         email,
		Username:      username,
		Password:      password,
		Role:          RoleUser,
		EmailVerified: false,
		LastLogin:     nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// SetRole はユーザーの役割を設定する
func (u *User) SetRole(role string) error {
	if role != RoleUser && role != RoleAdmin {
		return fmt.Errorf("invalid role: %s", role)
	}
	u.Role = role
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateLastLogin は最終ログイン時刻を更新する
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
	u.UpdatedAt = now
}

// IsAdmin はユーザーが管理者かどうかを返す
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	Token     string     `json:"-"`
	UserID    uuid.UUID  `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	IssuedAt  time.Time  `json:"issued_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NewRefreshToken は新しいRefreshTokenを作成する
func NewRefreshToken(userID uuid.UUID, token string, expirationDuration time.Duration) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		ID:        uuid.New(),
		Token:     token,
		UserID:    userID,
		ExpiresAt: now.Add(expirationDuration),
		IssuedAt:  now,
		RevokedAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsExpired はトークンが期限切れかどうかを判定する
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked はトークンが無効化されているかどうかを判定する
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// Revoke はトークンを無効化する
func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.RevokedAt = &now
	rt.UpdatedAt = now
}
