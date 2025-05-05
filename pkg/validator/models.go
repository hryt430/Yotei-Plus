package validator

import (
	"time"
)

// UserValidator はユーザー関連のバリデーションを行う構造体
type UserValidator struct {
	ID              string
	Username        string
	Email           string
	Password        string
	ConfirmPassword string
	FirstName       string
	LastName        string
	PhoneNumber     string
	BirthDate       time.Time
}

// Validate はUserValidatorのフィールドを検証します
func (u *UserValidator) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	// ユーザー名のバリデーション
	if String.IsEmpty(u.Username) {
		errors.Add("username", "ユーザー名は必須です")
	} else if !String.Length(u.Username, 3, 30) {
		errors.Add("username", "ユーザー名は3〜30文字である必要があります")
	} else if !String.IsUsername(u.Username) {
		errors.Add("username", "ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます")
	}

	// メールアドレスのバリデーション
	if String.IsEmpty(u.Email) {
		errors.Add("email", "メールアドレスは必須です")
	} else if !String.IsEmail(u.Email) {
		errors.Add("email", "有効なメールアドレスを入力してください")
	}

	// パスワードのバリデーション
	if String.IsEmpty(u.Password) {
		errors.Add("password", "パスワードは必須です")
	} else if !String.MinLength(u.Password, 8) {
		errors.Add("password", "パスワードは8文字以上である必要があります")
	} else if !String.HasUpperCase(u.Password) || !String.HasLowerCase(u.Password) ||
		!String.HasNumber(u.Password) || !String.HasSpecialChar(u.Password) {
		errors.Add("password", "パスワードには大文字、小文字、数字、特殊文字をそれぞれ1つ以上含める必要があります")
	}

	// パスワード確認のバリデーション
	if u.Password != u.ConfirmPassword {
		errors.Add("confirmPassword", "パスワードが一致しません")
	}

	// 氏名のバリデーション
	if String.IsNotEmpty(u.FirstName) && !String.MaxLength(u.FirstName, 50) {
		errors.Add("firstName", "名前は50文字以内である必要があります")
	}
	if String.IsNotEmpty(u.LastName) && !String.MaxLength(u.LastName, 50) {
		errors.Add("lastName", "苗字は50文字以内である必要があります")
	}

	// 電話番号のバリデーション（任意）
	if String.IsNotEmpty(u.PhoneNumber) && !String.IsPhoneNumber(u.PhoneNumber) {
		errors.Add("phoneNumber", "有効な電話番号を入力してください")
	}

	// 生年月日のバリデーション（任意）
	if !u.BirthDate.IsZero() && !Date.IsPast(u.BirthDate) {
		errors.Add("birthDate", "生年月日は過去の日付である必要があります")
	}

	return errors
}

// TaskValidator はタスク関連のバリデーションを行う構造体
type TaskValidator struct {
	ID          string
	Title       string
	Description string
	DueDate     time.Time
	Priority    int
	Status      string
	Tags        []string
	AssigneeIDs []string
}

// Validate はTaskValidatorのフィールドを検証します
func (t *TaskValidator) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	// タイトルのバリデーション
	if String.IsEmpty(t.Title) {
		errors.Add("title", "タイトルは必須です")
	} else if !String.Length(t.Title, 1, 100) {
		errors.Add("title", "タイトルは1〜100文字である必要があります")
	}

	// 説明のバリデーション（任意）
	if String.IsNotEmpty(t.Description) && !String.MaxLength(t.Description, 1000) {
		errors.Add("description", "説明は1000文字以内である必要があります")
	}

	// 期日のバリデーション
	if !t.DueDate.IsZero() && Date.IsPast(t.DueDate) && !Date.IsToday(t.DueDate) {
		errors.Add("dueDate", "期日は今日以降の日付である必要があります")
	}

	// 優先度のバリデーション
	if t.Priority != 0 && !Number.Between(t.Priority, 1, 5) {
		errors.Add("priority", "優先度は1〜5の間である必要があります")
	}

	// ステータスのバリデーション
	validStatuses := []string{"未着手", "進行中", "完了", "保留"}
	isValidStatus := false
	for _, status := range validStatuses {
		if t.Status == status {
			isValidStatus = true
			break
		}
	}
	if String.IsNotEmpty(t.Status) && !isValidStatus {
		errors.Add("status", "無効なステータスです")
	}

	// タグのバリデーション
	if len(t.Tags) > 0 {
		// 各タグをバリデーション
		for _, tag := range t.Tags {
			if String.IsEmpty(tag) {
				errors.Add("tags", "空のタグは許可されていません")
				break
			}
			if !String.MaxLength(tag, 30) {
				errors.Add("tags", "タグは30文字以内である必要があります")
				break
			}
		}

		// 重複チェック
		arrValidator := ArrayValidator{}
		if !arrValidator.NoDuplicateStrings(t.Tags) {
			errors.Add("tags", "タグに重複があります")
		}
	}

	// 担当者IDのバリデーション
	if len(t.AssigneeIDs) > 0 {
		// 各担当者IDをバリデーション
		for _, assigneeID := range t.AssigneeIDs {
			if String.IsEmpty(assigneeID) {
				errors.Add("assigneeIDs", "無効な担当者IDがあります")
				break
			}
		}

		// 重複チェック
		arrValidator := ArrayValidator{}
		if !arrValidator.NoDuplicateStrings(t.AssigneeIDs) {
			errors.Add("assigneeIDs", "担当者IDに重複があります")
		}
	}

	return errors
}

