package def

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

	procdec "smecalculus/rolevod/app/proc/dec"
	procdef "smecalculus/rolevod/app/proc/def"
	proceval "smecalculus/rolevod/app/proc/eval"
	typedec "smecalculus/rolevod/app/type/dec"
	typedef "smecalculus/rolevod/app/type/def"
)

type PoolSpec struct {
	SigQN sym.ADT
	SupID id.ADT
}

type PoolRef struct {
	PoolID id.ADT
	ProcID id.ADT // main
}

type PoolRec struct {
	PoolID id.ADT
	ProcID id.ADT // main
	SupID  id.ADT
	PoolRN rn.ADT
}

type PoolSnap struct {
	PoolID id.ADT
	Title  string
	Subs   []PoolRef
}

type StepSpec struct {
	PoolID id.ADT
	ProcID id.ADT
	Term   procdef.TermSpec
}

// Port
type API interface {
	Create(PoolSpec) (PoolRef, error)
	Retrieve(id.ADT) (PoolSnap, error)
	RetreiveRefs() ([]PoolRef, error)
	Spawn(proceval.Spec) (proceval.Ref, error)
	Take(StepSpec) error
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	pools    Repo
	sigs     procdec.SigRepo
	roles    typedec.Repo
	states   typedef.TermRepo
	operator data.Operator
	log      *slog.Logger
}

func newService(
	pools Repo,
	sigs procdec.SigRepo,
	roles typedec.Repo,
	states typedef.TermRepo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	return &service{pools, sigs, roles, states, operator, l}
}

func (s *service) Create(spec PoolSpec) (PoolRef, error) {
	ctx := context.Background()
	s.log.Debug("creation started", slog.Any("spec", spec))
	impl := PoolRec{
		PoolID: id.New(),
		ProcID: id.New(),
		SupID:  spec.SupID,
		PoolRN: rn.Initial(),
	}
	liab := proceval.Liab{
		PoolID: impl.PoolID,
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
	s.log.Debug("creation succeeded", slog.Any("poolID", impl.PoolID))
	return ConvertRecToRef(impl), nil
}

func (s *service) Spawn(spec proceval.Spec) (_ proceval.Ref, err error) {
	procAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("spawning started", procAttr)
	return proceval.Ref{}, nil
}

func (s *service) Take(spec StepSpec) (err error) {
	idAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("taking started", idAttr)
	ctx := context.Background()
	// initial values
	poolID := spec.PoolID
	procID := spec.ProcID
	termSpec := spec.Term
	for termSpec != nil {
		var procCfg proceval.Cfg
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
		var sigs map[id.ADT]procdec.SigRec
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			sigs, err = s.sigs.SelectEnv(ds, sigIDs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("sigs", sigIDs))
			return err
		}
		roleQNs := procdec.CollectEnv(maps.Values(sigs))
		var roles map[sym.ADT]typedec.TypeRec
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			roles, err = s.roles.SelectEnv(ds, roleQNs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("roles", roleQNs))
			return err
		}
		envIDs := typedec.CollectEnv(maps.Values(roles))
		ctxIDs := CollectCtx(maps.Values(procCfg.Chnls))
		var types map[id.ADT]typedef.TermRec
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			types, err = s.states.SelectEnv(ds, append(envIDs, ctxIDs...))
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("env", envIDs), slog.Any("ctx", ctxIDs))
			return err
		}
		procEnv := proceval.Env{ProcSigs: sigs, Roles: roles, TypeTerms: types}
		procCtx := convertToCtx(poolID, maps.Values(procCfg.Chnls), types)
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
		termSpec = nextSpec.Term
	}
	s.log.Debug("taking succeeded", idAttr)
	return nil
}

