package pooldec

import (
	"log/slog"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/termctx"
)

// Port
type API interface {
	Create(DecSpec) (DecRef, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type DecSpec struct {
	PoolNS qualsym.ADT
	PoolSN qualsym.ADT
	// endpoint where pool acts as a provider for insiders
	InsiderProvisionEP termctx.BindClaim
	// endpoint where pool acts as a client for insiders
	InsiderReceptionEP termctx.BindClaim
	// endpoint where pool acts as a provider for outsiders
	OutsiderProvisionEP termctx.BindClaim
	// endpoints where pool acts as a client for outsiders
	OutsiderReceptionEPs []termctx.BindClaim
}

type DecRef struct {
	DecID identity.ADT
}

type decRec struct {
	DecID identity.ADT
}

type service struct {
	sigs     repo
	operator db.Operator
	log      *slog.Logger
}

func (s *service) Create(spec DecSpec) (DecRef, error) {
	return DecRef{}, nil
}
