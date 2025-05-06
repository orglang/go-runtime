package step

import (
	"fmt"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"
	"smecalculus/rolevod/lib/sym"
)

type ID = id.ADT

type Ref interface {
	id.Identifiable
}

type ProcRef struct {
	ID id.ADT
}

func (r ProcRef) Ident() id.ADT { return r.ID }

type MsgRef struct {
	ID id.ADT
}

func (r MsgRef) Ident() id.ADT { return r.ID }

type SrvRef struct {
	ID id.ADT
}

func (r SrvRef) Ident() id.ADT { return r.ID }

type Root interface {
	step() id.ADT
}

func ChnlID(r Root) id.ADT { return r.step() }

// aka exec.Proc
type ProcRoot struct {
	ID   id.ADT
	PID  id.ADT
	Term Term
}

func (r ProcRoot) step() id.ADT { return r.PID }

// aka exec.Msg
type MsgRoot struct {
	ID  id.ADT
	PID id.ADT
	VID id.ADT
	Val Value
}

func (r MsgRoot) step() id.ADT { return r.VID }

type MsgRoot2 struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Val    Val
	PoolRN rn.ADT
	ProcRN rn.ADT
}

func (r MsgRoot2) step() id.ADT { return r.ChnlID }

type SrvRoot struct {
	ID   id.ADT
	PID  id.ADT
	VID  id.ADT
	Cont Continuation
}

func (r SrvRoot) step() id.ADT { return r.VID }

type SvcRoot struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Cont   Cont
	PoolRN rn.ADT
}

func (r SvcRoot) step() id.ADT { return r.ChnlID }

// aka Expression or Term
type Term interface {
	Via() sym.ADT
}

type Impl interface {
	Term
	impl()
}

// aka ast.Msg
type Value interface {
	Term
	val()
}

type Val interface {
	Impl
	val2()
}

type Continuation interface {
	Term
	cont()
}

type Cont interface {
	Impl
	cont2()
}

type Action interface {
	Term
	act()
}

type CloseSpec struct {
	X sym.ADT
}

func (s CloseSpec) Via() sym.ADT { return s.X }

func (CloseSpec) val() {}

type WaitSpec struct {
	X    sym.ADT
	Cont Term
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
	Cont Term
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
	Conts map[sym.ADT]Term
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
	Cont   Term
}

func (s SpawnSpec) Via() sym.ADT { return s.X }

// аналог RecvSpec, но значения принимаются балком
type SpawnSpec2 struct {
	X     sym.ADT
	SigPH sym.ADT // export
	Cont  Term
}

func (s SpawnSpec2) Via() sym.ADT { return s.SigPH }

type CloseImpl struct {
	X sym.ADT
}

func (i CloseImpl) Via() sym.ADT { return i.X }

func (CloseImpl) impl() {}

func (CloseImpl) val2() {}

type WaitImpl struct {
	X    sym.ADT
	Cont Term
}

func (i WaitImpl) Via() sym.ADT { return i.X }

func (WaitImpl) impl() {}

func (WaitImpl) cont2() {}

type SendImpl struct {
	X sym.ADT
	A id.ADT
	B id.ADT
	S id.ADT
}

func (i SendImpl) Via() sym.ADT { return i.X }

func (SendImpl) impl() {}

func (SendImpl) val2() {}

type RecvImpl struct {
	X    sym.ADT
	A    id.ADT
	Y    sym.ADT
	Cont Term
}

func (i RecvImpl) Via() sym.ADT { return i.X }

func (RecvImpl) impl() {}

func (RecvImpl) cont2() {}

type LabImpl struct {
	X sym.ADT
	A id.ADT
	L sym.ADT
}

func (i LabImpl) Via() sym.ADT { return i.X }

func (LabImpl) impl() {}

func (LabImpl) val2() {}

type CaseImpl struct {
	X     sym.ADT
	A     id.ADT
	Conts map[sym.ADT]Term
	// States map[sym.ADT]state.ID
}

func (i CaseImpl) Via() sym.ADT { return i.X }

func (CaseImpl) impl() {}

func (CaseImpl) cont2() {}

type FwdImpl struct {
	X sym.ADT
	B id.ADT // to
}

func (i FwdImpl) Via() sym.ADT { return i.X }

func (FwdImpl) impl() {}

func (FwdImpl) val2() {}

func (FwdImpl) cont2() {}

type Repo interface {
	Insert(data.Source, ...Root) error
	SelectAll(data.Source) ([]Ref, error)
	SelectByID(data.Source, id.ADT) (Root, error)
	SelectByPID(data.Source, id.ADT) (Root, error)
	SelectByVID(data.Source, id.ADT) (Root, error)
}

func CollectEnv(t Term) []id.ADT {
	return collectEnvRec(t, []id.ADT{})
}

func collectEnvRec(t Term, env []id.ADT) []id.ADT {
	switch term := t.(type) {
	case RecvSpec:
		return collectEnvRec(term.Cont, env)
	case CaseSpec:
		for _, cont := range term.Conts {
			env = collectEnvRec(cont, env)
		}
		return env
	case SpawnSpec:
		return collectEnvRec(term.Cont, append(env, term.SigID))
	default:
		return env
	}
}

func ErrDoesNotExist(want ID) error {
	return fmt.Errorf("root doesn't exist: %v", want)
}

func ErrRootTypeUnexpected(got Root) error {
	return fmt.Errorf("root type unexpected: %T", got)
}

func ErrRootTypeMismatch(got, want Root) error {
	return fmt.Errorf("root type mismatch: want %T, got %T", want, got)
}

func ErrTermTypeUnexpected(got Term) error {
	return fmt.Errorf("term type unexpected: %T", got)
}

func ErrImplTypeUnexpected(got Impl) error {
	return fmt.Errorf("impl type unexpected: %T", got)
}

func ErrTermTypeMismatch(got, want Term) error {
	return fmt.Errorf("term type mismatch: want %T, got %T", want, got)
}

func ErrTermValueNil(pid id.ADT) error {
	return fmt.Errorf("proc %q term is nil", pid)
}

func ErrValTypeUnexpected(got Value) error {
	return fmt.Errorf("value type unexpected: %T", got)
}

func ErrValTypeUnexpected2(got Val) error {
	return fmt.Errorf("value type unexpected: %T", got)
}

func ErrContTypeUnexpected(got Continuation) error {
	return fmt.Errorf("continuation type unexpected: %T", got)
}

func ErrContTypeUnexpected2(got Cont) error {
	return fmt.Errorf("continuation type unexpected: %T", got)
}

func ErrMissingInCfg(want sym.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCfg2(want id.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}
