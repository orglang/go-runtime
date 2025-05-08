package dec

import (
	"context"
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/app/type/def"

	"smecalculus/rolevod/internal/alias"
)

type TypeSpec struct {
	TypeNS sym.ADT
	TypeSN sym.ADT
	TypeTS def.TermSpec
}

type TypeRef struct {
	TypeID id.ADT
	Title  string
	TypeRN rn.ADT
}

// aka TpDef
type TypeRec struct {
	TypeID id.ADT
	Title  string
	TermID id.ADT
	TypeRN rn.ADT
}

type TypeSnap struct {
	TypeID id.ADT
	Title  string
	TypeQN sym.ADT
	TypeTS def.TermSpec
	TypeRN rn.ADT
}

type API interface {
	Incept(sym.ADT) (TypeRef, error)
	Create(TypeSpec) (TypeSnap, error)
	Modify(TypeSnap) (TypeSnap, error)
	Retrieve(id.ADT) (TypeSnap, error)
	retrieveSnap(TypeRec) (TypeSnap, error)
	RetreiveRefs() ([]TypeRef, error)
}

type service struct {
	roles    Repo
	states   def.TermRepo
	aliases  alias.Repo
	operator data.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	roles Repo,
	states def.TermRepo,
	aliases alias.Repo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	return &service{roles, states, aliases, operator, l}
}

func (s *service) Incept(qn sym.ADT) (_ TypeRef, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("roleQN", qn)
	s.log.Debug("inception started", qnAttr)
	newAlias := alias.Root{Sym: qn, ID: id.New(), RN: rn.Initial()}
	newType := TypeRec{TypeID: newAlias.ID, TypeRN: newAlias.RN, Title: newAlias.Sym.SN()}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.aliases.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.roles.Insert(ds, newType)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", qnAttr)
		return TypeRef{}, err
	}
	s.log.Debug("inception succeeded", qnAttr, slog.Any("roleID", newType.TypeID))
	return ConvertRecToRef(newType), nil
}

func (s *service) Create(ts TypeSpec) (_ TypeSnap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("roleQN", ts.TypeSN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", ts))
	newAlias := alias.Root{Sym: ts.TypeSN, ID: id.New(), RN: rn.Initial()}
	newState := def.ConvertSpecToRec(ts.TypeTS)
	newRoot := TypeRec{
		TypeID: newAlias.ID,
		TypeRN: newAlias.RN,
		Title:  newAlias.Sym.SN(),
		TermID: newState.Ident(),
	}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.aliases.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.states.Insert(ds, newState)
		if err != nil {
			return err
		}
		err = s.roles.Insert(ds, newRoot)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return TypeSnap{}, err
	}
	s.log.Debug("creation succeeded", qnAttr, slog.Any("roleID", newRoot.TypeID))
	return TypeSnap{
		TypeID: newRoot.TypeID,
		TypeRN: newRoot.TypeRN,
		Title:  newRoot.Title,
		TypeQN: newAlias.Sym,
		TypeTS: def.ConvertRecToSpec(newState),
	}, nil
}

func (s *service) Modify(newSnap TypeSnap) (_ TypeSnap, err error) {
	ctx := context.Background()
	idAttr := slog.Any("roleID", newSnap.TypeID)
	s.log.Debug("modification started", idAttr)
	var curRoot TypeRec
	s.operator.Implicit(ctx, func(ds data.Source) error {
		curRoot, err = s.roles.SelectByID(ds, newSnap.TypeID)
		return err
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, err
	}
	if newSnap.TypeRN != curRoot.TypeRN {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, errConcurrentModification(newSnap.TypeRN, curRoot.TypeRN)
	} else {
		newSnap.TypeRN = rn.Next(newSnap.TypeRN)
	}
	curSnap, err := s.retrieveSnap(curRoot)
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, err
	}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		if def.CheckSpec(newSnap.TypeTS, curSnap.TypeTS) != nil {
			newState := def.ConvertSpecToRec(newSnap.TypeTS)
			err = s.states.Insert(ds, newState)
			if err != nil {
				return err
			}
			curRoot.TermID = newState.Ident()
			curRoot.TypeRN = newSnap.TypeRN
		}
		if curRoot.TypeRN == newSnap.TypeRN {
			err = s.roles.Update(ds, curRoot)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return TypeSnap{}, err
	}
	s.log.Debug("modification succeeded", idAttr)
	return newSnap, nil
}

func (s *service) Retrieve(typeID id.ADT) (_ TypeSnap, err error) {
	ctx := context.Background()
	var root TypeRec
	s.operator.Implicit(ctx, func(ds data.Source) error {
		root, err = s.roles.SelectByID(ds, typeID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", typeID))
		return TypeSnap{}, err
	}
	return s.retrieveSnap(root)
}

func (s *service) RetrieveImpl(typeID id.ADT) (enty TypeRec, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) error {
		enty, err = s.roles.SelectByID(ds, typeID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", typeID))
		return TypeRec{}, err
	}
	return enty, nil
}

func (s *service) retrieveSnap(enty TypeRec) (_ TypeSnap, err error) {
	ctx := context.Background()
	var curState def.TermRec
	s.operator.Implicit(ctx, func(ds data.Source) error {
		curState, err = s.states.SelectByID(ds, enty.TermID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", enty.TypeID))
		return TypeSnap{}, err
	}
	return TypeSnap{
		TypeID: enty.TypeID,
		TypeRN: enty.TypeRN,
		Title:  enty.Title,
		TypeTS: def.ConvertRecToSpec(curState),
	}, nil
}

func (s *service) RetreiveRefs() (refs []TypeRef, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) error {
		refs, err = s.roles.SelectRefs(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectEnv(roots []TypeRec) []id.ADT {
	termIDs := []id.ADT{}
	for _, r := range roots {
		termIDs = append(termIDs, r.TermID)
	}
	return termIDs
}

type Repo interface {
	Insert(data.Source, TypeRec) error
	Update(data.Source, TypeRec) error
	SelectRefs(data.Source) ([]TypeRef, error)
	SelectByID(data.Source, id.ADT) (TypeRec, error)
	SelectByIDs(data.Source, []id.ADT) ([]TypeRec, error)
	SelectByQN(data.Source, sym.ADT) (TypeRec, error)
	SelectByQNs(data.Source, []sym.ADT) ([]TypeRec, error)
	SelectEnv(data.Source, []sym.ADT) (map[sym.ADT]TypeRec, error)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/app/type/def:Convert.*
var (
	ConvertRecToRef  func(TypeRec) TypeRef
	ConvertSnapToRef func(TypeSnap) TypeRef
)

func ErrMissingInEnv(want sym.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func errConcurrentModification(got rn.ADT, want rn.ADT) error {
	return fmt.Errorf("entity concurrent modification: want revision %v, got revision %v", want, got)
}

func errOptimisticUpdate(got rn.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}
