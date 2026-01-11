package procexec

import (
	"context"
	"fmt"
	"log/slog"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procdec"
	"orglang/orglang/adt/procdef"
	"orglang/orglang/adt/procexp"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/revnum"
	"orglang/orglang/adt/typedef"
	"orglang/orglang/adt/typeexp"
)

type API interface {
	Run(ExecSpec) error
	Retrieve(identity.ADT) (ExecSnap, error)
}

type SemRec interface {
	step() identity.ADT
}

func ChnlID(r SemRec) identity.ADT { return r.step() }

type MsgRec struct {
	PoolID identity.ADT
	ProcID identity.ADT
	ChnlID identity.ADT
	Val    procexp.ExpRec
	PoolRN revnum.ADT
	ProcRN revnum.ADT
}

func (r MsgRec) step() identity.ADT { return r.ChnlID }

type SvcRec struct {
	PoolID identity.ADT
	ProcID identity.ADT
	ChnlID identity.ADT
	Cont   procexp.ExpRec
	PoolRN revnum.ADT
}

func (r SvcRec) step() identity.ADT { return r.ChnlID }

type ExecSpec struct {
	PoolID identity.ADT
	ExecID identity.ADT
	ProcES procexp.ExpSpec
}

type ExecRef struct {
	ExecID identity.ADT
}

type ExecSnap struct {
	ExecID identity.ADT
}

type MainCfg struct {
	ExecID identity.ADT
	Bnds   map[qualsym.ADT]EP2
	Acts   map[identity.ADT]SemRec
	PoolID identity.ADT
	ExecRN revnum.ADT
}

// aka Configuration
type Cfg struct {
	ExecID identity.ADT
	Chnls  map[qualsym.ADT]EP
	Steps  map[identity.ADT]SemRec
	PoolID identity.ADT
	PoolRN revnum.ADT
	ExecRN revnum.ADT
}

type Env struct {
	ProcDecs  map[identity.ADT]procdec.DecRec
	TypeDefs  map[qualsym.ADT]typedef.DefRec
	TypeTerms map[identity.ADT]typeexp.ExpRec
	Locks     map[qualsym.ADT]Lock
}

type EP struct {
	ChnlPH qualsym.ADT
	ChnlID identity.ADT
	ExpID  identity.ADT
	// provider
	PoolID identity.ADT
}

type EP2 struct {
	CordPH qualsym.ADT
	CordID identity.ADT
	// provider
	PoolID identity.ADT
}

type Lock struct {
	PoolID identity.ADT
	PoolRN revnum.ADT
}

func ChnlPH(rec EP) qualsym.ADT { return rec.ChnlPH }

// ответственность за процесс
type Liab struct {
	PoolID identity.ADT
	ProcID identity.ADT
	// позитивное значение при вручении
	// негативное значение при лишении
	PoolRN revnum.ADT
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
	ProcID identity.ADT
	ChnlPH qualsym.ADT
	ChnlID identity.ADT
	ExpID  identity.ADT
	PoolRN revnum.ADT
	ProcRN revnum.ADT
}

type service struct {
	procs    execRepo
	operator db.Operator
	log      *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	procs execRepo,
	operator db.Operator,
	l *slog.Logger,
) *service {
	return &service{procs, operator, l}
}

func (s *service) Run(spec ExecSpec) (err error) {
	idAttr := slog.Any("procID", spec.ExecID)
	s.log.Debug("creation started", idAttr)
	ctx := context.Background()
	var mainCfg MainCfg
	err = s.operator.Implicit(ctx, func(ds db.Source) error {
		mainCfg, err = s.procs.SelectMain(ds, spec.ExecID)
		return err
	})
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return err
	}
	var mainEnv Env
	err = s.checkType(spec.PoolID, mainEnv, mainCfg, spec.ProcES)
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return err
	}
	mainMod, err := s.createWith(mainEnv, mainCfg, spec.ProcES)
	if err != nil {
		s.log.Error("creation failed", idAttr)
		return err
	}
	err = s.operator.Explicit(ctx, func(ds db.Source) error {
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

func (s *service) Retrieve(procID identity.ADT) (_ ExecSnap, err error) {
	return ExecSnap{}, nil
}

func (s *service) checkType(
	poolID identity.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	termSpec procexp.ExpSpec,
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
	poolID identity.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	ts procexp.ExpSpec,
) error {
	return nil
}

func (s *service) checkClient(
	poolID identity.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	ts procexp.ExpSpec,
) error {
	return nil
}

func (s *service) createWith(
	mainEnv Env,
	procCfg MainCfg,
	ts procexp.ExpSpec,
) (
	procMod MainMod,
	_ error,
) {
	switch termSpec := ts.(type) {
	case procexp.CallSpec:
		viaCord, ok := procCfg.Bnds[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.CommPH)
			s.log.Error("coordination failed")
			return MainMod{}, err
		}
		viaAttr := slog.Any("cordID", viaCord.CordID)
		for _, valPH := range termSpec.ValPHs {
			sndrValBnd := Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: valPH,
				ProcRN: -procCfg.ExecRN.Next(),
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
	case procexp.SpawnSpec:
		s.log.Debug("coordination succeed")
		return procMod, nil
	default:
		panic(procexp.ErrExpTypeUnexpected(ts))
	}
}

func ErrMissingChnl(want qualsym.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func errMissingPool(want qualsym.ADT) error {
	return fmt.Errorf("pool missing in env: %v", want)
}

func errMissingSig(want identity.ADT) error {
	return fmt.Errorf("sig missing in env: %v", want)
}

func errMissingRole(want qualsym.ADT) error {
	return fmt.Errorf("role missing in env: %v", want)
}

func ErrRootTypeUnexpected(got SemRec) error {
	return fmt.Errorf("sem rec unexpected: %T", got)
}

func ErrRootTypeMismatch(got, want SemRec) error {
	return fmt.Errorf("sem rec mismatch: want %T, got %T", want, got)
}
