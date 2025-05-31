package def

import (
	"fmt"
	"log/slog"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type ProcSpec struct {
	ProcQN sym.ADT // or dec.ProcID
	ProcTS TermSpec
}

type ProcRef struct {
	ProcID id.ADT
}

type ProcRec struct {
	ProcID id.ADT
}

type ProcSnap struct {
	ProcID id.ADT
}

type TermSpec interface {
	Via() sym.ADT
}

type CloseSpec struct {
	CommPH sym.ADT
}

func (s CloseSpec) Via() sym.ADT { return s.CommPH }

type WaitSpec struct {
	CommPH sym.ADT
	ContTS TermSpec
}

func (s WaitSpec) Via() sym.ADT { return s.CommPH }

type SendSpec struct {
	CommPH sym.ADT // via
	ValPH  sym.ADT // val
}

func (s SendSpec) Via() sym.ADT { return s.CommPH }

type RecvSpec struct {
	CommPH sym.ADT
	BindPH sym.ADT
	ContTS TermSpec
}

func (s RecvSpec) Via() sym.ADT { return s.CommPH }

type LabSpec struct {
	CommPH sym.ADT
	Label  sym.ADT
	ContTS TermSpec
}

func (s LabSpec) Via() sym.ADT { return s.CommPH }

type CaseSpec struct {
	CommPH sym.ADT
	Conts  map[sym.ADT]TermSpec
}

func (s CaseSpec) Via() sym.ADT { return s.CommPH }

// aka ExpName
type LinkSpec struct {
	ProcQN sym.ADT
	X      id.ADT
	Ys     []id.ADT
}

func (s LinkSpec) Via() sym.ADT { return "" }

type FwdSpec struct {
	X sym.ADT // old via (from)
	Y sym.ADT // new via (to)
}

func (s FwdSpec) Via() sym.ADT { return s.X }

// аналог SendSpec, но значения отправляются балком
type CallSpecOld struct {
	X     sym.ADT
	SigPH sym.ADT // import
	Ys    []sym.ADT
}

func (s CallSpecOld) Via() sym.ADT { return s.SigPH }

type CallSpec struct {
	CommPH sym.ADT
	BindPH sym.ADT
	ProcSN sym.ADT   // label
	ValPHs []sym.ADT // channel bulk
	ContTS TermSpec
}

func (s CallSpec) Via() sym.ADT { return s.CommPH }

// аналог RecvSpec, но значения принимаются балком
type SpawnSpecOld struct {
	X      sym.ADT
	SigID  id.ADT
	Ys     []sym.ADT
	PoolQN sym.ADT
	Cont   TermSpec
}

func (s SpawnSpecOld) Via() sym.ADT { return s.X }

type SpawnSpec struct {
	CommPH sym.ADT
	ProcSN sym.ADT
	ContTS TermSpec
}

func (s SpawnSpec) Via() sym.ADT { return s.CommPH }

type AcqureSpec struct {
	CommPH sym.ADT
	ContTS TermSpec
}

func (s AcqureSpec) Via() sym.ADT { return s.CommPH }

type AcceptSpec struct {
	CommPH sym.ADT
	ContTS TermSpec
}

func (s AcceptSpec) Via() sym.ADT { return s.CommPH }

type DetachSpec struct {
	CommPH sym.ADT
}

func (s DetachSpec) Via() sym.ADT { return s.CommPH }

type ReleaseSpec struct {
	CommPH sym.ADT
}

func (s ReleaseSpec) Via() sym.ADT { return s.CommPH }

type TermRec interface {
	TermSpec
	impl()
}

type CloseRec struct {
	X sym.ADT
}

func (r CloseRec) Via() sym.ADT { return r.X }

func (CloseRec) impl() {}

type WaitRec struct {
	X    sym.ADT
	Cont TermSpec
}

func (r WaitRec) Via() sym.ADT { return r.X }

func (WaitRec) impl() {}

type SendRec struct {
	X      sym.ADT
	A      id.ADT
	B      id.ADT
	TermID id.ADT
}

func (r SendRec) Via() sym.ADT { return r.X }

func (SendRec) impl() {}

type RecvRec struct {
	X    sym.ADT
	A    id.ADT
	Y    sym.ADT
	Cont TermSpec
}

func (r RecvRec) Via() sym.ADT { return r.X }

func (RecvRec) impl() {}

type LabRec struct {
	X     sym.ADT
	A     id.ADT
	Label sym.ADT
}

func (r LabRec) Via() sym.ADT { return r.X }

func (LabRec) impl() {}

type CaseRec struct {
	X     sym.ADT
	A     id.ADT
	Conts map[sym.ADT]TermSpec
}

func (r CaseRec) Via() sym.ADT { return r.X }

func (CaseRec) impl() {}

type FwdRec struct {
	X sym.ADT
	B id.ADT // to
}

func (r FwdRec) Via() sym.ADT { return r.X }

func (FwdRec) impl() {}

func CollectEnv(spec TermSpec) []id.ADT {
	return collectEnvRec(spec, []id.ADT{})
}

type API interface {
	Create(ProcSpec) (ProcRef, error)
	Retrieve(id.ADT) (ProcRec, error)
}

// for compilation purposes
func newAPI() API {
	return &service{}
}

type service struct {
	procs    Repo
	operator data.Operator
	log      *slog.Logger
}

func newService(
	procs Repo,
	operator data.Operator,
	l *slog.Logger,
) *service {
	return &service{procs, operator, l}
}

func (s *service) Create(spec ProcSpec) (ProcRef, error) {
	return ProcRef{}, nil
}

func (s *service) Retrieve(recID id.ADT) (ProcRec, error) {
	return ProcRec{}, nil
}

type Repo interface {
	InsertProc(data.Source, ProcRec) error
}

func collectEnvRec(s TermSpec, env []id.ADT) []id.ADT {
	switch spec := s.(type) {
	case RecvSpec:
		return collectEnvRec(spec.ContTS, env)
	case CaseSpec:
		for _, cont := range spec.Conts {
			env = collectEnvRec(cont, env)
		}
		return env
	case SpawnSpecOld:
		return collectEnvRec(spec.Cont, append(env, spec.SigID))
	default:
		return env
	}
}

func ErrDoesNotExist(want id.ADT) error {
	return fmt.Errorf("rec doesn't exist: %v", want)
}

func ErrTermTypeUnexpected(got TermSpec) error {
	return fmt.Errorf("term spec unexpected: %T", got)
}

func ErrRecTypeUnexpected(got TermRec) error {
	return fmt.Errorf("term rec unexpected: %T", got)
}

func ErrTermTypeMismatch(got, want TermSpec) error {
	return fmt.Errorf("term spec mismatch: want %T, got %T", want, got)
}

func ErrTermValueNil(pid id.ADT) error {
	return fmt.Errorf("proc %q term is nil", pid)
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
