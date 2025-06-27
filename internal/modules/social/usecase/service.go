package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// SocialServiceImpl はSocialServiceの実装
type SocialServiceImpl struct {
	friendshipRepo FriendshipRepository
	invitationRepo InvitationRepository
	userValidator  commonDomain.UserValidator
	eventPublisher SocialEventPublisher
	urlGateway     URLGateway
	logger         *logger.Logger
}

// SocialEventPublisher はソーシャルイベント発行のインターフェース
type SocialEventPublisher interface {
	PublishFriendRequestSent(ctx context.Context, friendship *domain.Friendship, message string) error
	PublishFriendRequestAccepted(ctx context.Context, friendship *domain.Friendship) error
	PublishFriendRequestDeclined(ctx context.Context, friendship *domain.Friendship) error
	PublishFriendRemoved(ctx context.Context, userID, friendID uuid.UUID) error
	PublishUserBlocked(ctx context.Context, userID, targetID uuid.UUID) error
	PublishInvitationCreated(ctx context.Context, invitation *domain.Invitation) error
	PublishInvitationAccepted(ctx context.Context, invitation *domain.Invitation) error
	PublishInvitationDeclined(ctx context.Context, invitation *domain.Invitation) error
}

// URLGateway はURL生成のインターフェース
type URLGateway interface {
	GenerateInviteURL(ctx context.Context, invitationID uuid.UUID, code string) (string, error)
}

// NewSocialServiceImpl は新しいSocialServiceImplを作成する
func NewSocialServiceImpl(
	friendshipRepo FriendshipRepository,
	invitationRepo InvitationRepository,
	userValidator commonDomain.UserValidator,
	eventPublisher SocialEventPublisher,
	urlGateway URLGateway,
	logger *logger.Logger,
) SocialService {
	return &SocialServiceImpl{
		friendshipRepo: friendshipRepo,
		invitationRepo: invitationRepo,
		userValidator:  userValidator,
		eventPublisher: eventPublisher,
		urlGateway:     urlGateway,
		logger:         logger,
	}
}

// === 友達関係管理 ===

