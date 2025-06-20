package server

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hryt430/Yotei+/config"
	socialDomain "github.com/hryt430/Yotei+/internal/modules/social/domain"

	"github.com/hryt430/Yotei+/pkg/logger"
	"github.com/hryt430/Yotei+/pkg/token"

	// Common domain and validator (統一インターフェース)
	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	commonValidator "github.com/hryt430/Yotei+/internal/common/validator"

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
	notificationDatabaseInfra "github.com/hryt430/Yotei+/internal/modules/notification/infrastructure/database"
	notificationGateway "github.com/hryt430/Yotei+/internal/modules/notification/infrastructure/gateway"
	notificationMessaging "github.com/hryt430/Yotei+/internal/modules/notification/infrastructure/messaging"
	notificationDatabase "github.com/hryt430/Yotei+/internal/modules/notification/interface/database"
	"github.com/hryt430/Yotei+/internal/modules/notification/interface/websocket"
	notificationUseCase "github.com/hryt430/Yotei+/internal/modules/notification/usecase"
	notificationOutput "github.com/hryt430/Yotei+/internal/modules/notification/usecase/output"
	notificationPersistence "github.com/hryt430/Yotei+/internal/modules/notification/usecase/persistence"

	// Task module
	taskDatabaseInfra "github.com/hryt430/Yotei+/internal/modules/task/infrastructure/database"
	taskMessaging "github.com/hryt430/Yotei+/internal/modules/task/infrastructure/messaging"
	taskDatabase "github.com/hryt430/Yotei+/internal/modules/task/interface/database"
	taskUseCase "github.com/hryt430/Yotei+/internal/modules/task/usecase"

	// Social module
	socialDatabaseInfra "github.com/hryt430/Yotei+/internal/modules/social/infrastructure/database"
	socialDatabase "github.com/hryt430/Yotei+/internal/modules/social/interface/database"
	socialUseCase "github.com/hryt430/Yotei+/internal/modules/social/usecase"

	// Group module
	groupDatabaseInfra "github.com/hryt430/Yotei+/internal/modules/group/infrastructure/database"
	groupDatabase "github.com/hryt430/Yotei+/internal/modules/group/interface/database"
	groupUseCase "github.com/hryt430/Yotei+/internal/modules/group/usecase"
)

