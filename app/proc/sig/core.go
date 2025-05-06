package sig

import (
	"context"
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/alias"

	"smecalculus/rolevod/app/proc/bnd"
)

type Spec struct {
	X     bnd.Spec // via
	SigNS sym.ADT
	SigSN sym.ADT
	Ys    []bnd.Spec // vals
}

type Ref struct {
	SigID id.ADT
	Title string
	SigRN rn.ADT
}

// aka ExpDec or ExpDecDef without expression
type Impl struct {
	X     bnd.Spec
	SigID id.ADT
	Ys    []bnd.Spec
	Title string
	SigRN rn.ADT
}

type Snap struct {
	X     bnd.Spec
	SigID id.ADT
	Ys    []bnd.Spec
	Title string
	SigRN rn.ADT
}

type API interface {
	Incept(sym.ADT) (Ref, error)
	Create(Spec) (Impl, error)
	Retrieve(id.ADT) (Impl, error)
	RetreiveRefs() ([]Ref, error)
}

type service struct {
	sigs     Repo
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

func (s *service) Incept(sigQN sym.ADT) (_ Ref, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("sigQN", sigQN)
	s.log.Debug("inception started", qnAttr)
	newAlias := alias.Root{Sym: sigQN, ID: id.New(), RN: rn.Initial()}
	newImpl := Impl{SigID: newAlias.ID, SigRN: newAlias.RN, Title: newAlias.Sym.SN()}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.aliases.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.sigs.Insert(ds, newImpl)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", qnAttr)
		return Ref{}, err
	}
	s.log.Debug("inception succeeded", qnAttr, slog.Any("sigId", newImpl.SigID))
	return ConvertRootToRef(newImpl), nil
}

func (s *service) Create(spec Spec) (_ Impl, err error) {
	ctx := context.Background()
	qnAttr := slog.Any("sigQN", spec.SigSN)
	s.log.Debug("creation started", qnAttr, slog.Any("spec", spec))
	newImpl := Impl{
		X:     spec.X,
		SigID: id.New(),
		Ys:    spec.Ys,
		SigRN: rn.Initial(),
	}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.sigs.Insert(ds, newImpl)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", qnAttr)
		return Impl{}, err
	}
	s.log.Debug("creation succeeded", qnAttr, slog.Any("sigId", newImpl.SigID))
	return newImpl, nil
}

func (s *service) Retrieve(sigID id.ADT) (impl Impl, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) error {
		impl, err = s.sigs.SelectByID(ds, sigID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("sigId", sigID))
		return Impl{}, err
	}
	return impl, nil
}

func (s *service) RetreiveRefs() (refs []Ref, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) error {
		refs, err = s.sigs.SelectAll(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

type Repo interface {
	Insert(data.Source, Impl) error
	SelectAll(data.Source) ([]Ref, error)
	SelectByID(data.Source, id.ADT) (Impl, error)
	SelectByIDs(data.Source, []id.ADT) ([]Impl, error)
	SelectEnv(data.Source, []id.ADT) (map[id.ADT]Impl, error)
}

func CollectEnv(sigs []Impl) []sym.ADT {
	roleQNs := []sym.ADT{}
	for _, sig := range sigs {
		roleQNs = append(roleQNs, sig.X.RoleQN)
		for _, y := range sig.Ys {
			roleQNs = append(roleQNs, y.RoleQN)
		}
	}
	return roleQNs
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	ConvertRootToRef func(Impl) Ref
)

func ErrRootMissingInEnv(rid id.ADT) error {
	return fmt.Errorf("root missing in env: %v", rid)
}
