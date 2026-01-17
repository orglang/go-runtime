package procstep

import (
	"database/sql"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/procexp"
)

type Repo interface {
	InsertRecs(db.Source, ...StepRec) error
	SelectRecByID(db.Source, identity.ADT) (StepRec, error)
}

type StepRecDS struct {
	ID     string           `db:"id"`
	K      stepKindDS       `db:"kind"`
	PID    sql.NullString   `db:"pid"`
	VID    sql.NullString   `db:"vid"`
	ProcER procexp.ExpRecDS `db:"spec"`
}

type stepKindDS int

const (
	nonStep = stepKindDS(iota)
	msgStep
	svcStep
)
