package pool

import (
	"database/sql"
	"fmt"
	"slices"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"

	"orglang/go-runtime/lib/e2e"

	"github.com/orglang/go-sdk/lib/rc"

	"github.com/orglang/go-sdk/adt/poolbind"
	"github.com/orglang/go-sdk/adt/pooldec"
	"github.com/orglang/go-sdk/adt/poolexec"
	"github.com/orglang/go-sdk/adt/poolexp"
	"github.com/orglang/go-sdk/adt/poolstep"
	"github.com/orglang/go-sdk/adt/procbind"
	"github.com/orglang/go-sdk/adt/procdec"
	"github.com/orglang/go-sdk/adt/procexp"
	"github.com/orglang/go-sdk/adt/procstep"
	"github.com/orglang/go-sdk/adt/typedef"
	"github.com/orglang/go-sdk/adt/typeexp"
	"github.com/orglang/go-sdk/adt/xactdef"
	"github.com/orglang/go-sdk/adt/xactexp"
)

func TestPool(t *testing.T) {
	s := suite{}
	s.beforeAll(t)
	t.Run("CreateRetreive", s.createRetreive)
	t.Run("WaitClose", s.waitClose)
	t.Run("RecvSend", s.recvSend)
	t.Run("CaseLab", s.caseLab)
	t.Run("Call", s.call)
	t.Run("Fwd", s.fwd)

}

type suite struct {
	poolDecAPI  e2e.PoolDecAPI
	poolExecAPI e2e.PoolExecAPI
	xactDefAPI  e2e.XactDefAPI
	procDecAPI  e2e.ProcDecAPI
	procExecAPI e2e.ProcExecAPI
	typeDefAPI  e2e.TypeDefAPI
	db          *sql.DB
}

func (s *suite) beforeAll(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://orglang:orglang@localhost:5432/orglang")
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() { db.Close() })
	s.db = db
	app := fx.New(rc.Module, e2e.Module,
		fx.Populate(
			s.poolDecAPI,
			s.poolExecAPI,
			s.xactDefAPI,
			s.procDecAPI,
			s.procExecAPI,
			s.typeDefAPI,
		))
	t.Cleanup(func() { app.Stop(t.Context()) })
}

func (s *suite) beforeEach(t *testing.T) {
	tables := []string{
		"syn_decs",
		"pool_decs", "pool_execs", "pool_liabs", "pool_steps",
		"proc_decs", "proc_execs", "proc_binds", "proc_steps",
		"dec_pes", "dec_ces",
		"type_defs", "type_exps",
	}
	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("truncate table %v", table))
		if err != nil {
			t.Fatal(err)
		}
	}
}

func (s *suite) createRetreive(t *testing.T) {
	// given
	poolSpec1 := poolexec.ExecSpec{PoolQN: "pool-1"}
	poolRef1, err := s.poolExecAPI.Create(poolSpec1)
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolSpec2 := poolexec.ExecSpec{PoolQN: "pool-2", SupID: poolRef1.ID}
	poolRef2, err := s.poolExecAPI.Create(poolSpec2)
	if err != nil {
		t.Fatal(err)
	}
	// when
	poolSnap1, err := s.poolExecAPI.Retrieve(poolRef1)
	if err != nil {
		t.Fatal(err)
	}
	// then
	if !slices.Contains(poolSnap1.SubExecs, poolRef2) {
		t.Errorf("unexpected subs in %q; want: %+v, got: %+v",
			poolSpec1.PoolQN, poolRef2, poolSnap1.SubExecs)
	}
}

