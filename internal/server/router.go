package api

import (
	"github.com/gin-gonic/gin"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/pkg/logger"

	authMiddleware "github.com/hryt430/Yotei+/internal/modules/auth/infrastructure/middleware"
	authController "github.com/hryt430/Yotei+/internal/modules/auth/interface/controller"
	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	notificationController "github.com/hryt430/Yotei+/internal/modules/notification/interface/controller"
	// その他必要なコントローラやミドルウェアをインポート
)

// SetupRouter はAPIルーターをセットアップする
func SetupRouter(cfg *config.Config, log logger.Logger) *gin.Engine {
	// リリースモードの設定
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// ルーターの作成
	router := gin.New()

	// 共通ミドルウェアの適用
	router.Use(middleware.RecoveryMiddleware(log))
	router.Use(middleware.LoggerMiddleware(log))
	router.Use(middleware.CORSMiddleware(cfg))

	// Next.jsとのCSRF連携
	if cfg.EnableCSRF() {
		router.Use(middleware.SetCSRFToken())
		router.Use(middleware.CSRFProtection())
	}

	// APIグループ
	api := router.Group("/api")

	// 各モジュールのルート設定
	setupAuthRoutes(api, cfg, log)
	setupNotificationRoutes(api, cfg, log)
	// 他のモジュールのルート設定

	return router
}

// setupAuthRoutes は認証モジュールのルートをセットアップする
func setupAuthRoutes(router *gin.RouterGroup, cfg *config.Config, log logger.Logger) {
	// 認証コントローラの初期化（依存関係の注入はここで簡略化）
	authCtrl := getAuthController(cfg, log)
	tokenUseCase := getTokenUseCase(cfg, log)

	// 認証ミドルウェアの初期化
	authMw := authMiddleware.NewAuthMiddleware(tokenUseCase)

	// 認証ルートグループ
	authRoutes := router.Group("/auth")
	{
		// パブリックエンドポイント
		authRoutes.POST("/register", authCtrl.Register)
		authRoutes.POST("/login", authCtrl.Login)
		authRoutes.POST("/refresh-token", authCtrl.RefreshToken)

		// 認証が必要なエンドポイント
		authenticated := authRoutes.Group("")
		authenticated.Use(authMw.AuthRequired())
		{
			authenticated.POST("/logout", authCtrl.Logout)
			authenticated.GET("/me", authCtrl.Me)
		}

		// 管理者専用エンドポイント
		admin := authRoutes.Group("/admin")
		admin.Use(authMw.AuthRequired(), authMw.RoleRequired("admin"))
		{
			// 管理者機能
		}
	}
}

// setupNotificationRoutes は通知モジュールのルートをセットアップする
func setupNotificationRoutes(router *gin.RouterGroup, cfg *config.Config, log logger.Logger) {
	// 通知コントローラの初期化
	notificationCtrl := getNotificationController(cfg, log)
	tokenUseCase := getTokenUseCase(cfg, log)

	// 認証ミドルウェアの初期化
	authMw := authMiddleware.NewAuthMiddleware(tokenUseCase)

	// 通知ルートグループ（認証が必要）
	notificationRoutes := router.Group("/notifications")
	notificationRoutes.Use(authMw.AuthRequired())

	// 通知ルートの登録
	notificationController.RegisterNotificationRoutes(notificationRoutes, notificationCtrl)
}

// 以下はモック関数（実際の実装では依存性注入コンテナや専用のファクトリ関数を使用）
func getAuthController(cfg *config.Config, log logger.Logger) *authController.AuthController {
	// 実際の初期化ロジック
	return &authController.AuthController{}
}

func getTokenUseCase(cfg *config.Config, log logger.Logger) tokenService.TokenUseCase {
	// 実際の初期化ロジック
	return nil
}

func getNotificationController(cfg *config.Config, log logger.Logger) *notificationController.NotificationController {
	// 実際の初期化ロジック
	return &notificationController.NotificationController{}
}
