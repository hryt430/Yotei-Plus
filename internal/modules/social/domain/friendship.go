package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

// FriendshipStatus は友達関係のステータス
type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "PENDING"  // 承認待ち
	FriendshipStatusAccepted FriendshipStatus = "ACCEPTED" // 承認済み
	FriendshipStatusBlocked  FriendshipStatus = "BLOCKED"  // ブロック
)

// Friendship は友達関係を表すドメインエンティティ
type Friendship struct {
	ID          uuid.UUID        `json:"id"`
	RequesterID uuid.UUID        `json:"requester_id"` // 申請者
	AddresseeID uuid.UUID        `json:"addressee_id"` // 申請先
	Status      FriendshipStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	AcceptedAt  *time.Time       `json:"accepted_at,omitempty"`
	BlockedAt   *time.Time       `json:"blocked_at,omitempty"`
}

// NewFriendship は新しい友達申請を作成する
func NewFriendship(requesterID, addresseeID uuid.UUID) *Friendship {
	now := time.Now()
	return &Friendship{
		ID:          uuid.New(),
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      FriendshipStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Accept は友達申請を承認する
func (f *Friendship) Accept() {
	f.Status = FriendshipStatusAccepted
	now := time.Now()
	f.AcceptedAt = &now
	f.UpdatedAt = now
}

// Block はユーザーをブロックする
func (f *Friendship) Block() {
	f.Status = FriendshipStatusBlocked
	now := time.Now()
	f.BlockedAt = &now
	f.UpdatedAt = now
}

// IsFriend は友達関係が成立しているかチェック
func (f *Friendship) IsFriend() bool {
	return f.Status == FriendshipStatusAccepted
}

// IsBlocked はブロックされているかチェック
func (f *Friendship) IsBlocked() bool {
	return f.Status == FriendshipStatusBlocked
}

// InvitationType は招待の種類
type InvitationType string

const (
	InvitationTypeFriend InvitationType = "FRIEND" // 友達招待
	InvitationTypeGroup  InvitationType = "GROUP"  // グループ招待
)

// InvitationMethod は招待方法
type InvitationMethod string

const (
	MethodInApp InvitationMethod = "IN_APP" // アプリ内通知
	MethodCode  InvitationMethod = "CODE"   // 招待コード
	MethodURL   InvitationMethod = "URL"    // URL共有
)

// InvitationStatus は招待のステータス
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "PENDING"  // 送信済み（未処理）
	InvitationStatusAccepted InvitationStatus = "ACCEPTED" // 承認済み
	InvitationStatusDeclined InvitationStatus = "DECLINED" // 拒否
	InvitationStatusExpired  InvitationStatus = "EXPIRED"  // 期限切れ
	InvitationStatusCanceled InvitationStatus = "CANCELED" // キャンセル
)

// Invitation は招待を表すドメインエンティティ
type Invitation struct {
	ID          uuid.UUID        `json:"id"`
	Type        InvitationType   `json:"type"`
	Method      InvitationMethod `json:"method"`
	Status      InvitationStatus `json:"status"`
	InviterID   uuid.UUID        `json:"inviter_id"`             // 招待者
	InviteeID   *uuid.UUID       `json:"invitee_id,omitempty"`   // 被招待者（登録済みの場合）
	InviteeInfo *InviteeInfo     `json:"invitee_info,omitempty"` // 被招待者情報（未登録の場合）

	// 招待対象（friend招待の場合は空、group招待の場合はgroupID）
	TargetID *uuid.UUID `json:"target_id,omitempty"`

	// 招待コード・URL用
	Code string `json:"code,omitempty"`
	URL  string `json:"url,omitempty"`

	// メタデータ
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`

	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

// InviteeInfo は未登録ユーザーの招待情報
type InviteeInfo struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// NewInvitation は新しい招待を作成する
func NewInvitation(
	invitationType InvitationType,
	method InvitationMethod,
	inviterID uuid.UUID,
	message string,
	expirationHours int,
) *Invitation {
	now := time.Now()
	invitation := &Invitation{
		ID:        uuid.New(),
		Type:      invitationType,
		Method:    method,
		Status:    InvitationStatusPending,
		InviterID: inviterID,
		Message:   message,
		Metadata:  make(map[string]string),
		ExpiresAt: now.Add(time.Duration(expirationHours) * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 招待コード・URLの生成
	if method == MethodCode || method == MethodURL {
		invitation.Code = generateInvitationCode()
		if method == MethodURL {
			invitation.URL = generateInvitationURL(invitation)
		}
	}

	return invitation
}

// generateInvitationCode は招待コードを生成する
func generateInvitationCode() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateInvitationURL は招待URL用データを生成する
func generateInvitationURL(invitation *Invitation) string {
	// 実際の実装では、フロントエンドのURLを使用
	return "https://yotei-plus.com/invite/" + invitation.Code
}

// SetInvitee は被招待者を設定する（登録済みユーザー）
func (i *Invitation) SetInvitee(userID uuid.UUID) {
	i.InviteeID = &userID
	i.UpdatedAt = time.Now()
}

// SetInviteeInfo は被招待者情報を設定する（未登録ユーザー）
func (i *Invitation) SetInviteeInfo(info InviteeInfo) {
	i.InviteeInfo = &info
	i.UpdatedAt = time.Now()
}

// SetTarget は招待対象を設定する（グループ招待の場合）
func (i *Invitation) SetTarget(targetID uuid.UUID) {
	i.TargetID = &targetID
	i.UpdatedAt = time.Now()
}

// Accept は招待を承認する
func (i *Invitation) Accept() error {
	if i.IsExpired() {
		return ErrInvitationExpired
	}
	if i.Status != InvitationStatusPending {
		return ErrInvalidInvitationStatus
	}

	i.Status = InvitationStatusAccepted
	now := time.Now()
	i.AcceptedAt = &now
	i.UpdatedAt = now
	return nil
}

// Decline は招待を拒否する
func (i *Invitation) Decline() error {
	if i.Status != InvitationStatusPending {
		return ErrInvalidInvitationStatus
	}

	i.Status = InvitationStatusDeclined
	i.UpdatedAt = time.Now()
	return nil
}

// Cancel は招待をキャンセルする
func (i *Invitation) Cancel() error {
	if i.Status != InvitationStatusPending {
		return ErrInvalidInvitationStatus
	}

	i.Status = InvitationStatusCanceled
	i.UpdatedAt = time.Now()
	return nil
}

// MarkAsExpired は招待を期限切れにする
func (i *Invitation) MarkAsExpired() {
	i.Status = InvitationStatusExpired
	i.UpdatedAt = time.Now()
}

// IsExpired は招待が期限切れかチェック
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsValid は招待が有効かチェック
func (i *Invitation) IsValid() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

// AddMetadata はメタデータを追加する
func (i *Invitation) AddMetadata(key, value string) {
	if i.Metadata == nil {
		i.Metadata = make(map[string]string)
	}
	i.Metadata[key] = value
	i.UpdatedAt = time.Now()
}

// エラー定義
var (
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvalidInvitationStatus = errors.New("invalid invitation status")
)
