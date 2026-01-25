package poolexec

import (
	"database/sql"

	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/lib/db"
)

// Port
type Repo interface {
	InsertRec(db.Source, ExecRec) error
	InsertLiab(db.Source, Liab) error
	SelectRefs(db.Source) ([]ExecRef, error)
	SelectSubs(db.Source, ExecRef) (ExecSnap, error)
}

type execRefDS = uniqref.Data

type execSnapDS struct {
	ID       string      `db:"exec_id"`
	RN       int64       `db:"exec_rn"`
	Title    string      `db:"title"`
	SubExecs []execRefDS `db:"subs"`
}

type execRecDS struct {
	ID    string         `db:"exec_id"`
	RN    int64          `db:"exec_rn"`
	SupID sql.NullString `db:"sup_exec_id"`
}

type liabDS struct {
	ID     string `db:"exec_id"`
	RN     int64  `db:"exec_rn"`
	ProcID string `db:"proc_id"`
}
