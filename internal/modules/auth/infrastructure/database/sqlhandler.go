package databaseInfra

import (
	"database/sql"
	"fmt"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/modules/auth/interface/database"
)

type SqlHandler struct {
	Conn *sql.DB
}

// インターフェースは既存のまま維持する想定

func NewSqlHandler() SqlHandler {
	config, err := config.LoadConfig(".")
	if err != nil {
		panic(err.Error())
	}

	// common/databaseのハンドラーを取得
	commonHandler, err := commonDB.NewMySQLHandler(config)
	if err != nil {
		panic(err.Error()) // または既存の処理に合わせたエラーハンドリング
	}

	// common/databaseのコネクションを取得して、infrastructure側のSqlHandlerに設定
	sqlHandler := new(SqlHandler)
	sqlHandler.Conn = commonHandler.(*commonDB.MySQLHandler).Conn

	return *sqlHandler
}

func (h *SqlHandler) Execute(statement string, args ...interface{}) (database.Result, error) {
	res, err := h.Conn.Exec(statement, args...)
	if err != nil {
		return nil, fmt.Errorf("ステートメント実行失敗: %w", err)
	}
	return &SqlResult{res}, nil
}

func (h *SqlHandler) Query(statement string, args ...interface{}) (database.Row, error) {
	rows, err := h.Conn.Query(statement, args...)
	if err != nil {
		return nil, fmt.Errorf("クエリ実行失敗: %w", err)
	}
	return &SqlRow{Rows: rows}, nil
}

func (h *SqlHandler) Close() error {
	return h.Conn.Close()
}

type SqlResult struct {
	Result sql.Result
}

func (r *SqlResult) LastInsertId() (int64, error) {
	return r.Result.LastInsertId()
}

func (r *SqlResult) RowsAffected() (int64, error) {
	return r.Result.RowsAffected()
}

type SqlRow struct {
	Rows *sql.Rows
}

func (r *SqlRow) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r *SqlRow) Next() bool {
	return r.Rows.Next()
}

func (r *SqlRow) Close() error {
	return r.Rows.Close()
}
