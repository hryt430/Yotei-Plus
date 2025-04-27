package databaseInfra

import (
	"auth-service/internal/interface/database"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type GormHandler struct {
	DB *gorm.DB
}

func NewGormHandler() database.SqlHandler {
	// dsn := config.GetDSN()
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	panic(fmt.Sprintf("DB接続失敗: %v", err))
	// }

	// fmt.Println("✅ GORM経由でDB接続成功しました!")

	// // 初期化SQLを流したい場合（ただしGORM経由だと生接続が必要）
	// sqlDB, err := db.DB()
	// if err != nil {
	// 	panic(fmt.Sprintf("GORMからsql.DB取得失敗: %v", err))
	// }

	// sqlBytes, err := os.ReadFile("mysql/init.sql")
	// if err != nil {
	// 	fmt.Printf("❌ SQL読み取り失敗: %v", err)
	// }

	// if _, err := sqlDB.Exec(string(sqlBytes)); err != nil {
	// 	fmt.Printf("❌ SQL実行失敗: %v", err)
	// }

	// return &GormHandler{DB: db}
}

func (handler *GormHandler) Execute(statement string, args ...interface{}) (database.Result, error) {
	sqlDB, err := handler.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("DB取得失敗: %w", err)
	}
	res, err := sqlDB.Exec(statement, args...)
	if err != nil {
		return nil, err
	}
	return &GormResult{Result: res}, nil
}

func (handler *GormHandler) Query(statement string, args ...interface{}) (database.Row, error) {
	sqlDB, err := handler.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("DB取得失敗: %w", err)
	}
	rows, err := sqlDB.Query(statement, args...)
	if err != nil {
		return nil, err
	}
	return &GormRow{Rows: rows}, nil
}

func (handler *GormHandler) Close() error {
	sqlDB, err := handler.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

type GormResult struct {
	Result sql.Result
}

func (r *GormResult) LastInsertId() (int64, error) {
	return r.Result.LastInsertId()
}

func (r *GormResult) RowsAffected() (int64, error) {
	return r.Result.RowsAffected()
}

type GormRow struct {
	Rows *sql.Rows
}

func (r *GormRow) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r *GormRow) Next() bool {
	return r.Rows.Next()
}

func (r *GormRow) Close() error {
	return r.Rows.Close()
}
