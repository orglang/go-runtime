package procexp

import (
	"fmt"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
)

func MsgFromExpSpecNilable(spec ExpSpec) *ExpSpecME {
	if spec == nil {
		return nil
	}
	dto := MsgFromExpSpec(spec)
	return &dto
}

func MsgFromExpSpec(t ExpSpec) ExpSpecME {
	switch term := t.(type) {
	case CloseSpec:
		return ExpSpecME{
			K: Close,
			Close: &CloseSpecME{
				CommPH: qualsym.ConvertToString(term.CommPH),
			},
		}
	case WaitSpec:
		return ExpSpecME{
			K: Wait,
			Wait: &WaitSpecME{
				CommPH: qualsym.ConvertToString(term.CommPH),
				ContES: MsgFromExpSpec(term.ContES),
			},
		}
	case SendSpec:
		return ExpSpecME{
			K: Send,
			Send: &SendSpecME{
				CommPH: qualsym.ConvertToString(term.CommPH),
				ValPH:  qualsym.ConvertToString(term.ValPH),
			},
		}
	case RecvSpec:
		return ExpSpecME{
			K: Recv,
			Recv: &RecvSpecME{
				CommPH: qualsym.ConvertToString(term.CommPH),
				BindPH: qualsym.ConvertToString(term.CommPH),
				ContES: MsgFromExpSpec(term.ContES),
			},
		}
	case LabSpec:
		return ExpSpecME{
			K: Lab,
			Lab: &LabSpecME{
				CommPH: qualsym.ConvertToString(term.CommPH),
				Label:  string(term.Label),
			},
		}
	case CaseSpec:
		brs := []BranchSpecME{}
		for l, t := range term.ContESs {
			brs = append(brs, BranchSpecME{Label: string(l), ContES: MsgFromExpSpec(t)})
		}
		return ExpSpecME{
			K: Case,
			Case: &CaseSpecME{
				CommPH:  qualsym.ConvertToString(term.CommPH),
				ContBSs: brs,
			},
		}
	case SpawnSpec:
		return ExpSpecME{
			K: Spawn,
			Spawn: &SpawnSpecME{
				CommPH:  qualsym.ConvertToString(term.CommPH),
				ProcQN:  qualsym.ConvertToString(term.ProcQN),
				BindPHs: qualsym.ConvertToStrings(term.BindPHs),
				ContES:  MsgFromExpSpecNilable(term.ContES),
			},
		}
	case FwdSpec:
		return ExpSpecME{
			K: Fwd,
			Fwd: &FwdSpecME{
				X: qualsym.ConvertToString(term.X),
				Y: qualsym.ConvertToString(term.Y),
			},
		}
	default:
		panic(ErrExpTypeUnexpected(t))
	}
}

func MsgToExpSpecNilable(dto *ExpSpecME) (ExpSpec, error) {
	if dto == nil {
		return nil, nil
	}
	return MsgToExpSpec(*dto)
}

func MsgToExpSpec(dto ExpSpecME) (ExpSpec, error) {
	switch dto.K {
	case Close:
		x, err := qualsym.ConvertFromString(dto.Close.CommPH)
		if err != nil {
			return nil, err
		}
		return CloseSpec{CommPH: x}, nil
	case Wait:
		x, err := qualsym.ConvertFromString(dto.Wait.CommPH)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToExpSpec(dto.Wait.ContES)
		if err != nil {
			return nil, err
		}
		return WaitSpec{CommPH: x, ContES: cont}, nil
	case Send:
		x, err := qualsym.ConvertFromString(dto.Send.CommPH)
		if err != nil {
			return nil, err
		}
		y, err := qualsym.ConvertFromString(dto.Send.ValPH)
		if err != nil {
			return nil, err
		}
		return SendSpec{CommPH: x, ValPH: y}, nil
	case Recv:
		x, err := qualsym.ConvertFromString(dto.Recv.CommPH)
		if err != nil {
			return nil, err
		}
		y, err := qualsym.ConvertFromString(dto.Recv.BindPH)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToExpSpec(dto.Recv.ContES)
		if err != nil {
			return nil, err
		}
		return RecvSpec{CommPH: x, BindPH: y, ContES: cont}, nil
	case Lab:
		x, err := qualsym.ConvertFromString(dto.Lab.CommPH)
		if err != nil {
			return nil, err
		}
		return LabSpec{CommPH: x, Label: qualsym.ADT(dto.Lab.Label)}, nil
	case Case:
		x, err := qualsym.ConvertFromString(dto.Case.CommPH)
		if err != nil {
			return nil, err
		}
		conts := make(map[qualsym.ADT]ExpSpec, len(dto.Case.ContBSs))
		for _, b := range dto.Case.ContBSs {
			cont, err := MsgToExpSpec(b.ContES)
			if err != nil {
				return nil, err
			}
			conts[qualsym.ADT(b.Label)] = cont
		}
		return CaseSpec{CommPH: x, ContESs: conts}, nil
	case Spawn:
		commPH, err := qualsym.ConvertFromString(dto.Spawn.CommPH)
		if err != nil {
			return nil, err
		}
		procQN, err := qualsym.ConvertFromString(dto.Spawn.ProcQN)
		if err != nil {
			return nil, err
		}
		bindPHs, err := qualsym.ConvertFromStrings(dto.Spawn.BindPHs)
		if err != nil {
			return nil, err
		}
		contES, err := MsgToExpSpecNilable(dto.Spawn.ContES)
		if err != nil {
			return nil, err
		}
		return SpawnSpec{CommPH: commPH, ProcQN: procQN, BindPHs: bindPHs, ContES: contES}, nil
	case Call:
		commPH, err := qualsym.ConvertFromString(dto.Call.CommPH)
		if err != nil {
			return nil, err
		}
		procQN, err := qualsym.ConvertFromString(dto.Call.ProcQN)
		if err != nil {
			return nil, err
		}
		valPHs, err := qualsym.ConvertFromStrings(dto.Call.ValPHs)
		if err != nil {
			return nil, err
		}
		return CallSpec{CommPH: commPH, ProcQN: procQN, ValPHs: valPHs}, nil
	case Fwd:
		x, err := qualsym.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		y, err := qualsym.ConvertFromString(dto.Fwd.Y)
		if err != nil {
			return nil, err
		}
		return FwdSpec{X: x, Y: y}, nil
	default:
		panic(ErrUnexpectedExpKind(dto.K))
	}
}

