package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/internal/modules/auth/domain"
)

// TokenStorage はデータベースを使用したトークンストレージの実装
type TokenStorage struct {
	SqlHandler
}

func (t *TokenStorage) SaveRefreshToken(token *domain.RefreshToken) error {
	query := `INSERT INTO ` + "`Yotei-Plus`" + `.refresh_tokens 
		(id, token, user_id, expires_at, issued_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := t.Execute(query,
		token.ID.String(),
		token.Token,
		token.UserID.String(),
		token.ExpiresAt,
		token.IssuedAt,
		token.CreatedAt,
		token.UpdatedAt,
	)
	return err
}

func (t *TokenStorage) FindRefreshTokenByToken(token string) (*domain.RefreshToken, error) {
	query := `SELECT id, token, user_id, expires_at, issued_at, revoked_at, created_at, updated_at 
		FROM ` + "`Yotei-Plus`" + `.refresh_tokens 
		WHERE token = ? AND revoked_at IS NULL`

	row, err := t.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, nil // トークンが見つからない
	}

	var refreshToken domain.RefreshToken
	var revokedAt sql.NullTime
	var idStr, userIDStr string

	if err = row.Scan(
		&idStr,
		&refreshToken.Token,
		&userIDStr,
		&refreshToken.ExpiresAt,
		&refreshToken.IssuedAt,
		&revokedAt,
		&refreshToken.CreatedAt,
		&refreshToken.UpdatedAt,
	); err != nil {
		return nil, err
	}

	//　UUIDパース
	if parsedID, err := uuid.Parse(idStr); err == nil {
		refreshToken.ID = parsedID
	} else {
		return nil, err
	}

	if parsedUserID, err := uuid.Parse(userIDStr); err == nil {
		refreshToken.UserID = parsedUserID
	} else {
		return nil, err
	}

	if revokedAt.Valid {
		refreshToken.RevokedAt = &revokedAt.Time
	}

	return &refreshToken, nil
}

func (t *TokenStorage) RevokeRefreshToken(token string) error {
	query := `UPDATE ` + "`Yotei-Plus`" + `.refresh_tokens 
		SET revoked_at = NOW() 
		WHERE token = ?`
	_, err := t.Execute(query, token)
	return err
}

func (r *TokenStorage) DeleteExpiredRefreshTokens() error {
	query := `DELETE FROM ` + "`Yotei-Plus`" + `.refresh_tokens 
		WHERE expires_at < ?`
	_, err := r.Execute(query, time.Now())
	return err
}
