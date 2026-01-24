package procexec

import (
	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/procbind"
	"orglang/go-runtime/adt/procstep"
	"orglang/go-runtime/adt/uniqref"
)

type Repo interface {
	SelectSnap(db.Source, ExecRef) (ExecSnap, error)
	UpdateProc(db.Source, ExecMod) error
}

type execModDS struct {
	Locks []execRefDS
	Binds []procbind.BindRecDS
	Steps []procstep.StepRecDS
}

type execRefDS = uniqref.Data

type liabDS struct {
	PoolID string `db:"pool_id"`
	ProcID string `db:"proc_id"`
	PoolRN int64  `db:"rev"`
}
