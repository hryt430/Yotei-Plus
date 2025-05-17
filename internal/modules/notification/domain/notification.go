package domain

import (
	"time"
)

// NotificationType は通知の種類を表す
type NotificationType string

const (
	TaskAssigned  NotificationType = "TASK_ASSIGNED"  // タスク割り当て
	TaskCompleted NotificationType = "TASK_COMPLETED" // タスク完了
	TaskDueSoon   NotificationType = "TASK_DUE_SOON"  // タスク期限間近
	SystemNotice  NotificationType = "SYSTEM_NOTICE"  // システムからの通知
)

// NotificationStatus は通知の状態を表す
type NotificationStatus string

const (
	StatusUnread  NotificationStatus = "UNREAD"  // 未読
	StatusRead    NotificationStatus = "READ"    // 既読え
	StatusDeleted NotificationStatus = "DELETED" // 削除済み
)

// Notification は通知情報を保持するエンティティ
type Notification struct {
	ID        uint                   `json:"id"`
	UserID    uint                   `json:"user_id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Status    NotificationStatus     `json:"status"`
	RelatedID *uint                  `json:"related_id,omitempty"` // 関連するリソースのID（タスクIDなど）
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Channels  []Channel              `json:"-"` // 送信チャネルのリスト
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

// NewNotification は新しい通知エンティティを作成する
func NewNotification(
	userID uint,
	notificationType NotificationType,
	title string,
	content string,
	relatedID *uint,
	metadata map[string]interface{},
) *Notification {
	now := time.Now()
	return &Notification{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Content:   content,
		Status:    StatusUnread,
		RelatedID: relatedID,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MarkAsRead は通知を既読にする
func (n *Notification) MarkAsRead() {
	if n.Status == StatusUnread {
		n.Status = StatusRead
		now := time.Now()
		n.ReadAt = &now
		n.UpdatedAt = now
	}
}

// Delete は通知を削除する
func (n *Notification) Delete() {
	n.Status = StatusDeleted
	n.UpdatedAt = time.Time{}
}

// AddChannel は通知に送信チャネルを追加する
func (n *Notification) AddChannel(channel Channel) {
	n.Channels = append(n.Channels, channel)
}
