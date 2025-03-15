package proc

import (
	"fmt"

	"smecalculus/rolevod/app/role"
	"smecalculus/rolevod/app/sig"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/rev"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/chnl"
	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"
)

// aka Configuration
type Cfg struct {
	ProcID id.ADT
	Chnls  map[ph.ADT]Chnl
	Steps  map[chnl.ID]step.Root
	PoolID id.ADT
	Rev    rev.ADT
}

type Env struct {
	Sigs   map[sig.ID]sig.Root
	Roles  map[role.QN]role.Root
	States map[state.ID]state.Root
	Locks  map[sym.ADT]Lock
}

type Chnl struct {
	ChnlPH  ph.ADT
	ChnlID  id.ADT
	StateID id.ADT
	// provider
	PoolID id.ADT
	ProcID id.ADT
}

type EP struct {
	ChnlPH  ph.ADT
	ChnlID  id.ADT
	StateID id.ADT
}

type Lock struct {
	PoolID id.ADT
	Rev    rev.ADT
}

func ChnlPH(ch Chnl) ph.ADT { return ch.ChnlPH }

func ChnlID(ch Chnl) id.ADT { return ch.ChnlID }

// ответственность за процесс
type Liab struct {
	ProcID id.ADT
	PoolID id.ADT
	// позитивное значение при возникновении
	// негативное значение при лишении
	Rev rev.ADT
}

type Mod struct {
	Locks []Lock
	Bnds  []Bnd
	Steps []step.Root
	Liabs []Liab
}

type Bnd struct {
	ProcID  id.ADT
	ChnlPH  ph.ADT
	ChnlID  id.ADT
	StateID id.ADT
	Rev     rev.ADT
}

func ErrMissingChnl(want ph.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}
