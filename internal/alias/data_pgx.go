package alias

import (
	"log/slog"
	"math"
	"smecalculus/rolevod/lib/data"

	"github.com/jackc/pgx/v5"
)

// Adapter
type repoPgx struct {
	log *slog.Logger
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	name := slog.String("name", "aliasRepoPgx")
	return &repoPgx{l.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &repoPgx{}
}

func (r *repoPgx) Insert(source data.Source, root Root) error {
	ds := data.MustConform[data.SourcePgx](source)
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
