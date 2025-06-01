package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
)

type IUserRepository struct {
	SqlHandler
}

// UserExists はユーザーが存在するかチェック（軽量版）
func (r *IUserRepository) UserExists(userID string) (bool, error) {
	query := `SELECT 1 FROM ` + "`Yotei-Plus`" + `.users WHERE id = ? LIMIT 1`

	row, err := r.Query(query, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			// ログ出力（実際の実装ではloggerを使用）
			fmt.Printf("Warning: failed to close row: %v\n", closeErr)
		}
	}()

	return row.Next(), nil
}

// GetUserBasicInfo はユーザーの基本情報のみ取得（軽量版）
func (r *IUserRepository) GetUserBasicInfo(userID string) (*UserBasicInfo, error) {
	query := `SELECT id, username, email FROM ` + "`Yotei-Plus`" + `.users WHERE id = ? LIMIT 1`

	row, err := r.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user basic info: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close row: %v\n", closeErr)
		}
	}()

	if !row.Next() {
		return nil, nil // ユーザーが見つからない
	}

	var info UserBasicInfo
	if err := row.Scan(&info.ID, &info.Username, &info.Email); err != nil {
		return nil, fmt.Errorf("failed to scan user basic info: %w", err)
	}

	return &info, nil
}

// GetUsersBasicInfoBatch は複数ユーザーの基本情報を一括取得（N+1問題解決）
func (r *IUserRepository) GetUsersBasicInfoBatch(userIDs []string) (map[string]*UserBasicInfo, error) {
	if len(userIDs) == 0 {
		return make(map[string]*UserBasicInfo), nil
	}

	// プレースホルダーを動的に生成
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id, username, email FROM ` + "`Yotei-Plus`" + `.users 
		WHERE id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := r.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users basic info batch: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	result := make(map[string]*UserBasicInfo)
	for rows.Next() {
		var info UserBasicInfo
		if err := rows.Scan(&info.ID, &info.Username, &info.Email); err != nil {
			return nil, fmt.Errorf("failed to scan user basic info: %w", err)
		}
		result[info.ID] = &info
	}

	return result, nil
}

// UserBasicInfo はユーザーの基本情報
type UserBasicInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// CreateUser は新しいユーザーを作成する（コネクション管理改善）
func (r *IUserRepository) CreateUser(user *domain.User) error {
	// Role のバリデーションとデフォルト値設定
	if user.Role == "" {
		user.Role = domain.RoleUser
	}
	if user.Role != domain.RoleUser && user.Role != domain.RoleAdmin {
		return fmt.Errorf("invalid role: %s", user.Role)
	}

	// 時刻フィールドのバリデーションと修正
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

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
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindUserByEmail はメールアドレスでユーザーを検索する（コネクション管理改善）
func (r *IUserRepository) FindUserByEmail(email string) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.users 
		WHERE email = ? LIMIT 1`

	row, err := r.Query(query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close row: %v\n", closeErr)
		}
	}()

	if !row.Next() {
		return nil, nil // ユーザーが見つからない
	}

	return r.scanUser(row)
}

// FindUserByID はIDでユーザーを検索する（コネクション管理改善）
func (r *IUserRepository) FindUserByID(id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.users 
		WHERE id = ? LIMIT 1`

	row, err := r.Query(query, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query user by ID: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close row: %v\n", closeErr)
		}
	}()

	if !row.Next() {
		return nil, nil // ユーザーが見つからない
	}

	return r.scanUser(row)
}

// FindUserByUsername はユーザー名による検索（コネクション管理改善）
func (r *IUserRepository) FindUserByUsername(username string) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, email_verified, last_login, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.users 
		WHERE username = ? LIMIT 1`

	row, err := r.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by username: %w", err)
	}
	defer func() {
		if closeErr := row.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close row: %v\n", closeErr)
		}
	}()

	if !row.Next() {
		return nil, nil // ユーザーが見つからない
	}

	return r.scanUser(row)
}

// FindUsers はユーザー一覧取得（検索機能付き、コネクション管理改善）
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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var users []*domain.User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser はユーザーを更新する（コネクション管理改善）
func (r *IUserRepository) UpdateUser(user *domain.User) error {
	// Role のバリデーション
	if user.Role != domain.RoleUser && user.Role != domain.RoleAdmin {
		return fmt.Errorf("invalid role: %s", user.Role)
	}

	// UpdatedAt を現在時刻に設定
	user.UpdatedAt = time.Now()

	query := `UPDATE ` + "`Yotei-Plus`" + `.users 
		SET username = ?, email = ?, password = ?, role = ?, email_verified = ?, last_login = ?, updated_at = ? 
		WHERE id = ?`

	result, err := r.Execute(query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.EmailVerified,
		user.LastLogin,
		user.UpdatedAt,
		user.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID.String())
	}

	return nil
}

// scanUser は共通のスキャン処理（重複コード削減）
func (r *IUserRepository) scanUser(row Row) (*domain.User, error) {
	var user domain.User
	var idStr string
	var lastLogin sql.NullTime

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
		return nil, fmt.Errorf("failed to scan user fields: %w", err)
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

	return &user, nil
}

// UserValidator の実装
type UserValidator struct {
	userRepo *IUserRepository
}

func NewUserValidator(userRepo *IUserRepository) *UserValidator {
	return &UserValidator{userRepo: userRepo}
}

func (v *UserValidator) UserExists(userID string) (bool, error) {
	return v.userRepo.UserExists(userID)
}

func (v *UserValidator) GetUserBasicInfo(userID string) (*UserBasicInfo, error) {
	return v.userRepo.GetUserBasicInfo(userID)
}
