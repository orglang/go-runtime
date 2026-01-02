package procdecl

import (
	"context"
	"fmt"
	"log/slog"

	"orglang/orglang/lib/sd"

	"orglang/orglang/adt/expalias"
	"orglang/orglang/adt/expctx"
	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/revnum"
)

type API interface {
	Incept(qualsym.ADT) (ProcRef, error)
	Create(ProcSpec) (ProcSnap, error)
	Retrieve(identity.ADT) (ProcSnap, error)
	RetreiveRefs() ([]ProcRef, error)
}

type ProcSpec struct {
	ProcNS qualsym.ADT
	ProcSN qualsym.ADT
	// endpoint where process acts as a provider
	ProvisionEP expctx.BindClaim
	// endpoints where process acts as a client
	ReceptionEPs []expctx.BindClaim
}

type ProcRef struct {
	DecID identity.ADT
	Title string
	DecRN revnum.ADT
}

type ProcRec struct {
	X     expctx.BindClaim
	DecID identity.ADT
	Ys    []expctx.BindClaim
	Title string
	DecRN revnum.ADT
}

// aka ExpDec or ExpDecDef without expression
type ProcSnap struct {
	X     expctx.BindClaim
	DecID identity.ADT
	Ys    []expctx.BindClaim
	Title string
	DecRN revnum.ADT
}

type service struct {
	procs    Repo
	aliases  expalias.Repo
	operator sd.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(procs Repo, aliases expalias.Repo, operator sd.Operator, l *slog.Logger) *service {
	return &service{procs, aliases, operator, l}
}

func (s *service) Incept(procQN qualsym.ADT) (_ ProcRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("procQN", procQN)
	s.log.Debug("inception started", qnAttr)
	newAlias := expalias.Root{QN: procQN, ID: identity.New(), RN: revnum.Initial()}
	newRec := ProcRec{DecID: newAlias.ID, DecRN: newAlias.RN, Title: newAlias.QN.SN()}
	s.operator.Explicit(ctx, func(ds sd.Source) error {
		err = s.aliases.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.procs.Insert(ds, newRec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", qnAttr)
		return ProcRef{}, err
	}
	s.log.Debug("inception succeed", qnAttr, slog.Any("sigID", newRec.DecID))
	return ConvertRecToRef(newRec), nil
}

func (s *service) Create(spec ProcSpec) (_ ProcSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("sigQN", spec.ProcSN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newRec := ProcRec{
		X:     spec.ProvisionEP,
		DecID: identity.New(),
		Ys:    spec.ReceptionEPs,
		DecRN: revnum.Initial(),
	}
	s.operator.Explicit(ctx, func(ds sd.Source) error {
		err = s.procs.Insert(ds, newRec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return ProcSnap{}, err
	}
	s.log.Debug("creation succeed", qnAttr, slog.Any("sigID", newRec.DecID))
	return ConvertRecToSnap(newRec), nil
}

func (s *service) Retrieve(sigID identity.ADT) (snap ProcSnap, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds sd.Source) error {
		snap, err = s.procs.SelectByID(ds, sigID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("sigID", sigID))
		return ProcSnap{}, err
	}
	return snap, nil
}

func (s *service) RetreiveRefs() (refs []ProcRef, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds sd.Source) error {
		refs, err = s.procs.SelectAll(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectEnv(recs []ProcRec) []qualsym.ADT {
	typeQNs := []qualsym.ADT{}
	for _, rec := range recs {
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
