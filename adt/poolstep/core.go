package poolstep

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/exp/maps"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/polarity"
	"orglang/orglang/adt/procdec"
	"orglang/orglang/adt/procdef"
	"orglang/orglang/adt/procexec"
	"orglang/orglang/adt/procexp"
	"orglang/orglang/adt/qualsym"
	"orglang/orglang/adt/revnum"
	"orglang/orglang/adt/typedef"
	"orglang/orglang/adt/typeexp"
)

// Port
type API interface {
	Take(StepSpec) error
}

type StepSpec struct {
	ExecID identity.ADT
	ProcID identity.ADT
	ProcES procexp.ExpSpec
}

type PollSpec struct {
	ExecID identity.ADT
}

type service struct {
	poolSteps Repo
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
	poolSteps Repo,
	procDecs procdec.Repo,
	typeDefs typedef.Repo,
	typeExps typeexp.Repo,
	operator db.Operator,
	l *slog.Logger,
) *service {
	return &service{poolSteps, procDecs, typeDefs, typeExps, operator, l}
}

func (s *service) Take(spec StepSpec) (err error) {
	idAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("taking started", idAttr)
	ctx := context.Background()
	// initial values
	poolID := spec.ExecID
	procID := spec.ProcID
	termSpec := spec.ProcES
	for termSpec != nil {
		var procCfg procexec.Cfg
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			procCfg, err = s.poolSteps.SelectProc(ds, procID)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr)
			return err
		}
		if len(procCfg.Chnls) == 0 {
			panic("zero channels")
		}
		sigIDs := procexp.CollectEnv(termSpec)
		var sigs map[identity.ADT]procdec.DecRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			sigs, err = s.procDecs.SelectEnv(ds, sigIDs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("sigs", sigIDs))
			return err
		}
		typeQNs := procdec.CollectEnv(maps.Values(sigs))
		var types map[qualsym.ADT]typedef.DefRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			types, err = s.typeDefs.SelectEnv(ds, typeQNs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("types", typeQNs))
			return err
		}
		envIDs := typedef.CollectEnv(maps.Values(types))
		ctxIDs := CollectCtx(maps.Values(procCfg.Chnls))
		var terms map[identity.ADT]typeexp.ExpRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			terms, err = s.typeExps.SelectEnv(ds, append(envIDs, ctxIDs...))
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("env", envIDs), slog.Any("ctx", ctxIDs))
			return err
		}
		procEnv := procexec.Env{ProcDecs: sigs, TypeDefs: types, TypeTerms: terms}
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
		err = s.operator.Explicit(ctx, func(ds db.Source) error {
			err = s.poolSteps.UpdateProc(ds, procMod)
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
		termSpec = nextSpec.ProcES
	}
	s.log.Debug("taking succeed", idAttr)
	return nil
}

