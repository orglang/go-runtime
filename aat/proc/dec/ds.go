package dec

import (
	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
)

type Repo interface {
	Insert(sd.Source, ProcRec) error
	SelectAll(sd.Source) ([]ProcRef, error)
	SelectByID(sd.Source, id.ADT) (ProcSnap, error)
	SelectByIDs(sd.Source, []id.ADT) ([]ProcRec, error)
	SelectEnv(sd.Source, []id.ADT) (map[id.ADT]ProcRec, error)
}

type bndSpecDS struct {
	ChnlPH string `json:"chnl_ph"`
	TypeQN string `json:"role_qn"`
}

type sigRefDS struct {
	SigID string `db:"sig_id"`
	SigRN int64  `db:"rev"`
	Title string `db:"title"`
}

type sigRecDS struct {
	SigID string      `db:"sig_id"`
	Title string      `db:"title"`
	Ys    []bndSpecDS `db:"ys"`
	X     bndSpecDS   `db:"x"`
	SigRN int64       `db:"rev"`
}

type sigSnapDS struct {
	SigID string      `db:"sig_id"`
	Title string      `db:"title"`
	Ys    []bndSpecDS `db:"ys"`
	X     bndSpecDS   `db:"x"`
	SigRN int64       `db:"rev"`
}
