package typedef

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
	"orglang/orglang/adt/typeexp"
)

type API interface {
	Incept(qualsym.ADT) (DefRef, error)
	Create(DefSpec) (DefSnap, error)
	Modify(DefSnap) (DefSnap, error)
	Retrieve(identity.ADT) (DefSnap, error)
	retrieveSnap(DefRec) (DefSnap, error)
	RetreiveRefs() ([]DefRef, error)
}

type DefSpec struct {
	TypeQN qualsym.ADT
	TypeES typeexp.ExpSpec
}

type DefRef struct {
	DefID identity.ADT
	DefRN revnum.ADT
}

// aka TpDef
type DefRec struct {
	DefID identity.ADT
	Title string
	ExpID identity.ADT
	DefRN revnum.ADT
}

type DefSnap struct {
	DefID  identity.ADT
	Title  string
	TypeQN qualsym.ADT
	TypeES typeexp.ExpSpec
	DefRN  revnum.ADT
}

type Context struct {
	Assets map[qualsym.ADT]typeexp.ExpRec
	Liabs  map[qualsym.ADT]typeexp.ExpRec
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

func (s *service) Incept(typeQN qualsym.ADT) (_ DefRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("typeQN", typeQN)
	s.log.Debug("inception started", qnAttr)
	newAlias := syndec.DecRec{DecQN: typeQN, DecID: identity.New(), DecRN: revnum.Initial()}
	newType := DefRec{DefID: newAlias.DecID, DefRN: newAlias.DecRN, Title: newAlias.DecQN.SN()}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.synDecs.Insert(ds, newAlias)
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
	s.log.Debug("inception succeed", qnAttr, slog.Any("defID", newType.DefID))
	return ConvertRecToRef(newType), nil
}

func (s *service) Create(spec DefSpec) (_ DefSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("typeQN", spec.TypeQN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newAlias := syndec.DecRec{DecQN: spec.TypeQN, DecID: identity.New(), DecRN: revnum.Initial()}
	newTerm := typeexp.ConvertSpecToRec(spec.TypeES)
	newType := DefRec{
		DefID: newAlias.DecID,
		DefRN: newAlias.DecRN,
		Title: newAlias.DecQN.SN(),
		ExpID: newTerm.Ident(),
	}
	s.operator.Explicit(ctx, func(ds db.Source) error {
		err = s.synDecs.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.typeExps.InsertRec(ds, newTerm)
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
	s.log.Debug("creation succeed", qnAttr, slog.Any("defID", newType.DefID))
	return DefSnap{
		DefID:  newType.DefID,
		DefRN:  newType.DefRN,
		Title:  newType.Title,
		TypeQN: newAlias.DecQN,
		TypeES: typeexp.ConvertRecToSpec(newTerm),
	}, nil
}

func (s *service) Modify(snap DefSnap) (_ DefSnap, err error) {
	ctx := context.Background()
	idAttr := slog.Any("defID", snap.DefID)
	s.log.Debug("modification started", idAttr)
	var rec DefRec
	s.operator.Implicit(ctx, func(ds db.Source) error {
		rec, err = s.typeDefs.SelectRecByID(ds, snap.DefID)
		return err
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return DefSnap{}, err
	}
	if snap.DefRN != rec.DefRN {
		s.log.Error("modification failed", idAttr)
		return DefSnap{}, errConcurrentModification(snap.DefRN, rec.DefRN)
	} else {
		snap.DefRN = revnum.Next(snap.DefRN)
	}
	curSnap, err := s.retrieveSnap(rec)
	if err != nil {
		s.log.Error("modification failed", idAttr)
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
			rec.DefRN = snap.DefRN
		}
		if rec.DefRN == snap.DefRN {
			err = s.typeDefs.Update(ds, rec)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return DefSnap{}, err
	}
	s.log.Debug("modification succeed", idAttr)
	return snap, nil
}

func (s *service) Retrieve(defID identity.ADT) (_ DefSnap, err error) {
	ctx := context.Background()
	var root DefRec
	s.operator.Implicit(ctx, func(ds db.Source) error {
		root, err = s.typeDefs.SelectRecByID(ds, defID)
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
		s.log.Error("retrieval failed", slog.Any("defID", rec.DefID))
		return DefSnap{}, err
	}
	return DefSnap{
		DefID:  rec.DefID,
		DefRN:  rec.DefRN,
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

func ErrSymMissingInEnv(want qualsym.ADT) error {
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

func ErrMissingInCtx(want qualsym.ADT) error {
	return fmt.Errorf("root missing in ctx: %v", want)
}
