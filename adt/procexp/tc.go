package procexp

import (
	"fmt"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/symbol"
	"orglang/go-runtime/adt/uniqsym"

	"github.com/orglang/go-sdk/adt/procexp"
)

func MsgFromExpSpecNilable(spec ExpSpec) *procexp.ExpSpecME {
	if spec == nil {
		return nil
	}
	dto := MsgFromExpSpec(spec)
	return &dto
}

func MsgFromExpSpec(s ExpSpec) procexp.ExpSpecME {
	switch spec := s.(type) {
	case CloseSpec:
		return procexp.ExpSpecME{
			K: procexp.Close,
			Close: &procexp.CloseSpecME{
				CommPH: symbol.ConvertToString(spec.CommChnlPH),
			},
		}
	case WaitSpec:
		return procexp.ExpSpecME{
			K: procexp.Wait,
			Wait: &procexp.WaitSpecME{
				CommPH: symbol.ConvertToString(spec.CommChnlPH),
				ContES: MsgFromExpSpec(spec.ContES),
			},
		}
	case SendSpec:
		return procexp.ExpSpecME{
			K: procexp.Send,
			Send: &procexp.SendSpecME{
				CommPH: symbol.ConvertToString(spec.CommChnlPH),
				ValPH:  symbol.ConvertToString(spec.ValChnlPH),
			},
		}
	case RecvSpec:
		return procexp.ExpSpecME{
			K: procexp.Recv,
			Recv: &procexp.RecvSpecME{
				CommPH: symbol.ConvertToString(spec.CommChnlPH),
				BindPH: symbol.ConvertToString(spec.CommChnlPH),
				ContES: MsgFromExpSpec(spec.ContES),
			},
		}
	case LabSpec:
		return procexp.ExpSpecME{
			K: procexp.Lab,
			Lab: &procexp.LabSpecME{
				CommPH: symbol.ConvertToString(spec.CommChnlPH),
				Label:  uniqsym.ConvertToString(spec.LabelQN),
			},
		}
	case CaseSpec:
		brs := []procexp.BranchSpecME{}
		for l, t := range spec.ContESs {
			brs = append(brs, procexp.BranchSpecME{Label: uniqsym.ConvertToString(l), ContES: MsgFromExpSpec(t)})
		}
		return procexp.ExpSpecME{
			K: procexp.Case,
			Case: &procexp.CaseSpecME{
				CommPH:  symbol.ConvertToString(spec.CommChnlPH),
				ContBSs: brs,
			},
		}
	case SpawnSpec:
		return procexp.ExpSpecME{
			K: procexp.Spawn,
			Spawn: &procexp.SpawnSpecME{
				CommPH:  symbol.ConvertToString(spec.CommChnlPH),
				ProcQN:  uniqsym.ConvertToString(spec.ProcQN),
				BindPHs: symbol.ConvertToStrings(spec.BindChnlPHs),
				ContES:  MsgFromExpSpecNilable(spec.ContES),
			},
		}
	case FwdSpec:
		return procexp.ExpSpecME{
			K: procexp.Fwd,
			Fwd: &procexp.FwdSpecME{
				X: symbol.ConvertToString(spec.CommChnlPH),
				Y: symbol.ConvertToString(spec.ContChnlPH),
			},
		}
	default:
		panic(ErrExpTypeUnexpected(s))
	}
}

func MsgToExpSpecNilable(dto *procexp.ExpSpecME) (ExpSpec, error) {
	if dto == nil {
		return nil, nil
	}
	return MsgToExpSpec(*dto)
}

