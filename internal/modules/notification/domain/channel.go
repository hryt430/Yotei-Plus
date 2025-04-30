package domain

// ChannelType は通知チャネルの種類を表す
type ChannelType string

const (
	AppInternal ChannelType = "APP_INTERNAL" // アプリ内通知
	LineMessage ChannelType = "LINE"         // LINE通知
)

// Channel は通知チャネルを表すインターフェース
type Channel interface {
	GetType() ChannelType
}

// AppChannel はアプリ内通知チャネル
type AppChannel struct {
	UserID uint
}

// GetType はチャネルタイプを返す
func (c *AppChannel) GetType() ChannelType {
	return AppInternal
}

// NewAppChannel は新しいアプリ内通知チャネルを作成する
func NewAppChannel(userID uint) *AppChannel {
	return &AppChannel{
		UserID: userID,
	}
}

// LineChannel はLINE通知チャネル
type LineChannel struct {
	UserID      uint
	LineUserID  string // LINEユーザーID
	AccessToken string // LINEアクセストークン
}

// GetType はチャネルタイプを返す
func (c *LineChannel) GetType() ChannelType {
	return LineMessage
}

// NewLineChannel は新しいLINE通知チャネルを作成する
func NewLineChannel(userID uint, lineUserID, accessToken string) *LineChannel {
	return &LineChannel{
		UserID:      userID,
		LineUserID:  lineUserID,
		AccessToken: accessToken,
	}
}
