package step

import (
	"database/sql"
	"fmt"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

type RootData struct {
	ID   string         `db:"id"`
	K    stepKind       `db:"kind"`
	PID  sql.NullString `db:"pid"`
	VID  sql.NullString `db:"vid"`
	Spec specData       `db:"spec"`
}

type stepKind int

const (
	nonstep = stepKind(iota)
	proc
	msg
	srv
)

type specData struct {
	K     termKind   `json:"k"`
	Close *closeData `json:"close,omitempty"`
	Wait  *waitData  `json:"wait,omitempty"`
	Send  *sendData  `json:"send,omitempty"`
	Recv  *recvData  `json:"recv,omitempty"`
	Lab   *labData   `json:"lab,omitempty"`
	Case  *caseData  `json:"case,omitempty"`
	Fwd   *fwdData   `json:"fwd,omitempty"`
	CTA   *ctaData   `json:"cta,omitempty"`
}

type closeData struct {
	A string `json:"a"`
}

type waitData struct {
	X    string   `json:"x"`
	Cont specData `json:"cont"`
}

type sendData struct {
	A string `json:"a"`
	B string `json:"b"`
}

type recvData struct {
	X    string   `json:"x"`
	Y    string   `json:"y"`
	Cont specData `json:"cont"`
}

type labData struct {
	A string `json:"a"`
	L string `json:"l"`
}

type caseData struct {
	X   string       `json:"x"`
	Brs []branchData `json:"brs"`
}

type branchData struct {
	L    string   `json:"l"`
	Cont specData `json:"cont"`
}

type fwdData struct {
	C string `json:"c"`
	D string `json:"d"`
}

type ctaData struct {
	AK  string `json:"ak"`
	Sig string `json:"sig"`
}

type termKind int

const (
	nonterm = termKind(iota)
	close
	wait
	send
	recv
	lab
	caze
	cta
	link
	spawn
	fwd
)

// goverter:variables
// goverter:output:format assign-variable
// goverter:extend data.*
var (
	DataToRoots    func([]RootData) ([]Root, error)
	DataFromRoots  func([]Root) ([]RootData, error)
	DataToTerms    func([]specData) ([]Term, error)
	DataFromTerms  func([]Term) ([]specData, error)
	DataToValues   func([]specData) ([]Value, error)
	DataFromValues func([]Value) []specData
	DataToConts    func([]specData) ([]Continuation, error)
	DataFromConts  func([]Continuation) ([]specData, error)
)

func dataFromRoot(r Root) (RootData, error) {
	if r == nil {
		return RootData{}, nil
	}
	switch root := r.(type) {
	case ProcRoot:
		pid := id.ConvertToNullString(root.PID)
		spec, err := dataFromTerm(root.Term)
		if err != nil {
			return RootData{}, err
		}
		return RootData{
			K:    proc,
			ID:   root.ID.String(),
			PID:  pid,
			Spec: spec,
		}, nil
	case MsgRoot:
		pid := id.ConvertToNullString(root.PID)
		vid := id.ConvertToNullString(root.VID)
		return RootData{
			K:    msg,
			ID:   root.ID.String(),
			PID:  pid,
			VID:  vid,
			Spec: dataFromValue(root.Val),
		}, nil
	case SrvRoot:
		pid := id.ConvertToNullString(root.PID)
		vid := id.ConvertToNullString(root.VID)
		spec, err := dataFromCont(root.Cont)
		if err != nil {
			return RootData{}, err
		}
		return RootData{
			K:    srv,
			ID:   root.ID.String(),
			PID:  pid,
			VID:  vid,
			Spec: spec,
		}, nil
	default:
		panic(ErrRootTypeUnexpected(root))
	}
}

func dataToRoot(dto RootData) (Root, error) {
	var nilData RootData
	if dto == nilData {
		return nil, nil
	}
	ident, err := id.ConvertFromString(dto.ID)
	if err != nil {
		return nil, err
	}
	pid, err := id.ConvertFromNullString(dto.PID)
	if err != nil {
		return nil, err
	}
	vid, err := id.ConvertFromNullString(dto.VID)
	if err != nil {
		return nil, err
	}
	switch dto.K {
	case proc:
		term, err := dataToTerm(dto.Spec)
		if err != nil {
			return nil, err
		}
		return ProcRoot{ID: ident, PID: pid, Term: term}, nil
	case msg:
		val, err := dataToValue(dto.Spec)
		if err != nil {
			return nil, err
		}
		return MsgRoot{ID: ident, PID: pid, VID: vid, Val: val}, nil
	case srv:
		cont, err := dataToCont(dto.Spec)
		if err != nil {
			return nil, err
		}
		return SrvRoot{ID: ident, PID: pid, VID: vid, Cont: cont}, nil
	default:
		panic(errUnexpectedStepKind(dto.K))
	}
}

func dataFromTerm(t Term) (specData, error) {
	switch term := t.(type) {
	case CloseSpec:
		return dataFromValue(term), nil
	case WaitSpec:
		return dataFromCont(term)
	case SendSpec:
		return dataFromValue(term), nil
	case RecvSpec:
		return dataFromCont(term)
	case LabSpec:
		return dataFromValue(term), nil
	case CaseSpec:
		return dataFromCont(term)
	case FwdSpec:
		return dataFromValue(term), nil
	default:
		panic(ErrTermTypeUnexpected(term))
	}
}

func dataToTerm(dto specData) (Term, error) {
	switch dto.K {
	case close:
		return dataToValue(dto)
	case wait:
		return dataToCont(dto)
	case send:
		return dataToValue(dto)
	case recv:
		return dataToCont(dto)
	case lab:
		return dataToValue(dto)
	case caze:
		return dataToCont(dto)
	case fwd:
		return dataToValue(dto)
	default:
		panic(errUnexpectedTermKind(dto.K))
	}
}

func dataFromValue(v Value) specData {
	switch val := v.(type) {
	case CloseSpec:
		return specData{
			K:     close,
			Close: &closeData{sym.ConvertToString(val.X)},
		}
	case SendSpec:
		return specData{
			K:    send,
			Send: &sendData{sym.ConvertToString(val.X), sym.ConvertToString(val.Y)},
		}
	case LabSpec:
		return specData{
			K:   lab,
			Lab: &labData{sym.ConvertToString(val.X), string(val.Label)},
		}
	case FwdSpec:
		return specData{
			K: fwd,
			Fwd: &fwdData{
				C: sym.ConvertToString(val.X),
				D: sym.ConvertToString(val.Y),
			},
		}
	default:
		panic(ErrValTypeUnexpected(val))
	}
}

func dataToValue(dto specData) (Value, error) {
	switch dto.K {
	case close:
		a, err := sym.ConvertFromString(dto.Close.A)
		if err != nil {
			return nil, err
		}
		return CloseSpec{X: a}, nil
	case send:
		a, err := sym.ConvertFromString(dto.Send.A)
		if err != nil {
			return nil, err
		}
		b, err := sym.ConvertFromString(dto.Send.B)
		if err != nil {
			return nil, err
		}
		return SendSpec{X: a, Y: b}, nil
	case lab:
		a, err := sym.ConvertFromString(dto.Lab.A)
		if err != nil {
			return nil, err
		}
		return LabSpec{X: a, Label: sym.ADT(dto.Lab.L)}, nil
	case fwd:
		c, err := sym.ConvertFromString(dto.Fwd.C)
		if err != nil {
			return nil, err
		}
		d, err := sym.ConvertFromString(dto.Fwd.D)
		if err != nil {
			return nil, err
		}
		return FwdSpec{X: c, Y: d}, nil
	default:
		panic(errUnexpectedTermKind(dto.K))
	}
}

func dataFromCont(c Continuation) (specData, error) {
	switch cont := c.(type) {
	case WaitSpec:
		dto, err := dataFromTerm(cont.Cont)
		if err != nil {
			return specData{}, err
		}
		return specData{
			K: wait,
			Wait: &waitData{
				X:    sym.ConvertToString(cont.X),
				Cont: dto,
			},
		}, nil
	case RecvSpec:
		dto, err := dataFromTerm(cont.Cont)
		if err != nil {
			return specData{}, err
		}
		return specData{
			K: recv,
			Recv: &recvData{
				X:    sym.ConvertToString(cont.X),
				Y:    sym.ConvertToString(cont.Y),
				Cont: dto,
			},
		}, nil
	case CaseSpec:
		brs := []branchData{}
		for l, cont := range cont.Conts {
			dto, err := dataFromTerm(cont)
			if err != nil {
				return specData{}, err
			}
			brs = append(brs, branchData{L: string(l), Cont: dto})
		}
		return specData{
			K: caze,
			Case: &caseData{
				X:   sym.ConvertToString(cont.X),
				Brs: brs,
			},
		}, nil
	case FwdSpec:
		return specData{
			K: fwd,
			Fwd: &fwdData{
				C: sym.ConvertToString(cont.X),
				D: sym.ConvertToString(cont.Y),
			},
		}, nil
	default:
		panic(ErrContTypeUnexpected(cont))
	}
}

func dataToCont(dto specData) (Continuation, error) {
	switch dto.K {
	case wait:
		x, err := sym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToTerm(dto.Wait.Cont)
		if err != nil {
			return nil, err
		}
		return WaitSpec{X: x, Cont: cont}, nil
	case recv:
		x, err := sym.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := sym.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := dataToTerm(dto.Recv.Cont)
		if err != nil {
			return nil, err
		}
		return RecvSpec{X: x, Y: y, Cont: cont}, nil
	case caze:
		x, err := sym.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[sym.ADT]Term, len(dto.Case.Brs))
		for _, b := range dto.Case.Brs {
			cont, err := dataToTerm(b.Cont)
			if err != nil {
				return nil, err
			}
			conts[sym.ADT(b.L)] = cont
		}
		return CaseSpec{X: x, Conts: conts}, nil
	case fwd:
		c, err := sym.ConvertFromString(dto.Fwd.C)
		if err != nil {
			return nil, err
		}
		d, err := sym.ConvertFromString(dto.Fwd.D)
		if err != nil {
			return nil, err
		}
		return FwdSpec{X: c, Y: d}, nil
	default:
		panic(errUnexpectedTermKind(dto.K))
	}
}

func errUnexpectedTermKind(k termKind) error {
	return fmt.Errorf("unexpected term kind: %v", k)
}

func errUnexpectedStepKind(k stepKind) error {
	return fmt.Errorf("unexpected step kind: %v", k)
}
