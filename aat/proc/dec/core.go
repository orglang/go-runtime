package dec

import (
	"context"
	"fmt"
	"log/slog"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/rn"
	"orglang/orglang/avt/sym"

	"orglang/orglang/aet/alias"
)

type API interface {
	Incept(sym.ADT) (ProcRef, error)
	Create(ProcSpec) (ProcSnap, error)
	Retrieve(id.ADT) (ProcSnap, error)
	RetreiveRefs() ([]ProcRef, error)
}

type ProcSpec struct {
	ProcNS sym.ADT
	ProcSN sym.ADT
	// endpoint where process acts as a provider
	ProvisionEP ChnlSpec
	// endpoints where process acts as a client
	ReceptionEPs []ChnlSpec
}

// channel endpoint
type ChnlSpec struct {
	CommPH sym.ADT // may be blank
	TypeQN sym.ADT
}

type ProcRef struct {
	DecID id.ADT
	Title string
	DecRN rn.ADT
}

type ProcRec struct {
	X     ChnlSpec
	DecID id.ADT
	Ys    []ChnlSpec
	Title string
	DecRN rn.ADT
}

// aka ExpDec or ExpDecDef without expression
type ProcSnap struct {
	X     ChnlSpec
	DecID id.ADT
	Ys    []ChnlSpec
	Title string
	DecRN rn.ADT
}

type service struct {
	procs    Repo
	aliases  alias.Repo
	operator sd.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(procs Repo, aliases alias.Repo, operator sd.Operator, l *slog.Logger) *service {
	return &service{procs, aliases, operator, l}
}

func (s *service) Incept(procQN sym.ADT) (_ ProcRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("procQN", procQN)
	s.log.Debug("inception started", qnAttr)
	newAlias := alias.Root{QN: procQN, ID: id.New(), RN: rn.Initial()}
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
		DecID: id.New(),
		Ys:    spec.ReceptionEPs,
		DecRN: rn.Initial(),
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

func (s *service) Retrieve(sigID id.ADT) (snap ProcSnap, err error) {
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

func CollectEnv(recs []ProcRec) []sym.ADT {
	typeQNs := []sym.ADT{}
	for _, rec := range recs {
		typeQNs = append(typeQNs, rec.X.TypeQN)
		for _, y := range rec.Ys {
			typeQNs = append(typeQNs, y.TypeQN)
		}
	}
	return typeQNs
}

func ErrRootMissingInEnv(rid id.ADT) error {
	return fmt.Errorf("root missing in env: %v", rid)
}
