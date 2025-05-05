package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// ValidationErrors は複数のバリデーションエラーを格納する構造体
type ValidationErrors struct {
	Errors map[string]string `json:"errors"`
}

// NewValidationErrors は新しいValidationErrorsインスタンスを作成します
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make(map[string]string),
	}
}

// Add はバリデーションエラーを追加します
func (v *ValidationErrors) Add(field, message string) {
	if v.Errors == nil {
		v.Errors = make(map[string]string)
	}
	v.Errors[field] = message
}

// HasErrors はエラーが存在するか確認します
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// ErrorMessages はすべてのエラーメッセージを文字列のスライスとして返します
func (v *ValidationErrors) ErrorMessages() []string {
	messages := make([]string, 0, len(v.Errors))
	for field, message := range v.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, message))
	}
	return messages
}

// Validator はフィールドの検証を行うインターフェース
type Validator interface {
	Validate() *ValidationErrors
}

// StringValidator は文字列のバリデーション関数群
type StringValidator struct{}

// IsEmail は有効なメールアドレスかどうかを検証します
func (StringValidator) IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// MinLength は文字列が指定された最小文字数を満たすか検証します
func (StringValidator) MinLength(s string, min int) bool {
	return utf8.RuneCountInString(s) >= min
}

// MaxLength は文字列が指定された最大文字数を超えないか検証します
func (StringValidator) MaxLength(s string, max int) bool {
	return utf8.RuneCountInString(s) <= max
}

// Length は文字列が指定された文字数範囲内かを検証します
func (StringValidator) Length(s string, min, max int) bool {
	count := utf8.RuneCountInString(s)
	return count >= min && count <= max
}

// MatchesPattern は文字列が正規表現パターンにマッチするかを検証します
func (StringValidator) MatchesPattern(s string, pattern string) bool {
	match, _ := regexp.MatchString(pattern, s)
	return match
}

// IsAlphanumeric は文字列が英数字のみで構成されているかを検証します
func (StringValidator) IsAlphanumeric(s string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s)
}

// IsUsername はユーザー名として有効かを検証します（英数字、アンダースコア、ハイフンのみ）
func (StringValidator) IsUsername(s string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(s)
}

// HasSpecialChar は文字列に特殊文字が含まれているかを検証します
func (StringValidator) HasSpecialChar(s string) bool {
	return regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(s)
}

// HasUpperCase は文字列に大文字が含まれているかを検証します
func (StringValidator) HasUpperCase(s string) bool {
	return regexp.MustCompile(`[A-Z]`).MatchString(s)
}

// HasLowerCase は文字列に小文字が含まれているかを検証します
func (StringValidator) HasLowerCase(s string) bool {
	return regexp.MustCompile(`[a-z]`).MatchString(s)
}

// HasNumber は文字列に数字が含まれているかを検証します
func (StringValidator) HasNumber(s string) bool {
	return regexp.MustCompile(`[0-9]`).MatchString(s)
}

// IsStrongPassword はパスワードが強力かどうかを検証します
func (v StringValidator) IsStrongPassword(password string) bool {
	return v.MinLength(password, 8) &&
		v.HasUpperCase(password) &&
		v.HasLowerCase(password) &&
		v.HasNumber(password) &&
		v.HasSpecialChar(password)
}

// IsEmpty は文字列が空かどうかを検証します
func (StringValidator) IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty は文字列が空でないかどうかを検証します
func (v StringValidator) IsNotEmpty(s string) bool {
	return !v.IsEmpty(s)
}

// IsPhoneNumber は電話番号が有効かどうかを検証します
func (StringValidator) IsPhoneNumber(s string) bool {
	// 基本的な電話番号フォーマット (国によって異なる場合は調整が必要)
	return regexp.MustCompile(`^[+]?[\d\s-]{10,15}$`).MatchString(s)
}

