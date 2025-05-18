package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hryt430/Yotei+/pkg/logger"
)

// MessageBroker はメッセージブローカーのインターフェース
type MessageBroker interface {
	// Publish はトピックにメッセージを公開する
	Publish(ctx context.Context, topic string, message []byte) error

	// Subscribe はトピックからメッセージを購読する
	Subscribe(ctx context.Context, topic string, handler func([]byte) error) error

	// Close はブローカー接続を閉じる
	Close() error
}

// InMemoryMessageBroker はメモリ上で動作するシンプルなメッセージブローカー
// 開発・テスト用途に適しています
type InMemoryMessageBroker struct {
	subscribers map[string][]func([]byte) error
	logger      logger.Logger
}

// NewInMemoryMessageBroker は新しいInMemoryMessageBrokerを作成する
func NewInMemoryMessageBroker(logger logger.Logger) MessageBroker {
	return &InMemoryMessageBroker{
		subscribers: make(map[string][]func([]byte) error),
		logger:      logger,
	}
}

// Publish はトピックにメッセージを公開する
func (b *InMemoryMessageBroker) Publish(ctx context.Context, topic string, message []byte) error {
	b.logger.Info("Publishing message", "topic", topic, "size", len(message))

	for _, handler := range b.subscribers[topic] {
		if err := handler(message); err != nil {
			b.logger.Error("Error handling message", "topic", topic, "error", err)
		}
	}

	return nil
}

// Subscribe はトピックからメッセージを購読する
func (b *InMemoryMessageBroker) Subscribe(ctx context.Context, topic string, handler func([]byte) error) error {
	b.logger.Info("Subscribing to topic", "topic", topic)
	b.subscribers[topic] = append(b.subscribers[topic], handler)
	return nil
}

// Close はブローカー接続を閉じる
func (b *InMemoryMessageBroker) Close() error {
	b.subscribers = make(map[string][]func([]byte) error)
	return nil
}

// WebSocketMessage はWebSocketに送信するメッセージ
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSMessageBroker はWebSocket向けのメッセージブローカー
type WSMessageBroker struct {
	broker MessageBroker
	logger logger.Logger
}

// NewWSMessageBroker は新しいWSMessageBrokerを作成する
func NewWSMessageBroker(broker MessageBroker, logger logger.Logger) *WSMessageBroker {
	return &WSMessageBroker{
		broker: broker,
		logger: logger,
	}
}

// PublishToUser は特定ユーザー向けにメッセージを公開する
func (b *WSMessageBroker) PublishToUser(ctx context.Context, userID string, messageType string, payload interface{}) error {
	message := WebSocketMessage{
		Type:    messageType,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	topic := fmt.Sprintf("user.%s.notifications", userID)
	return b.broker.Publish(ctx, topic, data)
}

// PublishNotification は通知メッセージを公開する
func (b *WSMessageBroker) PublishNotification(ctx context.Context, userID string, notification interface{}) error {
	return b.PublishToUser(ctx, userID, "notification", notification)
}

// PublishNotificationRead は通知既読メッセージを公開する
func (b *WSMessageBroker) PublishNotificationRead(ctx context.Context, userID string, notificationID string) error {
	payload := map[string]interface{}{
		"notification_id": notificationID,
		"read_at":         true,
	}
	return b.PublishToUser(ctx, userID, "notification_read", payload)
}

// SubscribeToUserNotifications はユーザー通知を購読する
func (b *WSMessageBroker) SubscribeToUserNotifications(ctx context.Context, userID string, handler func([]byte) error) error {
	topic := fmt.Sprintf("user.%s.notifications", userID)
	return b.broker.Subscribe(ctx, topic, handler)
}

// Close はブローカー接続を閉じる
func (b *WSMessageBroker) Close() error {
	return b.broker.Close()
}
