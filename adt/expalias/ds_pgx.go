package expalias

import (
	"log/slog"
	"math"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/sd"
)

// Adapter
type daoPgx struct {
	log *slog.Logger
}

func newDaoPgx(l *slog.Logger) *daoPgx {
	name := slog.String("name", reflect.TypeFor[daoPgx]().Name())
	return &daoPgx{l.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &daoPgx{}
}

func (r *daoPgx) Insert(source sd.Source, root Root) error {
	ds := sd.MustConform[sd.SourcePgx](source)
	idAttr := slog.Any("id", root.ID)
	dto, err := DataFromRoot(root)
	if err != nil {
		r.log.Error("model mapping failed", idAttr)
		return err
	}
	query := `
		insert into aliases (
			id, rev_from, rev_to, sym
		) values (
			@id, @rev_from, @rev_to, @sym
		)`
	args := pgx.NamedArgs{
		"id":       dto.ID,
		"rev_from": dto.RN,
		"rev_to":   math.MaxInt64,
		"sym":      dto.Sym,
	}
	_, err = ds.Conn.Exec(ds.Ctx, query, args)
	if err != nil {
		r.log.Error("query execution failed", idAttr, slog.String("q", query))
		return err
	}
	return nil
}
