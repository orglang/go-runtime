package poolexp

import (
	"orglang/orglang/lib/db"
)

type Repo interface {
	InsertRec(db.Source, ExpSpec) error
}
