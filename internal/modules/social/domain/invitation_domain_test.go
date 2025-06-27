package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInvitation(t *testing.T) {
	tests := []struct {
		name              string
		invitationType    InvitationType
		method            InvitationMethod
		inviterID         uuid.UUID
		message           string
		expirationHours   int
		wantCodeGenerated bool
		wantURLGenerated  bool
	}{
		{
			name:              "Friend invitation with code method",
			invitationType:    InvitationTypeFriend,
			method:            MethodCode,
			inviterID:         uuid.New(),
			message:           "Let's be friends!",
			expirationHours:   24,
			wantCodeGenerated: true,
			wantURLGenerated:  false,
		},
		{
			name:              "Group invitation with URL method",
			invitationType:    InvitationTypeGroup,
			method:            MethodURL,
			inviterID:         uuid.New(),
			message:           "Join our group!",
			expirationHours:   48,
			wantCodeGenerated: true,
			wantURLGenerated:  true,
		},
		{
			name:              "In-app invitation",
			invitationType:    InvitationTypeFriend,
			method:            MethodInApp,
			inviterID:         uuid.New(),
			message:           "In-app invitation",
			expirationHours:   168,
			wantCodeGenerated: false,
			wantURLGenerated:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := NewInvitation(tt.invitationType, tt.method, tt.inviterID, tt.message, tt.expirationHours)

			// Basic validations
			require.NotNil(t, invitation)
			assert.NotEqual(t, uuid.Nil, invitation.ID, "Invitation ID should be generated")
			assert.Equal(t, tt.invitationType, invitation.Type)
			assert.Equal(t, tt.method, invitation.Method)
			assert.Equal(t, InvitationStatusPending, invitation.Status)
			assert.Equal(t, tt.inviterID, invitation.InviterID)
			assert.Equal(t, tt.message, invitation.Message)
			assert.Nil(t, invitation.InviteeID, "InviteeID should be nil initially")
			assert.Nil(t, invitation.InviteeInfo, "InviteeInfo should be nil initially")
			assert.Nil(t, invitation.TargetID, "TargetID should be nil initially")
			assert.Nil(t, invitation.AcceptedAt, "AcceptedAt should be nil initially")

			// Metadata should be initialized
			require.NotNil(t, invitation.Metadata)
			assert.Equal(t, 0, len(invitation.Metadata))

			// Time validations
			now := time.Now()
			assert.WithinDuration(t, now, invitation.CreatedAt, time.Second)
			assert.WithinDuration(t, now, invitation.UpdatedAt, time.Second)

			// Expiration validation
			expectedExpiration := now.Add(time.Duration(tt.expirationHours) * time.Hour)
			assert.WithinDuration(t, expectedExpiration, invitation.ExpiresAt, time.Second)

			// Code and URL generation validation
			if tt.wantCodeGenerated {
				assert.NotEmpty(t, invitation.Code, "Code should be generated")
				assert.Len(t, invitation.Code, 16, "Code should be 16 characters long") // 8 bytes * 2 (hex)
			} else {
				assert.Empty(t, invitation.Code, "Code should not be generated")
			}

			if tt.wantURLGenerated {
				assert.NotEmpty(t, invitation.URL, "URL should be generated")
				assert.Contains(t, invitation.URL, invitation.Code, "URL should contain the code")
			} else {
				assert.Empty(t, invitation.URL, "URL should not be generated")
			}
		})
	}
}

