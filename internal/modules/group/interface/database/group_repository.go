package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
	groupUsecase "github.com/hryt430/Yotei+/internal/modules/group/usecase"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type GroupRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewGroupRepository(db *sql.DB, logger logger.Logger) groupUsecase.GroupRepository {
	return &GroupRepository{
		db:     db,
		logger: logger,
	}
}

// CreateGroup はグループを作成する
func (r *GroupRepository) CreateGroup(ctx context.Context, group *domain.Group) error {
	query := `
		INSERT INTO groups (
			id, name, description, type, owner_id, member_count, 
			is_public, allow_member_invite, require_approval, enable_notifications,
			default_privacy_level, allow_schedule_details, enable_gantt_chart, enable_task_dependency,
			created_at, updated_at, version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		group.ID.String(),
		group.Name,
		group.Description,
		string(group.Type),
		group.OwnerID.String(),
		group.MemberCount,
		group.Settings.IsPublic,
		group.Settings.AllowMemberInvite,
		group.Settings.RequireApproval,
		group.Settings.EnableNotifications,
		group.Settings.DefaultPrivacyLevel,
		group.Settings.AllowScheduleDetails,
		group.Settings.EnableGanttChart,
		group.Settings.EnableTaskDependency,
		group.CreatedAt,
		group.UpdatedAt,
		group.Version,
	)

	if err != nil {
		r.logger.Error("Failed to create group", logger.Error(err))
		return fmt.Errorf("failed to create group: %w", err)
	}

	// オーナーをメンバーとして追加
	member := domain.NewGroupMember(group.ID, group.OwnerID, domain.RoleOwner)
	return r.AddMember(ctx, member)
}

// GetGroupByID はIDでグループを取得する
func (r *GroupRepository) GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	query := `
		SELECT id, name, description, type, owner_id, member_count,
			   is_public, allow_member_invite, require_approval, enable_notifications,
			   default_privacy_level, allow_schedule_details, enable_gantt_chart, enable_task_dependency,
			   created_at, updated_at, version
		FROM groups
		WHERE id = ?
	`

	var group domain.Group
	var idStr, ownerIDStr string
	var defaultPrivacyLevel, allowScheduleDetails, enableGanttChart, enableTaskDependency sql.NullString

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&group.Name,
		&group.Description,
		(*string)(&group.Type),
		&ownerIDStr,
		&group.MemberCount,
		&group.Settings.IsPublic,
		&group.Settings.AllowMemberInvite,
		&group.Settings.RequireApproval,
		&group.Settings.EnableNotifications,
		&defaultPrivacyLevel,
		&allowScheduleDetails,
		&enableGanttChart,
		&enableTaskDependency,
		&group.CreatedAt,
		&group.UpdatedAt,
		&group.Version,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get group by ID", logger.Error(err))
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	group.ID, _ = uuid.Parse(idStr)
	group.OwnerID, _ = uuid.Parse(ownerIDStr)

	// Optional fieldsの処理
	if defaultPrivacyLevel.Valid {
		group.Settings.DefaultPrivacyLevel = domain.PrivacyLevel(defaultPrivacyLevel.String)
	}
	if allowScheduleDetails.Valid {
		group.Settings.AllowScheduleDetails = allowScheduleDetails.String == "1"
	}
	if enableGanttChart.Valid {
		group.Settings.EnableGanttChart = enableGanttChart.String == "1"
	}
	if enableTaskDependency.Valid {
		group.Settings.EnableTaskDependency = enableTaskDependency.String == "1"
	}

	return &group, nil
}

// UpdateGroup はグループを更新する
func (r *GroupRepository) UpdateGroup(ctx context.Context, group *domain.Group) error {
	query := `
		UPDATE groups
		SET name = ?, description = ?, member_count = ?, 
			is_public = ?, allow_member_invite = ?, require_approval = ?, enable_notifications = ?,
			default_privacy_level = ?, allow_schedule_details = ?, enable_gantt_chart = ?, enable_task_dependency = ?,
			updated_at = ?, version = ?
		WHERE id = ? AND version = ?
	`

	oldVersion := group.Version - 1

	result, err := r.db.ExecContext(ctx, query,
		group.Name,
		group.Description,
		group.MemberCount,
		group.Settings.IsPublic,
		group.Settings.AllowMemberInvite,
		group.Settings.RequireApproval,
		group.Settings.EnableNotifications,
		group.Settings.DefaultPrivacyLevel,
		group.Settings.AllowScheduleDetails,
		group.Settings.EnableGanttChart,
		group.Settings.EnableTaskDependency,
		group.UpdatedAt,
		group.Version,
		group.ID.String(),
		oldVersion,
	)

	if err != nil {
		r.logger.Error("Failed to update group", logger.Error(err))
		return fmt.Errorf("failed to update group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group not found or version conflict")
	}

	return nil
}

// DeleteGroup はグループを削除する
func (r *GroupRepository) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// メンバーを削除
	_, err = tx.ExecContext(ctx, "DELETE FROM group_members WHERE group_id = ?", id.String())
	if err != nil {
		r.logger.Error("Failed to delete group members", logger.Error(err))
		return fmt.Errorf("failed to delete group members: %w", err)
	}

	// グループを削除
	_, err = tx.ExecContext(ctx, "DELETE FROM groups WHERE id = ?", id.String())
	if err != nil {
		r.logger.Error("Failed to delete group", logger.Error(err))
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return tx.Commit()
}

// ListGroupsByOwner はオーナーでグループを検索する
func (r *GroupRepository) ListGroupsByOwner(ctx context.Context, ownerID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Group, int, error) {
	// 総数を取得
	countQuery := "SELECT COUNT(*) FROM groups WHERE owner_id = ?"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, ownerID.String()).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count groups by owner", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to count groups: %w", err)
	}

	// データを取得
	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT id, name, description, type, owner_id, settings, member_count, created_at, updated_at, version
		FROM groups
		WHERE owner_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID.String(), pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to list groups by owner", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	groups, err := r.scanGroups(rows)
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// ListGroupsByMember はメンバーでグループを検索する
func (r *GroupRepository) ListGroupsByMember(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Group, int, error) {
	// 総数を取得
	countQuery := `
		SELECT COUNT(*)
		FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
	`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID.String()).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count groups by member", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to count groups: %w", err)
	}

	// データを取得
	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT g.id, g.name, g.description, g.type, g.owner_id, g.settings, g.member_count, g.created_at, g.updated_at, g.version
		FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
		ORDER BY g.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID.String(), pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to list groups by member", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	groups, err := r.scanGroups(rows)
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// SearchGroups はグループを検索する
func (r *GroupRepository) SearchGroups(ctx context.Context, query string, groupType *domain.GroupType, pagination commonDomain.Pagination) ([]*domain.Group, int, error) {
	// 条件構築
	conditions := []string{"(g.name LIKE ? OR g.description LIKE ?)"}
	args := []interface{}{"%" + query + "%", "%" + query + "%"}

	if groupType != nil {
		conditions = append(conditions, "g.type = ?")
		args = append(args, string(*groupType))
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 総数を取得
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM groups g %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count search results", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// データを取得
	offset := (pagination.Page - 1) * pagination.PageSize
	searchQuery := fmt.Sprintf(`
		SELECT g.id, g.name, g.description, g.type, g.owner_id, g.settings, g.member_count, g.created_at, g.updated_at, g.version
		FROM groups g
		%s
		ORDER BY g.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pagination.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, searchQuery, args...)
	if err != nil {
		r.logger.Error("Failed to search groups", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to search groups: %w", err)
	}
	defer rows.Close()

	groups, err := r.scanGroups(rows)
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// AddMember はメンバーを追加する
func (r *GroupRepository) AddMember(ctx context.Context, member *domain.GroupMember) error {
	query := `
		INSERT INTO group_members (id, group_id, user_id, role, joined_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		member.ID.String(),
		member.GroupID.String(),
		member.UserID.String(),
		string(member.Role),
		member.JoinedAt,
		member.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to add member", logger.Error(err))
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

// GetMember はメンバーを取得する
func (r *GroupRepository) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error) {
	query := `
		SELECT id, group_id, user_id, role, joined_at, updated_at
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`

	var member domain.GroupMember
	var idStr, groupIDStr, userIDStr string

	err := r.db.QueryRowContext(ctx, query, groupID.String(), userID.String()).Scan(
		&idStr,
		&groupIDStr,
		&userIDStr,
		(*string)(&member.Role),
		&member.JoinedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get member", logger.Error(err))
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	member.ID, _ = uuid.Parse(idStr)
	member.GroupID, _ = uuid.Parse(groupIDStr)
	member.UserID, _ = uuid.Parse(userIDStr)

	return &member, nil
}

// UpdateMemberRole はメンバーの権限を更新する
func (r *GroupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role domain.MemberRole) error {
	query := `
		UPDATE group_members
		SET role = ?, updated_at = ?
		WHERE group_id = ? AND user_id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		string(role),
		sql.Named("updated_at", "NOW()"),
		groupID.String(),
		userID.String(),
	)

	if err != nil {
		r.logger.Error("Failed to update member role", logger.Error(err))
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

// RemoveMember はメンバーを削除する
func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	query := "DELETE FROM group_members WHERE group_id = ? AND user_id = ?"

	_, err := r.db.ExecContext(ctx, query, groupID.String(), userID.String())
	if err != nil {
		r.logger.Error("Failed to remove member", logger.Error(err))
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// ListMembers はメンバー一覧を取得する
func (r *GroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.GroupMember, error) {
	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT id, group_id, user_id, role, joined_at, updated_at
		FROM group_members
		WHERE group_id = ?
		ORDER BY joined_at ASC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, groupID.String(), pagination.PageSize, offset)
	if err != nil {
		r.logger.Error("Failed to list members", logger.Error(err))
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()

	var members []*domain.GroupMember
	for rows.Next() {
		var member domain.GroupMember
		var idStr, groupIDStr, userIDStr string

		err := rows.Scan(
			&idStr,
			&groupIDStr,
			&userIDStr,
			(*string)(&member.Role),
			&member.JoinedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan member", logger.Error(err))
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}

		member.ID, _ = uuid.Parse(idStr)
		member.GroupID, _ = uuid.Parse(groupIDStr)
		member.UserID, _ = uuid.Parse(userIDStr)

		members = append(members, &member)
	}

	return members, nil
}

// IsMember はメンバーかどうかチェックする
func (r *GroupRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	query := "SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?"

	var count int
	err := r.db.QueryRowContext(ctx, query, groupID.String(), userID.String()).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to check membership", logger.Error(err))
		return false, fmt.Errorf("failed to check membership: %w", err)
	}

	return count > 0, nil
}

// GetMemberRole はメンバーの権限を取得する
func (r *GroupRepository) GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (domain.MemberRole, error) {
	query := "SELECT role FROM group_members WHERE group_id = ? AND user_id = ?"

	var role string
	err := r.db.QueryRowContext(ctx, query, groupID.String(), userID.String()).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("member not found")
		}
		r.logger.Error("Failed to get member role", logger.Error(err))
		return "", fmt.Errorf("failed to get member role: %w", err)
	}

	return domain.MemberRole(role), nil
}

