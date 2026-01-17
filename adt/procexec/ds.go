package procexec

import (
	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/procstep"
)

type Repo interface {
	SelectProc(db.Source, identity.ADT) (Cfg, error)
	UpdateProc(db.Source, Mod) error
	SelectMain(db.Source, identity.ADT) (MainCfg, error)
	UpdateMain(db.Source, MainMod) error
}

type modDS struct {
	Locks []lockDS
	Binds []bindDS
	Steps []procstep.StepRecDS
}

type lockDS struct {
	PoolID string
	PoolRN int64
}

type bindDS struct {
	ExecID  string
	ChnlPH  string
	ChnlID  string
	StateID string
	PoolRN  int64
}

type liabDS struct {
	PoolID string `db:"pool_id"`
	ProcID string `db:"proc_id"`
	PoolRN int64  `db:"rev"`
}

type epDS struct {
	ProcID   string  `db:"proc_id"`
	ChnlPH   string  `db:"chnl_ph"`
	ChnlID   string  `db:"chnl_id"`
	StateID  string  `db:"state_id"`
	PoolID   string  `db:"pool_id"`
	SrvID    string  `db:"srv_id"`
	SrvRevs  []int64 `db:"srv_revs"`
	ClntID   string  `db:"clnt_id"`
	ClntRevs []int64 `db:"clnt_revs"`
}
