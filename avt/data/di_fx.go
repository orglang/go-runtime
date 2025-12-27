package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"orglang/orglang/avt/core"
)

var Module = fx.Module("lib/data",
	fx.Provide(
		newPgx,
		fx.Annotate(newOperator, fx.As(new(Operator))),
	),
	fx.Provide(
		fx.Private,
		newCfg,
	),
)

func newOperator(pool *pgxpool.Pool) Operator {
	return &OperatorPgx{pool}
}

func newCfg(k core.Keeper) (*props, error) {
	props := &props{}
	err := k.Load("storage", props)
	if err != nil {
		return nil, err
	}
	return props, nil
}

func newPgx(p *props, lc fx.Lifecycle) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(p.Protocol.Postgres.Url)
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
