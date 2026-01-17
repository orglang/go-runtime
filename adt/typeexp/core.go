package typeexp

import (
	"fmt"

	"orglang/go-runtime/adt/identity"
	"orglang/go-runtime/adt/polarity"
	"orglang/go-runtime/adt/revnum"
	"orglang/go-runtime/adt/uniqsym"
)

type ExpSpec interface {
	spec()
}

type OneSpec struct{}

func (OneSpec) spec() {}

// aka TpName
type LinkSpec struct {
	TypeQN uniqsym.ADT
}

func (LinkSpec) spec() {}

type TensorSpec struct {
	Y ExpSpec // val to send
	Z ExpSpec // cont
}

func (TensorSpec) spec() {}

type LolliSpec struct {
	Y ExpSpec // val to receive
	Z ExpSpec // cont
}

func (LolliSpec) spec() {}

// aka Internal Choice
type PlusSpec struct {
	Zs map[uniqsym.ADT]ExpSpec // conts
}

func (PlusSpec) spec() {}

// aka External Choice
type WithSpec struct {
	Zs map[uniqsym.ADT]ExpSpec // conts
}

func (WithSpec) spec() {}

type UpSpec struct {
	Z ExpSpec // cont
}

func (UpSpec) spec() {}

type DownSpec struct {
	Z ExpSpec // cont
}

func (DownSpec) spec() {}

type XactSpec struct {
	Zs map[uniqsym.ADT]ExpSpec // conts
}

func (XactSpec) spec() {}

type ExpRef interface {
	identity.Identifiable
}

type OneRef struct {
	ExpID identity.ADT
}

func (r OneRef) Ident() identity.ADT { return r.ExpID }

type LinkRef struct {
	ExpID identity.ADT
}

func (r LinkRef) Ident() identity.ADT { return r.ExpID }

type PlusRef struct {
	ExpID identity.ADT
}

func (r PlusRef) Ident() identity.ADT { return r.ExpID }

type WithRef struct {
	ExpID identity.ADT
}

func (r WithRef) Ident() identity.ADT { return r.ExpID }

type TensorRef struct {
	ExpID identity.ADT
}

func (r TensorRef) Ident() identity.ADT { return r.ExpID }

type LolliRef struct {
	ExpID identity.ADT
}

func (r LolliRef) Ident() identity.ADT { return r.ExpID }

type UpRef struct {
	ExpID identity.ADT
}

func (r UpRef) Ident() identity.ADT { return r.ExpID }

type DownRef struct {
	ExpID identity.ADT
}

func (r DownRef) Ident() identity.ADT { return r.ExpID }

// aka Stype
type ExpRec interface {
	identity.Identifiable
	polarity.Polarizable
}

type ProdRec interface {
	Next() identity.ADT
}

type SumRec interface {
	Next(uniqsym.ADT) identity.ADT
}

type OneRec struct {
	ExpID identity.ADT
}

func (OneRec) spec() {}

func (r OneRec) Ident() identity.ADT { return r.ExpID }

func (OneRec) Pol() polarity.ADT { return polarity.Pos }

// aka TpName
type LinkRec struct {
	ExpID  identity.ADT
	TypeQN uniqsym.ADT
}

func (LinkRec) spec() {}

func (r LinkRec) Ident() identity.ADT { return r.ExpID }

func (LinkRec) Pol() polarity.ADT { return polarity.Zero }

// aka Internal Choice
type PlusRec struct {
	ExpID identity.ADT
	Zs    map[uniqsym.ADT]ExpRec
}

func (PlusRec) spec() {}

func (r PlusRec) Ident() identity.ADT { return r.ExpID }

func (r PlusRec) Next(l uniqsym.ADT) identity.ADT { return r.Zs[l].Ident() }

func (PlusRec) Pol() polarity.ADT { return polarity.Pos }

// aka External Choice
type WithRec struct {
	ExpID identity.ADT
	Zs    map[uniqsym.ADT]ExpRec
}

func (WithRec) spec() {}

func (r WithRec) Ident() identity.ADT { return r.ExpID }

func (r WithRec) Next(l uniqsym.ADT) identity.ADT { return r.Zs[l].Ident() }

func (WithRec) Pol() polarity.ADT { return polarity.Neg }

