package proc

import (
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	procsig "smecalculus/rolevod/app/proc/sig"
	roleroot "smecalculus/rolevod/app/role/root"
)

type Spec struct {
	PoolID id.ADT
	ProcID id.ADT
	Term   step.CallSpec
}

type Ref struct {
	ProcID id.ADT
}

type Snap struct {
	ProcID id.ADT
}

// aka Configuration
type Cfg struct {
	ProcID id.ADT
	Chnls  map[ph.ADT]EP
	Steps  map[id.ADT]step.Root
	PoolID id.ADT
	PoolRN rn.ADT // ProcRN?
}

type Env struct {
	Sigs   map[id.ADT]procsig.Impl
	Roles  map[sym.ADT]roleroot.Impl
	States map[state.ID]state.Root
	Locks  map[sym.ADT]Lock
}

type EP struct {
	ChnlPH  ph.ADT
	ChnlID  id.ADT
	StateID id.ADT
	// provider
	PoolID id.ADT
}

type Lock struct {
	PoolID id.ADT
	PoolRN rn.ADT
}

func ChnlPH(ch EP) ph.ADT { return ch.ChnlPH }

func ChnlID(ch EP) id.ADT { return ch.ChnlID }

// ответственность за процесс
type Liab struct {
	PoolID id.ADT
	ProcID id.ADT
	// позитивное значение при вручении
	// негативное значение при лишении
	PoolRN rn.ADT
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
	PoolRN  rn.ADT
}

type API interface {
	Create(Spec) (Ref, error)
	Retrieve(id.ADT) (Snap, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	procs    Repo
	operator data.Operator
	log      *slog.Logger
}

func newService(
	procs Repo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	return &service{procs, operator, l}
}

func (s *service) Create(spec Spec) (_ Ref, err error) {
	idAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("creation started", idAttr)
	return Ref{}, nil
}

func (s *service) Retrieve(procID id.ADT) (_ Snap, err error) {
	return Snap{}, nil
}

type Repo interface {
}

func ErrMissingChnl(want ph.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}
