package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType は通知の種類を表す
type NotificationType string

const (
	AppNotification NotificationType = "APP_NOTIFICATION" // アプリ内通知
	TaskAssigned    NotificationType = "TASK_ASSIGNED"    // タスク割り当て
	TaskCompleted   NotificationType = "TASK_COMPLETED"   // タスク完了
	TaskDueSoon     NotificationType = "TASK_DUE_SOON"    // タスク期限間近
	SystemNotice    NotificationType = "SYSTEM_NOTICE"    // システムからの通知
)

// NotificationStatus は通知の状態を表す
type NotificationStatus string

const (
	StatusPending NotificationStatus = "PENDING" // 保留中
	StatusSent    NotificationStatus = "SENT"    // 送信済み
	StatusRead    NotificationStatus = "READ"    // 既読
	StatusFailed  NotificationStatus = "FAILED"  // 送信失敗
)

// Notification は通知情報を保持するエンティティ
type Notification struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	Type      NotificationType   `json:"type"`
	Title     string             `json:"title"`
	Message   string             `json:"message"`
	Status    NotificationStatus `json:"status"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
	Channels  []Channel          `json:"-"` // 送信チャネルのリスト
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	SentAt    *time.Time         `json:"sent_at,omitempty"`
}

// NewNotification は新しい通知エンティティを作成する
func NewNotification(
	userID string,
	notificationType NotificationType,
	title string,
	message string,
	metadata map[string]string,
) *Notification {
	now := time.Now()
	return &Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Status:    StatusPending,
		Metadata:  metadata,
		Channels:  []Channel{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MarkAsSent は通知を送信済みにする
func (n *Notification) MarkAsSent() {
	n.Status = StatusSent
	now := time.Now()
	n.SentAt = &now
	n.UpdatedAt = now
}

// MarkAsRead は通知を既読にする
func (n *Notification) MarkAsRead() {
	n.Status = StatusRead
	n.UpdatedAt = time.Now()
}

// MarkAsFailed は通知を送信失敗にする
func (n *Notification) MarkAsFailed() {
	n.Status = StatusFailed
	n.UpdatedAt = time.Now()
}

// AddChannel は通知に送信チャネルを追加する
func (n *Notification) AddChannel(channel Channel) {
	n.Channels = append(n.Channels, channel)
}

// AddMetadata はメタデータに項目を追加する
func (n *Notification) AddMetadata(key, value string) {
	if n.Metadata == nil {
		n.Metadata = make(map[string]string)
	}
	n.Metadata[key] = value
	n.UpdatedAt = time.Now()
}
