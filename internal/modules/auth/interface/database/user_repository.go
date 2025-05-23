package database

import (
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"

	"github.com/google/uuid"
)

type IUserRepository struct {
	SqlHandler
}

func (r *IUserRepository) CreateUser(user *domain.User) error {
	query := "INSERT INTO users (id, name, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := r.Execute(query, user.ID, user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *IUserRepository) FindUserByEmail(email string) (*domain.User, error) {
	query := "SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ? LIMIT 1"
	row, err := r.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var user domain.User
	if !row.Next() {
		return nil, nil // NotFound扱い
	}
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *IUserRepository) FindUserByID(id uuid.UUID) (*domain.User, error) {
	query := "SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = ? LIMIT 1"
	row, err := r.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var user domain.User
	if !row.Next() {
		return nil, nil
	}
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *IUserRepository) UpdateUser(user *domain.User) error {
	query := "UPDATE users SET name = ?, email = ?, password = ?, updated_at = ? WHERE id = ?"
	_, err := r.Execute(query, user.Username, user.Email, user.Password, time.Now(), user.ID)
	return err
}
