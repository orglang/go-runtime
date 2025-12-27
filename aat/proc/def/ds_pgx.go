package def

import (
	"log/slog"
)

// Adapter
type repoPgx struct {
	log *slog.Logger
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	name := slog.String("name", "stepRepoPgx")
	return &repoPgx{l.With(name)}
}
