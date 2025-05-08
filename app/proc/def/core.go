package def

import (
	"fmt"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"
)

type SemRec interface {
	step() id.ADT
}

func ChnlID(r SemRec) id.ADT { return r.step() }

type MsgRec struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Val    TermVal
	PoolRN rn.ADT
	ProcRN rn.ADT
}

func (r MsgRec) step() id.ADT { return r.ChnlID }

type SvcRec struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Cont   TermCont
	PoolRN rn.ADT
}

func (r SvcRec) step() id.ADT { return r.ChnlID }

type TermSpec interface {
	Via() sym.ADT
}

// aka ast.Msg
type Value interface {
	TermSpec
	val()
}

type Continuation interface {
	TermSpec
	cont()
}

type CloseSpec struct {
	X sym.ADT
}

func (s CloseSpec) Via() sym.ADT { return s.X }

func (CloseSpec) val() {}

type WaitSpec struct {
	X    sym.ADT
	Cont TermSpec
}

func (s WaitSpec) Via() sym.ADT { return s.X }

func (WaitSpec) cont() {}

type SendSpec struct {
	X sym.ADT // via
	Y sym.ADT // val
}

func (s SendSpec) Via() sym.ADT { return s.X }

func (SendSpec) val() {}

type RecvSpec struct {
	X    sym.ADT // via
	Y    sym.ADT // val
	Cont TermSpec
}

func (s RecvSpec) Via() sym.ADT { return s.X }

func (RecvSpec) cont() {}

type LabSpec struct {
	X     sym.ADT
	Label sym.ADT
}

func (s LabSpec) Via() sym.ADT { return s.X }

func (LabSpec) val() {}

type CaseSpec struct {
	X     sym.ADT
	Conts map[sym.ADT]TermSpec
}

func (s CaseSpec) Via() sym.ADT { return s.X }

func (CaseSpec) cont() {}

// aka ExpName
type LinkSpec struct {
	SigQN sym.ADT
	X     id.ADT
	Ys    []id.ADT
}

func (s LinkSpec) Via() sym.ADT { return "" }

type FwdSpec struct {
	X sym.ADT // old via (from)
	Y sym.ADT // new via (to)
}

func (s FwdSpec) Via() sym.ADT { return s.X }

func (FwdSpec) val() {}

func (FwdSpec) cont() {}

// аналог SendSpec, но значения отправляются балком
type CallSpec struct {
	X     sym.ADT
	SigPH sym.ADT // import
	Ys    []sym.ADT
}

func (s CallSpec) Via() sym.ADT { return s.SigPH }

// аналог RecvSpec, но значения принимаются балком
type SpawnSpec struct {
	X      sym.ADT
	SigID  id.ADT
	Ys     []sym.ADT
	PoolQN sym.ADT
	Cont   TermSpec
}

func (s SpawnSpec) Via() sym.ADT { return s.X }

// аналог RecvSpec, но значения принимаются балком
type SpawnSpec2 struct {
	X     sym.ADT
	SigPH sym.ADT // export
	Cont  TermSpec
}

func (s SpawnSpec2) Via() sym.ADT { return s.SigPH }

type TermRec interface {
	TermSpec
	impl()
}

type TermVal interface {
	TermRec
	val2()
}

type TermCont interface {
	TermRec
	cont2()
}

type CloseRec struct {
	X sym.ADT
}

func (r CloseRec) Via() sym.ADT { return r.X }

func (CloseRec) impl() {}

func (CloseRec) val2() {}

type WaitRec struct {
	X    sym.ADT
	Cont TermSpec
}

func (r WaitRec) Via() sym.ADT { return r.X }

func (WaitRec) impl() {}

func (WaitRec) cont2() {}

type SendRec struct {
	X      sym.ADT
	A      id.ADT
	B      id.ADT
	TermID id.ADT
}

func (r SendRec) Via() sym.ADT { return r.X }

func (SendRec) impl() {}

func (SendRec) val2() {}

type RecvRec struct {
	X    sym.ADT
	A    id.ADT
	Y    sym.ADT
	Cont TermSpec
}

func (r RecvRec) Via() sym.ADT { return r.X }

func (RecvRec) impl() {}

func (RecvRec) cont2() {}

type LabRec struct {
	X     sym.ADT
	A     id.ADT
	Label sym.ADT
}

func (r LabRec) Via() sym.ADT { return r.X }

func (LabRec) impl() {}

func (LabRec) val2() {}

type CaseRec struct {
	X     sym.ADT
	A     id.ADT
	Conts map[sym.ADT]TermSpec
}

func (r CaseRec) Via() sym.ADT { return r.X }

func (CaseRec) impl() {}

func (CaseRec) cont2() {}

type FwdRec struct {
	X sym.ADT
	B id.ADT // to
}

func (r FwdRec) Via() sym.ADT { return r.X }

func (FwdRec) impl() {}

func (FwdRec) val2() {}

func (FwdRec) cont2() {}

type SemRepo interface {
	Insert(data.Source, ...SemRec) error
	SelectByID(data.Source, id.ADT) (SemRec, error)
	SelectByPID(data.Source, id.ADT) (SemRec, error)
	SelectByVID(data.Source, id.ADT) (SemRec, error)
}

func CollectEnv(spec TermSpec) []id.ADT {
	return collectEnvRec(spec, []id.ADT{})
}

func collectEnvRec(s TermSpec, env []id.ADT) []id.ADT {
	switch spec := s.(type) {
	case RecvSpec:
		return collectEnvRec(spec.Cont, env)
	case CaseSpec:
		for _, cont := range spec.Conts {
			env = collectEnvRec(cont, env)
		}
		return env
	case SpawnSpec:
		return collectEnvRec(spec.Cont, append(env, spec.SigID))
	default:
		return env
	}
}

func ErrDoesNotExist(want id.ADT) error {
	return fmt.Errorf("root doesn't exist: %v", want)
}

func ErrRootTypeUnexpected(got SemRec) error {
	return fmt.Errorf("root type unexpected: %T", got)
}

func ErrRootTypeMismatch(got, want SemRec) error {
	return fmt.Errorf("root type mismatch: want %T, got %T", want, got)
}

func ErrTermTypeUnexpected(got TermSpec) error {
	return fmt.Errorf("term type unexpected: %T", got)
}

func ErrImplTypeUnexpected(got TermRec) error {
	return fmt.Errorf("impl type unexpected: %T", got)
}

func ErrTermTypeMismatch(got, want TermSpec) error {
	return fmt.Errorf("term type mismatch: want %T, got %T", want, got)
}

func ErrTermValueNil(pid id.ADT) error {
	return fmt.Errorf("proc %q term is nil", pid)
}

func ErrValTypeUnexpected(got Value) error {
	return fmt.Errorf("value type unexpected: %T", got)
}

func ErrValTypeUnexpected2(got TermVal) error {
	return fmt.Errorf("value type unexpected: %T", got)
}

func ErrContTypeUnexpected(got Continuation) error {
	return fmt.Errorf("continuation type unexpected: %T", got)
}

func ErrContTypeUnexpected2(got TermCont) error {
	return fmt.Errorf("continuation type unexpected: %T", got)
}

func ErrMissingInCfg(want sym.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCfg2(want id.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCtx(want sym.ADT) error {
	return fmt.Errorf("channel missing in ctx: %v", want)
}
