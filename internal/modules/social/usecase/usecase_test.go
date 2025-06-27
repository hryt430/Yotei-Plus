package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/domain"
	"github.com/hryt430/Yotei+/internal/modules/social/usecase/mocks"
	"github.com/hryt430/Yotei+/pkg/logger"
)

//go:generate mockgen -source=../domain/friendship_repository.go -destination=mocks/mock_friendship_repository.go -package=mocks
//go:generate mockgen -source=../domain/invitation_repository.go -destination=mocks/mock_invitation_repository.go -package=mocks
//go:generate mockgen -source=user_validator.go -destination=mocks/mock_user_validator.go -package=mocks
//go:generate mockgen -source=event_publisher.go -destination=mocks/mock_event_publisher.go -package=mocks
//go:generate mockgen -source=url_gateway.go -destination=mocks/mock_url_gateway.go -package=mocks

func TestSocialService_SendFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		requesterID   uuid.UUID
		addresseeID   uuid.UUID
		message       string
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful friend request",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			message:     "Let's be friends!",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)

				mockFriendshipRepo.EXPECT().
					CreateFriendship(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, friendship *domain.Friendship) {
						assert.Equal(t, domain.FriendshipStatusPending, friendship.Status)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishFriendRequestSent(gomock.Any(), gomock.Any(), "Let's be friends!").
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "self friend request",
			requesterID: uuid.New(),
			addresseeID: func() uuid.UUID { id := uuid.New(); return id }(),
			message:     "",
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "cannot send friend request to yourself",
		},
		{
			name:        "addressee user not found",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			message:     "",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(false, nil)
			},
			expectedError: "addressee user not found",
		},
		{
			name:        "already friends",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			message:     "",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				existingFriendship := &domain.Friendship{
					Status: domain.FriendshipStatusAccepted,
				}
				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(existingFriendship, nil)
			},
			expectedError: "already friends",
		},
		{
			name:        "friend request already pending",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			message:     "",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				existingFriendship := &domain.Friendship{
					Status: domain.FriendshipStatusPending,
				}
				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(existingFriendship, nil)
			},
			expectedError: "friend request already pending",
		},
		{
			name:        "user blocked",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			message:     "",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				existingFriendship := &domain.Friendship{
					Status: domain.FriendshipStatusBlocked,
				}
				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(existingFriendship, nil)
			},
			expectedError: "user is blocked",
		},
		{
			name:        "user validation error",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			message:     "",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(false, errors.New("database error"))
			},
			expectedError: "failed to validate addressee",
		},
		{
			name:        "repository error",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			message:     "",
			setupMocks: func() {
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to check existing friendship",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle self request case
			if tt.name == "self friend request" {
				tt.addresseeID = tt.requesterID
			}

			tt.setupMocks()

			result, err := service.SendFriendRequest(context.Background(), tt.requesterID, tt.addresseeID, tt.message)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.requesterID, result.RequesterID)
				assert.Equal(t, tt.addresseeID, result.AddresseeID)
				assert.Equal(t, domain.FriendshipStatusPending, result.Status)
			}
		})
	}
}

