package database

import (
	"database/sql"
	"fmt"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/common/infrastructure/database"
)

// SqlHandler はGroupモジュール用のSQLハンドラー
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

// InitializeTables はGroupモジュール用のテーブルを初期化する
func (h *SqlHandler) InitializeTables() error {
	// グループテーブル
	groupsTableSQL := `
	CREATE TABLE IF NOT EXISTS groups (
		id CHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		type ENUM('PUBLIC', 'PRIVATE', 'SECRET') NOT NULL DEFAULT 'PRIVATE',
		owner_id CHAR(36) NOT NULL,
		settings JSON NULL COMMENT 'グループ設定',
		member_count INT DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_owner_id (owner_id),
		INDEX idx_type (type),
		INDEX idx_name (name),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// グループメンバーテーブル
	groupMembersTableSQL := `
	CREATE TABLE IF NOT EXISTS group_members (
		id CHAR(36) PRIMARY KEY,
		group_id CHAR(36) NOT NULL,
		user_id CHAR(36) NOT NULL,
		role ENUM('OWNER', 'ADMIN', 'MEMBER') NOT NULL DEFAULT 'MEMBER',
		joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_group_id (group_id),
		INDEX idx_user_id (user_id),
		INDEX idx_role (role),
		INDEX idx_joined_at (joined_at),
		UNIQUE KEY unique_group_member (group_id, user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// グループタスクテーブル
	groupTasksTableSQL := `
	CREATE TABLE IF NOT EXISTS group_tasks (
		id CHAR(36) PRIMARY KEY,
		group_id CHAR(36) NOT NULL,
		task_id CHAR(36) NOT NULL,
		assigned_by CHAR(36) NOT NULL,
		assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_group_id (group_id),
		INDEX idx_task_id (task_id),
		INDEX idx_assigned_by (assigned_by),
		INDEX idx_assigned_at (assigned_at),
		UNIQUE KEY unique_group_task (group_id, task_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// テーブル作成
	tables := []string{groupsTableSQL, groupMembersTableSQL, groupTasksTableSQL}

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
		// グループの複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_group_owner_type ON groups (owner_id, type)`,
		`CREATE INDEX IF NOT EXISTS idx_group_type_created ON groups (type, created_at)`,

		// グループメンバーの複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_member_group_role ON group_members (group_id, role)`,
		`CREATE INDEX IF NOT EXISTS idx_member_user_role ON group_members (user_id, role)`,

		// グループタスクの複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_group_task_assigned ON group_tasks (group_id, assigned_at)`,
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