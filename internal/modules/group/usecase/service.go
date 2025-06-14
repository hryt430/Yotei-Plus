package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

type groupService struct {
	groupRepo     GroupRepository
	userValidator commonDomain.UserValidator
	logger        logger.Logger
}

func NewGroupService(
	groupRepo GroupRepository,
	userValidator commonDomain.UserValidator,
	logger logger.Logger,
) GroupService {
	return &groupService{
		groupRepo:     groupRepo,
		userValidator: userValidator,
		logger:        logger,
	}
}

// CreateGroup はグループを作成する
func (s *groupService) CreateGroup(ctx context.Context, input CreateGroupInput) (*domain.Group, error) {
	// 入力バリデーション
	if err := s.validateCreateGroupInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// オーナーIDを入力から取得（修正）
	ownerID := input.OwnerID

	exists, err := s.userValidator.UserExists(ctx, ownerID.String())
	if err != nil {
		s.logger.Error("Failed to validate owner existence", logger.Error(err))
		return nil, fmt.Errorf("failed to validate owner: %w", err)
	}
	if !exists {
		return nil, errors.New("owner not found")
	}

	// グループ作成
	group := domain.NewGroup(input.Name, input.Description, input.Type, ownerID)
	group.UpdateSettings(input.Settings)

	err = s.groupRepo.CreateGroup(ctx, group)
	if err != nil {
		s.logger.Error("Failed to create group", logger.Error(err))
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	s.logger.Info("Group created successfully", logger.Any("groupID", group.ID))
	return group, nil
}

// GetGroup はグループ詳細を取得する
func (s *groupService) GetGroup(ctx context.Context, groupID uuid.UUID, requesterID uuid.UUID) (*GroupWithMembers, error) {
	// グループ取得
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return nil, errors.New("group not found")
	}

	// メンバーシップ確認
	isMember, err := s.groupRepo.IsMember(ctx, groupID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember && !group.Settings.IsPublic {
		return nil, errors.New("access denied")
	}

	// リクエスターの権限取得
	var myRole domain.MemberRole
	if isMember {
		myRole, err = s.groupRepo.GetMemberRole(ctx, groupID, requesterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get member role: %w", err)
		}
	}

	// メンバー一覧取得
	pagination := commonDomain.Pagination{Page: 1, PageSize: 100}
	members, err := s.groupRepo.ListMembers(ctx, groupID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	// ユーザー情報を一括取得
	memberWithUserInfo, err := s.enrichMembersWithUserInfo(ctx, members)
	if err != nil {
		s.logger.Error("Failed to enrich members with user info", logger.Error(err))
		// エラーでも継続（ユーザー情報なしで返す）
		memberWithUserInfo = make([]*MemberWithUserInfo, len(members))
		for i, member := range members {
			memberWithUserInfo[i] = &MemberWithUserInfo{
				Member:   member,
				UserInfo: nil,
			}
		}
	}

	return &GroupWithMembers{
		Group:   group,
		Members: memberWithUserInfo,
		MyRole:  myRole,
	}, nil
}

// UpdateGroup はグループ情報を更新する
func (s *groupService) UpdateGroup(ctx context.Context, groupID uuid.UUID, input UpdateGroupInput, requesterID uuid.UUID) (*domain.Group, error) {
	// 権限チェック
	hasPermission, err := s.CheckPermission(ctx, groupID, requesterID, ActionEditGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("insufficient permissions")
	}

	// グループ取得
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return nil, errors.New("group not found")
	}

	// 更新適用
	hasChanges := false
	if input.Name != nil && *input.Name != group.Name {
		group.Name = *input.Name
		hasChanges = true
	}
	if input.Description != nil && *input.Description != group.Description {
		group.Description = *input.Description
		hasChanges = true
	}
	if input.Settings != nil {
		group.UpdateSettings(*input.Settings)
		hasChanges = true
	}

	if !hasChanges {
		return group, nil
	}

	// 更新実行
	err = s.groupRepo.UpdateGroup(ctx, group)
	if err != nil {
		s.logger.Error("Failed to update group", logger.Error(err))
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	s.logger.Info("Group updated successfully", logger.Any("groupID", groupID))
	return group, nil
}

// DeleteGroup はグループを削除する
func (s *groupService) DeleteGroup(ctx context.Context, groupID uuid.UUID, requesterID uuid.UUID) error {
	// 権限チェック（オーナーのみ）
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return errors.New("group not found")
	}
	if group.OwnerID != requesterID {
		return errors.New("only owner can delete group")
	}

	// 削除実行
	err = s.groupRepo.DeleteGroup(ctx, groupID)
	if err != nil {
		s.logger.Error("Failed to delete group", logger.Error(err))
		return fmt.Errorf("failed to delete group: %w", err)
	}

	s.logger.Info("Group deleted successfully", logger.Any("groupID", groupID))
	return nil
}

// GetMyGroups は自分のグループ一覧を取得する
func (s *groupService) GetMyGroups(ctx context.Context, userID uuid.UUID, groupType *domain.GroupType, pagination commonDomain.Pagination) ([]*domain.Group, int, error) {
	// オーナーのグループ取得
	ownedGroups, ownedTotal, err := s.groupRepo.ListGroupsByOwner(ctx, userID, pagination)
	if err != nil {
		s.logger.Error("Failed to get owned groups", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to get owned groups: %w", err)
	}

	// メンバーのグループ取得
	memberGroups, memberTotal, err := s.groupRepo.ListGroupsByMember(ctx, userID, pagination)
	if err != nil {
		s.logger.Error("Failed to get member groups", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to get member groups: %w", err)
	}

	// 重複除去してマージ
	groupMap := make(map[uuid.UUID]*domain.Group)
	for _, group := range ownedGroups {
		if groupType == nil || group.Type == *groupType {
			groupMap[group.ID] = group
		}
	}
	for _, group := range memberGroups {
		if groupType == nil || group.Type == *groupType {
			groupMap[group.ID] = group
		}
	}

	// 結果をスライスに変換
	groups := make([]*domain.Group, 0, len(groupMap))
	for _, group := range groupMap {
		groups = append(groups, group)
	}

	total := ownedTotal + memberTotal
	return groups, total, nil
}

// SearchGroups はグループを検索する
func (s *groupService) SearchGroups(ctx context.Context, query string, groupType *domain.GroupType, pagination commonDomain.Pagination) ([]*domain.Group, int, error) {
	groups, total, err := s.groupRepo.SearchGroups(ctx, query, groupType, pagination)
	if err != nil {
		s.logger.Error("Failed to search groups", logger.Error(err))
		return nil, 0, fmt.Errorf("failed to search groups: %w", err)
	}

	return groups, total, nil
}

// AddMember はメンバーを追加する
func (s *groupService) AddMember(ctx context.Context, groupID, userID, inviterID uuid.UUID, role domain.MemberRole) error {
	// 権限チェック
	hasPermission, err := s.CheckPermission(ctx, groupID, inviterID, ActionInviteMembers)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return errors.New("insufficient permissions")
	}

	// ユーザー存在確認
	exists, err := s.userValidator.UserExists(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return errors.New("user not found")
	}

	// 既にメンバーかチェック
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return errors.New("user is already a member")
	}

	// メンバー追加
	member := domain.NewGroupMember(groupID, userID, role)
	err = s.groupRepo.AddMember(ctx, member)
	if err != nil {
		s.logger.Error("Failed to add member", logger.Error(err))
		return fmt.Errorf("failed to add member: %w", err)
	}

	// グループのメンバー数更新
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group for member count update: %w", err)
	}
	group.AddMember()
	err = s.groupRepo.UpdateGroup(ctx, group)
	if err != nil {
		s.logger.Error("Failed to update group member count", logger.Error(err))
	}

	s.logger.Info("Member added successfully",
		logger.Any("groupID", groupID),
		logger.Any("userID", userID))
	return nil
}