// GroupValidator はグループ関連のバリデーションを行う構造体
type GroupValidator struct {
	ID          string
	Name        string
	Description string
	MemberIDs   []string
	AdminIDs    []string
}

// Validate はGroupValidatorのフィールドを検証します
func (g *GroupValidator) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	// グループ名のバリデーション
	if String.IsEmpty(g.Name) {
		errors.Add("name", "グループ名は必須です")
	} else if !String.Length(g.Name, 3, 50) {
		errors.Add("name", "グループ名は3〜50文字である必要があります")
	}

	// 説明のバリデーション（任意）
	if String.IsNotEmpty(g.Description) && !String.MaxLength(g.Description, 500) {
		errors.Add("description", "説明は500文字以内である必要があります")
	}

	// メンバーIDのバリデーション
	if len(g.MemberIDs) == 0 {
		errors.Add("memberIDs", "グループには少なくとも1人のメンバーが必要です")
	} else {
		arrValidator := ArrayValidator{}
		if !arrValidator.NoDuplicateStrings(g.MemberIDs) {
			errors.Add("memberIDs", "メンバーIDに重複があります")
		}

		// 各メンバーIDをバリデーション
		for _, memberID := range g.MemberIDs {
			if String.IsEmpty(memberID) {
				errors.Add("memberIDs", "無効なメンバーIDがあります")
				break
			}
		}
	}

	// 管理者IDのバリデーション
	if len(g.AdminIDs) == 0 {
		errors.Add("adminIDs", "グループには少なくとも1人の管理者が必要です")
	} else {
		arrValidator := ArrayValidator{}
		if !arrValidator.NoDuplicateStrings(g.AdminIDs) {
			errors.Add("adminIDs", "管理者IDに重複があります")
		}

		// 各管理者IDをバリデーション
		for _, adminID := range g.AdminIDs {
			if String.IsEmpty(adminID) {
				errors.Add("adminIDs", "無効な管理者IDがあります")
				break
			}
		}

		// すべての管理者がメンバーであることを確認
		for _, adminID := range g.AdminIDs {
			isAdmin := false
			for _, memberID := range g.MemberIDs {
				if adminID == memberID {
					isAdmin = true
					break
				}
			}
			if !isAdmin {
				errors.Add("adminIDs", "管理者はグループのメンバーである必要があります")
				break
			}
		}
	}

	return errors
}

// EventValidator はイベント/予定関連のバリデーションを行う構造体
type EventValidator struct {
	ID                string
	Title             string
	Description       string
	StartTime         time.Time
	EndTime           time.Time
	Location          string
	ParticipantIDs    []string
	IsAllDay          bool
	Recurring         bool
	RecurrencePattern string
}

// Validate はEventValidatorのフィールドを検証します
func (e *EventValidator) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	// タイトルのバリデーション
	if String.IsEmpty(e.Title) {
		errors.Add("title", "タイトルは必須です")
	} else if !String.Length(e.Title, 1, 100) {
		errors.Add("title", "タイトルは1〜100文字である必要があります")
	}

	// 説明のバリデーション（任意）
	if String.IsNotEmpty(e.Description) && !String.MaxLength(e.Description, 1000) {
		errors.Add("description", "説明は1000文字以内である必要があります")
	}

	// 開始時間のバリデーション
	if e.StartTime.IsZero() {
		errors.Add("startTime", "開始時間は必須です")
	}

	// 終了時間のバリデーション
	if e.EndTime.IsZero() {
		errors.Add("endTime", "終了時間は必須です")
	} else if !e.StartTime.IsZero() && !e.EndTime.After(e.StartTime) {
		errors.Add("endTime", "終了時間は開始時間より後である必要があります")
	}

	// 場所のバリデーション（任意）
	if String.IsNotEmpty(e.Location) && !String.MaxLength(e.Location, 200) {
		errors.Add("location", "場所は200文字以内である必要があります")
	}

	// 参加者IDのバリデーション
	if len(e.ParticipantIDs) > 0 {
		arrValidator := ArrayValidator{}
		if !arrValidator.NoDuplicateStrings(e.ParticipantIDs) {
			errors.Add("participantIDs", "参加者IDに重複があります")
		}

		// 各参加者IDをバリデーション
		for _, participantID := range e.ParticipantIDs {
			if String.IsEmpty(participantID) {
				errors.Add("participantIDs", "無効な参加者IDがあります")
				break
			}
		}
	}

	// 繰り返しパターンのバリデーション
	if e.Recurring {
		validPatterns := []string{"daily", "weekly", "biweekly", "monthly", "yearly"}
		isValidPattern := false
		for _, pattern := range validPatterns {
			if e.RecurrencePattern == pattern {
				isValidPattern = true
				break
			}
		}
		if !isValidPattern {
			errors.Add("recurrencePattern", "無効な繰り返しパターンです")
		}
	}

	return errors
}