// SendFriendRequest は友達申請を送信する
func (s *SocialServiceImpl) SendFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID, message string) (*domain.Friendship, error) {
	// 自分自身への申請チェック
	if requesterID == addresseeID {
		return nil, errors.New("cannot send friend request to yourself")
	}

	// ユーザー存在確認
	exists, err := s.userValidator.UserExists(ctx, addresseeID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to validate addressee: %w", err)
	}
	if !exists {
		return nil, errors.New("addressee user not found")
	}

	// 既存の友達関係をチェック
	existingFriendship, err := s.friendshipRepo.GetFriendship(ctx, requesterID, addresseeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing friendship: %w", err)
	}

	if existingFriendship != nil {
		switch existingFriendship.Status {
		case domain.FriendshipStatusAccepted:
			return nil, errors.New("already friends")
		case domain.FriendshipStatusPending:
			return nil, errors.New("friend request already pending")
		case domain.FriendshipStatusBlocked:
			return nil, errors.New("user is blocked")
		}
	}

	// 友達申請作成
	friendship := domain.NewFriendship(requesterID, addresseeID)

	if err := s.friendshipRepo.CreateFriendship(ctx, friendship); err != nil {
		s.logger.Error("Failed to create friendship",
			logger.Any("requesterID", requesterID),
			logger.Any("addresseeID", addresseeID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to create friendship: %w", err)
	}

	// イベント発行
	if err := s.eventPublisher.PublishFriendRequestSent(ctx, friendship, message); err != nil {
		s.logger.Error("Failed to publish friend request sent event", logger.Error(err))
		// イベント発行失敗は非致命的
	}

	s.logger.Info("Friend request sent successfully",
		logger.Any("requesterID", requesterID),
		logger.Any("addresseeID", addresseeID))

	return friendship, nil
}

// AcceptFriendRequest は友達申請を承認する
func (s *SocialServiceImpl) AcceptFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID) (*domain.Friendship, error) {
	friendship, err := s.friendshipRepo.GetFriendship(ctx, requesterID, addresseeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friendship: %w", err)
	}

	if friendship == nil {
		return nil, errors.New("friend request not found")
	}

	if friendship.Status != domain.FriendshipStatusPending {
		return nil, errors.New("friend request is not pending")
	}

	// addresseeIDが申請の受信者であることを確認
	if friendship.AddresseeID != addresseeID {
		return nil, errors.New("not authorized to accept this friend request")
	}

	// 友達申請を承認
	friendship.Accept()

	if err := s.friendshipRepo.UpdateFriendship(ctx, friendship); err != nil {
		s.logger.Error("Failed to update friendship",
			logger.Any("friendshipID", friendship.ID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to update friendship: %w", err)
	}

	// イベント発行
	if err := s.eventPublisher.PublishFriendRequestAccepted(ctx, friendship); err != nil {
		s.logger.Error("Failed to publish friend request accepted event", logger.Error(err))
	}

	s.logger.Info("Friend request accepted successfully",
		logger.Any("friendshipID", friendship.ID))

	return friendship, nil
}

// DeclineFriendRequest は友達申請を拒否する
func (s *SocialServiceImpl) DeclineFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	friendship, err := s.friendshipRepo.GetFriendship(ctx, requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("failed to get friendship: %w", err)
	}

	if friendship == nil {
		return errors.New("friend request not found")
	}

	if friendship.Status != domain.FriendshipStatusPending {
		return errors.New("friend request is not pending")
	}

	// 友達申請を削除（拒否）
	if err := s.friendshipRepo.DeleteFriendship(ctx, requesterID, addresseeID); err != nil {
		s.logger.Error("Failed to delete friendship",
			logger.Any("requesterID", requesterID),
			logger.Any("addresseeID", addresseeID),
			logger.Error(err))
		return fmt.Errorf("failed to delete friendship: %w", err)
	}

	// イベント発行
	if err := s.eventPublisher.PublishFriendRequestDeclined(ctx, friendship); err != nil {
		s.logger.Error("Failed to publish friend request declined event", logger.Error(err))
	}

	s.logger.Info("Friend request declined successfully",
		logger.Any("requesterID", requesterID),
		logger.Any("addresseeID", addresseeID))

	return nil
}

// RemoveFriend は友達を削除する
func (s *SocialServiceImpl) RemoveFriend(ctx context.Context, userID, friendID uuid.UUID) error {
	// 友達関係が存在するかチェック
	areFriends, err := s.friendshipRepo.AreFriends(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to check friendship: %w", err)
	}

	if !areFriends {
		return errors.New("not friends")
	}

	// 友達関係を削除
	if err := s.friendshipRepo.DeleteFriendship(ctx, userID, friendID); err != nil {
		s.logger.Error("Failed to remove friend",
			logger.Any("userID", userID),
			logger.Any("friendID", friendID),
			logger.Error(err))
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	// イベント発行
	if err := s.eventPublisher.PublishFriendRemoved(ctx, userID, friendID); err != nil {
		s.logger.Error("Failed to publish friend removed event", logger.Error(err))
	}

	s.logger.Info("Friend removed successfully",
		logger.Any("userID", userID),
		logger.Any("friendID", friendID))

	return nil
}

// BlockUser はユーザーをブロックする
func (s *SocialServiceImpl) BlockUser(ctx context.Context, userID, targetID uuid.UUID) error {
	// 既存の関係をチェック
	existingFriendship, err := s.friendshipRepo.GetFriendship(ctx, userID, targetID)
	if err != nil {
		return fmt.Errorf("failed to check existing relationship: %w", err)
	}

	if existingFriendship != nil {
		// 既存の関係をブロック状態に更新
		existingFriendship.Block()
		if err := s.friendshipRepo.UpdateFriendship(ctx, existingFriendship); err != nil {
			return fmt.Errorf("failed to update friendship to blocked: %w", err)
		}
	} else {
		// 新規ブロック関係を作成
		friendship := domain.NewFriendship(userID, targetID)
		friendship.Block()
		if err := s.friendshipRepo.CreateFriendship(ctx, friendship); err != nil {
			return fmt.Errorf("failed to create blocked relationship: %w", err)
		}
	}

	// イベント発行
	if err := s.eventPublisher.PublishUserBlocked(ctx, userID, targetID); err != nil {
		s.logger.Error("Failed to publish user blocked event", logger.Error(err))
	}

	s.logger.Info("User blocked successfully",
		logger.Any("userID", userID),
		logger.Any("targetID", targetID))

	return nil
}

// UnblockUser はブロックを解除する
func (s *SocialServiceImpl) UnblockUser(ctx context.Context, userID, targetID uuid.UUID) error {
	// ブロック関係を削除
	if err := s.friendshipRepo.DeleteFriendship(ctx, userID, targetID); err != nil {
		s.logger.Error("Failed to unblock user",
			logger.Any("userID", userID),
			logger.Any("targetID", targetID),
			logger.Error(err))
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	s.logger.Info("User unblocked successfully",
		logger.Any("userID", userID),
		logger.Any("targetID", targetID))

	return nil
}

// === 友達一覧・検索 ===

// GetFriends は友達一覧を取得する
func (s *SocialServiceImpl) GetFriends(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*FriendWithUserInfo, error) {
	friendships, err := s.friendshipRepo.GetFriends(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	if len(friendships) == 0 {
		return []*FriendWithUserInfo{}, nil
	}

	// ユーザー情報を一括取得
	userIDs := make([]string, 0, len(friendships))
	for _, friendship := range friendships {
		if friendship.RequesterID == userID {
			userIDs = append(userIDs, friendship.AddresseeID.String())
		} else {
			userIDs = append(userIDs, friendship.RequesterID.String())
		}
	}

	userInfoMap, err := s.userValidator.GetUsersInfoBatch(ctx, userIDs)
	if err != nil {
		s.logger.Error("Failed to get user info batch", logger.Error(err))
		userInfoMap = make(map[string]*commonDomain.UserInfo)
	}

	// 結果を組み立て
	result := make([]*FriendWithUserInfo, len(friendships))
	for i, friendship := range friendships {
		friendID := friendship.AddresseeID
		if friendship.RequesterID != userID {
			friendID = friendship.RequesterID
		}

		result[i] = &FriendWithUserInfo{
			Friendship: friendship,
			UserInfo:   userInfoMap[friendID.String()],
		}
	}

	return result, nil
}

// GetPendingRequests は受信した友達申請を取得する
func (s *SocialServiceImpl) GetPendingRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*FriendshipWithUserInfo, error) {
	friendships, err := s.friendshipRepo.GetPendingRequests(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}

	if len(friendships) == 0 {
		return []*FriendshipWithUserInfo{}, nil
	}

	// 申請者のユーザー情報を一括取得
	userIDs := make([]string, len(friendships))
	for i, friendship := range friendships {
		userIDs[i] = friendship.RequesterID.String()
	}

	userInfoMap, err := s.userValidator.GetUsersInfoBatch(ctx, userIDs)
	if err != nil {
		s.logger.Error("Failed to get user info batch", logger.Error(err))
		userInfoMap = make(map[string]*commonDomain.UserInfo)
	}

	// 結果を組み立て
	result := make([]*FriendshipWithUserInfo, len(friendships))
	for i, friendship := range friendships {
		result[i] = &FriendshipWithUserInfo{
			Friendship: friendship,
			UserInfo:   userInfoMap[friendship.RequesterID.String()],
		}
	}

	return result, nil
}

// GetSentRequests は送信した友達申請を取得する
func (s *SocialServiceImpl) GetSentRequests(ctx context.Context, userID uuid.UUID, pagination commonDomain.Pagination) ([]*FriendshipWithUserInfo, error) {
	friendships, err := s.friendshipRepo.GetSentRequests(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent requests: %w", err)
	}

	if len(friendships) == 0 {
		return []*FriendshipWithUserInfo{}, nil
	}

	// 申請先のユーザー情報を一括取得
	userIDs := make([]string, len(friendships))
	for i, friendship := range friendships {
		userIDs[i] = friendship.AddresseeID.String()
	}

	userInfoMap, err := s.userValidator.GetUsersInfoBatch(ctx, userIDs)
	if err != nil {
		s.logger.Error("Failed to get user info batch", logger.Error(err))
		userInfoMap = make(map[string]*commonDomain.UserInfo)
	}

	// 結果を組み立て
	result := make([]*FriendshipWithUserInfo, len(friendships))
	for i, friendship := range friendships {
		result[i] = &FriendshipWithUserInfo{
			Friendship: friendship,
			UserInfo:   userInfoMap[friendship.AddresseeID.String()],
		}
	}

	return result, nil
}

// GetMutualFriends は共通の友達を取得する
func (s *SocialServiceImpl) GetMutualFriends(ctx context.Context, userID, targetID uuid.UUID) ([]*FriendWithUserInfo, error) {
	friendships, err := s.friendshipRepo.GetMutualFriends(ctx, userID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mutual friends: %w", err)
	}

	if len(friendships) == 0 {
		return []*FriendWithUserInfo{}, nil
	}

	// 共通の友達のユーザー情報を一括取得
	userIDs := make([]string, 0, len(friendships))
	for _, friendship := range friendships {
		if friendship.RequesterID != userID && friendship.RequesterID != targetID {
			userIDs = append(userIDs, friendship.RequesterID.String())
		}
		if friendship.AddresseeID != userID && friendship.AddresseeID != targetID {
			userIDs = append(userIDs, friendship.AddresseeID.String())
		}
	}

	userInfoMap, err := s.userValidator.GetUsersInfoBatch(ctx, userIDs)
	if err != nil {
		s.logger.Error("Failed to get user info batch", logger.Error(err))
		userInfoMap = make(map[string]*commonDomain.UserInfo)
	}

	// 結果を組み立て
	result := make([]*FriendWithUserInfo, 0, len(friendships))
	for _, friendship := range friendships {
		var friendID uuid.UUID
		if friendship.RequesterID != userID && friendship.RequesterID != targetID {
			friendID = friendship.RequesterID
		} else if friendship.AddresseeID != userID && friendship.AddresseeID != targetID {
			friendID = friendship.AddresseeID
		} else {
			continue
		}

		result = append(result, &FriendWithUserInfo{
			Friendship: friendship,
			UserInfo:   userInfoMap[friendID.String()],
		})
	}

	return result, nil
}

// === 招待管理 ===

// CreateInvitation は招待を作成する
func (s *SocialServiceImpl) CreateInvitation(ctx context.Context, input CreateInvitationInput) (*domain.Invitation, error) {
	// 招待作成
	invitation := domain.NewInvitation(input.Type, input.Method, input.InviterID, input.Message, input.ExpiresHours)

	// ターゲット設定
	if input.TargetID != nil {
		invitation.SetTarget(*input.TargetID)
	}

	// 被招待者情報設定
	if input.InviteeEmail != nil {
		inviteeInfo := domain.InviteeInfo{
			Email: *input.InviteeEmail,
		}
		invitation.SetInviteeInfo(inviteeInfo)
	}

	// データベースに保存
	if err := s.invitationRepo.CreateInvitation(ctx, invitation); err != nil {
		s.logger.Error("Failed to create invitation",
			logger.Any("invitation", invitation),
			logger.Error(err))
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// イベント発行
	if err := s.eventPublisher.PublishInvitationCreated(ctx, invitation); err != nil {
		s.logger.Error("Failed to publish invitation created event", logger.Error(err))
	}

	s.logger.Info("Invitation created successfully",
		logger.Any("invitationID", invitation.ID))

	return invitation, nil
}

// GetInvitation は招待詳細を取得する
func (s *SocialServiceImpl) GetInvitation(ctx context.Context, invitationID uuid.UUID) (*domain.Invitation, error) {
	return s.invitationRepo.GetInvitationByID(ctx, invitationID)
}

// GetInvitationByCode は招待コードから招待を取得する
func (s *SocialServiceImpl) GetInvitationByCode(ctx context.Context, code string) (*domain.Invitation, error) {
	return s.invitationRepo.GetInvitationByCode(ctx, code)
}

// AcceptInvitation は招待を受諾する
func (s *SocialServiceImpl) AcceptInvitation(ctx context.Context, code string, userID uuid.UUID) (*InvitationResult, error) {
	invitation, err := s.invitationRepo.GetInvitationByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	if !invitation.IsValid() {
		return nil, errors.New("invitation is not valid")
	}

	// 招待を受諾
	if err := invitation.Accept(); err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	invitation.SetInvitee(userID)

	if err := s.invitationRepo.UpdateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// 招待タイプに応じた処理
	result := &InvitationResult{
		Success: true,
		Message: "招待を受諾しました",
	}

	switch invitation.Type {
	case domain.InvitationTypeFriend:
		// 友達関係を作成
		friendship, err := s.SendFriendRequest(ctx, invitation.InviterID, userID, "招待から")
		if err != nil {
			// 既に友達の場合などは警告レベル
			s.logger.Warn("Failed to create friendship from invitation", logger.Error(err))
		} else {
			result.Friendship = friendship
		}
	case domain.InvitationTypeGroup:
		// グループメンバー追加（グループモジュールとの連携が必要）
		result.Message = "グループ招待を受諾しました"
	}

	// イベント発行
	if err := s.eventPublisher.PublishInvitationAccepted(ctx, invitation); err != nil {
		s.logger.Error("Failed to publish invitation accepted event", logger.Error(err))
	}

	return result, nil
}

// DeclineInvitation は招待を拒否する
func (s *SocialServiceImpl) DeclineInvitation(ctx context.Context, invitationID, userID uuid.UUID) error {
	invitation, err := s.invitationRepo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation: %w", err)
	}

	if invitation == nil {
		return errors.New("invitation not found")
	}

	// 権限チェック
	if invitation.InviteeID == nil || *invitation.InviteeID != userID {
		return errors.New("not authorized to decline this invitation")
	}

	if err := invitation.Decline(); err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
	}

	if err := s.invitationRepo.UpdateInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// イベント発行
	if err := s.eventPublisher.PublishInvitationDeclined(ctx, invitation); err != nil {
		s.logger.Error("Failed to publish invitation declined event", logger.Error(err))
	}

	return nil
}

// CancelInvitation は招待をキャンセルする
func (s *SocialServiceImpl) CancelInvitation(ctx context.Context, invitationID, inviterID uuid.UUID) error {
	invitation, err := s.invitationRepo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation: %w", err)
	}

	if invitation == nil {
		return errors.New("invitation not found")
	}

	// 権限チェック
	if invitation.InviterID != inviterID {
		return errors.New("not authorized to cancel this invitation")
	}

	if err := invitation.Cancel(); err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	if err := s.invitationRepo.UpdateInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	return nil
}

// GetSentInvitations は送信した招待一覧を取得する
func (s *SocialServiceImpl) GetSentInvitations(ctx context.Context, inviterID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error) {
	return s.invitationRepo.GetSentInvitations(ctx, inviterID, pagination)
}

// GetReceivedInvitations は受信した招待一覧を取得する
func (s *SocialServiceImpl) GetReceivedInvitations(ctx context.Context, inviteeID uuid.UUID, pagination commonDomain.Pagination) ([]*domain.Invitation, error) {
	return s.invitationRepo.GetReceivedInvitations(ctx, inviteeID, pagination)
}

// GenerateInviteURL は招待URLを生成する
func (s *SocialServiceImpl) GenerateInviteURL(ctx context.Context, invitationID uuid.UUID) (string, error) {
	invitation, err := s.invitationRepo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return "", fmt.Errorf("failed to get invitation: %w", err)
	}

	if invitation == nil {
		return "", errors.New("invitation not found")
	}

	if invitation.Code == "" {
		return "", errors.New("invitation does not have a code")
	}

	return s.urlGateway.GenerateInviteURL(ctx, invitationID, invitation.Code)
}

// ValidateInviteCode は招待コードの妥当性を確認する
func (s *SocialServiceImpl) ValidateInviteCode(ctx context.Context, code string) (*domain.Invitation, error) {
	return s.invitationRepo.GetInvitationByCode(ctx, code)
}

// GetRelationship はユーザー間の関係を取得する
func (s *SocialServiceImpl) GetRelationship(ctx context.Context, userID, targetID uuid.UUID) (*UserRelationship, error) {
	relationship := &UserRelationship{}

	// 友達関係をチェック
	areFriends, err := s.friendshipRepo.AreFriends(ctx, userID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to check friendship: %w", err)
	}
	relationship.IsFriend = areFriends

	// ブロック関係をチェック
	isBlocked, err := s.friendshipRepo.IsBlocked(ctx, userID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to check block status: %w", err)
	}
	relationship.IsBlocked = isBlocked

	// 申請状況をチェック
	friendship, err := s.friendshipRepo.GetFriendship(ctx, userID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friendship: %w", err)
	}

	if friendship != nil && friendship.Status == domain.FriendshipStatusPending {
		if friendship.RequesterID == userID {
			relationship.RequestSent = true
		} else {
			relationship.RequestReceived = true
		}
	}

	return relationship, nil
}
