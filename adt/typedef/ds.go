package typedef

import (
	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
)

type Repo interface {
	Insert(db.Source, DefRec) error
	Update(db.Source, DefRec) error
	SelectRefs(db.Source) ([]DefRef, error)
	SelectRecByID(db.Source, identity.ADT) (DefRec, error)
	SelectRecsByIDs(db.Source, []identity.ADT) ([]DefRec, error)
	SelectRecByQN(db.Source, qualsym.ADT) (DefRec, error)
	SelectRecsByQNs(db.Source, []qualsym.ADT) ([]DefRec, error)
	SelectEnv(db.Source, []qualsym.ADT) (map[qualsym.ADT]DefRec, error)
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
