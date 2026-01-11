package procexp

import (
	"fmt"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
)

type ExpSpec interface {
	Via() qualsym.ADT
}

type CloseSpec struct {
	CommPH qualsym.ADT
}

func (s CloseSpec) Via() qualsym.ADT { return s.CommPH }

type WaitSpec struct {
	CommPH qualsym.ADT
	ContES ExpSpec
}

func (s WaitSpec) Via() qualsym.ADT { return s.CommPH }

type SendSpec struct {
	CommPH qualsym.ADT // via
	ValPH  qualsym.ADT // val
}

func (s SendSpec) Via() qualsym.ADT { return s.CommPH }

type RecvSpec struct {
	CommPH qualsym.ADT
	BindPH qualsym.ADT
	ContES ExpSpec
}

func (s RecvSpec) Via() qualsym.ADT { return s.CommPH }

type LabSpec struct {
	CommPH qualsym.ADT
	Label  qualsym.ADT
	ContES ExpSpec
}

func (s LabSpec) Via() qualsym.ADT { return s.CommPH }

type CaseSpec struct {
	CommPH  qualsym.ADT
	ContESs map[qualsym.ADT]ExpSpec
}

func (s CaseSpec) Via() qualsym.ADT { return s.CommPH }

// aka ExpName
type LinkSpec struct {
	ProcQN qualsym.ADT
	X      identity.ADT
	Ys     []identity.ADT
}

func (s LinkSpec) Via() qualsym.ADT { return "" }

type FwdSpec struct {
	X qualsym.ADT // old via (from)
	Y qualsym.ADT // new via (to)
}

func (s FwdSpec) Via() qualsym.ADT { return s.X }

type CallSpec struct {
	CommPH qualsym.ADT
	BindPH qualsym.ADT
	ProcQN qualsym.ADT
	ValPHs []qualsym.ADT // channel bulk
	ContES ExpSpec
}

func (s CallSpec) Via() qualsym.ADT { return s.CommPH }

// аналог RecvSpec, но значения принимаются балком
type SpawnSpecOld struct {
	X      qualsym.ADT
	SigID  identity.ADT
	Ys     []qualsym.ADT
	PoolQN qualsym.ADT
	ContES ExpSpec
}

func (s SpawnSpecOld) Via() qualsym.ADT { return s.X }

type SpawnSpec struct {
	CommPH  qualsym.ADT
	ProcQN  qualsym.ADT
	BindPHs []qualsym.ADT
	ContES  ExpSpec
}

func (s SpawnSpec) Via() qualsym.ADT { return s.CommPH }

type AcqureSpec struct {
	CommPH qualsym.ADT
	ContES ExpSpec
}

func (s AcqureSpec) Via() qualsym.ADT { return s.CommPH }

type AcceptSpec struct {
	CommPH qualsym.ADT
	ContES ExpSpec
}

func (s AcceptSpec) Via() qualsym.ADT { return s.CommPH }

type DetachSpec struct {
	CommPH qualsym.ADT
}

func (s DetachSpec) Via() qualsym.ADT { return s.CommPH }

type ReleaseSpec struct {
	CommPH qualsym.ADT
}

func (s ReleaseSpec) Via() qualsym.ADT { return s.CommPH }

type ExpRec interface {
	ExpSpec
	impl()
}

type CloseRec struct {
	X qualsym.ADT
}

func (r CloseRec) Via() qualsym.ADT { return r.X }

func (CloseRec) impl() {}

type WaitRec struct {
	X      qualsym.ADT
	ContES ExpSpec
}

func (r WaitRec) Via() qualsym.ADT { return r.X }

func (WaitRec) impl() {}

type SendRec struct {
	X     qualsym.ADT
	A     identity.ADT
	B     identity.ADT
	ExpID identity.ADT
}

func (r SendRec) Via() qualsym.ADT { return r.X }

func (SendRec) impl() {}

type RecvRec struct {
	X      qualsym.ADT
	A      identity.ADT
	Y      qualsym.ADT
	ContES ExpSpec
}

func (r RecvRec) Via() qualsym.ADT { return r.X }

func (RecvRec) impl() {}

type LabRec struct {
	X     qualsym.ADT
	A     identity.ADT
	Label qualsym.ADT
}

func (r LabRec) Via() qualsym.ADT { return r.X }

func (LabRec) impl() {}

type CaseRec struct {
	X       qualsym.ADT
	A       identity.ADT
	ContESs map[qualsym.ADT]ExpSpec
}

func (r CaseRec) Via() qualsym.ADT { return r.X }

func (CaseRec) impl() {}

type FwdRec struct {
	X qualsym.ADT
	B identity.ADT // to
}

func (r FwdRec) Via() qualsym.ADT { return r.X }

func (FwdRec) impl() {}

func CollectEnv(spec ExpSpec) []identity.ADT {
	return collectEnvRec(spec, []identity.ADT{})
}

func collectEnvRec(s ExpSpec, env []identity.ADT) []identity.ADT {
	switch spec := s.(type) {
	case RecvSpec:
		return collectEnvRec(spec.ContES, env)
	case CaseSpec:
		for _, cont := range spec.ContESs {
			env = collectEnvRec(cont, env)
		}
		return env
	case SpawnSpec:
		return collectEnvRec(spec.ContES, env)
	default:
		return env
	}
}

func ErrExpTypeUnexpected(got ExpSpec) error {
	return fmt.Errorf("term spec unexpected: %T", got)
}

func ErrRecTypeUnexpected(got ExpRec) error {
	return fmt.Errorf("term rec unexpected: %T", got)
}

func ErrExpTypeMismatch(got, want ExpSpec) error {
	return fmt.Errorf("term spec mismatch: want %T, got %T", want, got)
}

func ErrExpValueNil(pid identity.ADT) error {
	return fmt.Errorf("proc %q term is nil", pid)
}
