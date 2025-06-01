package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hryt430/Yotei+/config"
)

func NewMySQLConnection(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	fmt.Printf("DSN: %q\n", dsn)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 接続確認
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// コネクションプールの設定
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("✅ DB接続成功しました!")

	// 初期化スクリプトの実行
	if err := executeInitSQL(conn, "mysql/init.sql"); err != nil {
		fmt.Printf("⚠️ 初期化SQLの実行に失敗しました: %v\n", err)
		// 必要に応じてここでエラーを返すかどうか決める
		// return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return conn, nil
}

// 初期化SQLを実行する関数
func executeInitSQL(db *sql.DB, filepath string) error {
	// 既に初期化済みかチェック
	if isAlreadyInitialized(db) {
		fmt.Println("✅ データベースは既に初期化済みです。スキップします。")
		return nil
	}

	// ファイルの存在確認
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fmt.Printf("初期化ファイル %s が見つかりません。スキップします。\n", filepath)
		return nil
	}

	// ファイル読み込み
	initSQL, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read init SQL file: %w", err)
	}

	// SQL文を分割して実行
	statements := splitSQLStatements(string(initSQL))

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		fmt.Printf("実行中 [%d/%d]: %.50s...\n", i+1, len(statements), stmt)

		if _, err := db.Exec(stmt); err != nil {
			// エラーが発生した場合、どの文でエラーになったかを明示
			fmt.Printf("❌ SQL実行エラー (文 %d): %v\n", i+1, err)
			fmt.Printf("問題のSQL: %s\n", stmt)

			// 一部のエラーは無視する（例：既にテーブルが存在する場合など）
			if isIgnorableError(err) {
				fmt.Printf("⚠️ エラーを無視して継続します\n")
				continue
			}

			return fmt.Errorf("failed to execute SQL statement %d: %w", i+1, err)
		}
	}

	fmt.Println("✅ 初期化SQL実行完了")
	return nil
}

// データベースが既に初期化済みかチェック
func isAlreadyInitialized(db *sql.DB) bool {
	// usersテーブルの存在をチェック
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'Yotei-Plus' AND table_name = 'users'"
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// SQL文を分割する関数
func splitSQLStatements(sqlContent string) []string {
	// コメント行を除去
	lines := strings.Split(sqlContent, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 空行やコメント行をスキップ
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	// 再結合してセミコロンで分割
	cleanSQL := strings.Join(cleanLines, " ")
	statements := strings.Split(cleanSQL, ";")

	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}

// 無視可能なエラーかどうかを判定
func isIgnorableError(err error) bool {
	errStr := strings.ToLower(err.Error())
	ignorableErrors := []string{
		"table already exists",
		"database exists",
		"duplicate entry",
		"duplicate key name", // INDEXの重複エラーを追加
		"key already exists",
	}

	for _, ignorable := range ignorableErrors {
		if strings.Contains(errStr, ignorable) {
			return true
		}
	}

	return false
}
