package def

import (
	"database/sql"
	"fmt"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type SemRecData struct {
	ID    string         `db:"id"`
	K     semKind        `db:"kind"`
	PID   sql.NullString `db:"pid"`
	VID   sql.NullString `db:"vid"`
	SemTR TermRecData    `db:"spec"`
}

type semKind int

const (
	nonsem = semKind(iota)
	msgKind
	svcKind
)

type TermRecData struct {
	K     termKind      `json:"k"`
	Close *closeRecData `json:"close,omitempty"`
	Wait  *waitRecData  `json:"wait,omitempty"`
	Send  *sendRecData  `json:"send,omitempty"`
	Recv  *recvRecData  `json:"recv,omitempty"`
	Lab   *labRecData   `json:"lab,omitempty"`
	Case  *caseRecData  `json:"case,omitempty"`
	Fwd   *fwdRecData   `json:"fwd,omitempty"`
}

type TermSpecData struct {
	K     termKind       `json:"k"`
	Close *closeSpecData `json:"close,omitempty"`
	Wait  *waitSpecData  `json:"wait,omitempty"`
	Send  *sendSpecData  `json:"send,omitempty"`
	Recv  *recvSpecData  `json:"recv,omitempty"`
	Lab   *labSpecData   `json:"lab,omitempty"`
	Case  *caseSpecData  `json:"case,omitempty"`
	Fwd   *fwdSpecData   `json:"fwd,omitempty"`
}

type termKind int

const (
	nonterm = termKind(iota)
	closeKind
	waitKind
	sendKind
	recvKind
	labKind
	caseKind
	linkKind
	spawnKind
	fwdKind
)

type closeSpecData struct {
	X string `json:"x"`
}

type closeRecData struct {
	X string `json:"x"`
}

type waitSpecData struct {
	X    string       `json:"x"`
	Cont TermSpecData `json:"cont"`
}

type waitRecData struct {
	X    string       `json:"x"`
	Cont TermSpecData `json:"cont"`
}

type sendSpecData struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type sendRecData struct {
	X string `json:"x"`
	A string `json:"a"`
	B string `json:"b"`
}

type recvSpecData struct {
	X    string       `json:"x"`
	Y    string       `json:"y"`
	Cont TermSpecData `json:"cont"`
}

type recvRecData struct {
	X    string       `json:"x"`
	A    string       `json:"a"`
	Y    string       `json:"y"`
	Cont TermSpecData `json:"cont"`
}

type labSpecData struct {
	X     string `json:"x"`
	Label string `json:"lab"`
}

type labRecData struct {
	X     string `json:"x"`
	Label string `json:"lab"`
}

type caseSpecData struct {
	X        string           `json:"x"`
	Branches []branchSpecData `json:"brs"`
}

type caseRecData struct {
	X        string          `json:"x"`
	Branches []branchRecData `json:"brs"`
}

type branchSpecData struct {
	Label string       `json:"lab"`
	Cont  TermSpecData `json:"cont"`
}

type branchRecData struct {
	Label string       `json:"lab"`
	Cont  TermSpecData `json:"cont"`
}

type fwdSpecData struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type fwdRecData struct {
	X string `json:"x"`
	B string `json:"b"`
}

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend data.*
var (
	DataToSemRecs     func([]SemRecData) ([]SemRec, error)
	DataFromSemRecs   func([]SemRec) ([]SemRecData, error)
	DataToTermSpecs   func([]TermRecData) ([]TermSpec, error)
	DataFromTermSpecs func([]TermSpec) ([]TermRecData, error)
	DataToTermRecs    func([]TermRecData) ([]TermRec, error)
	DataFromTermRecs  func([]TermRec) ([]TermRecData, error)
	DataToTermVals    func([]TermRecData) ([]Value, error)
	DataFromTermVals  func([]Value) []TermRecData
	DataToTermConts   func([]TermRecData) ([]Continuation, error)
	DataFromTermConts func([]Continuation) ([]TermRecData, error)
)

func dataFromSemRec(r SemRec) (SemRecData, error) {
	if r == nil {
		return SemRecData{}, nil
	}
	switch rec := r.(type) {
	case MsgRec:
		return SemRecData{
			K:     msgKind,
			SemTR: dataFromTermVal(rec.Val),
		}, nil
	case SvcRec:
		cont, err := dataFromTermCont(rec.Cont)
		if err != nil {
			return SemRecData{}, err
		}
		return SemRecData{
			K:     svcKind,
			SemTR: cont,
		}, nil
	default:
		panic(ErrRootTypeUnexpected(rec))
	}
}

func dataToSemRec(dto SemRecData) (SemRec, error) {
	var nilData SemRecData
	if dto == nilData {
		return nil, nil
	}
	switch dto.K {
	case msgKind:
		val, err := dataToTermVal(dto.SemTR)
		if err != nil {
			return nil, err
		}
		return MsgRec{Val: val}, nil
	case svcKind:
		cont, err := dataToTermCont(dto.SemTR)
		if err != nil {
			return nil, err
		}
		return SvcRec{Cont: cont}, nil
	default:
		panic(errUnexpectedStepKind(dto.K))
	}
}

func dataFromTermRec(r TermRec) (TermRecData, error) {
	switch rec := r.(type) {
	case CloseRec:
		return dataFromTermVal(rec), nil
	case WaitRec:
		return dataFromTermCont(rec)
	case SendRec:
		return dataFromTermVal(rec), nil
	case RecvRec:
		return dataFromTermCont(rec)
	case LabRec:
		return dataFromTermVal(rec), nil
	case CaseRec:
		return dataFromTermCont(rec)
	case FwdRec:
		return dataFromTermVal(rec), nil
	default:
		panic(ErrTermTypeUnexpected(rec))
	}
}

func dataToTermRec(dto TermRecData) (TermRec, error) {
	switch dto.K {
	case closeKind:
		return dataToTermVal(dto)
	case waitKind:
		return dataToTermCont(dto)
	case sendKind:
		return dataToTermVal(dto)
	case recvKind:
		return dataToTermCont(dto)
	case labKind:
		return dataToTermVal(dto)
	case caseKind:
		return dataToTermCont(dto)
	case fwdKind:
		return dataToTermVal(dto)
	default:
		panic(errUnexpectedTermKind(dto.K))
	}
}

func dataFromTermVal(v TermVal) TermRecData {
	switch val := v.(type) {
	case CloseRec:
		return TermRecData{
			K:     closeKind,
			Close: &closeRecData{sym.ConvertToString(val.X)},
		}
	case SendRec:
		return TermRecData{
			K: sendKind,
			Send: &sendRecData{
				X: sym.ConvertToString(val.X),
				A: id.ConvertToString(val.A),
				B: id.ConvertToString(val.B),
			},
		}
	case LabRec:
		return TermRecData{
			K:   labKind,
			Lab: &labRecData{sym.ConvertToString(val.X), string(val.Label)},
		}
	case FwdRec:
		return TermRecData{
			K: fwdKind,
			Fwd: &fwdRecData{
				X: sym.ConvertToString(val.X),
				B: id.ConvertToString(val.B),
			},
		}
	default:
		panic(ErrValTypeUnexpected2(val))
	}
}

func dataToTermVal(dto TermRecData) (TermVal, error) {
	switch dto.K {
	case closeKind:
		a, err := sym.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseRec{X: a}, nil
	case sendKind:
		x, err := sym.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		a, err := id.ConvertFromString(dto.Send.A)
		if err != nil {
			return nil, err
		}
		return SendRec{X: x, A: a}, nil
	case labKind:
		a, err := sym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabRec{X: a, Label: sym.ADT(dto.Lab.Label)}, nil
	case fwdKind:
		x, err := sym.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		b, err := id.ConvertFromString(dto.Fwd.B)
		if err != nil {
			return nil, err
		}
		return FwdRec{X: x, B: b}, nil
	default:
		panic(errUnexpectedTermKind(dto.K))
	}
}

func dataFromTermCont(c TermCont) (TermRecData, error) {
	switch cont := c.(type) {
	case WaitRec:
		dto, err := dataFromTermSpec(cont.Cont)
		if err != nil {
			return TermRecData{}, err
		}
		return TermRecData{
			K: waitKind,
			Wait: &waitRecData{
				X:    sym.ConvertToString(cont.X),
				Cont: dto,
			},
		}, nil
	case RecvRec:
		dto, err := dataFromTermSpec(cont.Cont)
		if err != nil {
			return TermRecData{}, err
		}
		return TermRecData{
			K: recvKind,
			Recv: &recvRecData{
				X:    sym.ConvertToString(cont.X),
				Y:    sym.ConvertToString(cont.Y),
				Cont: dto,
			},
		}, nil
	case CaseRec:
		brs := []branchRecData{}
		for l, cont := range cont.Conts {
			dto, err := dataFromTermSpec(cont)
			if err != nil {
				return TermRecData{}, err
			}
			brs = append(brs, branchRecData{Label: string(l), Cont: dto})
		}
		return TermRecData{
			K: caseKind,
			Case: &caseRecData{
				X:        sym.ConvertToString(cont.X),
				Branches: brs,
			},
		}, nil
	case FwdRec:
		return TermRecData{
			K: fwdKind,
			Fwd: &fwdRecData{
				X: sym.ConvertToString(cont.X),
				B: id.ConvertToString(cont.B),
			},
		}, nil
	default:
		panic(ErrContTypeUnexpected2(cont))
	}
}

func dataToTermCont(dto TermRecData) (TermCont, error) {
	switch dto.K {
	case waitKind:
		x, err := sym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToTermSpec(dto.Wait.Cont)
		if err != nil {
			return nil, err
		}
		return WaitRec{X: x, Cont: cont}, nil
	case recvKind:
		x, err := sym.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := dataToTermSpec(dto.Recv.Cont)
		if err != nil {
			return nil, err
		}
		return RecvRec{X: x, Y: y, Cont: cont}, nil
	case caseKind:
		x, err := sym.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[sym.ADT]TermSpec, len(dto.Case.Branches))
		for _, branch := range dto.Case.Branches {
			cont, err := dataToTermSpec(branch.Cont)
			if err != nil {
				return nil, err
			}
			conts[sym.ADT(branch.Label)] = cont
		}
		return CaseRec{X: x, Conts: conts}, nil
	case fwdKind:
		x, err := sym.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		b, err := id.ConvertFromString(dto.Fwd.B)
		if err != nil {
			return nil, err
		}
		return FwdRec{X: x, B: b}, nil
	default:
		panic(errUnexpectedTermKind(dto.K))
	}
}

