package dec

import (
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type BndSpec struct {
	ChnlPH sym.ADT // may be blank
	TypeQN sym.ADT
}

type SigSpec struct {
	X     BndSpec // export
	SigNS sym.ADT
	SigSN sym.ADT   // label
	Ys    []BndSpec // imports
}

type SigRef struct {
	SigID id.ADT
}

type sigRec struct {
	SigID id.ADT
}

type API interface {
	Create(SigSpec) (SigRef, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	sigs     repo
	operator data.Operator
	log      *slog.Logger
}

func (s *service) Create(spec SigSpec) (SigRef, error) {
	return SigRef{}, nil
}

// Port
type repo interface {
	Insert(data.Source, sigRec) error
}
