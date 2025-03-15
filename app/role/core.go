package role

import (
	"context"
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rev"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/alias"
	"smecalculus/rolevod/internal/state"
)

// for external readability
type ID = id.ADT
type Rev = rev.ADT
type QN = sym.ADT
type Title = string

type Spec struct {
	QN    sym.ADT
	State state.Spec
}

type Ref struct {
	ID    id.ADT
	Rev   rev.ADT
	Title string
}

type Snap struct {
	ID    id.ADT
	Rev   rev.ADT
	Title string
	QN    sym.ADT
	State state.Spec
	// Parts   []Ref
}

// aka TpDef
type Root struct {
	ID    id.ADT
	Rev   rev.ADT
	Title string
	// specification relation
	StateID state.ID
	// composition relation
	WholeID id.ADT
}

type API interface {
	Incept(sym.ADT) (Ref, error)
	Create(Spec) (Snap, error)
	Modify(Snap) (Snap, error)
	Retrieve(id.ADT) (Snap, error)
	RetrieveRoot(id.ADT) (Root, error)
	RetrieveSnap(Root) (Snap, error)
	RetreiveRefs() ([]Ref, error)
}

type service struct {
	roles    Repo
	states   state.Repo
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
	states state.Repo,
	aliases alias.Repo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	name := slog.String("name", "roleService")
	return &service{roles, states, aliases, operator, l.With(name)}
}

func (s *service) Incept(fqn sym.ADT) (_ Ref, err error) {
	ctx := context.Background()
	fqnAttr := slog.Any("fqn", fqn)
	s.log.Debug("inception started", fqnAttr)
	newAlias := alias.Root{Sym: fqn, ID: id.New(), Rev: rev.Initial()}
	newRoot := Root{ID: newAlias.ID, Rev: newAlias.Rev, Title: newAlias.Sym.Name()}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.aliases.Insert(ds, newAlias)
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
		s.log.Error("inception failed", fqnAttr)
		return Ref{}, err
	}
	s.log.Debug("inception succeeded", fqnAttr, slog.Any("id", newRoot.ID))
	return ConvertRootToRef(newRoot), nil
}

func (s *service) Create(spec Spec) (_ Snap, err error) {
	ctx := context.Background()
	fqnAttr := slog.Any("fqn", spec.QN)
	s.log.Debug("creation started", fqnAttr, slog.Any("spec", spec))
	newAlias := alias.Root{Sym: spec.QN, ID: id.New(), Rev: rev.Initial()}
	newState := state.ConvertSpecToRoot(spec.State)
	newRoot := Root{
		ID:      newAlias.ID,
		Rev:     newAlias.Rev,
		Title:   newAlias.Sym.Name(),
		StateID: newState.Ident(),
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
		s.log.Error("creation failed", fqnAttr)
		return Snap{}, err
	}
	s.log.Debug("creation succeeded", fqnAttr, slog.Any("id", newRoot.ID))
	return Snap{
		ID:    newRoot.ID,
		Rev:   newRoot.Rev,
		Title: newRoot.Title,
		QN:    newAlias.Sym,
		State: state.ConvertRootToSpec(newState),
	}, nil
}

func (s *service) Modify(newSnap Snap) (_ Snap, err error) {
	ctx := context.Background()
	idAttr := slog.Any("id", newSnap.ID)
	s.log.Debug("modification started", idAttr)
	var curRoot Root
	s.operator.Implicit(ctx, func(ds data.Source) {
		curRoot, err = s.roles.SelectByID(ds, newSnap.ID)
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return Snap{}, err
	}
	if newSnap.Rev != curRoot.Rev {
		s.log.Error("modification failed", idAttr)
		return Snap{}, errConcurrentModification(newSnap.Rev, curRoot.Rev)
	} else {
		newSnap.Rev = rev.Next(newSnap.Rev)
	}
	curSnap, err := s.RetrieveSnap(curRoot)
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return Snap{}, err
	}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		if state.CheckSpec(newSnap.State, curSnap.State) != nil {
			newState := state.ConvertSpecToRoot(newSnap.State)
			err = s.states.Insert(ds, newState)
			if err != nil {
				return err
			}
			curRoot.StateID = newState.Ident()
			curRoot.Rev = newSnap.Rev
		}
		if curRoot.Rev == newSnap.Rev {
			err = s.roles.Update(ds, curRoot)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return Snap{}, err
	}
	s.log.Debug("modification succeeded", idAttr)
	return newSnap, nil
}

func (s *service) Retrieve(rid ID) (_ Snap, err error) {
	ctx := context.Background()
	var root Root
	s.operator.Implicit(ctx, func(ds data.Source) {
		root, err = s.roles.SelectByID(ds, rid)
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("id", rid))
		return Snap{}, err
	}
	return s.RetrieveSnap(root)
}

func (s *service) RetrieveRoot(rid ID) (root Root, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) {
		root, err = s.roles.SelectByID(ds, rid)
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("id", rid))
		return Root{}, err
	}
	return root, nil
}

func (s *service) RetrieveSnap(root Root) (_ Snap, err error) {
	ctx := context.Background()
	var curState state.Root
	s.operator.Implicit(ctx, func(ds data.Source) {
		curState, err = s.states.SelectByID(ds, root.StateID)
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("id", root.ID))
		return Snap{}, err
	}
	return Snap{
		ID:    root.ID,
		Rev:   root.Rev,
		Title: root.Title,
		State: state.ConvertRootToSpec(curState),
	}, nil
}

func (s *service) RetreiveRefs() (refs []Ref, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) {
		refs, err = s.roles.SelectRefs(ds)
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectEnv(roles []Root) []id.ADT {
	stateIDs := []id.ADT{}
	for _, r := range roles {
		stateIDs = append(stateIDs, r.StateID)
	}
	return stateIDs
}

type Repo interface {
	Insert(data.Source, Root) error
	Update(data.Source, Root) error
	SelectRefs(data.Source) ([]Ref, error)
	SelectByID(data.Source, id.ADT) (Root, error)
	SelectByIDs(data.Source, []id.ADT) ([]Root, error)
	SelectByFQN(data.Source, sym.ADT) (Root, error)
	SelectByFQNs(data.Source, []sym.ADT) ([]Root, error)
	SelectEnv(data.Source, []sym.ADT) (map[sym.ADT]Root, error)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/internal/state:Convert.*
var (
	ConvertRootToRef func(Root) Ref
	ConvertSnapToRef func(Snap) Ref
)

func ErrMissingInEnv(want QN) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func errConcurrentModification(got rev.ADT, want rev.ADT) error {
	return fmt.Errorf("entity concurrent modification: want revision %v, got revision %v", want, got)
}

func errOptimisticUpdate(got rev.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}
