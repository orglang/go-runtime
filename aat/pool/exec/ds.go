package exec

import (
	"database/sql"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"

	procexec "orglang/orglang/aat/proc/exec"
)

// Port
type Repo interface {
	Insert(sd.Source, PoolRec) error
	InsertLiab(sd.Source, procexec.Liab) error
	SelectRefs(sd.Source) ([]PoolRef, error)
	SelectSubs(sd.Source, id.ADT) (PoolSnap, error)
	SelectProc(sd.Source, id.ADT) (procexec.Cfg, error)
	UpdateProc(sd.Source, procexec.Mod) error
}

type poolRefDS struct {
	PoolID string `json:"pool_id"`
	ProcID string `json:"proc_id"`
}

type poolSnapDS struct {
	PoolID string      `db:"pool_id"`
	Title  string      `db:"title"`
	Subs   []poolRefDS `db:"subs"`
}

type poolRecDS struct {
	PoolID string         `db:"pool_id"`
	ProcID string         `db:"proc_id"`
	SupID  sql.NullString `db:"sup_pool_id"`
	PoolRN int64          `db:"rev"`
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
