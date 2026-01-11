package procdec

import (
	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/termctx"
)

type Repo interface {
	Insert(db.Source, DecRec) error
	SelectAll(db.Source) ([]DecRef, error)
	SelectByID(db.Source, identity.ADT) (DecSnap, error)
	SelectByIDs(db.Source, []identity.ADT) ([]DecRec, error)
	SelectEnv(db.Source, []identity.ADT) (map[identity.ADT]DecRec, error)
}

type decRefDS struct {
	DecID string `db:"dec_id"`
	DecRN int64  `db:"dec_rn"`
}

type decRecDS struct {
	DecID string                `db:"dec_id"`
	Title string                `db:"title"`
	Ys    []termctx.BindClaimDS `db:"ys"`
	X     termctx.BindClaimDS   `db:"x"`
	DecRN int64                 `db:"dec_rn"`
}

type decSnapDS struct {
	DecID string                `db:"dec_id"`
	Title string                `db:"title"`
	Ys    []termctx.BindClaimDS `db:"ys"`
	X     termctx.BindClaimDS   `db:"x"`
	DecRN int64                 `db:"dec_rn"`
}
