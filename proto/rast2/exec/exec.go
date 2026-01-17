package exec

import (
	"errors"
	"fmt"
	"maps"
	a "orglang/go-runtime/proto/rast2/ast"
)

type Proc struct {
	C a.Chan
	P a.Expression
}

type Msg struct {
	D a.Chan
	M a.Msg
}

type Sem interface {
	sem()
}

func (Proc) sem() {}
func (Msg) sem()  {}

type Configuration struct {
	Conf   map[a.Chan]Sem
	Conts  map[a.Chan]a.Chan
	Shared map[a.Chan]a.Chan
}

type Pol string

const (
	Pos Pol = "pos"
	Neg Pol = "neg"
)

var chanNum uint32 = 0

func lfresh() a.Chan {
	n := chanNum
	chanNum += 1
	return a.Chan{V: fmt.Sprintf("ch%d", n)}
}

func createConfig(sem Sem) *Configuration {
	conf := make(map[a.Chan]Sem)
	switch s := sem.(type) {
	case Proc:
		conf[s.C] = sem
	case Msg:
		conf[s.D] = sem
	default:
		panic(ErrUnexpectedSem)
	}
	return &Configuration{conf, make(map[a.Chan]a.Chan), make(map[a.Chan]a.Chan)}
}

func step(env a.Environment, config *Configuration) (*Configuration, error) {
	stepped := true
	for stepped {
		stepped = false
		for _, sem := range maps.Clone(config.Conf) {
			changed, err := matchAndOneStep(env, sem, config)
			if err != nil {
				return nil, err
			}
			if changed {
				stepped = true
			}
		}
	}
	return config, nil
}

func matchAndOneStep(env a.Environment, sem Sem, config *Configuration) (bool, error) {
	switch s := sem.(type) {
	case Proc:
		switch s.P.(type) {
		case a.Fwd:
			return fwd(s.C, config)
		case a.Spawn:
			return spawn(env, s.C, config)
		case a.ExpName:
			return expand(env, s.C, config)
		case a.Close:
			return oneS(s.C, config)
		case a.Wait:
			return oneR(s.C, config)
		default:
			panic(a.ErrUnexpectedExp)
		}
	case Msg:
		return false, nil
	default:
		panic(ErrUnexpectedSem)
	}
}

func fwd(ch a.Chan, config *Configuration) (bool, error) {
	s, ok := config.Conf[ch]
	if !ok {
		return false, ErrExecImpossible
	}
	switch sem := s.(type) {
	case Proc:
		fwd, ok := sem.P.(a.Fwd)
		if !ok {
			return false, ErrExecImpossible
		}
		if sem.C != fwd.X {
			return false, ErrChannelMismatch
		}
		// try to apply fwd+ rule
		msgp, err := findMsg(fwd.Y, config, Pos)
		if err != nil {
			return false, err
		}
		if msgp == nil {
			// try to apply fwd- rule
			msgn, err := findMsg(sem.C, config, Neg)
			if err != nil {
				return false, err
			}
			switch msg := msgn.(type) {
			case Msg:
				switch m := msg.M.(type) {
				case a.MLabI, a.MSendL, a.MClose:
					delete(config.Conf, sem.C)
					e, ok := config.Conts[sem.C]
					if !ok {
						return false, ErrExecImpossible
					}
					delete(config.Conf, e)
					msg := Msg{e, a.Msubst(fwd.Y, sem.C, m)}
					config.Conf[e] = msg
					delete(config.Conts, sem.C)
					addCont(fwd.Y, e, config)
					return true, nil
				default:
					return false, nil
				}
			default:
				return false, nil
			}
		}
		switch msg := msgp.(type) {
		case Msg:
			switch m := msg.M.(type) {
			case a.MLabI, a.MSendL:
				delete(config.Conf, sem.C)
				delete(config.Conf, fwd.Y)
				msg := Msg{sem.C, a.Msubst(sem.C, fwd.Y, m)}
				config.Conf[sem.C] = msg
				d, ok := config.Conts[fwd.Y]
				if !ok {
					return false, ErrExecImpossible
				}
				delete(config.Conts, fwd.Y)
				addCont(sem.C, d, config)
				return true, nil
			default:
				return false, nil
			}
		default:
			return false, nil
		}
	default:
		return false, ErrExecImpossible
	}
}

