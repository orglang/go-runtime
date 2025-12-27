package exec

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	a "orglang/orglang/proto/rast2/ast"
)

func setupSubtest() {
	chanNum = 0
}

func TestExecSucc(t *testing.T) {
	f := func(env a.Environment, f a.Expname, expected *Configuration) {
		t.Helper()
		actual, err := Exec(env, f)
		if err != nil {
			t.Fatalf("unexpected error %q", err)
		}
		opt := cmpopts.EquateEmpty()
		if diff := cmp.Diff(expected, actual, opt); diff != "" {
			t.Errorf("unexpected config (-want +got):\n%s", diff)
		}
	}

	t.Run("close", func(t *testing.T) {
		setupSubtest()

		z := a.Chan{V: "z"}
		main := "main"
		env := map[string]a.Decl{
			main: a.ExpDecDef{
				F: main,
				Zc: a.ChanTp{
					X: z,
					A: a.One{},
				},
				P: a.Close{X: z},
			},
		}
		ch0 := a.Chan{V: "ch0"}
		config := &Configuration{
			Conf: map[a.Chan]Sem{
				ch0: Msg{D: ch0, M: a.MClose{X: ch0}},
			},
		}
		f(env, main, config)
	})

	t.Run("spawn", func(t *testing.T) {
		setupSubtest()

		foo := "foo"
		z := a.Chan{V: "z"}
		main := "main"
		d := a.Chan{V: "d"}
		env := map[string]a.Decl{
			foo: a.ExpDecDef{
				F: foo,
				Zc: a.ChanTp{
					X: z,
					A: a.One{},
				},
				P: a.Close{X: z},
			},
			main: a.ExpDecDef{
				F: main,
				Zc: a.ChanTp{
					X: d,
					A: a.One{},
				},
				P: a.Spawn{
					X: z,
					F: foo,
					Q: a.Wait{X: z, P: a.Close{X: d}},
				},
			},
		}
		ch0 := a.Chan{V: "ch0"}
		config := &Configuration{
			Conf: map[a.Chan]Sem{
				ch0: Msg{D: ch0, M: a.MClose{X: ch0}},
			},
		}
		f(env, main, config)
	})
}

func TestExecErr(t *testing.T) {
	f := func(env a.Environment, f a.Expname, expected error) {
		t.Helper()
		_, actual := Exec(env, f)
		if actual == nil {
			t.Fatalf("expecting error %q", expected)
		}
		if !errors.Is(actual, expected) {
			t.Errorf("unexpected error; got %q; want %q", actual, expected)
		}
	}

	f(make(a.Environment), "foo", ErrUndefinedProcess)
}
