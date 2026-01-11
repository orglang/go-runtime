package syndec

import (
	"log/slog"
	"math"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/db"
)

type pgxDAO struct {
	log *slog.Logger
}

func newPgxDAO(l *slog.Logger) *pgxDAO {
	name := slog.String("name", reflect.TypeFor[pgxDAO]().Name())
	return &pgxDAO{l.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &pgxDAO{}
}

func (dao *pgxDAO) Insert(source db.Source, root DecRec) error {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("id", root.DecID)
	dto, err := DataFromDecRec(root)
	if err != nil {
		dao.log.Error("model conversion failed", idAttr)
		return err
	}
	query := `
		insert into aliases (
			id, from_rn, to_rn, sym
		) values (
			@id, @from_rn, @to_rn, @sym
		)`
	args := pgx.NamedArgs{
		"id":      dto.DecID,
		"from_rn": dto.DecRN,
		"to_rn":   math.MaxInt64,
		"sym":     dto.DecQN,
	}
	_, err = ds.Conn.Exec(ds.Ctx, query, args)
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", query))
		return err
	}
	return nil
}
