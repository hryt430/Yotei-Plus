package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hryt430/task-management/pkg/logger"

	"github.com/hryt430/task-management/config"
	"github.com/hryt430/task-management/internal/modules/notification/usecase/output"
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
func NewLineGateway(config *config.Config, logger logger.Logger) *LineGateway {
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
		lineUserID = userID
	}

	return g.SendLineNotification(ctx, lineUserID, message)
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
		g.logger.Error("Failed to marshal LINE message", "error", err)
		return fmt.Errorf("failed to marshal LINE message: %w", err)
	}

	// リクエストの作成
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		g.logger.Error("Failed to create HTTP request", "error", err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// ヘッダーの設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.config.LineChannelToken)

	// リクエストの送信
	resp, err := g.httpClient.Do(req)
	if err != nil {
		g.logger.Error("Failed to send LINE notification", "error", err)
		return fmt.Errorf("failed to send LINE notification: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスの確認
	if resp.StatusCode != http.StatusOK {
		g.logger.Error("LINE API returned non-OK status", "status", resp.Status)
		return fmt.Errorf("LINE API returned non-OK status: %s", resp.Status)
	}

	g.logger.Info("Successfully sent LINE notification", "lineUserID", lineUserID)
	return nil
}

// LineWebhookHandler はLINE Webhookのハンドラ
func (g *LineGateway) LineWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Webhookリクエストの検証
	signature := r.Header.Get("X-Line-Signature")
	if signature == "" {
		g.logger.Error("Missing LINE signature")
		http.Error(w, "Missing signature", http.StatusBadRequest)
		return
	}

	// リクエストボディの読み取り
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		g.logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// ここでLINE Webhookの処理を実装
	// 例: 友達追加イベントやメッセージイベントの処理

	// 正常応答
	w.WriteHeader(http.StatusOK)
}

// LineWebhookOutput はLINE Webhook出力の実装
type LineWebhookOutput struct {
	config     *config.Config
	httpClient *http.Client
	logger     logger.Logger
}

// NewLineWebhookOutput は新しいLineWebhookOutputを作成する
func NewLineWebhookOutput(config *config.Config, logger logger.Logger) output.WebhookOutput {
	return &LineWebhookOutput{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendWebhook はWebhookを送信する
func (g *LineWebhookOutput) SendWebhook(ctx context.Context, event output.WebhookEvent, payload interface{}) error {
	// ユーザーが設定したWebhookURLがある場合のみ送信
	if g.config.WebhookURL == "" {
		return nil
	}

	// ペイロードの構築
	webhookPayload := map[string]interface{}{
		"event":     event,
		"payload":   payload,
		"timestamp": time.Now().Unix(),
	}

	// JSONに変換
	jsonData, err := json.Marshal(webhookPayload)
	if err != nil {
		g.logger.Error("Failed to marshal webhook payload", "error", err)
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// リクエストの作成
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		g.logger.Error("Failed to create webhook request", "error", err)
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// ヘッダーの設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", g.config.WebhookSecret)

	// リクエストの送信
	resp, err := g.httpClient.Do(req)
	if err != nil {
		g.logger.Error("Failed to send webhook", "error", err)
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスの確認
	if resp.StatusCode != http.StatusOK {
		g.logger.Error("Webhook endpoint returned non-OK status", "status", resp.Status)
		return fmt.Errorf("webhook endpoint returned non-OK status: %s", resp.Status)
	}

	g.logger.Info("Successfully sent webhook", "event", event)
	return nil
}
