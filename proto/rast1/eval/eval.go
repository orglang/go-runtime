package eval

import (
	"errors"
	a "orglang/go-runtime/proto/rast1/ast"
)

type Value interface {
	value()
}

func (Lab) value()     {}
func (Send) value()    {}
func (Close) value()   {}
func (CloCase) value() {}
func (CloRecv) value() {}

type Lab struct {
	K a.Label
	V Value
}

type Send struct {
	W Value
	V Value
}

type Close struct{}

type CloCase struct {
	Eta      Eta
	Branches a.Branches
	Z        a.Channel
}

type CloRecv struct {
	Eta     Eta
	ChanExp ChanExp
	Z       a.Channel
}

type ChanExp struct {
	X a.Channel
	P a.Exp
}

type Eta map[a.Channel]Value

func Evaluate(env *a.Env, P a.Exp, z a.Channel) (Value, error) {
	return eval(env, make(Eta), P, z)
}

func eval(env *a.Env, eta Eta, P a.Exp, z a.Channel) (Value, error) {
	switch exp := P.(type) {
	case a.Id:
		return eta[exp.Y], nil
	case a.Spawn:
		eta2, err := evalCall(env, eta, exp.P)
		if err != nil {
			return nil, err
		}
		return eval(env, eta2, exp.Q, z)
	case a.Lab:
		if exp.X == z {
			v, err := eval(env, eta, exp.P, z)
			if err != nil {
				return nil, err
			}
			return Lab{exp.K, v}, nil
		}
		clo := eta[exp.X].(CloCase)
		v, err := eval(env, clo.Eta, clo.Branches[exp.K], z)
		if err != nil {
			return nil, err
		}
		eta[exp.X] = v
		return eval(env, eta, exp.P, z)
	case a.Case:
		if exp.X == z {
			return CloCase{eta, exp.Branches, z}, nil
		}
		lab := eta[exp.X].(Lab)
		eta[exp.X] = lab.V
		return eval(env, eta, exp.Branches[lab.K], z)
	case a.Send:
		w := eta[exp.Y]
		if exp.X == z {
			delete(eta, exp.Y)
			v, err := eval(env, eta, exp.P, z)
			if err != nil {
				return nil, err
			}
			return Send{w, v}, nil
		}
		clo := eta[exp.X].(CloRecv)
		clo.Eta[clo.ChanExp.X] = w
		v, err := eval(env, clo.Eta, clo.ChanExp.P, z)
		if err != nil {
			return nil, err
		}
		eta[exp.X] = v
		delete(eta, exp.Y)
		return eval(env, eta, exp.P, z)
	case a.Recv:
		if exp.X == z {
			return CloRecv{eta, ChanExp{exp.Y, exp.P}, z}, nil
		}
		send := eta[exp.X].(Send)
		eta[exp.X] = send.V
		eta[exp.Y] = send.W
		return eval(env, eta, exp.P, z)
	case a.Close:
		return Close{}, nil
	case a.Wait:
		_, ok := eta[exp.X].(Close)
		if !ok {
			return nil, ErrUnexpectedWait
		}
		delete(eta, exp.X)
		return eval(env, eta, exp.P, z)
	case nil:
		return nil, nil
	default:
		panic(ErrUnexpectedExp)
	}
}

var (
	ErrUnexpectedExp  = errors.New("unexpected exp")
	ErrUnexpectedWait = errors.New("unexpected wait")
)

// Simplified version which based on ExpDec instead of ExpDef
func evalCall(env *a.Env, eta1 Eta, P a.Exp) (Eta, error) {
	switch exp := P.(type) {
	case a.ExpName:
		expDec := env.ExpDecs[exp.F]
		ys := make([]a.Channel, len(expDec.Delta))
		for i, y := range expDec.Delta {
			ys[i] = y.Ch
		}
		if len(exp.Ys) != len(ys) {
			return nil, errors.New("lenghts should be equal")
		}
		eta2 := make(Eta, len(ys))
		for i, y := range exp.Ys {
			eta2[ys[i]] = eta1[y]
			delete(eta1, y)
		}
		v, err := eval(env, eta1, P, expDec.Zc.Ch)
		if err != nil {
			return nil, err
		}
		eta2[exp.X] = v
		return eta2, nil
	default:
		return nil, ErrUnexpectedExp
	}
}
