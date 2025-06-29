package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/server"
	appLogger "github.com/hryt430/Yotei+/pkg/logger"

	// Swagger関連のimport
	_ "github.com/hryt430/Yotei+/docs" // swag initで自動生成されるドキュメント
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Yotei+ Task Management API
// @version         1.0
// @description     高機能タスク管理システムのREST API
// @description
// @description     ## 概要
// @description     Yotei+は、個人およびチーム向けの包括的なタスク管理システムです。
// @description
// @description     ## 主要機能
// @description     - 🔐 **認証・認可**: JWT ベースの安全な認証システム
// @description     - 📋 **タスク管理**: 作成、更新、削除、割り当て機能
// @description     - 📊 **統計・分析**: 詳細な進捗統計とダッシュボード
// @description     - 🔔 **通知システム**: リアルタイム通知機能
// @description     - 👥 **ユーザー管理**: プロフィール管理と権限制御
// @description
// @description     ## 認証方法
// @description     このAPIはJWTトークンベースの認証を使用します。
// @description     1. `/api/v1/auth/login` でログインしてアクセストークンを取得
// @description     2. 保護されたエンドポイントには `Authorization: Bearer <token>` ヘッダーを付与

// @termsOfService  https://yotei-plus.example.com/terms
// @contact.name    Yotei+ API Support
// @contact.url     https://yotei-plus.example.com/support
// @contact.email   support@yotei-plus.example.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT認証トークン。値の形式: "Bearer {token}"

// @tag.name auth
// @tag.description 認証・認可関連のAPI

// @tag.name users
// @tag.description ユーザー管理関連のAPI

// @tag.name tasks
// @tag.description タスク管理関連のAPI

// @tag.name stats
// @tag.description 統計・分析関連のAPI

// @tag.name notifications
// @tag.description 通知管理関連のAPI

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

	// バックグラウンドサービスの開始
	// server.StartBackgroundServices(deps)
	// defer server.StopBackgroundServices(deps)

	// ルーターの設定
	router := server.SetupRouter(deps)

	// Swagger UIの追加（開発環境でのみ有効）
	if !cfg.IsProduction() {
		// Swagger UIエンドポイント
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// API仕様書への直接アクセス
		router.GET("/api-docs", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		})

		logger.Info("Swagger UI enabled",
			appLogger.Any("url", "http://"+cfg.GetServerAddress()+"/swagger/index.html"),
		)
	}

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
