package exec

import (
	"database/sql"

	procexec "orglang/orglang/aat/proc/exec"
)

type poolRefData struct {
	PoolID string `json:"pool_id"`
	ProcID string `json:"proc_id"`
}

type poolSnapData struct {
	PoolID string        `db:"pool_id"`
	Title  string        `db:"title"`
	Subs   []poolRefData `db:"subs"`
}

type poolRecData struct {
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
// goverter:extend orglang/orglang/avt/id:Convert.*
var (
	DataToPoolRef    func(poolRefData) (PoolRef, error)
	DataFromPoolRef  func(PoolRef) poolRefData
	DataToPoolRefs   func([]poolRefData) ([]PoolRef, error)
	DataFromPoolRefs func([]PoolRef) []poolRefData
	DataToPoolRec    func(poolRecData) (PoolRec, error)
	DataFromPoolRec  func(PoolRec) poolRecData
	DataToLiab       func(liabData) (procexec.Liab, error)
	DataFromLiab     func(procexec.Liab) liabData
	DataToPoolSnap   func(poolSnapData) (PoolSnap, error)
	DataFromPoolSnap func(PoolSnap) poolSnapData
	DataToEPs        func([]epData) ([]procexec.EP, error)
)
