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
	ID         string                `db:"dec_id"`
	RN         int64                 `db:"dec_rn"`
	ClientBSs  []procbind.BindSpecDS `db:"ys"`
	ProviderBS procbind.BindSpecDS   `db:"x"`
}

type decSnapDS struct {
	ID         string                `db:"id"`
	RN         int64                 `db:"rn"`
	ClientBSs  []procbind.BindSpecDS `db:"ys"`
	ProviderBS procbind.BindSpecDS   `db:"x"`
}
