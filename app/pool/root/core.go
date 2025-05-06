package pool

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/exp/maps"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/pol"
	rn "smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/chnl"
	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	procroot "smecalculus/rolevod/app/proc/root"
	procsig "smecalculus/rolevod/app/proc/sig"
	rolesig "smecalculus/rolevod/app/role/sig"
)

type Spec struct {
	SigQN sym.ADT
	SupID id.ADT
}

type Ref struct {
	PoolID id.ADT
	ProcID id.ADT // main
}

type impl struct {
	PoolID id.ADT
	ProcID id.ADT // main
	SupID  id.ADT
	PoolRN rn.ADT
}

type Snap struct {
	PoolID id.ADT
	Title  string
	Subs   []Ref
}

type StepSpec struct {
	PoolID id.ADT
	ProcID id.ADT
	Term   step.Term
}

// Port
type API interface {
	Create(Spec) (Ref, error)
	Retrieve(id.ADT) (Snap, error)
	RetreiveRefs() ([]Ref, error)
	Spawn(procroot.Spec) (procroot.Ref, error)
	Take(StepSpec) error
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	pools    Repo
	sigs     procsig.Repo
	roles    rolesig.Repo
	states   state.Repo
	operator data.Operator
	log      *slog.Logger
}

func newService(
	pools Repo,
	sigs procsig.Repo,
	roles rolesig.Repo,
	states state.Repo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	return &service{pools, sigs, roles, states, operator, l}
}

func (s *service) Create(spec Spec) (Ref, error) {
	ctx := context.Background()
	s.log.Debug("creation started", slog.Any("spec", spec))
	impl := impl{
		PoolID: id.New(),
		ProcID: id.New(),
		SupID:  spec.SupID,
		PoolRN: rn.Initial(),
	}
	liab := procroot.Liab{
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
		return Ref{}, err
	}
	s.log.Debug("creation succeeded", slog.Any("poolID", impl.PoolID))
	return ConvertRootToRef(impl), nil
}

