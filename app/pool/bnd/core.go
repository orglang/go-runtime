package bnd

import (
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/sym"
)

type Spec struct {
	ProcPH ph.ADT // may be blank
	ProcQN sym.ADT
}

type Impl struct {
	ProcPH ph.ADT
	ProcID id.ADT
}
