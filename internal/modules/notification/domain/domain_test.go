package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===================
// Domain Tests
// ===================

func TestNewNotification(t *testing.T) {
	userID := "user123"
	notificationType := TaskAssigned
	title := "Test Notification"
	message := "This is a test notification"
	metadata := map[string]string{
		"task_id":  "task123",
		"priority": "high",
	}

	notification := NewNotification(userID, notificationType, title, message, metadata)

	require.NotNil(t, notification)
	assert.NotEmpty(t, notification.ID)
	assert.Equal(t, userID, notification.UserID)
	assert.Equal(t, notificationType, notification.Type)
	assert.Equal(t, title, notification.Title)
	assert.Equal(t, message, notification.Message)
	assert.Equal(t, StatusPending, notification.Status)
	assert.Equal(t, metadata, notification.Metadata)
	assert.Empty(t, notification.Channels)
	assert.False(t, notification.CreatedAt.IsZero())
	assert.False(t, notification.UpdatedAt.IsZero())
	assert.Nil(t, notification.SentAt)
}

func TestNotification_GetMethods(t *testing.T) {
	notification := NewNotification(
		"user123",
		AppNotification,
		"Test Title",
		"Test Message",
		map[string]string{"key": "value"},
	)

	assert.Equal(t, notification.ID, notification.GetID())
	assert.Equal(t, "user123", notification.GetUserID())
	assert.Equal(t, "Test Title", notification.GetTitle())
}

func TestNotification_MarkAsSent(t *testing.T) {
	notification := NewNotification(
		"user123",
		AppNotification,
		"Test",
		"Message",
		nil,
	)

	originalUpdatedAt := notification.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	notification.MarkAsSent()

	assert.Equal(t, StatusSent, notification.Status)
	assert.NotNil(t, notification.SentAt)
	assert.True(t, notification.UpdatedAt.After(originalUpdatedAt))

	// SentAt and UpdatedAt should be very close
	timeDiff := notification.UpdatedAt.Sub(*notification.SentAt)
	assert.True(t, timeDiff < time.Millisecond)
}

func TestNotification_MarkAsRead(t *testing.T) {
	notification := NewNotification(
		"user123",
		AppNotification,
		"Test",
		"Message",
		nil,
	)

	originalUpdatedAt := notification.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	notification.MarkAsRead()

	assert.Equal(t, StatusRead, notification.Status)
	assert.True(t, notification.UpdatedAt.After(originalUpdatedAt))
}

func TestNotification_MarkAsFailed(t *testing.T) {
	notification := NewNotification(
		"user123",
		AppNotification,
		"Test",
		"Message",
		nil,
	)

	originalUpdatedAt := notification.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	notification.MarkAsFailed()

	assert.Equal(t, StatusFailed, notification.Status)
	assert.True(t, notification.UpdatedAt.After(originalUpdatedAt))
}

func TestNotification_AddChannel(t *testing.T) {
	notification := NewNotification(
		"user123",
		AppNotification,
		"Test",
		"Message",
		nil,
	)

	// Initially no channels
	assert.Empty(t, notification.Channels)

	// Add app channel
	appChannel := NewAppChannel("user123")
	notification.AddChannel(appChannel)

	assert.Len(t, notification.Channels, 1)
	assert.Equal(t, AppInternal, notification.Channels[0].GetType())

	// Add line channel
	lineChannel := NewLineChannel("user123", "line_user_123", "access_token")
	notification.AddChannel(lineChannel)

	assert.Len(t, notification.Channels, 2)
	assert.Equal(t, LineMessage, notification.Channels[1].GetType())
}

func TestNotification_AddMetadata(t *testing.T) {
	notification := NewNotification(
		"user123",
		AppNotification,
		"Test",
		"Message",
		map[string]string{"existing": "value"},
	)

	originalUpdatedAt := notification.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	notification.AddMetadata("new_key", "new_value")

	assert.Equal(t, "value", notification.Metadata["existing"])
	assert.Equal(t, "new_value", notification.Metadata["new_key"])
	assert.True(t, notification.UpdatedAt.After(originalUpdatedAt))
}

func TestNotification_AddMetadataToNilMap(t *testing.T) {
	notification := NewNotification(
		"user123",
		AppNotification,
		"Test",
		"Message",
		nil,
	)

	notification.AddMetadata("key", "value")

	require.NotNil(t, notification.Metadata)
	assert.Equal(t, "value", notification.Metadata["key"])
}

// Channel Tests
func TestNewAppChannel(t *testing.T) {
	userID := "user123"
	channel := NewAppChannel(userID)

	require.NotNil(t, channel)
	assert.Equal(t, userID, channel.UserID)
	assert.Equal(t, AppInternal, channel.GetType())
}

func TestNewLineChannel(t *testing.T) {
	userID := "user123"
	lineUserID := "line_user_456"
	accessToken := "access_token_789"

	channel := NewLineChannel(userID, lineUserID, accessToken)

	require.NotNil(t, channel)
	assert.Equal(t, userID, channel.UserID)
	assert.Equal(t, lineUserID, channel.LineUserID)
	assert.Equal(t, accessToken, channel.AccessToken)
	assert.Equal(t, LineMessage, channel.GetType())
}

// Notification Type Constants Test
func TestNotificationTypeConstants(t *testing.T) {
	assert.Equal(t, NotificationType("APP_NOTIFICATION"), AppNotification)
	assert.Equal(t, NotificationType("TASK_ASSIGNED"), TaskAssigned)
	assert.Equal(t, NotificationType("TASK_COMPLETED"), TaskCompleted)
	assert.Equal(t, NotificationType("TASK_DUE_SOON"), TaskDueSoon)
	assert.Equal(t, NotificationType("SYSTEM_NOTICE"), SystemNotice)
}

// Notification Status Constants Test
func TestNotificationStatusConstants(t *testing.T) {
	assert.Equal(t, NotificationStatus("PENDING"), StatusPending)
	assert.Equal(t, NotificationStatus("SENT"), StatusSent)
	assert.Equal(t, NotificationStatus("READ"), StatusRead)
	assert.Equal(t, NotificationStatus("FAILED"), StatusFailed)
}

// Channel Type Constants Test
func TestChannelTypeConstants(t *testing.T) {
	assert.Equal(t, ChannelType("APP_INTERNAL"), AppInternal)
	assert.Equal(t, ChannelType("LINE"), LineMessage)
}
