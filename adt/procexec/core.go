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
	"orglang/go-runtime/adt/procbind"
	"orglang/go-runtime/adt/procdec"
	"orglang/go-runtime/adt/procdef"
	"orglang/go-runtime/adt/procexp"
	"orglang/go-runtime/adt/procstep"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/typedef"
	"orglang/go-runtime/adt/typeexp"
	"orglang/go-runtime/adt/uniqref"
	"orglang/go-runtime/adt/uniqsym"
)

type API interface {
	Take(procstep.StepSpec) error
	RetrieveSnap(ExecRef) (ExecSnap, error)
}

type ExecSpec struct {
	ProcQN     uniqsym.ADT
	ValChnlIDs []identity.ADT
	ProcES     procexp.ExpSpec
}

type ExecRef = uniqref.ADT

// aka Configuration
type ExecSnap struct {
	ExecRef ExecRef
	ChnlBRs map[symbol.ADT]procbind.BindRec
	ProcSRs map[identity.ADT]procstep.StepRec
}

type Env struct {
	TypeDefs map[uniqsym.ADT]typedef.DefRec
	TypeExps map[identity.ADT]typeexp.ExpRec
	ProcDecs map[identity.ADT]procdec.DecRec
}

func ChnlPH(rec procbind.BindRec) symbol.ADT { return rec.ChnlPH }

type ExecMod struct {
	Locks []ExecRef
	Binds []procbind.BindRec
	Steps []procstep.StepRec
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

func (s *service) RetrieveSnap(ref ExecRef) (_ ExecSnap, err error) {
	return ExecSnap{}, nil
}

func ErrMissingChnl(want symbol.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func (s *service) Take(spec procstep.StepSpec) (err error) {
	refAttr := slog.Any("execRef", spec.ExecRef)
	s.log.Debug("taking started", refAttr)
	ctx := context.Background()
	// initial values
	execRef := spec.ExecRef
	expSpec := spec.ProcES
	for expSpec != nil {
		var execSnap ExecSnap
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			execSnap, err = s.procExecs.SelectSnap(ds, execRef)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", refAttr)
			return err
		}
		if len(execSnap.ChnlBRs) == 0 {
			panic("zero channel binds")
		}
		decIDs := procexp.CollectEnv(expSpec)
		var procDRs map[identity.ADT]procdec.DecRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			procDRs, err = s.procDecs.SelectEnv(ds, decIDs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", refAttr, slog.Any("decs", decIDs))
			return err
		}
		typeQNs := procdec.CollectEnv(maps.Values(procDRs))
		var typeDefs map[uniqsym.ADT]typedef.DefRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			typeDefs, err = s.typeDefs.SelectEnv(ds, typeQNs)
			return err
		})
		if err != nil {
			s.log.Error("taking failed", refAttr, slog.Any("types", typeQNs))
			return err
		}
		envIDs := typedef.CollectEnv(maps.Values(typeDefs))
		ctxIDs := CollectCtx(maps.Values(execSnap.ChnlBRs))
		var typeExps map[identity.ADT]typeexp.ExpRec
		err = s.operator.Implicit(ctx, func(ds db.Source) error {
			typeExps, err = s.typeExps.SelectEnv(ds, append(envIDs, ctxIDs...))
			return err
		})
		if err != nil {
			s.log.Error("taking failed", refAttr, slog.Any("env", envIDs), slog.Any("ctx", ctxIDs))
			return err
		}
		procEnv := Env{ProcDecs: procDRs, TypeDefs: typeDefs, TypeExps: typeExps}
		procCtx := convertToCtx(maps.Values(execSnap.ChnlBRs), typeExps)
		// type checking
		err = s.checkType(procEnv, procCtx, execSnap, expSpec)
		if err != nil {
			s.log.Error("taking failed", refAttr)
			return err
		}
		// step taking
		nextSpec, procMod, err := s.takeWith(procEnv, execSnap, expSpec)
		if err != nil {
			s.log.Error("taking failed", refAttr)
			return err
		}
		err = s.operator.Explicit(ctx, func(ds db.Source) error {
			err = s.procExecs.UpdateProc(ds, procMod)
			if err != nil {
				s.log.Error("taking failed", refAttr)
				return err
			}
			return nil
		})
		if err != nil {
			s.log.Error("taking failed", refAttr)
			return err
		}
		// next values
		execRef = nextSpec.ExecRef
		expSpec = nextSpec.ProcES
	}
	s.log.Debug("taking succeed", refAttr)
	return nil
}

