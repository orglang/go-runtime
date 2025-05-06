package step

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type RefMsg struct {
	ID string `json:"id" param:"id"`
}

type StepKind string

const (
	Proc = StepKind("proc")
	Msg  = StepKind("msg")
	Srv  = StepKind("srv")
)

var stepKindRequired = []validation.Rule{
	validation.Required,
	validation.In(Proc, Msg, Srv),
}

type RootMsg struct {
	ID string   `json:"id"`
	K  StepKind `json:"kind"`
}

type ProcRootMsg struct {
	ID   string   `json:"id"`
	PID  string   `json:"pid"`
	Term *TermMsg `json:"term"`
}

type TermKind string

const (
	Close = TermKind("close")
	Wait  = TermKind("wait")
	Send  = TermKind("send")
	Recv  = TermKind("recv")
	Lab   = TermKind("lab")
	Case  = TermKind("case")
	Call  = TermKind("call")
	Link  = TermKind("link")
	Spawn = TermKind("spawn")
	Fwd   = TermKind("fwd")
)

var termKindRequired = []validation.Rule{
	validation.Required,
	validation.In(Close, Wait, Send, Recv, Lab, Case, Spawn, Fwd, Call),
}

type TermMsg struct {
	K     TermKind  `json:"kind"`
	Close *CloseMsg `json:"close,omitempty"`
	Wait  *WaitMsg  `json:"wait,omitempty"`
	Send  *SendMsg  `json:"send,omitempty"`
	Recv  *RecvMsg  `json:"recv,omitempty"`
	Lab   *LabMsg   `json:"lab,omitempty"`
	Case  *CaseMsg  `json:"case,omitempty"`
	Spawn *SpawnMsg `json:"spawn,omitempty"`
	Fwd   *FwdMsg   `json:"fwd,omitempty"`
	Call  *CallMsg  `json:"call,omitempty"`
}

func (dto TermMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.K, termKindRequired...),
		validation.Field(&dto.Close, validation.Required.When(dto.K == Close)),
		validation.Field(&dto.Wait, validation.Required.When(dto.K == Wait)),
		validation.Field(&dto.Send, validation.Required.When(dto.K == Send)),
		validation.Field(&dto.Recv, validation.Required.When(dto.K == Recv)),
		validation.Field(&dto.Lab, validation.Required.When(dto.K == Lab)),
		validation.Field(&dto.Case, validation.Required.When(dto.K == Case)),
		validation.Field(&dto.Spawn, validation.Required.When(dto.K == Spawn)),
		validation.Field(&dto.Fwd, validation.Required.When(dto.K == Fwd)),
		validation.Field(&dto.Call, validation.Required.When(dto.K == Call)),
	)
}

type CloseMsg struct {
	X string `json:"a"`
}

func (dto CloseMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
	)
}

type WaitMsg struct {
	X    string  `json:"x"`
	Cont TermMsg `json:"cont"`
}

func (dto WaitMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type SendMsg struct {
	X string `json:"a"`
	Y string `json:"b"`
}

func (dto SendMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Y, validation.Required),
	)
}

type RecvMsg struct {
	X    string  `json:"x"`
	Y    string  `json:"y"`
	Cont TermMsg `json:"cont"`
}

func (dto RecvMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Y, validation.Required),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type LabMsg struct {
	X     string `json:"a"`
	Label string `json:"label"`
}

func (dto LabMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Label, core.NameRequired...),
	)
}

type CaseMsg struct {
	X   string      `json:"x"`
	Brs []BranchMsg `json:"branches"`
}

func (dto CaseMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Brs,
			validation.Required,
			validation.Length(1, 10),
			validation.Each(validation.Required),
		),
	)
}

type BranchMsg struct {
	Label string  `json:"label"`
	Cont  TermMsg `json:"cont"`
}

func (dto BranchMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Label, core.NameRequired...),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type CallMsg struct {
	X     string   `json:"x"`
	SigPH string   `json:"sig_ph"`
	Ys    []string `json:"ys"`
}

func (dto CallMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigPH, id.Required...),
		validation.Field(&dto.Ys, core.CtxOptional...),
	)
}

type SpawnMsg struct {
	X     string   `json:"x"`
	SigID string   `json:"sig_id"`
	Ys    []string `json:"ys"`
	Cont  *TermMsg `json:"cont"`
}

func (dto SpawnMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigID, id.Required...),
		validation.Field(&dto.Ys, core.CtxOptional...),
		// validation.Field(&dto.Cont, validation.Required),
	)
}

type FwdMsg struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func (dto FwdMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Y, validation.Required),
	)
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend smecalculus/rolevod/lib/id:Convert.*
// goverter:extend MsgFromTerm
// goverter:extend MsgToTerm
// goverter:extend MsgFromTermNilable
// goverter:extend MsgToTermNilable
var (
	MsgFromProcRoot func(ProcRoot) ProcRootMsg
	MsgToProcRoot   func(ProcRootMsg) (ProcRoot, error)
)

func MsgFromTermNilable(t Term) *TermMsg {
	if t == nil {
		return nil
	}
	term := MsgFromTerm(t)
	return &term
}

