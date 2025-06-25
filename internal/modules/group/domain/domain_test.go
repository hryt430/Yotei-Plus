package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGroup(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		description string
		groupType   GroupType
		ownerID     uuid.UUID
		wantError   bool
	}{
		{
			name:        "Valid project group creation",
			groupName:   "Test Project",
			description: "Test project description",
			groupType:   GroupTypeProject,
			ownerID:     uuid.New(),
			wantError:   false,
		},
		{
			name:        "Valid schedule group creation",
			groupName:   "Schedule Group",
			description: "Schedule sharing group",
			groupType:   GroupTypeSchedule,
			ownerID:     uuid.New(),
			wantError:   false,
		},
		{
			name:        "Empty name should work",
			groupName:   "",
			description: "Empty name test",
			groupType:   GroupTypeProject,
			ownerID:     uuid.New(),
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := NewGroup(tt.groupName, tt.description, tt.groupType, tt.ownerID)

			// Basic validations
			require.NotNil(t, group)
			assert.NotEqual(t, uuid.Nil, group.ID, "Group ID should be generated")
			assert.Equal(t, tt.groupName, group.Name)
			assert.Equal(t, tt.description, group.Description)
			assert.Equal(t, tt.groupType, group.Type)
			assert.Equal(t, tt.ownerID, group.OwnerID)
			assert.Equal(t, 1, group.MemberCount, "Initial member count should be 1 (owner)")
			assert.Equal(t, 1, group.Version, "Initial version should be 1")

			// Time validations
			now := time.Now()
			assert.WithinDuration(t, now, group.CreatedAt, time.Second)
			assert.WithinDuration(t, now, group.UpdatedAt, time.Second)

			// Settings validation based on type
			if tt.groupType == GroupTypeProject {
				assert.True(t, group.Settings.EnableGanttChart, "Project groups should have Gantt chart enabled by default")
				assert.False(t, group.Settings.EnableTaskDependency, "Project groups should have task dependency disabled by default")
			}

			if tt.groupType == GroupTypeSchedule {
				assert.Equal(t, PrivacyLevelBusy, group.Settings.DefaultPrivacyLevel, "Schedule groups should have BUSY privacy level by default")
				assert.False(t, group.Settings.AllowScheduleDetails, "Schedule groups should not allow schedule details by default")
			}

			// Common default settings
			assert.False(t, group.Settings.IsPublic, "Groups should be private by default")
			assert.True(t, group.Settings.AllowMemberInvite, "Groups should allow member invite by default")
			assert.True(t, group.Settings.RequireApproval, "Groups should require approval by default")
			assert.True(t, group.Settings.EnableNotifications, "Groups should have notifications enabled by default")
		})
	}
}

