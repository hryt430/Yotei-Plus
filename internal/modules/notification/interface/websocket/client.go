package websocket

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hryt430/Yotei+/pkg/logger"
	"go.uber.org/zap"
)

const (
	// クライアント→サーバーのメッセージ読み取りタイムアウト
	pongWait = 60 * time.Second

	// サーバー→クライアントのPing送信間隔
	pingPeriod = (pongWait * 9) / 10

	// メッセージの最大サイズ
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 開発用に全てのオリジンを許可（本番環境では適切に制限すること）
	},
}

// Client はWebSocketクライアント（ロガー追加版）
type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	UserID      string
	remoteAddr  string
	logger      logger.Logger
	connectedAt time.Time
}

// NewClient は新しいWebSocketクライアントを作成（ロガー追加版）
func NewClient(hub *Hub, conn *websocket.Conn, userID string, logger logger.Logger) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		UserID:      userID,
		remoteAddr:  conn.RemoteAddr().String(),
		logger:      logger,
		connectedAt: time.Now(),
	}
}

// ReadPump はクライアントからのメッセージ読み取りループ（ロガー追加版）
func (c *Client) ReadPump() {
	defer func() {
		c.logger.Info("Read pump stopped, cleaning up",
			zap.Any("userID", c.UserID),
			zap.Any("remoteAddr", c.remoteAddr),
			zap.Any("connectionDuration", time.Since(c.connectedAt)))
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.logger.Debug("Pong received", zap.Any("userID", c.UserID))
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	c.logger.Info("Starting read pump",
		zap.Any("userID", c.UserID),
		zap.Any("remoteAddr", c.remoteAddr))

	// クライアント側からの切断を検知するための無限ループ
	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error",
					zap.Any("userID", c.UserID),
					zap.Any("remoteAddr", c.remoteAddr),
					zap.Error(err))
			} else {
				c.logger.Info("Client disconnected normally",
					zap.Any("userID", c.UserID),
					zap.Any("remoteAddr", c.remoteAddr))
			}
			break
		}

		// メッセージを受信した場合のログ（デバッグ用）
		c.logger.Debug("Message received from client",
			zap.Any("userID", c.UserID),
			zap.Any("messageType", messageType),
			zap.Any("messageSize", len(message)))

		// 実際にはクライアントからのメッセージ処理は今回必要ないが、
		// 接続維持のためにReadMessageを呼び出す
	}
}

// WritePump はクライアントへのメッセージ送信ループ（ロガー追加版）
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.logger.Info("Write pump stopped",
			zap.Any("userID", c.UserID),
			zap.Any("remoteAddr", c.remoteAddr))
		c.conn.Close()
	}()

	c.logger.Info("Starting write pump",
		zap.Any("userID", c.UserID),
		zap.Any("remoteAddr", c.remoteAddr))

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// ハブがチャネルを閉じた
				c.logger.Info("Send channel closed by hub",
					zap.Any("userID", c.UserID))
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.logger.Error("Failed to get next writer",
					zap.Any("userID", c.UserID),
					zap.Error(err))
				return
			}
			w.Write(message)

			// キューに残っているメッセージがあれば、追加で送信
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				c.logger.Error("Failed to close writer",
					zap.Any("userID", c.UserID),
					zap.Error(err))
				return
			}

			c.logger.Debug("Message sent to client",
				zap.Any("userID", c.UserID),
				zap.Any("messageSize", len(message)),
				zap.Any("batchSize", n+1))

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Error("Failed to send ping",
					zap.Any("userID", c.UserID),
					zap.Error(err))
				return
			}
			c.logger.Debug("Ping sent", zap.Any("userID", c.UserID))
		}
	}
}

// ServeWs はWebSocket接続をハンドリングする（ロガー追加版）
func ServeWs(hub *Hub, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ミドルウェアで設定されたユーザーIDを取得
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Warn("WebSocket connection without user authentication",
				zap.Any("remoteAddr", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		userIDStr := userID.(string)
		logger.Info("WebSocket upgrade request",
			zap.Any("userID", userIDStr),
			zap.Any("remoteAddr", c.ClientIP()))

		// WebSocketにアップグレード
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("Failed to upgrade to WebSocket",
				zap.Any("userID", userIDStr),
				zap.Any("remoteAddr", c.ClientIP()),
				zap.Error(err))
			return
		}

		logger.Info("WebSocket connection established",
			zap.Any("userID", userIDStr),
			zap.Any("remoteAddr", conn.RemoteAddr().String()))

		// クライアント作成
		client := NewClient(hub, conn, userIDStr, logger)
		client.hub.register <- client

		// クライアントのループを開始
		go client.WritePump()
		go client.ReadPump()
	}
}