// FriendshipValidator は友達関係のバリデーションを行う構造体
type FriendshipValidator struct {
	ID        string
	UserID    string
	FriendID  string
	Status    string
	CreatedAt time.Time
}

// Validate はFriendshipValidatorのフィールドを検証します
func (f *FriendshipValidator) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	// ユーザーIDのバリデーション
	if String.IsEmpty(f.UserID) {
		errors.Add("userID", "ユーザーIDは必須です")
	}

	// 友達IDのバリデーション
	if String.IsEmpty(f.FriendID) {
		errors.Add("friendID", "友達IDは必須です")
	} else if f.FriendID == f.UserID {
		errors.Add("friendID", "自分自身を友達に追加することはできません")
	}

	// ステータスのバリデーション
	validStatuses := []string{"pending", "accepted", "rejected", "blocked"}
	isValidStatus := false
	for _, status := range validStatuses {
		if f.Status == status {
			isValidStatus = true
			break
		}
	}
	if String.IsEmpty(f.Status) {
		errors.Add("status", "ステータスは必須です")
	} else if !isValidStatus {
		errors.Add("status", "無効なステータスです")
	}

	return errors
}

// NotificationValidator は通知のバリデーションを行う構造体
type NotificationValidator struct {
	ID        string
	UserID    string
	Type      string
	Title     string
	Message   string
	IsRead    bool
	RelatedID string
	CreatedAt time.Time
}

// Validate はNotificationValidatorのフィールドを検証します
func (n *NotificationValidator) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	// ユーザーIDのバリデーション
	if String.IsEmpty(n.UserID) {
		errors.Add("userID", "ユーザーIDは必須です")
	}

	// タイプのバリデーション
	validTypes := []string{"task", "friend_request", "event_invitation", "reminder", "system"}
	isValidType := false
	for _, typ := range validTypes {
		if n.Type == typ {
			isValidType = true
			break
		}
	}
	if String.IsEmpty(n.Type) {
		errors.Add("type", "通知タイプは必須です")
	} else if !isValidType {
		errors.Add("type", "無効な通知タイプです")
	}

	// タイトルのバリデーション
	if String.IsEmpty(n.Title) {
		errors.Add("title", "タイトルは必須です")
	} else if !String.MaxLength(n.Title, 100) {
		errors.Add("title", "タイトルは100文字以内である必要があります")
	}

	// メッセージのバリデーション
	if String.IsEmpty(n.Message) {
		errors.Add("message", "メッセージは必須です")
	} else if !String.MaxLength(n.Message, 500) {
		errors.Add("message", "メッセージは500文字以内である必要があります")
	}

	return errors
}

// JWTClaimsValidator はJWTクレームのバリデーションを行う構造体
type JWTClaimsValidator struct {
	UserID    string
	Username  string
	Email     string
	ExpiresAt int64
	IssuedAt  int64
}

// Validate はJWTClaimsValidatorのフィールドを検証します
func (j *JWTClaimsValidator) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	// ユーザーIDのバリデーション
	if String.IsEmpty(j.UserID) {
		errors.Add("userID", "ユーザーIDは必須です")
	}

	// ユーザー名のバリデーション
	if String.IsEmpty(j.Username) {
		errors.Add("username", "ユーザー名は必須です")
	}

	// メールアドレスのバリデーション
	if String.IsEmpty(j.Email) {
		errors.Add("email", "メールアドレスは必須です")
	} else if !String.IsEmail(j.Email) {
		errors.Add("email", "有効なメールアドレスを入力してください")
	}

	// 有効期限のバリデーション
	now := time.Now().Unix()
	if j.ExpiresAt <= now {
		errors.Add("expiresAt", "トークンの有効期限が切れています")
	}

	// 発行時刻のバリデーション
	if j.IssuedAt > now {
		errors.Add("issuedAt", "トークンの発行時刻が未来です")
	}

	return errors
}
