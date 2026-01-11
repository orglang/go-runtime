package procdef

import (
	"fmt"
	"log/slog"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procexp"
	"orglang/orglang/adt/qualsym"
)

type API interface {
	Create(DefSpec) (DefRef, error)
	Retrieve(identity.ADT) (DefRec, error)
}

type DefSpec struct {
	ProcQN qualsym.ADT // or dec.ProcID
	ProcES procexp.ExpSpec
}

type DefRef struct {
	DefID identity.ADT
}

type DefRec struct {
	DefID identity.ADT
}

type DefSnap struct {
	DefID identity.ADT
}

type service struct {
	procs    Repo
	operator db.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	procs Repo,
	operator db.Operator,
	l *slog.Logger,
) *service {
	return &service{procs, operator, l}
}

func (s *service) Create(spec DefSpec) (DefRef, error) {
	return DefRef{}, nil
}

func (s *service) Retrieve(recID identity.ADT) (DefRec, error) {
	return DefRec{}, nil
}

func ErrDoesNotExist(want identity.ADT) error {
	return fmt.Errorf("rec doesn't exist: %v", want)
}

func ErrMissingInCfg(want qualsym.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCfg2(want identity.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCtx(want qualsym.ADT) error {
	return fmt.Errorf("channel missing in ctx: %v", want)
}
