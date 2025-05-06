package sig

import (
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/app/pool/bnd"
)

type Spec struct {
	X     bnd.Spec // export
	SigNS sym.ADT
	SigSN sym.ADT    // label
	Ys    []bnd.Spec // imports
}

type Ref struct {
	SigID id.ADT
}

type impl struct {
	SigID id.ADT
}

type API interface {
	Create(Spec) (Ref, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	sigs     Repo
	operator data.Operator
	log      *slog.Logger
}

func (s *service) Create(spec Spec) (Ref, error) {
	return Ref{}, nil
}

// Port
type Repo interface {
	Insert(data.Source, impl) error
}
