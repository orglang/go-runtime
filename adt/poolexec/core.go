package poolexec

import (
	"context"
	"log/slog"
	"reflect"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/poolstep"
	"orglang/go-runtime/adt/procdec"
	"orglang/go-runtime/adt/procexec"
	"orglang/go-runtime/adt/typedef"
	"orglang/go-runtime/adt/typeexp"
	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/adt/uniqsym"
)

// Port
type API interface {
	Run(ExecSpec) (ExecRef, error) // aka Create
	RetrieveSnap(ExecRef) (ExecSnap, error)
	RetreiveRefs() ([]ExecRef, error)
	Take(poolstep.StepSpec) error
	Poll(PollSpec) (procexec.ExecRef, error)
}

type ExecSpec struct {
	PoolQN uniqsym.ADT
	SupID  identity.ADT
}

type ExecRef = uniqref.ADT

type ExecRec struct {
	ExecRef ExecRef
	SupID   identity.ADT
}

type ExecSnap struct {
	ExecRef  ExecRef
	Title    string
	SubExecs []ExecRef
}

type PollSpec struct {
	ExecID identity.ADT
}

// ответственность за процесс
type Liab struct {
	// позитивное значение при вручении
	// негативное значение при лишении
	ExecRef ExecRef
	ProcID  identity.ADT
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
	log *slog.Logger,
) *service {
	name := slog.String("name", reflect.TypeFor[service]().Name())
	return &service{poolExecs, procDecs, typeDefs, typeExps, operator, log.With(name)}
}

func (s *service) Run(spec ExecSpec) (ExecRef, error) {
	ctx := context.Background()
	s.log.Debug("creation started", slog.Any("spec", spec))
	execRec := ExecRec{
		ExecRef: uniqref.New(),
		SupID:   spec.SupID,
	}
	err := s.operator.Explicit(ctx, func(ds db.Source) error {
		err := s.poolExecs.InsertRec(ds, execRec)
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
	s.log.Debug("creation succeed", slog.Any("execRef", execRec.ExecRef))
	return execRec.ExecRef, nil
}

func (s *service) Poll(spec PollSpec) (procexec.ExecRef, error) {
	return procexec.ExecRef{}, nil
}

func (s *service) Take(spec poolstep.StepSpec) (err error) {
	qnAttr := slog.Any("procQN", spec.ProcQN)
	s.log.Debug("spawning started", qnAttr)
	return nil
}

func (s *service) RetrieveSnap(ref ExecRef) (snap ExecSnap, err error) {
	ctx := context.Background()
	err = s.operator.Implicit(ctx, func(ds db.Source) error {
		snap, err = s.poolExecs.SelectSubs(ds, ref)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("execID", ref))
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
