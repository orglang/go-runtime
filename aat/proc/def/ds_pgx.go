package def

import (
	"log/slog"
)

// Adapter
type daoPgx struct {
	log *slog.Logger
}

func newDaoPgx(l *slog.Logger) *daoPgx {
	name := slog.String("name", "stepRepoPgx")
	return &daoPgx{l.With(name)}
}
