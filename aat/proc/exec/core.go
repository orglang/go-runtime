package exec

import (
	"context"
	"fmt"
	"log/slog"

	"orglang/orglang/lib/sd"

	"orglang/orglang/avt/id"
	"orglang/orglang/avt/rn"
	"orglang/orglang/avt/sym"

	procdec "orglang/orglang/aat/proc/dec"
	procdef "orglang/orglang/aat/proc/def"
	typedef "orglang/orglang/aat/type/def"
)

type API interface {
	Run(ProcSpec) error
	Retrieve(id.ADT) (ProcSnap, error)
}

type SemRec interface {
	step() id.ADT
}

func ChnlID(r SemRec) id.ADT { return r.step() }

type MsgRec struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Val    procdef.TermRec
	PoolRN rn.ADT
	ProcRN rn.ADT
}

func (r MsgRec) step() id.ADT { return r.ChnlID }

type SvcRec struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Cont   procdef.TermRec
	PoolRN rn.ADT
}

func (r SvcRec) step() id.ADT { return r.ChnlID }

type ProcSpec struct {
	PoolID id.ADT
	ExecID id.ADT
	ProcTS procdef.TermSpec
}

type ProcRef struct {
	ExecID id.ADT
}

type ProcSnap struct {
	ExecID id.ADT
}

type MainCfg struct {
	ProcID id.ADT
	Bnds   map[sym.ADT]EP2
	Acts   map[id.ADT]SemRec
	PoolID id.ADT
	ProcRN rn.ADT
}

// aka Configuration
type Cfg struct {
	ProcID id.ADT
	Chnls  map[sym.ADT]EP
	Steps  map[id.ADT]SemRec
	PoolID id.ADT
	PoolRN rn.ADT
	ProcRN rn.ADT
}

type Env struct {
	ProcSigs  map[id.ADT]procdec.ProcRec
	Types     map[sym.ADT]typedef.TypeRec
	TypeTerms map[id.ADT]typedef.TermRec
	Locks     map[sym.ADT]Lock
}

type EP struct {
	ChnlPH sym.ADT
	ChnlID id.ADT
	TermID id.ADT
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

func ChnlPH(rec EP) sym.ADT { return rec.ChnlPH }

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
	Steps []SemRec
	Liabs []Liab
}

type MainMod struct {
	Bnds []Bnd
	Acts []SemRec
}

type Bnd struct {
	ProcID id.ADT
	ChnlPH sym.ADT
	ChnlID id.ADT
	TermID id.ADT
	PoolRN rn.ADT
	ProcRN rn.ADT
}

type service struct {
	procs    repo
	operator sd.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	procs repo,
	operator sd.Operator,
	l *slog.Logger,
) *service {
	return &service{procs, operator, l}
}

func (s *service) Run(spec ProcSpec) (err error) {
	idAttr := slog.Any("procID", spec.ExecID)
	s.log.Debug("creation started", idAttr)
	ctx := context.Background()
	var mainCfg MainCfg
	err = s.operator.Implicit(ctx, func(ds sd.Source) error {
		mainCfg, err = s.procs.SelectMain(ds, spec.ExecID)
		return err
	})
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return err
	}
	var mainEnv Env
	err = s.checkType(spec.PoolID, mainEnv, mainCfg, spec.ProcTS)
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return err
	}
	mainMod, err := s.createWith(mainEnv, mainCfg, spec.ProcTS)
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return err
	}
	err = s.operator.Explicit(ctx, func(ds sd.Source) error {
		err = s.procs.UpdateMain(ds, mainMod)
		if err != nil {
			s.log.Error("creation failed", idAttr)
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return err
	}
	return nil
}

func (s *service) Retrieve(procID id.ADT) (_ ProcSnap, err error) {
	return ProcSnap{}, nil
}

func (s *service) checkType(
	poolID id.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	termSpec procdef.TermSpec,
) error {
	imp, ok := mainCfg.Bnds[termSpec.Via()]
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
	ts procdef.TermSpec,
) error {
	return nil
}

func (s *service) checkClient(
	poolID id.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	ts procdef.TermSpec,
) error {
	return nil
}

func (s *service) createWith(
	mainEnv Env,
	procCfg MainCfg,
	ts procdef.TermSpec,
) (
	procMod MainMod,
	_ error,
) {
	switch termSpec := ts.(type) {
	case procdef.CallSpecOld:
		viaCord, ok := procCfg.Bnds[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
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
			sndrAct := MsgRec{}
			procMod.Acts = append(procMod.Acts, sndrAct)
			s.log.Debug("coordination half done", viaAttr)
			return procMod, nil
		}
		s.log.Debug("coordination succeed")
		return procMod, nil
	case procdef.SpawnSpecOld:
		s.log.Debug("coordination succeed")
		return procMod, nil
	default:
		panic(procdef.ErrTermTypeUnexpected(ts))
	}
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

func ErrRootTypeUnexpected(got SemRec) error {
	return fmt.Errorf("sem rec unexpected: %T", got)
}

func ErrRootTypeMismatch(got, want SemRec) error {
	return fmt.Errorf("sem rec mismatch: want %T, got %T", want, got)
}
