package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/hryt430/Yotei+/config"
	"github.com/hryt430/Yotei+/internal/modules/notification/interface/database"
)

type SqlHandler struct {
	Conn *sql.DB
}

func NewSqlHandler() database.SqlHandler {
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

func (handler *SqlHandler) Execute(statement string, args ...interface{}) (database.Result, error) {
	res := SqlResult{}
	result, err := handler.Conn.Exec(statement, args...)
	if err != nil {
		return res, err
	}
	res.Result = result
	return res, nil
}

func (handler *SqlHandler) Query(statement string, args ...interface{}) (database.Row, error) {
	rows, err := handler.Conn.Query(statement, args...)
	if err != nil {
		return new(SqlRow), err
	}
	row := new(SqlRow)
	row.Rows = rows
	return row, nil
}

func (handler *SqlHandler) Close() error {
	return handler.Conn.Close()
}

type SqlResult struct {
	Result sql.Result
}

func (r SqlResult) LastInsertId() (int64, error) {
	return r.Result.LastInsertId()
}

func (r SqlResult) RowsAffected() (int64, error) {
	return r.Result.RowsAffected()
}

type SqlRow struct {
	Rows *sql.Rows
}

func (r SqlRow) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r SqlRow) Next() bool {
	return r.Rows.Next()
}

func (r SqlRow) Close() error {
	return r.Rows.Close()
}