// RemoveMember はメンバーを削除する
func (s *groupService) RemoveMember(ctx context.Context, groupID, userID, requesterID uuid.UUID) error {
	// 権限チェック
	hasPermission, err := s.CheckPermission(ctx, groupID, requesterID, ActionRemoveMembers)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission && requesterID != userID {
		return errors.New("insufficient permissions")
	}

	// メンバー削除
	err = s.groupRepo.RemoveMember(ctx, groupID, userID)
	if err != nil {
		s.logger.Error("Failed to remove member", logger.Error(err))
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// グループのメンバー数更新
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group for member count update: %w", err)
	}
	err = group.RemoveMember()
	if err != nil {
		return fmt.Errorf("failed to update group member count: %w", err)
	}
	err = s.groupRepo.UpdateGroup(ctx, group)
	if err != nil {
		s.logger.Error("Failed to update group member count", logger.Error(err))
	}

	s.logger.Info("Member removed successfully",
		logger.Any("groupID", groupID),
		logger.Any("userID", userID))
	return nil
}

// UpdateMemberRole はメンバーの権限を変更する
func (s *groupService) UpdateMemberRole(ctx context.Context, groupID, userID, requesterID uuid.UUID, newRole domain.MemberRole) error {
	// 権限チェック
	hasPermission, err := s.CheckPermission(ctx, groupID, requesterID, ActionManageRoles)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return errors.New("insufficient permissions")
	}

	// オーナーの変更は不可
	requesterRole, err := s.groupRepo.GetMemberRole(ctx, groupID, requesterID)
	if err != nil {
		return fmt.Errorf("failed to get requester role: %w", err)
	}

	targetRole, err := s.groupRepo.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to get target role: %w", err)
	}

	if targetRole == domain.RoleOwner && requesterRole != domain.RoleOwner {
		return errors.New("cannot change owner role")
	}

	// 権限更新
	err = s.groupRepo.UpdateMemberRole(ctx, groupID, userID, newRole)
	if err != nil {
		s.logger.Error("Failed to update member role", logger.Error(err))
		return fmt.Errorf("failed to update member role: %w", err)
	}

	s.logger.Info("Member role updated successfully",
		logger.Any("groupID", groupID),
		logger.Any("userID", userID),
		logger.Any("newRole", newRole))
	return nil
}

