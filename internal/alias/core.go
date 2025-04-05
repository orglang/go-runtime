package alias

import (
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"
)

type Root struct {
	ID  id.ADT
	RN  rn.ADT
	Sym sym.ADT
}

type Repo interface {
	Insert(data.Source, Root) error
}