// GetMemberCount はメンバー数を取得する
func (r *GroupRepository) GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	query := "SELECT COUNT(*) FROM group_members WHERE group_id = ?"

	var count int
	err := r.db.QueryRowContext(ctx, query, groupID.String()).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to get member count", logger.Error(err))
		return 0, fmt.Errorf("failed to get member count: %w", err)
	}

	return count, nil
}

// GetGroupStats はグループ統計情報を取得する
func (r *GroupRepository) GetGroupStats(ctx context.Context, groupID uuid.UUID) (*groupUsecase.GroupStats, error) {
	stats := &groupUsecase.GroupStats{}

	// メンバー数取得
	memberCount, err := r.GetMemberCount(ctx, groupID)
	if err != nil {
		return nil, err
	}
	stats.MemberCount = memberCount

	// TODO: タスク数と予定数の取得は該当するモジュールのリポジトリと連携する必要がある
	// 現在は基本的な統計のみ実装
	stats.ActiveMembers = memberCount // 簡易実装

	return stats, nil
}

// === ヘルパーメソッド ===

func (r *GroupRepository) scanGroups(rows *sql.Rows) ([]*domain.Group, error) {
	var groups []*domain.Group

	for rows.Next() {
		var group domain.Group
		var settingsJSON string
		var idStr, ownerIDStr string

		err := rows.Scan(
			&idStr,
			&group.Name,
			&group.Description,
			(*string)(&group.Type),
			&ownerIDStr,
			&settingsJSON,
			&group.MemberCount,
			&group.CreatedAt,
			&group.UpdatedAt,
			&group.Version,
		)
		if err != nil {
			r.logger.Error("Failed to scan group", logger.Error(err))
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		group.ID, _ = uuid.Parse(idStr)
		group.OwnerID, _ = uuid.Parse(ownerIDStr)
		group.Settings = r.decodeGroupSettings(settingsJSON)

		groups = append(groups, &group)
	}

	return groups, nil
}

func (r *GroupRepository) encodeGroupSettings(settings domain.GroupSettings) string {
	// 簡易実装：JSONエンコード（実際の実装では適切なJSONライブラリを使用）
	return fmt.Sprintf(`{
		"is_public": %t,
		"allow_member_invite": %t,
		"require_approval": %t,
		"enable_notifications": %t,
		"default_privacy_level": "%s",
		"allow_schedule_details": %t,
		"enable_gantt_chart": %t,
		"enable_task_dependency": %t
	}`,
		settings.IsPublic,
		settings.AllowMemberInvite,
		settings.RequireApproval,
		settings.EnableNotifications,
		string(settings.DefaultPrivacyLevel),
		settings.AllowScheduleDetails,
		settings.EnableGanttChart,
		settings.EnableTaskDependency,
	)
}

func (r *GroupRepository) decodeGroupSettings(settingsJSON string) domain.GroupSettings {
	// 簡易実装：デフォルト値を返す（実際の実装ではJSONパースを行う）
	return domain.GroupSettings{
		IsPublic:             false,
		AllowMemberInvite:    true,
		RequireApproval:      true,
		EnableNotifications:  true,
		DefaultPrivacyLevel:  domain.PrivacyLevelBusy,
		AllowScheduleDetails: false,
		EnableGanttChart:     true,
		EnableTaskDependency: false,
	}
}
