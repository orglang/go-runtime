package typedef

import (
	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/uniqsym"
)

type Repo interface {
	Insert(db.Source, DefRec) error
	Update(db.Source, DefRec) error
	SelectRefs(db.Source) ([]DefRef, error)
	SelectRecByID(db.Source, identity.ADT) (DefRec, error)
	SelectRecsByIDs(db.Source, []identity.ADT) ([]DefRec, error)
	SelectRecByQN(db.Source, uniqsym.ADT) (DefRec, error)
	SelectRecsByQNs(db.Source, []uniqsym.ADT) ([]DefRec, error)
	SelectEnv(db.Source, []uniqsym.ADT) (map[uniqsym.ADT]DefRec, error)
}

type defRefDS struct {
	DefID string `db:"def_id"`
	DefRN int64  `db:"def_rn"`
}

type defRecDS struct {
	DefID string `db:"def_id"`
	Title string `db:"title"`
	ExpID string `db:"exp_id"`
	DefRN int64  `db:"def_rn"`
}
