package dec

import (
	"context"
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/alias"
)

// aka ChanTp
type ChnlSpec struct {
	ChnlPH sym.ADT // may be blank
	TypeQN sym.ADT
}

type SigSpec struct {
	X     ChnlSpec // via
	SigNS sym.ADT
	SigSN sym.ADT
	Ys    []ChnlSpec // vals
}

type SigRef struct {
	SigID id.ADT
	Title string
	SigRN rn.ADT
}

type SigRec struct {
	X     ChnlSpec
	SigID id.ADT
	Ys    []ChnlSpec
	Title string
	SigRN rn.ADT
}

// aka ExpDec or ExpDecDef without expression
type SigSnap struct {
	X     ChnlSpec
	SigID id.ADT
	Ys    []ChnlSpec
	Title string
	SigRN rn.ADT
}

type API interface {
	Incept(sym.ADT) (SigRef, error)
	Create(SigSpec) (SigSnap, error)
	Retrieve(id.ADT) (SigSnap, error)
	RetreiveRefs() ([]SigRef, error)
}

type service struct {
	procs    Repo
	aliases  alias.Repo
	operator data.Operator
	log      *slog.Logger
}

func newService(sigs Repo, aliases alias.Repo, operator data.Operator, l *slog.Logger) *service {
	return &service{sigs, aliases, operator, l}
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func (s *service) Incept(sigQN sym.ADT) (_ SigRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("sigQN", sigQN)
	s.log.Debug("inception started", qnAttr)
	newAlias := alias.Root{QN: sigQN, ID: id.New(), RN: rn.Initial()}
	newRec := SigRec{SigID: newAlias.ID, SigRN: newAlias.RN, Title: newAlias.QN.SN()}
	s.operator.Explicit(ctx, func(ds data.Source) error {
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
		return SigRef{}, err
	}
	s.log.Debug("inception succeeded", qnAttr, slog.Any("sigId", newRec.SigID))
	return ConvertRecToRef(newRec), nil
}

func (s *service) Create(spec SigSpec) (_ SigSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("sigQN", spec.SigSN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newRec := SigRec{
		X:     spec.X,
		SigID: id.New(),
		Ys:    spec.Ys,
		SigRN: rn.Initial(),
	}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.procs.Insert(ds, newRec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return SigSnap{}, err
	}
	s.log.Debug("creation succeeded", qnAttr, slog.Any("sigId", newRec.SigID))
	return ConvertRecToSnap(newRec), nil
}

func (s *service) Retrieve(sigID id.ADT) (snap SigSnap, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) error {
		snap, err = s.procs.SelectByID(ds, sigID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("sigId", sigID))
		return SigSnap{}, err
	}
	return snap, nil
}

func (s *service) RetreiveRefs() (refs []SigRef, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) error {
		refs, err = s.procs.SelectAll(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

type Repo interface {
	Insert(data.Source, SigRec) error
	SelectAll(data.Source) ([]SigRef, error)
	SelectByID(data.Source, id.ADT) (SigSnap, error)
	SelectByIDs(data.Source, []id.ADT) ([]SigRec, error)
	SelectEnv(data.Source, []id.ADT) (map[id.ADT]SigRec, error)
}

func CollectEnv(sigs []SigRec) []sym.ADT {
	typeQNs := []sym.ADT{}
	for _, sig := range sigs {
		typeQNs = append(typeQNs, sig.X.TypeQN)
		for _, y := range sig.Ys {
			typeQNs = append(typeQNs, y.TypeQN)
		}
	}
	return typeQNs
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	ConvertSnapToRef func(SigSnap) SigRef
	ConvertRecToRef  func(SigRec) SigRef
	ConvertRecToSnap func(SigRec) SigSnap
)

func ErrRootMissingInEnv(rid id.ADT) error {
	return fmt.Errorf("root missing in env: %v", rid)
}