func MsgToExpSpec(dto procexp.ExpSpecME) (ExpSpec, error) {
	switch dto.K {
	case procexp.Close:
		x, err := symbol.ConvertFromString(dto.Close.CommPH)
		if err != nil {
			return nil, err
		}
		return CloseSpec{CommChnlPH: x}, nil
	case procexp.Wait:
		x, err := symbol.ConvertFromString(dto.Wait.CommPH)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToExpSpec(dto.Wait.ContES)
		if err != nil {
			return nil, err
		}
		return WaitSpec{CommChnlPH: x, ContES: cont}, nil
	case procexp.Send:
		x, err := symbol.ConvertFromString(dto.Send.CommPH)
		if err != nil {
			return nil, err
		}
		y, err := symbol.ConvertFromString(dto.Send.ValPH)
		if err != nil {
			return nil, err
		}
		return SendSpec{CommChnlPH: x, ValChnlPH: y}, nil
	case procexp.Recv:
		x, err := symbol.ConvertFromString(dto.Recv.CommPH)
		if err != nil {
			return nil, err
		}
		y, err := symbol.ConvertFromString(dto.Recv.BindPH)
		if err != nil {
			return nil, err
		}
		cont, err := MsgToExpSpec(dto.Recv.ContES)
		if err != nil {
			return nil, err
		}
		return RecvSpec{CommChnlPH: x, BindChnlPH: y, ContES: cont}, nil
	case procexp.Lab:
		x, err := symbol.ConvertFromString(dto.Lab.CommPH)
		if err != nil {
			return nil, err
		}
		label, err := uniqsym.ConvertFromString(dto.Lab.Label)
		if err != nil {
			return nil, err
		}
		return LabSpec{CommChnlPH: x, LabelQN: label}, nil
	case procexp.Case:
		x, err := symbol.ConvertFromString(dto.Case.CommPH)
		if err != nil {
			return nil, err
		}
		conts := make(map[uniqsym.ADT]ExpSpec, len(dto.Case.ContBSs))
		for _, b := range dto.Case.ContBSs {
			cont, err := MsgToExpSpec(b.ContES)
			if err != nil {
				return nil, err
			}
			label, err := uniqsym.ConvertFromString(dto.Lab.Label)
			if err != nil {
				return nil, err
			}
			conts[label] = cont
		}
		return CaseSpec{CommChnlPH: x, ContESs: conts}, nil
	case procexp.Spawn:
		commPH, err := symbol.ConvertFromString(dto.Spawn.CommPH)
		if err != nil {
			return nil, err
		}
		procQN, err := uniqsym.ConvertFromString(dto.Spawn.ProcQN)
		if err != nil {
			return nil, err
		}
		bindPHs, err := symbol.ConvertFromStrings(dto.Spawn.BindPHs)
		if err != nil {
			return nil, err
		}
		contES, err := MsgToExpSpecNilable(dto.Spawn.ContES)
		if err != nil {
			return nil, err
		}
		return SpawnSpec{CommChnlPH: commPH, ProcQN: procQN, BindChnlPHs: bindPHs, ContES: contES}, nil
	case procexp.Call:
		commPH, err := symbol.ConvertFromString(dto.Call.CommPH)
		if err != nil {
			return nil, err
		}
		procQN, err := uniqsym.ConvertFromString(dto.Call.ProcQN)
		if err != nil {
			return nil, err
		}
		valPHs, err := symbol.ConvertFromStrings(dto.Call.ValPHs)
		if err != nil {
			return nil, err
		}
		return CallSpec{CommChnlPH: commPH, ProcQN: procQN, ValChnlPHs: valPHs}, nil
	case procexp.Fwd:
		x, err := symbol.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		y, err := symbol.ConvertFromString(dto.Fwd.Y)
		if err != nil {
			return nil, err
		}
		return FwdSpec{CommChnlPH: x, ContChnlPH: y}, nil
	default:
		panic(procexp.ErrUnexpectedExpKind(dto.K))
	}
}

