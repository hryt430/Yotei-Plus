package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/domain"
	"github.com/hryt430/Yotei+/internal/modules/group/usecase/mocks"
	"github.com/hryt430/Yotei+/pkg/logger"
)

//go:generate mockgen -source=../domain/group_repository.go -destination=mocks/mock_group_repository.go -package=mocks
//go:generate mockgen -source=user_validator.go -destination=mocks/mock_user_validator.go -package=mocks

func TestGroupService_CreateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		input         CreateGroupInput
		setupMocks    func()
		expectedError string
	}{
		{
			name: "successful group creation",
			input: CreateGroupInput{
				Name:        "Test Group",
				Description: "Test Description",
				Type:        domain.GroupTypeProject,
				OwnerID:     uuid.New(),
				Settings: domain.GroupSettings{
					IsPublic:            true,
					AllowMemberInvite:   true,
					RequireApproval:     false,
					EnableNotifications: true,
				},
			},
			setupMocks: func() {
				mockValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					CreateGroup(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, group *domain.Group) {
						assert.Equal(t, "Test Group", group.Name)
						assert.Equal(t, "Test Description", group.Description)
						assert.Equal(t, domain.GroupTypeProject, group.Type)
						assert.Equal(t, 1, group.MemberCount)
						assert.Equal(t, 2, group.Version)
						assert.NotEqual(t, uuid.Nil, group.ID)
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name: "empty name",
			input: CreateGroupInput{
				Name:        "",
				Description: "Test",
				Type:        domain.GroupTypeProject,
				OwnerID:     uuid.New(),
			},
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "name is required",
		},
		{
			name: "name too long",
			input: CreateGroupInput{
				Name:        string(make([]byte, 101)), // 101 characters
				Description: "Test",
				Type:        domain.GroupTypeProject,
				OwnerID:     uuid.New(),
			},
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "name too long",
		},
		{
			name: "description too long",
			input: CreateGroupInput{
				Name:        "Test",
				Description: string(make([]byte, 501)), // 501 characters
				Type:        domain.GroupTypeProject,
				OwnerID:     uuid.New(),
			},
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "description too long",
		},
		{
			name: "invalid group type",
			input: CreateGroupInput{
				Name:        "Test",
				Description: "Test",
				Type:        domain.GroupType("INVALID"),
				OwnerID:     uuid.New(),
			},
			setupMocks: func() {
				// No mocks needed - validation fails early
			},
			expectedError: "invalid group type",
		},
		{
			name: "owner not found",
			input: CreateGroupInput{
				Name:        "Test Group",
				Description: "Test Description",
				Type:        domain.GroupTypeProject,
				OwnerID:     uuid.New(),
			},
			setupMocks: func() {
				mockValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(false, nil)
			},
			expectedError: "owner not found",
		},
		{
			name: "user validation error",
			input: CreateGroupInput{
				Name:        "Test Group",
				Description: "Test Description",
				Type:        domain.GroupTypeProject,
				OwnerID:     uuid.New(),
			},
			setupMocks: func() {
				mockValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(false, errors.New("validation error"))
			},
			expectedError: "failed to validate owner",
		},
		{
			name: "repository error",
			input: CreateGroupInput{
				Name:        "Test Group",
				Description: "Test Description",
				Type:        domain.GroupTypeProject,
				OwnerID:     uuid.New(),
			},
			setupMocks: func() {
				mockValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					CreateGroup(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "failed to create group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.CreateGroup(context.Background(), tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, tt.input.Type, result.Type)
			}
		})
	}
}

func TestGroupService_GetGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		groupID       uuid.UUID
		requesterID   uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful get group as member",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				group := &domain.Group{
					ID:      uuid.New(),
					Name:    "Test Group",
					OwnerID: uuid.New(),
					Settings: domain.GroupSettings{
						IsPublic: false,
					},
				}
				members := []*domain.GroupMember{
					{
						ID:      uuid.New(),
						GroupID: uuid.New(),
						UserID:  uuid.New(),
						Role:    domain.RoleMember,
					},
				}
				userInfo := &commonDomain.UserInfo{
					ID:       uuid.New().String(),
					Username: "testuser",
					Email:    "test@example.com",
				}

				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleMember, nil)

				mockRepo.EXPECT().
					ListMembers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(members, nil)

				mockValidator.EXPECT().
					GetUsersInfoBatch(gomock.Any(), gomock.Any()).
					Return(map[string]*commonDomain.UserInfo{
						userInfo.ID: userInfo,
					}, nil)
			},
			expectedError: "",
		},
		{
			name:        "group not found",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError: "group not found",
		},
		{
			name:        "access denied to private group",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				group := &domain.Group{
					ID:      uuid.New(),
					Name:    "Private Group",
					OwnerID: uuid.New(),
					Settings: domain.GroupSettings{
						IsPublic: false,
					},
				}

				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)
			},
			expectedError: "access denied",
		},
		{
			name:        "public group access",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				group := &domain.Group{
					ID:      uuid.New(),
					Name:    "Public Group",
					OwnerID: uuid.New(),
					Settings: domain.GroupSettings{
						IsPublic: true,
					},
				}

				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)

				mockRepo.EXPECT().
					ListMembers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*domain.GroupMember{}, nil)

				mockValidator.EXPECT().
					GetUsersInfoBatch(gomock.Any(), []string{}).
					Return(map[string]*commonDomain.UserInfo{}, nil).
					AnyTimes()
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.GetGroup(context.Background(), tt.groupID, tt.requesterID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Group)
			}
		})
	}
}

