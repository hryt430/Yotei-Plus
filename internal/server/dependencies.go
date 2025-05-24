package server

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hryt430/Yotei+/config"

	// commonDB "github.com/hryt430/Yotei+/internal/common/infrastructure/database"
	"github.com/hryt430/Yotei+/pkg/logger"
	"github.com/hryt430/Yotei+/pkg/token"

	// Auth module
	authDomain "github.com/hryt430/Yotei+/internal/modules/auth/domain"
	authDatabaseInfra "github.com/hryt430/Yotei+/internal/modules/auth/infrastructure/database"
	authRedisInfra "github.com/hryt430/Yotei+/internal/modules/auth/infrastructure/redis"
	authDatabase "github.com/hryt430/Yotei+/internal/modules/auth/interface/database"
	authRedis "github.com/hryt430/Yotei+/internal/modules/auth/interface/redis"
	authService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/auth"
	tokenService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/token"
	userService "github.com/hryt430/Yotei+/internal/modules/auth/usecase/user"

	// Notification module
	notificationDomain "github.com/hryt430/Yotei+/internal/modules/notification/domain"
	notificationDatabaseInfra "github.com/hryt430/Yotei+/internal/modules/notification/infrastructure/database"
	notificationGateway "github.com/hryt430/Yotei+/internal/modules/notification/infrastructure/gateway"
	notificationDatabase "github.com/hryt430/Yotei+/internal/modules/notification/interface/database"
	"github.com/hryt430/Yotei+/internal/modules/notification/interface/websocket"
	notificationUseCase "github.com/hryt430/Yotei+/internal/modules/notification/usecase"
	notificationOutput "github.com/hryt430/Yotei+/internal/modules/notification/usecase/output"
	notificationPersistence "github.com/hryt430/Yotei+/internal/modules/notification/usecase/persistence"

	// Task module
	taskDatabaseInfra "github.com/hryt430/Yotei+/internal/modules/task/infrastructure/database"
	taskDatabase "github.com/hryt430/Yotei+/internal/modules/task/interface/database"
	taskUseCase "github.com/hryt430/Yotei+/internal/modules/task/usecase"
)

// NewDependencies は依存関係を初期化します
func NewDependencies(cfg *config.Config, log logger.Logger) (*Dependencies, error) {
	// データベース接続の初期化
	// db, err := commonDB.NewMySQLConnection(cfg)
	// if err != nil {
	// 	return nil, err
	// }

	// Redis接続の初期化
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	// JWTマネージャーの初期化
	accessTokenDuration, err := time.ParseDuration(cfg.GetJWTAccessTokenDuration())
	if err != nil {
		return nil, err
	}

	refreshTokenDuration, err := time.ParseDuration(cfg.GetJWTRefreshTokenDuration())
	if err != nil {
		return nil, err
	}

	jwtManager := token.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.Issuer)

	// Auth module dependencies
	authSqlHandler := authDatabaseInfra.NewSqlHandler()
	userRepository := &authDatabase.IUserRepository{
		SqlHandler: &authSqlHandler,
	}
	tokenStorage := &authDatabase.TokenStorage{
		SqlHandler: &authSqlHandler,
	}
	redisTokenCache := authRedisInfra.NewRedisTokenCache(redisClient)
	tokenRepository := authRedis.NewTokenRepositoryAdapter(redisTokenCache, tokenStorage)

	userSvc := userService.NewUserService(userRepository)
	tokenSvc := tokenService.NewTokenService(tokenRepository, jwtManager, accessTokenDuration, refreshTokenDuration)

	// AuthRepository の実装
	authRepository := &AuthRepositoryImpl{
		UserService:  *userSvc,
		TokenService: *tokenSvc,
	}
	authSvc := authService.NewAuthService(authRepository, *userSvc, *tokenSvc)

	// Notification module dependencies
	// notificationSqlHandler := notificationDatabaseInfra.NewSqlHandler()
	// notificationRepo := &notificationDatabase.NotificationServiceRepository{
	// 	SqlHandler: &notificationSqlHandler,
	// 	Logger:     log,
	// }

	// // Notification gateways
	// appGateway := notificationGateway.NewWebhookGateway(cfg, log)
	// lineGateway := notificationGateway.NewLineGateway(cfg, log)
	// webhookGateway := notificationGateway.NewWebhookGateway(cfg, log)

	// // Type assertions to ensure interface compliance
	// var notificationRepository notificationPersistence.NotificationRepository = notificationRepo
	// var appNotificationGateway notificationOutput.AppNotificationGateway = &appNotificationGatewayAdapter{webhookGateway, notificationRepo}
	// var lineNotificationGateway notificationOutput.LineNotificationGateway = lineGateway
	// var webhookOutput notificationOutput.WebhookOutput = webhookGateway

	// notificationUseCaseImpl := notificationUseCase.NewNotificationUseCase(
	// 	notificationRepository,
	// 	appNotificationGateway,
	// 	lineNotificationGateway,
	// 	webhookOutput,
	// )

	// Notification module dependencies
	notificationSqlHandler := notificationDatabaseInfra.NewSqlHandler()
	notificationRepo := &notificationDatabase.NotificationServiceRepository{
		SqlHandler: &notificationSqlHandler,
		Logger:     log,
	}

	// WebSocketハブの初期化（必要に応じて）
	wsHub := websocket.NewHub()
	go wsHub.Run() // 別のgoroutineで実行

	// Notification gateways
	// アプリ内通知用：WebSocket + データベース
	appGateway := notificationGateway.NewAppNotificationGateway(cfg, notificationRepo, wsHub, log)

	// LINE通知用
	lineGateway := notificationGateway.NewLineGateway(cfg, log)

	// Webhook送信用：外部システム連携
	// webhookGateway := notificationGateway.NewWebhookGateway(cfg, log)

	// Type assertions to ensure interface compliance
	var notificationRepository notificationPersistence.NotificationRepository = notificationRepo
	var appNotificationGateway notificationOutput.AppNotificationGateway = appGateway
	var lineNotificationGateway notificationOutput.LineNotificationGateway = lineGateway
	// var webhookOutput notificationOutput.WebhookOutput = webhookGateway

	notificationUseCaseImpl := notificationUseCase.NewNotificationUseCase(
		notificationRepository,
		appNotificationGateway,
		lineNotificationGateway,
		// webhookOutput,
	)

	// Task module dependencies
	taskSqlHandler := taskDatabaseInfra.NewSqlHandler()
	taskRepository := &taskDatabase.TaskRepository{
		SqlHandler: &taskSqlHandler,
	}

	taskService := taskUseCase.NewTaskService(taskRepository)

	return &Dependencies{
		AuthService:         *authSvc,
		TokenService:        *tokenSvc,
		UserService:         *userSvc,
		NotificationUseCase: notificationUseCaseImpl,
		TaskService:         *taskService,
		Logger:              log,
		Config:              cfg,
	}, nil
}

