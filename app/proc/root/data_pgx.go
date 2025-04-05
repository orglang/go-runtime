package proc

import "log/slog"

// Adapter
type repoPgx struct {
	log *slog.Logger
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	return &repoPgx{l}
}

// for compilation purposes
func newRepo() Repo {
	return &repoPgx{}
}
