package storage

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"ordermonitor/internal/config"
)

func NewMySQL(cfg config.MySQLConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(60 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

