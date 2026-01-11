package poolexec

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procdec"
	"orglang/orglang/adt/procexec"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/revnum"
	"orglang/orglang/adt/typedef"
	"orglang/orglang/adt/typeexp"
)

// Port
type API interface {
	Run(ExecSpec) (ExecRef, error) // aka Create
	Retrieve(identity.ADT) (ExecSnap, error)
	RetreiveRefs() ([]ExecRef, error)
	Spawn(procexec.ExecSpec) (procexec.ExecRef, error)
	Poll(PollSpec) (procexec.ExecRef, error)
}

type ExecSpec struct {
	PoolQN qualsym.ADT
	SupID  identity.ADT
}

type ExecRef struct {
	ExecID identity.ADT
	ProcID identity.ADT // main
}

type ExecRec struct {
	ExecID identity.ADT
	ProcID identity.ADT // main
	SupID  identity.ADT
	ExecRN revnum.ADT
}

type ExecSnap struct {
	ExecID identity.ADT
	Title  string
	Subs   []ExecRef
}

type PollSpec struct {
	ExecID identity.ADT
}

type service struct {
	poolExecs Repo
	procDecs  procdec.Repo
	typeDefs  typedef.Repo
	typeExps  typeexp.Repo
	operator  db.Operator
	log       *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	poolExecs Repo,
	procDecs procdec.Repo,
	typeDefs typedef.Repo,
	typeExps typeexp.Repo,
	operator db.Operator,
	l *slog.Logger,
) *service {
	name := slog.String("name", reflect.TypeFor[service]().Name())
	return &service{poolExecs, procDecs, typeDefs, typeExps, operator, l.With(name)}
}

func (s *service) Run(spec ExecSpec) (ExecRef, error) {
	ctx := context.Background()
	s.log.Debug("creation started", slog.Any("spec", spec))
	impl := ExecRec{
		ExecID: identity.New(),
		ProcID: identity.New(),
		SupID:  spec.SupID,
		ExecRN: revnum.Initial(),
	}
	liab := procexec.Liab{
		PoolID: impl.ExecID,
		ProcID: impl.ProcID,
		PoolRN: impl.ExecRN,
	}
	err := s.operator.Explicit(ctx, func(ds db.Source) error {
		err := s.poolExecs.Insert(ds, impl)
		if err != nil {
			s.log.Error("creation failed")
			return err
		}
		err = s.poolExecs.InsertLiab(ds, liab)
		if err != nil {
			s.log.Error("creation failed")
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed")
		return ExecRef{}, err
	}
	s.log.Debug("creation succeed", slog.Any("poolID", impl.ExecID))
	return ConvertRecToRef(impl), nil
}

func (s *service) Poll(spec PollSpec) (procexec.ExecRef, error) {
	return procexec.ExecRef{}, nil
}

func (s *service) Spawn(spec procexec.ExecSpec) (_ procexec.ExecRef, err error) {
	procAttr := slog.Any("procID", spec.ExecID)
	s.log.Debug("spawning started", procAttr)
	return procexec.ExecRef{}, nil
}

func (s *service) Retrieve(execID identity.ADT) (snap ExecSnap, err error) {
	ctx := context.Background()
	err = s.operator.Implicit(ctx, func(ds db.Source) error {
		snap, err = s.poolExecs.SelectSubs(ds, execID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("execID", execID))
		return ExecSnap{}, err
	}
	return snap, nil
}

func (s *service) RetreiveRefs() (refs []ExecRef, err error) {
	ctx := context.Background()
	err = s.operator.Implicit(ctx, func(ds db.Source) error {
		refs, err = s.poolExecs.SelectRefs(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func errOptimisticUpdate(got revnum.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}