// NewDependencies は依存関係を初期化します（統一インターフェース対応版）
func NewDependencies(cfg *config.Config, log logger.Logger) (*Dependencies, error) {
	// Redis接続の初期化
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	// Redis接続テスト
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Warn("Redis connection failed, continuing without Redis", logger.Error(err))
		// Redisが利用できない場合はnilを設定（開発環境対応）
		redisClient = nil
	}

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

	// Redis Token Cache（Redis利用可能時のみ）
	var tokenRepository tokenService.ITokenRepository
	if redisClient != nil {
		redisTokenCache := authRedisInfra.NewRedisTokenCache(redisClient)
		tokenRepository = authRedis.NewTokenRepositoryAdapter(redisTokenCache, tokenStorage)
	} else {
		// Redis不使用時はDBのみ使用するアダプタを作成（logger追加）
		tokenRepository = NewDBOnlyTokenRepository(tokenStorage, log)
	}

	userSvc := userService.NewUserService(userRepository)
	tokenSvc := tokenService.NewTokenService(tokenRepository, jwtManager, accessTokenDuration, refreshTokenDuration)

	// AuthRepository の実装
	authRepository := &AuthRepositoryImpl{
		UserService:  *userSvc,
		TokenService: *tokenSvc,
	}
	authSvc := authService.NewAuthService(authRepository, *userSvc, *tokenSvc)

	// **統一されたUserValidator の実装**
	var userValidator commonDomain.UserValidator = commonValidator.NewUserValidator(userRepository)

	// Notification module dependencies
	notificationSqlHandler := notificationDatabaseInfra.NewSqlHandler()
	notificationRepo := &notificationDatabase.NotificationServiceRepository{
		SqlHandler: &notificationSqlHandler,
		Logger:     log,
	}

	// WebSocketハブの初期化
	wsHub := websocket.NewHub(log)

	// Notification gateways
	appGateway := notificationGateway.NewAppNotificationGateway(cfg, notificationRepo, wsHub, log)
	lineGateway := notificationGateway.NewLineGateway(cfg, log)

	// Type assertions to ensure interface compliance
	var notificationRepository notificationPersistence.NotificationRepository = notificationRepo
	var appNotificationGateway notificationOutput.AppNotificationGateway = appGateway
	var lineNotificationGateway notificationOutput.LineNotificationGateway = lineGateway

	// **通知ユースケース（統一されたUserValidatorを使用）**
	notificationUseCaseImpl := notificationUseCase.NewNotificationUseCase(
		notificationRepository,
		appNotificationGateway,
		lineNotificationGateway,
		userValidator, // 統一されたUserValidatorを使用
		log,
	)

	// Task module dependencies
	taskSqlHandler := taskDatabaseInfra.NewSqlHandler()
	taskRepository := taskDatabase.NewTaskRepository(&taskSqlHandler, log)

	// 統計リポジトリの初期化
	statsRepository := taskDatabase.NewTaskStatsRepository(&taskSqlHandler, log)

	// Event Publisher（修正版：戻り値統一）
	notificationAdapter := taskMessaging.NewNotificationAdapter(notificationUseCaseImpl)
	eventPublisher := taskMessaging.NewTaskEventPublisher(notificationAdapter, log)

	// **Task Service（統一されたUserValidatorを使用）**
	taskService := taskUseCase.NewTaskService(
		taskRepository,
		userValidator, // 統一されたUserValidatorを使用
		eventPublisher,
		log,
	)

	// Stats Service
	statsService := taskUseCase.NewTaskStatsService(
		taskRepository,
		statsRepository,
		log,
	)

	// Social module dependencies
	socialSqlHandler := socialDatabaseInfra.NewSqlHandler()
	friendshipRepository := socialDatabase.NewFriendshipRepository(socialSqlHandler.GetConnection(), log)
	invitationRepository := socialDatabase.NewInvitationRepository(socialSqlHandler.GetConnection(), log)

	// Social event publisher (simplified for now)
	socialEventPublisher := &SimpleSocialEventPublisher{logger: log}

	// URL gateway (simplified for now)
	urlGateway := &SimpleURLGateway{baseURL: "http://localhost:8080"}

	socialService := socialUseCase.NewSocialServiceImpl(
		friendshipRepository,
		invitationRepository,
		userValidator, // using the existing userValidator
		socialEventPublisher,
		urlGateway,
		log,
	)

	// Group module dependencies
	groupSqlHandler := groupDatabaseInfra.NewSqlHandler()
	groupRepository := groupDatabase.NewGroupRepository(groupSqlHandler.GetConnection(), log)
	groupService := groupUseCase.NewGroupService(groupRepository, userValidator, log)

	// メッセージブローカーとスケジューラー
	messageBroker := notificationMessaging.NewInMemoryMessageBroker(log)

	// **タスク期限通知スケジューラー（統一されたUserValidatorを使用）**
	taskScheduler := taskMessaging.NewTaskDueNotificationScheduler(
		*taskService,
		notificationAdapter,
		eventPublisher,
		log,
	)

	return &Dependencies{
		AuthService:         *authSvc,
		TokenService:        *tokenSvc,
		UserService:         *userSvc,
		NotificationUseCase: notificationUseCaseImpl,
		TaskService:         *taskService,
		StatsService:        statsService,
		SocialService:       socialService,
		GroupService:        groupService,
		WSHub:               wsHub,
		TaskScheduler:       taskScheduler,
		MessageBroker:       messageBroker,
		Logger:              log,
		Config:              cfg,
		// context管理用フィールドは初期化時は設定しない
	}, nil
}

// DBOnlyTokenRepository はRedis不使用時のトークンリポジトリ実装（修正版）
type DBOnlyTokenRepository struct {
	tokenStorage *authDatabase.TokenStorage
	logger       logger.Logger
}

// NewDBOnlyTokenRepository は新しいDBOnlyTokenRepositoryを作成
func NewDBOnlyTokenRepository(tokenStorage *authDatabase.TokenStorage, logger logger.Logger) *DBOnlyTokenRepository {
	return &DBOnlyTokenRepository{
		tokenStorage: tokenStorage,
		logger:       logger,
	}
}

func (r *DBOnlyTokenRepository) SaveTokenToBlacklist(token string, ttl time.Duration) error {
	// DBのみの場合はブラックリスト機能を簡易実装
	r.logger.Warn("Token blacklist feature disabled (Redis not available)")
	return nil
}

func (r *DBOnlyTokenRepository) IsTokenBlacklisted(token string) bool {
	// DBのみの場合は常にfalse（ブラックリスト機能無効）
	return false
}

func (r *DBOnlyTokenRepository) SaveRefreshToken(token *authDomain.RefreshToken) error {
	return r.tokenStorage.SaveRefreshToken(token)
}

func (r *DBOnlyTokenRepository) FindRefreshToken(token string) (*authDomain.RefreshToken, error) {
	return r.tokenStorage.FindRefreshTokenByToken(token)
}

func (r *DBOnlyTokenRepository) RevokeRefreshToken(token string) error {
	return r.tokenStorage.RevokeRefreshToken(token)
}

