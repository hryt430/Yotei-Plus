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

// WebhookGateway はWebhook送信のゲートウェイ実装
type WebhookGateway struct {
	config     *config.Config
	httpClient *http.Client
	logger     logger.Logger
}

// NewWebhookGateway は新しいWebhookGatewayを作成する
func NewWebhookGateway(config *config.Config, logger logger.Logger) output.WebhookOutput {
	return &WebhookGateway{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendWebhook はWebhookを送信する
func (g *WebhookGateway) SendWebhook(ctx context.Context, event output.WebhookEvent, payload interface{}) error {
	// Webhookが設定されていない場合は何もしない
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
		g.logger.Error("Failed to marshal webhook payload", logger.Error(err))
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// リクエストの作成
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		g.logger.Error("Failed to create webhook request", logger.Error(err))
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// ヘッダーの設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", g.config.WebhookSecret)

	// リクエストの送信
	resp, err := g.httpClient.Do(req)
	if err != nil {
		g.logger.Error("Failed to send webhook", logger.Error(err))
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスの確認
	if resp.StatusCode != http.StatusOK {
		g.logger.Error("Webhook endpoint returned non-OK status", logger.Any("status", resp.Status))
		return fmt.Errorf("webhook endpoint returned non-OK status: %s", resp.Status)
	}

	g.logger.Info("Successfully sent webhook", logger.Any("event", event))
	return nil
}
