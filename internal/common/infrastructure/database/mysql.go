package database

import (
	"database/sql"
	"fmt"
	"os"
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

	// 初期化スクリプトの実行（必要に応じて）
	if initSQL, err := os.ReadFile("mysql/init.sql"); err == nil {
		if _, err := conn.Exec(string(initSQL)); err != nil {
			fmt.Printf("⚠️ 初期化SQLの実行に失敗しました: %v\n", err)
			// 致命的ではないのでエラーは返さない
		}
	}

	return conn, nil
}
