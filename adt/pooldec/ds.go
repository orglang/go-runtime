package pooldec

import (
	"orglang/orglang/lib/db"
)

// Port
type repo interface {
	Insert(db.Source, DecRec) error
}