func TestGroup_UpdateSettings(t *testing.T) {
	// Setup
	ownerID := uuid.New()
	group := NewGroup("Test Group", "Description", GroupTypeProject, ownerID)
	originalVersion := group.Version
	originalUpdatedAt := group.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Create new settings
	newSettings := GroupSettings{
		IsPublic:             true,
		AllowMemberInvite:    false,
		RequireApproval:      false,
		EnableNotifications:  false,
		EnableGanttChart:     false,
		EnableTaskDependency: true,
	}

	// Execute
	group.UpdateSettings(newSettings)

	// Verify
	assert.Equal(t, newSettings, group.Settings, "Settings should be updated")
	assert.Equal(t, originalVersion+1, group.Version, "Version should be incremented")
	assert.True(t, group.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestGroup_AddMember(t *testing.T) {
	// Setup
	ownerID := uuid.New()
	group := NewGroup("Test Group", "Description", GroupTypeProject, ownerID)
	originalMemberCount := group.MemberCount
	originalVersion := group.Version
	originalUpdatedAt := group.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	group.AddMember()

	// Verify
	assert.Equal(t, originalMemberCount+1, group.MemberCount, "Member count should be incremented")
	assert.Equal(t, originalVersion+1, group.Version, "Version should be incremented")
	assert.True(t, group.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestGroup_RemoveMember(t *testing.T) {
	tests := []struct {
		name             string
		initialMembers   int
		expectedError    bool
		expectedMembers  int
		expectedErrorMsg string
	}{
		{
			name:            "Remove member from group with multiple members",
			initialMembers:  3,
			expectedError:   false,
			expectedMembers: 2,
		},
		{
			name:            "Remove member from group with two members",
			initialMembers:  2,
			expectedError:   false,
			expectedMembers: 1,
		},
		{
			name:             "Cannot remove last member",
			initialMembers:   1,
			expectedError:    true,
			expectedMembers:  1,
			expectedErrorMsg: "cannot remove the last member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ownerID := uuid.New()
			group := NewGroup("Test Group", "Description", GroupTypeProject, ownerID)

			// Set initial member count
			group.MemberCount = tt.initialMembers
			originalVersion := group.Version
			originalUpdatedAt := group.UpdatedAt

			// Wait a bit to ensure timestamp difference
			time.Sleep(time.Millisecond)

			// Execute
			err := group.RemoveMember()

			// Verify
			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				assert.Equal(t, tt.expectedMembers, group.MemberCount, "Member count should not change on error")
				assert.Equal(t, originalVersion, group.Version, "Version should not change on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMembers, group.MemberCount, "Member count should be decremented")
				assert.Equal(t, originalVersion+1, group.Version, "Version should be incremented")
				assert.True(t, group.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func TestNewGroupMember(t *testing.T) {
	// Setup
	groupID := uuid.New()
	userID := uuid.New()
	role := RoleMember

	// Execute
	member := NewGroupMember(groupID, userID, role)

	// Verify
	require.NotNil(t, member)
	assert.NotEqual(t, uuid.Nil, member.ID, "Member ID should be generated")
	assert.Equal(t, groupID, member.GroupID)
	assert.Equal(t, userID, member.UserID)
	assert.Equal(t, role, member.Role)

	// Time validations
	now := time.Now()
	assert.WithinDuration(t, now, member.JoinedAt, time.Second)
	assert.WithinDuration(t, now, member.UpdatedAt, time.Second)
}

func TestGroupMember_CanManageGroup(t *testing.T) {
	tests := []struct {
		name     string
		role     MemberRole
		expected bool
	}{
		{
			name:     "Owner can manage group",
			role:     RoleOwner,
			expected: true,
		},
		{
			name:     "Admin can manage group",
			role:     RoleAdmin,
			expected: true,
		},
		{
			name:     "Member cannot manage group",
			role:     RoleMember,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := NewGroupMember(uuid.New(), uuid.New(), tt.role)
			result := member.CanManageGroup()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupMember_CanInviteMembers(t *testing.T) {
	tests := []struct {
		name              string
		role              MemberRole
		allowMemberInvite bool
		expected          bool
	}{
		{
			name:              "Owner can always invite",
			role:              RoleOwner,
			allowMemberInvite: false,
			expected:          true,
		},
		{
			name:              "Admin can always invite",
			role:              RoleAdmin,
			allowMemberInvite: false,
			expected:          true,
		},
		{
			name:              "Member can invite when allowed",
			role:              RoleMember,
			allowMemberInvite: true,
			expected:          true,
		},
		{
			name:              "Member cannot invite when not allowed",
			role:              RoleMember,
			allowMemberInvite: false,
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := NewGroupMember(uuid.New(), uuid.New(), tt.role)
			settings := GroupSettings{
				AllowMemberInvite: tt.allowMemberInvite,
			}
			result := member.CanInviteMembers(settings)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupMember_PromoteToAdmin(t *testing.T) {
	tests := []struct {
		name      string
		role      MemberRole
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Member can be promoted to admin",
			role:      RoleMember,
			wantError: false,
		},
		{
			name:      "Admin can be promoted to admin (no change)",
			role:      RoleAdmin,
			wantError: false,
		},
		{
			name:      "Owner cannot be promoted",
			role:      RoleOwner,
			wantError: true,
			errorMsg:  "owner cannot be promoted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := NewGroupMember(uuid.New(), uuid.New(), tt.role)
			originalUpdatedAt := member.UpdatedAt

			// Wait a bit to ensure timestamp difference
			time.Sleep(time.Millisecond)

			err := member.PromoteToAdmin()

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, tt.role, member.Role, "Role should not change on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, RoleAdmin, member.Role, "Role should be updated to admin")
				assert.True(t, member.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func TestGroupMember_DemoteToMember(t *testing.T) {
	tests := []struct {
		name      string
		role      MemberRole
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Admin can be demoted to member",
			role:      RoleAdmin,
			wantError: false,
		},
		{
			name:      "Member can be demoted to member (no change)",
			role:      RoleMember,
			wantError: false,
		},
		{
			name:      "Owner cannot be demoted",
			role:      RoleOwner,
			wantError: true,
			errorMsg:  "owner cannot be demoted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := NewGroupMember(uuid.New(), uuid.New(), tt.role)
			originalUpdatedAt := member.UpdatedAt

			// Wait a bit to ensure timestamp difference
			time.Sleep(time.Millisecond)

			err := member.DemoteToMember()

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, tt.role, member.Role, "Role should not change on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, RoleMember, member.Role, "Role should be updated to member")
				assert.True(t, member.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func TestGroupMember_TransferOwnership(t *testing.T) {
	// Setup
	member := NewGroupMember(uuid.New(), uuid.New(), RoleAdmin)
	originalUpdatedAt := member.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	member.TransferOwnership()

	// Verify
	assert.Equal(t, RoleOwner, member.Role, "Role should be updated to owner")
	assert.True(t, member.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestGroupTypes(t *testing.T) {
	// Test that group types are correctly defined
	assert.Equal(t, GroupType("PROJECT"), GroupTypeProject)
	assert.Equal(t, GroupType("SCHEDULE"), GroupTypeSchedule)
}

func TestMemberRoles(t *testing.T) {
	// Test that member roles are correctly defined
	assert.Equal(t, MemberRole("OWNER"), RoleOwner)
	assert.Equal(t, MemberRole("ADMIN"), RoleAdmin)
	assert.Equal(t, MemberRole("MEMBER"), RoleMember)
}

func TestPrivacyLevels(t *testing.T) {
	// Test that privacy levels are correctly defined
	assert.Equal(t, PrivacyLevel("NONE"), PrivacyLevelNone)
	assert.Equal(t, PrivacyLevel("BUSY"), PrivacyLevelBusy)
	assert.Equal(t, PrivacyLevel("TITLE"), PrivacyLevelTitle)
	assert.Equal(t, PrivacyLevel("DETAILS"), PrivacyLevelDetails)
}

func TestGroupSettings_TypeSpecificDefaults(t *testing.T) {
	tests := []struct {
		name      string
		groupType GroupType
		checkFunc func(t *testing.T, settings GroupSettings)
	}{
		{
			name:      "Project group default settings",
			groupType: GroupTypeProject,
			checkFunc: func(t *testing.T, settings GroupSettings) {
				assert.True(t, settings.EnableGanttChart, "Project groups should have Gantt chart enabled")
				assert.False(t, settings.EnableTaskDependency, "Project groups should have task dependency disabled")
			},
		},
		{
			name:      "Schedule group default settings",
			groupType: GroupTypeSchedule,
			checkFunc: func(t *testing.T, settings GroupSettings) {
				assert.Equal(t, PrivacyLevelBusy, settings.DefaultPrivacyLevel, "Schedule groups should have BUSY privacy level")
				assert.False(t, settings.AllowScheduleDetails, "Schedule groups should not allow schedule details")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := NewGroup("Test", "Description", tt.groupType, uuid.New())
			tt.checkFunc(t, group.Settings)
		})
	}
}

// Benchmark tests for performance
func BenchmarkNewGroup(b *testing.B) {
	ownerID := uuid.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewGroup("Test Group", "Description", GroupTypeProject, ownerID)
	}
}

func BenchmarkGroupMember_CanManageGroup(b *testing.B) {
	member := NewGroupMember(uuid.New(), uuid.New(), RoleAdmin)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		member.CanManageGroup()
	}
}

func BenchmarkGroupMember_CanInviteMembers(b *testing.B) {
	member := NewGroupMember(uuid.New(), uuid.New(), RoleMember)
	settings := GroupSettings{AllowMemberInvite: true}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		member.CanInviteMembers(settings)
	}
}
