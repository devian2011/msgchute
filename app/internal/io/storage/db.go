package storage

import (
	"context"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	maxIdleConnections = 4
	maxOpenConnections = 10
)

type Config struct {
	Driver       string `env:"APP_DB_DRIVER" yaml:"driver"`
	Dsn          string `env:"APP_DB_DSN" yaml:"dsn"`
	MigrationDir string `env:"APP_DB_MIGRATION_DIR" yaml:"migrationDir"`
}

func NewDB(cfg *Config) (*sqlx.DB, error) {
	database, dbInitErr := sqlx.Open(cfg.Driver, cfg.Dsn)
	if dbInitErr != nil {
		return nil, dbInitErr
	}

	database.SetConnMaxIdleTime(time.Hour)
	database.SetMaxIdleConns(maxIdleConnections)
	database.SetMaxOpenConns(maxOpenConnections)

	if pingErr := database.Ping(); pingErr != nil {
		return nil, pingErr
	}

	return database, nil
}

// txKey is a private type to avoid collisions with other context keys.
type txKey struct{}

// WithTx returns a new context with the given transaction.
func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx returns the transaction from the context, or nil if not present.
func ExtractTx(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

func InTransaction(ctx context.Context, db *sqlx.DB, fn func(ctx context.Context) error) error {
	tx, txErr := db.BeginTxx(ctx, nil)
	if txErr != nil {
		return txErr
	}

	execErr := fn(WithTx(ctx, tx))
	if execErr != nil {
		return errors.Join(tx.Rollback(), execErr)
	}

	return tx.Commit()
}
