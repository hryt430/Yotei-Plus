package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/server"
	appLogger "github.com/hryt430/Yotei+/pkg/logger"
)

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 設定の妥当性チェック
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// ロガーの初期化
	logger := server.NewLogger(cfg)

	// 依存関係の初期化
	deps, err := server.NewDependencies(cfg, *logger)
	if err != nil {
		logger.Fatal("Failed to initialize dependencies", appLogger.Error(err))
	}

	// ルーターの設定
	router := server.SetupRouter(deps)

	// HTTPサーバーの設定
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// サーバーをgoroutineで起動
	go func() {
		logger.Info("Starting server",
			appLogger.Any("address", srv.Addr),
			appLogger.Any("environment", cfg.Environment),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", appLogger.Error(err))
		}
	}()

	// Graceful shutdown の設定
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 30秒のタイムアウトでサーバーを停止
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", appLogger.Error(err))
	}

	logger.Info("Server exited")
}
