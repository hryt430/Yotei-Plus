package server

import (
	"github.com/gin-gonic/gin"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/common/middleware"
	"github.com/hryt430/Yotei+/pkg/logger"

	authMiddleware "github.com/hryt430/Yotei+/internal/modules/auth/infrastructure/middleware"
	authController "github.com/hryt430/Yotei+/internal/modules/auth/interface/controller"
	authService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/auth"
	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	userService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/user"

	notificationController "github.com/hryt430/Yotei+/internal/modules/notification/interface/controller"
	notificationUseCase "github.com/hryt430/Yotei+/internal/modules/notification/usecase/input"

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
	Logger              logger.Logger
	Config              *config.Config
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
		})
	})

	// APIグループ
	api := router.Group("/api/v1")

	// 各モジュールのルート設定
	setupAuthRoutes(api, deps)
	setupNotificationRoutes(api, deps)
	setupTaskRoutes(api, deps)

	return router
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
			// 管理者機能のエンドポイントをここに追加
		}
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
	}
}
