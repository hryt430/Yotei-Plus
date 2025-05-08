package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config はデータベース接続設定を表します
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewMySQLConnection は新しいMySQL接続を作成します
func NewMySQLConnection(config Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("データベース接続オープンエラー: %w", err)
	}

	// 接続設定
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	// 接続確認
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("データベース接続確認エラー: %w", err)
	}

	return db, nil
}

// Transaction はトランザクションを扱うヘルパー関数です
func Transaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
