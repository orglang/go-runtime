package procdec

import (
	"context"
	"fmt"
	"iter"
	"log/slog"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/revnum"
	"orglang/orglang/adt/syndec"
	"orglang/orglang/adt/termctx"
)

type API interface {
	Incept(qualsym.ADT) (DecRef, error)
	Create(DecSpec) (DecSnap, error)
	RetrieveSnap(identity.ADT) (DecSnap, error)
	RetreiveRefs() ([]DecRef, error)
}

type DecSpec struct {
	ProcQN qualsym.ADT
	// endpoint where process acts as a provider
	ProvisionEP termctx.BindClaim
	// endpoints where process acts as a client
	ReceptionEPs []termctx.BindClaim
}

type DecRef struct {
	DecID identity.ADT
	DecRN revnum.ADT
}

type DecRec struct {
	X     termctx.BindClaim
	DecID identity.ADT
	Ys    []termctx.BindClaim
	Title string
	DecRN revnum.ADT
}

// aka ExpDec or ExpDecDef without expression
type DecSnap struct {
	X     termctx.BindClaim
	DecID identity.ADT
	Ys    []termctx.BindClaim
	Title string
	DecRN revnum.ADT
}

type service struct {
	procDecs Repo
	synDecs  syndec.Repo
	operator db.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(procs Repo, aliases syndec.Repo, operator db.Operator, l *slog.Logger) *service {
	return &service{procs, aliases, operator, l}
}

func (s *service) Incept(procQN qualsym.ADT) (_ DecRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("procQN", procQN)
	s.log.Debug("inception started", qnAttr)
	newSyn := syndec.DecRec{DecQN: procQN, DecID: identity.New(), DecRN: revnum.Initial()}
	newRec := DecRec{DecID: newSyn.DecID, DecRN: newSyn.DecRN, Title: newSyn.DecQN.SN()}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.synDecs.Insert(ds, newSyn)
		if err != nil {
			return err
		}
		err = s.procDecs.Insert(ds, newRec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", qnAttr)
		return DecRef{}, err
	}
	s.log.Debug("inception succeed", qnAttr, slog.Any("decID", newRec.DecID))
	return ConvertRecToRef(newRec), nil
}

func (s *service) Create(spec DecSpec) (_ DecSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("procQN", spec.ProcQN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newRec := DecRec{
		X:     spec.ProvisionEP,
		DecID: identity.New(),
		Ys:    spec.ReceptionEPs,
		DecRN: revnum.Initial(),
	}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.procDecs.Insert(ds, newRec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return DecSnap{}, err
	}
	s.log.Debug("creation succeed", qnAttr, slog.Any("decID", newRec.DecID))
	return ConvertRecToSnap(newRec), nil
}

func (s *service) RetrieveSnap(decID identity.ADT) (snap DecSnap, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds db.Source) error {
		snap, err = s.procDecs.SelectByID(ds, decID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("decID", decID))
		return DecSnap{}, err
	}
	return snap, nil
}

func (s *service) RetreiveRefs() (refs []DecRef, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds db.Source) error {
		refs, err = s.procDecs.SelectAll(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectEnv(recs iter.Seq[DecRec]) []qualsym.ADT {
	typeQNs := []qualsym.ADT{}
	for rec := range recs {
		typeQNs = append(typeQNs, rec.X.TypeQN)
		for _, y := range rec.Ys {
			typeQNs = append(typeQNs, y.TypeQN)
		}
	}
	return typeQNs
}

func ErrRootMissingInEnv(rid identity.ADT) error {
	return fmt.Errorf("root missing in env: %v", rid)
}
