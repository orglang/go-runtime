package poolexec

import (
	"database/sql"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/procexec"
)

// Port
type Repo interface {
	InsertRec(db.Source, ExecRec) error
	InsertLiab(db.Source, procexec.Liab) error
	SelectRefs(db.Source) ([]ExecRef, error)
	SelectSubs(db.Source, ExecRef) (ExecSnap, error)
}

type execRefDS struct {
	PoolID string `json:"pool_id"`
	ProcID string `json:"proc_id"`
}

type execSnapDS struct {
	ExecID string      `db:"pool_id"`
	Title  string      `db:"title"`
	Subs   []execRefDS `db:"subs"`
}

type execRecDS struct {
	ExecID string         `db:"pool_id"`
	ProcID string         `db:"proc_id"`
	SupID  sql.NullString `db:"sup_pool_id"`
	ExecRN int64          `db:"rev"`
}

type liabDS struct {
	ExecID string `db:"pool_id"`
	ProcID string `db:"proc_id"`
	ExecRN int64  `db:"rev"`
}
