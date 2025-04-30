package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"your-app/notification/domain/entity"
	"your-app/notification/domain/repository"
)

type mysqlNotificationRepository struct {
	db *sql.DB
}

// NewMySQLNotificationRepository は通知リポジトリのMySQL実装を返します
func NewMySQLNotificationRepository(db *sql.DB) repository.NotificationRepository {
	return &mysqlNotificationRepository{
		db: db,
	}
}

// Store は通知をデータベースに保存します
func (r *mysqlNotificationRepository) Store(ctx context.Context, notification *entity.Notification) error {
	query := `
		INSERT INTO notifications 
		(user_id, title, content, type, status, metadata, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	// metadataをJSON形式に変換
	var metadataJSON []byte
	if notification.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(notification.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		notification.UserID,
		notification.Title,
		notification.Content,
		notification.Type,
		notification.Status,
		metadataJSON,
		notification.CreatedAt,
		notification.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	notification.ID = id

	return nil
}

// GetByID は指定されたIDの通知を取得します
func (r *mysqlNotificationRepository) GetByID(ctx context.Context, id int64) (*entity.Notification, error) {
	query := `
		SELECT id, user_id, title, content, type, status, metadata, created_at, updated_at, read_at
		FROM notifications
		WHERE id = ?
	`

	var notification entity.Notification
	var metadataJSON sql.NullString
	var readAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Content,
		&notification.Type,
		&notification.Status,
		&metadataJSON,
		&notification.CreatedAt,
		&notification.UpdatedAt,
		&readAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("notification not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// JSONをmap[string]stringに変換
	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &notification.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}

	return &notification, nil
}

// GetByUserID はユーザーIDに基づいて通知一覧を取得します
func (r *mysqlNotificationRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, title, content, type, status, metadata, created_at, updated_at, read_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*entity.Notification
	for rows.Next() {
		var notification entity.Notification
		var metadataJSON sql.NullString
		var readAt sql.NullTime

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Content,
			&notification.Type,
			&notification.Status,
			&metadataJSON,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&readAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// JSONをmap[string]stringに変換
		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}

		notifications = append(notifications, &notification)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return notifications, nil
}

// GetUnreadByUserID はユーザーの未読通知を取得します
func (r *mysqlNotificationRepository) GetUnreadByUserID(ctx context.Context, userID int64) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, title, content, type, status, metadata, created_at, updated_at, read_at
		FROM notifications
		WHERE user_id = ? AND status = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, entity.NotificationStatusUnread)
	if err != nil {
		return nil, fmt.Errorf("failed to query unread notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*entity.Notification
	for rows.Next() {
		var notification entity.Notification
		var metadataJSON sql.NullString
		var readAt sql.NullTime

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Content,
			&notification.Type,
			&notification.Status,
			&metadataJSON,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&readAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// JSONをmap[string]stringに変換
		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}

		notifications = append(notifications, &notification)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return notifications, nil
}

// UpdateStatus は通知のステータスを更新します
func (r *mysqlNotificationRepository) UpdateStatus(ctx context.Context, id int64, status entity.NotificationStatus) error {
	query := `
		UPDATE notifications
		SET status = ?, updated_at = ?, read_at = ?
		WHERE id = ?
	`

	now := time.Now()
	var readAt *time.Time
	if status == entity.NotificationStatusRead {
		readAt = &now
	}

	_, err := r.db.ExecContext(ctx, query, status, now, readAt, id)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

// GetUserLineID はユーザーIDに紐づくLINE IDを取得します
func (r *mysqlNotificationRepository) GetUserLineID(ctx context.Context, userID int64) (string, error) {
	query := `
		SELECT line_user_id 
		FROM user_line_mappings 
		WHERE user_id = ?
	`

	var lineUserID string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&lineUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("line user ID not found for user %d", userID)
		}
		return "", fmt.Errorf("failed to get LINE user ID: %w", err)
	}

	return lineUserID, nil
}

// SaveUserLineID はユーザーIDとLINE IDのマッピングを保存します
func (r *mysqlNotificationRepository) SaveUserLineID(ctx context.Context, userID int64, lineUserID string) error {
	query := `
		INSERT INTO user_line_mappings (user_id, line_user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
		line_user_id = VALUES(line_user_id),
		updated_at = VALUES(updated_at)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, lineUserID, now, now)
	if err != nil {
		return fmt.Errorf("failed to save LINE user ID: %w", err)
	}

	return nil
}

// GetUnreadCount はユーザーの未読通知数を取得します
func (r *mysqlNotificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM notifications 
		WHERE user_id = ? AND status = ?
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, entity.NotificationStatusUnread).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}

	return count, nil
}