func (s *suite) waitClose(t *testing.T) {
	s.beforeEach(t)
	// given
	closerProcQN := "closer-proc-qn"
	waiterProcQN := "waiter-proc-qn"
	// and
	withXactQN := "with-xact-qn"
	_, err := s.xactDefAPI.Create(xactdef.DefSpec{
		XactQN: withXactQN,
		XactES: xactexp.ExpSpec{
			K: xactexp.With,
			With: &xactexp.LaborSpec{
				Choices: []xactexp.ChoiceSpec{
					// пул заявляет способность работать closerProcQN
					{ProcQN: closerProcQN, ContES: xactexp.ExpSpec{
						K:    xactexp.Link,
						Link: &xactexp.LinkSpec{XactQN: withXactQN},
					}},
					// пул заявляет способность работать waiterProcQN
					{ProcQN: waiterProcQN, ContES: xactexp.ExpSpec{
						K:    xactexp.Link,
						Link: &xactexp.LinkSpec{XactQN: withXactQN},
					}},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	myPoolQN := "my-pool-qn"
	poolClientPH := "pool-client-ph"
	poolProviderPH := "pool-provider-ph"
	_, err = s.poolDecAPI.Create(pooldec.DecSpec{
		PoolQN: myPoolQN,
		ProviderBS: poolbind.BindSpec{
			ChnlPH: poolProviderPH,
			ChnlBM: poolbind.Self,
			XactQN: withXactQN,
		},
		ClientBS: poolbind.BindSpec{
			ChnlPH: poolClientPH,
			ChnlBM: poolbind.Self,
			XactQN: withXactQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	myPoolExec, err := s.poolExecAPI.Create(poolexec.ExecSpec{
		PoolQN: myPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolCloserPH := "pool-closer-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire, // пул делает предложение поработать в качестве closerProcQN
			Hire: &poolexp.HireSpec{
				CommPH: poolClientPH, // пул выступает в качестве клиента самого себя
				ProcQN: closerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Apply, // пул принимает предложение поработать в качестве closerProcQN
			Apply: &poolexp.ApplySpec{
				CommPH: poolProviderPH, // пул выступает в качестве провайдера самому себе
				ProcQN: closerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolCloserPH,
				ProcQN: closerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolWaiterPH := "pool-waiter-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				CommPH: poolClientPH,
				ProcQN: waiterProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Apply,
			Apply: &poolexp.ApplySpec{
				CommPH: poolProviderPH,
				ProcQN: waiterProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolWaiterPH,
				ProcQN: waiterProcQN,
				ValPHs: []string{poolCloserPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneTypeQN := "one-type-qn"
	_, err = s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: oneTypeQN,
		TypeES: typeexp.ExpSpec{K: typeexp.One},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: closerProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "closer-provider-ph",
			TypeQN: oneTypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:     waiterProcQN,
		ProviderBS: procbind.BindSpec{ChnlPH: "waiter-provider-ph", TypeQN: oneTypeQN},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "waiter-closer-ph", TypeQN: oneTypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: closerProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Close,
			Close: &procexp.CloseSpec{
				CommPH: closerProcDec.ProviderBS.ChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: waiterProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Wait,
			Wait: &procexp.WaitSpec{
				CommPH: waiterProcDec.ClientBSs[0].ChnlPH,
				ContES: procexp.ExpSpec{
					K: procexp.Close,
					Close: &procexp.CloseSpec{
						CommPH: waiterProcDec.ProviderBS.ChnlPH,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// then
	// TODO добавить проверку
}

func (s *suite) recvSend(t *testing.T) {
	s.beforeEach(t)
	// given
	senderProcQN := "sender-proc-qn"
	messageProcQN := "message-proc-qn"
	receiverProcQN := "receiver-proc-qn"
	// and
	withXactQN := "with-xact-qn"
	_, err := s.xactDefAPI.Create(xactdef.DefSpec{
		XactQN: withXactQN,
		XactES: xactexp.ExpSpec{
			K: xactexp.With,
			With: &xactexp.LaborSpec{
				Choices: []xactexp.ChoiceSpec{
					{ProcQN: senderProcQN, ContES: xactexp.ExpSpec{
						K:    xactexp.Link,
						Link: &xactexp.LinkSpec{XactQN: withXactQN},
					}},
					{ProcQN: receiverProcQN, ContES: xactexp.ExpSpec{
						K:    xactexp.Link,
						Link: &xactexp.LinkSpec{XactQN: withXactQN},
					}},
					{ProcQN: messageProcQN, ContES: xactexp.ExpSpec{
						K:    xactexp.Link,
						Link: &xactexp.LinkSpec{XactQN: withXactQN},
					}},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	myPoolQN := "my-pool-qn"
	poolProviderPH := "pool-provider-ph"
	poolClientPH := "pool-client-ph"
	_, err = s.poolDecAPI.Create(pooldec.DecSpec{
		PoolQN: myPoolQN,
		ProviderBS: poolbind.BindSpec{
			ChnlPH: poolProviderPH,
			ChnlBM: poolbind.Self,
			XactQN: withXactQN,
		},
		ClientBS: poolbind.BindSpec{
			ChnlPH: poolClientPH,
			ChnlBM: poolbind.Self,
			XactQN: withXactQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	myPoolExec, err := s.poolExecAPI.Create(poolexec.ExecSpec{
		PoolQN: myPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	lolliTypeQN := "lolli-type-qn"
	_, err = s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: lolliTypeQN,
		TypeES: typeexp.ExpSpec{
			K: typeexp.Lolli,
			Lolli: &typeexp.ProdSpec{
				ValES:  typeexp.ExpSpec{K: typeexp.One},
				ContES: typeexp.ExpSpec{K: typeexp.One},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneTypeQN := "one-type-qn"
	_, err = s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: oneTypeQN,
		TypeES: typeexp.ExpSpec{K: typeexp.One},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	receiverProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: receiverProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "receiver-provider-ph",
			TypeQN: lolliTypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	_, err = s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: messageProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "message-provider-ph",
			TypeQN: oneTypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	senderProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: senderProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "sender-provider-ph",
			TypeQN: oneTypeQN,
		},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "sender-receiver-ph", TypeQN: lolliTypeQN},
			{ChnlPH: "sender-message-ph", TypeQN: oneTypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolReceiverPH := "pool-receiver-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				CommPH: poolClientPH,
				ProcQN: receiverProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	receiverProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolReceiverPH,
				ProcQN: receiverProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolMessagePH := "pool-message-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				CommPH: poolClientPH,
				ProcQN: messageProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolMessagePH,
				ProcQN: messageProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolSenderPH := "pool-sender-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				CommPH: poolClientPH,
				ProcQN: senderProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	senderProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolSenderPH,
				ProcQN: senderProcQN,
				ValPHs: []string{poolReceiverPH, poolMessagePH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: receiverProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Recv,
			Recv: &procexp.RecvSpec{
				CommPH: receiverProcDec.ProviderBS.ChnlPH,
				BindPH: "receiver-message-ph",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: senderProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Send,
			Send: &procexp.SendSpec{
				CommPH: senderProcDec.ClientBSs[0].ChnlPH,
				ValPH:  senderProcDec.ClientBSs[1].ChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// then
	// TODO добавить проверку
}

func (s *suite) caseLab(t *testing.T) {
	s.beforeEach(t)
	// given
	myPoolQN := "my-pool-qn"
	myPoolExec, err := s.poolExecAPI.Create(poolexec.ExecSpec{
		PoolQN: myPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	labelQN := "label-qn"
	// and
	withType, err := s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: "with-type-qn",
		TypeES: typeexp.ExpSpec{
			With: &typeexp.SumSpec{
				Choices: []typeexp.ChoiceSpec{
					{LabQN: labelQN, ContES: typeexp.ExpSpec{K: typeexp.One}},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneType, err := s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: "one-type-qn",
		TypeES: typeexp.ExpSpec{K: typeexp.One},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	followerProcQN := "follower-proc-qn"
	followerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: followerProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "follower-provider-ph",
			TypeQN: withType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	deciderProcQN := "decider-proc-qn"
	_, err = s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:    deciderProcQN,
		ClientBSs: []procbind.BindSpec{followerProcDec.ProviderBS},
		ProviderBS: procbind.BindSpec{
			ChnlPH: "decider-provider-ph",
			TypeQN: oneType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolFollowerPH := "pool-follower-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				ProcQN: followerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	followerProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolFollowerPH,
				ProcQN: followerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolDeciderPH := "pool-decider-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				ProcQN: deciderProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	deciderProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolDeciderPH,
				ProcQN: deciderProcQN,
				ValPHs: []string{poolFollowerPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: followerProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Case,
			Case: &procexp.CaseSpec{
				CommPH: poolFollowerPH,
				ContBSs: []procexp.BranchSpec{
					{LabQN: labelQN, ContES: procexp.ExpSpec{
						K: procexp.Close,
						Close: &procexp.CloseSpec{
							CommPH: poolFollowerPH,
						},
					},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: deciderProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Lab,
			Lab: &procexp.LabSpec{
				CommPH: poolFollowerPH,
				InfoQN: labelQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// then
	// TODO добавить проверку
}

func (s *suite) call(t *testing.T) {
	s.beforeEach(t)
	// given
	myPoolQN := "my-pool-qn"
	myPoolExec, err := s.poolExecAPI.Create(poolexec.ExecSpec{
		PoolQN: myPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneType, err := s.typeDefAPI.Create(
		typedef.DefSpec{
			TypeQN: "one-type-qn",
			TypeES: typeexp.ExpSpec{K: typeexp.One},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	// and
	injecteeProcQN := "injectee-proc-qn"
	_, err = s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: injecteeProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "injectee-provider-ph",
			TypeQN: oneType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				ProcQN: injecteeProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolInjecteePH := "pool-injectee-ph"
	_, err = s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolInjecteePH,
				ProcQN: injecteeProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	callerProcQN := "caller-proc-qn"
	callerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: callerProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "caller-provider-ph",
			TypeQN: oneType.Spec.TypeQN,
		},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "caller-injectee-ph", TypeQN: oneType.Spec.TypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				ProcQN: callerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	callerProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: "pool-caller-ph",
				ProcQN: callerProcQN,
				ValPHs: []string{poolInjecteePH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	calleeProcQN := "callee-proc-qn"
	_, err = s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: calleeProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "callee-provider-ph",
			TypeQN: oneType.Spec.TypeQN,
		},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "callee-injectee-ph", TypeQN: oneType.Spec.TypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	callerCalleePH := "caller-callee-ph"
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: callerProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Call,
			Call: &procexp.CallSpec{
				BindPH: callerCalleePH,
				ProcQN: calleeProcQN,
				ValPHs: []string{callerProcDec.ClientBSs[0].ChnlPH},
				ContES: procexp.ExpSpec{
					K: procexp.Wait,
					Wait: &procexp.WaitSpec{
						CommPH: callerCalleePH,
						ContES: procexp.ExpSpec{
							K: procexp.Close,
							Close: &procexp.CloseSpec{
								CommPH: callerProcDec.ProviderBS.ChnlPH,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// then
	// TODO добавить проверку
}

func (s *suite) fwd(t *testing.T) {
	s.beforeEach(t)
	// given
	myPoolQN := "my-pool-qn"
	myPoolExec, err := s.poolExecAPI.Create(poolexec.ExecSpec{
		PoolQN: myPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneType, err := s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: "one-type-qn",
		TypeES: typeexp.ExpSpec{K: typeexp.One},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcQN := "closer-proc-qn"
	closerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: closerProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "closer-provider-ph",
			TypeQN: oneType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	forwarderProcQN := "forwarder-proc-qn"
	forwarderProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: forwarderProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "forwarder-provider-ph",
			TypeQN: oneType.Spec.TypeQN,
		},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "forwarder-closer-ph", TypeQN: oneType.Spec.TypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterProcQN := "waiter-proc-qn"
	waiterProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: waiterProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "waiter-provider-ph",
			TypeQN: oneType.Spec.TypeQN,
		},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "waiter-forwarder-ph", TypeQN: oneType.Spec.TypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				ProcQN: closerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolCloserPH := "pool-closer-ph"
	closerProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolCloserPH,
				ProcQN: closerProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				ProcQN: forwarderProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolForwarderPH := "pool-forwarder-ph"
	forwarderProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: poolForwarderPH,
				ProcQN: forwarderProcQN,
				ValPHs: []string{poolCloserPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Hire,
			Hire: &poolexp.HireSpec{
				ProcQN: waiterProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	waiterProcExec, err := s.poolExecAPI.Spawn(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			K: poolexp.Spawn,
			Spawn: &poolexp.SpawnSpec{
				BindPH: "pool-waiter-ph",
				ProcQN: waiterProcQN,
				ValPHs: []string{poolForwarderPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: closerProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Close,
			Close: &procexp.CloseSpec{
				CommPH: closerProcDec.ProviderBS.ChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: forwarderProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Fwd,
			Fwd: &procexp.FwdSpec{
				CommPH: forwarderProcDec.ProviderBS.ChnlPH,
				ContPH: forwarderProcDec.ClientBSs[0].ChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: waiterProcExec,
		ProcES: procexp.ExpSpec{
			K: procexp.Wait,
			Wait: &procexp.WaitSpec{
				CommPH: waiterProcDec.ClientBSs[0].ChnlPH,
				ContES: procexp.ExpSpec{
					K: procexp.Close,
					Close: &procexp.CloseSpec{
						CommPH: waiterProcDec.ProviderBS.ChnlPH,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// then
	// TODO добавить проверку
}
