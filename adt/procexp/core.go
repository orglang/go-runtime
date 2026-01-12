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
	CommChnlPH qualsym.ADT
}

func (s CloseSpec) Via() qualsym.ADT { return s.CommChnlPH }

type WaitSpec struct {
	CommChnlPH qualsym.ADT
	ContES     ExpSpec
}

func (s WaitSpec) Via() qualsym.ADT { return s.CommChnlPH }

type SendSpec struct {
	CommChnlPH qualsym.ADT
	ValChnlPH  qualsym.ADT
}

func (s SendSpec) Via() qualsym.ADT { return s.CommChnlPH }

type RecvSpec struct {
	CommChnlPH qualsym.ADT
	BindChnlPH qualsym.ADT
	ContES     ExpSpec
}

func (s RecvSpec) Via() qualsym.ADT { return s.CommChnlPH }

type LabSpec struct {
	CommChnlPH qualsym.ADT
	LabelQN    qualsym.ADT
	ContES     ExpSpec
}

func (s LabSpec) Via() qualsym.ADT { return s.CommChnlPH }

type CaseSpec struct {
	CommChnlPH qualsym.ADT
	ContESs    map[qualsym.ADT]ExpSpec
}

func (s CaseSpec) Via() qualsym.ADT { return s.CommChnlPH }

// aka ExpName
type LinkSpec struct {
	ProcQN qualsym.ADT
	X      identity.ADT
	Ys     []identity.ADT
}

func (s LinkSpec) Via() qualsym.ADT { return "" }

type FwdSpec struct {
	CommChnlPH qualsym.ADT // old via (from)
	ContChnlPH qualsym.ADT // new via (to)
}

func (s FwdSpec) Via() qualsym.ADT { return s.CommChnlPH }

type CallSpec struct {
	CommChnlPH qualsym.ADT
	BindChnlPH qualsym.ADT
	ProcQN     qualsym.ADT
	ValChnlPHs []qualsym.ADT // channel bulk
	ContES     ExpSpec
}

func (s CallSpec) Via() qualsym.ADT { return s.CommChnlPH }

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
	CommChnlPH  qualsym.ADT
	ProcQN      qualsym.ADT
	BindChnlPHs []qualsym.ADT
	ContES      ExpSpec
}

func (s SpawnSpec) Via() qualsym.ADT { return s.CommChnlPH }

type AcqureSpec struct {
	CommChnlPH qualsym.ADT
	ContES     ExpSpec
}

func (s AcqureSpec) Via() qualsym.ADT { return s.CommChnlPH }

type AcceptSpec struct {
	CommChnlPH qualsym.ADT
	ContES     ExpSpec
}

func (s AcceptSpec) Via() qualsym.ADT { return s.CommChnlPH }

type DetachSpec struct {
	CommChnlPH qualsym.ADT
}

func (s DetachSpec) Via() qualsym.ADT { return s.CommChnlPH }

type ReleaseSpec struct {
	CommChnlPH qualsym.ADT
}

func (s ReleaseSpec) Via() qualsym.ADT { return s.CommChnlPH }

type ExpRec interface {
	ExpSpec
	impl()
}

type CloseRec struct {
	CommChnlPH qualsym.ADT
}

func (r CloseRec) Via() qualsym.ADT { return r.CommChnlPH }

func (CloseRec) impl() {}

type WaitRec struct {
	CommChnlPH qualsym.ADT
	ContES     ExpSpec
}

func (r WaitRec) Via() qualsym.ADT { return r.CommChnlPH }

func (WaitRec) impl() {}

type SendRec struct {
	CommChnlPH qualsym.ADT
	ContChnlID identity.ADT
	ValChnlID  identity.ADT
	ValExpID   identity.ADT
}

func (r SendRec) Via() qualsym.ADT { return r.CommChnlPH }

func (SendRec) impl() {}

type RecvRec struct {
	CommChnlPH qualsym.ADT
	ContChnlID identity.ADT
	ValChnlPH  qualsym.ADT
	ContES     ExpSpec
}

func (r RecvRec) Via() qualsym.ADT { return r.CommChnlPH }

func (RecvRec) impl() {}

type LabRec struct {
	CommChnlPH qualsym.ADT
	ContChnlID identity.ADT
	LabelQN    qualsym.ADT
}

func (r LabRec) Via() qualsym.ADT { return r.CommChnlPH }

func (LabRec) impl() {}

type CaseRec struct {
	CommChnlPH qualsym.ADT
	ContChnlID identity.ADT
	ContESs    map[qualsym.ADT]ExpSpec
}

func (r CaseRec) Via() qualsym.ADT { return r.CommChnlPH }

func (CaseRec) impl() {}

type FwdRec struct {
	CommChnlPH qualsym.ADT
	ContChnlID identity.ADT
}

func (r FwdRec) Via() qualsym.ADT { return r.CommChnlPH }

func (FwdRec) impl() {}

func CollectEnv(spec ExpSpec) []identity.ADT {
	return collectEnvRec(spec, []identity.ADT{})
}

func collectEnvRec(es ExpSpec, env []identity.ADT) []identity.ADT {
	switch spec := es.(type) {
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