func dataFromTermSpec(s TermSpec) (TermSpecData, error) {
	switch spec := s.(type) {
	case CloseSpec:
		return TermSpecData{
			K:     closeKind,
			Close: &closeSpecData{sym.ConvertToString(spec.X)},
		}, nil
	case WaitSpec:
		dto, err := dataFromTermSpec(spec.Cont)
		if err != nil {
			return TermSpecData{}, err
		}
		return TermSpecData{
			K: waitKind,
			Wait: &waitSpecData{
				X:    sym.ConvertToString(spec.X),
				Cont: dto,
			},
		}, nil
	case SendSpec:
		return TermSpecData{
			K: sendKind,
			Send: &sendSpecData{
				X: sym.ConvertToString(spec.X),
				Y: sym.ConvertToString(spec.Y),
			},
		}, nil
	case RecvSpec:
		dto, err := dataFromTermSpec(spec.Cont)
		if err != nil {
			return TermSpecData{}, err
		}
		return TermSpecData{
			K: recvKind,
			Recv: &recvSpecData{
				X:    sym.ConvertToString(spec.X),
				Y:    sym.ConvertToString(spec.Y),
				Cont: dto,
			},
		}, nil
	case LabSpec:
		return TermSpecData{
			K:   labKind,
			Lab: &labSpecData{sym.ConvertToString(spec.X), string(spec.Label)},
		}, nil
	case CaseSpec:
		brs := []branchSpecData{}
		for l, cont := range spec.Conts {
			dto, err := dataFromTermSpec(cont)
			if err != nil {
				return TermSpecData{}, err
			}
			brs = append(brs, branchSpecData{Label: string(l), Cont: dto})
		}
		return TermSpecData{
			K: caseKind,
			Case: &caseSpecData{
				X:        sym.ConvertToString(spec.X),
				Branches: brs,
			},
		}, nil
	case FwdSpec:
		return TermSpecData{
			K: fwdKind,
			Fwd: &fwdSpecData{
				X: sym.ConvertToString(spec.X),
				Y: sym.ConvertToString(spec.Y),
			},
		}, nil
	default:
		panic(ErrTermTypeUnexpected(spec))
	}
}