func (s *service) takeWith(
	procEnv proceval.Env,
	procCfg proceval.Cfg,
	ts procdef.TermSpec,
) (
	tranSpec StepSpec,
	procMod proceval.Mod,
	_ error,
) {
	switch termSpec := ts.(type) {
	case procdef.CloseSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := proceval.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			sndrStep := procdef.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procdef.CloseRec{
					X: termSpec.X,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procdef.SvcRec)
		if !ok {
			panic(procdef.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case procdef.WaitRec:
			sndrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := proceval.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.X,
				PoolRN: -svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: svcStep.PoolID,
				ProcID: svcStep.ProcID,
				Term:   termImpl.Cont,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrContTypeUnexpected2(svcStep.Cont))
		}
	case procdef.WaitSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := proceval.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[viaChnl.ChnlID]
		if sndrStep == nil {
			rcvrStep := procdef.SvcRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procdef.WaitRec{
					X:    termSpec.X,
					Cont: termSpec.Cont,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(procdef.MsgRec)
		if !ok {
			panic(procdef.ErrRootTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.Val.(type) {
		case procdef.CloseRec:
			sndrViaBnd := proceval.Bnd{
				ProcID: msgStep.ProcID,
				ChnlPH: termImpl.X,
				PoolRN: -msgStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				Term:   termSpec.Cont,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		case procdef.FwdRec:
			rcvrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				ChnlID: termImpl.B,
				TermID: viaChnl.TermID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				Term:   termSpec,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrValTypeUnexpected2(msgStep.Val))
		}
	case procdef.SendSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := proceval.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.TermID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, proceval.Mod{}, err
		}
		viaStateID := viaState.(typedef.ProdRec).Next()
		valChnl, ok := procCfg.Chnls[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.Y)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, proceval.Mod{}, err
		}
		sndrValBnd := proceval.Bnd{
			ProcID: procCfg.ProcID,
			ChnlPH: termSpec.Y,
			PoolRN: -procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrValBnd)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newChnlID := id.New()
			sndrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				ChnlID: newChnlID,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procdef.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procdef.SendRec{
					X:      termSpec.X,
					A:      newChnlID,
					B:      valChnl.ChnlID,
					TermID: valChnl.TermID,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procdef.SvcRec)
		if !ok {
			panic(procdef.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case procdef.RecvRec:
			sndrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				ChnlID: termImpl.A,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := proceval.Bnd{
				ProcID: svcStep.ProcID,
				ChnlPH: termImpl.X,
				ChnlID: termImpl.A,
				TermID: viaStateID,
				PoolRN: svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := proceval.Bnd{
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
				Term:   termImpl.Cont,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrContTypeUnexpected2(svcStep.Cont))
		}
	case procdef.RecvSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := proceval.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrSemRec := procCfg.Steps[viaChnl.ChnlID]
		if sndrSemRec == nil {
			rcvrSemRec := procdef.SvcRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procdef.RecvRec{
					X:    termSpec.X,
					A:    id.New(),
					Y:    termSpec.Y,
					Cont: termSpec.Cont,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrSemRec)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		sndrMsgRec, ok := sndrSemRec.(procdef.MsgRec)
		if !ok {
			panic(procdef.ErrRootTypeUnexpected(sndrSemRec))
		}
		switch termRec := sndrMsgRec.Val.(type) {
		case procdef.SendRec:
			viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
			if !ok {
				err := typedef.ErrMissingInEnv(viaChnl.TermID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, proceval.Mod{}, err
			}
			rcvrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				ChnlID: termRec.A,
				TermID: viaState.(typedef.ProdRec).Next(),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.Y,
				ChnlID: termRec.B,
				TermID: termRec.TermID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				Term:   termSpec.Cont,
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrValTypeUnexpected2(sndrMsgRec.Val))
		}
	case procdef.LabSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := proceval.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.TermID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, proceval.Mod{}, err
		}
		viaStateID := viaState.(typedef.SumRec).Next(termSpec.Label)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newViaID := id.New()
			sndrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				ChnlID: newViaID,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := procdef.MsgRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: procdef.LabRec{
					X:     termSpec.X,
					A:     newViaID,
					Label: termSpec.Label,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(procdef.SvcRec)
		if !ok {
			panic(procdef.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case procdef.CaseRec:
			sndrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				ChnlID: termImpl.A,
				TermID: viaStateID,
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := proceval.Bnd{
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
				Term:   termImpl.Conts[termSpec.Label],
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrContTypeUnexpected2(svcStep.Cont))
		}
	case procdef.CaseSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := proceval.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[viaChnl.ChnlID]
		if sndrStep == nil {
			rcvrStep := procdef.SvcRec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: procdef.CaseRec{
					X:     termSpec.X,
					A:     id.New(),
					Conts: termSpec.Conts,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(procdef.MsgRec)
		if !ok {
			panic(procdef.ErrRootTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.Val.(type) {
		case procdef.LabRec:
			viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
			if !ok {
				err := typedef.ErrMissingInEnv(viaChnl.TermID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, proceval.Mod{}, err
			}
			rcvrViaBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				ChnlID: termImpl.A,
				TermID: viaState.(typedef.SumRec).Next(termImpl.Label),
				PoolRN: procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				Term:   termSpec.Conts[termImpl.Label],
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(procdef.ErrValTypeUnexpected2(msgStep.Val))
		}
	case procdef.SpawnSpec:
		rcvrSnap, ok := procEnv.Locks[termSpec.PoolQN]
		if !ok {
			err := errMissingPool(termSpec.PoolQN)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		rcvrLiab := proceval.Liab{
			ProcID: id.New(),
			PoolID: rcvrSnap.PoolID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Liabs = append(procMod.Liabs, rcvrLiab)
		rcvrSig, ok := procEnv.ProcSigs[termSpec.SigID]
		if !ok {
			err := errMissingSig(termSpec.SigID)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		rcvrRole, ok := procEnv.Roles[rcvrSig.X.TypeQN]
		if !ok {
			err := errMissingRole(rcvrSig.X.TypeQN)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		newViaID := id.New()
		sndrViaBnd := proceval.Bnd{
			ProcID: procCfg.ProcID,
			ChnlPH: termSpec.X,
			ChnlID: newViaID,
			TermID: rcvrRole.TermID,
			PoolRN: procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
		rcvrViaBnd := proceval.Bnd{
			ProcID: rcvrLiab.ProcID,
			ChnlPH: rcvrSig.X.ChnlPH,
			ChnlID: newViaID,
			TermID: rcvrRole.TermID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
		for i, chnlPH := range termSpec.Ys {
			valChnl, ok := procCfg.Chnls[chnlPH]
			if !ok {
				err := proceval.ErrMissingChnl(chnlPH)
				s.log.Error("taking failed")
				return StepSpec{}, proceval.Mod{}, err
			}
			sndrValBnd := proceval.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: chnlPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrValBnd)
			rcvrValBnd := proceval.Bnd{
				ProcID: rcvrLiab.ProcID,
				ChnlPH: rcvrSig.Ys[i].ChnlPH,
				ChnlID: valChnl.ChnlID,
				TermID: valChnl.TermID,
				PoolRN: rcvrSnap.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrValBnd)
		}
		tranSpec = StepSpec{
			PoolID: procCfg.PoolID,
			ProcID: procCfg.ProcID,
			Term:   termSpec.Cont,
		}
		s.log.Debug("taking succeeded")
		return tranSpec, procMod, nil
	case procdef.FwdSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		viaState, ok := procEnv.TypeTerms[viaChnl.TermID]
		if !ok {
			err := typedef.ErrMissingInEnv(viaChnl.TermID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, proceval.Mod{}, err
		}
		valChnl, ok := procCfg.Chnls[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCfg(termSpec.Y)
			s.log.Error("taking failed")
			return StepSpec{}, proceval.Mod{}, err
		}
		vs := procCfg.Steps[viaChnl.ChnlID]
		switch viaState.Pol() {
		case pol.Pos:
			switch viaStep := vs.(type) {
			case procdef.SvcRec:
				xBnd := proceval.Bnd{
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
					Term:   viaStep.Cont,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case procdef.MsgRec:
				yBnd := proceval.Bnd{
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
					Term:   viaStep.Val,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case nil:
				xBnd := proceval.Bnd{
					ProcID: procCfg.ProcID,
					ChnlPH: termSpec.X,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				yBnd := proceval.Bnd{
					ProcID: procCfg.ProcID,
					ChnlPH: termSpec.Y,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				msgStep := procdef.MsgRec{
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
				panic(procdef.ErrRootTypeUnexpected(vs))
			}
		case pol.Neg:
			switch viaStep := vs.(type) {
			case procdef.SvcRec:
				yBnd := proceval.Bnd{
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
					Term:   viaStep.Cont,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case procdef.MsgRec:
				xBnd := proceval.Bnd{
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
					Term:   viaStep.Val,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case nil:
				svcStep := procdef.SvcRec{
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
				panic(procdef.ErrRootTypeUnexpected(vs))
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

func CollectCtx(chnls []proceval.EP) []id.ADT {
	return nil
}

func convertToCtx(poolID id.ADT, chnls []proceval.EP, types map[id.ADT]typedef.TermRec) typedef.Context {
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
	procEnv proceval.Env,
	procCtx typedef.Context,
	procCfg proceval.Cfg,
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
	procEnv proceval.Env,
	procCtx typedef.Context,
	procCfg proceval.Cfg,
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
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVia, typedef.OneRec{})
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		delete(procCtx.Liabs, termSpec.X)
		return nil
	case procdef.WaitSpec:
		err := procdef.ErrTermTypeMismatch(ts, procdef.CloseSpec{})
		s.log.Error("checking failed")
		return err
	case procdef.SendSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.X)
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
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.B)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[termSpec.X] = wantVia.C
		delete(procCtx.Assets, termSpec.Y)
		return nil
	case procdef.RecvSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.X)
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
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Liabs[termSpec.X] = wantVia.Z
		procCtx.Assets[termSpec.Y] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	case procdef.LabSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.X)
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
		choice, ok := wantVia.Choices[termSpec.Label]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Choices), termSpec.Label)
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[termSpec.X] = choice
		return nil
	case procdef.CaseSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := typedef.ErrMissingInCtx(termSpec.X)
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
		if len(termSpec.Conts) != len(wantVia.Choices) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Choices), len(termSpec.Conts))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Choices {
			cont, ok := termSpec.Conts[label]
			if !ok {
				err := fmt.Errorf("label mismatch: want %q, got nothing", label)
				s.log.Error("checking failed")
				return err
			}
			procCtx.Liabs[termSpec.X] = choice
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
	procEnv proceval.Env,
	procCtx typedef.Context,
	procCfg proceval.Cfg,
	ts procdef.TermSpec,
) error {
	switch termSpec := ts.(type) {
	case procdef.CloseSpec:
		err := procdef.ErrTermTypeMismatch(ts, procdef.WaitSpec{})
		s.log.Error("checking failed")
		return err
	case procdef.WaitSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.X)
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
		delete(procCtx.Assets, termSpec.X)
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	case procdef.SendSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.X)
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
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[termSpec.X] = wantVia.Z
		delete(procCtx.Assets, termSpec.Y)
		return nil
	case procdef.RecvSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.X)
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
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := typedef.CheckRec(gotVal, wantVia.B)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[termSpec.X] = wantVia.C
		procCtx.Assets[termSpec.Y] = wantVia.B
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	case procdef.LabSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.X)
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
		choice, ok := wantVia.Choices[termSpec.Label]
		if !ok {
			err := fmt.Errorf("label mismatch: want %v, got %q", maps.Keys(wantVia.Choices), termSpec.Label)
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[termSpec.X] = choice
		return nil
	case procdef.CaseSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := procdef.ErrMissingInCtx(termSpec.X)
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
		if len(termSpec.Conts) != len(wantVia.Choices) {
			err := fmt.Errorf("state mismatch: want %v choices, got %v conts", len(wantVia.Choices), len(termSpec.Conts))
			s.log.Error("checking failed")
			return err
		}
		for label, choice := range wantVia.Choices {
			cont, ok := termSpec.Conts[label]
			if !ok {
				err := fmt.Errorf("label mismatch: want %q, got nothing", label)
				s.log.Error("checking failed")
				return err
			}
			procCtx.Assets[termSpec.X] = choice
			err := s.checkState(poolID, procEnv, procCtx, procCfg, cont)
			if err != nil {
				s.log.Error("checking failed")
				return err
			}
		}
		return nil
	case procdef.SpawnSpec:
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
			valRole, ok := procEnv.Roles[ep.TypeQN]
			if !ok {
				err := typedec.ErrMissingInEnv(ep.TypeQN)
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
				err := procdef.ErrMissingInCtx(ep.ChnlPH)
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
		viaRole, ok := procEnv.Roles[procSig.X.TypeQN]
		if !ok {
			err := typedec.ErrMissingInEnv(procSig.X.TypeQN)
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
	InsertLiab(data.Source, proceval.Liab) error
	SelectRefs(data.Source) ([]PoolRef, error)
	SelectSubs(data.Source, id.ADT) (PoolSnap, error)
	SelectProc(data.Source, id.ADT) (proceval.Cfg, error)
	UpdateProc(data.Source, proceval.Mod) error
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
