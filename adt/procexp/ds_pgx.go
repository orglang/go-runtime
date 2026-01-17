package procexp

import (
	"log/slog"
	"reflect"

	"orglang/go-runtime/lib/db"
)

// Adapter
type pgxDAO struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo() Repo {
	return &pgxDAO{}
}

func newPgxDAO(l *slog.Logger) *pgxDAO {
	name := slog.String("name", reflect.TypeFor[pgxDAO]().Name())
	return &pgxDAO{l.With(name)}
}

func (dao *pgxDAO) Insert(source db.Source, rec ExpRec) error {
	return nil
}
