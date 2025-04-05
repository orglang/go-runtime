package role

import (
	"context"
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/alias"
	"smecalculus/rolevod/internal/state"
)

type Spec struct {
	RoleQN sym.ADT
	State  state.Spec
}

type Ref struct {
	RoleID id.ADT
	Title  string
	RoleRN rn.ADT
}

// aka TpDef
type Impl struct {
	RoleID  id.ADT
	Title   string
	StateID state.ID
	RoleRN  rn.ADT
}

type Snap struct {
	RoleID id.ADT
	Title  string
	RoleQN sym.ADT
	State  state.Spec
	RoleRN rn.ADT
}

type API interface {
	Incept(sym.ADT) (Ref, error)
	Create(Spec) (Snap, error)
	Modify(Snap) (Snap, error)
	Retrieve(id.ADT) (Snap, error)
	RetrieveImpl(id.ADT) (Impl, error)
	RetrieveSnap(Impl) (Snap, error)
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
	return &service{roles, states, aliases, operator, l}
}

func (s *service) Incept(roleQN sym.ADT) (_ Ref, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("roleQN", roleQN)
	s.log.Debug("inception started", qnAttr)
	newAlias := alias.Root{Sym: roleQN, ID: id.New(), RN: rn.Initial()}
	newRoot := Impl{RoleID: newAlias.ID, RoleRN: newAlias.RN, Title: newAlias.Sym.SN()}
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
		s.log.Error("inception failed", qnAttr)
		return Ref{}, err
	}
	s.log.Debug("inception succeeded", qnAttr, slog.Any("roleID", newRoot.RoleID))
	return ConvertRootToRef(newRoot), nil
}

func (s *service) Create(spec Spec) (_ Snap, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("roleQN", spec.RoleQN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newAlias := alias.Root{Sym: spec.RoleQN, ID: id.New(), RN: rn.Initial()}
	newState := state.ConvertSpecToRoot(spec.State)
	newRoot := Impl{
		RoleID:  newAlias.ID,
		RoleRN:  newAlias.RN,
		Title:   newAlias.Sym.SN(),
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
		s.log.Error("creation failed", qnAttr)
		return Snap{}, err
	}
	s.log.Debug("creation succeeded", qnAttr, slog.Any("roleID", newRoot.RoleID))
	return Snap{
		RoleID: newRoot.RoleID,
		RoleRN: newRoot.RoleRN,
		Title:  newRoot.Title,
		RoleQN: newAlias.Sym,
		State:  state.ConvertRootToSpec(newState),
	}, nil
}

func (s *service) Modify(newSnap Snap) (_ Snap, err error) {
	ctx := context.Background()
	idAttr := slog.Any("roleID", newSnap.RoleID)
	s.log.Debug("modification started", idAttr)
	var curRoot Impl
	s.operator.Implicit(ctx, func(ds data.Source) error {
		curRoot, err = s.roles.SelectByID(ds, newSnap.RoleID)
		return err
	})
	if err != nil {
		s.log.Error("modification failed", idAttr)
		return Snap{}, err
	}
	if newSnap.RoleRN != curRoot.RoleRN {
		s.log.Error("modification failed", idAttr)
		return Snap{}, errConcurrentModification(newSnap.RoleRN, curRoot.RoleRN)
	} else {
		newSnap.RoleRN = rn.Next(newSnap.RoleRN)
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
			curRoot.RoleRN = newSnap.RoleRN
		}
		if curRoot.RoleRN == newSnap.RoleRN {
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

func (s *service) Retrieve(roleID id.ADT) (_ Snap, err error) {
	ctx := context.Background()
	var root Impl
	s.operator.Implicit(ctx, func(ds data.Source) error {
		root, err = s.roles.SelectByID(ds, roleID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", roleID))
		return Snap{}, err
	}
	return s.RetrieveSnap(root)
}

func (s *service) RetrieveImpl(roleID id.ADT) (root Impl, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) error {
		root, err = s.roles.SelectByID(ds, roleID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", roleID))
		return Impl{}, err
	}
	return root, nil
}

func (s *service) RetrieveSnap(root Impl) (_ Snap, err error) {
	ctx := context.Background()
	var curState state.Root
	s.operator.Implicit(ctx, func(ds data.Source) error {
		curState, err = s.states.SelectByID(ds, root.StateID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("roleID", root.RoleID))
		return Snap{}, err
	}
	return Snap{
		RoleID: root.RoleID,
		RoleRN: root.RoleRN,
		Title:  root.Title,
		State:  state.ConvertRootToSpec(curState),
	}, nil
}

func (s *service) RetreiveRefs() (refs []Ref, err error) {
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

func CollectEnv(roles []Impl) []id.ADT {
	stateIDs := []id.ADT{}
	for _, r := range roles {
		stateIDs = append(stateIDs, r.StateID)
	}
	return stateIDs
}

type Repo interface {
	Insert(data.Source, Impl) error
	Update(data.Source, Impl) error
	SelectRefs(data.Source) ([]Ref, error)
	SelectByID(data.Source, id.ADT) (Impl, error)
	SelectByIDs(data.Source, []id.ADT) ([]Impl, error)
	SelectByFQN(data.Source, sym.ADT) (Impl, error)
	SelectByFQNs(data.Source, []sym.ADT) ([]Impl, error)
	SelectEnv(data.Source, []sym.ADT) (map[sym.ADT]Impl, error)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend smecalculus/rolevod/internal/state:Convert.*
var (
	ConvertRootToRef func(Impl) Ref
	ConvertSnapToRef func(Snap) Ref
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
