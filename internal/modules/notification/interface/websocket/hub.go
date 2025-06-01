package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/notification/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// Hub はWebSocketクライアントを管理するハブ（改良版）
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

	// 停止シグナル
	stopCh chan struct{}

	// 実行状態
	isRunning bool
	runningMu sync.RWMutex

	// ロガー
	logger logger.Logger

	// メトリクス
	metrics *HubMetrics
}

// HubMetrics はハブのメトリクス情報
type HubMetrics struct {
	TotalConnections    int64
	ActiveConnections   int64
	TotalNotifications  int64
	FailedNotifications int64
	mu                  sync.RWMutex
}

// NewHub はWebSocketハブを作成する（改良版）
func NewHub(logger logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client, 1000),               // バッファサイズ追加
		unregister: make(chan *Client, 1000),               // バッファサイズ追加
		broadcast:  make(chan *domain.Notification, 10000), // 大きなバッファ
		stopCh:     make(chan struct{}),
		logger:     logger,
		metrics:    &HubMetrics{},
	}
}

// Run はWebSocketハブを起動する（改良版、goroutineリーク対策）
func (h *Hub) Run(ctx context.Context) {
	h.runningMu.Lock()
	if h.isRunning {
		h.runningMu.Unlock()
		h.logger.Warn("Hub is already running")
		return
	}
	h.isRunning = true
	h.runningMu.Unlock()

	h.logger.Info("Starting WebSocket hub")

	// メトリクス収集用のticker
	metricsTicker := time.NewTicker(30 * time.Second)
	defer metricsTicker.Stop()

	// クリーンアップ用のticker
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	defer func() {
		h.runningMu.Lock()
		h.isRunning = false
		h.runningMu.Unlock()
		h.logger.Info("WebSocket hub stopped")
	}()

	for {
		select {
		case client := <-h.register:
			h.handleClientRegister(client)

		case client := <-h.unregister:
			h.handleClientUnregister(client)

		case notification := <-h.broadcast:
			h.handleBroadcast(notification)

		case <-metricsTicker.C:
			h.logMetrics()

		case <-cleanupTicker.C:
			h.cleanupStaleConnections()

		case <-h.stopCh:
			h.logger.Info("Hub stop signal received")
			return

		case <-ctx.Done():
			h.logger.Info("Hub context cancelled")
			return
		}
	}
}

// Stop はハブを停止
func (h *Hub) Stop() {
	h.runningMu.RLock()
	if !h.isRunning {
		h.runningMu.RUnlock()
		return
	}
	h.runningMu.RUnlock()

	select {
	case h.stopCh <- struct{}{}:
	default:
		// チャネルが満杯の場合はスキップ
	}
}

// RegisterClient はクライアントを登録
func (h *Hub) RegisterClient(client *Client) {
	select {
	case h.register <- client:
	default:
		h.logger.Error("Failed to register client: channel full",
			logger.Any("userID", client.UserID))
		// チャネルが満杯の場合は接続を閉じる
		client.Close()
	}
}

// UnregisterClient はクライアントの登録を解除
func (h *Hub) UnregisterClient(client *Client) {
	select {
	case h.unregister <- client:
	default:
		h.logger.Error("Failed to unregister client: channel full",
			logger.Any("userID", client.UserID))
	}
}

// SendNotification は指定ユーザーに通知を送信する（改良版）
func (h *Hub) SendNotification(notification *domain.Notification) {
	select {
	case h.broadcast <- notification:
		h.metrics.mu.Lock()
		h.metrics.TotalNotifications++
		h.metrics.mu.Unlock()
	default:
		h.logger.Error("Failed to send notification: broadcast channel full",
			logger.Any("notificationID", notification.ID))
		h.metrics.mu.Lock()
		h.metrics.FailedNotifications++
		h.metrics.mu.Unlock()
	}
}

// handleClientRegister はクライアント登録を処理
func (h *Hub) handleClientRegister(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true

	h.metrics.mu.Lock()
	h.metrics.TotalConnections++
	h.metrics.ActiveConnections++
	h.metrics.mu.Unlock()

	h.logger.Info("Client registered",
		logger.Any("userID", client.UserID),
		logger.Any("clientAddr", client.RemoteAddr()))
}

// handleClientUnregister はクライアント登録解除を処理
func (h *Hub) handleClientUnregister(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)

			// チャネルを安全に閉じる
			select {
			case <-client.send:
			default:
				close(client.send)
			}

			// ユーザーIDに対応するクライアントがなくなった場合、マップエントリを削除
			if len(clients) == 0 {
				delete(h.clients, client.UserID)
			}

			h.metrics.mu.Lock()
			h.metrics.ActiveConnections--
			h.metrics.mu.Unlock()

			h.logger.Info("Client unregistered",
				logger.Any("userID", client.UserID),
				logger.Any("clientAddr", client.RemoteAddr()))
		}
	}
}

// handleBroadcast は通知の配信を処理
func (h *Hub) handleBroadcast(notification *domain.Notification) {
	h.clientsMu.RLock()
	clients, exists := h.clients[notification.UserID]
	if !exists {
		h.clientsMu.RUnlock()
		h.logger.Debug("No clients connected for user",
			logger.Any("userID", notification.UserID))
		return
	}

	// マップをコピーして安全に反復処理
	clientList := make([]*Client, 0, len(clients))
	for client := range clients {
		clientList = append(clientList, client)
	}
	h.clientsMu.RUnlock()

	// 通知をJSONにシリアライズ
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		h.logger.Error("Failed to marshal notification",
			logger.Any("notificationID", notification.ID),
			logger.Error(err))
		return
	}

	// 各クライアントに送信
	for _, client := range clientList {
		select {
		case client.send <- notificationJSON:
			// 送信成功
		default:
			// 送信失敗（チャネルが満杯）、クライアントを切断
			h.logger.Warn("Client send channel full, disconnecting",
				logger.Any("userID", client.UserID))
			h.UnregisterClient(client)
		}
	}
}

// cleanupStaleConnections は古い接続をクリーンアップ
func (h *Hub) cleanupStaleConnections() {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for userID, clients := range h.clients {
		for client := range clients {
			if client.IsStale() {
				h.logger.Info("Cleaning up stale connection",
					logger.Any("userID", userID))
				delete(clients, client)
				close(client.send)
			}
		}

		if len(clients) == 0 {
			delete(h.clients, userID)
		}
	}
}

// logMetrics はメトリクスをログ出力
func (h *Hub) logMetrics() {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()

	h.logger.Info("WebSocket Hub Metrics",
		logger.Any("totalConnections", h.metrics.TotalConnections),
		logger.Any("activeConnections", h.metrics.ActiveConnections),
		logger.Any("totalNotifications", h.metrics.TotalNotifications),
		logger.Any("failedNotifications", h.metrics.FailedNotifications))
}

// GetMetrics はメトリクス情報を取得
func (h *Hub) GetMetrics() HubMetrics {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()
	return *h.metrics
}
