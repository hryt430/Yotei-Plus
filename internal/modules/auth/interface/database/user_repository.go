package database

import (
	"database/sql"
	"fmt"
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
		user.Username, // ✅ name → username
		user.Email,
		user.Password,
		user.Role,          // ✅ 追加
		user.EmailVerified, // ✅ 追加
		user.LastLogin,     // ✅ 修正: &を削除（すでにポインタ型）
		user.CreatedAt,     // ✅ 検証済み時刻
		user.UpdatedAt,     // ✅ 検証済み時刻
	)
	return err
}

func (r *IUserRepository) FindUserByEmail(email string) (*domain.User, error) {
	// ✅ 修正: name → username、テーブル名にDB名追加、全フィールド対応
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
	var lastLogin sql.NullTime // NULL許可

	if !row.Next() {
		return nil, nil // NotFound扱い
	}

	// ✅ 修正: 全フィールドをスキャン
	if err := row.Scan(
		&idStr,
		&user.Username, // ✅ name → username
		&user.Email,
		&user.Password,
		&user.Role,          // ✅ 追加
		&user.EmailVerified, // ✅ 追加
		&lastLogin,          // ✅ 追加
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	// UUIDパース
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	user.ID = parsedID

	// last_loginのNULL処理
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time // ✅ アドレス演算子を追加
	}

	return &user, nil
}

func (r *IUserRepository) FindUserByID(id uuid.UUID) (*domain.User, error) {
	// ✅ 修正: name → username、テーブル名にDB名追加、全フィールド対応
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

	// ✅ 修正: 全フィールドをスキャン
	if err := row.Scan(
		&idStr,
		&user.Username, // ✅ name → username
		&user.Email,
		&user.Password,
		&user.Role,          // ✅ 追加
		&user.EmailVerified, // ✅ 追加
		&lastLogin,          // ✅ 追加
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	// UUIDパース
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	user.ID = parsedID

	// last_loginのNULL処理
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time // ✅ アドレス演算子を追加
	}

	return &user, nil
}

func (r *IUserRepository) UpdateUser(user *domain.User) error {
	// ✅ Role のバリデーション
	if user.Role != domain.RoleUser && user.Role != domain.RoleAdmin {
		return fmt.Errorf("invalid role: %s", user.Role)
	}

	// ✅ UpdatedAt を現在時刻に設定
	user.UpdatedAt = time.Now()

	// ✅ 修正: name → username、テーブル名にDB名追加、全フィールド対応
	query := `UPDATE ` + "`Yotei-Plus`" + `.users 
		SET username = ?, email = ?, password = ?, role = ?, email_verified = ?, last_login = ?, updated_at = ? 
		WHERE id = ?`
	_, err := r.Execute(query,
		user.Username, // ✅ name → username
		user.Email,
		user.Password,
		user.Role,          // ✅ 追加
		user.EmailVerified, // ✅ 追加
		user.LastLogin,     // ✅ 修正: &を削除（すでにポインタ型）
		user.UpdatedAt,     // ✅ 現在時刻を使用
		user.ID.String(),
	)
	return err
}