func DataFromExpRec(r ExpRec) (ExpRecDS, error) {
	switch rec := r.(type) {
	case CloseRec:
		return ExpRecDS{
			K:     closeExp,
			Close: &closeRecDS{symbol.ConvertToString(rec.CommChnlPH)},
		}, nil
	case WaitRec:
		dto, err := dataFromExpSpec(rec.ContES)
		if err != nil {
			return ExpRecDS{}, err
		}
		return ExpRecDS{
			K: waitExp,
			Wait: &waitRecDS{
				X:      symbol.ConvertToString(rec.CommChnlPH),
				ContES: dto,
			},
		}, nil
	case SendRec:
		return ExpRecDS{
			K: sendExp,
			Send: &sendRecDS{
				X: symbol.ConvertToString(rec.CommChnlPH),
				A: identity.ConvertToString(rec.ContChnlID),
				B: identity.ConvertToString(rec.ValChnlID),
			},
		}, nil
	case RecvRec:
		dto, err := dataFromExpSpec(rec.ContES)
		if err != nil {
			return ExpRecDS{}, err
		}
		return ExpRecDS{
			K: recvExp,
			Recv: &recvRecDS{
				X:      symbol.ConvertToString(rec.CommChnlPH),
				Y:      symbol.ConvertToString(rec.ValChnlPH),
				ContES: dto,
			},
		}, nil
	case LabRec:
		return ExpRecDS{
			K:   labExp,
			Lab: &labRecDS{symbol.ConvertToString(rec.CommChnlPH), uniqsym.ConvertToString(rec.LabelQN)},
		}, nil
	case CaseRec:
		brs := []branchRecDS{}
		for l, cont := range rec.ContESs {
			dto, err := dataFromExpSpec(cont)
			if err != nil {
				return ExpRecDS{}, err
			}
			brs = append(brs, branchRecDS{Label: uniqsym.ConvertToString(l), ContES: dto})
		}
		return ExpRecDS{
			K: caseExp,
			Case: &caseRecDS{
				X:        symbol.ConvertToString(rec.CommChnlPH),
				Branches: brs,
			},
		}, nil
	case FwdRec:
		return ExpRecDS{
			K: fwdExp,
			Fwd: &fwdRecDS{
				X: symbol.ConvertToString(rec.CommChnlPH),
				B: identity.ConvertToString(rec.ContChnlID),
			},
		}, nil
	default:
		panic(ErrExpTypeUnexpected(rec))
	}
}

func DataToExpRec(dto ExpRecDS) (ExpRec, error) {
	switch dto.K {
	case closeExp:
		a, err := symbol.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseRec{CommChnlPH: a}, nil
	case waitExp:
		x, err := symbol.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Wait.ContES)
		if err != nil {
			return nil, err
		}
		return WaitRec{CommChnlPH: x, ContES: cont}, nil
	case sendExp:
		x, err := symbol.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		a, err := identity.ConvertFromString(dto.Send.A)
		if err != nil {
			return nil, err
		}
		return SendRec{CommChnlPH: x, ContChnlID: a}, nil
	case recvExp:
		x, err := symbol.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := symbol.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Recv.ContES)
		if err != nil {
			return nil, err
		}
		return RecvRec{CommChnlPH: x, ValChnlPH: y, ContES: cont}, nil
	case labExp:
		a, err := symbol.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		label, err := uniqsym.ConvertFromString(dto.Lab.Label)
		if err != nil {
			return nil, err
		}
		return LabRec{CommChnlPH: a, LabelQN: label}, nil
	case caseExp:
		x, err := symbol.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[uniqsym.ADT]ExpSpec, len(dto.Case.Branches))
		for _, branch := range dto.Case.Branches {
			cont, err := dataToExpSpec(branch.ContES)
			if err != nil {
				return nil, err
			}
			label, err := uniqsym.ConvertFromString(dto.Lab.Label)
			if err != nil {
				return nil, err
			}
			conts[label] = cont
		}
		return CaseRec{CommChnlPH: x, ContESs: conts}, nil
	case fwdExp:
		x, err := symbol.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		b, err := identity.ConvertFromString(dto.Fwd.B)
		if err != nil {
			return nil, err
		}
		return FwdRec{CommChnlPH: x, ContChnlID: b}, nil
	default:
		panic(errUnexpectedExpKind(dto.K))
	}
}

