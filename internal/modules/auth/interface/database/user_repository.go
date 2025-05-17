package database

import (
	"context"
	"time"

	"auth-service/internal/domain/entity"

	"github.com/google/uuid"
)

type UserServiceRepository struct {
	SqlHandler
}

func (r *UserServiceRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := "INSERT INTO users (id, name, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := r.Execute(query, user.ID, user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *UserServiceRepository) FindUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := "SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ? LIMIT 1"
	row, err := r.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var user entity.User
	if !row.Next() {
		return nil, nil // NotFound扱い
	}
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserServiceRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := "SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = ? LIMIT 1"
	row, err := r.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var user entity.User
	if !row.Next() {
		return nil, nil
	}
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserServiceRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	query := "UPDATE users SET name = ?, email = ?, password = ?, updated_at = ? WHERE id = ?"
	_, err := r.Execute(query, user.Username, user.Email, user.Password, time.Now(), user.ID)
	return err
}

func (r *UserServiceRepository) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	query := "INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)"
	_, err := r.Execute(query, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *UserServiceRepository) FindRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	query := "SELECT id, user_id, token, expires_at, revoked_at, created_at FROM refresh_tokens WHERE token = ? LIMIT 1"
	row, err := r.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var rt entity.RefreshToken
	if !row.Next() {
		return nil, nil
	}
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *UserServiceRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	now := time.Now()
	query := "UPDATE refresh_tokens SET revoked_at = ? WHERE token = ?"
	_, err := r.Execute(query, now, token)
	return err
}

func (r *UserServiceRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	// 期限切れのリフレッシュトークンを削除（定期的なクリーンアップ用）
	query := "DELETE FROM refresh_tokens WHERE expires_at < ?"
	_, err := r.Execute(query, time.Now())
	return err
}
