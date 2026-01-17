package poolexp

import (
	"orglang/go-runtime/lib/db"
)

type Repo interface {
	InsertRec(db.Source, ExpSpec) error
}
