package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
)

// NotificationServiceRepository はSQLを使用した通知リポジトリの実装
type NotificationServiceRepository struct {
	SqlHandler
}

// Save は通知を保存する
func (r *NotificationServiceRepository) Save(ctx context.Context, notification *domain.Notification) error {
	// メタデータをJSON文字列に変換
	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		r.logger.Error("Failed to marshal metadata", "error", err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// 送信日時の処理
	var sentAt sql.NullTime
	if notification.SentAt != nil {
		sentAt = sql.NullTime{
			Time:  *notification.SentAt,
			Valid: true,
		}
	}

	// UPSERT操作の実行
	query := `
		INSERT INTO notifications (
			id, user_id, title, message, type, status, metadata, created_at, updated_at, sent_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) ON CONFLICT (id) DO UPDATE SET
			user_id = $2,
			title = $3,
			message = $4,
			type = $5,
			status = $6,
			metadata = $7,
			updated_at = $9,
			sent_at = $10
	`

	_, err = r.ExecContext(
		ctx,
		query,
		notification.ID,
		notification.UserID,
		notification.Title,
		notification.Message,
		notification.Type,
		notification.Status,
		metadataJSON,
		notification.CreatedAt,
		notification.UpdatedAt,
		sentAt,
	)

	if err != nil {
		r.logger.Error("Failed to save notification", "id", notification.ID, "error", err)
		return fmt.Errorf("failed to save notification: %w", err)
	}

	return nil
}

// FindByID は指定されたIDの通知を取得する
func (r *NotificationServiceRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	query := `
		SELECT 
			id, user_id, title, message, type, status, metadata, created_at, updated_at, sent_at
		FROM 
			notifications
		WHERE 
			id = $1
	`

	var (
		notification domain.Notification
		metadataJSON []byte
		sentAt       sql.NullTime
	)

	err := r.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Message,
		&notification.Type,
		&notification.Status,
		&metadataJSON,
		&notification.CreatedAt,
		&notification.UpdatedAt,
		&sentAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 通知が見つからない場合
		}
		r.logger.Error("Failed to find notification", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}

	// メタデータのデコード
	if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
		r.logger.Error("Failed to unmarshal metadata", "error", err)
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// 送信日時の処理
	if sentAt.Valid {
		notification.SentAt = &sentAt.Time
	}

	return &notification, nil
}

// FindByUserID は指定されたユーザーIDの通知を取得する
func (r *NotificationServiceRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Notification, error) {
	query := `
		SELECT 
			id, user_id, title, message, type, status, metadata, created_at, updated_at, sent_at
		FROM 
			notifications
		WHERE 
			user_id = $1
		ORDER BY 
			created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to query notifications", "userID", userID, "error", err)
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		var (
			notification domain.Notification
			metadataJSON []byte
			sentAt       sql.NullTime
		)

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.Status,
			&metadataJSON,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&sentAt,
		)

		if err != nil {
			r.logger.Error("Failed to scan notification row", "error", err)
			return nil, fmt.Errorf("failed to scan notification row: %w", err)
		}

		// メタデータのデコード
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			r.logger.Error("Failed to unmarshal metadata", "error", err)
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		// 送信日時の処理
		if sentAt.Valid {
			notification.SentAt = &sentAt.Time
		}

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating notification rows", "error", err)
		return nil, fmt.Errorf("error iterating notification rows: %w", err)
	}

	return notifications, nil
}

// UpdateStatus は通知のステータスを更新する
func (r *NotificationServiceRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error {
	query := `
		UPDATE notifications
		SET 
			status = $1,
			updated_at = $2,
			sent_at = CASE 
				WHEN $1 = 'sent' THEN $3
				ELSE sent_at
			END
		WHERE 
			id = $4
	`

	now := time.Now()
	var sentAt *time.Time
	if status == domain.StatusSent {
		sentAt = &now
	}

	result, err := r.db.ExecContext(ctx, query, status, now, sentAt, id)
	if err != nil {
		r.logger.Error("Failed to update notification status", "id", id, "error", err)
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %s", id)
	}

	return nil
}

// CountByUserIDAndStatus はユーザーIDとステータスに基づいて通知数を取得する
func (r *NotificationServiceRepository) CountByUserIDAndStatus(ctx context.Context, userID string, status domain.NotificationStatus) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND status = $2
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, status).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count notifications", "userID", userID, "status", status, "error", err)
		return 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	return count, nil
}

// FindPendingNotifications は保留中の通知を取得する
func (r *NotificationServiceRepository) FindPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	query := `
		SELECT 
			id, user_id, title, message, type, status, metadata, created_at, updated_at, sent_at
		FROM 
			notifications
		WHERE 
			status = $1
		ORDER BY 
			created_at ASC
		LIMIT $2
	`

	rows, err := r.QueryContext(ctx, query, domain.StatusPending, limit)
	if err != nil {
		r.logger.Error("Failed to query pending notifications", "error", err)
		return nil, fmt.Errorf("failed to query pending notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		var (
			notification domain.Notification
			metadataJSON []byte
			sentAt       sql.NullTime
		)

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.Status,
			&metadataJSON,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&sentAt,
		)

		if err != nil {
			r.logger.Error("Failed to scan notification row", "error", err)
			return nil, fmt.Errorf("failed to scan notification row: %w", err)
		}

		// メタデータのデコード
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			r.logger.Error("Failed to unmarshal metadata", "error", err)
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		// 送信日時の処理
		if sentAt.Valid {
			notification.SentAt = &sentAt.Time
		}

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating notification rows", "error", err)
		return nil, fmt.Errorf("error iterating notification rows: %w", err)
	}

	return notifications, nil
}
