package validator

import (
	"context"

	commonDomain "github.com/hryt430/Yotei+/internal/common/domain"
	authDB "github.com/hryt430/Yotei+/internal/modules/auth/interface/database"
)

// UserValidator は統一されたユーザー存在確認の実装
type UserValidator struct {
	userRepo *authDB.IUserRepository
}

// NewUserValidator は新しいUserValidatorを作成
func NewUserValidator(userRepo *authDB.IUserRepository) commonDomain.UserValidator {
	return &UserValidator{
		userRepo: userRepo,
	}
}

// UserExists はユーザーが存在するかチェック
func (v *UserValidator) UserExists(ctx context.Context, userID string) (bool, error) {
	return v.userRepo.UserExists(userID)
}

// GetUserInfo はユーザー情報を取得
func (v *UserValidator) GetUserInfo(ctx context.Context, userID string) (*commonDomain.UserInfo, error) {
	basicInfo, err := v.userRepo.GetUserBasicInfo(userID)
	if err != nil {
		return nil, err
	}
	if basicInfo == nil {
		return nil, nil
	}

	return &commonDomain.UserInfo{
		ID:       basicInfo.ID,
		Username: basicInfo.Username,
		Email:    basicInfo.Email,
	}, nil
}

// GetUsersInfoBatch は複数ユーザーの基本情報を一括取得
func (v *UserValidator) GetUsersInfoBatch(ctx context.Context, userIDs []string) (map[string]*commonDomain.UserInfo, error) {
	batchInfo, err := v.userRepo.GetUsersBasicInfoBatch(userIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*commonDomain.UserInfo)
	for userID, info := range batchInfo {
		result[userID] = &commonDomain.UserInfo{
			ID:       info.ID,
			Username: info.Username,
			Email:    info.Email,
		}
	}

	return result, nil
}
