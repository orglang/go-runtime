package def

import (
	"fmt"

	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/pol"
	"smecalculus/rolevod/lib/sym"
)

type TermSpec interface {
	spec()
}

type OneSpec struct{}

func (OneSpec) spec() {}

// aka TpName
type LinkSpec struct {
	TypeQN sym.ADT
}

func (LinkSpec) spec() {}

type TensorSpec struct {
	B TermSpec
	C TermSpec
}

func (TensorSpec) spec() {}

type LolliSpec struct {
	Y TermSpec
	Z TermSpec
}

func (LolliSpec) spec() {}

// aka Internal Choice
type PlusSpec struct {
	Choices map[sym.ADT]TermSpec
}

func (PlusSpec) spec() {}

// aka External Choice
type WithSpec struct {
	Choices map[sym.ADT]TermSpec
}

func (WithSpec) spec() {}

type UpSpec struct {
	X TermSpec
}

func (UpSpec) spec() {}

type DownSpec struct {
	X TermSpec
}

func (DownSpec) spec() {}

type TermRef interface {
	id.Identifiable
}

type OneRef struct {
	TermID id.ADT
}

func (r OneRef) Ident() id.ADT { return r.TermID }

type LinkRef struct {
	TermID id.ADT
}

func (r LinkRef) Ident() id.ADT { return r.TermID }

type PlusRef struct {
	TermID id.ADT
}

func (r PlusRef) Ident() id.ADT { return r.TermID }

type WithRef struct {
	TermID id.ADT
}

func (r WithRef) Ident() id.ADT { return r.TermID }

type TensorRef struct {
	TermID id.ADT
}

func (r TensorRef) Ident() id.ADT { return r.TermID }

type LolliRef struct {
	TermID id.ADT
}

func (r LolliRef) Ident() id.ADT { return r.TermID }

type UpRef struct {
	TermID id.ADT
}

func (r UpRef) Ident() id.ADT { return r.TermID }

type DownRef struct {
	TermID id.ADT
}

func (r DownRef) Ident() id.ADT { return r.TermID }

// aka Stype
type TermRec interface {
	id.Identifiable
	pol.Polarizable
}

type ProdRec interface {
	Next() id.ADT
}

type SumRec interface {
	Next(sym.ADT) id.ADT
}

type OneRec struct {
	TermID id.ADT
}

func (OneRec) spec() {}

func (r OneRec) Ident() id.ADT { return r.TermID }

func (OneRec) Pol() pol.ADT { return pol.Pos }

// aka TpName
type LinkRec struct {
	TermID id.ADT
	TypeQN sym.ADT
}

func (LinkRec) spec() {}

func (r LinkRec) Ident() id.ADT { return r.TermID }

func (LinkRec) Pol() pol.ADT { return pol.Zero }

// aka Internal Choice
type PlusRec struct {
	TermID  id.ADT
	Choices map[sym.ADT]TermRec
}

func (PlusRec) spec() {}

func (r PlusRec) Ident() id.ADT { return r.TermID }

func (r PlusRec) Next(l sym.ADT) id.ADT { return r.Choices[l].Ident() }

func (PlusRec) Pol() pol.ADT { return pol.Pos }

// aka External Choice
type WithRec struct {
	TermID  id.ADT
	Choices map[sym.ADT]TermRec
}

func (WithRec) spec() {}

func (r WithRec) Ident() id.ADT { return r.TermID }

func (r WithRec) Next(l sym.ADT) id.ADT { return r.Choices[l].Ident() }

func (WithRec) Pol() pol.ADT { return pol.Neg }

type TensorRec struct {
	TermID id.ADT
	B      TermRec // value
	C      TermRec // cont
}

func (TensorRec) spec() {}

func (r TensorRec) Ident() id.ADT { return r.TermID }

func (r TensorRec) Next() id.ADT { return r.C.Ident() }

func (TensorRec) Pol() pol.ADT { return pol.Pos }

type LolliRec struct {
	TermID id.ADT
	Y      TermRec // value
	Z      TermRec // cont
}

func (LolliRec) spec() {}

func (r LolliRec) Ident() id.ADT { return r.TermID }

func (r LolliRec) Next() id.ADT { return r.Z.Ident() }

func (LolliRec) Pol() pol.ADT { return pol.Neg }

type UpRec struct {
	TermID id.ADT
	X      TermRec
}

func (UpRec) spec() {}

func (r UpRec) Ident() id.ADT { return r.TermID }

func (UpRec) Pol() pol.ADT { return pol.Zero }

type DownRec struct {
	TermID id.ADT
	X      TermRec
}

func (DownRec) spec() {}

func (r DownRec) Ident() id.ADT { return r.TermID }

func (DownRec) Pol() pol.ADT { return pol.Zero }

type Context struct {
	Assets map[sym.ADT]TermRec
	Liabs  map[sym.ADT]TermRec
}

type TermRepo interface {
	Insert(data.Source, TermRec) error
	SelectAll(data.Source) ([]TermRef, error)
	SelectByID(data.Source, id.ADT) (TermRec, error)
	SelectByIDs(data.Source, []id.ADT) ([]TermRec, error)
	SelectEnv(data.Source, []id.ADT) (map[id.ADT]TermRec, error)
}