func MsgFromTerm(t Term) TermMsg {
	switch term := t.(type) {
	case CloseSpec:
		return TermMsg{
			K: Close,
			Close: &CloseMsg{
				X: sym.ConvertToString(term.X),
			},
		}
	case WaitSpec:
		return TermMsg{
			K: Wait,
			Wait: &WaitMsg{
				X:    sym.ConvertToString(term.X),
				Cont: MsgFromTerm(term.Cont),
			},
		}
	case SendSpec:
		return TermMsg{
			K: Send,
			Send: &SendMsg{
				X: sym.ConvertToString(term.X),
				Y: sym.ConvertToString(term.Y),
			},
		}
	case RecvSpec:
		return TermMsg{
			K: Recv,
			Recv: &RecvMsg{
				X:    sym.ConvertToString(term.X),
				Y:    sym.ConvertToString(term.Y),
				Cont: MsgFromTerm(term.Cont),
			},
		}
	case LabSpec:
		return TermMsg{
			K: Lab,
			Lab: &LabMsg{
				X:     sym.ConvertToString(term.X),
				Label: string(term.Label),
			},
		}
	case CaseSpec:
		brs := []BranchMsg{}
		for l, t := range term.Conts {
			brs = append(brs, BranchMsg{Label: string(l), Cont: MsgFromTerm(t)})
		}
		return TermMsg{
			K: Case,
			Case: &CaseMsg{
				X:   sym.ConvertToString(term.X),
				Brs: brs,
			},
		}
	case SpawnSpec:
		return TermMsg{
			K: Spawn,
			Spawn: &SpawnMsg{
				X:     sym.ConvertToString(term.X),
				SigID: id.ConvertToString(term.SigID),
				Ys:    sym.ConvertToStrings(term.Ys),
				Cont:  MsgFromTermNilable(term.Cont),
			},
		}
	case CallSpec:
		return TermMsg{
			K: Call,
			Call: &CallMsg{
				X:     sym.ConvertToString(term.X),
				SigPH: sym.ConvertToString(term.SigPH),
				Ys:    sym.ConvertToStrings(term.Ys),
			},
		}
	case FwdSpec:
		return TermMsg{
			K: Fwd,
			Fwd: &FwdMsg{
				X: sym.ConvertToString(term.X),
				Y: sym.ConvertToString(term.Y),
			},
		}
	default:
		panic(ErrTermTypeUnexpected(t))
	}
}

func MsgToTermNilable(dto *TermMsg) (Term, error) {
	if dto == nil {
		return nil, nil
	}
	return MsgToTerm(*dto)
}

func MsgToTerm(dto TermMsg) (Term, error) {
	switch dto.K {
	case Close:
		x, err := sym.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseSpec{X: x}, nil
	case Wait:
		x, err := sym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToTerm(dto.Wait.Cont)
		if err != nil {
			return nil, err
		}
		return WaitSpec{X: x, Cont: cont}, nil
	case Send:
		x, err := sym.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Send.Y)
		if err != nil {
			return nil, err
		}
		return SendSpec{X: x, Y: y}, nil
	case Recv:
		x, err := sym.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToTerm(dto.Recv.Cont)
		if err != nil {
			return nil, err
		}
		return RecvSpec{X: x, Y: y, Cont: cont}, nil
	case Lab:
		x, err := sym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabSpec{X: x, Label: sym.ADT(dto.Lab.Label)}, nil
	case Case:
		x, err := sym.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[sym.ADT]Term, len(dto.Case.Brs))
		for _, b := range dto.Case.Brs {
			cont, err := MsgToTerm(b.Cont)
			if err != nil {
				return nil, err
			}
			conts[sym.ADT(b.Label)] = cont
		}
		return CaseSpec{X: x, Conts: conts}, nil
	case Spawn:
		x, err := sym.ConvertFromString(dto.Spawn.X)
		if err != nil {
			return nil, err
		}
		sigID, err := id.ConvertFromString(dto.Spawn.SigID)
		if err != nil {
			return nil, err
		}
		ys, err := sym.ConvertFromStrings(dto.Spawn.Ys)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToTermNilable(dto.Spawn.Cont)
		if err != nil {
			return nil, err
		}
		return SpawnSpec{X: x, SigID: sigID, Ys: ys, Cont: cont}, nil
	case Call:
		x, err := sym.ConvertFromString(dto.Call.X)
		if err != nil {
			return nil, err
		}
		sigPH, err := sym.ConvertFromString(dto.Call.SigPH)
		if err != nil {
			return nil, err
		}
		ys, err := sym.ConvertFromStrings(dto.Call.Ys)
		if err != nil {
			return nil, err
		}
		return CallSpec{X: x, SigPH: sigPH, Ys: ys}, nil
	case Fwd:
		x, err := sym.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Fwd.Y)
		if err != nil {
			return nil, err
		}
		return FwdSpec{X: x, Y: y}, nil
	default:
		panic(ErrUnexpectedTermKind(dto.K))
	}
}

func ErrUnexpectedTermKind(k TermKind) error {
	return fmt.Errorf("unexpected term kind: %v", k)
}

func ErrUnexpectedStepKind(k StepKind) error {
	return fmt.Errorf("unexpected step kind: %v", k)
}