func ErrUnexpectedExpKind(k expKindME) error {
	return fmt.Errorf("unexpected term kind: %v", k)
}

func DataFromExpRec(r ExpRec) (ExpRecDS, error) {
	switch rec := r.(type) {
	case CloseRec:
		return ExpRecDS{
			K:     closeKind,
			Close: &closeRecDS{qualsym.ConvertToString(rec.X)},
		}, nil
	case WaitRec:
		dto, err := dataFromExpSpec(rec.ContES)
		if err != nil {
			return ExpRecDS{}, err
		}
		return ExpRecDS{
			K: waitKind,
			Wait: &waitRecDS{
				X:      qualsym.ConvertToString(rec.X),
				ContES: dto,
			},
		}, nil
	case SendRec:
		return ExpRecDS{
			K: sendKind,
			Send: &sendRecDS{
				X: qualsym.ConvertToString(rec.X),
				A: identity.ConvertToString(rec.A),
				B: identity.ConvertToString(rec.B),
			},
		}, nil
	case RecvRec:
		dto, err := dataFromExpSpec(rec.ContES)
		if err != nil {
			return ExpRecDS{}, err
		}
		return ExpRecDS{
			K: recvKind,
			Recv: &recvRecDS{
				X:      qualsym.ConvertToString(rec.X),
				Y:      qualsym.ConvertToString(rec.Y),
				ContES: dto,
			},
		}, nil
	case LabRec:
		return ExpRecDS{
			K:   labKind,
			Lab: &labRecDS{qualsym.ConvertToString(rec.X), string(rec.Label)},
		}, nil
	case CaseRec:
		brs := []branchRecDS{}
		for l, cont := range rec.ContESs {
			dto, err := dataFromExpSpec(cont)
			if err != nil {
				return ExpRecDS{}, err
			}
			brs = append(brs, branchRecDS{Label: string(l), ContES: dto})
		}
		return ExpRecDS{
			K: caseKind,
			Case: &caseRecDS{
				X:        qualsym.ConvertToString(rec.X),
				Branches: brs,
			},
		}, nil
	case FwdRec:
		return ExpRecDS{
			K: fwdKind,
			Fwd: &fwdRecDS{
				X: qualsym.ConvertToString(rec.X),
				B: identity.ConvertToString(rec.B),
			},
		}, nil
	default:
		panic(ErrExpTypeUnexpected(rec))
	}
}

func DataToExpRec(dto ExpRecDS) (ExpRec, error) {
	switch dto.K {
	case closeKind:
		a, err := qualsym.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseRec{X: a}, nil
	case waitKind:
		x, err := qualsym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Wait.ContES)
		if err != nil {
			return nil, err
		}
		return WaitRec{X: x, ContES: cont}, nil
	case sendKind:
		x, err := qualsym.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		a, err := identity.ConvertFromString(dto.Send.A)
		if err != nil {
			return nil, err
		}
		return SendRec{X: x, A: a}, nil
	case recvKind:
		x, err := qualsym.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := qualsym.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Recv.ContES)
		if err != nil {
			return nil, err
		}
		return RecvRec{X: x, Y: y, ContES: cont}, nil
	case labKind:
		a, err := qualsym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabRec{X: a, Label: qualsym.ADT(dto.Lab.Label)}, nil
	case caseKind:
		x, err := qualsym.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[qualsym.ADT]ExpSpec, len(dto.Case.Branches))
		for _, branch := range dto.Case.Branches {
			cont, err := dataToExpSpec(branch.ContES)
			if err != nil {
				return nil, err
			}
			conts[qualsym.ADT(branch.Label)] = cont
		}
		return CaseRec{X: x, ContESs: conts}, nil
	case fwdKind:
		x, err := qualsym.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		b, err := identity.ConvertFromString(dto.Fwd.B)
		if err != nil {
			return nil, err
		}
		return FwdRec{X: x, B: b}, nil
	default:
		panic(errUnexpectedExpKind(dto.K))
	}
}

