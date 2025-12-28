package dec

import (
	"log/slog"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/sym"
)

// Port
type API interface {
	Create(PoolSpec) (PoolRef, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type PoolSpec struct {
	PoolNS sym.ADT
	PoolSN sym.ADT
	// endpoint where pool acts as a provider for insiders
	InsiderProvisionEP ChnlSpec
	// endpoint where pool acts as a client for insiders
	InsiderReceptionEP ChnlSpec
	// endpoint where pool acts as a provider for outsiders
	OutsiderProvisionEP ChnlSpec
	// endpoints where pool acts as a client for outsiders
	OutsiderReceptionEPs []ChnlSpec
}

type ChnlSpec struct {
	CommPH sym.ADT // may be blank
	TypeQN sym.ADT
}

type PoolRef struct {
	DecID id.ADT
}

type poolRec struct {
	DecID id.ADT
}

type service struct {
	sigs     repo
	operator sd.Operator
	log      *slog.Logger
}

func (s *service) Create(spec PoolSpec) (PoolRef, error) {
	return PoolRef{}, nil
}
