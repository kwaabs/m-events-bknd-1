package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug" // 👈 add this

	"github.com/kwaabs/m-events/internal/config"
)

// New opens a Bun/PostgreSQL connection, verifies it, and returns the DB handle.
func New(cfg config.DatabaseConfig) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DSN)))

	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(5)
	sqldb.SetConnMaxLifetime(30 * time.Minute)

	db := bun.NewDB(sqldb, pgdialect.New())

	// 👇 ADD DEBUG HOOK HERE
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true), // logs queries + args + timing
	))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil
}