func (s *service) Spawn(spec procroot.Spec) (_ procroot.Ref, err error) {
	procAttr := slog.Any("procID", spec.ProcID)
	s.log.Debug("spawning started", procAttr)
	return procroot.Ref{}, nil
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
		var procCfg procroot.Cfg
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
		sigIDs := step.CollectEnv(termSpec)
		var sigs map[id.ADT]procsig.Impl
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			sigs, err = s.sigs.SelectEnv(ds, sigIDs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("sigs", sigIDs))
			return err
		}
		roleQNs := procsig.CollectEnv(maps.Values(sigs))
		var roles map[sym.ADT]rolesig.Impl
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			roles, err = s.roles.SelectEnv(ds, roleQNs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("roles", roleQNs))
			return err
		}
		envIDs := rolesig.CollectEnv(maps.Values(roles))
		ctxIDs := CollectCtx(maps.Values(procCfg.Chnls))
		var states map[state.ID]state.Root
		err = s.operator.Implicit(ctx, func(ds data.Source) error {
			states, err = s.states.SelectEnv(ds, append(envIDs, ctxIDs...))
			return err
		})
		if err != nil {
			s.log.Error("taking failed", idAttr, slog.Any("env", envIDs), slog.Any("ctx", ctxIDs))
			return err
		}
		procEnv := procroot.Env{Sigs: sigs, Roles: roles, States: states}
		procCtx := convertToCtx(poolID, maps.Values(procCfg.Chnls), states)
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
	procEnv procroot.Env,
	procCfg procroot.Cfg,
	ts step.Term,
) (
	tranSpec StepSpec,
	procMod procroot.Mod,
	_ error,
) {
	switch termSpec := ts.(type) {
	case step.CloseSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := procroot.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			sndrStep := step.MsgRoot2{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: step.CloseImpl{
					X: termSpec.X,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(step.SvcRoot)
		if !ok {
			panic(step.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case step.WaitImpl:
			sndrViaBnd := procroot.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: termSpec.X,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procroot.Bnd{
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
			panic(step.ErrContTypeUnexpected2(svcStep.Cont))
		}
	case step.WaitSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := procroot.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[viaChnl.ChnlID]
		if sndrStep == nil {
			rcvrStep := step.SvcRoot{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: step.WaitImpl{
					X:    termSpec.X,
					Cont: termSpec.Cont,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(step.MsgRoot2)
		if !ok {
			panic(step.ErrRootTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.Val.(type) {
		case step.CloseImpl:
			sndrViaBnd := procroot.Bnd{
				ProcID: msgStep.ProcID,
				ChnlPH: termImpl.X,
				PoolRN: -msgStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procroot.Bnd{
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
		case step.FwdImpl:
			rcvrViaBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.X,
				ChnlID:  termImpl.B,
				StateID: viaChnl.StateID,
				PoolRN:  procCfg.PoolRN.Next(),
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
			panic(step.ErrValTypeUnexpected2(msgStep.Val))
		}
	case step.SendSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := procroot.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.States[viaChnl.StateID]
		if !ok {
			err := state.ErrMissingInEnv(viaChnl.StateID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procroot.Mod{}, err
		}
		viaStateID := viaState.(state.Prod).Next()
		valChnl, ok := procCfg.Chnls[termSpec.Y]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.Y)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procroot.Mod{}, err
		}
		sndrValBnd := procroot.Bnd{
			ProcID: procCfg.ProcID,
			ChnlPH: termSpec.Y,
			PoolRN: -procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrValBnd)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newChnlID := id.New()
			sndrViaBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.X,
				ChnlID:  newChnlID,
				StateID: viaStateID,
				PoolRN:  procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := step.MsgRoot2{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: step.SendImpl{
					X: termSpec.X,
					A: newChnlID,
					B: valChnl.ChnlID,
					S: valChnl.StateID,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(step.SvcRoot)
		if !ok {
			panic(step.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case step.RecvImpl:
			sndrViaBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.X,
				ChnlID:  termImpl.A,
				StateID: viaStateID,
				PoolRN:  procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procroot.Bnd{
				ProcID:  svcStep.ProcID,
				ChnlPH:  termImpl.X,
				ChnlID:  termImpl.A,
				StateID: viaStateID,
				PoolRN:  svcStep.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := procroot.Bnd{
				ProcID:  svcStep.ProcID,
				ChnlPH:  termImpl.Y,
				ChnlID:  valChnl.ChnlID,
				StateID: valChnl.StateID,
				PoolRN:  svcStep.PoolRN.Next(),
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
			panic(step.ErrContTypeUnexpected2(svcStep.Cont))
		}
	case step.RecvSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := procroot.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[viaChnl.ChnlID]
		if sndrStep == nil {
			rcvrStep := step.SvcRoot{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: step.RecvImpl{
					X:    termSpec.X,
					A:    id.New(),
					Y:    termSpec.Y,
					Cont: termSpec.Cont,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(step.MsgRoot2)
		if !ok {
			panic(step.ErrRootTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.Val.(type) {
		case step.SendImpl:
			viaState, ok := procEnv.States[viaChnl.StateID]
			if !ok {
				err := state.ErrMissingInEnv(viaChnl.StateID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, procroot.Mod{}, err
			}
			rcvrViaBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.X,
				ChnlID:  termImpl.A,
				StateID: viaState.(state.Prod).Next(),
				PoolRN:  procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			rcvrValBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.Y,
				ChnlID:  termImpl.B,
				StateID: termImpl.S,
				PoolRN:  procCfg.PoolRN.Next(),
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
			panic(step.ErrValTypeUnexpected2(msgStep.Val))
		}
	case step.LabSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		sndrLock := procroot.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, sndrLock)
		viaState, ok := procEnv.States[viaChnl.StateID]
		if !ok {
			err := state.ErrMissingInEnv(viaChnl.StateID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procroot.Mod{}, err
		}
		viaStateID := viaState.(state.Sum).Next(termSpec.Label)
		rcvrStep := procCfg.Steps[viaChnl.ChnlID]
		if rcvrStep == nil {
			newViaID := id.New()
			sndrViaBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.X,
				ChnlID:  newViaID,
				StateID: viaStateID,
				PoolRN:  procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			sndrStep := step.MsgRoot2{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Val: step.LabImpl{
					X: termSpec.X,
					A: newViaID,
					L: termSpec.Label,
				},
			}
			procMod.Steps = append(procMod.Steps, sndrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		svcStep, ok := rcvrStep.(step.SvcRoot)
		if !ok {
			panic(step.ErrRootTypeUnexpected(rcvrStep))
		}
		switch termImpl := svcStep.Cont.(type) {
		case step.CaseImpl:
			sndrViaBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.X,
				ChnlID:  termImpl.A,
				StateID: viaStateID,
				PoolRN:  procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
			rcvrViaBnd := procroot.Bnd{
				ProcID:  svcStep.ProcID,
				ChnlPH:  termImpl.X,
				ChnlID:  termImpl.A,
				StateID: viaStateID,
				PoolRN:  svcStep.PoolRN.Next(),
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
			panic(step.ErrContTypeUnexpected2(svcStep.Cont))
		}
	case step.CaseSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		rcvrLock := procroot.Lock{
			PoolID: procCfg.PoolID,
			PoolRN: procCfg.PoolRN,
		}
		procMod.Locks = append(procMod.Locks, rcvrLock)
		sndrStep := procCfg.Steps[viaChnl.ChnlID]
		if sndrStep == nil {
			rcvrStep := step.SvcRoot{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				ChnlID: viaChnl.ChnlID,
				PoolRN: procCfg.PoolRN.Next(),
				Cont: step.CaseImpl{
					X:     termSpec.X,
					A:     id.New(),
					Conts: termSpec.Conts,
				},
			}
			procMod.Steps = append(procMod.Steps, rcvrStep)
			s.log.Debug("taking half done", viaAttr)
			return tranSpec, procMod, nil
		}
		msgStep, ok := sndrStep.(step.MsgRoot2)
		if !ok {
			panic(step.ErrRootTypeUnexpected(sndrStep))
		}
		switch termImpl := msgStep.Val.(type) {
		case step.LabImpl:
			viaState, ok := procEnv.States[viaChnl.StateID]
			if !ok {
				err := state.ErrMissingInEnv(viaChnl.StateID)
				s.log.Error("taking failed", viaAttr)
				return StepSpec{}, procroot.Mod{}, err
			}
			rcvrViaBnd := procroot.Bnd{
				ProcID:  procCfg.ProcID,
				ChnlPH:  termSpec.X,
				ChnlID:  termImpl.A,
				StateID: viaState.(state.Sum).Next(termImpl.L),
				PoolRN:  procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
			tranSpec = StepSpec{
				PoolID: procCfg.PoolID,
				ProcID: procCfg.ProcID,
				Term:   termSpec.Conts[termImpl.L],
			}
			s.log.Debug("taking succeeded", viaAttr)
			return tranSpec, procMod, nil
		default:
			panic(step.ErrValTypeUnexpected2(msgStep.Val))
		}
	case step.SpawnSpec:
		rcvrSnap, ok := procEnv.Locks[termSpec.PoolQN]
		if !ok {
			err := errMissingPool(termSpec.PoolQN)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		rcvrLiab := procroot.Liab{
			ProcID: id.New(),
			PoolID: rcvrSnap.PoolID,
			PoolRN: rcvrSnap.PoolRN.Next(),
		}
		procMod.Liabs = append(procMod.Liabs, rcvrLiab)
		rcvrSig, ok := procEnv.Sigs[termSpec.SigID]
		if !ok {
			err := errMissingSig(termSpec.SigID)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		rcvrRole, ok := procEnv.Roles[rcvrSig.X.RoleQN]
		if !ok {
			err := errMissingRole(rcvrSig.X.RoleQN)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		newViaID := id.New()
		sndrViaBnd := procroot.Bnd{
			ProcID:  procCfg.ProcID,
			ChnlPH:  termSpec.X,
			ChnlID:  newViaID,
			StateID: rcvrRole.StateID,
			PoolRN:  procCfg.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, sndrViaBnd)
		rcvrViaBnd := procroot.Bnd{
			ProcID:  rcvrLiab.ProcID,
			ChnlPH:  rcvrSig.X.ChnlPH,
			ChnlID:  newViaID,
			StateID: rcvrRole.StateID,
			PoolRN:  rcvrSnap.PoolRN.Next(),
		}
		procMod.Bnds = append(procMod.Bnds, rcvrViaBnd)
		for i, chnlPH := range termSpec.Ys {
			valChnl, ok := procCfg.Chnls[chnlPH]
			if !ok {
				err := procroot.ErrMissingChnl(chnlPH)
				s.log.Error("taking failed")
				return StepSpec{}, procroot.Mod{}, err
			}
			sndrValBnd := procroot.Bnd{
				ProcID: procCfg.ProcID,
				ChnlPH: chnlPH,
				PoolRN: -procCfg.PoolRN.Next(),
			}
			procMod.Bnds = append(procMod.Bnds, sndrValBnd)
			rcvrValBnd := procroot.Bnd{
				ProcID:  rcvrLiab.ProcID,
				ChnlPH:  rcvrSig.Ys[i].ChnlPH,
				ChnlID:  valChnl.ChnlID,
				StateID: valChnl.StateID,
				PoolRN:  rcvrSnap.PoolRN.Next(),
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
	case step.FwdSpec:
		viaChnl, ok := procCfg.Chnls[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.X)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		viaAttr := slog.Any("chnlID", viaChnl.ChnlID)
		viaState, ok := procEnv.States[viaChnl.StateID]
		if !ok {
			err := state.ErrMissingInEnv(viaChnl.StateID)
			s.log.Error("taking failed", viaAttr)
			return StepSpec{}, procroot.Mod{}, err
		}
		valChnl, ok := procCfg.Chnls[termSpec.Y]
		if !ok {
			err := chnl.ErrMissingInCfg(termSpec.Y)
			s.log.Error("taking failed")
			return StepSpec{}, procroot.Mod{}, err
		}
		vs := procCfg.Steps[viaChnl.ChnlID]
		switch viaState.Pol() {
		case pol.Pos:
			switch viaStep := vs.(type) {
			case step.SvcRoot:
				xBnd := procroot.Bnd{
					ProcID:  viaStep.ProcID,
					ChnlPH:  viaStep.Cont.Via(),
					ChnlID:  viaChnl.ChnlID,
					StateID: viaChnl.StateID,
					PoolRN:  viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				tranSpec = StepSpec{
					PoolID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					Term:   viaStep.Cont,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case step.MsgRoot2:
				yBnd := procroot.Bnd{
					ProcID:  viaStep.ProcID,
					ChnlPH:  viaStep.Val.Via(),
					ChnlID:  valChnl.ChnlID,
					StateID: valChnl.StateID,
					PoolRN:  viaStep.PoolRN.Next(),
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
				xBnd := procroot.Bnd{
					ProcID: procCfg.ProcID,
					ChnlPH: termSpec.X,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, xBnd)
				yBnd := procroot.Bnd{
					ProcID: procCfg.ProcID,
					ChnlPH: termSpec.Y,
					PoolRN: -procCfg.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				msgStep := step.MsgRoot2{
					PoolID: procCfg.PoolID,
					ProcID: procCfg.ProcID,
					ChnlID: viaChnl.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					Val: step.FwdImpl{
						B: valChnl.ChnlID,
					},
				}
				procMod.Steps = append(procMod.Steps, msgStep)
				s.log.Debug("taking half done", viaAttr)
				return tranSpec, procMod, nil
			default:
				panic(step.ErrRootTypeUnexpected(vs))
			}
		case pol.Neg:
			switch viaStep := vs.(type) {
			case step.SvcRoot:
				yBnd := procroot.Bnd{
					ProcID:  viaStep.ProcID,
					ChnlPH:  viaStep.Cont.Via(),
					ChnlID:  valChnl.ChnlID,
					StateID: valChnl.StateID,
					PoolRN:  viaStep.PoolRN.Next(),
				}
				procMod.Bnds = append(procMod.Bnds, yBnd)
				tranSpec = StepSpec{
					PoolID: viaStep.PoolID,
					ProcID: viaStep.ProcID,
					Term:   viaStep.Cont,
				}
				s.log.Debug("taking succeeded", viaAttr)
				return tranSpec, procMod, nil
			case step.MsgRoot2:
				xBnd := procroot.Bnd{
					ProcID:  viaStep.ProcID,
					ChnlPH:  viaStep.Val.Via(),
					ChnlID:  viaChnl.ChnlID,
					StateID: viaChnl.StateID,
					PoolRN:  viaStep.PoolRN.Next(),
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
				svcStep := step.SvcRoot{
					PoolID: procCfg.PoolID,
					ProcID: procCfg.ProcID,
					ChnlID: viaChnl.ChnlID,
					PoolRN: procCfg.PoolRN.Next(),
					Cont: step.FwdImpl{
						B: valChnl.ChnlID,
					},
				}
				procMod.Steps = append(procMod.Steps, svcStep)
				s.log.Debug("taking half done", viaAttr)
				return tranSpec, procMod, nil
			default:
				panic(step.ErrRootTypeUnexpected(vs))
			}
		default:
			panic(state.ErrPolarityUnexpected(viaState))
		}
	default:
		panic(step.ErrTermTypeUnexpected(ts))
	}
}

func (s *service) Retrieve(poolID id.ADT) (snap Snap, err error) {
	ctx := context.Background()
	err = s.operator.Implicit(ctx, func(ds data.Source) error {
		snap, err = s.pools.SelectSubs(ds, poolID)
		return err
	})
	if err != nil {
		s.log.Error("retrieval failed", slog.Any("poolID", poolID))
		return Snap{}, err
	}
	return snap, nil
}

func (s *service) RetreiveRefs() (refs []Ref, err error) {
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

func CollectCtx(chnls []procroot.EP) []state.ID {
	return nil
}

func convertToCtx(poolID id.ADT, chnls []procroot.EP, states map[state.ID]state.Root) state.Context {
	assets := make(map[sym.ADT]state.Root, len(chnls)-1)
	liabs := make(map[sym.ADT]state.Root, 1)
	for _, ch := range chnls {
		if poolID == ch.PoolID {
			liabs[ch.ChnlPH] = states[ch.StateID]
		} else {
			assets[ch.ChnlPH] = states[ch.StateID]
		}
	}
	return state.Context{Assets: assets, Liabs: liabs}
}

func (s *service) checkState(
	poolID id.ADT,
	procEnv procroot.Env,
	procCtx state.Context,
	procCfg procroot.Cfg,
	termSpec step.Term,
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
	procEnv procroot.Env,
	procCtx state.Context,
	procCfg procroot.Cfg,
	ts step.Term,
) error {
	switch termSpec := ts.(type) {
	case step.CloseSpec:
		// check ctx
		if len(procCtx.Assets) > 0 {
			err := fmt.Errorf("context mismatch: want 0 items, got %v items", len(procCtx.Assets))
			s.log.Error("checking failed")
			return err
		}
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := state.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		err := state.CheckRoot(gotVia, state.OneRoot{})
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		delete(procCtx.Liabs, termSpec.X)
		return nil
	case step.WaitSpec:
		err := step.ErrTermTypeMismatch(ts, step.CloseSpec{})
		s.log.Error("checking failed")
		return err
	case step.SendSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := state.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.TensorRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := state.CheckRoot(gotVal, wantVia.B)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// no cont to check
		procCtx.Liabs[termSpec.X] = wantVia.C
		delete(procCtx.Assets, termSpec.Y)
		return nil
	case step.RecvSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := state.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.LolliRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := state.CheckRoot(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Liabs[termSpec.X] = wantVia.Z
		procCtx.Assets[termSpec.Y] = wantVia.Y
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	case step.LabSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := state.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.PlusRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
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
	case step.CaseSpec:
		// check via
		gotVia, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := state.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.WithRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
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
	case step.FwdSpec:
		if len(procCtx.Assets) != 1 {
			err := fmt.Errorf("context mismatch: want 1 item, got %v items", len(procCtx.Assets))
			s.log.Error("checking failed")
			return err
		}
		viaSt, ok := procCtx.Liabs[termSpec.X]
		if !ok {
			err := state.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		fwdSt, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		if fwdSt.Pol() != viaSt.Pol() {
			err := state.ErrPolarityMismatch(fwdSt, viaSt)
			s.log.Error("checking failed")
			return err
		}
		err := state.CheckRoot(fwdSt, viaSt)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		delete(procCtx.Liabs, termSpec.X)
		delete(procCtx.Assets, termSpec.Y)
		return nil
	default:
		panic(step.ErrTermTypeUnexpected(ts))
	}
}

func (s *service) checkClient(
	poolID id.ADT,
	procEnv procroot.Env,
	procCtx state.Context,
	procCfg procroot.Cfg,
	ts step.Term,
) error {
	switch termSpec := ts.(type) {
	case step.CloseSpec:
		err := step.ErrTermTypeMismatch(ts, step.WaitSpec{})
		s.log.Error("checking failed")
		return err
	case step.WaitSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.OneRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check cont
		delete(procCtx.Assets, termSpec.X)
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	case step.SendSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.LolliRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := state.CheckRoot(gotVal, wantVia.Y)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		procCtx.Assets[termSpec.X] = wantVia.Z
		delete(procCtx.Assets, termSpec.Y)
		return nil
	case step.RecvSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.TensorRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
			s.log.Error("checking failed")
			return err
		}
		// check value
		gotVal, ok := procCtx.Assets[termSpec.Y]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.Y)
			s.log.Error("checking failed")
			return err
		}
		err := state.CheckRoot(gotVal, wantVia.B)
		if err != nil {
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[termSpec.X] = wantVia.C
		procCtx.Assets[termSpec.Y] = wantVia.B
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	case step.LabSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.WithRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
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
	case step.CaseSpec:
		// check via
		gotVia, ok := procCtx.Assets[termSpec.X]
		if !ok {
			err := chnl.ErrMissingInCtx(termSpec.X)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := gotVia.(state.PlusRoot)
		if !ok {
			err := state.ErrRootTypeMismatch(gotVia, wantVia)
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
	case step.SpawnSpec:
		procSig, ok := procEnv.Sigs[termSpec.SigID]
		if !ok {
			err := procsig.ErrRootMissingInEnv(termSpec.SigID)
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
			valRole, ok := procEnv.Roles[ep.RoleQN]
			if !ok {
				err := rolesig.ErrMissingInEnv(ep.RoleQN)
				s.log.Error("checking failed")
				return err
			}
			wantVal, ok := procEnv.States[valRole.StateID]
			if !ok {
				err := state.ErrMissingInEnv(valRole.StateID)
				s.log.Error("checking failed")
				return err
			}
			gotVal, ok := procCtx.Assets[termSpec.Ys[i]]
			if !ok {
				err := chnl.ErrMissingInCtx(ep.ChnlPH)
				s.log.Error("checking failed")
				return err
			}
			err := state.CheckRoot(gotVal, wantVal)
			if err != nil {
				s.log.Error("checking failed", slog.Any("want", wantVal), slog.Any("got", gotVal))
				return err
			}
			delete(procCtx.Assets, termSpec.Ys[i])
		}
		// check via
		viaRole, ok := procEnv.Roles[procSig.X.RoleQN]
		if !ok {
			err := rolesig.ErrMissingInEnv(procSig.X.RoleQN)
			s.log.Error("checking failed")
			return err
		}
		wantVia, ok := procEnv.States[viaRole.StateID]
		if !ok {
			err := state.ErrMissingInEnv(viaRole.StateID)
			s.log.Error("checking failed")
			return err
		}
		// check cont
		procCtx.Assets[termSpec.X] = wantVia
		return s.checkState(poolID, procEnv, procCtx, procCfg, termSpec.Cont)
	default:
		panic(step.ErrTermTypeUnexpected(ts))
	}
}

// Port
type Repo interface {
	Insert(data.Source, impl) error
	InsertLiab(data.Source, procroot.Liab) error
	SelectRefs(data.Source) ([]Ref, error)
	SelectSubs(data.Source, id.ADT) (Snap, error)
	SelectProc(data.Source, id.ADT) (procroot.Cfg, error)
	UpdateProc(data.Source, procroot.Mod) error
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
var (
	ConvertRootToRef func(impl) Ref
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
