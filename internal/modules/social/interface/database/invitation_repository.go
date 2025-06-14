package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type InvitationRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewInvitationRepository(db *sql.DB, logger logger.Logger) usecase.InvitationRepository {
	return &InvitationRepository{
		db:     db,
		logger: logger,
	}
}

// CreateInvitation は招待を作成する
func (r *InvitationRepository) CreateInvitation(ctx context.Context, invitation *domain.Invitation) error {
	query := `
		INSERT INTO invitations (
			id, type, method, status, inviter_id, invitee_id, invitee_email, invitee_username, invitee_phone,
			target_id, code, url, message, metadata, expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var metadataJSON []byte
	var err error
	if invitation.Metadata != nil {
		metadataJSON, err = json.Marshal(invitation.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	var inviteeEmail, inviteeUsername, inviteePhone *string
	if invitation.InviteeInfo != nil {
		if invitation.InviteeInfo.Email != "" {
			inviteeEmail = &invitation.InviteeInfo.Email
		}
		if invitation.InviteeInfo.Username != "" {
			inviteeUsername = &invitation.InviteeInfo.Username
		}
		if invitation.InviteeInfo.Phone != "" {
			inviteePhone = &invitation.InviteeInfo.Phone
		}
	}

	_, err = r.db.ExecContext(ctx, query,
		invitation.ID,
		invitation.Type,
		invitation.Method,
		invitation.Status,
		invitation.InviterID,
		invitation.InviteeID,
		inviteeEmail,
		inviteeUsername,
		inviteePhone,
		invitation.TargetID,
		invitation.Code,
		invitation.URL,
		invitation.Message,
		metadataJSON,
		invitation.ExpiresAt,
		invitation.CreatedAt,
		invitation.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create invitation",
			logger.Any("invitation", invitation),
			logger.Error(err))
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

// GetInvitationByID はIDで招待を取得する
func (r *InvitationRepository) GetInvitationByID(ctx context.Context, id uuid.UUID) (*domain.Invitation, error) {
	query := `
		SELECT id, type, method, status, inviter_id, invitee_id, invitee_email, invitee_username, invitee_phone,
			   target_id, code, url, message, metadata, expires_at, created_at, updated_at, accepted_at
		FROM invitations
		WHERE id = ?
	`

	invitation, err := r.scanInvitation(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get invitation by ID",
			logger.Any("id", id),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get invitation by ID: %w", err)
	}

	return invitation, nil
}

// GetInvitationByCode はコードで招待を取得する
func (r *InvitationRepository) GetInvitationByCode(ctx context.Context, code string) (*domain.Invitation, error) {
	query := `
		SELECT id, type, method, status, inviter_id, invitee_id, invitee_email, invitee_username, invitee_phone,
			   target_id, code, url, message, metadata, expires_at, created_at, updated_at, accepted_at
		FROM invitations
		WHERE code = ?
	`

	invitation, err := r.scanInvitation(r.db.QueryRowContext(ctx, query, code))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get invitation by code",
			logger.Any("code", code),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get invitation by code: %w", err)
	}

	return invitation, nil
}

// UpdateInvitation は招待を更新する
func (r *InvitationRepository) UpdateInvitation(ctx context.Context, invitation *domain.Invitation) error {
	query := `
		UPDATE invitations 
		SET status = ?, invitee_id = ?, updated_at = ?, accepted_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		invitation.Status,
		invitation.InviteeID,
		invitation.UpdatedAt,
		invitation.AcceptedAt,
		invitation.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update invitation",
			logger.Any("invitation", invitation),
			logger.Error(err))
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	return nil
}

// DeleteInvitation は招待を削除する
func (r *InvitationRepository) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invitations WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete invitation",
			logger.Any("id", id),
			logger.Error(err))
		return fmt.Errorf("failed to delete invitation: %w", err)
	}

	return nil
}

