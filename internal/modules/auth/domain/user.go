package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	Email         string         `json:"email" gorm:"unique;not null"`
	Username      string         `json:"username" gorm:"unique;not null"`
	Password      string         `json:"-" gorm:"not null"` // パスワードはJSONに含めない
	Role          string         `json:"role" gorm:"default:user"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	LastLogin     time.Time      `json:"last_login"`
	RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	Token     string     `json:"-" gorm:"unique;not null"`
	UserID    uuid.UUID  `json:"-" gorm:"type:uuid;not null"`
	ExpiresAt time.Time  `json:"expires_at"`
	IssuedAt  time.Time  `json:"issued_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
