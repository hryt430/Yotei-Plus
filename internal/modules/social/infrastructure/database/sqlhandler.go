package database

import (
	"database/sql"
	"fmt"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/common/infrastructure/database"
)

// SqlHandler はSocialモジュール用のSQLハンドラー
type SqlHandler struct {
	Conn *sql.DB
}

// NewSqlHandler は新しいSqlHandlerを作成する
func NewSqlHandler() SqlHandler {
	// 共通のMySQLコネクションを使用
	cfg, err := config.LoadConfig("")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	conn, err := database.NewMySQLConnection(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	return SqlHandler{
		Conn: conn,
	}
}

// Close はデータベース接続を閉じる
func (h *SqlHandler) Close() error {
	if h.Conn != nil {
		return h.Conn.Close()
	}
	return nil
}

// GetConnection はデータベース接続を取得する
func (h *SqlHandler) GetConnection() *sql.DB {
	return h.Conn
}

// Begin はトランザクションを開始する
func (h *SqlHandler) Begin() (*sql.Tx, error) {
	return h.Conn.Begin()
}

// ExecInTransaction はトランザクション内でクエリを実行する
func (h *SqlHandler) ExecInTransaction(txFunc func(*sql.Tx) error) error {
	tx, err := h.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = txFunc(tx)
	return err
}

// InitializeTables はSocialモジュール用のテーブルを初期化する
func (h *SqlHandler) InitializeTables() error {
	// 友達関係テーブル
	friendshipsTableSQL := `
	CREATE TABLE IF NOT EXISTS friendships (
		id CHAR(36) PRIMARY KEY,
		requester_id CHAR(36) NOT NULL,
		addressee_id CHAR(36) NOT NULL,
		status ENUM('PENDING', 'ACCEPTED', 'BLOCKED') NOT NULL DEFAULT 'PENDING',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		accepted_at TIMESTAMP NULL,
		blocked_at TIMESTAMP NULL,
		INDEX idx_requester_id (requester_id),
		INDEX idx_addressee_id (addressee_id),
		INDEX idx_status (status),
		INDEX idx_created_at (created_at),
		UNIQUE KEY unique_friendship (requester_id, addressee_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 招待テーブル
	invitationsTableSQL := `
	CREATE TABLE IF NOT EXISTS invitations (
		id CHAR(36) PRIMARY KEY,
		type ENUM('FRIEND', 'GROUP') NOT NULL,
		method ENUM('IN_APP', 'CODE', 'URL') NOT NULL,
		status ENUM('PENDING', 'ACCEPTED', 'DECLINED', 'EXPIRED', 'CANCELED') NOT NULL DEFAULT 'PENDING',
		inviter_id CHAR(36) NOT NULL,
		invitee_id CHAR(36) NULL,
		invitee_info JSON NULL COMMENT '未登録ユーザーの招待情報',
		target_id CHAR(36) NULL COMMENT 'グループ招待の場合のグループID',
		code VARCHAR(255) NULL COMMENT '招待コード',
		url TEXT NULL COMMENT '招待URL',
		message TEXT NOT NULL DEFAULT '',
		metadata JSON NULL COMMENT '追加のメタデータ',
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		accepted_at TIMESTAMP NULL,
		INDEX idx_inviter_id (inviter_id),
		INDEX idx_invitee_id (invitee_id),
		INDEX idx_type (type),
		INDEX idx_method (method),
		INDEX idx_status (status),
		INDEX idx_code (code),
		INDEX idx_expires_at (expires_at),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// テーブル作成
	tables := []string{friendshipsTableSQL, invitationsTableSQL}

	for _, tableSQL := range tables {
		if _, err := h.Conn.Exec(tableSQL); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// CreateIndexes は必要なインデックスを作成する
func (h *SqlHandler) CreateIndexes() error {
	indexes := []string{
		// 友達関係の複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_friendship_users ON friendships (requester_id, addressee_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_friendship_status_updated ON friendships (status, updated_at)`,

		// 招待の複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_invitation_type_status ON invitations (type, status)`,
		`CREATE INDEX IF NOT EXISTS idx_invitation_expires_status ON invitations (expires_at, status)`,
		`CREATE INDEX IF NOT EXISTS idx_invitation_target ON invitations (target_id, type, status)`,
	}

	for _, indexSQL := range indexes {
		if _, err := h.Conn.Exec(indexSQL); err != nil {
			// インデックス作成エラーは警告レベル（既に存在する場合など）
			fmt.Printf("Warning: Failed to create index: %v\n", err)
		}
	}

	return nil
}

// HealthCheck はデータベース接続の健全性をチェックする
func (h *SqlHandler) HealthCheck() error {
	return h.Conn.Ping()
}

// GetStats はデータベース統計情報を取得する
func (h *SqlHandler) GetStats() sql.DBStats {
	return h.Conn.Stats()
}