// IsURL はURLが有効かどうかを検証します
func (StringValidator) IsURL(s string) bool {
	pattern := `^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`
	return regexp.MustCompile(pattern).MatchString(s)
}

// NumberValidator は数値のバリデーション関数群
type NumberValidator struct{}

// Min は数値が指定された最小値以上かを検証します
func (NumberValidator) Min(n, min int) bool {
	return n >= min
}

// Max は数値が指定された最大値以下かを検証します
func (NumberValidator) Max(n, max int) bool {
	return n <= max
}

// Between は数値が指定された範囲内かを検証します
func (NumberValidator) Between(n, min, max int) bool {
	return n >= min && n <= max
}

// IsPositive は数値が正の値かを検証します
func (NumberValidator) IsPositive(n int) bool {
	return n > 0
}

// IsNegative は数値が負の値かを検証します
func (NumberValidator) IsNegative(n int) bool {
	return n < 0
}

// IsZero は数値がゼロかを検証します
func (NumberValidator) IsZero(n int) bool {
	return n == 0
}

// DateValidator は日付のバリデーション関数群
type DateValidator struct{}

// IsFuture は日付が未来かを検証します
func (DateValidator) IsFuture(date time.Time) bool {
	return date.After(time.Now())
}

// IsPast は日付が過去かを検証します
func (DateValidator) IsPast(date time.Time) bool {
	return date.Before(time.Now())
}

// IsToday は日付が今日かを検証します
func (DateValidator) IsToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day()
}

// IsAfter は日付が指定された日付より後かを検証します
func (DateValidator) IsAfter(date, after time.Time) bool {
	return date.After(after)
}

// IsBefore は日付が指定された日付より前かを検証します
func (DateValidator) IsBefore(date, before time.Time) bool {
	return date.Before(before)
}

// IsBetween は日付が指定された範囲内かを検証します
func (DateValidator) IsBetween(date, start, end time.Time) bool {
	return (date.After(start) || date.Equal(start)) && (date.Before(end) || date.Equal(end))
}

// IsValidFormat は日付が指定されたフォーマットに一致するかを検証します
func (DateValidator) IsValidFormat(dateStr, layout string) bool {
	_, err := time.Parse(layout, dateStr)
	return err == nil
}

// ArrayValidator は配列のバリデーション関数群
type ArrayValidator struct{}

// MinLength は配列が指定された最小長さを満たすか検証します
func (ArrayValidator) MinLength(arr []interface{}, min int) bool {
	return len(arr) >= min
}

// MaxLength は配列が指定された最大長さを超えないか検証します
func (ArrayValidator) MaxLength(arr []interface{}, max int) bool {
	return len(arr) <= max
}

// ContainsElement は配列に特定の要素が含まれているかを検証します
func (ArrayValidator) ContainsElement(arr []interface{}, element interface{}) bool {
	for _, item := range arr {
		if item == element {
			return true
		}
	}
	return false
}

// NoDuplicates は配列に重複がないかを検証します
func (ArrayValidator) NoDuplicates(arr []interface{}) bool {
	seen := make(map[interface{}]bool)
	for _, item := range arr {
		if seen[item] {
			return false
		}
		seen[item] = true
	}
	return true
}

// NoDuplicateStrings は文字列配列に重複がないかを検証します
func (ArrayValidator) NoDuplicateStrings(arr []string) bool {
	seen := make(map[string]bool)
	for _, item := range arr {
		if seen[item] {
			return false
		}
		seen[item] = true
	}
	return true
}

// NoDuplicateInts は整数配列に重複がないかを検証します
func (ArrayValidator) NoDuplicateInts(arr []int) bool {
	seen := make(map[int]bool)
	for _, item := range arr {
		if seen[item] {
			return false
		}
		seen[item] = true
	}
	return true
}

// Validator はタスク管理アプリのバリデーションを行うためのグローバルインスタンス
var (
	String StringValidator
	Number NumberValidator
	Date   DateValidator
	Array  ArrayValidator
)
