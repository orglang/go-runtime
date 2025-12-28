package sd

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func newOperator(pool *pgxpool.Pool) Operator {
	return &OperatorPgx{pool}
}

func newDriverPgx(pc storagePC, lc fx.Lifecycle) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(pc.Protocol.Postgres.Url)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 2
	pgx, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	lc.Append(
		fx.Hook{
			OnStart: pgx.Ping,
			OnStop: func(ctx context.Context) error {
				go pgx.Close()
				return nil
			},
		},
	)
	return pgx, nil
}
