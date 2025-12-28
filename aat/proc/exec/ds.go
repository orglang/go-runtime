package exec

import (
	"database/sql"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"

	procdef "orglang/orglang/aat/proc/def"
)

type repo interface {
	SelectMain(sd.Source, id.ADT) (MainCfg, error)
	UpdateMain(sd.Source, MainMod) error
}

type SemRepo interface {
	InsertSem(sd.Source, ...SemRec) error
	SelectSemByID(sd.Source, id.ADT) (SemRec, error)
}

type modDS struct {
	Locks []lockDS
	Bnds  []bndDS
	Steps []SemRecDS
}

type lockDS struct {
	PoolID string
	PoolRN int64
}

type bndDS struct {
	ProcID  string
	ChnlPH  string
	ChnlID  string
	StateID string
	PoolRN  int64
}

type SemRecDS struct {
	ID  string            `db:"id"`
	K   semKind           `db:"kind"`
	PID sql.NullString    `db:"pid"`
	VID sql.NullString    `db:"vid"`
	TR  procdef.TermRecDS `db:"spec"`
}

type semKind int

const (
	nonsem = semKind(iota)
	msgKind
	svcKind
)