func (s *service) takeWith(
	procEnv procexec.Env,
	procCfg procexec.Cfg,
	ts procexp.ExpSpec,
) (
	tranSpec StepSpec,
	procMod procexec.Mod,
	_ error,
) {
	switch termSpec := ts.(type) {
	case procexp.CloseSpec:
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
				ProcID: procCfg.ExecID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procexp.CloseRec{
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
		case procexp.WaitRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
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
				ExecID: svcStep.PoolID,
				ProcID: svcStep.ProcID,
				ProcES: termImpl.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(svcStep.Cont))
		}
	case procexp.WaitSpec:
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
				ProcID: procCfg.ExecID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procexp.WaitRec{
					X:      termSpec.CommPH,
					ContES: termSpec.ContES,
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
		case procexp.CloseRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: msgStep.ProcID,
				ChnlPH: termImpl.X,
				PoolRN: -msgStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: termSpec.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return tranSpec, procMod, nil
		case procexp.FwdRec:
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.B,
				ExpID:  viaChnl.ExpID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: termSpec,
			}
			s.log.Debug("taking succeed", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(msgStep.Val))
		}
	case procexp.SendSpec:
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
		viaState, ok := procEnv.TypeTerms[viaChnl.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.ExpID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procexec.Mod{}, err
		}
		viaStateID := viaState.(typeexp.ProdRec).Next()
		valChnl, ok := procCfg.Chnls[termSpec.ValPH]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.ValPH)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procexec.Mod{}, err
		}
		sndrValBnd := procexec.Bnd{
			ProcID: procCfg.ExecID,
			ChnlPH: termSpec.ValPH,
			PoolRN: -procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrValBnd)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newChnlID := identity.New()
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				ChnlID: newChnlID,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procexec.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procexp.SendRec{
					X:     termSpec.CommPH,
					A:     newChnlID,
					B:     valChnl.ChnlID,
					ExpID: valChnl.ExpID,
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
		case procexp.RecvRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.A,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procexec.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.X,
				ChnlID: termImpl.A,
				ExpID:  viaStateID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := procexec.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.Y,
				ChnlID: valChnl.ChnlID,
				ExpID:  valChnl.ExpID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
			tranSpec = StepSpec{
				ExecID: svcStep.PoolID,
				ProcID: svcStep.ProcID,
				ProcES: termImpl.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(svcStep.Cont))
		}
	case procexp.RecvSpec:
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
				ProcID: procCfg.ExecID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procexp.RecvRec{
					X:      termSpec.CommPH,
					A:      identity.New(),
					Y:      termSpec.BindPH,
					ContES: termSpec.ContES,
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
		case procexp.SendRec:
			viaState, ok := procEnv.TypeTerms[viaChnl.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(viaChnl.ExpID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, procexec.Mod{}, err
			}
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termRec.A,
				ExpID:  viaState.(typeexp.ProdRec).Next(),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.BindPH,
				ChnlID: termRec.B,
				ExpID:  termRec.ExpID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
			tranSpec = StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: termSpec.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(sndrMsgRec.Val))
		}
	case procexp.LabSpec:
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
		viaState, ok := procEnv.TypeTerms[viaChnl.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.ExpID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procexec.Mod{}, err
		}
		viaStateID := viaState.(typeexp.SumRec).Next(termSpec.Label)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newViaID := identity.New()
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				ChnlID: newViaID,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procexec.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procexp.LabRec{
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
		case procexp.CaseRec:
			sndrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.A,
				ExpID:  viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procexec.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.X,
				ChnlID: termImpl.A,
				ExpID:  viaStateID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				ExecID: svcStep.PoolID,
				ProcID: svcStep.ProcID,
				ProcES: termImpl.ContESs[termSpec.Label],
			}
			s.log.Debug("taking succeed", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(svcStep.Cont))
		}
	case procexp.CaseSpec:
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
				ProcID: procCfg.ExecID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procexp.CaseRec{
					X:       termSpec.CommPH,
					A:       identity.New(),
					ContESs: termSpec.ContESs,
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
		case procexp.LabRec:
			viaState, ok := procEnv.TypeTerms[viaChnl.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(viaChnl.ExpID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, procexec.Mod{}, err
			}
			rcvrViaBnd := procexec.Bnd{
				ProcID: procCfg.ExecID,
				ChnlPH: termSpec.CommPH,
				ChnlID: termImpl.A,
				ExpID:  viaState.(typeexp.SumRec).Next(termImpl.Label),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				ExecID: procCfg.PoolID,
				ProcID: procCfg.ExecID,
				ProcES: termSpec.ContESs[termImpl.Label],
			}
			s.log.Debug("taking succeed", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(msgStep.Val))
		}
	case procexp.SpawnSpecOld:
		rcvrSnap, ok := procEnv.Locks[termSpec.PoolQN]
		if !ok {
			err := errMissingPool(termSpec.PoolQN)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		rcvrLiab := procexec.Liab{
			ProcID: identity.New(),
			PoolID: rcvrSnap.PoolID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Liabs = append(procMod.Liabs, rcvrLiab)
		rcvrSig, ok := procEnv.ProcDecs[termSpec.SigID]
		if !ok {
			err := errMissingSig(termSpec.SigID)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		rcvrRole, ok := procEnv.TypeDefs[rcvrSig.X.TypeQN]
		if !ok {
			err := errMissingRole(rcvrSig.X.TypeQN)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		newViaID := identity.New()
		sndrViaBnd := procexec.Bnd{
			ProcID: procCfg.ExecID,
			ChnlPH: termSpec.X,
			ChnlID: newViaID,
			ExpID:  rcvrRole.ExpID,
			PoolRN: procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
		rcvrViaBnd := procexec.Bnd{
			ProcID: rcvrLiab.ProcID,
			ChnlPH: rcvrSig.X.BindPH,
			ChnlID: newViaID,
			ExpID:  rcvrRole.ExpID,
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
				ProcID: procCfg.ExecID,
				ChnlPH: chnlPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrValBnd)
			rcvrValBnd := procexec.Bnd{
				ProcID: rcvrLiab.ProcID,
				ChnlPH: rcvrSig.Ys[i].BindPH,
				ChnlID: valChnl.ChnlID,
				ExpID:  valChnl.ExpID,
				PoolRN: rcvrSnap.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
		}
		tranSpec = StepSpec{
			ExecID: procCfg.PoolID,
			ProcID: procCfg.ExecID,
			ProcES: termSpec.ContES,
		}
		s.log.Debug("taking succeed")
		return tranSpec, procMod, nil
	case procexp.FwdSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procexec.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		viaState, ok := procEnv.TypeTerms[viaChnl.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.ExpID)
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
		case polarity.Pos:
			switch viaStep := vs.(type) {
			case procexec.SvcRec:
				xBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Cont.Via(),
					ChnlID: viaChnl.ChnlID,
					ExpID:  viaChnl.ExpID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				tranSpec = StepSpec{
					ExecID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcES: viaStep.Cont,
				}
				s.log.Debug("taking succeed", viaAttr)
				return tranSpec, procMod, nil
			case procexec.MsgRec:
				yBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Val.Via(),
					ChnlID: valChnl.ChnlID,
					ExpID:  valChnl.ExpID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				tranSpec = StepSpec{
					ExecID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcES: viaStep.Val,
				}
				s.log.Debug("taking succeed", viaAttr)
				return tranSpec, procMod, nil
			case nil:
				xBnd := procexec.Bnd{
					ProcID: procCfg.ExecID,
					ChnlPH: termSpec.X,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				yBnd := procexec.Bnd{
					ProcID: procCfg.ExecID,
					ChnlPH: termSpec.Y,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				msgStep := procexec.MsgRec{
					PoolID: procCfg.PoolID,
					ProcID: procCfg.ExecID,
					ChnlID: viaChnl.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					Val: procexp.FwdRec{
						B: valChnl.ChnlID,
					},
				}
				procMod.Steps = append(procMod.Steps, msgStep)
				s.log.Debug("taking half done", viaAttr)
				return tranSpec, procMod, nil
			default:
				panic(procexec.ErrRootTypeUnexpected(vs))
			}
		case polarity.Neg:
			switch viaStep := vs.(type) {
			case procexec.SvcRec:
				yBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Cont.Via(),
					ChnlID: valChnl.ChnlID,
					ExpID:  valChnl.ExpID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				tranSpec = StepSpec{
					ExecID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcES: viaStep.Cont,
				}
				s.log.Debug("taking succeed", viaAttr)
				return tranSpec, procMod, nil
			case procexec.MsgRec:
				xBnd := procexec.Bnd{
					ProcID: viaStep.ProcID,
					ChnlPH: viaStep.Val.Via(),
					ChnlID: viaChnl.ChnlID,
					ExpID:  viaChnl.ExpID,
					PoolRN: viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				tranSpec = StepSpec{
					ExecID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					ProcES: viaStep.Val,
				}
				s.log.Debug("taking succeed", viaAttr)
				return tranSpec, procMod, nil
			case nil:
				svcStep := procexec.SvcRec{
					PoolID: procCfg.PoolID,
					ProcID: procCfg.ExecID,
					ChnlID: viaChnl.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					Cont: procexp.FwdRec{
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
			panic(typeexp.ErrPolarityUnexpected(viaState))
		}
	default:
		panic(procexp.ErrExpTypeUnexpected(ts))
	}
}

func CollectCtx(chnls []procexec.EP) []identity.ADT {
	return nil
}

func convertToCtx(poolID identity.ADT, chnls []procexec.EP, types map[identity.ADT]typeexp.ExpRec) typedef.Context {
	assets := make(map[qualsym.ADT]typeexp.ExpRec, len(chnls)-1)
	liabs := make(map[qualsym.ADT]typeexp.ExpRec, 1)
	for _, ch := range chnls {
		if poolID == ch.PoolID {
			liabs[ch.ChnlPH] = types[ch.ExpID]
		} else {
			assets[ch.ChnlPH] = types[ch.ExpID]
		}
	}
	return typedef.Context{Assets: assets, Liabs: liabs}
}

func (s *service) checkState(
	poolID identity.ADT,
	procEnv procexec.Env,
	procCtx typedef.Context,
	procCfg procexec.Cfg,
	termSpec procexp.ExpSpec,
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
	poolID identity.ADT,
	procEnv procexec.Env,
	procCtx typedef.Context,
	procCfg procexec.Cfg,
	ts procexp.ExpSpec,
) error {
	switch termSpec := ts.(type) {
	case procexp.CloseSpec:
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
		err := typeexp.CheckRec(gotVia, typeexp.OneRec{})
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		delete(procCtx.Liabs, termSpec.CommPH)
		return nil
	case procexp.WaitSpec:
		err := procexp.ErrExpTypeMismatch(ts, procexp.CloseSpec{})
		s.log.Error("checking failed")
		return err
	case procexp.SendSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
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
		gotVal, ok := procCtx.Assets[termSpec.ValPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.ValPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[termSpec.CommPH] = wantVia.Z
		delete(procCtx.Assets, termSpec.ValPH)
		return nil
	case procexp.RecvSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
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
		gotVal, ok := procCtx.Assets[termSpec.BindPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.BindPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Liabs[termSpec.CommPH] = wantVia.Z
		procCtx.Assets[termSpec.BindPH] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.ContES)
	case procexp.LabSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
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
		choice, ok := wantVia.Zs[termSpec.Label]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Zs), termSpec.Label)
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[termSpec.CommPH] = choice
		return nil
	case procexp.CaseSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.CommPH]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.CommPH)
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
		if len(termSpec.ContESs) != len(wantVia.Zs) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Zs), len(termSpec.ContESs))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Zs {
			cont, ok := termSpec.ContESs[label]
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
	case procexp.FwdSpec:
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
			err := typeexp.ErrPolarityMismatch(fwdSt, viaSt)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(fwdSt, viaSt)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		delete(procCtx.Liabs, termSpec.X)
		delete(procCtx.Assets, termSpec.Y)
		return nil
	default:
		panic(procexp.ErrExpTypeUnexpected(ts))
	}
}

func (s *service) checkClient(
	poolID identity.ADT,
	procEnv procexec.Env,
	procCtx typedef.Context,
	procCfg procexec.Cfg,
	ts procexp.ExpSpec,
) error {
	switch termSpec := ts.(type) {
	case procexp.CloseSpec:
		err := procexp.ErrExpTypeMismatch(ts, procexp.WaitSpec{})
		s.log.Error("checking failed")
		return err
	case procexp.WaitSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
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
		delete(procCtx.Assets, termSpec.CommPH)
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.ContES)
	case procexp.SendSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
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
		gotVal, ok := procCtx.Assets[termSpec.ValPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.ValPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[termSpec.CommPH] = wantVia.Z
		delete(procCtx.Assets, termSpec.ValPH)
		return nil
	case procexp.RecvSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
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
		gotVal, ok := procCtx.Assets[termSpec.BindPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.BindPH)
			s.log.Error("checking failed")
			return err
		}
		err := typeexp.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[termSpec.CommPH] = wantVia.Z
		procCtx.Assets[termSpec.BindPH] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.ContES)
	case procexp.LabSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
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
		choice, ok := wantVia.Zs[termSpec.Label]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Zs), termSpec.Label)
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[termSpec.CommPH] = choice
		return nil
	case procexp.CaseSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.CommPH]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.CommPH)
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
		if len(termSpec.ContESs) != len(wantVia.Zs) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Zs), len(termSpec.ContESs))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Zs {
			cont, ok := termSpec.ContESs[label]
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
	case procexp.SpawnSpecOld:
		procSig, ok := procEnv.ProcDecs[termSpec.SigID]
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
			valRole, ok := procEnv.TypeDefs[ep.TypeQN]
			if !ok {
				err := typedef.ErrSymMissingInEnv(ep.TypeQN)
				s.log.Error("checking failed")
				return err
			}
			wantVal, ok := procEnv.TypeTerms[valRole.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(valRole.ExpID)
				s.log.Error("checking failed")
				return err
			}
			gotVal, ok := procCtx.Assets[termSpec.Ys[i]]
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
			delete(procCtx.Assets, termSpec.Ys[i])
		}
		// check via
		viaRole, ok := procEnv.TypeDefs[procSig.X.TypeQN]
		if !ok {
			err := typedef.ErrSymMissingInEnv(procSig.X.TypeQN)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := procEnv.TypeTerms[viaRole.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaRole.ExpID)
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[termSpec.X] = wantVia
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.ContES)
	default:
		panic(procexp.ErrExpTypeUnexpected(ts))
	}
}

func errOptimisticUpdate(got revnum.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
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