// GetSentInvitations は送信した招待一覧を取得する
func (r *InvitationRepository) GetSentInvitations(ctx context.Context, inviterID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error) {
	offset := (pagination.Page - 1) * pagination.PageSize

	query := `
		SELECT id, type, method, status, inviter_id, invitee_id, invitee_email, invitee_username, invitee_phone,
			   target_id, code, url, message, metadata, expires_at, created_at, updated_at, accepted_at
		FROM invitations
		WHERE inviter_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, inviterID, pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to get sent invitations",
			logger.Any("inviterID", inviterID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get sent invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*domain.Invitation
	for rows.Next() {
		invitation, err := r.scanInvitationFromRows(rows)
		if err != nil {
			r.logger.Error("Failed to scan invitation", logger.Error(err))
			continue
		}
		invitations = append(invitations, invitation)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating invitation rows", logger.Error(err))
		return nil, fmt.Errorf("error iterating invitation rows: %w", err)
	}

	return invitations, nil
}

// GetReceivedInvitations は受信した招待一覧を取得する
func (r *InvitationRepository) GetReceivedInvitations(ctx context.Context, inviteeID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error) {
	offset := (pagination.Page - 1) * pagination.PageSize

	query := `
		SELECT id, type, method, status, inviter_id, invitee_id, invitee_email, invitee_username, invitee_phone,
			   target_id, code, url, message, metadata, expires_at, created_at, updated_at, accepted_at
		FROM invitations
		WHERE invitee_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, inviteeID, pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to get received invitations",
			logger.Any("inviteeID", inviteeID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get received invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*domain.Invitation
	for rows.Next() {
		invitation, err := r.scanInvitationFromRows(rows)
		if err != nil {
			r.logger.Error("Failed to scan invitation", logger.Error(err))
			continue
		}
		invitations = append(invitations, invitation)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating invitation rows", logger.Error(err))
		return nil, fmt.Errorf("error iterating invitation rows: %w", err)
	}

	return invitations, nil
}

// MarkExpiredInvitations は期限切れ招待をマークする
func (r *InvitationRepository) MarkExpiredInvitations(ctx context.Context) error {
	query := `
		UPDATE invitations 
		SET status = ? 
		WHERE status = ? AND expires_at < NOW()
	`

	_, err := r.db.ExecContext(ctx, query, domain.InvitationStatusExpired, domain.InvitationStatusPending)
	if err != nil {
		r.logger.Error("Failed to mark expired invitations", logger.Error(err))
		return fmt.Errorf("failed to mark expired invitations: %w", err)
	}

	return nil
}

// DeleteExpiredInvitations は期限切れ招待を削除する
func (r *InvitationRepository) DeleteExpiredInvitations(ctx context.Context, beforeDate time.Time) error {
	query := `
		DELETE FROM invitations 
		WHERE status = ? AND expires_at < ?
	`

	_, err := r.db.ExecContext(ctx, query, domain.InvitationStatusExpired, beforeDate)
	if err != nil {
		r.logger.Error("Failed to delete expired invitations", logger.Error(err))
		return fmt.Errorf("failed to delete expired invitations: %w", err)
	}

	return nil
}

// IsValidInvitation は招待コードの妥当性を確認する
func (r *InvitationRepository) IsValidInvitation(ctx context.Context, code string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM invitations
		WHERE code = ? AND status = ? AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, code, domain.InvitationStatusPending).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to validate invitation code",
			logger.Any("code", code),
			logger.Error(err))
		return false, fmt.Errorf("failed to validate invitation code: %w", err)
	}

	return count > 0, nil
}

// scanInvitation はsql.Rowから招待をスキャンする
func (r *InvitationRepository) scanInvitation(row *sql.Row) (*domain.Invitation, error) {
	var invitation domain.Invitation
	var inviteeEmail, inviteeUsername, inviteePhone sql.NullString
	var metadataJSON sql.NullString
	var acceptedAt sql.NullTime

	err := row.Scan(
		&invitation.ID,
		&invitation.Type,
		&invitation.Method,
		&invitation.Status,
		&invitation.InviterID,
		&invitation.InviteeID,
		&inviteeEmail,
		&inviteeUsername,
		&inviteePhone,
		&invitation.TargetID,
		&invitation.Code,
		&invitation.URL,
		&invitation.Message,
		&metadataJSON,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&acceptedAt,
	)

	if err != nil {
		return nil, err
	}

	// InviteeInfoの構築
	if inviteeEmail.Valid || inviteeUsername.Valid || inviteePhone.Valid {
		invitation.InviteeInfo = &domain.InviteeInfo{
			Email:    inviteeEmail.String,
			Username: inviteeUsername.String,
			Phone:    inviteePhone.String,
		}
	}

	// Metadataの解析
	if metadataJSON.Valid {
		err = json.Unmarshal([]byte(metadataJSON.String), &invitation.Metadata)
		if err != nil {
			r.logger.Warn("Failed to unmarshal metadata", logger.Error(err))
		}
	}

	// AcceptedAtの設定
	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}

	return &invitation, nil
}

// scanInvitationFromRows はsql.Rowsから招待をスキャンする
func (r *InvitationRepository) scanInvitationFromRows(rows *sql.Rows) (*domain.Invitation, error) {
	var invitation domain.Invitation
	var inviteeEmail, inviteeUsername, inviteePhone sql.NullString
	var metadataJSON sql.NullString
	var acceptedAt sql.NullTime

	err := rows.Scan(
		&invitation.ID,
		&invitation.Type,
		&invitation.Method,
		&invitation.Status,
		&invitation.InviterID,
		&invitation.InviteeID,
		&inviteeEmail,
		&inviteeUsername,
		&inviteePhone,
		&invitation.TargetID,
		&invitation.Code,
		&invitation.URL,
		&invitation.Message,
		&metadataJSON,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&acceptedAt,
	)

	if err != nil {
		return nil, err
	}

	// InviteeInfoの構築
	if inviteeEmail.Valid || inviteeUsername.Valid || inviteePhone.Valid {
		invitation.InviteeInfo = &domain.InviteeInfo{
			Email:    inviteeEmail.String,
			Username: inviteeUsername.String,
			Phone:    inviteePhone.String,
		}
	}

	// Metadataの解析
	if metadataJSON.Valid {
		err = json.Unmarshal([]byte(metadataJSON.String), &invitation.Metadata)
		if err != nil {
			r.logger.Warn("Failed to unmarshal metadata", logger.Error(err))
		}
	}

	// AcceptedAtの設定
	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}

	return &invitation, nil
}