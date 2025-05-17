package websocket

import (
	"encoding/json"
	"sync"

	"your-app/notification/domain/entity"
)

// Hub はWebSocketクライアントを管理するハブ
type Hub struct {
	// クライアントマップ（キー：ユーザーID）
	clients   map[uint]map[*Client]bool
	clientsMu sync.RWMutex

	// クライアント登録チャネル
	register chan *Client

	// クライアント登録解除チャネル
	unregister chan *Client

	// 通知送信チャネル
	broadcast chan *entity.Notification
}

// NewHub はWebSocketハブを作成する
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *entity.Notification),
	}
}

// Run はWebSocketハブを起動する
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clientsMu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.clientsMu.Unlock()

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

		case notification := <-h.broadcast:
			// 通知対象ユーザーのクライアント全てに送信
			h.clientsMu.RLock()
			if clients, ok := h.clients[notification.UserID]; ok {
				notificationJSON, err := json.Marshal(notification)
				if err != nil {
					continue
				}

				for client := range clients {
					select {
					case client.send <- notificationJSON:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.clientsMu.RUnlock()
		}
	}
}

// SendNotification は指定ユーザーに通知を送信する
func (h *Hub) SendNotification(notification *entity.Notification) {
	h.broadcast <- notification
}
