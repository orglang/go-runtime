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
)

func TestPool(t *testing.T) {
	s := suite{}
	s.beforeAll(t)
	t.Run("CreateRetreive", s.createRetreive)
	t.Run("WaitClose", s.waitClose)
	t.Run("RecvSend", s.recvSend)
	t.Run("CaseLab", s.caseLab)
	t.Run("SpawnCall", s.spawnCall)
	t.Run("Fwd", s.fwd)

}

type suite struct {
	poolDecAPI  e2e.PoolDecAPI
	poolExecAPI e2e.PoolExecAPI
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
			s.procDecAPI,
			s.procExecAPI,
			s.typeDefAPI,
		))
	t.Cleanup(func() { app.Stop(t.Context()) })
}

func (s *suite) beforeEach(t *testing.T) {
	tables := []string{
		"aliases",
		"pool_roots", "pool_liabs", "proc_bnds", "proc_steps",
		"sig_roots", "sig_pes", "sig_ces",
		"type_def_roots", "type_term_states",
		"type_term_states"}
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
	if !slices.Contains(poolSnap1.Subs, poolRef2) {
		t.Errorf("unexpected subs in %q; want: %+v, got: %+v",
			poolSpec1.PoolQN, poolRef2, poolSnap1.Subs)
	}
}

