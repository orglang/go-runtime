package sig

import (
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/app/pool/bnd"
)

type Spec struct {
	SigQN   sym.ADT
	Imports []bnd.Spec
	Exports []bnd.Spec
}

type Impl struct {
	SigID id.ADT
}

type API interface {
	Create(Spec) (Impl, error)
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

func (s *service) Create(spec Spec) (Impl, error) {
	return Impl{}, nil
}

// Port
type Repo interface {
	Insert(data.Source, Impl) error
}