func dataFromExpSpec(s ExpSpec) (ExpSpecDS, error) {
	switch spec := s.(type) {
	case CloseSpec:
		return ExpSpecDS{
			K:     closeKind,
			Close: &closeSpecDS{qualsym.ConvertToString(spec.CommPH)},
		}, nil
	case WaitSpec:
		dto, err := dataFromExpSpec(spec.ContES)
		if err != nil {
			return ExpSpecDS{}, err
		}
		return ExpSpecDS{
			K: waitKind,
			Wait: &waitSpecDS{
				X:      qualsym.ConvertToString(spec.CommPH),
				ContES: dto,
			},
		}, nil
	case SendSpec:
		return ExpSpecDS{
			K: sendKind,
			Send: &sendSpecDS{
				X: qualsym.ConvertToString(spec.CommPH),
				Y: qualsym.ConvertToString(spec.ValPH),
			},
		}, nil
	case RecvSpec:
		dto, err := dataFromExpSpec(spec.ContES)
		if err != nil {
			return ExpSpecDS{}, err
		}
		return ExpSpecDS{
			K: recvKind,
			Recv: &recvSpecDS{
				X:      qualsym.ConvertToString(spec.CommPH),
				Y:      qualsym.ConvertToString(spec.CommPH),
				ContES: dto,
			},
		}, nil
	case LabSpec:
		return ExpSpecDS{
			K:   labKind,
			Lab: &labSpecDS{qualsym.ConvertToString(spec.CommPH), string(spec.Label)},
		}, nil
	case CaseSpec:
		brs := []branchSpecDS{}
		for l, cont := range spec.ContESs {
			dto, err := dataFromExpSpec(cont)
			if err != nil {
				return ExpSpecDS{}, err
			}
			brs = append(brs, branchSpecDS{Label: string(l), ContES: dto})
		}
		return ExpSpecDS{
			K: caseKind,
			Case: &caseSpecDS{
				X:        qualsym.ConvertToString(spec.CommPH),
				Branches: brs,
			},
		}, nil
	case FwdSpec:
		return ExpSpecDS{
			K: fwdKind,
			Fwd: &fwdSpecDS{
				X: qualsym.ConvertToString(spec.X),
				Y: qualsym.ConvertToString(spec.Y),
			},
		}, nil
	default:
		panic(ErrExpTypeUnexpected(spec))
	}
}

func dataToExpSpec(dto ExpSpecDS) (ExpSpec, error) {
	switch dto.K {
	case closeKind:
		a, err := qualsym.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseSpec{CommPH: a}, nil
	case waitKind:
		x, err := qualsym.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Wait.ContES)
		if err != nil {
			return nil, err
		}
		return WaitSpec{CommPH: x, ContES: cont}, nil
	case sendKind:
		x, err := qualsym.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		y, err := qualsym.ConvertFromString(dto.Send.Y)
		if err != nil {
			return nil, err
		}
		return SendSpec{CommPH: x, ValPH: y}, nil
	case recvKind:
		x, err := qualsym.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := qualsym.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Recv.ContES)
		if err != nil {
			return nil, err
		}
		return RecvSpec{CommPH: x, BindPH: y, ContES: cont}, nil
	case labKind:
		x, err := qualsym.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		return LabSpec{CommPH: x, Label: qualsym.ADT(dto.Lab.Label)}, nil
	case caseKind:
		x, err := qualsym.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[qualsym.ADT]ExpSpec, len(dto.Case.Branches))
		for _, b := range dto.Case.Branches {
			cont, err := dataToExpSpec(b.ContES)
			if err != nil {
				return nil, err
			}
			conts[qualsym.ADT(b.Label)] = cont
		}
		return CaseSpec{CommPH: x, ContESs: conts}, nil
	case fwdKind:
		x, err := qualsym.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		y, err := qualsym.ConvertFromString(dto.Fwd.Y)
		if err != nil {
			return nil, err
		}
		return FwdSpec{X: x, Y: y}, nil
	default:
		panic(errUnexpectedExpKind(dto.K))
	}
}

func errUnexpectedExpKind(k expKindDS) error {
	return fmt.Errorf("unexpected term kind: %v", k)
}
