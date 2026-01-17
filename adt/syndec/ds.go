package syndec

import (
	"orglang/go-runtime/lib/db"
)

type Repo interface {
	Insert(db.Source, DecRec) error
}

type decRecDS struct {
	DecID string
	DecRN int64
	DecQN string
}
