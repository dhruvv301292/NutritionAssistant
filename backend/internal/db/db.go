package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool connects via Supabase's transaction-mode pooler (pgbouncer),
// which recycles physical connections across clients mid-session — pgx's
// default prepared-statement cache assumes a stable connection and errors
// with "prepared statement already exists" under that recycling. The simple
// protocol avoids server-side prepared statements entirely.
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	connPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return connPool, nil
}