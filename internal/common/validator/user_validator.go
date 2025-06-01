package validator

import (
	"context"

	authDB "github.com/hryt430/Yotei+/internal/modules/auth/interface/database"
	notificationUsecase "github.com/hryt430/Yotei+/internal/modules/notification/usecase"
)

// UserValidator はユーザー存在確認の実装
type UserValidator struct {
	userRepo *authDB.IUserRepository
}

// NewUserValidator は新しいUserValidatorを作成
func NewUserValidator(userRepo *authDB.IUserRepository) *UserValidator {
	return &UserValidator{
		userRepo: userRepo,
	}
}

// UserExists はユーザーが存在するかチェック
func (v *UserValidator) UserExists(ctx context.Context, userID string) (bool, error) {
	return v.userRepo.UserExists(userID)
}

// GetUserInfo はユーザー情報を取得
func (v *UserValidator) GetUserInfo(ctx context.Context, userID string) (*notificationUsecase.UserInfo, error) {
	basicInfo, err := v.userRepo.GetUserBasicInfo(userID)
	if err != nil {
		return nil, err
	}
	if basicInfo == nil {
		return nil, nil
	}

	return &notificationUsecase.UserInfo{
		ID:       basicInfo.ID,
		Username: basicInfo.Username,
		Email:    basicInfo.Email,
	}, nil
}
