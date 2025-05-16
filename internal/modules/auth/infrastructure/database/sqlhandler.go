package databaseInfra

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/modules/auth/interface/database"
)

type SqlHandler struct {
	Conn *sql.DB
}

func NewSqlHandler() database.SqlHandler {
	config, err := config.LoadConfig(".")
	dsn := config.GetDSN()
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}

	// DB接続が確立できてるかを確認
	if err := conn.Ping(); err != nil {
		panic(err.Error())
	}

	fmt.Println("✅ DB接続成功しました!")

	sqlBytes, err := os.ReadFile("mysql/init.sql")
	if err != nil {
		fmt.Printf("❌ SQL読み取り失敗: %v", err)
	}

	if _, err := conn.Exec(string(sqlBytes)); err != nil {
		fmt.Printf("❌ SQL実行失敗: %v", err)
	}

	sqlHandler := new(SqlHandler)
	sqlHandler.Conn = conn
	return sqlHandler
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
