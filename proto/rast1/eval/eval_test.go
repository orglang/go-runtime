package eval

import (
	"errors"
	"testing"

	a "orglang/orglang/proto/rast1/ast"

	"github.com/google/go-cmp/cmp"
)

const (
	// channels
	x a.Channel = "x"
	y a.Channel = "y"
	z a.Channel = "z"

	// labels
	k a.Label = "k"
)

var (
	// exps
	closeX = a.Close{x}

	// values
	closeV = Close{}
)

func TestEvalSucc(t *testing.T) {
	t.Parallel()
	tcs := map[string]struct {
		eta      Eta
		exp      a.Exp
		z        a.Channel
		expected Value
	}{
		"Id1":    {Eta{y: closeV}, a.Id{x, y}, x, closeV},
		"Id2":    {Eta{y: closeV}, a.Id{x, y}, z, closeV},
		"Lab1":   {Eta{}, a.Lab{x, k, nil}, x, Lab{k, nil}},
		"Lab2":   {Eta{x: CloCase{Eta{}, a.Branches{k: nil}, z}}, a.Lab{x, k, nil}, z, nil},
		"Case1":  {Eta{}, a.Case{x, a.Branches{k: nil}}, x, CloCase{Eta{}, a.Branches{k: nil}, x}},
		"Case2":  {Eta{x: Lab{k, nil}}, a.Case{x, a.Branches{k: nil}}, z, nil},
		"Send1":  {Eta{y: nil}, a.Send{x, y, nil}, x, Send{nil, nil}},
		"Send2":  {Eta{x: CloRecv{Eta{}, ChanExp{x, nil}, z}}, a.Send{x, y, nil}, z, nil},
		"Recv1":  {Eta{}, a.Recv{x, y, nil}, x, CloRecv{Eta{}, ChanExp{y, nil}, x}},
		"Recv2":  {Eta{x: Send{nil, nil}}, a.Recv{x, y, nil}, z, nil},
		"Close1": {Eta{}, closeX, x, closeV},
		"Close2": {Eta{}, closeX, z, closeV},
		"Wait1":  {Eta{x: closeV}, a.Wait{x, nil}, x, nil},
		"Wait2":  {Eta{x: closeV}, a.Wait{x, nil}, z, nil},
	}
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			actual, err := eval(&a.Env{}, tc.eta, tc.exp, tc.z)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(actual, tc.expected) {
				t.Errorf("\nactual: %T,\nexpected: %T,\ndiff: %s\n",
					actual, tc.expected, cmp.Diff(actual, tc.expected))
			}
		})
	}
}

func TestEvalErr(t *testing.T) {
	t.Parallel()
	tcs := map[string]struct {
		eta      Eta
		exp      a.Exp
		expected error
	}{
		"Wait": {Eta{}, a.Wait{x, nil}, ErrUnexpectedWait},
	}
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			_, actual := eval(&a.Env{}, tc.eta, tc.exp, z)
			if !errors.Is(actual, tc.expected) {
				t.Errorf("\nactual: %#v,\nexpected: %#v",
					actual, tc.expected)
			}
		})
	}
}