// GetMembers はメンバー一覧を取得する
func (s *groupService) GetMembers(ctx context.Context, groupID uuid.UUID, pagination commonDomain.Pagination) ([]*MemberWithUserInfo, error) {
	members, err := s.groupRepo.ListMembers(ctx, groupID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return s.enrichMembersWithUserInfo(ctx, members)
}

// CheckPermission は権限をチェックする
func (s *groupService) CheckPermission(ctx context.Context, groupID, userID uuid.UUID, action GroupAction) (bool, error) {
	// メンバーかどうかチェック
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return false, nil
	}

	// 権限取得
	role, err := s.groupRepo.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get member role: %w", err)
	}

	// 権限チェック
	return s.hasPermissionForAction(role, action), nil
}

// GetUserRole はユーザーの権限を取得する
func (s *groupService) GetUserRole(ctx context.Context, groupID, userID uuid.UUID) (domain.MemberRole, error) {
	return s.groupRepo.GetMemberRole(ctx, groupID, userID)
}

// GetGroupStats はグループ統計情報を取得する
func (s *groupService) GetGroupStats(ctx context.Context, groupID uuid.UUID, requesterID uuid.UUID) (*GroupStats, error) {
	// 権限チェック
	hasPermission, err := s.CheckPermission(ctx, groupID, requesterID, ActionViewGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("insufficient permissions")
	}

	return s.groupRepo.GetGroupStats(ctx, groupID)
}

// GetGroupActivity はグループ活動情報を取得する
func (s *groupService) GetGroupActivity(ctx context.Context, groupID uuid.UUID, days int) (*GroupActivity, error) {
	// 簡易実装
	return &GroupActivity{
		TasksCreated:    0,
		TasksCompleted:  0,
		SchedulesShared: 0,
		ActiveMembers:   0,
	}, nil
}

// === ヘルパーメソッド ===

func (s *groupService) validateCreateGroupInput(input CreateGroupInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	if len(input.Name) > 100 {
		return errors.New("name too long")
	}
	if len(input.Description) > 500 {
		return errors.New("description too long")
	}
	if input.Type != domain.GroupTypeProject && input.Type != domain.GroupTypeSchedule {
		return errors.New("invalid group type")
	}
	return nil
}

