// internal/modules/social/interface/database/friendship_repository.go
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type FriendshipRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewFriendshipRepository(db *sql.DB, logger logger.Logger) usecase.FriendshipRepository {
	return &FriendshipRepository{
		db:     db,
		logger: logger,
	}
}

// CreateFriendship は友達関係を作成する
func (r *FriendshipRepository) CreateFriendship(ctx context.Context, friendship *domain.Friendship) error {
	query := `
		INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		friendship.ID,
		friendship.RequesterID,
		friendship.AddresseeID,
		friendship.Status,
		friendship.CreatedAt,
		friendship.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create friendship",
			logger.Any("friendship", friendship),
			logger.Error(err))
		return fmt.Errorf("failed to create friendship: %w", err)
	}

	return nil
}

// GetFriendship は友達関係を取得する
func (r *FriendshipRepository) GetFriendship(ctx context.Context, requesterID, addresseeID uuid.UUID) (*domain.Friendship, error) {
	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at, accepted_at, blocked_at
		FROM friendships
		WHERE (requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)
	`

	var friendship domain.Friendship
	var acceptedAt, blockedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, requesterID, addresseeID, addresseeID, requesterID).Scan(
		&friendship.ID,
		&friendship.RequesterID,
		&friendship.AddresseeID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.UpdatedAt,
		&acceptedAt,
		&blockedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get friendship",
			logger.Any("requesterID", requesterID),
			logger.Any("addresseeID", addresseeID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get friendship: %w", err)
	}

	if acceptedAt.Valid {
		friendship.AcceptedAt = &acceptedAt.Time
	}
	if blockedAt.Valid {
		friendship.BlockedAt = &blockedAt.Time
	}

	return &friendship, nil
}

// GetFriendshipByID はIDで友達関係を取得する
func (r *FriendshipRepository) GetFriendshipByID(ctx context.Context, friendshipID uuid.UUID) (*domain.Friendship, error) {
	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at, accepted_at, blocked_at
		FROM friendships
		WHERE id = ?
	`

	var friendship domain.Friendship
	var acceptedAt, blockedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, friendshipID).Scan(
		&friendship.ID,
		&friendship.RequesterID,
		&friendship.AddresseeID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.UpdatedAt,
		&acceptedAt,
		&blockedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get friendship by ID",
			logger.Any("friendshipID", friendshipID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get friendship by ID: %w", err)
	}

	if acceptedAt.Valid {
		friendship.AcceptedAt = &acceptedAt.Time
	}
	if blockedAt.Valid {
		friendship.BlockedAt = &blockedAt.Time
	}

	return &friendship, nil
}

// UpdateFriendship は友達関係を更新する
func (r *FriendshipRepository) UpdateFriendship(ctx context.Context, friendship *domain.Friendship) error {
	query := `
		UPDATE friendships 
		SET status = ?, updated_at = ?, accepted_at = ?, blocked_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		friendship.Status,
		friendship.UpdatedAt,
		friendship.AcceptedAt,
		friendship.BlockedAt,
		friendship.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update friendship",
			logger.Any("friendship", friendship),
			logger.Error(err))
		return fmt.Errorf("failed to update friendship: %w", err)
	}

	return nil
}

// DeleteFriendship は友達関係を削除する
func (r *FriendshipRepository) DeleteFriendship(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	query := `
		DELETE FROM friendships 
		WHERE (requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)
	`

	_, err := r.db.ExecContext(ctx, query, requesterID, addresseeID, addresseeID, requesterID)
	if err != nil {
		r.logger.Error("Failed to delete friendship",
			logger.Any("requesterID", requesterID),
			logger.Any("addresseeID", addresseeID),
			logger.Error(err))
		return fmt.Errorf("failed to delete friendship: %w", err)
	}

	return nil
}

// GetFriends は友達一覧を取得する
func (r *FriendshipRepository) GetFriends(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Friendship, error) {
	offset := (pagination.Page - 1) * pagination.PageSize

	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at, accepted_at, blocked_at
		FROM friendships
		WHERE (requester_id = ? OR addressee_id = ?) AND status = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, userID, domain.FriendshipStatusAccepted, pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to get friends",
			logger.Any("userID", userID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}
	defer rows.Close()

	var friendships []*domain.Friendship
	for rows.Next() {
		var friendship domain.Friendship
		var acceptedAt, blockedAt sql.NullTime

		err := rows.Scan(
			&friendship.ID,
			&friendship.RequesterID,
			&friendship.AddresseeID,
			&friendship.Status,
			&friendship.CreatedAt,
			&friendship.UpdatedAt,
			&acceptedAt,
			&blockedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan friendship", logger.Error(err))
			continue
		}

		if acceptedAt.Valid {
			friendship.AcceptedAt = &acceptedAt.Time
		}
		if blockedAt.Valid {
			friendship.BlockedAt = &blockedAt.Time
		}

		friendships = append(friendships, &friendship)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating friendship rows", logger.Error(err))
		return nil, fmt.Errorf("error iterating friendship rows: %w", err)
	}

	return friendships, nil
}