func (s *suite) waitClose(t *testing.T) {
	s.beforeEach(t)
	// given
	poolTypeQN := "pool-type-qn"
	closerProcQN := "closer-proc-qn"
	waiterProcQN := "waiter-proc-qn"
	_, err := s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: poolTypeQN,
		TypeES: typeexp.ExpSpec{
			K: typeexp.UpExp,
			Up: &typeexp.ShiftSpec{
				ContES: typeexp.ExpSpec{
					K: typeexp.XactExp,
					Xact: &typeexp.XactSpec{
						ContESs: map[string]typeexp.ExpSpec{
							closerProcQN: {
								K: typeexp.DownExp,
								Down: &typeexp.ShiftSpec{
									ContES: typeexp.ExpSpec{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpec{TypeQN: poolTypeQN},
									},
								},
							},
							waiterProcQN: {
								K: typeexp.DownExp,
								Down: &typeexp.ShiftSpec{
									ContES: typeexp.ExpSpec{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpec{TypeQN: poolTypeQN},
									},
								},
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
	myPoolQN := "my-pool-qn"
	poolProvisionPH := "pool-provision-ph"
	poolReceptionPH := "pool-reception-ph"
	_, err = s.poolDecAPI.Create(pooldec.DecSpec{
		PoolQN: myPoolQN,
		InsiderProvisionBS: procbind.BindSpec{
			ChnlPH: poolProvisionPH,
			TypeQN: poolTypeQN,
		},
		InsiderReceptionBS: procbind.BindSpec{
			ChnlPH: poolReceptionPH,
			TypeQN: poolTypeQN,
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
	oneTypeQN := "one-type-qn"
	_, err = s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: oneTypeQN,
		TypeES: typeexp.ExpSpec{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: closerProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "closer-provision-ph",
			TypeQN: oneTypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcPH := "closer-proc-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Acquire: &poolexp.AcquireSpec{
				CommPH: poolReceptionPH,
				ContES: poolexp.ExpSpec{
					Hire: &poolexp.HireSpec{
						CommPH: poolReceptionPH,
						BindPH: closerProcPH,
						ProcQN: closerProcQN,
						ContES: poolexp.ExpSpec{
							Release: &poolexp.ReleaseSpec{
								CommPH: poolReceptionPH,
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
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Accept: &poolexp.AcceptSpec{
				CommPH: poolProvisionPH,
				ContES: poolexp.ExpSpec{
					Apply: &poolexp.ApplySpec{
						CommPH: poolProvisionPH,
						ProcQN: closerProcQN,
						ContES: poolexp.ExpSpec{
							Detach: &poolexp.DetachSpec{
								CommPH: poolProvisionPH,
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
	waiterProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:     waiterProcQN,
		ProviderBS: procbind.BindSpec{ChnlPH: "waiter-provision-ph", TypeQN: oneTypeQN},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "closer-reception-ph", TypeQN: oneTypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterProcPH := "waiter-proc-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Acquire: &poolexp.AcquireSpec{
				CommPH: poolProvisionPH,
				ContES: poolexp.ExpSpec{
					Hire: &poolexp.HireSpec{
						BindPH: poolProvisionPH,
						CommPH: waiterProcPH,
						ProcQN: waiterProcQN,
						ValPHs: []string{closerProcPH},
						ContES: poolexp.ExpSpec{
							Release: &poolexp.ReleaseSpec{
								CommPH: poolProvisionPH,
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
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Accept: &poolexp.AcceptSpec{
				CommPH: poolProvisionPH,
				ContES: poolexp.ExpSpec{
					Apply: &poolexp.ApplySpec{
						CommPH: poolProvisionPH,
						ProcQN: waiterProcQN,
						ContES: poolexp.ExpSpec{
							Detach: &poolexp.DetachSpec{
								CommPH: poolProvisionPH,
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
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: closerProcExec,
		ProcES: procexp.ExpSpec{
			Close: &procexp.CloseSpec{
				CommPH: closerProcDec.Spec.ProviderBS.ChnlPH,
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
			Wait: &procexp.WaitSpec{
				CommPH: waiterProcDec.Spec.ClientBSs[0].ChnlPH,
				ContES: procexp.ExpSpec{
					Close: &procexp.CloseSpec{
						CommPH: waiterProcDec.Spec.ProviderBS.ChnlPH,
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
	poolTypeQN := "pool-type-qn"
	senderProcQN := "sender-proc-qn"
	receiverProcQN := "receiver-proc-qn"
	messageProcQN := "message-proc-qn"
	_, err := s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: poolTypeQN,
		TypeES: typeexp.ExpSpec{
			K: typeexp.UpExp,
			Up: &typeexp.ShiftSpec{
				ContES: typeexp.ExpSpec{
					K: typeexp.XactExp,
					Xact: &typeexp.XactSpec{
						ContESs: map[string]typeexp.ExpSpec{
							senderProcQN: typeexp.ExpSpec{
								K: typeexp.DownExp,
								Down: &typeexp.ShiftSpec{
									ContES: typeexp.ExpSpec{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpec{TypeQN: poolTypeQN},
									},
								},
							},
							receiverProcQN: typeexp.ExpSpec{
								K: typeexp.DownExp,
								Down: &typeexp.ShiftSpec{
									ContES: typeexp.ExpSpec{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpec{TypeQN: poolTypeQN},
									},
								},
							},
							messageProcQN: typeexp.ExpSpec{
								K: typeexp.DownExp,
								Down: &typeexp.ShiftSpec{
									ContES: typeexp.ExpSpec{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpec{TypeQN: poolTypeQN},
									},
								},
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
	myPoolQN := "my-pool-qn"
	mainProvisionPH := "pool-provision-ph"
	mainReceptionPH := "pool-reception-ph"
	_, err = s.poolDecAPI.Create(pooldec.DecSpec{
		PoolQN: myPoolQN,
		InsiderProvisionBS: procbind.BindSpec{
			ChnlPH: mainProvisionPH,
			TypeQN: myPoolQN,
		},
		InsiderReceptionBS: procbind.BindSpec{
			ChnlPH: mainReceptionPH,
			TypeQN: myPoolQN,
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
			K: typeexp.LolliExp,
			Lolli: &typeexp.ProdSpec{
				ValES:  typeexp.ExpSpec{K: typeexp.OneExp},
				ContES: typeexp.ExpSpec{K: typeexp.OneExp},
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
		TypeES: typeexp.ExpSpec{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	receiverProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: receiverProcQN,
		ProviderBS: procbind.BindSpec{
			ChnlPH: "receiver-provision-ph",
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
			ChnlPH: "message-provision-ph",
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
			ChnlPH: "sender-provision-ph",
			TypeQN: oneTypeQN,
		},
		ClientBSs: []procbind.BindSpec{
			{ChnlPH: "receiver-reception-ph", TypeQN: lolliTypeQN},
			{ChnlPH: "message-reception-ph", TypeQN: oneTypeQN},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	receiverProcPH := "receiver-proc-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: mainReceptionPH,
				CommPH: receiverProcPH,
				ProcQN: receiverProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	messageProcPH := "message-proc-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: mainReceptionPH,
				CommPH: messageProcPH,
				ProcQN: messageProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	senderProcPH := "sender-proc-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: mainReceptionPH,
				CommPH: senderProcPH,
				ProcQN: senderProcQN,
				ValPHs: []string{receiverProcPH, messageProcPH},
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
			Recv: &procexp.RecvSpec{
				BindPH: receiverProcDec.Spec.ProviderBS.ChnlPH,
				CommPH: "message-reception-ph",
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
			Send: &procexp.SendSpec{
				CommPH: senderProcDec.Spec.ClientBSs[0].ChnlPH,
				ValPH:  senderProcDec.Spec.ClientBSs[1].ChnlPH,
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
	labelQN := "label-1"
	// and
	withType, err := s.typeDefAPI.Create(typedef.DefSpec{
		TypeQN: "with-type-qn",
		TypeES: typeexp.ExpSpec{
			With: &typeexp.SumSpec{
				Choices: []typeexp.ChoiceSpec{
					{LabQN: labelQN, ContES: typeexp.ExpSpec{K: typeexp.OneExp}},
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
		TypeES: typeexp.ExpSpec{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	followerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: "follower-proc-qn",
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-1",
			TypeQN: withType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	deciderProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:    "decider-proc-qn",
		ClientBSs: []procbind.BindSpec{followerProcDec.Spec.ProviderBS},
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-2",
			TypeQN: oneType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	followerPH := "follower-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: followerPH,
				ProcQN: followerProcDec.Spec.ProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	deciderPH := "decider-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: deciderPH,
				ProcQN: deciderProcDec.Spec.ProcQN,
				ValPHs: []string{followerPH},
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
			Case: &procexp.CaseSpec{
				CommPH: followerPH,
				ContBSs: []procexp.BranchSpec{
					procexp.BranchSpec{
						LabQN: labelQN,
						ContES: procexp.ExpSpec{
							Close: &procexp.CloseSpec{
								CommPH: followerPH,
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
			Lab: &procexp.LabSpec{
				CommPH: followerPH,
				LabQN:  labelQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// then
	// TODO добавить проверку
}

func (s *suite) spawnCall(t *testing.T) {
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
	oneTypeDef, err := s.typeDefAPI.Create(
		typedef.DefSpec{
			TypeQN: "one-type-qn",
			TypeES: typeexp.ExpSpec{K: typeexp.OneExp},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: "closer-proc-qn",
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-1",
			TypeQN: oneTypeDef.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	injecteeProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:    "injectee-proc-qn",
		ClientBSs: []procbind.BindSpec{closerProcDec.Spec.ProviderBS},
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-2",
			TypeQN: oneTypeDef.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	spawnerProcDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:    "spawner-proc-qn",
		ClientBSs: []procbind.BindSpec{closerProcDec.Spec.ProviderBS},
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-3",
			TypeQN: oneTypeDef.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	injecteePH := "injectee-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: injecteePH,
				ProcQN: injecteeProcDec.Spec.ProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	spawnerPH := "spawner-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: spawnerPH,
				ProcQN: spawnerProcDec.Spec.ProcQN,
				ValPHs: []string{injecteePH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	x := "x"
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: spawnerProcExec,
		ProcES: procexp.ExpSpec{
			Spawn: &procexp.SpawnSpec{
				BindPHs: []string{injecteePH},
				CommPH:  x,
				ContES: procexp.ExpSpec{
					Wait: &procexp.WaitSpec{
						CommPH: x,
						ContES: procexp.ExpSpec{
							Close: &procexp.CloseSpec{
								CommPH: spawnerPH,
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
		TypeES: typeexp.ExpSpec{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN: "closer-proc-qn",
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-1",
			TypeQN: oneType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	forwarderDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:    "forwarder-proc-qn",
		ClientBSs: []procbind.BindSpec{closerDec.Spec.ProviderBS},
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-2",
			TypeQN: oneType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterDec, err := s.procDecAPI.Create(procdec.DecSpec{
		ProcQN:    "waiter-proc-qn",
		ClientBSs: []procbind.BindSpec{closerDec.Spec.ProviderBS},
		ProviderBS: procbind.BindSpec{
			ChnlPH: "chnl-3",
			TypeQN: oneType.Spec.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerChnlPH := "closer-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: closerChnlPH,
				ProcQN: closerDec.Spec.ProcQN,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	forwarderChnlPH := "forwarder-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: forwarderChnlPH,
				ProcQN: forwarderDec.Spec.ProcQN,
				ValPHs: []string{closerChnlPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterChnlPH := "waiter-ph"
	err = s.poolExecAPI.Take(poolstep.StepSpec{
		ExecRef: myPoolExec,
		PoolES: poolexp.ExpSpec{
			Hire: &poolexp.HireSpec{
				BindPH: waiterChnlPH,
				ProcQN: waiterDec.Spec.ProcQN,
				ValPHs: []string{forwarderChnlPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: closerExec,
		ProcES: procexp.ExpSpec{
			Close: &procexp.CloseSpec{
				CommPH: closerChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: forwarderExec,
		ProcES: procexp.ExpSpec{
			Fwd: &procexp.FwdSpec{
				CommPH: forwarderChnlPH,
				ContPH: closerChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Take(procstep.StepSpec{
		ExecRef: waiterExec,
		ProcES: procexp.ExpSpec{
			Wait: &procexp.WaitSpec{
				CommPH: forwarderChnlPH,
				ContES: procexp.ExpSpec{
					Close: &procexp.CloseSpec{
						CommPH: waiterChnlPH,
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