func (s *groupService) hasPermissionForAction(role domain.MemberRole, action GroupAction) bool {
	switch action {
	case ActionViewGroup:
		return true // 全メンバーが閲覧可能
	case ActionEditGroup, ActionDeleteGroup:
		return role == domain.RoleOwner || role == domain.RoleAdmin
	case ActionInviteMembers:
		return role == domain.RoleOwner || role == domain.RoleAdmin
	case ActionRemoveMembers, ActionManageRoles:
		return role == domain.RoleOwner || role == domain.RoleAdmin
	case ActionCreateTasks, ActionEditTasks, ActionDeleteTasks:
		return role == domain.RoleOwner || role == domain.RoleAdmin
	case ActionViewTasks, ActionViewSchedules:
		return true // 全メンバーが閲覧可能
	default:
		return false
	}
}

func (s *groupService) enrichMembersWithUserInfo(ctx context.Context, members []*domain.GroupMember) ([]*MemberWithUserInfo, error) {
	if len(members) == 0 {
		return []*MemberWithUserInfo{}, nil
	}

	// ユーザーIDを収集
	userIDs := make([]string, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID.String()
	}

	// ユーザー情報を一括取得
	userInfoMap, err := s.userValidator.GetUsersInfoBatch(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info batch: %w", err)
	}

	// 結果を組み立て
	result := make([]*MemberWithUserInfo, len(members))
	for i, member := range members {
		result[i] = &MemberWithUserInfo{
			Member:   member,
			UserInfo: userInfoMap[member.UserID.String()],
		}
	}

	return result, nil
}

// === 友達招待（Social連携） ===

// InviteFriendsToGroup は友達をグループに招待する
func (s *groupService) InviteFriendsToGroup(ctx context.Context, groupID, inviterID uuid.UUID, friendIDs []uuid.UUID, message string) ([]*GroupInviteResult, error) {
	// 権限チェック
	hasPermission, err := s.CheckPermission(ctx, groupID, inviterID, ActionInviteMembers)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("insufficient permissions")
	}

	results := make([]*GroupInviteResult, len(friendIDs))

	for i, friendID := range friendIDs {
		result := &GroupInviteResult{
			FriendID: friendID,
		}

		// 既にメンバーかチェック
		isMember, err := s.groupRepo.IsMember(ctx, groupID, friendID)
		if err != nil {
			result.Success = false
			result.Error = "メンバーシップの確認に失敗しました"
			results[i] = result
			continue
		}

		if isMember {
			result.Success = false
			result.Error = "既にグループのメンバーです"
			results[i] = result
			continue
		}

		// ユーザー存在確認
		exists, err := s.userValidator.UserExists(ctx, friendID.String())
		if err != nil || !exists {
			result.Success = false
			result.Error = "ユーザーが見つかりません"
			results[i] = result
			continue
		}

		// TODO: Social モジュールとの連携でグループ招待を作成
		// 現在は直接メンバーとして追加
		member := domain.NewGroupMember(groupID, friendID, domain.RoleMember)
		err = s.groupRepo.AddMember(ctx, member)
		if err != nil {
			s.logger.Error("Failed to add member to group",
				logger.Any("groupID", groupID),
				logger.Any("friendID", friendID),
				logger.Error(err))
			result.Success = false
			result.Error = "グループへの追加に失敗しました"
		} else {
			// グループのメンバー数を更新
			group, err := s.groupRepo.GetGroupByID(ctx, groupID)
			if err == nil {
				group.AddMember()
				s.groupRepo.UpdateGroup(ctx, group)
			}

			result.Success = true
			result.Message = "グループに招待しました"
		}

		results[i] = result
	}

	return results, nil
}

// GetAvailableFriends は招待可能な友達一覧を取得する
func (s *groupService) GetAvailableFriends(ctx context.Context, groupID, userID uuid.UUID) ([]*AvailableFriend, error) {
	// メンバーシップ確認
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, errors.New("not a group member")
	}

	// TODO: Social モジュールとの連携で友達一覧を取得
	// 現在は空の配列を返す
	return []*AvailableFriend{}, nil
}