type TensorRec struct {
	ExpID identity.ADT
	Y     ExpRec
	Z     ExpRec
}

func (TensorRec) spec() {}

func (r TensorRec) Ident() identity.ADT { return r.ExpID }

func (r TensorRec) Next() identity.ADT { return r.Z.Ident() }

func (TensorRec) Pol() polarity.ADT { return polarity.Pos }

type LolliRec struct {
	ExpID identity.ADT
	Y     ExpRec
	Z     ExpRec
}

func (LolliRec) spec() {}

func (r LolliRec) Ident() identity.ADT { return r.ExpID }

func (r LolliRec) Next() identity.ADT { return r.Z.Ident() }

func (LolliRec) Pol() polarity.ADT { return polarity.Neg }

type UpRec struct {
	ExpID identity.ADT
	Z     ExpRec
}

func (UpRec) spec() {}

func (r UpRec) Ident() identity.ADT { return r.ExpID }

func (UpRec) Pol() polarity.ADT { return polarity.Zero }

type DownRec struct {
	ExpID identity.ADT
	Z     ExpRec
}

func (DownRec) spec() {}

func (r DownRec) Ident() identity.ADT { return r.ExpID }

func (DownRec) Pol() polarity.ADT { return polarity.Zero }

type Context struct {
	Assets map[uniqsym.ADT]ExpRec
	Liabs  map[uniqsym.ADT]ExpRec
}

func ErrSymMissingInEnv(want uniqsym.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func errConcurrentModification(got revnum.ADT, want revnum.ADT) error {
	return fmt.Errorf("entity concurrent modification: want revision %v, got revision %v", want, got)
}

func errOptimisticUpdate(got revnum.ADT) error {
	return fmt.Errorf("entity concurrent modification: got revision %v", got)
}

func CheckRef(got, want identity.ADT) error {
	if got != want {
		return fmt.Errorf("type mismatch: want %+v, got %+v", want, got)
	}
	return nil
}

// aka eqtp
func CheckSpec(got, want ExpSpec) error {
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
		err := CheckSpec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckSpec(gotSt.Z, wantSt.Z)
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
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
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
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
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
func CheckRec(got, want ExpRec) error {
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
		err := CheckRec(gotSt.Y, wantSt.Y)
		if err != nil {
			return err
		}
		return CheckRec(gotSt.Z, wantSt.Z)
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
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
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
		if len(gotSt.Zs) != len(wantSt.Zs) {
			return fmt.Errorf("choices mismatch: want %v items, got %v items", len(wantSt.Zs), len(gotSt.Zs))
		}
		for wantLab, wantChoice := range wantSt.Zs {
			gotChoice, ok := gotSt.Zs[wantLab]
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
		panic(ErrRecTypeUnexpected(want))
	}
}

func ErrSpecTypeUnexpected(got ExpSpec) error {
	return fmt.Errorf("spec type unexpected: %T", got)
}

func ErrRefTypeUnexpected(got ExpRef) error {
	return fmt.Errorf("ref type unexpected: %T", got)
}

func ErrDoesNotExist(want identity.ADT) error {
	return fmt.Errorf("root doesn't exist: %v", want)
}

func ErrMissingInEnv(want identity.ADT) error {
	return fmt.Errorf("root missing in env: %v", want)
}

func ErrMissingInCfg(want identity.ADT) error {
	return fmt.Errorf("root missing in cfg: %v", want)
}

func ErrMissingInCtx(want uniqsym.ADT) error {
	return fmt.Errorf("root missing in ctx: %v", want)
}

func ErrRecTypeUnexpected(got ExpRec) error {
	return fmt.Errorf("rec type unexpected: %T", got)
}

func ErrSpecTypeMismatch(got, want ExpSpec) error {
	return fmt.Errorf("spec type mismatch: want %T, got %T", want, got)
}

func ErrSnapTypeMismatch(got, want ExpRec) error {
	return fmt.Errorf("root type mismatch: want %T, got %T", want, got)
}

func ErrPolarityUnexpected(got ExpRec) error {
	return fmt.Errorf("root polarity unexpected: %v", got.Pol())
}

func ErrPolarityMismatch(a, b ExpRec) error {
	return fmt.Errorf("root polarity mismatch: %v != %v", a.Pol(), b.Pol())
}
