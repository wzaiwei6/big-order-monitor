package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"ordermonitor/internal/config"
)

func NewSQLite(cfg config.SQLiteConfig) (*sql.DB, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite dir: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(%d)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", cfg.Path, cfg.BusyTimeoutMS)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if err := ensureSQLiteSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func ensureSQLiteSchema(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS orders_filled (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	symbol TEXT NOT NULL,
	market_type TEXT NOT NULL,
	side TEXT NOT NULL,
	price REAL NOT NULL,
	quantity REAL NOT NULL,
	first_seen INTEGER NOT NULL,
	filled_time INTEGER NOT NULL,
	duration_seconds INTEGER NOT NULL,
	threshold REAL NOT NULL,
	threshold_op TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_filled_unique
ON orders_filled(symbol, market_type, side, price, filled_time, threshold, threshold_op);

CREATE INDEX IF NOT EXISTS idx_orders_filled_lookup
ON orders_filled(symbol, market_type, threshold, threshold_op, filled_time DESC);
`

	_, err := db.Exec(schema)
	return err
}

func CleanupOrdersBefore(db *sql.DB, cutoffUnix int64) (int64, error) {
	result, err := db.Exec(`DELETE FROM orders_filled WHERE filled_time < ?`, cutoffUnix)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func VacuumSQLite(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA wal_checkpoint(TRUNCATE);`); err != nil {
		return err
	}

	_, err := db.Exec(`VACUUM;`)
	return err
}
