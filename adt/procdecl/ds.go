package procdecl

import (
	"orglang/orglang/lib/sd"

	"orglang/orglang/adt/expctx"
	"orglang/orglang/adt/identity"
)

type Repo interface {
	Insert(sd.Source, ProcRec) error
	SelectAll(sd.Source) ([]ProcRef, error)
	SelectByID(sd.Source, identity.ADT) (ProcSnap, error)
	SelectByIDs(sd.Source, []identity.ADT) ([]ProcRec, error)
	SelectEnv(sd.Source, []identity.ADT) (map[identity.ADT]ProcRec, error)
}

type sigRefDS struct {
	SigID string `db:"sig_id"`
	SigRN int64  `db:"rev"`
	Title string `db:"title"`
}

type sigRecDS struct {
	SigID string               `db:"sig_id"`
	Title string               `db:"title"`
	Ys    []expctx.BindClaimDS `db:"ys"`
	X     expctx.BindClaimDS   `db:"x"`
	SigRN int64                `db:"rev"`
}

type sigSnapDS struct {
	SigID string               `db:"sig_id"`
	Title string               `db:"title"`
	Ys    []expctx.BindClaimDS `db:"ys"`
	X     expctx.BindClaimDS   `db:"x"`
	SigRN int64                `db:"rev"`
}
