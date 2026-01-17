package typecheck

import (
	"errors"
	"fmt"
	a "orglang/go-runtime/proto/rast2/ast"
)

func Contractive(tp a.Stype) bool {
	switch tp.(type) {
	case a.TpName:
		return false
	default:
		return true
	}
}

func EsyncTp(env a.Environment, tp a.Stype) error {
	return fmt.Errorf("not implemented yet")
}

func CheckExp(env a.Environment, delta a.Context, exp a.Expression, zc a.ChanTp) error {
	switch exp := exp.(type) {
	case a.Fwd:
		if exp.X != zc.X {
			return ErrUnknownVarRight(exp.X)
		}
		sdelta := delta.Shared
		ldelta := delta.Linear
		if a.IsShared(env, zc.A) {
			if len(sdelta) == 0 {
				return errors.New("shared context empty while offered channel is shared")
			}
			if !checkTp(exp.Y, sdelta) {
				return ErrUnknownVarCtx(exp.Y)
			}
			a, err := findStp(exp.Y, delta)
			if err != nil {
				return err
			}
			if eqTp(env, a, zc.A) {
				return nil
			}
			return fmt.Errorf("left type not equal to right type")
		} else {
			if len(ldelta) != 1 {
				return errors.New("linear context must have only one channel")
			}
			ty := ldelta[0].X
			if exp.Y != ty {
				return ErrUnknownVarCtx(exp.Y)
			}
			a := ldelta[0].A
			if eqTp(env, a, zc.A) {
				return nil
			}
			return fmt.Errorf("left type not equal to right type")
		}
	default:
		panic(a.ErrUnexpectedExp)
	}
}

func checkTp(c a.Chan, delta []a.ChanTp) bool {
	for _, d := range delta {
		if d.X.V == c.V {
			return true
		}
	}
	return false
}

func findTp(c a.Chan, delta []a.ChanTp) (a.Stype, error) {
	for _, d := range delta {
		if d.X.V == c.V {
			return d.A, nil
		}
	}
	return nil, ErrUnknownType
}

func findStp(c a.Chan, delta a.Context) (a.Stype, error) {
	if !modS(c) {
		return nil, fmt.Errorf("mode of channel %q not S", c.V)
	}
	return findTp(c, delta.Shared)
}

func modS(ch a.Chan) bool {
	switch ch.M {
	case a.Shared, a.Unknown:
		return true
	default:
		return false
	}
}

func eqTp(env a.Environment, tp1, tp2 a.Stype) bool {
	switch tp1 := tp1.(type) {
	case a.Plus:
		tp2, ok := tp2.(a.Plus)
		if !ok {
			return false
		}
		return eqChoice(env, tp1.Choices, tp2.Choices)
	case a.With:
		tp2, ok := tp2.(a.With)
		if !ok {
			return false
		}
		return eqChoice(env, tp1.Choices, tp2.Choices)
	case a.Tensor:
		tp2, ok := tp2.(a.Tensor)
		if !ok {
			return false
		}
		return eqTp(env, tp1.S, tp2.S) && eqTp(env, tp1.T, tp2.T)
	case a.Lolli:
		tp2, ok := tp2.(a.Lolli)
		if !ok {
			return false
		}
		return eqTp(env, tp1.S, tp2.S) && eqTp(env, tp1.T, tp2.T)
	case a.One:
		_, ok := tp2.(a.One)
		return ok
	case a.Up:
		tp2, ok := tp2.(a.Up)
		if !ok {
			return false
		}
		return eqTp(env, tp1.A, tp2.A)
	case a.Down:
		tp2, ok := tp2.(a.Down)
		if !ok {
			return false
		}
		return eqTp(env, tp1.A, tp2.A)
	case a.TpName:
		tp2, ok := tp2.(a.TpName)
		if !ok {
			tp1, err := a.ExpdTp(env, tp1.A)
			if err != nil {
				panic(err)
			}
			return eqTp(env, tp1, tp2)
		}
		return eqNameName(env, tp1.A, tp2.A)
	default:
		tbd, ok := tp2.(a.TpName)
		if !ok {
			return false
		}
		tp2, err := a.ExpdTp(env, tbd.A)
		if err != nil {
			panic(err)
		}
		return eqTp(env, tp1, tp2)
	}
}

func eqChoice(env a.Environment, cs1, cs2 a.Choices) bool {
	return false
}

func eqNameName(env a.Environment, a1, a2 a.Tpname) bool {
	return false
}

var (
	ErrUnknownType = errors.New("unknown type")
)

func ErrUnknownVarRight(ch a.Chan) error {
	return fmt.Errorf("unbound variable %q on the right", ch.V)
}

func ErrUnknownVarCtx(ch a.Chan) error {
	return fmt.Errorf("unbound variable %q in the context", ch.V)
}
