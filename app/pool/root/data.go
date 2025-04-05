package pool

import (
	"database/sql"

	procroot "smecalculus/rolevod/app/proc/root"
)

type refData struct {
	PoolID string `json:"pool_id"`
	ProcID string `json:"proc_id"`
}

type snapData struct {
	PoolID string    `db:"pool_id"`
	Title  string    `db:"title"`
	Subs   []refData `db:"subs"`
}

type implData struct {
	PoolID string         `db:"pool_id"`
	ProcID string         `db:"proc_id"`
	SupID  sql.NullString `db:"sup_pool_id"`
	PoolRN int64          `db:"rev"`
}

type liabData struct {
	PoolID string `db:"pool_id"`
	ProcID string `db:"proc_id"`
	PoolRN int64  `db:"rev"`
}

type epData struct {
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

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	DataToRef       func(refData) (Ref, error)
	DataFromRef     func(Ref) refData
	DataToRefs      func([]refData) ([]Ref, error)
	DataFromRefs    func([]Ref) []refData
	DataToRoot      func(implData) (impl, error)
	DataFromRoot    func(impl) implData
	DataToLiab      func(liabData) (procroot.Liab, error)
	DataFromLiab    func(procroot.Liab) liabData
	DataToSubSnap   func(snapData) (Snap, error)
	DataFromSubSnap func(Snap) snapData
	DataToEPs       func([]epData) ([]procroot.EP, error)
)
