package def

import (
	"database/sql"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
)

type Repo interface {
	InsertType(sd.Source, TypeRec) error
	UpdateType(sd.Source, TypeRec) error
	SelectTypeRefs(sd.Source) ([]TypeRef, error)
	SelectTypeRecByID(sd.Source, id.ADT) (TypeRec, error)
	SelectTypeRecsByIDs(sd.Source, []id.ADT) ([]TypeRec, error)
	SelectTypeRecByQN(sd.Source, sym.ADT) (TypeRec, error)
	SelectTypeRecsByQNs(sd.Source, []sym.ADT) ([]TypeRec, error)
	SelectTypeEnv(sd.Source, []sym.ADT) (map[sym.ADT]TypeRec, error)

	InsertTerm(sd.Source, TermRec) error
	SelectTermRecByID(sd.Source, id.ADT) (TermRec, error)
	SelectTermRecsByIDs(sd.Source, []id.ADT) ([]TermRec, error)
	SelectTermEnv(sd.Source, []id.ADT) (map[id.ADT]TermRec, error)
}

type typeRefDS struct {
	TypeID string `db:"role_id"`
	TypeRN int64  `db:"rev"`
	Title  string `db:"title"`
}

type typeRecDS struct {
	TypeID string `db:"role_id"`
	Title  string `db:"title"`
	TermID string `db:"state_id"`
	TypeRN int64  `db:"rev"`
}

type termKind int

const (
	nonterm termKind = iota
	oneKind
	linkKind
	tensorKind
	lolliKind
	plusKind
	withKind
)

type TermRefDS struct {
	ID string   `db:"id" json:"id"`
	K  termKind `db:"kind" json:"kind"`
}

type termRecDS struct {
	ID     string
	States []stateDS
}

type stateDS struct {
	ID     string         `db:"id"`
	K      termKind       `db:"kind"`
	FromID sql.NullString `db:"from_id"`
	Spec   specDS         `db:"spec"`
}

type specDS struct {
	Link   string  `json:"link,omitempty"`
	Tensor *prodDS `json:"tensor,omitempty"`
	Lolli  *prodDS `json:"lolli,omitempty"`
	Plus   []sumDS `json:"plus,omitempty"`
	With   []sumDS `json:"with,omitempty"`
}

type prodDS struct {
	Val  string `json:"on"`
	Cont string `json:"to"`
}

type sumDS struct {
	Lab  string `json:"on"`
	Cont string `json:"to"`
}
