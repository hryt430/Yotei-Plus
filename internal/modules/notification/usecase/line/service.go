package line

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"your-app/notification/domain/entity"
)

// LINE APIのエンドポイント
const (
	lineMessageAPIURL = "https://api.line.me/v2/bot/message/push"
)

// LineMessage はLINE APIに送信するメッセージ構造体
type LineMessage struct {
	To       string           `json:"to"`
	Messages []LineMessageObj `json:"messages"`
}

// LineMessageObj はLINE APIに送信するメッセージオブジェクト
type LineMessageObj struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// LineService はLINE通知に関するロジックを定義するインターフェース
type LineService interface {
	// SendMessage はLINEにメッセージを送信する
	SendMessage(ctx context.Context, channel *entity.LineChannel, message string) error
}

// DefaultLineService はLINEサービスの基本実装
type DefaultLineService struct {
	httpClient *http.Client
	botToken   string // LINE Bot API Token
}

// NewLineService はLineServiceのインスタンスを作成する
func NewLineService(botToken string) LineService {
	return &DefaultLineService{
		httpClient: &http.Client{},
		botToken:   botToken,
	}
}

// SendMessage はLINEにメッセージを送信する
func (s *DefaultLineService) SendMessage(ctx context.Context, channel *entity.LineChannel, message string) error {
	if channel.LineUserID == "" {
		return errors.New("line user id is empty")
	}

	lineMsg := LineMessage{
		To: channel.LineUserID,
		Messages: []LineMessageObj{
			{
				Type: "text",
				Text: message,
			},
		},
	}

	jsonData, err := json.Marshal(lineMsg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", lineMessageAPIURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.botToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send LINE message, status: " + resp.Status)
	}

	return nil
}
