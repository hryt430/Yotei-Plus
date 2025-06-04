package websocket

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// Hub はWebSocketクライアントを管理するハブ
type Hub struct {
	// クライアントマップ（キー：ユーザーID）
	clients   map[string]map[*Client]bool
	clientsMu sync.RWMutex

	// クライアント登録チャネル
	register chan *Client

	// クライアント登録解除チャネル
	unregister chan *Client

	// 通知送信チャネル
	broadcast chan *domain.Notification

	// ロガー
	logger logger.Logger
}

// NewHub はWebSocketハブを作成する
func NewHub(logger logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *domain.Notification),
		logger:     logger,
	}
}

// Run はWebSocketハブを起動する
func (h *Hub) Run(ctx context.Context) error {
	h.logger.Info("Starting WebSocket hub")

	// 停止時のクリーンアップ用
	defer h.cleanup()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("WebSocket hub stopping due to context cancellation")
			return ctx.Err()

		case client := <-h.register:
			h.clientsMu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.clientsMu.Unlock()

			h.logger.Info("Client registered",
				logger.Any("userID", client.UserID),
				logger.Any("totalClients", len(h.clients)))

		case client := <-h.unregister:
			h.clientsMu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients[client.UserID], client)
				close(client.send)

				// ユーザーIDに対応するクライアントがなくなった場合、マップエントリを削除
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.clientsMu.Unlock()

			h.logger.Info("Client unregistered",
				logger.Any("userID", client.UserID),
				logger.Any("totalClients", len(h.clients)))

		case notification := <-h.broadcast:
			// context がキャンセルされている場合は処理をスキップ
			if ctx.Err() != nil {
				h.logger.Debug("Skipping notification broadcast due to context cancellation",
					logger.Any("notificationID", notification.ID))
				continue
			}

			// 通知対象ユーザーのクライアント全てに送信
			h.clientsMu.RLock()
			if clients, ok := h.clients[notification.UserID]; ok {
				notificationJSON, err := json.Marshal(notification)
				if err != nil {
					h.logger.Error("Failed to marshal notification",
						logger.Any("notificationID", notification.ID),
						logger.Error(err))
					h.clientsMu.RUnlock()
					continue
				}

				sentCount := 0
				failedCount := 0

				for client := range clients {
					select {
					case client.send <- notificationJSON:
						sentCount++
					default:
						h.logger.Warn("Client send channel full, closing connection",
							logger.Any("userID", client.UserID))
						close(client.send)
						delete(clients, client)
						failedCount++
					}
				}

				h.logger.Info("Notification sent",
					logger.Any("notificationID", notification.ID),
					logger.Any("userID", notification.UserID),
					logger.Any("sentCount", sentCount),
					logger.Any("failedCount", failedCount))
			} else {
				h.logger.Debug("No clients connected for notification",
					logger.Any("notificationID", notification.ID),
					logger.Any("userID", notification.UserID))
			}
			h.clientsMu.RUnlock()
		}
	}
}

// cleanup は停止時のクリーンアップ処理を行う
func (h *Hub) cleanup() {
	h.logger.Info("Cleaning up WebSocket hub")

	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	// 全クライアントを閉じる
	totalClients := 0
	for userID, clients := range h.clients {
		for client := range clients {
			close(client.send)
			totalClients++
		}
		delete(h.clients, userID)
	}

	h.logger.Info("WebSocket hub cleanup completed",
		logger.Any("closedClients", totalClients))
}

// SendNotification は指定ユーザーに通知を送信する
func (h *Hub) SendNotification(notification *domain.Notification) {
	h.logger.Debug("Queueing notification for broadcast",
		logger.Any("notificationID", notification.ID),
		logger.Any("userID", notification.UserID))

	// ノンブロッキングで送信を試行
	select {
	case h.broadcast <- notification:
		// 正常に送信
	default:
		h.logger.Warn("Broadcast channel full, dropping notification",
			logger.Any("notificationID", notification.ID),
			logger.Any("userID", notification.UserID))
	}
}