func ConvertSpecToRec(s TermSpec) TermRec {
	if s == nil {
		return nil
	}
	switch spec := s.(type) {
	case OneSpec:
		return OneRec{TermID: id.New()}
	case LinkSpec:
		return LinkRec{TermID: id.New(), TypeQN: spec.TypeQN}
	case TensorSpec:
		return TensorRec{
			TermID: id.New(),
			B:      ConvertSpecToRec(spec.B),
			C:      ConvertSpecToRec(spec.C),
		}
	case LolliSpec:
		return LolliRec{
			TermID: id.New(),
			Y:      ConvertSpecToRec(spec.Y),
			Z:      ConvertSpecToRec(spec.Z),
		}
	case WithSpec:
		choices := make(map[sym.ADT]TermRec, len(spec.Choices))
		for lab, st := range spec.Choices {
			choices[lab] = ConvertSpecToRec(st)
		}
		return WithRec{TermID: id.New(), Choices: choices}
	case PlusSpec:
		choices := make(map[sym.ADT]TermRec, len(spec.Choices))
		for lab, st := range spec.Choices {
			choices[lab] = ConvertSpecToRec(st)
		}
		return PlusRec{TermID: id.New(), Choices: choices}
	default:
		panic(ErrSpecTypeUnexpected(spec))
	}
}

func ConvertRecToSpec(r TermRec) TermSpec {
	if r == nil {
		return nil
	}
	switch rec := r.(type) {
	case OneRec:
		return OneSpec{}
	case LinkRec:
		return LinkSpec{TypeQN: rec.TypeQN}
	case TensorRec:
		return TensorSpec{
			B: ConvertRecToSpec(rec.B),
			C: ConvertRecToSpec(rec.C),
		}
	case LolliRec:
		return LolliSpec{
			Y: ConvertRecToSpec(rec.Y),
			Z: ConvertRecToSpec(rec.Z),
		}
	case WithRec:
		choices := make(map[sym.ADT]TermSpec, len(rec.Choices))
		for lab, st := range rec.Choices {
			choices[lab] = ConvertRecToSpec(st)
		}
		return WithSpec{Choices: choices}
	case PlusRec:
		choices := make(map[sym.ADT]TermSpec, len(rec.Choices))
		for lab, st := range rec.Choices {
			choices[lab] = ConvertRecToSpec(st)
		}
		return PlusSpec{Choices: choices}
	default:
		panic(ErrSnapTypeUnexpected(rec))
	}
}

func CheckRef(got, want id.ADT) error {
	if got != want {
		return fmt.Errorf("type mismatch: want %+v, got %+v", want, got)
	}
	return nil
}

// aka eqtp
func CheckSpec(got, want TermSpec) error {
	switch wantSt := want.(type) {
	case OneSpec:
		_, ok := got.(OneSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		return nil
	case TensorSpec:
		gotSt, ok := got.(TensorSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		err := CheckSpec(gotSt.B, wantSt.B)
		if err != nil {
			return err
		}
		return CheckSpec(gotSt.C, wantSt.C)
	case LolliSpec:
		gotSt, ok := got.(LolliSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		err := CheckSpec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckSpec(gotSt.Z, wantSt.Z)
	case PlusSpec:
		gotSt, ok := got.(PlusSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckSpec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	case WithSpec:
		gotSt, ok := got.(WithSpec)
		if !ok {
			return ErrSpecTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckSpec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		panic(ErrSpecTypeUnexpected(want))
	}
}

// aka eqtp
func CheckRec(got, want TermRec) error {
	switch wantSt := want.(type) {
	case OneRec:
		_, ok := got.(OneRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		return nil
	case TensorRec:
		gotSt, ok := got.(TensorRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		err := CheckRec(gotSt.B, wantSt.B)
		if err != nil {
			return err
		}
		return CheckRec(gotSt.C, wantSt.C)
	case LolliRec:
		gotSt, ok := got.(LolliRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		err := CheckRec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckRec(gotSt.Z, wantSt.Z)
	case PlusRec:
		gotSt, ok := got.(PlusRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckRec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	case WithRec:
		gotSt, ok := got.(WithRec)
		if !ok {
			return ErrSnapTypeMismatch(got, want)
		}
		if len(gotSt.Choices) != len(wantSt.Choices) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Choices), len(gotSt.Choices))
		}
		for wantLab, wantChoice := range wantSt.Choices {
			gotChoice, ok := gotSt.Choices[wantLab]
			if !ok {
				return fmt.Errorf("label mismatch: want %q, got nothing", wantLab)
			}
			err := CheckRec(gotChoice, wantChoice)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		panic(ErrSnapTypeUnexpected(want))
	}
}

func ErrSpecTypeUnexpected(got TermSpec) error {
	return fmt.Errorf("spec type unexpected: %T", got)
}

func ErrRefTypeUnexpected(got TermRef) error {
	return fmt.Errorf("ref type unexpected: %T", got)
}

func ErrDoesNotExist(want id.ADT) error {
	return fmt.Errorf("root doesn't exist: %v", want)
}

func ErrMissingInEnv(want id.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func ErrMissingInCfg(want id.ADT) error {
	return fmt.Errorf("root missing in cfg: %v", want)
}

func ErrMissingInCtx(want sym.ADT) error {
	return fmt.Errorf("root missing in ctx: %v", want)
}

func ErrSnapTypeUnexpected(got TermRec) error {
	return fmt.Errorf("root type unexpected: %T", got)
}

func ErrSpecTypeMismatch(got, want TermSpec) error {
	return fmt.Errorf("spec type mismatch: want %T, got %T", want, got)
}

func ErrSnapTypeMismatch(got, want TermRec) error {
	return fmt.Errorf("root type mismatch: want %T, got %T", want, got)
}
