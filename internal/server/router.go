package server

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/pkg/logger"

	authMiddleware "github.com/hryt430/Yotei+/internal/modules/auth/infrastructure/middleware"
	authController "github.com/hryt430/Yotei+/internal/modules/auth/interface/controller"
	userController "github.com/hryt430/Yotei+/internal/modules/auth/interface/controller"
	authService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/auth"
	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	userService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/user"

	notificationMessaging "github.com/hryt430/Yotei+/internal/modules/notification/infrastructure/messaging"
	notificationController "github.com/hryt430/Yotei+/internal/modules/notification/interface/controller"
	"github.com/hryt430/Yotei+/internal/modules/notification/interface/websocket"
	notificationUseCase "github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"

	taskMessaging "github.com/hryt430/Yotei+/internal/modules/task/infrastructure/messaging"
	taskController "github.com/hryt430/Yotei+/internal/modules/task/interface/controller"
	taskUseCase "github.com/hryt430/Yotei+/internal/modules/task/usecase"
)

// Dependencies は各モジュールの依存関係を格納する構造体
type Dependencies struct {
	AuthService         authService.AuthService
	TokenService        tokenService.TokenService
	UserService         userService.UserService
	NotificationUseCase notificationUseCase.NotificationUseCase
	TaskService         taskUseCase.TaskService
	StatsService        *taskUseCase.TaskStatsService
	WSHub               *websocket.Hub
	TaskScheduler       *taskMessaging.TaskDueNotificationScheduler
	MessageBroker       notificationMessaging.MessageBroker
	Logger              logger.Logger
	Config              *config.Config

	// バックグラウンドサービス管理用
	cancelFunc   context.CancelFunc
	backgroundWg sync.WaitGroup
}

// SetupRouter はAPIルーターをセットアップする
func SetupRouter(deps *Dependencies) *gin.Engine {
	// リリースモードの設定
	if deps.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// ルーターの作成
	router := gin.New()

	// 共通ミドルウェアの適用
	router.Use(middleware.RecoveryMiddleware(deps.Logger))
	router.Use(middleware.LoggerMiddleware(deps.Logger))
	router.Use(middleware.CORSMiddleware(deps.Config))

	// セキュリティヘッダー
	router.Use(middleware.SecurityHeadersMiddleware())

	// Next.jsとのCSRF連携
	if deps.Config.EnableCSRF() {
		router.Use(middleware.SetCSRFToken())
		router.Use(middleware.CSRFProtection())
	}

	// ヘルスチェックエンドポイント
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "task-management-api",
			"version": "v1.0.0",
		})
	})

	// APIグループ
	api := router.Group("/api/v1")

	// WebSocketエンドポイント（認証必要）
	setupWebSocketRoutes(router, deps)

	// 各モジュールのルート設定
	setupAuthRoutes(api, deps)
	setupUserRoutes(api, deps)
	setupNotificationRoutes(api, deps)
	setupTaskRoutes(api, deps)

	return router
}

// setupWebSocketRoutes はWebSocketエンドポイントをセットアップする（context対応版）
func setupWebSocketRoutes(router *gin.Engine, deps *Dependencies) {
	if deps.WSHub == nil {
		deps.Logger.Warn("WebSocket hub not available, skipping WebSocket routes")
		return
	}

	// 認証ミドルウェアの初期化（notificationRoutesと同じパターン）
	authMw := authMiddleware.NewAuthMiddleware(deps.TokenService)

	// WebSocketエンドポイント
	wsGroup := router.Group("/ws")
	{
		// ✅ WebSocket用の認証ミドルウェアを追加
		wsGroup.GET("/notifications",
			authMw.WebSocketAuthRequired(), // ← 新しく追加する認証ミドルウェア
			websocket.ServeWs(deps.WSHub, deps.Logger))
	}
}

// setupAuthRoutes は認証モジュールのルートをセットアップする
func setupAuthRoutes(router *gin.RouterGroup, deps *Dependencies) {
	// 認証コントローラの初期化
	authCtrl := authController.NewAuthController(deps.AuthService, deps.Logger)

	// 認証ミドルウェアの初期化
	authMw := authMiddleware.NewAuthMiddleware(deps.TokenService)

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
			// 将来の管理者機能用
		}
	}
}

// setupUserRoutes はユーザー管理のルートをセットアップする
func setupUserRoutes(router *gin.RouterGroup, deps *Dependencies) {
	// ユーザーコントローラの初期化
	userCtrl := userController.NewUserController(deps.UserService, deps.Logger)

	// 認証ミドルウェアの初期化
	authMw := authMiddleware.NewAuthMiddleware(deps.TokenService)

	// ユーザールートグループ（認証が必要）
	userRoutes := router.Group("/users")
	userRoutes.Use(authMw.AuthRequired())
	{
		// ユーザー一覧取得（タスク割り当て用）
		userRoutes.GET("", userCtrl.GetUsers)

		// 現在のユーザー関連（互換性維持）
		userRoutes.GET("/me", userCtrl.GetCurrentUser)
		userRoutes.PUT("/me", userCtrl.UpdateCurrentUser)

		// 特定ユーザー関連
		userRoutes.GET("/:id", userCtrl.GetUser)
		userRoutes.PUT("/:id", userCtrl.UpdateUser)
	}
}

