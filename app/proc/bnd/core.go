package bnd

import (
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/sym"
)

// aka ChanTp
type Spec struct {
	ChnlPH ph.ADT // may be blank
	RoleQN sym.ADT
}
