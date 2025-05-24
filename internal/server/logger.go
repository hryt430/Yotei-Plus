package server

import (
	"github.com/hryt430/Yotei+/config"
	appLogger "github.com/hryt430/Yotei+/pkg/logger"
)

// CreateLoggerConfig はconfig.LogからLogger.Configに変換する
func CreateLoggerConfig(cfg *config.Config) *appLogger.Config {
	loggerConfig := &appLogger.Config{
		Level:       cfg.Log.Level,
		Output:      cfg.Log.Output,
		Development: cfg.IsDevelopment(),
	}

	// ファイル出力設定
	loggerConfig.File.Path = "logs/app.log"
	loggerConfig.File.MaxSize = 100
	loggerConfig.File.MaxBackups = 3
	loggerConfig.File.MaxAge = 28
	loggerConfig.File.Compress = true

	// 環境に応じた調整
	if cfg.IsProduction() {
		loggerConfig.Output = "file" // 本番では基本的にファイル出力
		loggerConfig.Development = false
	}

	return loggerConfig
}

// NewLogger はロガーを作成する便利関数
func NewLogger(cfg *config.Config) *appLogger.Logger {
	loggerConfig := CreateLoggerConfig(cfg)
	return appLogger.NewLogger(loggerConfig)
}