func dataFromExpSpec(s ExpSpec) (ExpSpecDS, error) {
	switch spec := s.(type) {
	case CloseSpec:
		return ExpSpecDS{
			K:     closeExp,
			Close: &closeSpecDS{symbol.ConvertToString(spec.CommChnlPH)},
		}, nil
	case WaitSpec:
		dto, err := dataFromExpSpec(spec.ContES)
		if err != nil {
			return ExpSpecDS{}, err
		}
		return ExpSpecDS{
			K: waitExp,
			Wait: &waitSpecDS{
				X:      symbol.ConvertToString(spec.CommChnlPH),
				ContES: dto,
			},
		}, nil
	case SendSpec:
		return ExpSpecDS{
			K: sendExp,
			Send: &sendSpecDS{
				X: symbol.ConvertToString(spec.CommChnlPH),
				Y: symbol.ConvertToString(spec.ValChnlPH),
			},
		}, nil
	case RecvSpec:
		dto, err := dataFromExpSpec(spec.ContES)
		if err != nil {
			return ExpSpecDS{}, err
		}
		return ExpSpecDS{
			K: recvExp,
			Recv: &recvSpecDS{
				X:      symbol.ConvertToString(spec.CommChnlPH),
				Y:      symbol.ConvertToString(spec.CommChnlPH),
				ContES: dto,
			},
		}, nil
	case LabSpec:
		return ExpSpecDS{
			K:   labExp,
			Lab: &labSpecDS{symbol.ConvertToString(spec.CommChnlPH), uniqsym.ConvertToString(spec.LabelQN)},
		}, nil
	case CaseSpec:
		brs := []branchSpecDS{}
		for l, cont := range spec.ContESs {
			dto, err := dataFromExpSpec(cont)
			if err != nil {
				return ExpSpecDS{}, err
			}
			brs = append(brs, branchSpecDS{Label: uniqsym.ConvertToString(l), ContES: dto})
		}
		return ExpSpecDS{
			K: caseExp,
			Case: &caseSpecDS{
				X:        symbol.ConvertToString(spec.CommChnlPH),
				Branches: brs,
			},
		}, nil
	case FwdSpec:
		return ExpSpecDS{
			K: fwdExp,
			Fwd: &fwdSpecDS{
				X: symbol.ConvertToString(spec.CommChnlPH),
				Y: symbol.ConvertToString(spec.ContChnlPH),
			},
		}, nil
	default:
		panic(ErrExpTypeUnexpected(spec))
	}
}

func dataToExpSpec(dto ExpSpecDS) (ExpSpec, error) {
	switch dto.K {
	case closeExp:
		a, err := symbol.ConvertFromString(dto.Close.X)
		if err != nil {
			return nil, err
		}
		return CloseSpec{CommChnlPH: a}, nil
	case waitExp:
		x, err := symbol.ConvertFromString(dto.Wait.X)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Wait.ContES)
		if err != nil {
			return nil, err
		}
		return WaitSpec{CommChnlPH: x, ContES: cont}, nil
	case sendExp:
		x, err := symbol.ConvertFromString(dto.Send.X)
		if err != nil {
			return nil, err
		}
		y, err := symbol.ConvertFromString(dto.Send.Y)
		if err != nil {
			return nil, err
		}
		return SendSpec{CommChnlPH: x, ValChnlPH: y}, nil
	case recvExp:
		x, err := symbol.ConvertFromString(dto.Recv.X)
		if err != nil {
			return nil, err
		}
		y, err := symbol.ConvertFromString(dto.Recv.Y)
		if err != nil {
			return nil, err
		}
		cont, err := dataToExpSpec(dto.Recv.ContES)
		if err != nil {
			return nil, err
		}
		return RecvSpec{CommChnlPH: x, BindChnlPH: y, ContES: cont}, nil
	case labExp:
		x, err := symbol.ConvertFromString(dto.Lab.X)
		if err != nil {
			return nil, err
		}
		label, err := uniqsym.ConvertFromString(dto.Lab.Label)
		if err != nil {
			return nil, err
		}
		return LabSpec{CommChnlPH: x, LabelQN: label}, nil
	case caseExp:
		x, err := symbol.ConvertFromString(dto.Case.X)
		if err != nil {
			return nil, err
		}
		conts := make(map[uniqsym.ADT]ExpSpec, len(dto.Case.Branches))
		for _, b := range dto.Case.Branches {
			cont, err := dataToExpSpec(b.ContES)
			if err != nil {
				return nil, err
			}
			label, err := uniqsym.ConvertFromString(dto.Lab.Label)
			if err != nil {
				return nil, err
			}
			conts[label] = cont
		}
		return CaseSpec{CommChnlPH: x, ContESs: conts}, nil
	case fwdExp:
		x, err := symbol.ConvertFromString(dto.Fwd.X)
		if err != nil {
			return nil, err
		}
		y, err := symbol.ConvertFromString(dto.Fwd.Y)
		if err != nil {
			return nil, err
		}
		return FwdSpec{CommChnlPH: x, ContChnlPH: y}, nil
	default:
		panic(errUnexpectedExpKind(dto.K))
	}
}

func errUnexpectedExpKind(k expKindDS) error {
	return fmt.Errorf("unexpected term kind: %v", k)
}
