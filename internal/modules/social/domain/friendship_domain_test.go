package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFriendship(t *testing.T) {
	// Setup
	requesterID := uuid.New()
	addresseeID := uuid.New()

	// Execute
	friendship := NewFriendship(requesterID, addresseeID)

	// Verify
	require.NotNil(t, friendship)
	assert.NotEqual(t, uuid.Nil, friendship.ID, "Friendship ID should be generated")
	assert.Equal(t, requesterID, friendship.RequesterID)
	assert.Equal(t, addresseeID, friendship.AddresseeID)
	assert.Equal(t, FriendshipStatusPending, friendship.Status, "New friendship should be pending")
	assert.Nil(t, friendship.AcceptedAt, "AcceptedAt should be nil for new friendship")
	assert.Nil(t, friendship.BlockedAt, "BlockedAt should be nil for new friendship")

	// Time validations
	now := time.Now()
	assert.WithinDuration(t, now, friendship.CreatedAt, time.Second)
	assert.WithinDuration(t, now, friendship.UpdatedAt, time.Second)
}

func TestFriendship_Accept(t *testing.T) {
	// Setup
	requesterID := uuid.New()
	addresseeID := uuid.New()
	friendship := NewFriendship(requesterID, addresseeID)
	originalUpdatedAt := friendship.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	friendship.Accept()

	// Verify
	assert.Equal(t, FriendshipStatusAccepted, friendship.Status, "Status should be accepted")
	require.NotNil(t, friendship.AcceptedAt, "AcceptedAt should be set")
	assert.WithinDuration(t, time.Now(), *friendship.AcceptedAt, time.Second)
	assert.True(t, friendship.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
	assert.Nil(t, friendship.BlockedAt, "BlockedAt should remain nil")
}

func TestFriendship_Block(t *testing.T) {
	// Setup
	requesterID := uuid.New()
	addresseeID := uuid.New()
	friendship := NewFriendship(requesterID, addresseeID)
	originalUpdatedAt := friendship.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	friendship.Block()

	// Verify
	assert.Equal(t, FriendshipStatusBlocked, friendship.Status, "Status should be blocked")
	require.NotNil(t, friendship.BlockedAt, "BlockedAt should be set")
	assert.WithinDuration(t, time.Now(), *friendship.BlockedAt, time.Second)
	assert.True(t, friendship.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
	assert.Nil(t, friendship.AcceptedAt, "AcceptedAt should remain nil")
}

func TestFriendship_IsFriend(t *testing.T) {
	tests := []struct {
		name     string
		status   FriendshipStatus
		expected bool
	}{
		{
			name:     "Pending friendship is not friend",
			status:   FriendshipStatusPending,
			expected: false,
		},
		{
			name:     "Accepted friendship is friend",
			status:   FriendshipStatusAccepted,
			expected: true,
		},
		{
			name:     "Blocked friendship is not friend",
			status:   FriendshipStatusBlocked,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friendship := NewFriendship(uuid.New(), uuid.New())
			friendship.Status = tt.status

			result := friendship.IsFriend()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFriendship_IsBlocked(t *testing.T) {
	tests := []struct {
		name     string
		status   FriendshipStatus
		expected bool
	}{
		{
			name:     "Pending friendship is not blocked",
			status:   FriendshipStatusPending,
			expected: false,
		},
		{
			name:     "Accepted friendship is not blocked",
			status:   FriendshipStatusAccepted,
			expected: false,
		},
		{
			name:     "Blocked friendship is blocked",
			status:   FriendshipStatusBlocked,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friendship := NewFriendship(uuid.New(), uuid.New())
			friendship.Status = tt.status

			result := friendship.IsBlocked()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFriendship_StatusTransitions(t *testing.T) {
	// Test multiple status transitions
	requesterID := uuid.New()
	addresseeID := uuid.New()
	friendship := NewFriendship(requesterID, addresseeID)

	// Initially pending
	assert.Equal(t, FriendshipStatusPending, friendship.Status)
	assert.False(t, friendship.IsFriend())
	assert.False(t, friendship.IsBlocked())

	// Accept friendship
	friendship.Accept()
	assert.Equal(t, FriendshipStatusAccepted, friendship.Status)
	assert.True(t, friendship.IsFriend())
	assert.False(t, friendship.IsBlocked())

	// Block after acceptance (edge case)
	friendship.Block()
	assert.Equal(t, FriendshipStatusBlocked, friendship.Status)
	assert.False(t, friendship.IsFriend())
	assert.True(t, friendship.IsBlocked())
}

func TestFriendship_SameUserIDs(t *testing.T) {
	// Test edge case: same user as requester and addressee
	userID := uuid.New()
	friendship := NewFriendship(userID, userID)

	require.NotNil(t, friendship)
	assert.Equal(t, userID, friendship.RequesterID)
	assert.Equal(t, userID, friendship.AddresseeID)
	assert.Equal(t, FriendshipStatusPending, friendship.Status)

	// Should still work with status changes
	friendship.Accept()
	assert.True(t, friendship.IsFriend())
}

func TestFriendshipStatus_Constants(t *testing.T) {
	// Test that friendship status constants are correctly defined
	assert.Equal(t, FriendshipStatus("PENDING"), FriendshipStatusPending)
	assert.Equal(t, FriendshipStatus("ACCEPTED"), FriendshipStatusAccepted)
	assert.Equal(t, FriendshipStatus("BLOCKED"), FriendshipStatusBlocked)
}

func TestFriendship_TimestampConsistency(t *testing.T) {
	// Test that timestamps are consistent and properly ordered
	friendship := NewFriendship(uuid.New(), uuid.New())

	createdTime := friendship.CreatedAt
	updatedTime := friendship.UpdatedAt

	// CreatedAt and UpdatedAt should be very close initially
	assert.WithinDuration(t, createdTime, updatedTime, time.Millisecond)

	// Wait and accept
	time.Sleep(time.Millisecond * 2)
	friendship.Accept()

	// UpdatedAt should be after CreatedAt
	assert.True(t, friendship.UpdatedAt.After(createdTime))
	assert.True(t, friendship.UpdatedAt.After(updatedTime))

	// AcceptedAt should be close to the updated UpdatedAt
	require.NotNil(t, friendship.AcceptedAt)
	assert.WithinDuration(t, friendship.UpdatedAt, *friendship.AcceptedAt, time.Millisecond)
}

func TestFriendship_BlockAfterAccept(t *testing.T) {
	// Test blocking after acceptance
	friendship := NewFriendship(uuid.New(), uuid.New())

	// Accept first
	friendship.Accept()
	acceptedAt := friendship.AcceptedAt
	require.NotNil(t, acceptedAt)

	// Wait and block
	time.Sleep(time.Millisecond)
	friendship.Block()

	// Should now be blocked
	assert.Equal(t, FriendshipStatusBlocked, friendship.Status)
	assert.True(t, friendship.IsBlocked())
	assert.False(t, friendship.IsFriend())

	// AcceptedAt should still be set (preserving history)
	assert.Equal(t, acceptedAt, friendship.AcceptedAt)

	// BlockedAt should be set and after AcceptedAt
	require.NotNil(t, friendship.BlockedAt)
	assert.True(t, friendship.BlockedAt.After(*acceptedAt))
}

func TestFriendship_AcceptAfterBlock(t *testing.T) {
	// Test accepting after blocking (edge case)
	friendship := NewFriendship(uuid.New(), uuid.New())

	// Block first
	friendship.Block()
	blockedAt := friendship.BlockedAt
	require.NotNil(t, blockedAt)

	// Wait and accept
	time.Sleep(time.Millisecond)
	friendship.Accept()

	// Should now be accepted
	assert.Equal(t, FriendshipStatusAccepted, friendship.Status)
	assert.True(t, friendship.IsFriend())
	assert.False(t, friendship.IsBlocked())

	// BlockedAt should still be set (preserving history)
	assert.Equal(t, blockedAt, friendship.BlockedAt)

	// AcceptedAt should be set and after BlockedAt
	require.NotNil(t, friendship.AcceptedAt)
	assert.True(t, friendship.AcceptedAt.After(*blockedAt))
}
