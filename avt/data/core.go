package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Source interface {
	source()
}

type SourcePgx struct {
	Ctx  context.Context
	Conn *pgx.Conn
}

func (SourcePgx) source() {}

type Operator interface {
	Explicit(context.Context, func(Source) error) error
	Implicit(context.Context, func(Source) error) error
}

type OperatorPgx struct {
	pool *pgxpool.Pool
}

func (o *OperatorPgx) Explicit(ctx context.Context, op func(Source) error) error {
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return err
	}
	err = op(SourcePgx{Ctx: ctx, Conn: tx.Conn()})
	if err != nil {
		return errors.Join(err, tx.Rollback(ctx))
	}
	return tx.Commit(ctx)
}

func (o *OperatorPgx) Implicit(ctx context.Context, op func(Source) error) error {
	conn, err := o.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return op(SourcePgx{Ctx: ctx, Conn: conn.Conn()})
}

func MustConform[T Source](got Source) T {
	ds, ok := got.(T)
	if !ok {
		var want T
		panic(fmt.Sprintf("must conform: want %T, got %T", want, got))
	}
	return ds
}
