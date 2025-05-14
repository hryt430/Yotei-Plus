package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// サーバー設定
	ServerPort         string
	ServerTimeout      time.Duration
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration

	// データベース設定
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis設定
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// JWT設定
	JWTSecretKey          string
	JWTIssuer             string
	JWTExpirationTime     time.Duration
	RefreshExpirationTime time.Duration

	// その他設定
	Environment string
	Debug       bool
}

func LoadConfig(path string) (*Config, error) {
	// .envファイルの読み込み
	err := godotenv.Load(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return nil, fmt.Errorf(".env file not found: %w", err)
	}

	// 値の読み込み
	config := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		ServerTimeout:      getEnvAsDuration("SERVER_TIMEOUT", 30*time.Second),
		ServerReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
		ServerWriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "app_db"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		JWTSecretKey:          getEnv("JWT_SECRET_KEY", "secret"),
		JWTIssuer:             getEnv("JWT_ISSUER", "app"),
		JWTExpirationTime:     getEnvAsDuration("JWT_EXPIRATION_TIME", 1*time.Hour),
		RefreshExpirationTime: getEnvAsDuration("REFRESH_EXPIRATION_TIME", 7*24*time.Hour),

		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getEnvAsBool("DEBUG", false),
	}

	return config, nil
}

// ヘルパー関数
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := os.Getenv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsDuration(name string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(name)
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	return defaultVal
}
