package database

import (
	"database/sql"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
)

// TokenStorage はデータベースを使用したトークンストレージの実装
type TokenStorage struct {
	SqlHandler
}

func (t *TokenStorage) SaveRefreshToken(token *domain.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, token, user_id, expires_at, issued_at, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := t.Execute(query, token.ID, token.Token, token.UserID,
		token.ExpiresAt, token.IssuedAt, token.CreatedAt, token.UpdatedAt)
	return err
}

func (t *TokenStorage) FindRefreshTokenByToken(token string) (*domain.RefreshToken, error) {
	query := `SELECT id, token, user_id, expires_at, issued_at, revoked_at, created_at, updated_at 
			  FROM refresh_tokens WHERE token = ? AND revoked_at IS NULL`

	var refreshToken domain.RefreshToken
	var revokedAt sql.NullTime

	row, err := t.Query(query, token)

	if err = row.Scan(
		&refreshToken.ID, &refreshToken.Token, &refreshToken.UserID,
		&refreshToken.ExpiresAt, &refreshToken.IssuedAt, &revokedAt,
		&refreshToken.CreatedAt, &refreshToken.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if revokedAt.Valid {
		refreshToken.RevokedAt = &revokedAt.Time
	}

	return &refreshToken, nil
}

func (t *TokenStorage) RevokeRefreshToken(token string) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = ?`
	_, err := t.Execute(query, token)
	return err
}

func (r *TokenStorage) DeleteExpiredRefreshTokens() error {
	// 期限切れのリフレッシュトークンを削除（定期的なクリーンアップ用）
	query := "DELETE FROM refresh_tokens WHERE expires_at < ?"
	_, err := r.Execute(query, time.Now())
	return err
}
