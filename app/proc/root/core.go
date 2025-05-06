package root

import (
	"context"
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/chnl"
	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	procsig "smecalculus/rolevod/app/proc/sig"
	"smecalculus/rolevod/app/proc/xact"
	rolesig "smecalculus/rolevod/app/role/sig"
)

type Spec struct {
	PoolID id.ADT
	ProcID id.ADT
	Term   xact.TermSpec
}

type Ref struct {
	ProcID id.ADT
}

type Snap struct {
	ProcID id.ADT
}

type MainCfg struct {
	ProcID id.ADT
	Bnds   map[sym.ADT]EP2
	Acts   map[id.ADT]xact.Sem
	PoolID id.ADT
	ProcRN rn.ADT
}

// aka Configuration
type Cfg struct {
	ProcID id.ADT
	Chnls  map[sym.ADT]EP
	Steps  map[id.ADT]step.Root
	PoolID id.ADT
	PoolRN rn.ADT
	ProcRN rn.ADT
}

type Env struct {
	Sigs   map[id.ADT]procsig.Impl
	Roles  map[sym.ADT]rolesig.Impl
	States map[state.ID]state.Root
	Locks  map[sym.ADT]Lock
}

type EP struct {
	ChnlPH  sym.ADT
	ChnlID  id.ADT
	StateID id.ADT
	// provider
	PoolID id.ADT
}

type EP2 struct {
	CordPH sym.ADT
	CordID id.ADT
	// provider
	PoolID id.ADT
}

type Lock struct {
	PoolID id.ADT
	PoolRN rn.ADT
}

func ChnlPH(ch EP) sym.ADT { return ch.ChnlPH }

func ChnlID(ch EP) id.ADT { return ch.ChnlID }

// ответственность за процесс
type Liab struct {
	PoolID id.ADT
	ProcID id.ADT
	// позитивное значение при вручении
	// негативное значение при лишении
	PoolRN rn.ADT
}

type Mod struct {
	Locks []Lock
	Bnds  []Bnd
	Steps []step.Root
	Liabs []Liab
}

type MainMod struct {
	Bnds []Bnd
	Acts []xact.Sem
}

type Bnd struct {
	ProcID  id.ADT
	ChnlPH  sym.ADT
	ChnlID  id.ADT
	StateID id.ADT
	PoolRN  rn.ADT
	ProcRN  rn.ADT
}

type API interface {
	Create(Spec) (Ref, error)
	Retrieve(id.ADT) (Snap, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	procs    repo
	operator data.Operator
	log      *slog.Logger
}

func newService(
	procs repo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	return &service{procs, operator, l}
}

func (s *service) Create(spec Spec) (_ Ref, err error) {
	idAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("creation started", idAttr)
	ctx := context.Background()
	var mainCfg MainCfg
	err = s.operator.Implicit(ctx, func(ds data.Source) error {
		mainCfg, err = s.procs.SelectMain(ds, spec.ProcID)
		return err
	})
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return Ref{}, err
	}
	var mainEnv Env
	err = s.checkType(spec.PoolID, mainEnv, mainCfg, spec.Term)
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return Ref{}, err
	}
	mainMod, err := s.createWith(mainEnv, mainCfg, spec.Term)
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return Ref{}, err
	}
	err = s.operator.Explicit(ctx, func(ds data.Source) error {
		err = s.procs.UpdateMain(ds, mainMod)
		if err != nil {
			s.log.Error("creation failed", idAttr)
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return Ref{}, err
	}
	return Ref{}, nil
}

func (s *service) Retrieve(procID id.ADT) (_ Snap, err error) {
	return Snap{}, nil
}

func (s *service) checkType(
	poolID id.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	termSpec xact.TermSpec,
) error {
	imp, ok := mainCfg.Bnds[termSpec.ConnPH()]
	if !ok {
		panic("no via in main cfg")
	}
	if poolID == imp.PoolID {
		return s.checkProvider(poolID, mainEnv, mainCfg, termSpec)
	} else {
		return s.checkClient(poolID, mainEnv, mainCfg, termSpec)
	}
}

func (s *service) checkProvider(
	poolID id.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	ts xact.TermSpec,
) error {
	return nil
}

func (s *service) checkClient(
	poolID id.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	ts xact.TermSpec,
) error {
	return nil
}

func (s *service) createWith(
	mainEnv Env,
	procCfg MainCfg,
	ts xact.TermSpec,
) (
	procMod MainMod,
	_ error,
) {
	switch termSpec := ts.(type) {
	case xact.CallSpec:
		viaCord, ok := procCfg.Bnds[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("coordination failed")
			return MainMod{}, err
		}
		viaAttr := slog.Any("cordID", viaCord.CordID)
		for _, chnlPH := range termSpec.Ys {
			sndrValBnd := Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: chnlPH,
				ProcRN: -procCfg.ProcRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrValBnd)
		}
		rcvrAct := procCfg.Acts[viaCord.CordID]
		if rcvrAct == nil {
			sndrAct := xact.MsgSem{}
			procMod.Acts = append(procMod.Acts, sndrAct)
			s.log.Debug("coordination half done", viaAttr)
			return procMod, nil
		}
		s.log.Debug("coordination succeeded")
		return procMod, nil
	case xact.SpawnSpec:
		s.log.Debug("coordination succeeded")
		return procMod, nil
	default:
		panic(xact.ErrTermTypeUnexpected(ts))
	}
}

type repo interface {
	SelectMain(data.Source, id.ADT) (MainCfg, error)
	UpdateMain(data.Source, MainMod) error
}

func ErrMissingChnl(want sym.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func errMissingPool(want sym.ADT) error {
	return fmt.Errorf("pool missing in env: %v", want)
}

func errMissingSig(want id.ADT) error {
	return fmt.Errorf("sig missing in env: %v", want)
}

func errMissingRole(want sym.ADT) error {
	return fmt.Errorf("role missing in env: %v", want)
}
