package procexec

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"reflect"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/polarity"
	"orglang/go-runtime/adt/procdec"
	"orglang/go-runtime/adt/procdef"
	"orglang/go-runtime/adt/procexp"
	"orglang/go-runtime/adt/procstep"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/typedef"
	"orglang/go-runtime/adt/typeexp"
	"orglang/go-runtime/adt/uniqsym"
)

type API interface {
	Run(ExecSpec) error
	Take(procstep.StepSpec) error
	Retrieve(identity.ADT) (ExecSnap, error)
}

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
	Bnds   map[symbol.ADT]EP2
	Acts   map[identity.ADT]procstep.StepRec
	PoolID identity.ADT
	ExecRN revnum.ADT
}

// aka Configuration
type Cfg struct {
	ExecID identity.ADT
	Chnls  map[symbol.ADT]EP
	Steps  map[identity.ADT]procstep.StepRec
	PoolID identity.ADT
	PoolRN revnum.ADT
	ExecRN revnum.ADT
}

type Env struct {
	ProcDecs map[identity.ADT]procdec.DecRec
	TypeDefs map[uniqsym.ADT]typedef.DefRec
	TypeExps map[identity.ADT]typeexp.ExpRec
	Locks    map[uniqsym.ADT]Lock
}

type EP struct {
	ChnlPH symbol.ADT
	ChnlID identity.ADT
	ExpID  identity.ADT
	// provider
	PoolID identity.ADT
}

type EP2 struct {
	ChnlPH symbol.ADT
	ChnlID identity.ADT
	// provider
	PoolID identity.ADT
}

type Lock struct {
	PoolID identity.ADT
	PoolRN revnum.ADT
}

func ChnlPH(rec EP) symbol.ADT { return rec.ChnlPH }

// ответственность за процесс
type Liab struct {
	PoolID identity.ADT
	ExecID identity.ADT
	// позитивное значение при вручении
	// негативное значение при лишении
	PoolRN revnum.ADT
}

type Mod struct {
	Locks []Lock
	Bnds  []Bnd
	Steps []procstep.StepRec
	Liabs []Liab
}

type MainMod struct {
	Bnds []Bnd
	Acts []procstep.StepRec
}

type Bnd struct {
	ExecID identity.ADT
	ChnlPH symbol.ADT
	ChnlID identity.ADT
	ExpID  identity.ADT
	PoolRN revnum.ADT
	ProcRN revnum.ADT
}

