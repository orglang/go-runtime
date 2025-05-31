package def

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type SemKind string

const (
	Msg = SemKind("msg")
	Svc = SemKind("svc")
)

var semKindRequired = []validation.Rule{
	validation.Required,
	validation.In(Msg, Svc),
}

type RootMsg struct {
	ID string  `json:"id"`
	K  SemKind `json:"kind"`
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

type TermSpecMsg struct {
	K     TermKind      `json:"kind"`
	Close *CloseSpecMsg `json:"close,omitempty"`
	Wait  *WaitSpecMsg  `json:"wait,omitempty"`
	Send  *SendSpecMsg  `json:"send,omitempty"`
	Recv  *RecvSpecMsg  `json:"recv,omitempty"`
	Lab   *LabSpecMsg   `json:"lab,omitempty"`
	Case  *CaseSpecMsg  `json:"case,omitempty"`
	Spawn *SpawnSpecMsg `json:"spawn,omitempty"`
	Fwd   *FwdSpecMsg   `json:"fwd,omitempty"`
	Call  *CallSpecMsg  `json:"call,omitempty"`
}

func (dto TermSpecMsg) Validate() error {
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

type CloseSpecMsg struct {
	X string `json:"x"`
}

func (dto CloseSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
	)
}

type WaitSpecMsg struct {
	X    string      `json:"x"`
	Cont TermSpecMsg `json:"cont"`
}

func (dto WaitSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type SendSpecMsg struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func (dto SendSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Y, validation.Required),
	)
}

type RecvSpecMsg struct {
	X    string      `json:"x"`
	Y    string      `json:"y"`
	Cont TermSpecMsg `json:"cont"`
}

func (dto RecvSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Y, validation.Required),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type LabSpecMsg struct {
	X     string `json:"x"`
	Label string `json:"label"`
}

func (dto LabSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Label, core.NameRequired...),
	)
}

type CaseSpecMsg struct {
	X   string          `json:"x"`
	Brs []BranchSpecMsg `json:"branches"`
}

func (dto CaseSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Brs,
			validation.Required,
			validation.Length(1, 10),
			validation.Each(validation.Required),
		),
	)
}

type BranchSpecMsg struct {
	Label string      `json:"label"`
	Cont  TermSpecMsg `json:"cont"`
}

func (dto BranchSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.Label, core.NameRequired...),
		validation.Field(&dto.Cont, validation.Required),
	)
}

type CallSpecMsg struct {
	X     string   `json:"x"`
	SigPH string   `json:"sig_ph"`
	Ys    []string `json:"ys"`
}

func (dto CallSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigPH, id.Required...),
		validation.Field(&dto.Ys, core.CtxOptional...),
	)
}

type SpawnSpecMsg struct {
	X     string       `json:"x"`
	SigID string       `json:"sig_id"`
	Ys    []string     `json:"ys"`
	Cont  *TermSpecMsg `json:"cont"`
}

func (dto SpawnSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.SigID, id.Required...),
		validation.Field(&dto.Ys, core.CtxOptional...),
		// validation.Field(&dto.Cont, validation.Required),
	)
}

type FwdSpecMsg struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func (dto FwdSpecMsg) Validate() error {
	return validation.ValidateStruct(&dto,
		validation.Field(&dto.X, validation.Required),
		validation.Field(&dto.Y, validation.Required),
	)
}

func MsgFromTermSpecNilable(t TermSpec) *TermSpecMsg {
	if t == nil {
		return nil
	}
	term := MsgFromTermSpec(t)
	return &term
}

