package def

import (
	"fmt"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
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
// goverter:extend Data.*
// goverter:extend data.*
var (
	DataToTermSpecs   func([]TermRecData) ([]TermSpec, error)
	DataFromTermSpecs func([]TermSpec) ([]TermRecData, error)
	DataToTermRecs    func([]TermRecData) ([]TermRec, error)
	DataFromTermRecs  func([]TermRec) ([]TermRecData, error)
)

func DataFromTermRec(r TermRec) (TermRecData, error) {
	switch rec := r.(type) {
	case CloseRec:
		return TermRecData{
			K:     closeKind,
			Close: &closeRecData{sym.ConvertToString(rec.X)},
		}, nil
	case WaitRec:
		dto, err := dataFromTermSpec(rec.Cont)
		if err != nil {
			return TermRecData{}, err
		}
		return TermRecData{
			K: waitKind,
			Wait: &waitRecData{
				X:    sym.ConvertToString(rec.X),
				Cont: dto,
			},
		}, nil
	case SendRec:
		return TermRecData{
			K: sendKind,
			Send: &sendRecData{
				X: sym.ConvertToString(rec.X),
				A: id.ConvertToString(rec.A),
				B: id.ConvertToString(rec.B),
			},
		}, nil
	case RecvRec:
		dto, err := dataFromTermSpec(rec.Cont)
		if err != nil {
			return TermRecData{}, err
		}
		return TermRecData{
			K: recvKind,
			Recv: &recvRecData{
				X:    sym.ConvertToString(rec.X),
				Y:    sym.ConvertToString(rec.Y),
				Cont: dto,
			},
		}, nil
	case LabRec:
		return TermRecData{
			K:   labKind,
			Lab: &labRecData{sym.ConvertToString(rec.X), string(rec.Label)},
		}, nil
	case CaseRec:
		brs := []branchRecData{}
		for l, cont := range rec.Conts {
			dto, err := dataFromTermSpec(cont)
			if err != nil {
				return TermRecData{}, err
			}
			brs = append(brs, branchRecData{Label: string(l), Cont: dto})
		}
		return TermRecData{
			K: caseKind,
			Case: &caseRecData{
				X:        sym.ConvertToString(rec.X),
				Branches: brs,
			},
		}, nil
	case FwdRec:
		return TermRecData{
			K: fwdKind,
			Fwd: &fwdRecData{
				X: sym.ConvertToString(rec.X),
				B: id.ConvertToString(rec.B),
			},
		}, nil
	default:
		panic(ErrTermTypeUnexpected(rec))
	}
}

func DataToTermRec(dto TermRecData) (TermRec, error) {
	switch dto.K {
	case closeKind:
		a, err := sym.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseRec{X: a}, nil
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
	case labKind:
		a, err := sym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabRec{X: a, Label: sym.ADT(dto.Lab.Label)}, nil
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
			Close: &closeSpecData{sym.ConvertToString(spec.CommPH)},
		}, nil
	case WaitSpec:
		dto, err := dataFromTermSpec(spec.ContTS)
		if err != nil {
			return TermSpecData{}, err
		}
		return TermSpecData{
			K: waitKind,
			Wait: &waitSpecData{
				X:    sym.ConvertToString(spec.CommPH),
				Cont: dto,
			},
		}, nil
	case SendSpec:
		return TermSpecData{
			K: sendKind,
			Send: &sendSpecData{
				X: sym.ConvertToString(spec.CommPH),
				Y: sym.ConvertToString(spec.ValPH),
			},
		}, nil
	case RecvSpec:
		dto, err := dataFromTermSpec(spec.ContTS)
		if err != nil {
			return TermSpecData{}, err
		}
		return TermSpecData{
			K: recvKind,
			Recv: &recvSpecData{
				X:    sym.ConvertToString(spec.CommPH),
				Y:    sym.ConvertToString(spec.BindPH),
				Cont: dto,
			},
		}, nil
	case LabSpec:
		return TermSpecData{
			K:   labKind,
			Lab: &labSpecData{sym.ConvertToString(spec.CommPH), string(spec.Label)},
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
				X:        sym.ConvertToString(spec.CommPH),
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
		return CloseSpec{CommPH: a}, nil
	case waitKind:
		x, err := sym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToTermSpec(dto.Wait.Cont)
		if err != nil {
			return nil, err
		}
		return WaitSpec{CommPH: x, ContTS: cont}, nil
	case sendKind:
		x, err := sym.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Send.Y)
		if err != nil {
			return nil, err
		}
		return SendSpec{CommPH: x, ValPH: y}, nil
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
		return RecvSpec{CommPH: x, BindPH: y, ContTS: cont}, nil
	case labKind:
		x, err := sym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabSpec{CommPH: x, Label: sym.ADT(dto.Lab.Label)}, nil
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
		return CaseSpec{CommPH: x, Conts: conts}, nil
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
