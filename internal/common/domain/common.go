package domain

import "context"

// UserInfo は統一されたユーザー基本情報構造体
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Pagination はページネーション情報
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// UserValidator は統一されたユーザーバリデーションインターフェース
type UserValidator interface {
	// ユーザーの存在確認
	UserExists(ctx context.Context, userID string) (bool, error)

	// 単一ユーザー情報取得
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)

	// 複数ユーザー情報の一括取得（N+1問題解決用）
	GetUsersInfoBatch(ctx context.Context, userIDs []string) (map[string]*UserInfo, error)
}