func MsgFromTermSpec(t TermSpec) TermSpecMsg {
	switch term := t.(type) {
	case CloseSpec:
		return TermSpecMsg{
			K: Close,
			Close: &CloseSpecMsg{
				X: sym.ConvertToString(term.CommPH),
			},
		}
	case WaitSpec:
		return TermSpecMsg{
			K: Wait,
			Wait: &WaitSpecMsg{
				X:    sym.ConvertToString(term.CommPH),
				Cont: MsgFromTermSpec(term.ContTS),
			},
		}
	case SendSpec:
		return TermSpecMsg{
			K: Send,
			Send: &SendSpecMsg{
				X: sym.ConvertToString(term.CommPH),
				Y: sym.ConvertToString(term.ValPH),
			},
		}
	case RecvSpec:
		return TermSpecMsg{
			K: Recv,
			Recv: &RecvSpecMsg{
				X:    sym.ConvertToString(term.CommPH),
				Y:    sym.ConvertToString(term.BindPH),
				Cont: MsgFromTermSpec(term.ContTS),
			},
		}
	case LabSpec:
		return TermSpecMsg{
			K: Lab,
			Lab: &LabSpecMsg{
				X:     sym.ConvertToString(term.CommPH),
				Label: string(term.Label),
			},
		}
	case CaseSpec:
		brs := []BranchSpecMsg{}
		for l, t := range term.Conts {
			brs = append(brs, BranchSpecMsg{Label: string(l), Cont: MsgFromTermSpec(t)})
		}
		return TermSpecMsg{
			K: Case,
			Case: &CaseSpecMsg{
				X:   sym.ConvertToString(term.CommPH),
				Brs: brs,
			},
		}
	case SpawnSpecOld:
		return TermSpecMsg{
			K: Spawn,
			Spawn: &SpawnSpecMsg{
				X:     sym.ConvertToString(term.X),
				SigID: id.ConvertToString(term.SigID),
				Ys:    sym.ConvertToStrings(term.Ys),
				Cont:  MsgFromTermSpecNilable(term.Cont),
			},
		}
	case CallSpecOld:
		return TermSpecMsg{
			K: Call,
			Call: &CallSpecMsg{
				X:     sym.ConvertToString(term.X),
				SigPH: sym.ConvertToString(term.SigPH),
				Ys:    sym.ConvertToStrings(term.Ys),
			},
		}
	case FwdSpec:
		return TermSpecMsg{
			K: Fwd,
			Fwd: &FwdSpecMsg{
				X: sym.ConvertToString(term.X),
				Y: sym.ConvertToString(term.Y),
			},
		}
	default:
		panic(ErrTermTypeUnexpected(t))
	}
}

func MsgToTermSpecNilable(dto *TermSpecMsg) (TermSpec, error) {
	if dto == nil {
		return nil, nil
	}
	return MsgToTermSpec(*dto)
}

func MsgToTermSpec(dto TermSpecMsg) (TermSpec, error) {
	switch dto.K {
	case Close:
		x, err := sym.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseSpec{CommPH: x}, nil
	case Wait:
		x, err := sym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToTermSpec(dto.Wait.Cont)
		if err != nil {
			return nil, err
		}
		return WaitSpec{CommPH: x, ContTS: cont}, nil
	case Send:
		x, err := sym.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Send.Y)
		if err != nil {
			return nil, err
		}
		return SendSpec{CommPH: x, ValPH: y}, nil
	case Recv:
		x, err := sym.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToTermSpec(dto.Recv.Cont)
		if err != nil {
			return nil, err
		}
		return RecvSpec{CommPH: x, BindPH: y, ContTS: cont}, nil
	case Lab:
		x, err := sym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabSpec{CommPH: x, Label: sym.ADT(dto.Lab.Label)}, nil
	case Case:
		x, err := sym.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[sym.ADT]TermSpec, len(dto.Case.Brs))
		for _, b := range dto.Case.Brs {
			cont, err := MsgToTermSpec(b.Cont)
			if err != nil {
				return nil, err
			}
			conts[sym.ADT(b.Label)] = cont
		}
		return CaseSpec{CommPH: x, Conts: conts}, nil
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
		cont, err := MsgToTermSpecNilable(dto.Spawn.Cont)
		if err != nil {
			return nil, err
		}
		return SpawnSpecOld{X: x, SigID: sigID, Ys: ys, Cont: cont}, nil
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
		return CallSpecOld{X: x, SigPH: sigPH, Ys: ys}, nil
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

func ErrUnexpectedSemKind(k SemKind) error {
	return fmt.Errorf("unexpected sem kind: %v", k)
}