func TestGroupService_UpdateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		groupID       uuid.UUID
		input         UpdateGroupInput
		requesterID   uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:    "successful update by owner",
			groupID: uuid.New(),
			input: UpdateGroupInput{
				Name:        &[]string{"Updated Group"}[0],
				Description: &[]string{"Updated Description"}[0],
				Settings: &domain.GroupSettings{
					IsPublic:            true,
					AllowMemberInvite:   false,
					RequireApproval:     true,
					EnableNotifications: false,
				},
			},
			requesterID: uuid.New(),
			setupMocks: func() {
				group := &domain.Group{
					ID:          uuid.New(),
					Name:        "Old Name",
					Description: "Old Description",
					OwnerID:     uuid.New(),
					Settings: domain.GroupSettings{
						IsPublic: false,
					},
					Version: 1,
				}

				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleOwner, nil)

				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					UpdateGroup(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, g *domain.Group) {
						assert.Equal(t, "Updated Group", g.Name)
						assert.Equal(t, "Updated Description", g.Description)
						assert.Equal(t, 2, g.Version) // Should be incremented
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:    "insufficient permissions",
			groupID: uuid.New(),
			input: UpdateGroupInput{
				Name: &[]string{"New Name"}[0],
			},
			requesterID: uuid.New(),
			setupMocks: func() {
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleMember, nil)
			},
			expectedError: "insufficient permissions",
		},
		{
			name:        "no changes",
			groupID:     uuid.New(),
			input:       UpdateGroupInput{}, // No changes
			requesterID: uuid.New(),
			setupMocks: func() {
				group := &domain.Group{
					ID:      uuid.New(),
					Name:    "Test Group",
					OwnerID: uuid.New(),
					Version: 1,
				}

				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleOwner, nil)

				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				// No UpdateGroup call expected since there are no changes
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.UpdateGroup(context.Background(), tt.groupID, tt.input, tt.requesterID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGroupService_DeleteGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		groupID       uuid.UUID
		requesterID   uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful delete by owner",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				ownerID := uuid.New()
				group := &domain.Group{
					ID:      uuid.New(),
					OwnerID: ownerID,
				}

				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					DeleteGroup(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "only owner can delete",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				group := &domain.Group{
					ID:      uuid.New(),
					OwnerID: uuid.New(), // Different from requester
				}

				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)
			},
			expectedError: "only owner can delete group",
		},
		{
			name:        "group not found",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError: "group not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle successful delete case
			if tt.name == "successful delete by owner" {
				// Update the mock to use the same requesterID as ownerID
				group := &domain.Group{
					ID:      tt.groupID,
					OwnerID: tt.requesterID,
				}
				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), tt.groupID).
					Return(group, nil)
				mockRepo.EXPECT().
					DeleteGroup(gomock.Any(), tt.groupID).
					Return(nil)
			} else {
				tt.setupMocks()
			}

			err := service.DeleteGroup(context.Background(), tt.groupID, tt.requesterID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGroupService_AddMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		groupID       uuid.UUID
		userID        uuid.UUID
		inviterID     uuid.UUID
		role          domain.MemberRole
		setupMocks    func()
		expectedError string
	}{
		{
			name:      "successful add member",
			groupID:   uuid.New(),
			userID:    uuid.New(),
			inviterID: uuid.New(),
			role:      domain.RoleMember,
			setupMocks: func() {
				// Permission check
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleAdmin, nil)

				// User validation
				mockValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				// Check if already member
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Add member
				mockRepo.EXPECT().
					AddMember(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, member *domain.GroupMember) {
						assert.Equal(t, domain.RoleMember, member.Role)
					}).
					Return(nil)

				// Update group member count
				group := &domain.Group{
					ID:          uuid.New(),
					MemberCount: 1,
					Version:     1,
				}
				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					UpdateGroup(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, g *domain.Group) {
						assert.Equal(t, 2, g.MemberCount)
						assert.Equal(t, 2, g.Version)
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:      "user already member",
			groupID:   uuid.New(),
			userID:    uuid.New(),
			inviterID: uuid.New(),
			role:      domain.RoleMember,
			setupMocks: func() {
				// Permission check
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleAdmin, nil)

				// User validation
				mockValidator.EXPECT().
					UserExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				// Check if already member
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
			},
			expectedError: "user is already a member",
		},
		{
			name:      "insufficient permissions",
			groupID:   uuid.New(),
			userID:    uuid.New(),
			inviterID: uuid.New(),
			role:      domain.RoleMember,
			setupMocks: func() {
				// Permission check
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleMember, nil)
			},
			expectedError: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.AddMember(context.Background(), tt.groupID, tt.userID, tt.inviterID, tt.role)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGroupService_RemoveMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		groupID       uuid.UUID
		userID        uuid.UUID
		requesterID   uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful remove member",
			groupID:     uuid.New(),
			userID:      uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				// Permission check
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleAdmin, nil)

				// Remove member
				mockRepo.EXPECT().
					RemoveMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				// Update group member count
				group := &domain.Group{
					ID:          uuid.New(),
					MemberCount: 2,
					Version:     1,
				}
				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					UpdateGroup(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, g *domain.Group) {
						assert.Equal(t, 1, g.MemberCount)
						assert.Equal(t, 2, g.Version)
					}).
					Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "self removal",
			groupID:     uuid.New(),
			userID:      uuid.New(),
			requesterID: func() uuid.UUID { id := uuid.New(); return id }(), // Will be set to same as userID
			setupMocks: func() {
				// Permission check still happens for self-removal
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleMember, nil)

				// Remove member
				mockRepo.EXPECT().
					RemoveMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				// Update group member count
				group := &domain.Group{
					ID:          uuid.New(),
					MemberCount: 2,
					Version:     1,
				}
				mockRepo.EXPECT().
					GetGroupByID(gomock.Any(), gomock.Any()).
					Return(group, nil)

				mockRepo.EXPECT().
					UpdateGroup(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle self removal case
			if tt.name == "self removal" {
				tt.requesterID = tt.userID
			}

			tt.setupMocks()

			err := service.RemoveMember(context.Background(), tt.groupID, tt.userID, tt.requesterID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGroupService_GetMyGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		userID        uuid.UUID
		groupType     *domain.GroupType
		pagination    commonDomain.Pagination
		setupMocks    func()
		expectedError string
		expectedCount int
		expectedTotal int
	}{
		{
			name:       "successful get all groups",
			userID:     uuid.New(),
			groupType:  nil,
			pagination: commonDomain.Pagination{Page: 1, PageSize: 10},
			setupMocks: func() {
				ownedGroups := []*domain.Group{
					{ID: uuid.New(), Name: "Owned Group", Type: domain.GroupTypeProject},
				}
				memberGroups := []*domain.Group{
					{ID: uuid.New(), Name: "Member Group", Type: domain.GroupTypeProject},
				}

				mockRepo.EXPECT().
					ListGroupsByOwner(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(ownedGroups, 1, nil)

				mockRepo.EXPECT().
					ListGroupsByMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(memberGroups, 1, nil)
			},
			expectedError: "",
			expectedCount: 2,
			expectedTotal: 2,
		},
		{
			name:       "with type filter",
			userID:     uuid.New(),
			groupType:  &[]domain.GroupType{domain.GroupTypeSchedule}[0],
			pagination: commonDomain.Pagination{Page: 1, PageSize: 10},
			setupMocks: func() {
				ownedGroups := []*domain.Group{
					{ID: uuid.New(), Name: "Project Group", Type: domain.GroupTypeProject},
					{ID: uuid.New(), Name: "Schedule Group", Type: domain.GroupTypeSchedule},
				}
				memberGroups := []*domain.Group{
					{ID: uuid.New(), Name: "Another Schedule Group", Type: domain.GroupTypeSchedule},
				}

				mockRepo.EXPECT().
					ListGroupsByOwner(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(ownedGroups, 2, nil)

				mockRepo.EXPECT().
					ListGroupsByMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(memberGroups, 1, nil)
			},
			expectedError: "",
			expectedCount: 2, // Only schedule groups
			expectedTotal: 3, // Total before filtering
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, total, err := service.GetMyGroups(context.Background(), tt.userID, tt.groupType, tt.pagination)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))
				assert.Equal(t, tt.expectedTotal, total)

				// Verify type filter if applied
				if tt.groupType != nil {
					for _, group := range result {
						assert.Equal(t, *tt.groupType, group.Type)
					}
				}
			}
		})
	}
}

