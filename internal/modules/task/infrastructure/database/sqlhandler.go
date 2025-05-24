package databaseInfra

import (
	"database/sql"
	"fmt"

	"github.com/hryt430/Yotei+/config"
	commonDB "github.com/hryt430/Yotei+/internal/common/infrastructure/database"
	"github.com/hryt430/Yotei+/internal/modules/task/interface/database"
)

type SqlHandler struct {
	Conn *sql.DB
}

func NewSqlHandler() SqlHandler {
	config, err := config.LoadConfig(".")
	if err != nil {
		panic(err.Error())
	}

	// common/databaseからDBコネクションを取得
	conn, err := commonDB.NewMySQLConnection(config)
	if err != nil {
		panic(err.Error())
	}

	sqlHandler := new(SqlHandler)
	sqlHandler.Conn = conn
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