type service struct {
	procExecs Repo
	procDecs  procdec.Repo
	typeDefs  typedef.Repo
	typeExps  typeexp.Repo
	operator  db.Operator
	log       *slog.Logger
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

func newService(
	procExecs Repo,
	procDecs procdec.Repo,
	typeDefs typedef.Repo,
	typeExps typeexp.Repo,
	operator db.Operator,
	l *slog.Logger,
) *service {
	name := slog.String("name", reflect.TypeFor[service]().Name())
	return &service{procExecs, procDecs, typeDefs, typeExps, operator, l.With(name)}
}

func (s *service) Run(spec ExecSpec) (err error) {
	idAttr := slog.Any("execID", spec.ExecID)
	s.log.Debug("creation started", idAttr)
	ctx := context.Background()
	var mainCfg MainCfg
	err = s.operator.Implicit(ctx, func(ds db.Source) error {
		mainCfg, err = s.procExecs.SelectMain(ds, spec.ExecID)
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
		err = s.procExecs.UpdateMain(ds, mainMod)
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

func (s *service) Retrieve(execID identity.ADT) (_ ExecSnap, err error) {
	return ExecSnap{}, nil
}

func (s *service) checkType(
	poolID identity.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	expSpec procexp.ExpSpec,
) error {
	chnlEP, ok := mainCfg.Bnds[expSpec.Via()]
	if !ok {
		panic("no via in main cfg")
	}
	if poolID == chnlEP.PoolID {
		return s.checkProviderMain(poolID, mainEnv, mainCfg, expSpec)
	} else {
		return s.checkClientMain(poolID, mainEnv, mainCfg, expSpec)
	}
}

func (s *service) checkProviderMain(
	poolID identity.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	es procexp.ExpSpec,
) error {
	return nil
}

func (s *service) checkClientMain(
	poolID identity.ADT,
	mainEnv Env,
	mainCfg MainCfg,
	es procexp.ExpSpec,
) error {
	return nil
}

func (s *service) createWith(
	mainEnv Env,
	procCfg MainCfg,
	es procexp.ExpSpec,
) (
	procMod MainMod,
	_ error,
) {
	switch expSpec := es.(type) {
	case procexp.CallSpec:
		commChnlEP, ok := procCfg.Bnds[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("coordination failed")
			return MainMod{}, err
		}
		viaAttr := slog.Any("cordID", commChnlEP.ChnlID)
		for _, valChnlPH := range expSpec.ValChnlPHs {
			sndrValBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: valChnlPH,
				ProcRN: -procCfg.ExecRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrValBnd)
		}
		rcvrAct := procCfg.Acts[commChnlEP.ChnlID]
		if rcvrAct == nil {
			sndrAct := procstep.MsgRec{}
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
		panic(procexp.ErrExpTypeUnexpected(es))
	}
}

func ErrMissingChnl(want symbol.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func (s *service) Take(spec procstep.StepSpec) (err error) {
	idAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("taking started", idAttr)
	ctx := context.Background()
	// initial values
	poolID := spec.ExecID
	procID := spec.ProcID
	expSpec := spec.ProcES
	for expSpec != nil {
		var procCfg Cfg
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			procCfg, err = s.procExecs.SelectProc(ds, procID)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		if len(procCfg.Chnls) == 0 {
			panic("zero channels")
		}
		decIDs := procexp.CollectEnv(expSpec)
		var procDecs map[identity.ADT]procdec.DecRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			procDecs, err = s.procDecs.SelectEnv(ds, decIDs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("decs", decIDs))
			return err
		}
		typeQNs := procdec.CollectEnv(maps.Values(procDecs))
		var typeDefs map[uniqsym.ADT]typedef.DefRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			typeDefs, err = s.typeDefs.SelectEnv(ds, typeQNs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("types", typeQNs))
			return err
		}
		envIDs := typedef.CollectEnv(maps.Values(typeDefs))
		ctxIDs := CollectCtx(maps.Values(procCfg.Chnls))
		var typeExps map[identity.ADT]typeexp.ExpRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			typeExps, err = s.typeExps.SelectEnv(ds, append(envIDs, ctxIDs...))
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("env", envIDs), slog.Any("ctx", ctxIDs))
			return err
		}
		procEnv := Env{ProcDecs: procDecs, TypeDefs: typeDefs, TypeExps: typeExps}
		procCtx := convertToCtx(poolID, maps.Values(procCfg.Chnls), typeExps)
		// type checking
		err = s.checkState(poolID, procEnv, procCtx, procCfg, expSpec)
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		// step taking
		nextSpec, procMod, err := s.takeWith(procEnv, procCfg, expSpec)
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		err = s.operator.Explicit(ctx, func(ds db.Source) error {
			err = s.procExecs.UpdateProc(ds, procMod)
			if err != nil {
				s.log.Error("taking failed", idAttr)
				return err
			}
			return nil
		})
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		// next values
		poolID = nextSpec.ExecID
		procID = nextSpec.ProcID
		expSpec = nextSpec.ProcES
	}
	s.log.Debug("taking succeed", idAttr)
	return nil
}

func (s *service) takeWith(
	procEnv Env,
	procCfg Cfg,
	es procexp.ExpSpec,
) (
	stepSpec procstep.StepSpec,
	procMod Mod,
	_ error,
) {
	switch expSpec := es.(type) {
	case procexp.CloseSpec:
		commChnlEP, ok := procCfg.Chnls[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlEP.ChnlID)
		sndrLock := Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		rcvrStep := procCfg.Steps[commChnlEP.ChnlID]
		if rcvrStep == nil {
			sndrStep := procstep.MsgRec{
				PoolID: procCfg.PoolID,
				ExecID: procCfg.ExecID,
				ChnlID: commChnlEP.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				ValER: procexp.CloseRec{
					CommChnlPH: expSpec.CommChnlPH,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procstep.SvcRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.ContER.(type) {
		case procexp.WaitRec:
			sndrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := Bnd{
				ExecID: svcStep.ExecID,
				ChnlPH: termImpl.CommChnlPH,
				PoolRN: -svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			stepSpec = procstep.StepSpec{
				ExecID: svcStep.PoolID,
				ProcID: svcStep.ExecID,
				ProcES: termImpl.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(svcStep.ContER))
		}
	case procexp.WaitSpec:
		commChnlEP, ok := procCfg.Chnls[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlEP.ChnlID)
		rcvrLock := Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[commChnlEP.ChnlID]
		if sndrStep == nil {
			rcvrStep := procstep.SvcRec{
				PoolID: procCfg.PoolID,
				ExecID: procCfg.ExecID,
				ChnlID: commChnlEP.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				ContER: procexp.WaitRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContES:     expSpec.ContES,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(procstep.MsgRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.ValER.(type) {
		case procexp.CloseRec:
			sndrViaBnd := Bnd{
				ExecID: msgStep.ExecID,
				ChnlPH: termImpl.CommChnlPH,
				PoolRN: -msgStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			stepSpec = procstep.StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: expSpec.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, procMod, nil
		case procexp.FwdRec:
			rcvrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: termImpl.ContChnlID,
				ExpID:  commChnlEP.ExpID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			stepSpec = procstep.StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: expSpec,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(msgStep.ValER))
		}
	case procexp.SendSpec:
		commChnlEP, ok := procCfg.Chnls[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlEP.ChnlID)
		sndrLock := Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.TypeExps[commChnlEP.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(commChnlEP.ExpID)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, Mod{}, err
		}
		viaStateID := viaState.(typeexp.ProdRec).Next()
		valChnl, ok := procCfg.Chnls[expSpec.ValChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.ValChnlPH)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, Mod{}, err
		}
		sndrValBnd := Bnd{
			ExecID: procCfg.ExecID,
			ChnlPH: expSpec.ValChnlPH,
			PoolRN: -procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrValBnd)
		rcvrStep := procCfg.Steps[commChnlEP.ChnlID]
		if rcvrStep == nil {
			newChnlID := identity.New()
			sndrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: newChnlID,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procstep.MsgRec{
				PoolID: procCfg.PoolID,
				ExecID: procCfg.ExecID,
				ChnlID: commChnlEP.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				ValER: procexp.SendRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: newChnlID,
					ValChnlID:  valChnl.ChnlID,
					ValExpID:   valChnl.ExpID,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procstep.SvcRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.ContER.(type) {
		case procexp.RecvRec:
			sndrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: termImpl.ContChnlID,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := Bnd{
				ExecID: svcStep.ExecID,
				ChnlPH: termImpl.CommChnlPH,
				ChnlID: termImpl.ContChnlID,
				ExpID:  viaStateID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := Bnd{
				ExecID: svcStep.ExecID,
				ChnlPH: termImpl.ValChnlPH,
				ChnlID: valChnl.ChnlID,
				ExpID:  valChnl.ExpID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
			stepSpec = procstep.StepSpec{
				ExecID: svcStep.PoolID,
				ProcID: svcStep.ExecID,
				ProcES: termImpl.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(svcStep.ContER))
		}
	case procexp.RecvSpec:
		commChnlEP, ok := procCfg.Chnls[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlEP.ChnlID)
		rcvrLock := Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrSemRec := procCfg.Steps[commChnlEP.ChnlID]
		if sndrSemRec == nil {
			rcvrSemRec := procstep.SvcRec{
				PoolID: procCfg.PoolID,
				ExecID: procCfg.ExecID,
				ChnlID: commChnlEP.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				ContER: procexp.RecvRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: identity.New(),
					ValChnlPH:  expSpec.BindChnlPH,
					ContES:     expSpec.ContES,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrSemRec)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, procMod, nil
		}
		sndrMsgRec, ok := sndrSemRec.(procstep.MsgRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(sndrSemRec))
		}
		switch termRec := sndrMsgRec.ValER.(type) {
		case procexp.SendRec:
			viaState, ok := procEnv.TypeExps[commChnlEP.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(commChnlEP.ExpID)
				s.log.Error("taking failed", viaAttr)
				return procstep.StepSpec{}, Mod{}, err
			}
			rcvrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: termRec.ContChnlID,
				ExpID:  viaState.(typeexp.ProdRec).Next(),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.BindChnlPH,
				ChnlID: termRec.ValChnlID,
				ExpID:  termRec.ValExpID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
			stepSpec = procstep.StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: expSpec.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(sndrMsgRec.ValER))
		}
	case procexp.LabSpec:
		commChnlEP, ok := procCfg.Chnls[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlEP.ChnlID)
		sndrLock := Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.TypeExps[commChnlEP.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(commChnlEP.ExpID)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, Mod{}, err
		}
		viaStateID := viaState.(typeexp.SumRec).Next(expSpec.LabelQN)
		rcvrStep := procCfg.Steps[commChnlEP.ChnlID]
		if rcvrStep == nil {
			newViaID := identity.New()
			sndrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: newViaID,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procstep.MsgRec{
				PoolID: procCfg.PoolID,
				ExecID: procCfg.ExecID,
				ChnlID: commChnlEP.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				ValER: procexp.LabRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: newViaID,
					LabelQN:    expSpec.LabelQN,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procstep.SvcRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.ContER.(type) {
		case procexp.CaseRec:
			sndrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: termImpl.ContChnlID,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := Bnd{
				ExecID: svcStep.ExecID,
				ChnlPH: termImpl.CommChnlPH,
				ChnlID: termImpl.ContChnlID,
				ExpID:  viaStateID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			stepSpec = procstep.StepSpec{
				ExecID: svcStep.PoolID,
				ProcID: svcStep.ExecID,
				ProcES: termImpl.ContESs[expSpec.LabelQN],
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(svcStep.ContER))
		}
	case procexp.CaseSpec:
		commChnlEP, ok := procCfg.Chnls[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlEP.ChnlID)
		rcvrLock := Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[commChnlEP.ChnlID]
		if sndrStep == nil {
			rcvrStep := procstep.SvcRec{
				PoolID: procCfg.PoolID,
				ExecID: procCfg.ExecID,
				ChnlID: commChnlEP.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				ContER: procexp.CaseRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: identity.New(),
					ContESs:    expSpec.ContESs,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(procstep.MsgRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.ValER.(type) {
		case procexp.LabRec:
			viaState, ok := procEnv.TypeExps[commChnlEP.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(commChnlEP.ExpID)
				s.log.Error("taking failed", viaAttr)
				return procstep.StepSpec{}, Mod{}, err
			}
			rcvrViaBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: termImpl.ContChnlID,
				ExpID:  viaState.(typeexp.SumRec).Next(termImpl.LabelQN),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			stepSpec = procstep.StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: expSpec.ContESs[termImpl.LabelQN],
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(msgStep.ValER))
		}
	case procexp.SpawnSpecOld:
		rcvrSnap, ok := procEnv.Locks[expSpec.PoolQN]
		if !ok {
			err := errMissingPool(expSpec.PoolQN)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		rcvrLiab := Liab{
			ExecID: identity.New(),
			PoolID: rcvrSnap.PoolID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Liabs = append(procMod.Liabs, rcvrLiab)
		rcvrProcDec, ok := procEnv.ProcDecs[expSpec.SigID]
		if !ok {
			err := errMissingSig(expSpec.SigID)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		rcvrTypeDef, ok := procEnv.TypeDefs[rcvrProcDec.X.TypeQN]
		if !ok {
			err := errMissingRole(rcvrProcDec.X.TypeQN)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		contChnlID := identity.New()
		sndrViaBnd := Bnd{
			ExecID: procCfg.ExecID,
			ChnlPH: expSpec.X,
			ChnlID: contChnlID,
			ExpID:  rcvrTypeDef.ExpID,
			PoolRN: procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
		rcvrViaBnd := Bnd{
			ExecID: rcvrLiab.ExecID,
			ChnlPH: rcvrProcDec.X.BindPH,
			ChnlID: contChnlID,
			ExpID:  rcvrTypeDef.ExpID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
		for i, valChnlPH := range expSpec.Ys {
			valChnlEP, ok := procCfg.Chnls[valChnlPH]
			if !ok {
				err := ErrMissingChnl(valChnlPH)
				s.log.Error("taking failed")
				return procstep.StepSpec{}, Mod{}, err
			}
			sndrValBnd := Bnd{
				ExecID: procCfg.ExecID,
				ChnlPH: valChnlPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrValBnd)
			rcvrValBnd := Bnd{
				ExecID: rcvrLiab.ExecID,
				ChnlPH: rcvrProcDec.Ys[i].BindPH,
				ChnlID: valChnlEP.ChnlID,
				ExpID:  valChnlEP.ExpID,
				PoolRN: rcvrSnap.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
		}
		stepSpec = procstep.StepSpec{
			ExecID: procCfg.PoolID,
			ProcID: procCfg.ExecID,
			ProcES: expSpec.ContES,
		}
		s.log.Debug("taking succeed")
		return stepSpec, procMod, nil
	case procexp.FwdSpec:
		commChnlEP, ok := procCfg.Chnls[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlEP.ChnlID)
		commChnlER, ok := procEnv.TypeExps[commChnlEP.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(commChnlEP.ExpID)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, Mod{}, err
		}
		valChnlEP, ok := procCfg.Chnls[expSpec.ContChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.ContChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, Mod{}, err
		}
		commChnlSR := procCfg.Steps[commChnlEP.ChnlID]
		switch commChnlER.Pol() {
		case polarity.Pos:
			switch stepRec := commChnlSR.(type) {
			case procstep.SvcRec:
				xBnd := Bnd{
					ExecID: stepRec.ExecID,
					ChnlPH: stepRec.ContER.Via(),
					ChnlID: commChnlEP.ChnlID,
					ExpID:  commChnlEP.ExpID,
					PoolRN: stepRec.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				stepSpec = procstep.StepSpec{
					ExecID: stepRec.PoolID,
					ProcID: stepRec.ExecID,
					ProcES: stepRec.ContER,
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, procMod, nil
			case procstep.MsgRec:
				yBnd := Bnd{
					ExecID: stepRec.ExecID,
					ChnlPH: stepRec.ValER.Via(),
					ChnlID: valChnlEP.ChnlID,
					ExpID:  valChnlEP.ExpID,
					PoolRN: stepRec.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				stepSpec = procstep.StepSpec{
					ExecID: stepRec.PoolID,
					ProcID: stepRec.ExecID,
					ProcES: stepRec.ValER,
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, procMod, nil
			case nil:
				xBnd := Bnd{
					ExecID: procCfg.ExecID,
					ChnlPH: expSpec.CommChnlPH,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				yBnd := Bnd{
					ExecID: procCfg.ExecID,
					ChnlPH: expSpec.ContChnlPH,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				msgStep := procstep.MsgRec{
					PoolID: procCfg.PoolID,
					ExecID: procCfg.ExecID,
					ChnlID: commChnlEP.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					ValER: procexp.FwdRec{
						ContChnlID: valChnlEP.ChnlID,
					},
				}
				procMod.Steps = append(procMod.Steps, msgStep)
				s.log.Debug("taking half done", viaAttr)
				return stepSpec, procMod, nil
			default:
				panic(procstep.ErrRecTypeUnexpected(commChnlSR))
			}
		case polarity.Neg:
			switch stepRec := commChnlSR.(type) {
			case procstep.SvcRec:
				yBnd := Bnd{
					ExecID: stepRec.ExecID,
					ChnlPH: stepRec.ContER.Via(),
					ChnlID: valChnlEP.ChnlID,
					ExpID:  valChnlEP.ExpID,
					PoolRN: stepRec.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				stepSpec = procstep.StepSpec{
					ExecID: stepRec.PoolID,
					ProcID: stepRec.ExecID,
					ProcES: stepRec.ContER,
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, procMod, nil
			case procstep.MsgRec:
				xBnd := Bnd{
					ExecID: stepRec.ExecID,
					ChnlPH: stepRec.ValER.Via(),
					ChnlID: commChnlEP.ChnlID,
					ExpID:  commChnlEP.ExpID,
					PoolRN: stepRec.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				stepSpec = procstep.StepSpec{
					ExecID: stepRec.PoolID,
					ProcID: stepRec.ExecID,
					ProcES: stepRec.ValER,
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, procMod, nil
			case nil:
				svcStep := procstep.SvcRec{
					PoolID: procCfg.PoolID,
					ExecID: procCfg.ExecID,
					ChnlID: commChnlEP.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					ContER: procexp.FwdRec{
						ContChnlID: valChnlEP.ChnlID,
					},
				}
				procMod.Steps = append(procMod.Steps, svcStep)
				s.log.Debug("taking half done", viaAttr)
				return stepSpec, procMod, nil
			default:
				panic(procstep.ErrRecTypeUnexpected(commChnlSR))
			}
		default:
			panic(typeexp.ErrPolarityUnexpected(commChnlER))
		}
	default:
		panic(procexp.ErrExpTypeUnexpected(es))
	}
}

func CollectCtx(chnls iter.Seq[EP]) []identity.ADT {
	return nil
}

func convertToCtx(poolID identity.ADT, chnlEPs iter.Seq[EP], typeExps map[identity.ADT]typeexp.ExpRec) typedef.Context {
	assets := make(map[symbol.ADT]typeexp.ExpRec, 1)
	liabs := make(map[symbol.ADT]typeexp.ExpRec, 1)
	for ep := range chnlEPs {
		if poolID == ep.PoolID {
			liabs[ep.ChnlPH] = typeExps[ep.ExpID]
		} else {
			assets[ep.ChnlPH] = typeExps[ep.ExpID]
		}
	}
	return typedef.Context{Assets: assets, Liabs: liabs}
}

func (s *service) checkState(
	poolID identity.ADT,
	procEnv Env,
	procCtx typedef.Context,
	procCfg Cfg,
	expSpec procexp.ExpSpec,
) error {
	chnlEP, ok := procCfg.Chnls[expSpec.Via()]
	if !ok {
		panic("no via in proc snap")
	}
	if poolID == chnlEP.PoolID {
		return s.checkProvider(poolID, procEnv, procCtx, procCfg, expSpec)
	} else {
		return s.checkClient(poolID, procEnv, procCtx, procCfg, expSpec)
	}
}

func (s *service) checkProvider(
	poolID identity.ADT,
	procEnv Env,
	procCtx typedef.Context,
	procCfg Cfg,
	es procexp.ExpSpec,
) error {
	switch expSpec := es.(type) {
	case procexp.CloseSpec:
		// check ctx
		if len(procCtx.Assets) > 0 {
			err := fmt.Errorf("context mismatch: want 0 items, got %v items", len(procCtx.Assets))
			s.log.Error("checking failed")
			return err
		}
		// check via
		gotVia, ok := procCtx.Liabs[expSpec.CommChnlPH]
		if !ok {
			err := typedef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVia, typeexp.OneRec{})
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		delete(procCtx.Liabs, expSpec.CommChnlPH)
		return nil
	case procexp.WaitSpec:
		err := procexp.ErrExpTypeMismatch(es, procexp.CloseSpec{})
		s.log.Error("checking failed")
		return err
	case procexp.SendSpec:
		// check via
		gotVia, ok := procCtx.Liabs[expSpec.CommChnlPH]
		if !ok {
			err := typedef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.TensorRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[expSpec.ValChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.ValChnlPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[expSpec.CommChnlPH] = wantVia.Z
		delete(procCtx.Assets, expSpec.ValChnlPH)
		return nil
	case procexp.RecvSpec:
		// check via
		gotVia, ok := procCtx.Liabs[expSpec.CommChnlPH]
		if !ok {
			err := typedef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.LolliRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[expSpec.BindChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.BindChnlPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Liabs[expSpec.CommChnlPH] = wantVia.Z
		procCtx.Assets[expSpec.BindChnlPH] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, expSpec.ContES)
	case procexp.LabSpec:
		// check via
		gotVia, ok := procCtx.Liabs[expSpec.CommChnlPH]
		if !ok {
			err := typedef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.PlusRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check label
		choice, ok := wantVia.Zs[expSpec.LabelQN]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Zs), expSpec.LabelQN)
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[expSpec.CommChnlPH] = choice
		return nil
	case procexp.CaseSpec:
		// check via
		gotVia, ok := procCtx.Liabs[expSpec.CommChnlPH]
		if !ok {
			err := typedef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.WithRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check conts
		if len(expSpec.ContESs) != len(wantVia.Zs) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Zs), len(expSpec.ContESs))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Zs {
			cont, ok := expSpec.ContESs[label]
			if !ok {
				err := fmt.Errorf("label mismatch: want %q, got nothing", label)
				s.log.Error("checking failed")
				return err
			}
			procCtx.Liabs[expSpec.CommChnlPH] = choice
			err := s.checkState(poolID, procEnv, procCtx, procCfg, cont)
			if err != nil {
				s.log.Error("checking failed")
				return err
			}
		}
		return nil
	case procexp.FwdSpec:
		if len(procCtx.Assets) != 1 {
			err := fmt.Errorf("context mismatch: want 1 item, got %v items", len(procCtx.Assets))
			s.log.Error("checking failed")
			return err
		}
		viaSt, ok := procCtx.Liabs[expSpec.CommChnlPH]
		if !ok {
			err := typedef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		fwdSt, ok := procCtx.Assets[expSpec.ContChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.ContChnlPH)
			s.log.Error("checking failed")
			return err
		}
		if fwdSt.Pol() != viaSt.Pol() {
			err := typeexp.ErrPolarityMismatch(fwdSt, viaSt)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(fwdSt, viaSt)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		delete(procCtx.Liabs, expSpec.CommChnlPH)
		delete(procCtx.Assets, expSpec.ContChnlPH)
		return nil
	default:
		panic(procexp.ErrExpTypeUnexpected(es))
	}
}

func (s *service) checkClient(
	poolID identity.ADT,
	procEnv Env,
	procCtx typedef.Context,
	procCfg Cfg,
	es procexp.ExpSpec,
) error {
	switch expSpec := es.(type) {
	case procexp.CloseSpec:
		err := procexp.ErrExpTypeMismatch(es, procexp.WaitSpec{})
		s.log.Error("checking failed")
		return err
	case procexp.WaitSpec:
		// check via
		gotVia, ok := procCtx.Assets[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.OneRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check cont
		delete(procCtx.Assets, expSpec.CommChnlPH)
		return s.checkState(poolID, procEnv, procCtx, procCfg, expSpec.ContES)
	case procexp.SendSpec:
		// check via
		gotVia, ok := procCtx.Assets[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.LolliRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[expSpec.ValChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.ValChnlPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[expSpec.CommChnlPH] = wantVia.Z
		delete(procCtx.Assets, expSpec.ValChnlPH)
		return nil
	case procexp.RecvSpec:
		// check via
		gotVia, ok := procCtx.Assets[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.TensorRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[expSpec.BindChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.BindChnlPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[expSpec.CommChnlPH] = wantVia.Z
		procCtx.Assets[expSpec.BindChnlPH] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, expSpec.ContES)
	case procexp.LabSpec:
		// check via
		gotVia, ok := procCtx.Assets[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.WithRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check label
		choice, ok := wantVia.Zs[expSpec.LabelQN]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Zs), expSpec.LabelQN)
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[expSpec.CommChnlPH] = choice
		return nil
	case procexp.CaseSpec:
		// check via
		gotVia, ok := procCtx.Assets[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCtx(expSpec.CommChnlPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typeexp.PlusRec)
		if !ok {
			err := typeexp.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check conts
		if len(expSpec.ContESs) != len(wantVia.Zs) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Zs), len(expSpec.ContESs))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Zs {
			cont, ok := expSpec.ContESs[label]
			if !ok {
				err := fmt.Errorf("label mismatch: want %q, got nothing", label)
				s.log.Error("checking failed")
				return err
			}
			procCtx.Assets[expSpec.CommChnlPH] = choice
			err := s.checkState(poolID, procEnv, procCtx, procCfg, cont)
			if err != nil {
				s.log.Error("checking failed")
				return err
			}
		}
		return nil
	case procexp.SpawnSpecOld:
		procSig, ok := procEnv.ProcDecs[expSpec.SigID]
		if !ok {
			err := procdec.ErrRootMissingInEnv(expSpec.SigID)
			s.log.Error("checking failed")
			return err
		}
		// check vals
		if len(expSpec.Ys) != len(procSig.Ys) {
			err := fmt.Errorf("context mismatch: want %v items, got %v items", len(procSig.Ys), len(expSpec.Ys))
			s.log.Error("checking failed", slog.Any("want", procSig.Ys), slog.Any("got", expSpec.Ys))
			return err
		}
		if len(expSpec.Ys) == 0 {
			return nil
		}
		for i, ep := range procSig.Ys {
			valRole, ok := procEnv.TypeDefs[ep.TypeQN]
			if !ok {
				err := typedef.ErrSymMissingInEnv(ep.TypeQN)
				s.log.Error("checking failed")
				return err
			}
			wantVal, ok := procEnv.TypeExps[valRole.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(valRole.ExpID)
				s.log.Error("checking failed")
				return err
			}
			gotVal, ok := procCtx.Assets[expSpec.Ys[i]]
			if !ok {
				err := procdef.ErrMissingInCtx(ep.BindPH)
				s.log.Error("checking failed")
				return err
			}
			err := typeexp.CheckRec(gotVal, wantVal)
			if err != nil {
				s.log.Error("checking failed", slog.Any("want", wantVal), slog.Any("got", gotVal))
				return err
			}
			delete(procCtx.Assets, expSpec.Ys[i])
		}
		// check via
		viaRole, ok := procEnv.TypeDefs[procSig.X.TypeQN]
		if !ok {
			err := typedef.ErrSymMissingInEnv(procSig.X.TypeQN)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := procEnv.TypeExps[viaRole.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaRole.ExpID)
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[expSpec.X] = wantVia
		return s.checkState(poolID, procEnv, procCtx, procCfg, expSpec.ContES)
	default:
		panic(procexp.ErrExpTypeUnexpected(es))
	}
}

func errOptimisticUpdate(got revnum.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}

func errMissingPool(want uniqsym.ADT) error {
	return fmt.Errorf("pool missing in env: %v", want)
}

func errMissingSig(want identity.ADT) error {
	return fmt.Errorf("sig missing in env: %v", want)
}

func errMissingRole(want uniqsym.ADT) error {
	return fmt.Errorf("role missing in env: %v", want)
}
