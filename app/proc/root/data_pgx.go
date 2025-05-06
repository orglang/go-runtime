package root

import (
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
)

// Adapter
type repoPgx struct {
	log *slog.Logger
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	return &repoPgx{l}
}

// for compilation purposes
func newRepo() repo {
	return &repoPgx{}
}

func (r *repoPgx) SelectMain(data.Source, id.ADT) (MainCfg, error) {
	return MainCfg{}, nil
}

func (r *repoPgx) UpdateMain(data.Source, MainMod) error {
	return nil
}