// setupNotificationRoutes は通知モジュールのルートをセットアップする
func setupNotificationRoutes(router *gin.RouterGroup, deps *Dependencies) {
	// 通知コントローラの初期化
	notificationCtrl := notificationController.NewNotificationController(deps.NotificationUseCase, deps.Logger)

	// 認証ミドルウェアの初期化
	authMw := authMiddleware.NewAuthMiddleware(deps.TokenService)

	// 通知ルートグループ（認証が必要）
	notificationRoutes := router.Group("/notifications")
	notificationRoutes.Use(authMw.AuthRequired())

	// 通知ルートの登録
	notificationController.RegisterNotificationRoutes(notificationRoutes, notificationCtrl)
}

// setupTaskRoutes はタスクモジュールのルートをセットアップする
func setupTaskRoutes(router *gin.RouterGroup, deps *Dependencies) {
	// タスクコントローラの初期化
	taskCtrl := taskController.NewTaskController(deps.TaskService)

	// 統計コントローラの初期化
	statsCtrl := taskController.NewTaskStatsController(deps.StatsService)

	// 認証ミドルウェアの初期化
	authMw := authMiddleware.NewAuthMiddleware(deps.TokenService)

	// タスクルートグループ（認証が必要）
	taskRoutes := router.Group("/tasks")
	taskRoutes.Use(authMw.AuthRequired())
	{
		// タスクCRUD操作
		taskRoutes.POST("", taskCtrl.CreateTask)
		taskRoutes.GET("/:id", taskCtrl.GetTask)
		taskRoutes.PUT("/:id", taskCtrl.UpdateTask)
		taskRoutes.DELETE("/:id", taskCtrl.DeleteTask)

		// タスク一覧・検索
		taskRoutes.GET("", taskCtrl.ListTasks)
		taskRoutes.GET("/search", taskCtrl.SearchTasks)

		// タスクの状態管理
		taskRoutes.PUT("/:id/assign", taskCtrl.AssignTask)
		taskRoutes.PUT("/:id/status", taskCtrl.ChangeTaskStatus)

		// 特定条件でのタスク取得
		taskRoutes.GET("/overdue", taskCtrl.GetOverdueTasks)
		taskRoutes.GET("/my", taskCtrl.GetMyTasks)
		taskRoutes.GET("/user/:user_id", taskCtrl.GetUserTasks)

		// === 統計情報API ===
		statsGroup := taskRoutes.Group("/stats")
		{
			// ダッシュボード統計
			statsGroup.GET("/dashboard", statsCtrl.GetDashboardStats)

			// 日次統計
			statsGroup.GET("/today", statsCtrl.GetTodayStats)
			statsGroup.GET("/daily/:date", statsCtrl.GetDailyStats)

			// 週次・月次統計
			statsGroup.GET("/weekly", statsCtrl.GetWeeklyStats)
			statsGroup.GET("/monthly", statsCtrl.GetMonthlyStats)

			// 進捗情報
			statsGroup.GET("/progress-summary", statsCtrl.GetProgressSummary)
			statsGroup.GET("/progress-level", statsCtrl.GetProgressLevel)

			// 分析情報
			statsGroup.GET("/category-breakdown", statsCtrl.GetCategoryBreakdown)
			statsGroup.GET("/priority-breakdown", statsCtrl.GetPriorityBreakdown)
		}
	}
}

// StartBackgroundServices はバックグラウンドサービスを開始する（context対応版）
func StartBackgroundServices(deps *Dependencies) {
	// キャンセル可能なcontextを作成
	ctx, cancel := context.WithCancel(context.Background())
	deps.cancelFunc = cancel

	// WebSocketハブの起動（context対応）
	if deps.WSHub != nil {
		deps.backgroundWg.Add(1)
		go func() {
			defer deps.backgroundWg.Done()

			if err := deps.WSHub.Run(ctx); err != nil && err != context.Canceled {
				deps.Logger.Error("WebSocket hub stopped with error", logger.Error(err))
			} else {
				deps.Logger.Info("WebSocket hub stopped gracefully")
			}
		}()
		deps.Logger.Info("WebSocket hub started")
	}

	// タスクスケジューラーの起動
	if deps.TaskScheduler != nil {
		deps.TaskScheduler.Start(ctx)
		deps.Logger.Info("Task due notification scheduler started")
	}
}

// StopBackgroundServices はバックグラウンドサービスを停止する（context対応版）
func StopBackgroundServices(deps *Dependencies) {
	deps.Logger.Info("Stopping background services...")

	// contextをキャンセルしてWebSocketハブを停止
	if deps.cancelFunc != nil {
		deps.cancelFunc()
		deps.Logger.Info("Background context cancelled")
	}

	// タスクスケジューラーの停止
	if deps.TaskScheduler != nil {
		deps.TaskScheduler.Stop()
		deps.Logger.Info("Task due notification scheduler stopped")
	}

	// メッセージブローカーの停止
	if deps.MessageBroker != nil {
		deps.MessageBroker.Close()
		deps.Logger.Info("Message broker stopped")
	}

	// 全てのバックグラウンドサービスの完了を待機
	deps.backgroundWg.Wait()
	deps.Logger.Info("All background services stopped")
}