func TestSocialService_AcceptFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		requesterID   uuid.UUID
		addresseeID   uuid.UUID
		setupMocks    func(requesterID, addresseeID uuid.UUID)
		expectedError string
	}{
		{
			name:        "successful accept",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			setupMocks: func(requesterID, addresseeID uuid.UUID) {
				friendship := &domain.Friendship{
					ID:          uuid.New(),
					RequesterID: requesterID,
					AddresseeID: addresseeID,
					Status:      domain.FriendshipStatusPending,
				}

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(friendship, nil)

				mockFriendshipRepo.EXPECT().
					UpdateFriendship(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, f *domain.Friendship) {
						assert.Equal(t, domain.FriendshipStatusAccepted, f.Status)
						assert.NotNil(t, f.AcceptedAt)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishFriendRequestAccepted(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "friend request not found",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			setupMocks: func(requesterID, addresseeID uuid.UUID) {
				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError: "friend request not found",
		},
		{
			name:        "friend request not pending",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			setupMocks: func(requesterID, addresseeID uuid.UUID) {
				friendship := &domain.Friendship{
					Status: domain.FriendshipStatusAccepted,
				}

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(friendship, nil)
			},
			expectedError: "friend request is not pending",
		},
		{
			name:        "wrong addressee",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			setupMocks: func(requesterID, addresseeID uuid.UUID) {
				friendship := &domain.Friendship{
					RequesterID: requesterID,
					AddresseeID: uuid.New(), // Different from test addresseeID
					Status:      domain.FriendshipStatusPending,
				}

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(friendship, nil)
			},
			expectedError: "not authorized to accept this friend request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.requesterID, tt.addresseeID)

			result, err := service.AcceptFriendRequest(context.Background(), tt.requesterID, tt.addresseeID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, domain.FriendshipStatusAccepted, result.Status)
			}
		})
	}
}

func TestSocialService_DeclineFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		requesterID   uuid.UUID
		addresseeID   uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful decline",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			setupMocks: func() {
				friendship := &domain.Friendship{
					ID:          uuid.New(),
					RequesterID: uuid.New(),
					AddresseeID: uuid.New(),
					Status:      domain.FriendshipStatusPending,
				}

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(friendship, nil)

				mockFriendshipRepo.EXPECT().
					DeleteFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishFriendRequestDeclined(gomock.Any(), friendship).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "friend request not found",
			requesterID: uuid.New(),
			addresseeID: uuid.New(),
			setupMocks: func() {
				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError: "friend request not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.DeclineFriendRequest(context.Background(), tt.requesterID, tt.addresseeID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSocialService_RemoveFriend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		userID        uuid.UUID
		friendID      uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:     "successful remove friend",
			userID:   uuid.New(),
			friendID: uuid.New(),
			setupMocks: func() {
				mockFriendshipRepo.EXPECT().
					AreFriends(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockFriendshipRepo.EXPECT().
					DeleteFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishFriendRemoved(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:     "not friends",
			userID:   uuid.New(),
			friendID: uuid.New(),
			setupMocks: func() {
				mockFriendshipRepo.EXPECT().
					AreFriends(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)
			},
			expectedError: "not friends",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.RemoveFriend(context.Background(), tt.userID, tt.friendID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSocialService_BlockUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		userID        uuid.UUID
		targetID      uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:     "new block",
			userID:   uuid.New(),
			targetID: uuid.New(),
			setupMocks: func() {
				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)

				mockFriendshipRepo.EXPECT().
					CreateFriendship(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, friendship *domain.Friendship) {
						assert.Equal(t, domain.FriendshipStatusBlocked, friendship.Status)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishUserBlocked(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:     "update existing friendship",
			userID:   uuid.New(),
			targetID: uuid.New(),
			setupMocks: func() {
				existingFriendship := &domain.Friendship{
					Status: domain.FriendshipStatusAccepted,
				}

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(existingFriendship, nil)

				mockFriendshipRepo.EXPECT().
					UpdateFriendship(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, friendship *domain.Friendship) {
						assert.Equal(t, domain.FriendshipStatusBlocked, friendship.Status)
						assert.NotNil(t, friendship.BlockedAt)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishUserBlocked(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.BlockUser(context.Background(), tt.userID, tt.targetID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSocialService_GetFriends(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		userID        uuid.UUID
		pagination    commonDomain.Pagination
		setupMocks    func(userID uuid.UUID)
		expectedError string
		expectedCount int
	}{
		{
			name:       "successful get friends",
			userID:     uuid.New(),
			pagination: commonDomain.Pagination{Page: 1, PageSize: 10},
			setupMocks: func(userID uuid.UUID) {
				friendID := uuid.New()
				friendships := []*domain.Friendship{
					{
						RequesterID: userID,
						AddresseeID: friendID,
						Status:      domain.FriendshipStatusAccepted,
					},
				}

				userInfo := &commonDomain.UserInfo{
					ID:       friendID.String(),
					Username: "friend_user",
					Email:    "friend@example.com",
				}

				mockFriendshipRepo.EXPECT().
					GetFriends(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(friendships, nil)

				mockUserValidator.EXPECT().
					GetUsersInfoBatch(gomock.Any(), []string{friendID.String()}).
					Return(map[string]*commonDomain.UserInfo{
						friendID.String(): userInfo,
					}, nil)
			},
			expectedError: "",
			expectedCount: 1,
		},
		{
			name:       "empty result",
			userID:     uuid.New(),
			pagination: commonDomain.Pagination{Page: 1, PageSize: 10},
			setupMocks: func(userID uuid.UUID) {
				mockFriendshipRepo.EXPECT().
					GetFriends(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*domain.Friendship{}, nil)
			},
			expectedError: "",
			expectedCount: 0,
		},
		{
			name:       "repository error",
			userID:     uuid.New(),
			pagination: commonDomain.Pagination{Page: 1, PageSize: 10},
			setupMocks: func(userID uuid.UUID) {
				mockFriendshipRepo.EXPECT().
					GetFriends(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get friends",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.userID)

			result, err := service.GetFriends(context.Background(), tt.userID, tt.pagination)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestSocialService_CreateInvitation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		input         CreateInvitationInput
		setupMocks    func()
		expectedError string
	}{
		{
			name: "successful friend invitation creation",
			input: CreateInvitationInput{
				Type:         domain.InvitationTypeFriend,
				Method:       domain.MethodCode,
				InviterID:    uuid.New(),
				Message:      "Join us!",
				ExpiresHours: 24,
				InviteeEmail: &[]string{"invitee@example.com"}[0],
			},
			setupMocks: func() {
				mockInvitationRepo.EXPECT().
					CreateInvitation(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, invitation *domain.Invitation) {
						assert.Equal(t, domain.InvitationTypeFriend, invitation.Type)
						assert.Equal(t, domain.MethodCode, invitation.Method)
						assert.Equal(t, domain.InvitationStatusPending, invitation.Status)
						assert.NotEmpty(t, invitation.Code)
					}).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishInvitationCreated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "group invitation with target",
			input: CreateInvitationInput{
				Type:         domain.InvitationTypeGroup,
				Method:       domain.MethodURL,
				InviterID:    uuid.New(),
				Message:      "Join our group!",
				ExpiresHours: 48,
				TargetID:     &[]uuid.UUID{uuid.New()}[0],
			},
			setupMocks: func() {
				mockInvitationRepo.EXPECT().
					CreateInvitation(gomock.Any(), gomock.Any()).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishInvitationCreated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.CreateInvitation(context.Background(), tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Type, result.Type)
				assert.Equal(t, tt.input.InviterID, result.InviterID)
			}
		})
	}
}

func TestSocialService_AcceptInvitation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		code          string
		userID        uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:   "successful friend invitation accept",
			code:   "TEST123456",
			userID: uuid.New(),
			setupMocks: func() {
				invitation := &domain.Invitation{
					ID:        uuid.New(),
					Type:      domain.InvitationTypeFriend,
					InviterID: uuid.New(),
					Status:    domain.InvitationStatusPending,
					ExpiresAt: time.Now().Add(time.Hour),
				}

				mockInvitationRepo.EXPECT().
					GetInvitationByCode(gomock.Any(), "TEST123456").
					Return(invitation, nil)

				mockInvitationRepo.EXPECT().
					UpdateInvitation(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, inv *domain.Invitation) {
						assert.Equal(t, domain.InvitationStatusAccepted, inv.Status)
					}).
					Return(nil)

				// Friend request creation expectations
				mockUserValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)

				mockFriendshipRepo.EXPECT().
					CreateFriendship(gomock.Any(), gomock.Any()).
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishFriendRequestSent(gomock.Any(), gomock.Any(), "招待から").
					Return(nil)

				mockEventPublisher.EXPECT().
					PublishInvitationAccepted(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "invitation not found",
			code:   "INVALID123",
			userID: uuid.New(),
			setupMocks: func() {
				mockInvitationRepo.EXPECT().
					GetInvitationByCode(gomock.Any(), "INVALID123").
					Return(nil, nil)
			},
			expectedError: "invitation not found",
		},
		{
			name:   "expired invitation",
			code:   "EXPIRED123",
			userID: uuid.New(),
			setupMocks: func() {
				invitation := &domain.Invitation{
					ID:        uuid.New(),
					Type:      domain.InvitationTypeFriend,
					Status:    domain.InvitationStatusPending,
					ExpiresAt: time.Now().Add(-time.Hour), // Expired
				}

				mockInvitationRepo.EXPECT().
					GetInvitationByCode(gomock.Any(), "EXPIRED123").
					Return(invitation, nil)
			},
			expectedError: "invitation is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.AcceptInvitation(context.Background(), tt.code, tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
			}
		})
	}
}

func TestSocialService_GetRelationship(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name             string
		userID           uuid.UUID
		targetID         uuid.UUID
		setupMocks       func()
		expectedError    string
		expectedFriend   bool
		expectedBlocked  bool
		expectedSent     bool
		expectedReceived bool
	}{
		{
			name:     "friends relationship",
			userID:   uuid.New(),
			targetID: uuid.New(),
			setupMocks: func() {
				mockFriendshipRepo.EXPECT().
					AreFriends(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockFriendshipRepo.EXPECT().
					IsBlocked(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError:    "",
			expectedFriend:   true,
			expectedBlocked:  false,
			expectedSent:     false,
			expectedReceived: false,
		},
		{
			name:     "pending request sent",
			userID:   uuid.New(),
			targetID: uuid.New(),
			setupMocks: func() {
				userID := uuid.New()
				targetID := uuid.New()
				friendship := &domain.Friendship{
					RequesterID: userID,
					AddresseeID: targetID,
					Status:      domain.FriendshipStatusPending,
				}

				mockFriendshipRepo.EXPECT().
					AreFriends(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)

				mockFriendshipRepo.EXPECT().
					IsBlocked(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)

				mockFriendshipRepo.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(friendship, nil)
			},
			expectedError:    "",
			expectedFriend:   false,
			expectedBlocked:  false,
			expectedSent:     true,
			expectedReceived: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.GetRelationship(context.Background(), tt.userID, tt.targetID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedFriend, result.IsFriend)
				assert.Equal(t, tt.expectedBlocked, result.IsBlocked)
			}
		})
	}
}

func TestSocialService_GenerateInviteURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendshipRepo := mocks.NewMockFriendshipRepository(ctrl)
	mockInvitationRepo := mocks.NewMockInvitationRepository(ctrl)
	mockUserValidator := mocks.NewMockUserValidator(ctrl)
	mockEventPublisher := mocks.NewMockSocialEventPublisher(ctrl)
	mockURLGateway := mocks.NewMockURLGateway(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})

	service := NewSocialServiceImpl(
		mockFriendshipRepo,
		mockInvitationRepo,
		mockUserValidator,
		mockEventPublisher,
		mockURLGateway,
		&mockLogger,
	)

	tests := []struct {
		name          string
		invitationID  uuid.UUID
		setupMocks    func()
		expectedError string
		expectedURL   string
	}{
		{
			name:         "successful URL generation",
			invitationID: uuid.New(),
			setupMocks: func() {
				invitation := &domain.Invitation{
					ID:   uuid.New(),
					Code: "TEST123456",
				}
				expectedURL := "https://example.com/invite/TEST123456"

				mockInvitationRepo.EXPECT().
					GetInvitationByID(gomock.Any(), gomock.Any()).
					Return(invitation, nil)

				mockURLGateway.EXPECT().
					GenerateInviteURL(gomock.Any(), gomock.Any(), invitation.Code).
					Return(expectedURL, nil)
			},
			expectedError: "",
			expectedURL:   "https://example.com/invite/TEST123456",
		},
		{
			name:         "invitation not found",
			invitationID: uuid.New(),
			setupMocks: func() {
				mockInvitationRepo.EXPECT().
					GetInvitationByID(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError: "invitation not found",
			expectedURL:   "",
		},
		{
			name:         "invitation without code",
			invitationID: uuid.New(),
			setupMocks: func() {
				invitation := &domain.Invitation{
					ID:   uuid.New(),
					Code: "", // No code
				}

				mockInvitationRepo.EXPECT().
					GetInvitationByID(gomock.Any(), gomock.Any()).
					Return(invitation, nil)
			},
			expectedError: "invitation does not have a code",
			expectedURL:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.GenerateInviteURL(context.Background(), tt.invitationID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectedURL != "" {
					assert.Equal(t, tt.expectedURL, result)
				}
			}
		})
	}
}
