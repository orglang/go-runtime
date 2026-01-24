package procdec

import (
	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/procbind"
	"orglang/go-runtime/adt/uniqref"
)

type Repo interface {
	InsertRec(db.Source, DecRec) error
	SelectRefs(db.Source) ([]DecRef, error)
	SelectSnap(db.Source, DecRef) (DecSnap, error)
	SelectRecs(db.Source, []identity.ADT) ([]DecRec, error)
	SelectEnv(db.Source, []identity.ADT) (map[identity.ADT]DecRec, error)
}

type decRefDS = uniqref.Data

type decRecDS struct {
	DecID string                `db:"id"`
	Title string                `db:"title"`
	Ys    []procbind.BindSpecDS `db:"ys"`
	X     procbind.BindSpecDS   `db:"x"`
	DecRN int64                 `db:"rn"`
}

type decSnapDS struct {
	DecID string                `db:"id"`
	Title string                `db:"title"`
	Ys    []procbind.BindSpecDS `db:"ys"`
	X     procbind.BindSpecDS   `db:"x"`
	DecRN int64                 `db:"rn"`
}
