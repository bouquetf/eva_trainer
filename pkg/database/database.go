package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	globalDB   *sql.DB
	globalDBMu sync.RWMutex
)

func SetDB(db *sql.DB) {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()
	globalDB = db
}

func DB() *sql.DB {
	globalDBMu.RLock()
	defer globalDBMu.RUnlock()
	return globalDB
}

func Connect(cfg Config) (*sql.DB, error) {
	var dsn string
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=%d",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode, int(cfg.Timeout.Seconds()),
		)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func HealthCheck(db *sql.DB) error {
	return db.Ping()
}