func dataToTermSpec(dto TermSpecData) (TermSpec, error) {
	switch dto.K {
	case closeKind:
		a, err := sym.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseSpec{X: a}, nil
	case waitKind:
		x, err := sym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToTermSpec(dto.Wait.Cont)
		if err != nil {
			return nil, err
		}
		return WaitSpec{X: x, Cont: cont}, nil
	case sendKind:
		x, err := sym.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Send.Y)
		if err != nil {
			return nil, err
		}
		return SendSpec{X: x, Y: y}, nil
	case recvKind:
		x, err := sym.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := dataToTermSpec(dto.Recv.Cont)
		if err != nil {
			return nil, err
		}
		return RecvSpec{X: x, Y: y, Cont: cont}, nil
	case labKind:
		x, err := sym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabSpec{X: x, Label: sym.ADT(dto.Lab.Label)}, nil
	case caseKind:
		x, err := sym.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[sym.ADT]TermSpec, len(dto.Case.Branches))
		for _, b := range dto.Case.Branches {
			cont, err := dataToTermSpec(b.Cont)
			if err != nil {
				return nil, err
			}
			conts[sym.ADT(b.Label)] = cont
		}
		return CaseSpec{X: x, Conts: conts}, nil
	case fwdKind:
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
		panic(errUnexpectedTermKind(dto.K))
	}
}

func errUnexpectedTermKind(k termKind) error {
	return fmt.Errorf("unexpected term kind: %v", k)
}

func errUnexpectedStepKind(k semKind) error {
	return fmt.Errorf("unexpected step kind: %v", k)
}
