package step

import (
	"fmt"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/rev"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/chnl"
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
	step() chnl.ID
}

func ChnlID(r Root) chnl.ID { return r.step() }

// aka exec.Proc
type ProcRoot struct {
	ID   id.ADT
	PID  chnl.ID
	Term Term
}

func (r ProcRoot) step() chnl.ID { return r.PID }

// aka exec.Msg
type MsgRoot struct {
	ID  id.ADT
	PID chnl.ID
	VID chnl.ID
	Val Value
}

func (r MsgRoot) step() chnl.ID { return r.VID }

type MsgRoot2 struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Val    Val
	Rev    rev.ADT
}

func (r MsgRoot2) step() chnl.ID { return r.ChnlID }

type SrvRoot struct {
	ID   id.ADT
	PID  chnl.ID
	VID  chnl.ID
	Cont Continuation
}

func (r SrvRoot) step() chnl.ID { return r.VID }

type SvcRoot2 struct {
	PoolID id.ADT
	ProcID id.ADT
	ChnlID id.ADT
	Cont   Cont
	Rev    rev.ADT
}

func (r SvcRoot2) step() chnl.ID { return r.ChnlID }

type TbdRoot struct {
	ID  id.ADT
	PID chnl.ID
	VID chnl.ID
	Act Action
}

func (TbdRoot) step() {}

// aka Expression
type Term interface {
	Via() ph.ADT
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
	X ph.ADT
}

func (s CloseSpec) Via() ph.ADT { return s.X }

func (CloseSpec) val() {}

type WaitSpec struct {
	X    ph.ADT
	Cont Term
}

func (s WaitSpec) Via() ph.ADT { return s.X }

func (WaitSpec) cont() {}

type SendSpec struct {
	X ph.ADT // via
	Y ph.ADT // val
	// Cont  Term
}

func (s SendSpec) Via() ph.ADT { return s.X }

func (SendSpec) val() {}

type RecvSpec struct {
	X    ph.ADT // via
	Y    ph.ADT // val
	Cont Term
}

func (s RecvSpec) Via() ph.ADT { return s.X }

func (RecvSpec) cont() {}

type LabSpec struct {
	X ph.ADT
	L core.Label
	// Cont Term
}

func (s LabSpec) Via() ph.ADT { return s.X }

func (LabSpec) val() {}

type CaseSpec struct {
	X     ph.ADT
	Conts map[core.Label]Term
}

func (s CaseSpec) Via() ph.ADT { return s.X }

func (CaseSpec) cont() {}

// aka ExpName
type LinkSpec struct {
	PE    chnl.ID
	CEs   []chnl.ID
	SigQN sym.ADT
}

func (s LinkSpec) Via() ph.ADT { return "" }

// аналог SendSpec, но без продолжения с новым via
type FwdSpec struct {
	X ph.ADT // from
	Y ph.ADT // to
}

func (s FwdSpec) Via() ph.ADT { return s.X }

func (FwdSpec) val() {}

func (FwdSpec) cont() {}

// аналог балкового SendSpec
type SpawnSpec struct {
	X      ph.ADT
	Ys     []ph.ADT
	Ys2    []chnl.ID
	SigID  id.ADT // TODO ссылаться по QN
	PoolQN sym.ADT
	Cont   Term
}

func (s SpawnSpec) Via() ph.ADT { return s.X }

type EP struct {
	ChnlPH  ph.ADT
	ChnlID  id.ADT
	StateID id.ADT
}

type CloseImpl struct {
	X ph.ADT
}

func (i CloseImpl) Via() ph.ADT { return i.X }

func (CloseImpl) impl() {}

func (CloseImpl) val2() {}

type WaitImpl struct {
	X    ph.ADT
	Cont Term
}

func (i WaitImpl) Via() ph.ADT { return i.X }

func (WaitImpl) impl() {}

func (WaitImpl) cont2() {}

type SendImpl struct {
	X ph.ADT
	A id.ADT
	B id.ADT
	S id.ADT
}

func (i SendImpl) Via() ph.ADT { return i.X }

func (SendImpl) impl() {}

func (SendImpl) val2() {}

type RecvImpl struct {
	X    ph.ADT
	A    id.ADT
	Y    ph.ADT
	Cont Term
}

func (i RecvImpl) Via() ph.ADT { return i.X }

func (RecvImpl) impl() {}

func (RecvImpl) cont2() {}

type LabImpl struct {
	X ph.ADT
	A id.ADT
	L core.Label
}

func (i LabImpl) Via() ph.ADT { return i.X }

func (LabImpl) impl() {}

func (LabImpl) val2() {}

type CaseImpl struct {
	X     ph.ADT
	A     id.ADT
	Conts map[core.Label]Term
	// States map[core.Label]state.ID
}

func (i CaseImpl) Via() ph.ADT { return i.X }

func (CaseImpl) impl() {}

func (CaseImpl) cont2() {}

type FwdImpl struct {
	X ph.ADT
	B id.ADT // to
}

func (i FwdImpl) Via() ph.ADT { return i.X }

func (FwdImpl) impl() {}

func (FwdImpl) val2() {}

func (FwdImpl) cont2() {}

type Repo interface {
	Insert(data.Source, ...Root) error
	SelectAll(data.Source) ([]Ref, error)
	SelectByID(data.Source, id.ADT) (Root, error)
	SelectByPID(data.Source, chnl.ID) (Root, error)
	SelectByVID(data.Source, chnl.ID) (Root, error)
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

func ErrTermValueNil(pid chnl.ID) error {
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

func ErrMissingInCfg(want ph.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCfg2(want id.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}
