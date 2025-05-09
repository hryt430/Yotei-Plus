package messaging

import (
	"encoding/json"
	"fmt"
	"log"
)

// MessageType はメッセージの種類を定義します
type MessageType string

const (
	// TaskCreated はタスク作成イベントを表します
	TaskCreated MessageType = "task.created"
	// TaskAssigned はタスク割り当てイベントを表します
	TaskAssigned MessageType = "task.assigned"
	// TaskStatusChanged はタスクステータス変更イベントを表します
	TaskStatusChanged MessageType = "task.status_changed"
	// TaskCommentAdded はタスクコメント追加イベントを表します
	TaskCommentAdded MessageType = "task.comment_added"
)

// Message はメッセージングシステムを通じて送信されるメッセージを表します
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client はメッセージングシステムとの通信を行うクライアントです
// 実際のアプリケーションではRabbitMQ、Kafka、Redis Pubsubなどを使用します
type Client struct {
	// 実際の実装では接続情報やクライアントインスタンスが入ります
	subscribers map[MessageType][]func([]byte) error
}

// NewClient は新しいメッセージングクライアントを作成します
func NewClient() *Client {
	return &Client{
		subscribers: make(map[MessageType][]func([]byte) error),
	}
}

// Publish はメッセージを公開します
func (c *Client) Publish(msgType MessageType, payload interface{}) error {
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}

	// JSONにエンコード
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("メッセージのJSONエンコードエラー: %w", err)
	}

	// 実際のアプリケーションではここでメッセージングシステムに送信
	log.Printf("メッセージ送信: %s", string(data))

	// 登録されたサブスクライバーに通知
	for _, handler := range c.subscribers[msgType] {
		if err := handler(data); err != nil {
			log.Printf("メッセージハンドラーエラー: %v", err)
		}
	}

	return nil
}

// Subscribe はメッセージタイプに対するサブスクライバーを登録します
func (c *Client) Subscribe(msgType MessageType, handler func([]byte) error) {
	c.subscribers[msgType] = append(c.subscribers[msgType], handler)
}

// Close はメッセージングクライアントの接続を閉じます
func (c *Client) Close() error {
	// 実際の実装では接続のクローズ処理を行います
	return nil
}
