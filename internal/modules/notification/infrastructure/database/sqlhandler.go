package database

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/hryt430/Yotei+/config"
	commonDB "github.com/hryt430/Yotei+/internal/common/infrastructure/database"
	"github.com/hryt430/Yotei+/internal/modules/notification/interface/database"
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

func (handler *SqlHandler) Execute(statement string, args ...interface{}) (database.Result, error) {
	res := SqlResult{}
	result, err := handler.Conn.Exec(statement, args...)
	if err != nil {
		return res, err
	}
	res.Result = result
	return res, nil
}

func (handler *SqlHandler) Query(statement string, args ...interface{}) (database.Rows, error) {
	rows, err := handler.Conn.Query(statement, args...)
	if err != nil {
		return new(SqlRows), err
	}
	rowsStruct := new(SqlRows)
	rowsStruct.Rows = rows
	return rowsStruct, nil
}

func (handler *SqlHandler) Close() error {
	return handler.Conn.Close()
}

func (handler *SqlHandler) ExecContext(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	result, err := handler.Conn.ExecContext(ctx, query, args...)
	if err != nil {
		return new(SqlResult), err
	}
	return &SqlResult{Result: result}, nil
}

func (handler *SqlHandler) QueryContext(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	rows, err := handler.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return new(SqlRows), err
	}
	return &SqlRows{Rows: rows}, nil
}

func (handler *SqlHandler) QueryRowContext(ctx context.Context, query string, args ...interface{}) database.Row {
	row := handler.Conn.QueryRowContext(ctx, query, args...)
	return &SqlRow{Row: row}
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

type SqlRows struct {
	Rows *sql.Rows
}

func (r SqlRows) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r SqlRows) Next() bool {
	return r.Rows.Next()
}

func (r SqlRows) Close() error {
	return r.Rows.Close()
}

func (r SqlRows) Err() error {
	return r.Rows.Err()
}

type SqlRow struct {
	Row *sql.Row
}

func (r SqlRow) Scan(dest ...interface{}) error {
	return r.Row.Scan(dest...)
}
