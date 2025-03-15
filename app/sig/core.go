package sig

import (
	"context"
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/rev"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/alias"
	"smecalculus/rolevod/internal/chnl"

	"smecalculus/rolevod/app/role"
)

type ID = id.ADT
type QN = sym.ADT
type Name = string

type Spec struct {
	QN sym.ADT
	Ys []chnl.Spec // vals
	X  chnl.Spec   // via
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
	CEs   []chnl.Spec
	PE    chnl.Spec
}

// aka ExpDec or ExpDecDef without expression
type Root struct {
	SigID id.ADT
	Rev   rev.ADT
	Title string
	Ys2   []chnl.Spec
	X2    chnl.Spec
	Ys    []EP
	X     EP
}

// aka ChanTp
type EP struct {
	ChnlPH ph.ADT
	RoleQN sym.ADT
}

type API interface {
	Incept(QN) (Ref, error)
	Create(Spec) (Root, error)
	Retrieve(id.ADT) (Root, error)
	RetreiveRefs() ([]Ref, error)
}

type service struct {
	sigs     Repo
	aliases  alias.Repo
	operator data.Operator
	log      *slog.Logger
}

func newService(sigs Repo, aliases alias.Repo, operator data.Operator, l *slog.Logger) *service {
	name := slog.String("name", "sigService")
	return &service{sigs, aliases, operator, l.With(name)}
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func (s *service) Incept(fqn sym.ADT) (_ Ref, err error) {
	ctx := context.Background()
	fqnAttr := slog.Any("fqn", fqn)
	s.log.Debug("inception started", fqnAttr)
	newAlias := alias.Root{Sym: fqn, ID: id.New(), Rev: rev.Initial()}
	newRoot := Root{SigID: newAlias.ID, Rev: newAlias.Rev, Title: newAlias.Sym.Name()}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.aliases.Insert(ds, newAlias)
		if err != nil {
			return err
		}
		err = s.sigs.Insert(ds, newRoot)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("inception failed", fqnAttr)
		return Ref{}, err
	}
	s.log.Debug("inception succeeded", fqnAttr, slog.Any("id", newRoot.SigID))
	return ConvertRootToRef(newRoot), nil
}

func (s *service) Create(spec Spec) (_ Root, err error) {
	ctx := context.Background()
	fqnAttr := slog.Any("fqn", spec.QN)
	s.log.Debug("creation started", fqnAttr, slog.Any("spec", spec))
	root := Root{
		SigID: id.New(),
		Rev:   rev.Initial(),
		Title: spec.QN.Name(),
		X2:    spec.X,
		Ys2:   spec.Ys,
	}
	s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.sigs.Insert(ds, root)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", fqnAttr)
		return Root{}, err
	}
	s.log.Debug("creation succeeded", fqnAttr, slog.Any("id", root.SigID))
	return root, nil
}

func (s *service) Retrieve(rid ID) (root Root, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) {
		root, err = s.sigs.SelectByID(ds, rid)
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("id", rid))
		return Root{}, err
	}
	return root, nil
}

func (s *service) RetreiveRefs() (refs []Ref, err error) {
	ctx := context.Background()
	s.operator.Implicit(ctx, func(ds data.Source) {
		refs, err = s.sigs.SelectAll(ds)
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

type Repo interface {
	Insert(data.Source, Root) error
	SelectAll(data.Source) ([]Ref, error)
	SelectByID(data.Source, ID) (Root, error)
	SelectByIDs(data.Source, []ID) ([]Root, error)
	SelectEnv(data.Source, []ID) (map[ID]Root, error)
}

func CollectEnv(sigs []Root) []role.QN {
	roleFQNs := []role.QN{}
	for _, sig := range sigs {
		roleFQNs = append(roleFQNs, sig.X2.RoleQN)
		for _, ce := range sig.Ys2 {
			roleFQNs = append(roleFQNs, ce.RoleQN)
		}
	}
	return roleFQNs
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	ConvertRootToRef func(Root) Ref
)

func ErrRootMissingInEnv(rid ID) error {
	return fmt.Errorf("root missing in env: %v", rid)
}