func (s *service) takeWith(
	procEnv Env,
	execSnap ExecSnap,
	es procexp.ExpSpec,
) (
	stepSpec procstep.StepSpec,
	execMod ExecMod,
	_ error,
) {
	switch expSpec := es.(type) {
	case procexp.CloseSpec:
		commChnlBR, ok := execSnap.ChnlBRs[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlBR.ChnlID)
		execMod.Locks = append(execMod.Locks, execSnap.ExecRef)
		recieverSR := execSnap.ProcSRs[commChnlBR.ChnlID]
		if recieverSR == nil {
			senderSR := procstep.MsgRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlID: commChnlBR.ChnlID,
				ValER: procexp.CloseRec{
					CommChnlPH: expSpec.CommChnlPH,
				},
			}
			execMod.Steps = append(execMod.Steps, senderSR)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, execMod, nil
		}
		serviceSR, ok := recieverSR.(procstep.SvcRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(recieverSR))
		}
		switch procER := serviceSR.ContER.(type) {
		case procexp.WaitRec:
			senderBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: -execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
			}
			execMod.Binds = append(execMod.Binds, senderBR)
			recieverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: serviceSR.ExecRef.ID,
					RN: -serviceSR.ExecRef.RN.Next(),
				},
				ChnlPH: procER.CommChnlPH,
			}
			execMod.Binds = append(execMod.Binds, recieverBR)
			stepSpec = procstep.StepSpec{
				ExecRef: serviceSR.ExecRef,
				ProcES:  procER.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, execMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(serviceSR.ContER))
		}
	case procexp.WaitSpec:
		commChnlBR, ok := execSnap.ChnlBRs[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlBR.ChnlID)
		execMod.Locks = append(execMod.Locks, execSnap.ExecRef)
		senderSR := execSnap.ProcSRs[commChnlBR.ChnlID]
		if senderSR == nil {
			recieverSR := procstep.SvcRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlID: commChnlBR.ChnlID,
				ContER: procexp.WaitRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContES:     expSpec.ContES,
				},
			}
			execMod.Steps = append(execMod.Steps, recieverSR)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, execMod, nil
		}
		messageSR, ok := senderSR.(procstep.MsgRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(senderSR))
		}
		switch procER := messageSR.ValER.(type) {
		case procexp.CloseRec:
			senderBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: messageSR.ExecRef.ID,
					RN: -messageSR.ExecRef.RN.Next(),
				},
				ChnlPH: procER.CommChnlPH,
			}
			execMod.Binds = append(execMod.Binds, senderBR)
			recieverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: -execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
			}
			execMod.Binds = append(execMod.Binds, recieverBR)
			stepSpec = procstep.StepSpec{
				ExecRef: execSnap.ExecRef,
				ProcES:  expSpec.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, execMod, nil
		case procexp.FwdRec:
			recieverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: procER.ContChnlID,
				ExpID:  commChnlBR.ExpID,
			}
			execMod.Binds = append(execMod.Binds, recieverBR)
			stepSpec = procstep.StepSpec{
				ExecRef: execSnap.ExecRef,
				ProcES:  expSpec,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, execMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(messageSR.ValER))
		}
	case procexp.SendSpec:
		commChnlBR, ok := execSnap.ChnlBRs[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlBR.ChnlID)
		execMod.Locks = append(execMod.Locks, execSnap.ExecRef)
		typeER, ok := procEnv.TypeExps[commChnlBR.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(commChnlBR.ExpID)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, ExecMod{}, err
		}
		nextExpID := typeER.(typeexp.ProdRec).Next()
		valueEP, ok := execSnap.ChnlBRs[expSpec.ValChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.ValChnlPH)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, ExecMod{}, err
		}
		senderBR := procbind.BindRec{
			ExecRef: ExecRef{
				ID: execSnap.ExecRef.ID,
				RN: -execSnap.ExecRef.RN.Next(),
			},
			ChnlPH: expSpec.ValChnlPH,
		}
		execMod.Binds = append(execMod.Binds, senderBR)
		recieverSR := execSnap.ProcSRs[commChnlBR.ChnlID]
		if recieverSR == nil {
			newChnlID := identity.New()
			senderBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: newChnlID,
				ExpID:  nextExpID,
			}
			execMod.Binds = append(execMod.Binds, senderBR)
			senderSR := procstep.MsgRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlID: commChnlBR.ChnlID,
				ValER: procexp.SendRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: newChnlID,
					ValChnlID:  valueEP.ChnlID,
					ValExpID:   valueEP.ExpID,
				},
			}
			execMod.Steps = append(execMod.Steps, senderSR)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, execMod, nil
		}
		serviceSR, ok := recieverSR.(procstep.SvcRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(recieverSR))
		}
		switch expRec := serviceSR.ContER.(type) {
		case procexp.RecvRec:
			senderBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: expRec.ContChnlID,
				ExpID:  nextExpID,
			}
			execMod.Binds = append(execMod.Binds, senderBR)
			recieverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: serviceSR.ExecRef.ID,
					RN: serviceSR.ExecRef.RN.Next(),
				},
				ChnlPH: expRec.CommChnlPH,
				ChnlID: expRec.ContChnlID,
				ExpID:  nextExpID,
			}
			execMod.Binds = append(execMod.Binds, recieverBR)
			receiverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: serviceSR.ExecRef.ID,
					RN: serviceSR.ExecRef.RN.Next(),
				},
				ChnlPH: expRec.ValChnlPH,
				ChnlID: valueEP.ChnlID,
				ExpID:  valueEP.ExpID,
			}
			execMod.Binds = append(execMod.Binds, receiverBR)
			stepSpec = procstep.StepSpec{
				ExecRef: serviceSR.ExecRef,
				ProcES:  expRec.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, execMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(serviceSR.ContER))
		}
	case procexp.RecvSpec:
		commChnlBR, ok := execSnap.ChnlBRs[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlBR.ChnlID)
		execMod.Locks = append(execMod.Locks, execSnap.ExecRef)
		senderSR := execSnap.ProcSRs[commChnlBR.ChnlID]
		if senderSR == nil {
			receiverSR := procstep.SvcRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlID: commChnlBR.ChnlID,
				ContER: procexp.RecvRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: identity.New(),
					ValChnlPH:  expSpec.BindChnlPH,
					ContES:     expSpec.ContES,
				},
			}
			execMod.Steps = append(execMod.Steps, receiverSR)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, execMod, nil
		}
		sndrMsgRec, ok := senderSR.(procstep.MsgRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(senderSR))
		}
		switch expRec := sndrMsgRec.ValER.(type) {
		case procexp.SendRec:
			typeER, ok := procEnv.TypeExps[commChnlBR.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(commChnlBR.ExpID)
				s.log.Error("taking failed", viaAttr)
				return procstep.StepSpec{}, ExecMod{}, err
			}
			recieverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: expRec.ContChnlID,
				ExpID:  typeER.(typeexp.ProdRec).Next(),
			}
			execMod.Binds = append(execMod.Binds, recieverBR)
			receiverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.BindChnlPH,
				ChnlID: expRec.ValChnlID,
				ExpID:  expRec.ValExpID,
			}
			execMod.Binds = append(execMod.Binds, receiverBR)
			stepSpec = procstep.StepSpec{
				ExecRef: execSnap.ExecRef,
				ProcES:  expSpec.ContES,
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, execMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(sndrMsgRec.ValER))
		}
	case procexp.LabSpec:
		commChnlBR, ok := execSnap.ChnlBRs[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlBR.ChnlID)
		execMod.Locks = append(execMod.Locks, execSnap.ExecRef)
		typeER, ok := procEnv.TypeExps[commChnlBR.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(commChnlBR.ExpID)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, ExecMod{}, err
		}
		nextExpID := typeER.(typeexp.SumRec).Next(expSpec.LabelQN)
		recieverSR := execSnap.ProcSRs[commChnlBR.ChnlID]
		if recieverSR == nil {
			newViaID := identity.New()
			senderBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: newViaID,
				ExpID:  nextExpID,
			}
			execMod.Binds = append(execMod.Binds, senderBR)
			senderSR := procstep.MsgRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlID: commChnlBR.ChnlID,
				ValER: procexp.LabRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: newViaID,
					LabelQN:    expSpec.LabelQN,
				},
			}
			execMod.Steps = append(execMod.Steps, senderSR)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, execMod, nil
		}
		serviceSR, ok := recieverSR.(procstep.SvcRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(recieverSR))
		}
		switch expRec := serviceSR.ContER.(type) {
		case procexp.CaseRec:
			senderBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: expRec.ContChnlID,
				ExpID:  nextExpID,
			}
			execMod.Binds = append(execMod.Binds, senderBR)
			recieverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: serviceSR.ExecRef.ID,
					RN: serviceSR.ExecRef.RN.Next(),
				},
				ChnlPH: expRec.CommChnlPH,
				ChnlID: expRec.ContChnlID,
				ExpID:  nextExpID,
			}
			execMod.Binds = append(execMod.Binds, recieverBR)
			stepSpec = procstep.StepSpec{
				ExecRef: serviceSR.ExecRef,
				ProcES:  expRec.ContESs[expSpec.LabelQN],
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, execMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(serviceSR.ContER))
		}
	case procexp.CaseSpec:
		commChnlBR, ok := execSnap.ChnlBRs[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlBR.ChnlID)
		execMod.Locks = append(execMod.Locks, execSnap.ExecRef)
		senderSR := execSnap.ProcSRs[commChnlBR.ChnlID]
		if senderSR == nil {
			recieverSR := procstep.SvcRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlID: commChnlBR.ChnlID,
				ContER: procexp.CaseRec{
					CommChnlPH: expSpec.CommChnlPH,
					ContChnlID: identity.New(),
					ContESs:    expSpec.ContESs,
				},
			}
			execMod.Steps = append(execMod.Steps, recieverSR)
			s.log.Debug("taking half done", viaAttr)
			return stepSpec, execMod, nil
		}
		messageSR, ok := senderSR.(procstep.MsgRec)
		if !ok {
			panic(procstep.ErrRecTypeUnexpected(senderSR))
		}
		switch procER := messageSR.ValER.(type) {
		case procexp.LabRec:
			typeER, ok := procEnv.TypeExps[commChnlBR.ExpID]
			if !ok {
				err := typedef.ErrMissingInEnv(commChnlBR.ExpID)
				s.log.Error("taking failed", viaAttr)
				return procstep.StepSpec{}, ExecMod{}, err
			}
			recieverBR := procbind.BindRec{
				ExecRef: ExecRef{
					ID: execSnap.ExecRef.ID,
					RN: execSnap.ExecRef.RN.Next(),
				},
				ChnlPH: expSpec.CommChnlPH,
				ChnlID: procER.ContChnlID,
				ExpID:  typeER.(typeexp.SumRec).Next(procER.LabelQN),
			}
			execMod.Binds = append(execMod.Binds, recieverBR)
			stepSpec = procstep.StepSpec{
				ExecRef: execSnap.ExecRef,
				ProcES:  expSpec.ContESs[procER.LabelQN],
			}
			s.log.Debug("taking succeed", viaAttr)
			return stepSpec, execMod, nil
		default:
			panic(procexp.ErrRecTypeUnexpected(messageSR.ValER))
		}
	case procexp.FwdSpec:
		commChnlBR, ok := execSnap.ChnlBRs[expSpec.CommChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.CommChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		viaAttr := slog.Any("chnlID", commChnlBR.ChnlID)
		commChnlER, ok := procEnv.TypeExps[commChnlBR.ExpID]
		if !ok {
			err := typedef.ErrMissingInEnv(commChnlBR.ExpID)
			s.log.Error("taking failed", viaAttr)
			return procstep.StepSpec{}, ExecMod{}, err
		}
		valChnlBR, ok := execSnap.ChnlBRs[expSpec.ContChnlPH]
		if !ok {
			err := procdef.ErrMissingInCfg(expSpec.ContChnlPH)
			s.log.Error("taking failed")
			return procstep.StepSpec{}, ExecMod{}, err
		}
		commChnlSR := execSnap.ProcSRs[commChnlBR.ChnlID]
		switch commChnlER.Pol() {
		case polarity.Pos:
			switch stepRec := commChnlSR.(type) {
			case procstep.SvcRec:
				xBnd := procbind.BindRec{
					ExecRef: ExecRef{
						ID: stepRec.ExecRef.ID,
						RN: stepRec.ExecRef.RN.Next(),
					},
					ChnlPH: stepRec.ContER.Via(),
					ChnlID: commChnlBR.ChnlID,
					ExpID:  commChnlBR.ExpID,
				}
				execMod.Binds = append(execMod.Binds, xBnd)
				stepSpec = procstep.StepSpec{
					ExecRef: stepRec.ExecRef,
					ProcES:  stepRec.ContER,
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, execMod, nil
			case procstep.MsgRec:
				yBnd := procbind.BindRec{
					ExecRef: ExecRef{
						ID: stepRec.ExecRef.ID,
						RN: stepRec.ExecRef.RN.Next(),
					},
					ChnlPH: stepRec.ValER.Via(),
					ChnlID: valChnlBR.ChnlID,
					ExpID:  valChnlBR.ExpID,
				}
				execMod.Binds = append(execMod.Binds, yBnd)
				stepSpec = procstep.StepSpec{
					ExecRef: stepRec.ExecRef,
					ProcES:  stepRec.ValER,
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, execMod, nil
			case nil:
				xBnd := procbind.BindRec{
					ExecRef: ExecRef{
						ID: execSnap.ExecRef.ID,
						RN: -execSnap.ExecRef.RN.Next(),
					},
					ChnlPH: expSpec.CommChnlPH,
				}
				execMod.Binds = append(execMod.Binds, xBnd)
				yBnd := procbind.BindRec{
					ExecRef: ExecRef{
						ID: execSnap.ExecRef.ID,
						RN: -execSnap.ExecRef.RN.Next(),
					},
					ChnlPH: expSpec.ContChnlPH,
				}
				execMod.Binds = append(execMod.Binds, yBnd)
				messageSR := procstep.MsgRec{
					ExecRef: ExecRef{
						ID: execSnap.ExecRef.ID,
						RN: execSnap.ExecRef.RN.Next(),
					},
					ChnlID: commChnlBR.ChnlID,
					ValER: procexp.FwdRec{
						ContChnlID: valChnlBR.ChnlID,
					},
				}
				execMod.Steps = append(execMod.Steps, messageSR)
				s.log.Debug("taking half done", viaAttr)
				return stepSpec, execMod, nil
			default:
				panic(procstep.ErrRecTypeUnexpected(commChnlSR))
			}
		case polarity.Neg:
			switch stepRec := commChnlSR.(type) {
			case procstep.SvcRec:
				yBnd := procbind.BindRec{
					ExecRef: ExecRef{
						ID: stepRec.ExecRef.ID,
						RN: stepRec.ExecRef.RN.Next(),
					},
					ChnlPH: stepRec.ContER.Via(),
					ChnlID: valChnlBR.ChnlID,
					ExpID:  valChnlBR.ExpID,
				}
				execMod.Binds = append(execMod.Binds, yBnd)
				stepSpec = procstep.StepSpec{
					ExecRef: stepRec.ExecRef,
					ProcES:  stepRec.ContER,
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, execMod, nil
			case procstep.MsgRec:
				xBnd := procbind.BindRec{
					ExecRef: ExecRef{
						ID: stepRec.ExecRef.ID,
						RN: stepRec.ExecRef.RN.Next(),
					},
					ChnlPH: stepRec.ValER.Via(),
					ChnlID: commChnlBR.ChnlID,
					ExpID:  commChnlBR.ExpID,
				}
				execMod.Binds = append(execMod.Binds, xBnd)
				stepSpec = procstep.StepSpec{
					ExecRef: stepRec.ExecRef,
					ProcES:  stepRec.ValER, // TODO: несовпадение типов
				}
				s.log.Debug("taking succeed", viaAttr)
				return stepSpec, execMod, nil
			case nil:
				serviceSR := procstep.SvcRec{
					ExecRef: ExecRef{
						ID: execSnap.ExecRef.ID,
						RN: execSnap.ExecRef.RN.Next(),
					},
					ChnlID: commChnlBR.ChnlID,
					ContER: procexp.FwdRec{
						ContChnlID: valChnlBR.ChnlID,
					},
				}
				execMod.Steps = append(execMod.Steps, serviceSR)
				s.log.Debug("taking half done", viaAttr)
				return stepSpec, execMod, nil
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

func CollectCtx(chnls iter.Seq[procbind.BindRec]) []identity.ADT {
	return nil
}

func convertToCtx(chnlBinds iter.Seq[procbind.BindRec], typeExps map[identity.ADT]typeexp.ExpRec) typedef.Context {
	assets := make(map[symbol.ADT]typeexp.ExpRec, 1)
	liabs := make(map[symbol.ADT]typeexp.ExpRec, 1)
	for bind := range chnlBinds {
		if bind.ChnlBS == procbind.ProviderSide {
			liabs[bind.ChnlPH] = typeExps[bind.ExpID]
		} else {
			assets[bind.ChnlPH] = typeExps[bind.ExpID]
		}
	}
	return typedef.Context{Assets: assets, Liabs: liabs}
}

func (s *service) checkType(
	procEnv Env,
	procCtx typedef.Context,
	execSnap ExecSnap,
	expSpec procexp.ExpSpec,
) error {
	chnlBR, ok := execSnap.ChnlBRs[expSpec.Via()]
	if !ok {
		panic("no comm chnl in proc snap")
	}
	if chnlBR.ChnlBS == procbind.ProviderSide {
		return s.checkProvider(procEnv, procCtx, execSnap, expSpec)
	} else {
		return s.checkClient(procEnv, procCtx, execSnap, expSpec)
	}
}

func (s *service) checkProvider(
	procEnv Env,
	procCtx typedef.Context,
	procCfg ExecSnap,
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
		return s.checkType(procEnv, procCtx, procCfg, expSpec.ContES)
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
			err := s.checkType(procEnv, procCtx, procCfg, cont)
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
	procEnv Env,
	procCtx typedef.Context,
	procCfg ExecSnap,
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
		return s.checkType(procEnv, procCtx, procCfg, expSpec.ContES)
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
		return s.checkType(procEnv, procCtx, procCfg, expSpec.ContES)
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
			err := s.checkType(procEnv, procCtx, procCfg, cont)
			if err != nil {
				s.log.Error("checking failed")
				return err
			}
		}
		return nil
	case procexp.SpawnSpecOld:
		procDec, ok := procEnv.ProcDecs[expSpec.SigID]
		if !ok {
			err := procdec.ErrRootMissingInEnv(expSpec.SigID)
			s.log.Error("checking failed")
			return err
		}
		// check vals
		if len(expSpec.Ys) != len(procDec.ClientBSs) {
			err := fmt.Errorf("context mismatch: want %v items, got %v items", len(procDec.ClientBSs), len(expSpec.Ys))
			s.log.Error("checking failed", slog.Any("want", procDec.ClientBSs), slog.Any("got", expSpec.Ys))
			return err
		}
		if len(expSpec.Ys) == 0 {
			return nil
		}
		for i, ep := range procDec.ClientBSs {
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
				err := procdef.ErrMissingInCtx(ep.ChnlPH)
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
		viaRole, ok := procEnv.TypeDefs[procDec.ProviderBS.TypeQN]
		if !ok {
			err := typedef.ErrSymMissingInEnv(procDec.ProviderBS.TypeQN)
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
		return s.checkType(procEnv, procCtx, procCfg, expSpec.ContES)
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
