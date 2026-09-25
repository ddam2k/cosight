package database

import (
	"context"
	_ "embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/wasming/cosight/internal/config"
)

//go:embed schema.sql
var schemaSQL string

func OpenPostgreSQL(ctx context.Context, cfg config.PostgreSQL) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	if err := EnsureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// EnsureSchema creates the core tables when they do not exist. The transaction-level
// advisory lock serializes startup across multiple API instances sharing a database.
func EnsureSchema(ctx context.Context, db *sqlx.DB) (resultErr error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin schema initialization: %w", err)
	}
	defer func() {
		if resultErr != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('cosight.schema'))`); err != nil {
		return fmt.Errorf("lock schema initialization: %w", err)
	}
	if _, err = tx.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("initialize PostgreSQL schema: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit schema initialization: %w", err)
	}
	return nil
}
