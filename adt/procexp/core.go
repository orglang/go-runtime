package procexp

import (
	"fmt"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/uniqsym"
)

type ExpSpec interface {
	Via() symbol.ADT
}

type CloseSpec struct {
	CommChnlPH symbol.ADT
}

func (s CloseSpec) Via() symbol.ADT { return s.CommChnlPH }

type WaitSpec struct {
	CommChnlPH symbol.ADT
	ContES     ExpSpec
}

func (s WaitSpec) Via() symbol.ADT { return s.CommChnlPH }

type SendSpec struct {
	CommChnlPH symbol.ADT
	ValChnlPH  symbol.ADT
}

func (s SendSpec) Via() symbol.ADT { return s.CommChnlPH }

type RecvSpec struct {
	CommChnlPH symbol.ADT
	BindChnlPH symbol.ADT
	ContES     ExpSpec
}

func (s RecvSpec) Via() symbol.ADT { return s.CommChnlPH }

type LabSpec struct {
	CommChnlPH symbol.ADT
	LabelQN    uniqsym.ADT
	ContES     ExpSpec
}

func (s LabSpec) Via() symbol.ADT { return s.CommChnlPH }

type CaseSpec struct {
	CommChnlPH symbol.ADT
	ContESs    map[uniqsym.ADT]ExpSpec
}

func (s CaseSpec) Via() symbol.ADT { return s.CommChnlPH }

// aka ExpName
type LinkSpec struct {
	ProcQN uniqsym.ADT
	X      identity.ADT
	Ys     []identity.ADT
}

func (s LinkSpec) Via() symbol.ADT { return "" }

type FwdSpec struct {
	CommChnlPH symbol.ADT // old via / from / x
	ContChnlPH symbol.ADT // new via / to / y
}

func (s FwdSpec) Via() symbol.ADT { return s.CommChnlPH }

type CallSpec struct {
	CommChnlPH symbol.ADT
	BindChnlPH symbol.ADT
	ProcQN     uniqsym.ADT
	ValChnlPHs []symbol.ADT // channel bulk
	ContES     ExpSpec
}

func (s CallSpec) Via() symbol.ADT { return s.CommChnlPH }

// аналог RecvSpec, но значения принимаются балком
type SpawnSpecOld struct {
	X      symbol.ADT
	SigID  identity.ADT
	Ys     []symbol.ADT
	PoolQN uniqsym.ADT
	ContES ExpSpec
}

func (s SpawnSpecOld) Via() symbol.ADT { return s.X }

type SpawnSpec struct {
	CommChnlPH  symbol.ADT
	ProcQN      uniqsym.ADT
	BindChnlPHs []symbol.ADT
	ContES      ExpSpec
}

func (s SpawnSpec) Via() symbol.ADT { return s.CommChnlPH }

type AcqureSpec struct {
	CommChnlPH symbol.ADT
	ContES     ExpSpec
}

func (s AcqureSpec) Via() symbol.ADT { return s.CommChnlPH }

type AcceptSpec struct {
	CommChnlPH symbol.ADT
	ContES     ExpSpec
}

func (s AcceptSpec) Via() symbol.ADT { return s.CommChnlPH }

type DetachSpec struct {
	CommChnlPH symbol.ADT
}

func (s DetachSpec) Via() symbol.ADT { return s.CommChnlPH }

type ReleaseSpec struct {
	CommChnlPH symbol.ADT
}

func (s ReleaseSpec) Via() symbol.ADT { return s.CommChnlPH }

type ExpRec interface {
	ExpSpec
	impl()
}

type CloseRec struct {
	CommChnlPH symbol.ADT
}

func (r CloseRec) Via() symbol.ADT { return r.CommChnlPH }

func (CloseRec) impl() {}

type WaitRec struct {
	CommChnlPH symbol.ADT
	ContES     ExpSpec
}

func (r WaitRec) Via() symbol.ADT { return r.CommChnlPH }

func (WaitRec) impl() {}

type SendRec struct {
	CommChnlPH symbol.ADT
	ContChnlID identity.ADT
	ValChnlID  identity.ADT
	ValExpID   identity.ADT
}

func (r SendRec) Via() symbol.ADT { return r.CommChnlPH }

func (SendRec) impl() {}

type RecvRec struct {
	CommChnlPH symbol.ADT
	ContChnlID identity.ADT
	ValChnlPH  symbol.ADT
	ContES     ExpSpec
}

func (r RecvRec) Via() symbol.ADT { return r.CommChnlPH }

func (RecvRec) impl() {}

type LabRec struct {
	CommChnlPH symbol.ADT
	ContChnlID identity.ADT
	LabelQN    uniqsym.ADT
}

func (r LabRec) Via() symbol.ADT { return r.CommChnlPH }

func (LabRec) impl() {}

type CaseRec struct {
	CommChnlPH symbol.ADT
	ContChnlID identity.ADT
	ContESs    map[uniqsym.ADT]ExpSpec
}

func (r CaseRec) Via() symbol.ADT { return r.CommChnlPH }

func (CaseRec) impl() {}

type FwdRec struct {
	CommChnlPH symbol.ADT
	ContChnlID identity.ADT
}

func (r FwdRec) Via() symbol.ADT { return r.CommChnlPH }

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
