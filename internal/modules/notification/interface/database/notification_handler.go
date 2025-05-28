package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// NotificationServiceRepository はSQLを使用した通知リポジトリの実装
type NotificationServiceRepository struct {
	SqlHandler
	Logger logger.Logger
}

// Save は通知を保存する
func (r *NotificationServiceRepository) Save(ctx context.Context, notification *domain.Notification) error {
	// メタデータをJSON文字列に変換
	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		r.Logger.Error("Failed to marshal metadata", logger.Error(err))
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// 送信日時の処理
	var sentAt interface{}
	if notification.SentAt != nil {
		sentAt = *notification.SentAt
	} else {
		sentAt = nil
	}

	query := `
		INSERT INTO ` + "`Yotei-Plus`" + `.notifications (
			id, user_id, title, message, type, status, metadata, created_at, updated_at, sent_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			title = VALUES(title),
			message = VALUES(message),
			type = VALUES(type),
			status = VALUES(status),
			metadata = VALUES(metadata),
			updated_at = VALUES(updated_at),
			sent_at = VALUES(sent_at)
	`

	_, err = r.Execute(
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
		r.Logger.Error("Failed to save notification", logger.Any("id", notification.ID), logger.Error(err))
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
			` + "`Yotei-Plus`" + `.notifications
		WHERE 
			id = ?
	`

	row, err := r.Query(query, id)
	if err != nil {
		r.Logger.Error("Failed to query notification", logger.Any("id", id), logger.Error(err))
		return nil, fmt.Errorf("failed to query notification: %w", err)
	}
	defer row.Close()

	if !row.Next() {
		return nil, nil // 通知が見つからない場合
	}

	var (
		notification domain.Notification
		metadataJSON []byte
		sentAt       sql.NullTime
	)

	err = row.Scan(
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
		r.Logger.Error("Failed to scan notification", logger.Any("id", id), logger.Error(err))
		return nil, fmt.Errorf("failed to scan notification: %w", err)
	}

	// メタデータのデコード
	if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
		r.Logger.Error("Failed to unmarshal metadata", logger.Error(err))
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
			` + "`Yotei-Plus`" + `.notifications
		WHERE 
			user_id = ?
		ORDER BY 
			created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.Query(query, userID, limit, offset)
	if err != nil {
		r.Logger.Error("Failed to query notifications", logger.Any("userID", userID), logger.Error(err))
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
			r.Logger.Error("Failed to scan notification row", logger.Error(err))
			return nil, fmt.Errorf("failed to scan notification row: %w", err)
		}

		// メタデータのデコード
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			r.Logger.Error("Failed to unmarshal metadata", logger.Error(err))
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		// 送信日時の処理
		if sentAt.Valid {
			notification.SentAt = &sentAt.Time
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// UpdateStatus は通知のステータスを更新する
func (r *NotificationServiceRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error {
	now := time.Now()

	query := `
		UPDATE ` + "`Yotei-Plus`" + `.notifications
		SET 
			status = ?,
			updated_at = ?,
			sent_at = CASE 
				WHEN ? = 'SENT' THEN ?
				ELSE sent_at
			END
		WHERE 
			id = ?
	`

	var sentAt interface{}
	if status == domain.StatusSent {
		sentAt = now
	} else {
		sentAt = nil
	}

	result, err := r.Execute(query, status, now, status, sentAt, id)
	if err != nil {
		r.Logger.Error("Failed to update notification status", logger.Any("id", id), logger.Error(err))
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.Logger.Error("Failed to get rows affected", logger.Error(err))
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
		FROM ` + "`Yotei-Plus`" + `.notifications
		WHERE user_id = ? AND status = ?
	`

	row, err := r.Query(query, userID, status)
	if err != nil {
		r.Logger.Error("Failed to query notification count", logger.Any("userID", userID), logger.Any("status", status), logger.Error(err))
		return 0, fmt.Errorf("failed to query notification count: %w", err)
	}
	defer row.Close()

	var count int
	if row.Next() {
		if err := row.Scan(&count); err != nil {
			r.Logger.Error("Failed to scan count", logger.Error(err))
			return 0, fmt.Errorf("failed to scan count: %w", err)
		}
	}

	return count, nil
}

// FindPendingNotifications は保留中の通知を取得する
func (r *NotificationServiceRepository) FindPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	query := `
		SELECT 
			id, user_id, title, message, type, status, metadata, created_at, updated_at, sent_at
		FROM 
			` + "`Yotei-Plus`" + `.notifications
		WHERE 
			status = ?
		ORDER BY 
			created_at ASC
		LIMIT ?
	`

	rows, err := r.Query(query, domain.StatusPending, limit)
	if err != nil {
		r.Logger.Error("Failed to query pending notifications", logger.Error(err))
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
			r.Logger.Error("Failed to scan notification row", logger.Error(err))
			return nil, fmt.Errorf("failed to scan notification row: %w", err)
		}

		// メタデータのデコード
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			r.Logger.Error("Failed to unmarshal metadata", logger.Error(err))
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		// 送信日時の処理
		if sentAt.Valid {
			notification.SentAt = &sentAt.Time
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// 通知を削除するメソッド
func (r *NotificationServiceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM ` + "`Yotei-Plus`" + `.notifications WHERE id = ?`

	result, err := r.Execute(query, id)
	if err != nil {
		r.Logger.Error("Failed to delete notification", logger.Any("id", id), logger.Error(err))
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.Logger.Error("Failed to get rows affected", logger.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %s", id)
	}

	return nil
}

func (r *NotificationServiceRepository) MarkAsRead(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, domain.StatusRead)
}

func (r *NotificationServiceRepository) MarkAsSent(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, domain.StatusSent)
}

func (r *NotificationServiceRepository) MarkAsFailed(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, domain.StatusFailed)
}
