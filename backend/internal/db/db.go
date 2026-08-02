package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	connPool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return connPool, nil 
}