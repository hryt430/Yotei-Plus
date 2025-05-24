package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/modules/notification/usecase/output"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// LineMessage はLINE Messaging APIに送信するメッセージ形式
type LineMessage struct {
	To       string              `json:"to"`
	Messages []LineMessageDetail `json:"messages"`
}

// LineMessageDetail はLINEメッセージの詳細
type LineMessageDetail struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// LineGateway はLINE通知のゲートウェイ実装
type LineGateway struct {
	config     *config.Config
	httpClient *http.Client
	logger     logger.Logger
}

// NewLineGateway は新しいLineGatewayを作成する
func NewLineGateway(config *config.Config, logger logger.Logger) output.LineNotificationGateway {
	return &LineGateway{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendNotification は通知を送信する
func (g *LineGateway) SendNotification(ctx context.Context, userID, title, message string, metadata map[string]string) error {
	lineUserID, ok := metadata["line_user_id"]
	if !ok {
		g.logger.Warn("LINE user ID not found in metadata", logger.Any("userID", userID))
		return nil // LINEユーザーIDがない場合は何もしない
	}

	return g.SendLineNotification(ctx, lineUserID, title+"\n"+message)
}

// SendLineNotification はLINE通知を送信する
func (g *LineGateway) SendLineNotification(ctx context.Context, lineUserID, message string) error {
	// LINE Messaging APIのエンドポイント
	url := "https://api.line.me/v2/bot/message/push"

	// メッセージの構築
	lineMsg := LineMessage{
		To: lineUserID,
		Messages: []LineMessageDetail{
			{
				Type: "text",
				Text: message,
			},
		},
	}

	// JSONに変換
	jsonData, err := json.Marshal(lineMsg)
	if err != nil {
		g.logger.Error("Failed to marshal LINE message", logger.Error(err))
		return fmt.Errorf("failed to marshal LINE message: %w", err)
	}

	// リクエストの作成
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		g.logger.Error("Failed to create HTTP request", logger.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// ヘッダーの設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.config.External.LineChannelToken)

	// リクエストの送信
	resp, err := g.httpClient.Do(req)
	if err != nil {
		g.logger.Error("Failed to send LINE notification", logger.Error(err))
		return fmt.Errorf("failed to send LINE notification: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスの確認
	if resp.StatusCode != http.StatusOK {
		g.logger.Error("LINE API returned non-OK status", logger.Any("status", resp.Status))
		return fmt.Errorf("LINE API returned non-OK status: %s", resp.Status)
	}

	g.logger.Info("Successfully sent LINE notification", logger.Any("lineUserID", lineUserID))
	return nil
}