func spawn(_ a.Environment, ch a.Chan, config *Configuration) (bool, error) {
	s, ok := config.Conf[ch]
	if !ok {
		return false, ErrExecImpossible
	}
	delete(config.Conf, ch)
	switch sem := s.(type) {
	case Proc:
		spawn, ok := sem.P.(a.Spawn)
		if !ok {
			return false, ErrExecImpossible
		}
		c := lfresh()
		proc1 := Proc{c, a.ExpName{X: c, F: spawn.F, Xs: spawn.Xs}}
		proc2 := Proc{sem.C, a.Subst(c, spawn.X, spawn.Q)}
		config.Conf[c] = proc1
		config.Conf[sem.C] = proc2
		return true, nil
	default:
		return false, ErrExecImpossible
	}
}

func expand(env a.Environment, ch a.Chan, config *Configuration) (bool, error) {
	s, ok := config.Conf[ch]
	if !ok {
		return false, ErrExecImpossible
	}
	delete(config.Conf, ch)
	switch sem := s.(type) {
	case Proc:
		en, ok := sem.P.(a.ExpName)
		if !ok {
			return false, ErrExecImpossible
		}
		p, err := expdDef(env, en.X, en.F, en.Xs)
		if err != nil {
			return false, err
		}
		proc := Proc{sem.C, a.Subst(sem.C, en.X, p)}
		config.Conf[sem.C] = proc
		return true, nil
	default:
		return false, ErrExecImpossible
	}
}

func oneS(ch a.Chan, config *Configuration) (bool, error) {
	s, ok := config.Conf[ch]
	if !ok {
		return false, ErrExecImpossible
	}
	delete(config.Conf, ch)
	switch sem := s.(type) {
	case Proc:
		close, ok := sem.P.(a.Close)
		if !ok {
			return false, ErrExecImpossible
		}
		if sem.C != close.X {
			return false, ErrChannelMismatch
		}
		msg := Msg{sem.C, a.MClose{X: sem.C}}
		config.Conf[sem.C] = msg
		return true, nil
	default:
		return false, ErrExecImpossible
	}
}

func oneR(ch a.Chan, config *Configuration) (bool, error) {
	s, ok := config.Conf[ch]
	if !ok {
		return false, ErrExecImpossible
	}
	switch sem := s.(type) {
	case Proc:
		wait, ok := sem.P.(a.Wait)
		if !ok {
			return false, ErrExecImpossible
		}
		if sem.C == wait.X {
			return false, ErrChannelMismatch
		}
		m, err := findMsg(wait.X, config, Pos)
		if err != nil {
			return false, err
		}
		switch msg := m.(type) {
		case Msg:
			if msg.D != wait.X {
				return false, ErrChannelMismatch
			}
			_, ok := msg.M.(a.MClose)
			if !ok {
				return false, nil
			}
			proc := Proc{sem.C, wait.P}
			delete(config.Conf, ch)
			delete(config.Conf, wait.X)
			config.Conf[sem.C] = proc
			return true, nil
		default:
			return false, nil
		}
	default:
		return false, ErrExecImpossible
	}
}

func expdDef(env a.Environment, x a.Chan, f a.Expname, xs []a.Chan) (a.Expression, error) {
	decl, ok := env[f]
	if !ok {
		return nil, ErrUndefinedProcess
	}
	defdec, ok := decl.(a.ExpDecDef)
	if !ok {
		return nil, ErrExecImpossible
	}
	exp := a.Subst(x, defdec.Zc.X, defdec.P)
	return a.SubstCtx(xs, fst(defdec.Ctx.Ordered), exp), nil
}

func fst(as []a.ChanTp) []a.Chan {
	bs := make([]a.Chan, len(as))
	for _, a := range as {
		bs = append(bs, a.X)
	}
	return bs
}

func addCont(c1 a.Chan, c2 a.Chan, config *Configuration) {
	cs, ok := config.Shared[c1]
	if ok {
		delete(config.Shared, c1)
		config.Shared[c2] = cs
	}
	config.Conts[c1] = c2
}

func findMsg(c1 a.Chan, config *Configuration, dual Pol) (Sem, error) {
	switch dual {
	case Neg:
		c2, ok := config.Conts[c1]
		if !ok {
			return nil, nil
		}
		return config.Conf[c2], nil
	case Pos:
		return config.Conf[c1], nil
	default:
		return nil, ErrExecImpossible
	}
}

func Exec(env a.Environment, f a.Expname) (*Configuration, error) {
	c := lfresh()
	sem := Proc{c, a.ExpName{X: c, F: f, Xs: []a.Chan{}}}
	config, err := step(env, createConfig(sem))
	if err != nil {
		return nil, err
	}
	return config, nil
}

var (
	ErrExecImpossible   = errors.New("exec impossible")
	ErrChannelMismatch  = errors.New("channel mismatch")
	ErrUndefinedProcess = errors.New("udefined process")
	ErrUnexpectedSem    = errors.New("unexpected sem")
)
