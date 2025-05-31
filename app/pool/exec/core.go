package exec

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/exp/maps"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/pol"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/app/pool/def"
	procdec "smecalculus/rolevod/app/proc/dec"
	procdef "smecalculus/rolevod/app/proc/def"
	procexec "smecalculus/rolevod/app/proc/exec"
	typedef "smecalculus/rolevod/app/type/def"
)

type PoolSpec struct {
	PoolQN sym.ADT
	SupID  id.ADT
}

type PoolRef struct {
	ExecID id.ADT
	ProcID id.ADT // main
}

type PoolRec struct {
	ExecID id.ADT
	ProcID id.ADT // main
	SupID  id.ADT
	PoolRN rn.ADT
}

type PoolSnap struct {
	ExecID id.ADT
	Title  string
	Subs   []PoolRef
}

type StepSpec struct {
	PoolID id.ADT
	ProcID id.ADT
	ProcTS procdef.TermSpec
}

type PollSpec struct {
	PoolID id.ADT
	PoolTS def.TermSpec
}

// Port
type API interface {
	Create(PoolSpec) (PoolRef, error)
	Retrieve(id.ADT) (PoolSnap, error)
	RetreiveRefs() ([]PoolRef, error)
	Spawn(procexec.ProcSpec) (procexec.ProcRef, error)
	Take(StepSpec) error
	Poll(PollSpec) (procexec.ProcRef, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	pools    Repo
	procs    procdec.Repo
	types    typedef.Repo
	operator data.Operator
	log      *slog.Logger
}

func newService(
	pools Repo,
	procs procdec.Repo,
	types typedef.Repo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	return &service{pools, procs, types, operator, l}
}

func (s *service) Create(spec PoolSpec) (PoolRef, error) {
	ctx := context.Background()
	s.log.Debug("creation started", slog.Any("spec", spec))
	impl := PoolRec{
		ExecID: id.New(),
		ProcID: id.New(),
		SupID:  spec.SupID,
		PoolRN: rn.Initial(),
	}
	liab := procexec.Liab{
		PoolID: impl.ExecID,
		ProcID: impl.ProcID,
		PoolRN: impl.PoolRN,
	}
	err := s.operator.Explicit(ctx, func(ds data.Source) error {
		err := s.pools.Insert(ds, impl)
		if err != nil {
			s.log.Error("creation failed")
			return err
		}
		err = s.pools.InsertLiab(ds, liab)
		if err != nil {
			s.log.Error("creation failed")
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("creation failed")
		return PoolRef{}, err
	}
	s.log.Debug("creation succeeded", slog.Any("poolID", impl.ExecID))
	return ConvertRecToRef(impl), nil
}

func (s *service) Poll(spec PollSpec) (procexec.ProcRef, error) {
	switch spec.PoolTS.(type) {
	default:
		return procexec.ProcRef{}, nil
	}
	return procexec.ProcRef{}, nil
}

func (s *service) Spawn(spec procexec.ProcSpec) (_ procexec.ProcRef, err error) {
	procAttr := slog.Any("procID", spec.ExecID)
	s.log.Debug("spawning started", procAttr)
	return procexec.ProcRef{}, nil
}

func (s *service) Take(spec StepSpec) (err error) {
	idAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("taking started", idAttr)
	ctx := context.Background()
	// initial values
	poolID := spec.PoolID
	procID := spec.ProcID
	termSpec := spec.ProcTS
	for termSpec != nil {
		var procCfg procexec.Cfg
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			procCfg, err = s.pools.SelectProc(ds, procID)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		if len(procCfg.Chnls) == 0 {
			panic("zero channels")
		}
		sigIDs := procdef.CollectEnv(termSpec)
		var sigs map[id.ADT]procdec.ProcRec
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			sigs, err = s.procs.SelectEnv(ds, sigIDs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("sigs", sigIDs))
			return err
		}
		typeQNs := procdec.CollectEnv(maps.Values(sigs))
		var types map[sym.ADT]typedef.TypeRec
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			types, err = s.types.SelectTypeEnv(ds, typeQNs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("types", typeQNs))
			return err
		}
		envIDs := typedef.CollectEnv(maps.Values(types))
		ctxIDs := CollectCtx(maps.Values(procCfg.Chnls))
		var terms map[id.ADT]typedef.TermRec
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			terms, err = s.types.SelectTermEnv(ds, append(envIDs, ctxIDs...))
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("env", envIDs), slog.Any("ctx", ctxIDs))
			return err
		}
		procEnv := procexec.Env{ProcSigs: sigs, Types: types, TypeTerms: terms}
		procCtx := convertToCtx(poolID, maps.Values(procCfg.Chnls), terms)
		// type checking
		err = s.checkState(poolID, procEnv, procCtx, procCfg, termSpec)
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		// step taking
		nextSpec, procMod, err := s.takeWith(procEnv, procCfg, termSpec)
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		err = s.operator.Explicit(ctx, func(ds data.Source) error {
			err = s.pools.UpdateProc(ds, procMod)
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
		poolID = nextSpec.PoolID
		procID = nextSpec.ProcID
		termSpec = nextSpec.ProcTS
	}
	s.log.Debug("taking succeeded", idAttr)
	return nil
}

func (s *service) takeWith(
	procEnv procexec.Env,
	procCfg procexec.Cfg,
	ts procdef.TermSpec,
) (
	tranSpec StepSpec,
	procMod procexec.Mod,
	_ error,
) {
	switch termSpec := ts.(type) {
	case procdef.CloseSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.CommPH)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := procexec.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			sndrStep := procexec.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procdef.CloseRec{
					X: termSpec.CommPH,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procexec.SvcRec)
		if !ok {
			panic(procexec.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case procdef.WaitRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procexec.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.X,
				PoolRN: -svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: svcStep.PoolID,
				ProcID: svcStep.ProcID,
				ProcTS: termImpl.Cont,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrRecTypeUnexpected(svcStep.Cont))
		}
	case procdef.WaitSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.CommPH)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := procexec.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[viaChnl.ChnlID]
		if sndrStep == nil {
			rcvrStep := procexec.SvcRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procdef.WaitRec{
					X:    termSpec.CommPH,
					Cont: termSpec.ContTS,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(procexec.MsgRec)
		if !ok {
			panic(procexec.ErrRootTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.Val.(type) {
		case procdef.CloseRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: msgStep.ProcID,
				ChnlPH: termImpl.X,
				PoolRN: -msgStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ProcTS: termSpec.ContTS,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		case procdef.FwdRec:
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.B,
				TermID: viaChnl.TermID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ProcTS: termSpec,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrRecTypeUnexpected(msgStep.Val))
		}
	case procdef.SendSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.CommPH)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := procexec.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.TermID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procexec.Mod{}, err
		}
		viaStateID := viaState.(typedef.ProdRec).Next()
		valChnl, ok := procCfg.Chnls[termSpec.ValPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.ValPH)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procexec.Mod{}, err
		}
		sndrValBnd := procexec.Bnd{
			ProcID: procCfg.ProcID,
			ChnlPH: termSpec.ValPH,
			PoolRN: -procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrValBnd)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newChnlID := id.New()
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				ChnlID: newChnlID,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procexec.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procdef.SendRec{
					X:      termSpec.CommPH,
					A:      newChnlID,
					B:      valChnl.ChnlID,
					TermID: valChnl.TermID,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procexec.SvcRec)
		if !ok {
			panic(procexec.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case procdef.RecvRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.A,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procexec.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.X,
				ChnlID: termImpl.A,
				TermID: viaStateID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := procexec.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.Y,
				ChnlID: valChnl.ChnlID,
				TermID: valChnl.TermID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
			tranSpec = StepSpec{
				PoolID: svcStep.PoolID,
				ProcID: svcStep.ProcID,
				ProcTS: termImpl.Cont,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrRecTypeUnexpected(svcStep.Cont))
		}
	case procdef.RecvSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.CommPH)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := procexec.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrSemRec := procCfg.Steps[viaChnl.ChnlID]
		if sndrSemRec == nil {
			rcvrSemRec := procexec.SvcRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procdef.RecvRec{
					X:    termSpec.CommPH,
					A:    id.New(),
					Y:    termSpec.BindPH,
					Cont: termSpec.ContTS,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrSemRec)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		sndrMsgRec, ok := sndrSemRec.(procexec.MsgRec)
		if !ok {
			panic(procexec.ErrRootTypeUnexpected(sndrSemRec))
		}
		switch termRec := sndrMsgRec.Val.(type) {
		case procdef.SendRec:
			viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
			if !ok {
				err := typedef.ErrMissingInEnv(viaChnl.TermID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, procexec.Mod{}, err
			}
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termRec.A,
				TermID: viaState.(typedef.ProdRec).Next(),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.BindPH,
				ChnlID: termRec.B,
				TermID: termRec.TermID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ProcTS: termSpec.ContTS,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrRecTypeUnexpected(sndrMsgRec.Val))
		}
	case procdef.LabSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.CommPH)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := procexec.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.TermID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procexec.Mod{}, err
		}
		viaStateID := viaState.(typedef.SumRec).Next(termSpec.Label)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newViaID := id.New()
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				ChnlID: newViaID,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procexec.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procdef.LabRec{
					X:     termSpec.CommPH,
					A:     newViaID,
					Label: termSpec.Label,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procexec.SvcRec)
		if !ok {
			panic(procexec.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case procdef.CaseRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.A,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procexec.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.X,
				ChnlID: termImpl.A,
				TermID: viaStateID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: svcStep.PoolID,
				ProcID: svcStep.ProcID,
				ProcTS: termImpl.Conts[termSpec.Label],
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrRecTypeUnexpected(svcStep.Cont))
		}
	case procdef.CaseSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.CommPH)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := procexec.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[viaChnl.ChnlID]
		if sndrStep == nil {
			rcvrStep := procexec.SvcRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procdef.CaseRec{
					X:     termSpec.CommPH,
					A:     id.New(),
					Conts: termSpec.Conts,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(procexec.MsgRec)
		if !ok {
			panic(procexec.ErrRootTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.Val.(type) {
		case procdef.LabRec:
			viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
			if !ok {
				err := typedef.ErrMissingInEnv(viaChnl.TermID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, procexec.Mod{}, err
			}
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.A,
				TermID: viaState.(typedef.SumRec).Next(termImpl.Label),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ProcTS: termSpec.Conts[termImpl.Label],
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrRecTypeUnexpected(msgStep.Val))
		}
	case procdef.SpawnSpecOld:
		rcvrSnap, ok := procEnv.Locks[termSpec.PoolQN]
		if !ok {
			err := errMissingPool(termSpec.PoolQN)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		rcvrLiab := procexec.Liab{
			ProcID: id.New(),
			PoolID: rcvrSnap.PoolID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Liabs = append(procMod.Liabs, rcvrLiab)
		rcvrSig, ok := procEnv.ProcSigs[termSpec.SigID]
		if !ok {
			err := errMissingSig(termSpec.SigID)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		rcvrRole, ok := procEnv.Types[rcvrSig.X.TypeQN]
		if !ok {
			err := errMissingRole(rcvrSig.X.TypeQN)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		newViaID := id.New()
		sndrViaBnd := procexec.Bnd{
			ProcID: procCfg.ProcID,
			ChnlPH: termSpec.X,
			ChnlID: newViaID,
			TermID: rcvrRole.TermID,
			PoolRN: procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
		rcvrViaBnd := procexec.Bnd{
			ProcID: rcvrLiab.ProcID,
			ChnlPH: rcvrSig.X.CommPH,
			ChnlID: newViaID,
			TermID: rcvrRole.TermID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
		for i, chnlPH := range termSpec.Ys {
			valChnl, ok := procCfg.Chnls[chnlPH]
			if !ok {
				err := procexec.ErrMissingChnl(chnlPH)
				s.log.Error("taking failed")
				return StepSpec{}, procexec.Mod{}, err
			}
			sndrValBnd := procexec.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: chnlPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrValBnd)
			rcvrValBnd := procexec.Bnd{
				ProcID: rcvrLiab.ProcID,
				ChnlPH: rcvrSig.Ys[i].CommPH,
				ChnlID: valChnl.ChnlID,
				TermID: valChnl.TermID,
				PoolRN: rcvrSnap.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
		}
		tranSpec = StepSpec{
			PoolID: procCfg.PoolID,
			ProcID: procCfg.ProcID,
			ProcTS: termSpec.Cont,
		}
		s.log.Debug("taking succeeded")
		return tranSpec, procMod, nil
	case procdef.FwdSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.TermID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procexec.Mod{}, err
		}
		valChnl, ok := procCfg.Chnls[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.Y)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		vs := procCfg.Steps[viaChnl.ChnlID]
		switch viaState.Pol() {
		case pol.Pos:
			switch viaStep := vs.(type) {
			case procexec.SvcRec:
				xBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Cont.Via(),
					ChnlID: viaChnl.ChnlID,
					TermID: viaChnl.TermID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				tranSpec = StepSpec{
					PoolID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcTS: viaStep.Cont,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case procexec.MsgRec:
				yBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Val.Via(),
					ChnlID: valChnl.ChnlID,
					TermID: valChnl.TermID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				tranSpec = StepSpec{
					PoolID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcTS: viaStep.Val,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case nil:
				xBnd := procexec.Bnd{
					ProcID: procCfg.ProcID,
					ChnlPH: termSpec.X,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				yBnd := procexec.Bnd{
					ProcID: procCfg.ProcID,
					ChnlPH: termSpec.Y,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				msgStep := procexec.MsgRec{
					PoolID: procCfg.PoolID,
					ProcID: procCfg.ProcID,
					ChnlID: viaChnl.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					Val: procdef.FwdRec{
						B: valChnl.ChnlID,
					},
				}
				procMod.Steps = append(procMod.Steps, msgStep)
				s.log.Debug("taking half done", viaAttr)
				return tranSpec, procMod, nil
			default:
				panic(procexec.ErrRootTypeUnexpected(vs))
			}
		case pol.Neg:
			switch viaStep := vs.(type) {
			case procexec.SvcRec:
				yBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Cont.Via(),
					ChnlID: valChnl.ChnlID,
					TermID: valChnl.TermID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				tranSpec = StepSpec{
					PoolID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcTS: viaStep.Cont,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case procexec.MsgRec:
				xBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Val.Via(),
					ChnlID: viaChnl.ChnlID,
					TermID: viaChnl.TermID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				tranSpec = StepSpec{
					PoolID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcTS: viaStep.Val,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case nil:
				svcStep := procexec.SvcRec{
					PoolID: procCfg.PoolID,
					ProcID: procCfg.ProcID,
					ChnlID: viaChnl.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					Cont: procdef.FwdRec{
						B: valChnl.ChnlID,
					},
				}
				procMod.Steps = append(procMod.Steps, svcStep)
				s.log.Debug("taking half done", viaAttr)
				return tranSpec, procMod, nil
			default:
				panic(procexec.ErrRootTypeUnexpected(vs))
			}
		default:
			panic(typedef.ErrPolarityUnexpected(viaState))
		}
	default:
		panic(procdef.ErrTermTypeUnexpected(ts))
	}
}

func (s *service) Retrieve(poolID id.ADT) (snap PoolSnap, err error) {
	ctx := context.Background()
	err = s.operator.Implicit(ctx, func(ds data.Source) error {
		snap, err = s.pools.SelectSubs(ds, poolID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("poolID", poolID))
		return PoolSnap{}, err
	}
	return snap, nil
}

func (s *service) RetreiveRefs() (refs []PoolRef, err error) {
	ctx := context.Background()
	err = s.operator.Implicit(ctx, func(ds data.Source) error {
		refs, err = s.pools.SelectRefs(ds)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed")
		return nil, err
	}
	return refs, nil
}

func CollectCtx(chnls []procexec.EP) []id.ADT {
	return nil
}

func convertToCtx(poolID id.ADT, chnls []procexec.EP, types map[id.ADT]typedef.TermRec) typedef.Context {
	assets := make(map[sym.ADT]typedef.TermRec, len(chnls)-1)
	liabs := make(map[sym.ADT]typedef.TermRec, 1)
	for _, ch := range chnls {
		if poolID == ch.PoolID {
			liabs[ch.ChnlPH] = types[ch.TermID]
		} else {
			assets[ch.ChnlPH] = types[ch.TermID]
		}
	}
	return typedef.Context{Assets: assets, Liabs: liabs}
}

func (s *service) checkState(
	poolID id.ADT,
	procEnv procexec.Env,
	procCtx typedef.Context,
	procCfg procexec.Cfg,
	termSpec procdef.TermSpec,
) error {
	ch, ok := procCfg.Chnls[termSpec.Via()]
	if !ok {
		panic("no via in proc snap")
	}
	if poolID == ch.PoolID {
		return s.checkProvider(poolID, procEnv, procCtx, procCfg, termSpec)
	} else {
		return s.checkClient(poolID, procEnv, procCtx, procCfg, termSpec)
	}
}

func (s *service) checkProvider(
	poolID id.ADT,
	procEnv procexec.Env,
	procCtx typedef.Context,
	procCfg procexec.Cfg,
	ts procdef.TermSpec,
) error {
	switch termSpec := ts.(type) {
	case procdef.CloseSpec:
		// check ctx
		if len(procCtx.Assets) > 0 {
			err := fmt.Errorf("context mismatch: want 0 items, got %v items", len(procCtx.Assets))
			s.log.Error("checking failed")
			return err
		}
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVia, typedef.OneRec{})
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		delete(procCtx.Liabs, termSpec.CommPH)
		return nil
	case procdef.WaitSpec:
		err := procdef.ErrTermTypeMismatch(ts, procdef.CloseSpec{})
		s.log.Error("checking failed")
		return err
	case procdef.SendSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.TensorRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.ValPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.ValPH)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[termSpec.CommPH] = wantVia.Z
		delete(procCtx.Assets, termSpec.ValPH)
		return nil
	case procdef.RecvSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.LolliRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.BindPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.BindPH)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Liabs[termSpec.CommPH] = wantVia.Z
		procCtx.Assets[termSpec.BindPH] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.ContTS)
	case procdef.LabSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.PlusRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check label
		choice, ok := wantVia.Zs[termSpec.Label]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Zs), termSpec.Label)
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[termSpec.CommPH] = choice
		return nil
	case procdef.CaseSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.WithRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check conts
		if len(termSpec.Conts) != len(wantVia.Zs) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Zs), len(termSpec.Conts))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Zs {
			cont, ok := termSpec.Conts[label]
			if !ok {
				err := fmt.Errorf("label mismatch: want %q, got nothing", label)
				s.log.Error("checking failed")
				return err
			}
			procCtx.Liabs[termSpec.CommPH] = choice
			err := s.checkState(poolID, procEnv, procCtx, procCfg, cont)
			if err != nil {
				s.log.Error("checking failed")
				return err
			}
		}
		return nil
	case procdef.FwdSpec:
		if len(procCtx.Assets) != 1 {
			err := fmt.Errorf("context mismatch: want 1 item, got %v items", len(procCtx.Assets))
			s.log.Error("checking failed")
			return err
		}
		viaSt, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		fwdSt, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		if fwdSt.Pol() != viaSt.Pol() {
			err := typedef.ErrPolarityMismatch(fwdSt, viaSt)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(fwdSt, viaSt)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		delete(procCtx.Liabs, termSpec.X)
		delete(procCtx.Assets, termSpec.Y)
		return nil
	default:
		panic(procdef.ErrTermTypeUnexpected(ts))
	}
}

func (s *service) checkClient(
	poolID id.ADT,
	procEnv procexec.Env,
	procCtx typedef.Context,
	procCfg procexec.Cfg,
	ts procdef.TermSpec,
) error {
	switch termSpec := ts.(type) {
	case procdef.CloseSpec:
		err := procdef.ErrTermTypeMismatch(ts, procdef.WaitSpec{})
		s.log.Error("checking failed")
		return err
	case procdef.WaitSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.OneRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check cont
		delete(procCtx.Assets, termSpec.CommPH)
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.ContTS)
	case procdef.SendSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.LolliRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.ValPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.ValPH)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[termSpec.CommPH] = wantVia.Z
		delete(procCtx.Assets, termSpec.ValPH)
		return nil
	case procdef.RecvSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.TensorRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.BindPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.BindPH)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[termSpec.CommPH] = wantVia.Z
		procCtx.Assets[termSpec.BindPH] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.ContTS)
	case procdef.LabSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.WithRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check label
		choice, ok := wantVia.Zs[termSpec.Label]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Zs), termSpec.Label)
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[termSpec.CommPH] = choice
		return nil
	case procdef.CaseSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(typedef.PlusRec)
		if !ok {
			err := typedef.ErrSnapTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check conts
		if len(termSpec.Conts) != len(wantVia.Zs) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Zs), len(termSpec.Conts))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Zs {
			cont, ok := termSpec.Conts[label]
			if !ok {
				err := fmt.Errorf("label mismatch: want %q, got nothing", label)
				s.log.Error("checking failed")
				return err
			}
			procCtx.Assets[termSpec.CommPH] = choice
			err := s.checkState(poolID, procEnv, procCtx, procCfg, cont)
			if err != nil {
				s.log.Error("checking failed")
				return err
			}
		}
		return nil
	case procdef.SpawnSpecOld:
		procSig, ok := procEnv.ProcSigs[termSpec.SigID]
		if !ok {
			err := procdec.ErrRootMissingInEnv(termSpec.SigID)
			s.log.Error("checking failed")
			return err
		}
		// check vals
		if len(termSpec.Ys) != len(procSig.Ys) {
			err := fmt.Errorf("context mismatch: want %v items, got %v items", len(procSig.Ys), len(termSpec.Ys))
			s.log.Error("checking failed", slog.Any("want", procSig.Ys), slog.Any("got", termSpec.Ys))
			return err
		}
		if len(termSpec.Ys) == 0 {
			return nil
		}
		for i, ep := range procSig.Ys {
			valRole, ok := procEnv.Types[ep.TypeQN]
			if !ok {
				err := typedef.ErrSymMissingInEnv(ep.TypeQN)
				s.log.Error("checking failed")
				return err
			}
			wantVal, ok := procEnv.TypeTerms[valRole.TermID]
			if !ok {
				err := typedef.ErrMissingInEnv(valRole.TermID)
				s.log.Error("checking failed")
				return err
			}
			gotVal, ok := procCtx.Assets[termSpec.Ys[i]]
			if !ok {
				err := procdef.ErrMissingInCtx(ep.CommPH)
				s.log.Error("checking failed")
				return err
			}
			err := typedef.CheckRec(gotVal, wantVal)
			if err != nil {
				s.log.Error("checking failed", slog.Any("want", wantVal), slog.Any("got", gotVal))
				return err
			}
			delete(procCtx.Assets, termSpec.Ys[i])
		}
		// check via
		viaRole, ok := procEnv.Types[procSig.X.TypeQN]
		if !ok {
			err := typedef.ErrSymMissingInEnv(procSig.X.TypeQN)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := procEnv.TypeTerms[viaRole.TermID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaRole.TermID)
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[termSpec.X] = wantVia
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	default:
		panic(procdef.ErrTermTypeUnexpected(ts))
	}
}

// Port
type Repo interface {
	Insert(data.Source, PoolRec) error
	InsertLiab(data.Source, procexec.Liab) error
	SelectRefs(data.Source) ([]PoolRef, error)
	SelectSubs(data.Source, id.ADT) (PoolSnap, error)
	SelectProc(data.Source, id.ADT) (procexec.Cfg, error)
	UpdateProc(data.Source, procexec.Mod) error
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	ConvertRecToRef func(PoolRec) PoolRef
)

func errOptimisticUpdate(got rn.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
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
