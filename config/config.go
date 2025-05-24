package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config はアプリケーション設定を格納する構造体
type Config struct {
	Environment string   `mapstructure:"ENVIRONMENT"`
	Server      Server   `mapstructure:",squash"`
	Database    Database `mapstructure:",squash"`
	Redis       Redis    `mapstructure:",squash"`
	JWT         JWT      `mapstructure:",squash"`
	CORS        CORS     `mapstructure:",squash"`
	Security    Security `mapstructure:",squash"`
	Log         Log      `mapstructure:",squash"`
	External    External `mapstructure:",squash"`
}

// Server はサーバー設定
type Server struct {
	Host           string `mapstructure:"SERVER_HOST"`
	Port           string `mapstructure:"SERVER_PORT"`
	MaxRequestSize int64  `mapstructure:"MAX_REQUEST_SIZE"`
	ReadTimeout    int    `mapstructure:"READ_TIMEOUT"`
	WriteTimeout   int    `mapstructure:"WRITE_TIMEOUT"`
}

// Database はデータベース設定
type Database struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSL      bool   `mapstructure:"DB_SSL"`
	TimeZone string `mapstructure:"DB_TIMEZONE"`
}

// Redis はRedis設定
type Redis struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       string `mapstructure:"REDIS_DB"`
	URL      string `mapstructure:"REDIS_URL"`
}