func (r *DBOnlyTokenRepository) DeleteExpiredRefreshTokens() error {
	return r.tokenStorage.DeleteExpiredRefreshTokens()
}

// AuthRepositoryImpl はAuthRepositoryの実装
type AuthRepositoryImpl struct {
	UserService  userService.UserService
	TokenService tokenService.TokenService
}

func (r *AuthRepositoryImpl) Register(ctx context.Context, email, username, password string) (*authDomain.User, error) {
	user := &authDomain.User{
		Email:    email,
		Username: username,
		Password: password,
	}

	return r.UserService.CreateUser(user)
}

func (r *AuthRepositoryImpl) Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error) {
	user, err := r.UserService.FindUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	if user != nil {
		accessToken, err := r.TokenService.GenerateAccessToken(user)
		if err != nil {
			return "", "", err
		}

		refreshToken, err := r.TokenService.GenerateRefreshToken(user)
		if err != nil {
			return "", "", err
		}

		return accessToken, refreshToken, nil
	}

	return "", "", nil
}

func (r *AuthRepositoryImpl) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	tokenEntity, err := r.TokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	user, err := r.UserService.FindUserByID(tokenEntity.UserID)
	if err != nil {
		return "", "", err
	}

	newAccessToken, err = r.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = r.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	err = r.TokenService.RevokeToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (r *AuthRepositoryImpl) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if err := r.TokenService.RevokeAccessToken(accessToken); err != nil {
		return err
	}

	return r.TokenService.RevokeToken(refreshToken)
}

// SimpleSocialEventPublisher は簡単なソーシャルイベントパブリッシャー実装
type SimpleSocialEventPublisher struct {
	logger logger.Logger
}

func (p *SimpleSocialEventPublisher) PublishFriendRequestSent(ctx context.Context, friendship *socialDomain.Friendship, message string) error {
	p.logger.Info("Friend request sent",
		logger.Any("friendshipID", friendship.ID),
		logger.Any("requesterID", friendship.RequesterID),
		logger.Any("addresseeID", friendship.AddresseeID),
		logger.Any("message", message))
	return nil
}

func (p *SimpleSocialEventPublisher) PublishFriendRequestAccepted(ctx context.Context, friendship *socialDomain.Friendship) error {
	p.logger.Info("Friend request accepted",
		logger.Any("friendshipID", friendship.ID),
		logger.Any("requesterID", friendship.RequesterID),
		logger.Any("addresseeID", friendship.AddresseeID))
	return nil
}

func (p *SimpleSocialEventPublisher) PublishFriendRequestDeclined(ctx context.Context, friendship *socialDomain.Friendship) error {
	p.logger.Info("Friend request declined",
		logger.Any("friendshipID", friendship.ID),
		logger.Any("requesterID", friendship.RequesterID),
		logger.Any("addresseeID", friendship.AddresseeID))
	return nil
}

func (p *SimpleSocialEventPublisher) PublishFriendRemoved(ctx context.Context, userID, friendID uuid.UUID) error {
	p.logger.Info("Friend removed", logger.Any("userID", userID), logger.Any("friendID", friendID))
	return nil
}

func (p *SimpleSocialEventPublisher) PublishUserBlocked(ctx context.Context, userID, targetID uuid.UUID) error {
	p.logger.Info("User blocked", logger.Any("userID", userID), logger.Any("targetID", targetID))
	return nil
}

func (p *SimpleSocialEventPublisher) PublishInvitationCreated(ctx context.Context, invitation *socialDomain.Invitation) error {
	p.logger.Info("Invitation created",
		logger.Any("invitationID", invitation.ID),
		logger.Any("inviterID", invitation.InviterID),
		logger.Any("type", invitation.Type))
	return nil
}

func (p *SimpleSocialEventPublisher) PublishInvitationAccepted(ctx context.Context, invitation *socialDomain.Invitation) error {
	p.logger.Info("Invitation accepted",
		logger.Any("invitationID", invitation.ID),
		logger.Any("inviterID", invitation.InviterID),
		logger.Any("inviteeID", invitation.InviteeID))
	return nil
}

func (p *SimpleSocialEventPublisher) PublishInvitationDeclined(ctx context.Context, invitation *socialDomain.Invitation) error {
	p.logger.Info("Invitation declined",
		logger.Any("invitationID", invitation.ID),
		logger.Any("inviterID", invitation.InviterID),
		logger.Any("inviteeID", invitation.InviteeID))
	return nil
}

// SimpleURLGateway は簡単なURL生成ゲートウェイ実装
type SimpleURLGateway struct {
	baseURL string
}

func (g *SimpleURLGateway) GenerateInviteURL(ctx context.Context, invitationID uuid.UUID, code string) (string, error) {
	return fmt.Sprintf("%s/invite/%s", g.baseURL, code), nil
}