func TestGroupService_UpdateMemberRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		groupID       uuid.UUID
		userID        uuid.UUID
		requesterID   uuid.UUID
		newRole       domain.MemberRole
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "cannot change owner role",
			groupID:     uuid.New(),
			userID:      uuid.New(),
			requesterID: uuid.New(),
			newRole:     domain.RoleMember,
			setupMocks: func() {
				// Permission check
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleAdmin, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleAdmin, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleOwner, nil)
			},
			expectedError: "cannot change owner role",
		},
		{
			name:        "insufficient permissions",
			groupID:     uuid.New(),
			userID:      uuid.New(),
			requesterID: uuid.New(),
			newRole:     domain.RoleMember,
			setupMocks: func() {
				// Permission check
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleMember, nil)
			},
			expectedError: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := service.UpdateMemberRole(context.Background(), tt.groupID, tt.userID, tt.requesterID, tt.newRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGroupService_GetGroupStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGroupRepository(ctrl)
	mockValidator := mocks.NewMockUserValidator(ctrl)
	mockLogger := *logger.NewLogger(&logger.Config{
		Level:       "error", // Only log errors to reduce noise in tests
		Output:      "console",
		Development: false,
	})
	service := NewGroupService(mockRepo, mockValidator, &mockLogger)

	tests := []struct {
		name          string
		groupID       uuid.UUID
		requesterID   uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:        "successful get stats",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				expectedStats := &domain.GroupStats{
					MemberCount:   5,
					ActiveMembers: 3,
				}

				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)

				mockRepo.EXPECT().
					GetMemberRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RoleMember, nil)

				mockRepo.EXPECT().
					GetGroupStats(gomock.Any(), gomock.Any()).
					Return(expectedStats, nil)
			},
			expectedError: "",
		},
		{
			name:        "not a member",
			groupID:     uuid.New(),
			requesterID: uuid.New(),
			setupMocks: func() {
				mockRepo.EXPECT().
					IsMember(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil)
			},
			expectedError: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := service.GetGroupStats(context.Background(), tt.groupID, tt.requesterID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGroupService_hasPermissionForAction(t *testing.T) {
	service := &groupService{}

	tests := []struct {
		name     string
		role     domain.MemberRole
		action   GroupAction
		expected bool
	}{
		{"Owner can edit group", domain.RoleOwner, ActionEditGroup, true},
		{"Admin can edit group", domain.RoleAdmin, ActionEditGroup, true},
		{"Member cannot edit group", domain.RoleMember, ActionEditGroup, false},
		{"All members can view group", domain.RoleMember, ActionViewGroup, true},
		{"All members can view tasks", domain.RoleMember, ActionViewTasks, true},
		{"Unknown action denied", domain.RoleOwner, GroupAction("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.hasPermissionForAction(tt.role, tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}