// GetPendingRequests は受信した友達申請を取得する
func (r *FriendshipRepository) GetPendingRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Friendship, error) {
	offset := (pagination.Page - 1) * pagination.PageSize

	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at, accepted_at, blocked_at
		FROM friendships
		WHERE addressee_id = ? AND status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, domain.FriendshipStatusPending, pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to get pending requests",
			logger.Any("userID", userID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}
	defer rows.Close()

	var friendships []*domain.Friendship
	for rows.Next() {
		var friendship domain.Friendship
		var acceptedAt, blockedAt sql.NullTime

		err := rows.Scan(
			&friendship.ID,
			&friendship.RequesterID,
			&friendship.AddresseeID,
			&friendship.Status,
			&friendship.CreatedAt,
			&friendship.UpdatedAt,
			&acceptedAt,
			&blockedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan pending request", logger.Error(err))
			continue
		}

		if acceptedAt.Valid {
			friendship.AcceptedAt = &acceptedAt.Time
		}
		if blockedAt.Valid {
			friendship.BlockedAt = &blockedAt.Time
		}

		friendships = append(friendships, &friendship)
	}

	return friendships, nil
}

// GetSentRequests は送信した友達申請を取得する
func (r *FriendshipRepository) GetSentRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Friendship, error) {
	offset := (pagination.Page - 1) * pagination.PageSize

	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at, accepted_at, blocked_at
		FROM friendships
		WHERE requester_id = ? AND status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, domain.FriendshipStatusPending, pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to get sent requests",
			logger.Any("userID", userID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get sent requests: %w", err)
	}
	defer rows.Close()

	var friendships []*domain.Friendship
	for rows.Next() {
		var friendship domain.Friendship
		var acceptedAt, blockedAt sql.NullTime

		err := rows.Scan(
			&friendship.ID,
			&friendship.RequesterID,
			&friendship.AddresseeID,
			&friendship.Status,
			&friendship.CreatedAt,
			&friendship.UpdatedAt,
			&acceptedAt,
			&blockedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan sent request", logger.Error(err))
			continue
		}

		if acceptedAt.Valid {
			friendship.AcceptedAt = &acceptedAt.Time
		}
		if blockedAt.Valid {
			friendship.BlockedAt = &blockedAt.Time
		}

		friendships = append(friendships, &friendship)
	}

	return friendships, nil
}

// AreFriends は友達関係が成立しているかチェックする
func (r *FriendshipRepository) AreFriends(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*) FROM friendships
		WHERE ((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?))
		AND status = ?
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID1, userID2, userID2, userID1, domain.FriendshipStatusAccepted).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to check if users are friends",
			logger.Any("userID1", userID1),
			logger.Any("userID2", userID2),
			logger.Error(err))
		return false, fmt.Errorf("failed to check if users are friends: %w", err)
	}

	return count > 0, nil
}

// IsBlocked はブロック関係があるかチェックする
func (r *FriendshipRepository) IsBlocked(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*) FROM friendships
		WHERE ((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?))
		AND status = ?
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID1, userID2, userID2, userID1, domain.FriendshipStatusBlocked).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to check if user is blocked",
			logger.Any("userID1", userID1),
			logger.Any("userID2", userID2),
			logger.Error(err))
		return false, fmt.Errorf("failed to check if user is blocked: %w", err)
	}

	return count > 0, nil
}

// GetFriendCount は友達数を取得する
func (r *FriendshipRepository) GetFriendCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM friendships
		WHERE (requester_id = ? OR addressee_id = ?) AND status = ?
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, userID, domain.FriendshipStatusAccepted).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to get friend count",
			logger.Any("userID", userID),
			logger.Error(err))
		return 0, fmt.Errorf("failed to get friend count: %w", err)
	}

	return count, nil
}

// GetMutualFriends は共通の友達を取得する
func (r *FriendshipRepository) GetMutualFriends(ctx context.Context, userID1, userID2 uuid.UUID) ([]*domain.Friendship, error) {
	query := `
		SELECT DISTINCT f1.id, f1.requester_id, f1.addressee_id, f1.status, f1.created_at, f1.updated_at, f1.accepted_at, f1.blocked_at
		FROM friendships f1
		JOIN friendships f2 ON (
			(f1.requester_id = f2.requester_id OR f1.requester_id = f2.addressee_id OR 
			 f1.addressee_id = f2.requester_id OR f1.addressee_id = f2.addressee_id)
			AND f1.id != f2.id
		)
		WHERE f1.status = ? AND f2.status = ?
		AND ((f1.requester_id = ? OR f1.addressee_id = ?) AND (f2.requester_id = ? OR f2.addressee_id = ?))
		AND f1.requester_id != ? AND f1.addressee_id != ? AND f2.requester_id != ? AND f2.addressee_id != ?
	`

	rows, err := r.db.QueryContext(ctx, query,
		domain.FriendshipStatusAccepted, domain.FriendshipStatusAccepted,
		userID1, userID1, userID2, userID2,
		userID1, userID1, userID2, userID2)
	if err != nil {
		r.logger.Error("Failed to get mutual friends",
			logger.Any("userID1", userID1),
			logger.Any("userID2", userID2),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get mutual friends: %w", err)
	}
	defer rows.Close()

	var friendships []*domain.Friendship
	for rows.Next() {
		var friendship domain.Friendship
		var acceptedAt, blockedAt sql.NullTime

		err := rows.Scan(
			&friendship.ID,
			&friendship.RequesterID,
			&friendship.AddresseeID,
			&friendship.Status,
			&friendship.CreatedAt,
			&friendship.UpdatedAt,
			&acceptedAt,
			&blockedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan mutual friend", logger.Error(err))
			continue
		}

		if acceptedAt.Valid {
			friendship.AcceptedAt = &acceptedAt.Time
		}
		if blockedAt.Valid {
			friendship.BlockedAt = &blockedAt.Time
		}

		friendships = append(friendships, &friendship)
	}

	return friendships, nil
}