// AuthRepositoryImpl はAuthRepositoryの簡易実装
type AuthRepositoryImpl struct {
	UserService  userService.UserService
	TokenService tokenService.TokenService
}

func (r *AuthRepositoryImpl) Register(ctx context.Context, email, username, password string) (*authDomain.User, error) {
	// ユーザー作成処理
	user := &authDomain.User{
		Email:    email,
		Username: username,
		Password: password,
	}

	return r.UserService.CreateUser(user)
}

func (r *AuthRepositoryImpl) Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error) {
	// ユーザー認証処理
	user, err := r.UserService.FindUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	// パスワード検証の実装が必要
	// 簡易的な実装
	if user != nil {
		// アクセストークン生成
		accessToken, err := r.TokenService.GenerateAccessToken(user)
		if err != nil {
			return "", "", err
		}

		// リフレッシュトークン生成
		refreshToken, err := r.TokenService.GenerateRefreshToken(user)
		if err != nil {
			return "", "", err
		}

		return accessToken, refreshToken, nil
	}

	return "", "", nil
}

func (r *AuthRepositoryImpl) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	// リフレッシュトークンの検証と新しいトークンの生成
	tokenEntity, err := r.TokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	user, err := r.UserService.FindUserByID(tokenEntity.UserID)
	if err != nil {
		return "", "", err
	}

	// 新しいアクセストークン生成
	newAccessToken, err = r.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// 新しいリフレッシュトークン生成
	newRefreshToken, err = r.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	// 古いリフレッシュトークンを無効化
	err = r.TokenService.RevokeToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (r *AuthRepositoryImpl) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// アクセストークンをブラックリストに追加
	if err := r.TokenService.RevokeAccessToken(accessToken); err != nil {
		return err
	}

	// リフレッシュトークンを無効化
	return r.TokenService.RevokeToken(refreshToken)
}

// appNotificationGatewayAdapter はWebhookGatewayをAppNotificationGatewayインターフェースに適合させるアダプター
type appNotificationGatewayAdapter struct {
	webhookGateway notificationOutput.WebhookOutput
	repository     notificationPersistence.NotificationRepository
}

func (a *appNotificationGatewayAdapter) SendNotification(ctx context.Context, userID, title, message string, metadata map[string]string) error {
	// Webhookとして通知を送信
	return a.webhookGateway.SendWebhook(ctx, notificationOutput.EventNotificationSent, map[string]interface{}{
		"user_id":  userID,
		"title":    title,
		"message":  message,
		"metadata": metadata,
	})
}

func (a *appNotificationGatewayAdapter) MarkAsRead(ctx context.Context, notificationID string) error {
	// 通知を既読としてマーク
	return a.repository.UpdateStatus(ctx, notificationID, notificationDomain.StatusRead)
}

func (a *appNotificationGatewayAdapter) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	// 未読通知数を取得
	return a.repository.CountByUserIDAndStatus(ctx, userID, notificationDomain.StatusPending)
}
