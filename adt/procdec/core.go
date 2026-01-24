package procdec

import (
	"context"
	"fmt"
	"iter"
	"log/slog"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/procbind"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/syndec"
	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/adt/uniqsym"
)

type API interface {
	Incept(uniqsym.ADT) (DecRef, error)
	Create(DecSpec) (DecSnap, error)
	RetrieveSnap(DecRef) (DecSnap, error)
	RetreiveRefs() ([]DecRef, error)
}

type DecRef = uniqref.ADT

type DecSpec struct {
	ProcQN uniqsym.ADT
	// endpoint where process acts as a provider
	ProviderBS procbind.BindSpec
	// endpoints where process acts as a client
	ClientBSs []procbind.BindSpec
}

type DecRec struct {
	DecRef     DecRef
	ProviderBS procbind.BindSpec
	ClientBSs  []procbind.BindSpec
	Title      string
}

// aka ExpDec or ExpDecDef without expression
type DecSnap struct {
	DecRef     DecRef
	ProviderBS procbind.BindSpec
	ClientBSs  []procbind.BindSpec
	Title      string
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

func newService(procDecs Repo, synDecs syndec.Repo, operator db.Operator, log *slog.Logger) *service {
	return &service{procDecs, synDecs, operator, log}
}

func (s *service) Incept(procQN uniqsym.ADT) (_ DecRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("procQN", procQN)
	s.log.Debug("inception started", qnAttr)
	newSyn := syndec.DecRec{DecQN: procQN, DecID: identity.New(), DecRN: revnum.New()}
	newRec := DecRec{DecRef: DecRef{ID: newSyn.DecID, RN: newSyn.DecRN}, Title: symbol.ConvertToString(newSyn.DecQN.Sym())}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.synDecs.Insert(ds, newSyn)
		if err != nil {
			return err
		}
		err = s.procDecs.InsertRec(ds, newRec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", qnAttr)
		return DecRef{}, err
	}
	s.log.Debug("inception succeed", qnAttr, slog.Any("decRef", newRec.DecRef))
	return ConvertRecToRef(newRec), nil
}

func (s *service) Create(spec DecSpec) (_ DecSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("procQN", spec.ProcQN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newRec := DecRec{
		DecRef:     DecRef{ID: identity.New(), RN: revnum.New()},
		ProviderBS: spec.ProviderBS,
		ClientBSs:  spec.ClientBSs,
	}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.procDecs.InsertRec(ds, newRec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return DecSnap{}, err
	}
	s.log.Debug("creation succeed", qnAttr, slog.Any("decRef", newRec.DecRef))
	return ConvertRecToSnap(newRec), nil
}

func (s *service) RetrieveSnap(ref DecRef) (snap DecSnap, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds db.Source) error {
		snap, err = s.procDecs.SelectSnap(ds, ref)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("decRef", ref))
		return DecSnap{}, err
	}
	return snap, nil
}

func (s *service) RetreiveRefs() (refs []DecRef, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds db.Source) error {
		refs, err = s.procDecs.SelectRefs(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectEnv(recs iter.Seq[DecRec]) []uniqsym.ADT {
	typeQNs := []uniqsym.ADT{}
	for rec := range recs {
		typeQNs = append(typeQNs, rec.ProviderBS.TypeQN)
		for _, y := range rec.ClientBSs {
			typeQNs = append(typeQNs, y.TypeQN)
		}
	}
	return typeQNs
}

func ErrRootMissingInEnv(rid identity.ADT) error {
	return fmt.Errorf("root missing in env: %v", rid)
}
