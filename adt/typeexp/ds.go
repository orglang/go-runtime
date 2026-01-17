package typeexp

import (
	"database/sql"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
)

type Repo interface {
	InsertRec(db.Source, ExpRec) error
	SelectRecByID(db.Source, identity.ADT) (ExpRec, error)
	SelectRecsByIDs(db.Source, []identity.ADT) ([]ExpRec, error)
	SelectEnv(db.Source, []identity.ADT) (map[identity.ADT]ExpRec, error)
}

type expKindDS int

const (
	nonExp expKindDS = iota
	oneExp
	linkExp
	tensorExp
	lolliExp
	plusExp
	withExp
)

type ExpRefDS struct {
	ExpID string    `db:"exp_id" json:"exp_id"`
	K     expKindDS `db:"kind" json:"kind"`
}

type expRecDS struct {
	ExpID  string
	States []stateDS
}

type stateDS struct {
	ExpID  string         `db:"exp_id"`
	K      expKindDS      `db:"kind"`
	FromID sql.NullString `db:"from_id"`
	Spec   expSpecDS      `db:"spec"`
}

type expSpecDS struct {
	Link   string  `json:"link,omitempty"`
	Tensor *prodDS `json:"tensor,omitempty"`
	Lolli  *prodDS `json:"lolli,omitempty"`
	Plus   []sumDS `json:"plus,omitempty"`
	With   []sumDS `json:"with,omitempty"`
}

type prodDS struct {
	ValES  string `json:"on"`
	ContES string `json:"to"`
}

type sumDS struct {
	Lab    string `json:"on"`
	ContES string `json:"to"`
}
