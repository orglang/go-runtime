package pool

import (
	"database/sql"
	"fmt"
	"slices"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"

	"orglang/go-runtime/lib/e2e"
	"orglang/go-runtime/lib/rc"

	"github.com/orglang/go-sdk/adt/pooldec"
	"github.com/orglang/go-sdk/adt/poolexec"
	"github.com/orglang/go-sdk/adt/procdec"
	"github.com/orglang/go-sdk/adt/procexec"
	"github.com/orglang/go-sdk/adt/procexp"
	"github.com/orglang/go-sdk/adt/termctx"
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
	poolSpec1 := poolexec.ExecSpecME{PoolQN: "pool-1"}
	poolRef1, err := s.poolExecAPI.Create(poolSpec1)
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolSpec2 := poolexec.ExecSpecME{PoolQN: "pool-2", SupID: poolRef1.ExecID}
	poolRef2, err := s.poolExecAPI.Create(poolSpec2)
	if err != nil {
		t.Fatal(err)
	}
	// when
	poolSnap1, err := s.poolExecAPI.Retrieve(poolRef1.ExecID)
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
	mainTypeQN := "main-type-qn"
	closerProcQN := "closer-proc-qn"
	waiterProcQN := "waiter-proc-qn"
	_, err := s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: mainTypeQN,
		TypeES: typeexp.ExpSpecME{
			K: typeexp.UpExp,
			Up: &typeexp.FooSpecME{
				ContES: typeexp.ExpSpecME{
					K: typeexp.XactExp,
					Xact: &typeexp.XactSpecME{
						ContESs: map[string]typeexp.ExpSpecME{
							closerProcQN: {
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeQN},
									},
								},
							},
							waiterProcQN: {
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeQN},
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
	mainPoolQN := "main-pool-qn"
	mainProvisionPH := "main-provision-ph"
	mainReceptionPH := "main-reception-ph"
	_, err = s.poolDecAPI.Create(pooldec.DecSpecME{
		PoolSN: mainPoolQN,
		InsiderProvisionBC: termctx.BindClaimME{
			BindPH: mainProvisionPH,
			TypeQN: mainTypeQN,
		},
		InsiderReceptionBC: termctx.BindClaimME{
			BindPH: mainReceptionPH,
			TypeQN: mainTypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneTypeQN := "one-type-qn"
	_, err = s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: oneTypeQN,
		TypeES: typeexp.ExpSpecME{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerDecSpec := procdec.DecSpecME{
		ProcQN: closerProcQN,
		X: termctx.BindClaimME{
			BindPH: "closer-provision-ph",
			TypeQN: oneTypeQN,
		},
	}
	_, err = s.procDecAPI.Create(closerDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcPH := "closer-proc-ph"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Acquire: &procexp.AcqureSpecME{
				CommPH: mainReceptionPH,
				ContES: procexp.ExpSpecME{
					Call: &procexp.CallSpecME{
						CommPH: mainReceptionPH,
						BindPH: closerProcPH,
						ProcQN: closerProcQN,
						ContES: procexp.ExpSpecME{
							Release: &procexp.ReleaseSpecME{
								CommPH: mainReceptionPH,
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
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Accept: &procexp.AcceptSpecME{
				CommPH: mainProvisionPH,
				ContES: procexp.ExpSpecME{
					Spawn: &procexp.SpawnSpecME{
						CommPH: mainProvisionPH,
						ProcQN: closerProcQN,
						ContES: &procexp.ExpSpecME{
							Detach: &procexp.DetachSpecME{
								CommPH: mainProvisionPH,
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
	closerExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterDecSpec := procdec.DecSpecME{
		ProcQN: waiterProcQN,
		X:      termctx.BindClaimME{BindPH: "waiter-provision-ph", TypeQN: oneTypeQN},
		Ys: []termctx.BindClaimME{
			{BindPH: "closer-reception-ph", TypeQN: oneTypeQN},
		},
	}
	_, err = s.procDecAPI.Create(waiterDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Acquire: &procexp.AcqureSpecME{
				CommPH: mainProvisionPH,
				ContES: procexp.ExpSpecME{
					Call: &procexp.CallSpecME{
						BindPH: mainProvisionPH,
						CommPH: "waiter-proc-ph",
						ProcQN: waiterProcQN,
						ValPHs: []string{closerProcPH},
						ContES: procexp.ExpSpecME{
							Release: &procexp.ReleaseSpecME{
								CommPH: mainProvisionPH,
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
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Accept: &procexp.AcceptSpecME{
				CommPH: mainProvisionPH,
				ContES: procexp.ExpSpecME{
					Spawn: &procexp.SpawnSpecME{
						CommPH: mainProvisionPH,
						ProcQN: waiterProcQN,
						ContES: &procexp.ExpSpecME{
							Detach: &procexp.DetachSpecME{
								CommPH: mainProvisionPH,
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
	waiterExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: closerExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Close: &procexp.CloseSpecME{
				CommPH: closerDecSpec.X.BindPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: waiterExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Wait: &procexp.WaitSpecME{
				CommPH: waiterDecSpec.Ys[0].BindPH,
				ContES: procexp.ExpSpecME{
					Close: &procexp.CloseSpecME{
						CommPH: waiterDecSpec.X.BindPH,
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
	mainTypeQN := "main-type-qn"
	senderProcQN := "sender-proc-qn"
	receiverProcQN := "receiver-proc-qn"
	messageProcQN := "message-proc-qn"
	_, err := s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: mainTypeQN,
		TypeES: typeexp.ExpSpecME{
			K: typeexp.UpExp,
			Up: &typeexp.FooSpecME{
				ContES: typeexp.ExpSpecME{
					K: typeexp.XactExp,
					Xact: &typeexp.XactSpecME{
						ContESs: map[string]typeexp.ExpSpecME{
							senderProcQN: typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeQN},
									},
								},
							},
							receiverProcQN: typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeQN},
									},
								},
							},
							messageProcQN: typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeQN},
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
	mainPoolQN := "main-pool-qn"
	mainProvisionPH := "main-provision-ph"
	mainReceptionPH := "main-reception-ph"
	_, err = s.poolDecAPI.Create(pooldec.DecSpecME{
		PoolSN: mainPoolQN,
		InsiderProvisionBC: termctx.BindClaimME{
			BindPH: mainProvisionPH,
			TypeQN: mainPoolQN,
		},
		InsiderReceptionBC: termctx.BindClaimME{
			BindPH: mainReceptionPH,
			TypeQN: mainPoolQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	lolliTypeQN := "lolli-type-qn"
	_, err = s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: lolliTypeQN,
		TypeES: typeexp.ExpSpecME{
			K: typeexp.LolliExp,
			Lolli: &typeexp.ProdSpecME{
				ValES:  typeexp.ExpSpecME{K: typeexp.OneExp},
				ContES: typeexp.ExpSpecME{K: typeexp.OneExp},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneTypeQN := "one-type-qn"
	_, err = s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: oneTypeQN,
		TypeES: typeexp.ExpSpecME{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	receiverDecSpec := procdec.DecSpecME{
		ProcQN: receiverProcQN,
		X: termctx.BindClaimME{
			BindPH: "receiver-provision-ph",
			TypeQN: lolliTypeQN,
		},
	}
	_, err = s.procDecAPI.Create(receiverDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	messageDecSpec := procdec.DecSpecME{
		ProcQN: messageProcQN,
		X: termctx.BindClaimME{
			BindPH: "message-provision-ph",
			TypeQN: oneTypeQN,
		},
	}
	_, err = s.procDecAPI.Create(messageDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	senderDecSpec := procdec.DecSpecME{
		ProcQN: senderProcQN,
		X: termctx.BindClaimME{
			BindPH: "sender-provision-ph",
			TypeQN: oneTypeQN,
		},
		Ys: []termctx.BindClaimME{
			{BindPH: "receiver-reception-ph", TypeQN: lolliTypeQN},
			{BindPH: "message-reception-ph", TypeQN: oneTypeQN},
		},
	}
	_, err = s.procDecAPI.Create(senderDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	receiverProcPH := "receiver-proc-ph"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
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
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
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
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
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
	// and
	receiverExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: receiverExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Recv: &procexp.RecvSpecME{
				BindPH: receiverDecSpec.X.BindPH,
				CommPH: "message-reception-ph",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	senderExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: senderExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Send: &procexp.SendSpecME{
				CommPH: senderDecSpec.Ys[0].BindPH,
				ValPH:  senderDecSpec.Ys[1].BindPH,
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
	mainPoolQN := "main-pool-qn"
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	label := "label-1"
	// and
	withRoleSpec := typedef.DefSpecME{
		TypeQN: "with-role",
		TypeES: typeexp.ExpSpecME{
			With: &typeexp.SumSpecME{
				Choices: []typeexp.ChoiceSpecME{
					{Label: label, ContES: typeexp.ExpSpecME{K: typeexp.OneExp}},
				},
			},
		},
	}
	withRole, err := s.typeDefAPI.Create(withRoleSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneRoleSpec := typedef.DefSpecME{
		TypeQN: "one-role",
		TypeES: typeexp.ExpSpecME{K: typeexp.OneExp},
	}
	oneRole, err := s.typeDefAPI.Create(oneRoleSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	withSigSpec := procdec.DecSpecME{
		ProcQN: "sig-1",
		X: termctx.BindClaimME{
			BindPH: "chnl-1",
			TypeQN: withRole.TypeQN,
		},
	}
	withSig, err := s.procDecAPI.Create(withSigSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneSigSpec := procdec.DecSpecME{
		ProcQN: "sig-2",
		Ys:     []termctx.BindClaimME{withSig.X},
		X: termctx.BindClaimME{
			BindPH: "chnl-2",
			TypeQN: oneRole.TypeQN,
		},
	}
	_, err = s.procDecAPI.Create(oneSigSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	followerPH := "follower"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: followerPH,
				ProcQN: "tbd",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	deciderPH := "decider"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: deciderPH,
				ProcQN: "tbd",
				ValPHs: []string{followerPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	followerExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: followerExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Case: &procexp.CaseSpecME{
				CommPH: followerPH,
				ContBSs: []procexp.BranchSpecME{
					procexp.BranchSpecME{
						Label: label,
						ContES: procexp.ExpSpecME{
							Close: &procexp.CloseSpecME{
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
	deciderExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: deciderExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Lab: &procexp.LabSpecME{
				CommPH: followerPH,
				Label:  label,
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
	mainPoolQN := "main-pool-qn"
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneDef, err := s.typeDefAPI.Create(
		typedef.DefSpecME{
			TypeQN: "one-type",
			TypeES: typeexp.ExpSpecME{K: typeexp.OneExp},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneDec1, err := s.procDecAPI.Create(procdec.DecSpecME{
		ProcQN: "one-proc-1",
		X: termctx.BindClaimME{
			BindPH: "chnl-1",
			TypeQN: oneDef.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	_, err = s.procDecAPI.Create(procdec.DecSpecME{
		ProcQN: "one-proc-2",
		Ys:     []termctx.BindClaimME{oneDec1.X},
		X: termctx.BindClaimME{
			BindPH: "chnl-2",
			TypeQN: oneDef.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneDec3, err := s.procDecAPI.Create(procdec.DecSpecME{
		ProcQN: "one-proc-3",
		Ys:     []termctx.BindClaimME{oneDec1.X},
		X: termctx.BindClaimME{
			BindPH: "chnl-3",
			TypeQN: oneDef.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	poolExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: "pool-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	injecteePH := "injectee"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: poolExecRef.ExecID,
		ExecID: poolExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: injecteePH,
				ProcQN: "tbd",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	spawnerPH := "spawner"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: poolExecRef.ExecID,
		ExecID: poolExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: spawnerPH,
				ProcQN: "tbd",
				ValPHs: []string{injecteePH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	x := "x"
	// and
	spawnerExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: poolExecRef.ExecID,
		ExecID: spawnerExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Spawn: &procexp.SpawnSpecME{
				DecID:   oneDec3.DecID,
				BindPHs: []string{injecteePH},
				CommPH:  x,
				ContES: &procexp.ExpSpecME{
					Wait: &procexp.WaitSpecME{
						CommPH: x,
						ContES: procexp.ExpSpecME{
							Close: &procexp.CloseSpecME{
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
	mainPoolQN := "main-pool-qn"
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolQN,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneDefSnap, err := s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: "one-role",
		TypeES: typeexp.ExpSpecME{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneDecSnap, err := s.procDecAPI.Create(procdec.DecSpecME{
		ProcQN: "sig-1",
		X: termctx.BindClaimME{
			BindPH: "chnl-1",
			TypeQN: oneDefSnap.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	_, err = s.procDecAPI.Create(procdec.DecSpecME{
		ProcQN: "sig-2",
		Ys:     []termctx.BindClaimME{oneDecSnap.X},
		X: termctx.BindClaimME{
			BindPH: "chnl-2",
			TypeQN: oneDefSnap.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	_, err = s.procDecAPI.Create(procdec.DecSpecME{
		ProcQN: "sig-3",
		Ys:     []termctx.BindClaimME{oneDecSnap.X},
		X: termctx.BindClaimME{
			BindPH: "chnl-3",
			TypeQN: oneDefSnap.TypeQN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerChnlPH := "closer"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: closerChnlPH,
				ProcQN: "tbd",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	forwarderChnlPH := "forwarder"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: forwarderChnlPH,
				ProcQN: "tbd",
				ValPHs: []string{closerChnlPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterChnlPH := "waiter"
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: waiterChnlPH,
				ProcQN: "tbd",
				ValPHs: []string{forwarderChnlPH},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: closerExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Close: &procexp.CloseSpecME{
				CommPH: closerChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	forwarderExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// when
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: forwarderExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Fwd: &procexp.FwdSpecME{
				X: forwarderChnlPH,
				Y: closerChnlPH,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterExecRef, err := s.poolExecAPI.Poll(poolexec.PollSpecME{
		ExecID: mainExecRef.ExecID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: waiterExecRef.ExecID,
		ProcES: procexp.ExpSpecME{
			Wait: &procexp.WaitSpecME{
				CommPH: forwarderChnlPH,
				ContES: procexp.ExpSpecME{
					Close: &procexp.CloseSpecME{
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
