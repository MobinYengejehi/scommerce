package dbsamples

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserAccountID = uint64

type PostgreDatabase struct {
	_         sync.Mutex
	PgxPool   *pgxpool.Pool
	TxOptions pgx.TxOptions
}

func NewPostgreDatabase(ctx context.Context, config *pgxpool.Config) (*PostgreDatabase, error) {
	pgxPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return &PostgreDatabase{
		PgxPool: pgxPool,
		TxOptions: pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		},
	}, nil
}

func (db *PostgreDatabase) Ping(ctx context.Context) error {
	return db.PgxPool.Ping(ctx)
}

func (db *PostgreDatabase) Close() error {
	db.PgxPool.Close()
	return nil
}