func TestInvitation_SetInvitee(t *testing.T) {
	// Setup
	invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
	userID := uuid.New()
	originalUpdatedAt := invitation.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	invitation.SetInvitee(userID)

	// Verify
	require.NotNil(t, invitation.InviteeID)
	assert.Equal(t, userID, *invitation.InviteeID)
	assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestInvitation_SetInviteeInfo(t *testing.T) {
	// Setup
	invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
	inviteeInfo := InviteeInfo{
		Email:    "invitee@example.com",
		Username: "invitee_user",
		Phone:    "+1234567890",
	}
	originalUpdatedAt := invitation.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	invitation.SetInviteeInfo(inviteeInfo)

	// Verify
	require.NotNil(t, invitation.InviteeInfo)
	assert.Equal(t, inviteeInfo, *invitation.InviteeInfo)
	assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestInvitation_SetTarget(t *testing.T) {
	// Setup
	invitation := NewInvitation(InvitationTypeGroup, MethodCode, uuid.New(), "Test", 24)
	targetID := uuid.New()
	originalUpdatedAt := invitation.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	invitation.SetTarget(targetID)

	// Verify
	require.NotNil(t, invitation.TargetID)
	assert.Equal(t, targetID, *invitation.TargetID)
	assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestInvitation_Accept(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*Invitation)
		expectedError bool
		errorType     error
	}{
		{
			name:          "Accept valid invitation",
			setupFunc:     func(i *Invitation) {}, // No setup needed
			expectedError: false,
		},
		{
			name: "Cannot accept expired invitation",
			setupFunc: func(i *Invitation) {
				i.ExpiresAt = time.Now().Add(-time.Hour)
			},
			expectedError: true,
			errorType:     ErrInvitationExpired,
		},
		{
			name: "Cannot accept already accepted invitation",
			setupFunc: func(i *Invitation) {
				i.Status = InvitationStatusAccepted
			},
			expectedError: true,
			errorType:     ErrInvalidInvitationStatus,
		},
		{
			name: "Cannot accept declined invitation",
			setupFunc: func(i *Invitation) {
				i.Status = InvitationStatusDeclined
			},
			expectedError: true,
			errorType:     ErrInvalidInvitationStatus,
		},
		{
			name: "Cannot accept canceled invitation",
			setupFunc: func(i *Invitation) {
				i.Status = InvitationStatusCanceled
			},
			expectedError: true,
			errorType:     ErrInvalidInvitationStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
			tt.setupFunc(invitation)

			originalUpdatedAt := invitation.UpdatedAt
			time.Sleep(time.Millisecond)

			err := invitation.Accept()

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
				// Should not modify the invitation on error
				assert.Nil(t, invitation.AcceptedAt)
			} else {
				require.NoError(t, err)
				assert.Equal(t, InvitationStatusAccepted, invitation.Status)
				require.NotNil(t, invitation.AcceptedAt)
				assert.WithinDuration(t, time.Now(), *invitation.AcceptedAt, time.Second)
				assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func TestInvitation_Decline(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus InvitationStatus
		expectedError bool
		errorType     error
	}{
		{
			name:          "Decline pending invitation",
			initialStatus: InvitationStatusPending,
			expectedError: false,
		},
		{
			name:          "Cannot decline accepted invitation",
			initialStatus: InvitationStatusAccepted,
			expectedError: true,
			errorType:     ErrInvalidInvitationStatus,
		},
		{
			name:          "Cannot decline already declined invitation",
			initialStatus: InvitationStatusDeclined,
			expectedError: true,
			errorType:     ErrInvalidInvitationStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
			invitation.Status = tt.initialStatus

			originalUpdatedAt := invitation.UpdatedAt
			time.Sleep(time.Millisecond)

			err := invitation.Decline()

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, InvitationStatusDeclined, invitation.Status)
				assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func TestInvitation_Cancel(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus InvitationStatus
		expectedError bool
	}{
		{
			name:          "Cancel pending invitation",
			initialStatus: InvitationStatusPending,
			expectedError: false,
		},
		{
			name:          "Cannot cancel accepted invitation",
			initialStatus: InvitationStatusAccepted,
			expectedError: true,
		},
		{
			name:          "Cannot cancel declined invitation",
			initialStatus: InvitationStatusDeclined,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
			invitation.Status = tt.initialStatus

			originalUpdatedAt := invitation.UpdatedAt
			time.Sleep(time.Millisecond)

			err := invitation.Cancel()

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, InvitationStatusCanceled, invitation.Status)
				assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func TestInvitation_MarkAsExpired(t *testing.T) {
	// Setup
	invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
	originalUpdatedAt := invitation.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	invitation.MarkAsExpired()

	// Verify
	assert.Equal(t, InvitationStatusExpired, invitation.Status)
	assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestInvitation_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "Not expired invitation",
			expiresAt: time.Now().Add(time.Hour),
			expected:  false,
		},
		{
			name:      "Expired invitation",
			expiresAt: time.Now().Add(-time.Hour),
			expected:  true,
		},
		{
			name:      "Just expired invitation",
			expiresAt: time.Now().Add(-time.Minute),
			expected:  true,
		},
		{
			name:      "Future invitation",
			expiresAt: time.Now().Add(24 * time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
			invitation.ExpiresAt = tt.expiresAt

			result := invitation.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInvitation_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Invitation)
		expected  bool
	}{
		{
			name:      "Valid pending invitation",
			setupFunc: func(i *Invitation) {}, // No setup needed
			expected:  true,
		},
		{
			name: "Invalid expired invitation",
			setupFunc: func(i *Invitation) {
				i.ExpiresAt = time.Now().Add(-time.Hour)
			},
			expected: false,
		},
		{
			name: "Invalid accepted invitation",
			setupFunc: func(i *Invitation) {
				i.Status = InvitationStatusAccepted
			},
			expected: false,
		},
		{
			name: "Invalid declined invitation",
			setupFunc: func(i *Invitation) {
				i.Status = InvitationStatusDeclined
			},
			expected: false,
		},
		{
			name: "Invalid canceled invitation",
			setupFunc: func(i *Invitation) {
				i.Status = InvitationStatusCanceled
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
			tt.setupFunc(invitation)

			result := invitation.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInvitation_AddMetadata(t *testing.T) {
	// Setup
	invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
	originalUpdatedAt := invitation.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond)

	// Execute
	invitation.AddMetadata("key1", "value1")
	invitation.AddMetadata("key2", "value2")

	// Verify
	require.NotNil(t, invitation.Metadata)
	assert.Equal(t, "value1", invitation.Metadata["key1"])
	assert.Equal(t, "value2", invitation.Metadata["key2"])
	assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
}

func TestInvitation_AddMetadata_NilMetadata(t *testing.T) {
	// Setup invitation with nil metadata
	invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
	invitation.Metadata = nil

	// Execute
	invitation.AddMetadata("key", "value")

	// Verify metadata is initialized and set
	require.NotNil(t, invitation.Metadata)
	assert.Equal(t, "value", invitation.Metadata["key"])
}

func TestInvitationTypes(t *testing.T) {
	// Test that invitation types are correctly defined
	assert.Equal(t, InvitationType("FRIEND"), InvitationTypeFriend)
	assert.Equal(t, InvitationType("GROUP"), InvitationTypeGroup)
}

func TestInvitationMethods(t *testing.T) {
	// Test that invitation methods are correctly defined
	assert.Equal(t, InvitationMethod("IN_APP"), MethodInApp)
	assert.Equal(t, InvitationMethod("CODE"), MethodCode)
	assert.Equal(t, InvitationMethod("URL"), MethodURL)
}

func TestInvitationStatuses(t *testing.T) {
	// Test that invitation statuses are correctly defined
	assert.Equal(t, InvitationStatus("PENDING"), InvitationStatusPending)
	assert.Equal(t, InvitationStatus("ACCEPTED"), InvitationStatusAccepted)
	assert.Equal(t, InvitationStatus("DECLINED"), InvitationStatusDeclined)
	assert.Equal(t, InvitationStatus("EXPIRED"), InvitationStatusExpired)
	assert.Equal(t, InvitationStatus("CANCELED"), InvitationStatusCanceled)
}

func TestInvitation_ErrorConstants(t *testing.T) {
	// Test that error constants are properly defined
	assert.NotNil(t, ErrInvitationExpired)
	assert.NotNil(t, ErrInvalidInvitationStatus)
	assert.Contains(t, ErrInvitationExpired.Error(), "expired")
	assert.Contains(t, ErrInvalidInvitationStatus.Error(), "invalid")
}

func TestInvitation_CompleteWorkflow(t *testing.T) {
	// Test a complete invitation workflow
	inviterID := uuid.New()
	inviteeID := uuid.New()
	targetID := uuid.New()

	// Create invitation
	invitation := NewInvitation(InvitationTypeGroup, MethodURL, inviterID, "Join our team!", 48)

	// Set target and invitee info
	invitation.SetTarget(targetID)
	invitation.SetInviteeInfo(InviteeInfo{
		Email:    "invitee@example.com",
		Username: "invitee",
	})

	// Add metadata
	invitation.AddMetadata("source", "web")
	invitation.AddMetadata("campaign", "Q1_2024")

	// Set invitee when they register
	invitation.SetInvitee(inviteeID)

	// Accept invitation
	err := invitation.Accept()
	require.NoError(t, err)

	// Verify final state
	assert.Equal(t, InvitationStatusAccepted, invitation.Status)
	assert.Equal(t, inviterID, invitation.InviterID)
	assert.Equal(t, inviteeID, *invitation.InviteeID)
	assert.Equal(t, targetID, *invitation.TargetID)
	require.NotNil(t, invitation.InviteeInfo)
	assert.Equal(t, "invitee@example.com", invitation.InviteeInfo.Email)
	assert.Equal(t, "web", invitation.Metadata["source"])
	assert.NotNil(t, invitation.AcceptedAt)
	assert.True(t, invitation.IsFriend() == false) // Group invitation, not friend
}

// Benchmark tests
func BenchmarkNewInvitation(b *testing.B) {
	inviterID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewInvitation(InvitationTypeFriend, MethodCode, inviterID, "Test", 24)
	}
}

func BenchmarkInvitation_IsValid(b *testing.B) {
	invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invitation.IsValid()
	}
}

func BenchmarkInvitation_Accept(b *testing.B) {
	invitations := make([]*Invitation, b.N)
	for i := 0; i < b.N; i++ {
		invitations[i] = NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invitations[i].Accept()
	}
}

func BenchmarkInvitation_AddMetadata(b *testing.B) {
	invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invitation.AddMetadata("key", "value")
	}
}

// Edge case tests
func TestInvitation_EdgeCases(t *testing.T) {
	t.Run("Zero expiration hours", func(t *testing.T) {
		invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 0)
		// Should expire immediately
		assert.True(t, invitation.IsExpired())
		assert.False(t, invitation.IsValid())
	})

	t.Run("Negative expiration hours", func(t *testing.T) {
		invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", -1)
		// Should be expired
		assert.True(t, invitation.IsExpired())
		assert.False(t, invitation.IsValid())
	})

	t.Run("Empty message", func(t *testing.T) {
		invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "", 24)
		assert.Equal(t, "", invitation.Message)
		assert.True(t, invitation.IsValid())
	})

	t.Run("Very long expiration", func(t *testing.T) {
		invitation := NewInvitation(InvitationTypeFriend, MethodCode, uuid.New(), "Test", 8760) // 1 year
		assert.False(t, invitation.IsExpired())
		assert.True(t, invitation.IsValid())

		expectedExpiration := time.Now().Add(8760 * time.Hour)
		assert.WithinDuration(t, expectedExpiration, invitation.ExpiresAt, time.Second)
	})
}
