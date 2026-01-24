package typedef

import (
	"context"
	"fmt"
	"iter"
	"log/slog"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/syndec"
	"orglang/go-runtime/adt/typeexp"
	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/adt/uniqsym"
)

type API interface {
	Incept(uniqsym.ADT) (DefRef, error)
	Create(DefSpec) (DefSnap, error)
	Modify(DefSnap) (DefSnap, error)
	RetrieveSnap(DefRef) (DefSnap, error)
	retrieveSnap(DefRec) (DefSnap, error)
	RetreiveRefs() ([]DefRef, error)
}

type DefRef = uniqref.ADT

type DefSpec struct {
	TypeQN uniqsym.ADT
	TypeES typeexp.ExpSpec
}

// aka TpDef
type DefRec struct {
	DefRef DefRef
	Title  string
	ExpID  identity.ADT
}

type DefSnap struct {
	DefRef DefRef
	Title  string
	TypeQN uniqsym.ADT
	TypeES typeexp.ExpSpec
}

type Context struct {
	Assets map[symbol.ADT]typeexp.ExpRec
	Liabs  map[symbol.ADT]typeexp.ExpRec
}

type service struct {
	typeDefs Repo
	typeExps typeexp.Repo
	synDecs  syndec.Repo
	operator db.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	typeDefs Repo,
	typeExps typeexp.Repo,
	synDecs syndec.Repo,
	operator db.Operator,
	l *slog.Logger,
) *service {
	return &service{typeDefs, typeExps, synDecs, operator, l}
}

func (s *service) Incept(typeQN uniqsym.ADT) (_ DefRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("typeQN", typeQN)
	s.log.Debug("inception started", qnAttr)
	newSyn := syndec.DecRec{DecQN: typeQN, DecID: identity.New(), DecRN: revnum.New()}
	newType := DefRec{DefRef: DefRef{ID: newSyn.DecID, RN: newSyn.DecRN}, Title: symbol.ConvertToString(newSyn.DecQN.Sym())}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.synDecs.Insert(ds, newSyn)
		if err != nil {
			return err
		}
		err = s.typeDefs.Insert(ds, newType)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", qnAttr)
		return DefRef{}, err
	}
	s.log.Debug("inception succeed", qnAttr, slog.Any("defRef", newType.DefRef))
	return ConvertRecToRef(newType), nil
}

func (s *service) Create(spec DefSpec) (_ DefSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("typeQN", spec.TypeQN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newSyn := syndec.DecRec{DecQN: spec.TypeQN, DecID: identity.New(), DecRN: revnum.New()}
	newExp := typeexp.ConvertSpecToRec(spec.TypeES)
	newType := DefRec{
		DefRef: DefRef{ID: newSyn.DecID, RN: newSyn.DecRN},
		Title:  symbol.ConvertToString(newSyn.DecQN.Sym()),
		ExpID:  newExp.Ident(),
	}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.synDecs.Insert(ds, newSyn)
		if err != nil {
			return err
		}
		err = s.typeExps.InsertRec(ds, newExp)
		if err != nil {
			return err
		}
		err = s.typeDefs.Insert(ds, newType)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return DefSnap{}, err
	}
	s.log.Debug("creation succeed", qnAttr, slog.Any("defRef", newType.DefRef))
	return DefSnap{
		DefRef: newType.DefRef,
		Title:  newType.Title,
		TypeQN: newSyn.DecQN,
		TypeES: typeexp.ConvertRecToSpec(newExp),
	}, nil
}

func (s *service) Modify(snap DefSnap) (_ DefSnap, err error) {
	ctx := context.Background()
	refAttr := slog.Any("defRef", snap.DefRef)
	s.log.Debug("modification started", refAttr)
	var rec DefRec
	s.operator.Implicit(ctx, func(ds db.Source) error {
		rec, err = s.typeDefs.SelectRecByRef(ds, snap.DefRef)
		return err
	})
	if err != nil {
		s.log.Error("modification failed", refAttr)
		return DefSnap{}, err
	}
	if snap.DefRef.RN != rec.DefRef.RN {
		s.log.Error("modification failed", refAttr)
		return DefSnap{}, errConcurrentModification(snap.DefRef.RN, rec.DefRef.RN)
	} else {
		snap.DefRef.RN = revnum.Next(snap.DefRef.RN)
	}
	curSnap, err := s.retrieveSnap(rec)
	if err != nil {
		s.log.Error("modification failed", refAttr)
		return DefSnap{}, err
	}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		if typeexp.CheckSpec(snap.TypeES, curSnap.TypeES) != nil {
			newTerm := typeexp.ConvertSpecToRec(snap.TypeES)
			err = s.typeExps.InsertRec(ds, newTerm)
			if err != nil {
				return err
			}
			rec.ExpID = newTerm.Ident()
			rec.DefRef.RN = snap.DefRef.RN
		}
		if rec.DefRef.RN == snap.DefRef.RN {
			err = s.typeDefs.Update(ds, rec)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.log.Error("modification failed", refAttr)
		return DefSnap{}, err
	}
	s.log.Debug("modification succeed", refAttr)
	return snap, nil
}

func (s *service) RetrieveSnap(defID DefRef) (_ DefSnap, err error) {
	ctx := context.Background()
	var root DefRec
	s.operator.Implicit(ctx, func(ds db.Source) error {
		root, err = s.typeDefs.SelectRecByRef(ds, defID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("defID", defID))
		return DefSnap{}, err
	}
	return s.retrieveSnap(root)
}

func (s *service) retrieveSnap(rec DefRec) (_ DefSnap, err error) {
	ctx := context.Background()
	var termRec typeexp.ExpRec
	s.operator.Implicit(ctx, func(ds db.Source) error {
		termRec, err = s.typeExps.SelectRecByID(ds, rec.ExpID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("defRef", rec.DefRef))
		return DefSnap{}, err
	}
	return DefSnap{
		DefRef: rec.DefRef,
		Title:  rec.Title,
		TypeES: typeexp.ConvertRecToSpec(termRec),
	}, nil
}

func (s *service) RetreiveRefs() (refs []DefRef, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds db.Source) error {
		refs, err = s.typeDefs.SelectRefs(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectEnv(recs iter.Seq[DefRec]) []identity.ADT {
	termIDs := []identity.ADT{}
	for r := range recs {
		termIDs = append(termIDs, r.ExpID)
	}
	return termIDs
}

func ErrSymMissingInEnv(want uniqsym.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func errConcurrentModification(got revnum.ADT, want revnum.ADT) error {
	return fmt.Errorf("entity concurrent modification: want revision %v, got revision %v", want, got)
}

func errOptimisticUpdate(got revnum.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}

func ErrDoesNotExist(want identity.ADT) error {
	return fmt.Errorf("root doesn't exist: %v", want)
}

func ErrMissingInEnv(want identity.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func ErrMissingInCfg(want identity.ADT) error {
	return fmt.Errorf("root missing in cfg: %v", want)
}

func ErrMissingInCtx(want symbol.ADT) error {
	return fmt.Errorf("root missing in ctx: %v", want)
}
