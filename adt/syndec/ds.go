package syndec

import (
	"orglang/orglang/lib/db"
)

type Repo interface {
	Insert(db.Source, DecRec) error
}

type decRecDS struct {
	DecID string
	DecRN int64
	DecQN string
}
