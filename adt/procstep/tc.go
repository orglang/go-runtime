package procstep

import (
	"fmt"

	"orglang/orglang/adt/procexp"
)

func dataFromStepRec(r StepRec) (StepRecDS, error) {
	if r == nil {
		return StepRecDS{}, nil
	}
	switch rec := r.(type) {
	case MsgRec:
		msgVal, err := procexp.DataFromExpRec(rec.ValER)
		if err != nil {
			return StepRecDS{}, err
		}
		return StepRecDS{
			K:      msgStep,
			ProcER: msgVal,
		}, nil
	case SvcRec:
		svcCont, err := procexp.DataFromExpRec(rec.ContER)
		if err != nil {
			return StepRecDS{}, err
		}
		return StepRecDS{
			K:      svcStep,
			ProcER: svcCont,
		}, nil
	default:
		panic(ErrRecTypeUnexpected(rec))
	}
}

func dataToStepRec(dto StepRecDS) (StepRec, error) {
	var nilData StepRecDS
	if dto == nilData {
		return nil, nil
	}
	switch dto.K {
	case msgStep:
		val, err := procexp.DataToExpRec(dto.ProcER)
		if err != nil {
			return nil, err
		}
		return MsgRec{ValER: val}, nil
	case svcStep:
		cont, err := procexp.DataToExpRec(dto.ProcER)
		if err != nil {
			return nil, err
		}
		return SvcRec{ContER: cont}, nil
	default:
		panic(errUnexpectedStepKind(dto.K))
	}
}

func errUnexpectedStepKind(k stepKindDS) error {
	return fmt.Errorf("unexpected step kind: %v", k)
}
