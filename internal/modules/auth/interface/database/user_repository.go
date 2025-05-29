package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"

	"github.com/google/uuid"
)

type IUserRepository struct {
	SqlHandler
}

func (r *IUserRepository) CreateUser(user *domain.User) error {
	// ✅ Role のバリデーションとデフォルト値設定
	if user.Role == "" {
		user.Role = domain.RoleUser // デフォルト値
	}
	if user.Role != domain.RoleUser && user.Role != domain.RoleAdmin {
		return fmt.Errorf("invalid role: %s", user.Role)
	}

	// ✅ 時刻フィールドのバリデーションと修正
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	// ✅ デバッグ用ログ（開発時のみ）
	fmt.Printf("Creating user - CreatedAt: %v, UpdatedAt: %v, Role: %s\n",
		user.CreatedAt, user.UpdatedAt, user.Role)

	// ✅ 修正: name → username、テーブル名にDB名追加、全フィールド対応
	query := `INSERT INTO ` + "`Yotei-Plus`" + `.users 
		(id, username, email, password, role, email_verified, last_login, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.Execute(query,
		user.ID.String(),
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.EmailVerified,
		user.LastLogin,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *IUserRepository) FindUserByEmail(email string) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.users 
		WHERE email = ? LIMIT 1`
	row, err := r.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var user domain.User
	var idStr string
	var lastLogin sql.NullTime

	if !row.Next() {
		return nil, nil
	}

	if err := row.Scan(
		&idStr,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.EmailVerified,
		&lastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	user.ID = parsedID

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

func (r *IUserRepository) FindUserByID(id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.users 
		WHERE id = ? LIMIT 1`
	row, err := r.Query(query, id.String())
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var user domain.User
	var idStr string
	var lastLogin sql.NullTime

	if !row.Next() {
		return nil, nil
	}

	if err := row.Scan(
		&idStr,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.EmailVerified,
		&lastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	user.ID = parsedID

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

// ✅ 新規追加: ユーザー名による検索
func (r *IUserRepository) FindUserByUsername(username string) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.users 
		WHERE username = ? LIMIT 1`
	row, err := r.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var user domain.User
	var idStr string
	var lastLogin sql.NullTime

	if !row.Next() {
		return nil, nil
	}

	if err := row.Scan(
		&idStr,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.EmailVerified,
		&lastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	user.ID = parsedID

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

// ✅ 新規追加: ユーザー一覧取得（検索機能付き）
func (r *IUserRepository) FindUsers(search string) ([]*domain.User, error) {
	var query string
	var args []interface{}

	if search != "" {
		search = strings.TrimSpace(search)
		searchPattern := "%" + search + "%"
		query = `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
			FROM ` + "`Yotei-Plus`" + `.users 
			WHERE username LIKE ? OR email LIKE ? 
			ORDER BY username ASC 
			LIMIT 100`
		args = []interface{}{searchPattern, searchPattern}
	} else {
		query = `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
			FROM ` + "`Yotei-Plus`" + `.users 
			ORDER BY username ASC 
			LIMIT 100`
		args = []interface{}{}
	}

	rows, err := r.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var idStr string
		var lastLogin sql.NullTime

		if err := rows.Scan(
			&idStr,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.EmailVerified,
			&lastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// UUIDパース
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user ID: %w", err)
		}
		user.ID = parsedID

		// last_loginのNULL処理
		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, &user)
	}

	return users, nil
}

func (r *IUserRepository) UpdateUser(user *domain.User) error {
	// ✅ Role のバリデーション
	if user.Role != domain.RoleUser && user.Role != domain.RoleAdmin {
		return fmt.Errorf("invalid role: %s", user.Role)
	}

	// ✅ UpdatedAt を現在時刻に設定
	user.UpdatedAt = time.Now()

	query := `UPDATE ` + "`Yotei-Plus`" + `.users 
		SET username = ?, email = ?, password = ?, role = ?, email_verified = ?, last_login = ?, updated_at = ? 
		WHERE id = ?`
	_, err := r.Execute(query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.EmailVerified,
		user.LastLogin,
		user.UpdatedAt,
		user.ID.String(),
	)
	return err
}