// JWT はJWT設定
type JWT struct {
	SecretKey            string `mapstructure:"JWT_SECRET_KEY"`
	AccessTokenDuration  string `mapstructure:"JWT_ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration string `mapstructure:"JWT_REFRESH_TOKEN_DURATION"`
	Issuer               string `mapstructure:"JWT_ISSUER"`
}

// CORS はCORS設定
type CORS struct {
	AllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	AllowedMethods string `mapstructure:"CORS_ALLOWED_METHODS"`
	AllowedHeaders string `mapstructure:"CORS_ALLOWED_HEADERS"`
}

// Security はセキュリティ設定
type Security struct {
	EnableCSRF    bool   `mapstructure:"ENABLE_CSRF"`
	RateLimitRPS  int    `mapstructure:"RATE_LIMIT_RPS"`
	SessionSecret string `mapstructure:"SESSION_SECRET"`
}

// Log はログ設定
type Log struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
	Output string `mapstructure:"LOG_OUTPUT"`
}

// External は外部サービス設定
type External struct {
	LineChannelToken  string `mapstructure:"LINE_CHANNEL_TOKEN"`
	LineChannelSecret string `mapstructure:"LINE_CHANNEL_SECRET"`
	WebhookURL        string `mapstructure:"WEBHOOK_URL"`
	WebhookSecret     string `mapstructure:"WEBHOOK_SECRET"`
}

// LoadConfig は設定を環境変数から読み込みます
func LoadConfig(path string) (*Config, error) {
	// .envファイルの読み込み（存在する場合）
	if path != "" {
		if err := godotenv.Load(path + "/.env"); err != nil {
			// .envファイルがない場合はエラーにしない
		}
	}

	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: Server{
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			Port:           getEnv("SERVER_PORT", "8080"),
			MaxRequestSize: getEnvAsInt64("MAX_REQUEST_SIZE", 10<<20), // 10MB
			ReadTimeout:    getEnvAsInt("READ_TIMEOUT", 30),
			WriteTimeout:   getEnvAsInt("WRITE_TIMEOUT", 30),
		},
		Database: Database{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "task_management"),
			SSL:      getEnvAsBool("DB_SSL", false),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Tokyo"),
		},
		Redis: Redis{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnv("REDIS_DB", "0"),
			URL:      getEnv("REDIS_URL", ""),
		},
		JWT: JWT{
			SecretKey:            getEnv("JWT_SECRET_KEY", "your-secret-key"),
			AccessTokenDuration:  getEnv("JWT_ACCESS_TOKEN_DURATION", "1h"),
			RefreshTokenDuration: getEnv("JWT_REFRESH_TOKEN_DURATION", "168h"),
			Issuer:               getEnv("JWT_ISSUER", "app"),
		},
		CORS: CORS{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001"),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-CSRF-Token"),
		},
		Security: Security{
			EnableCSRF:    getEnvAsBool("ENABLE_CSRF", false),
			RateLimitRPS:  getEnvAsInt("RATE_LIMIT_RPS", 100),
			SessionSecret: getEnv("SESSION_SECRET", "session-secret"),
		},
		Log: Log{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		External: External{
			LineChannelToken:  getEnv("LINE_CHANNEL_TOKEN", ""),
			LineChannelSecret: getEnv("LINE_CHANNEL_SECRET", ""),
			WebhookURL:        getEnv("WEBHOOK_URL", ""),
			WebhookSecret:     getEnv("WEBHOOK_SECRET", ""),
		},
	}

	return config, nil
}

// GetDSN はデータベース接続文字列を取得します
func (c *Config) GetDSN() string {
	ssl := "false"
	if c.Database.SSL {
		ssl = "true"
	}

	user := url.QueryEscape(c.Database.User)
	pass := url.QueryEscape(c.Database.Password)
	name := url.QueryEscape(c.Database.Name)
	tz := url.QueryEscape(c.Database.TimeZone) // "Asia/Tokyo" → "Asia%2FTokyo"

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=%s&tls=%s",
		user, pass,
		c.Database.Host, c.Database.Port,
		name,
		tz,
		ssl,
	)
}

// IsProduction は本番環境かどうかを判定します
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production" || strings.ToLower(c.Environment) == "prod"
}

// IsDevelopment は開発環境かどうかを判定します
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development" || strings.ToLower(c.Environment) == "dev"
}

// IsTest はテスト環境かどうかを判定します
func (c *Config) IsTest() bool {
	return strings.ToLower(c.Environment) == "test"
}

// EnableCSRF はCSRF保護を有効にするかどうかを判定します
func (c *Config) EnableCSRF() bool {
	// 本番環境では必ずCSRF保護を有効にする
	if c.IsProduction() {
		return true
	}

	// 開発環境ではオプション（環境変数で制御）
	return c.Security.EnableCSRF
}

// GetAllowedOrigins は許可されたCORSオリジンのリストを取得します
func (c *Config) GetAllowedOrigins() []string {
	if c.CORS.AllowedOrigins == "" {
		// デフォルトは開発用のローカルホスト
		return []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		}
	}

	// カンマ区切りの文字列を配列に変換
	origins := strings.Split(c.CORS.AllowedOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}

// GetJWTAccessTokenDuration はアクセストークンの有効期限を取得します
func (c *Config) GetJWTAccessTokenDuration() string {
	if c.JWT.AccessTokenDuration == "" {
		return "1h" // デフォルト1時間
	}
	return c.JWT.AccessTokenDuration
}

// GetJWTRefreshTokenDuration はリフレッシュトークンの有効期限を取得します
func (c *Config) GetJWTRefreshTokenDuration() string {
	if c.JWT.RefreshTokenDuration == "" {
		return "168h" // デフォルト7日間
	}
	return c.JWT.RefreshTokenDuration
}

// GetLogLevel はログレベルを取得します
func (c *Config) GetLogLevel() string {
	if c.Log.Level == "" {
		if c.IsProduction() {
			return "info"
		}
		return "debug"
	}
	return c.Log.Level
}

// GetServerAddress はサーバーのアドレスを取得します
func (c *Config) GetServerAddress() string {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	return c.Server.Host + ":" + c.Server.Port
}

// Validate は設定の妥当性をチェックします
func (c *Config) Validate() error {
	// 必須設定のチェック
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if c.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}

	// 本番環境での追加チェック
	if c.IsProduction() {
		if c.JWT.SecretKey == "your-secret-key" {
			return fmt.Errorf("default JWT secret key cannot be used in production")
		}

		if !c.Database.SSL {
			return fmt.Errorf("SSL must be enabled for database connection in production")
		}
	}

	return nil
}

// getEnv は環境変数を取得し、デフォルト値を返します
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt は環境変数を整数として取得します
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 は環境変数を64bit整数として取得します
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool は環境変数をブール値として取得します
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
