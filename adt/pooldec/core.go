package pooldec

import (
	"log/slog"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/qualsym"
	"orglang/go-runtime/adt/termctx"
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
	PoolQN qualsym.ADT
	// Endpoints where pool acts as a provider for insiders
	InsiderProvisionBCs []termctx.BindClaim
	// Endpoints where pool acts as a client for insiders
	InsiderReceptionBCs []termctx.BindClaim
	// Endpoints where pool acts as a provider for outsiders
	OutsiderProvisionBCs []termctx.BindClaim
	// Endpoints where pool acts as a client for outsiders
	OutsiderReceptionBCs []termctx.BindClaim
}

type DecRef struct {
	DecID identity.ADT
}

type DecRec struct {
	DecID identity.ADT
}

type service struct {
	poolDecs repo
	operator db.Operator
	log      *slog.Logger
}

func (s *service) Create(spec DecSpec) (DecRef, error) {
	return DecRef{}, nil
}
