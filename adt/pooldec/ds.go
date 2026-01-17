package pooldec

import (
	"orglang/go-runtime/lib/db"
)

// Port
type repo interface {
	Insert(db.Source, DecRec) error
}
