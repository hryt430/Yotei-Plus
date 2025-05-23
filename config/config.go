package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config はアプリケーション全体の設定を保持する
type Config struct {
	// サーバー設定
	Port        int    `mapstructure:"PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`

	// データベース設定
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     int    `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`

	// CORS設定
	AllowedOrigins string `mapstructure:"ALLOWED_ORIGINS"`

	// セキュリティ設定
	JWTSecret             string `mapstructure:"JWT_SECRET"`
	JWTExpirationHours    int    `mapstructure:"JWT_EXPIRATION_HOURS"`
	RefreshExpirationDays int    `mapstructure:"REFRESH_EXPIRATION_DAYS"`
	EnabledCSRF           bool   `mapstructure:"ENABLE_CSRF"`

	// 通知設定
	WebhookURL    string `mapstructure:"WEBHOOK_URL"`
	WebhookSecret string `mapstructure:"WEBHOOK_SECRET"`

	// LINE設定
	LineChannelToken string `mapstructure:"LINE_CHANNEL_TOKEN"`
}

// LoadConfig は設定ファイルを読み込む
func LoadConfig(path string) (*Config, error) {
	// Viperの設定
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// デフォルト値の設定
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 3306)
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("ALLOWED_ORIGINS", "*")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 1)
	viper.SetDefault("REFRESH_EXPIRATION_DAYS", 7)
	viper.SetDefault("ENABLE_CSRF", true)

	// 設定ファイルの読み込み
	if err := viper.ReadInConfig(); err != nil {
		// 設定ファイルが見つからなくても環境変数で上書きする可能性があるためエラーとしない
		fmt.Printf("Warning: 設定ファイルが見つかりません。環境変数を使用します: %v\n", err)
	}

	// 設定構造体にマッピング
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("設定の読み込みに失敗しました: %w", err)
	}

	// 必須の設定値を検証
	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET が設定されていません")
	}

	return &config, nil
}

// GetDSN はデータベース接続文字列を返す
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// IsProduction は本番環境かどうかを返す
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetAllowedOrigins は許可されたオリジンを返す
func (c *Config) GetAllowedOrigins() string {
	return c.AllowedOrigins
}

// GetAllowedOriginsSlice は許可されたオリジンをスライスとして返す
func (c *Config) GetAllowedOriginsSlice() []string {
	if c.AllowedOrigins == "*" {
		return []string{"*"}
	}
	return strings.Split(c.AllowedOrigins, ",")
}

// EnableCSRF はCSRF保護を有効にするかどうかを返す
func (c *Config) EnableCSRF() bool {
	return c.EnabledCSRF
}

// 設定ファイルのサンプル作成
func CreateSampleConfig() error {
	sampleConfig := []byte(`# アプリケーション設定
PORT=8080
ENVIRONMENT=development # development, production
LOG_LEVEL=info # debug, info, warn, error

# データベース設定
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=yotei_plus

# CORS設定
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# セキュリティ設定
JWT_SECRET=your-jwt-secret-key-here
JWT_EXPIRATION_HOURS=1
REFRESH_EXPIRATION_DAYS=7
ENABLE_CSRF=true

# 通知設定
WEBHOOK_URL=https://your-webhook-url.com
WEBHOOK_SECRET=your-webhook-secret

# LINE設定
LINE_CHANNEL_TOKEN=your-line-channel-token
`)

	return os.WriteFile("app.env.example", sampleConfig, 0644)
}
