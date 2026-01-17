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

	"orglang/go-runtime/adt/qualsym"

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
	mainTypeSN := qualsym.New("main-type-sn")
	closerProcQN := qualsym.New("closer-proc-qn")
	waiterProcQN := qualsym.New("waiter-proc-qn")
	_, err := s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: mainTypeSN.String(),
		TypeES: typeexp.ExpSpecME{
			K: typeexp.UpExp,
			Up: &typeexp.FooSpecME{
				ContES: typeexp.ExpSpecME{
					K: typeexp.XactExp,
					Xact: &typeexp.XactSpecME{
						ContESs: map[string]typeexp.ExpSpecME{
							closerProcQN.String(): typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeSN.String()},
									},
								},
							},
							waiterProcQN.String(): typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeSN.String()},
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
	mainPoolSN := qualsym.New("main-pool-sn")
	mainProvisionPH := qualsym.New("main-provision-ph")
	mainReceptionPH := qualsym.New("main-reception-ph")
	_, err = s.poolDecAPI.Create(pooldec.DecSpecME{
		PoolSN: mainPoolSN.String(),
		InsiderProvisionBC: termctx.BindClaimME{
			BindPH: mainProvisionPH.String(),
			TypeQN: mainTypeSN.String(),
		},
		InsiderReceptionBC: termctx.BindClaimME{
			BindPH: mainReceptionPH.String(),
			TypeQN: mainTypeSN.String(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolSN.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	oneTypeSN := qualsym.New("one-type-sn")
	_, err = s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: oneTypeSN.String(),
		TypeES: typeexp.ExpSpecME{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerDecSpec := procdec.DecSpecME{
		ProcQN: closerProcQN.String(),
		X: termctx.BindClaimME{
			BindPH: "closer-provision-ph",
			TypeQN: oneTypeSN.String(),
		},
	}
	_, err = s.procDecAPI.Create(closerDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	closerProcPH := qualsym.New("closer-proc-ph")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Acquire: &procexp.AcqureSpecME{
				CommPH: mainReceptionPH.String(),
				ContES: procexp.ExpSpecME{
					Call: &procexp.CallSpecME{
						CommPH: mainReceptionPH.String(),
						BindPH: closerProcPH.String(),
						ProcQN: closerProcQN.String(),
						ContES: procexp.ExpSpecME{
							Release: &procexp.ReleaseSpecME{
								CommPH: mainReceptionPH.String(),
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
				CommPH: mainProvisionPH.String(),
				ContES: procexp.ExpSpecME{
					Spawn: &procexp.SpawnSpecME{
						CommPH: mainProvisionPH.String(),
						ProcQN: closerProcQN.String(),
						ContES: &procexp.ExpSpecME{
							Detach: &procexp.DetachSpecME{
								CommPH: mainProvisionPH.String(),
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
		ProcQN: waiterProcQN.String(),
		X:      termctx.BindClaimME{BindPH: "waiter-provision-ph", TypeQN: oneTypeSN.String()},
		Ys: []termctx.BindClaimME{
			{BindPH: "closer-reception-ph", TypeQN: oneTypeSN.String()},
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
				CommPH: mainProvisionPH.String(),
				ContES: procexp.ExpSpecME{
					Call: &procexp.CallSpecME{
						BindPH: mainProvisionPH.String(),
						CommPH: qualsym.New("waiter-proc-ph").String(),
						ProcQN: waiterProcQN.String(),
						ValPHs: []string{closerProcPH.String()},
						ContES: procexp.ExpSpecME{
							Release: &procexp.ReleaseSpecME{
								CommPH: mainProvisionPH.String(),
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
				CommPH: mainProvisionPH.String(),
				ContES: procexp.ExpSpecME{
					Spawn: &procexp.SpawnSpecME{
						CommPH: mainProvisionPH.String(),
						ProcQN: waiterProcQN.String(),
						ContES: &procexp.ExpSpecME{
							Detach: &procexp.DetachSpecME{
								CommPH: mainProvisionPH.String(),
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
	mainTypeSN := qualsym.New("main-type-sn")
	senderProcSN := qualsym.New("sender-proc-sn")
	receiverProcSN := qualsym.New("receiver-proc-sn")
	messageProcSN := qualsym.New("message-proc-sn")
	_, err := s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: mainTypeSN.String(),
		TypeES: typeexp.ExpSpecME{
			K: typeexp.UpExp,
			Up: &typeexp.FooSpecME{
				ContES: typeexp.ExpSpecME{
					K: typeexp.XactExp,
					Xact: &typeexp.XactSpecME{
						ContESs: map[string]typeexp.ExpSpecME{
							senderProcSN.String(): typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeSN.String()},
									},
								},
							},
							receiverProcSN.String(): typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeSN.String()},
									},
								},
							},
							messageProcSN.String(): typeexp.ExpSpecME{
								K: typeexp.DownExp,
								Down: &typeexp.FooSpecME{
									ContES: typeexp.ExpSpecME{
										K:    typeexp.LinkExp,
										Link: &typeexp.LinkSpecME{TypeQN: mainTypeSN.String()},
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
	mainPoolSN := qualsym.New("main-pool-sn")
	mainProvisionPH := qualsym.New("main-provision-ph")
	mainReceptionPH := qualsym.New("main-reception-ph")
	_, err = s.poolDecAPI.Create(pooldec.DecSpecME{
		PoolSN: mainPoolSN.String(),
		InsiderProvisionBC: termctx.BindClaimME{
			BindPH: mainProvisionPH.String(),
			TypeQN: mainTypeSN.String(),
		},
		InsiderReceptionBC: termctx.BindClaimME{
			BindPH: mainReceptionPH.String(),
			TypeQN: mainTypeSN.String(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolSN.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	lolliTypeSN := qualsym.New("lolli-type-sn")
	_, err = s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: lolliTypeSN.String(),
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
	oneTypeSN := qualsym.New("one-type-sn")
	_, err = s.typeDefAPI.Create(typedef.DefSpecME{
		TypeQN: oneTypeSN.String(),
		TypeES: typeexp.ExpSpecME{K: typeexp.OneExp},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	receiverDecSpec := procdec.DecSpecME{
		ProcQN: receiverProcSN.String(),
		X: termctx.BindClaimME{
			BindPH: "receiver-provision-ph",
			TypeQN: lolliTypeSN.String(),
		},
	}
	_, err = s.procDecAPI.Create(receiverDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	messageDecSpec := procdec.DecSpecME{
		ProcQN: messageProcSN.String(),
		X: termctx.BindClaimME{
			BindPH: "message-provision-ph",
			TypeQN: oneTypeSN.String(),
		},
	}
	_, err = s.procDecAPI.Create(messageDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	senderDecSpec := procdec.DecSpecME{
		ProcQN: senderProcSN.String(),
		X: termctx.BindClaimME{
			BindPH: "sender-provision-ph",
			TypeQN: oneTypeSN.String(),
		},
		Ys: []termctx.BindClaimME{
			{BindPH: "receiver-reception-ph", TypeQN: lolliTypeSN.String()},
			{BindPH: "message-reception-ph", TypeQN: oneTypeSN.String()},
		},
	}
	_, err = s.procDecAPI.Create(senderDecSpec)
	if err != nil {
		t.Fatal(err)
	}
	// and
	receiverProcPH := qualsym.New("receiver-proc-ph")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: mainReceptionPH.String(),
				CommPH: receiverProcPH.String(),
				ProcQN: receiverProcSN.String(),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	messageProcPH := qualsym.New("message-proc-ph")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: mainReceptionPH.String(),
				CommPH: messageProcPH.String(),
				ProcQN: messageProcSN.String(),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	senderProcPH := qualsym.New("sender-proc-ph")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: mainReceptionPH.String(),
				CommPH: senderProcPH.String(),
				ProcQN: senderProcSN.String(),
				ValPHs: []string{receiverProcPH.String(), messageProcPH.String()},
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
				CommPH: qualsym.New("message-reception-ph").String(),
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
	mainPoolSN := qualsym.New("main-pool-sn")
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolSN.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	label := qualsym.ADT("label-1")
	// and
	withRoleSpec := typedef.DefSpecME{
		TypeQN: "with-role",
		TypeES: typeexp.ExpSpecME{
			With: &typeexp.SumSpecME{
				Choices: []typeexp.ChoiceSpecME{
					{Label: label.String(), ContES: typeexp.ExpSpecME{K: typeexp.OneExp}},
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
	followerPH := qualsym.New("follower")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: followerPH.String(),
				ProcQN: "tbd",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	deciderPH := qualsym.New("decider")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: deciderPH.String(),
				ProcQN: "tbd",
				ValPHs: []string{followerPH.String()},
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
				CommPH: followerPH.String(),
				ContBSs: []procexp.BranchSpecME{
					procexp.BranchSpecME{
						Label: label.String(),
						ContES: procexp.ExpSpecME{
							Close: &procexp.CloseSpecME{
								CommPH: followerPH.String(),
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
				CommPH: followerPH.String(),
				Label:  label.String(),
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
	mainPoolSN := qualsym.New("main-pool-sn")
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolSN.String(),
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
	injecteePH := qualsym.New("injectee")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: poolExecRef.ExecID,
		ExecID: poolExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: injecteePH.String(),
				ProcQN: "tbd",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	spawnerPH := qualsym.New("spawner")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: poolExecRef.ExecID,
		ExecID: poolExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: spawnerPH.String(),
				ProcQN: "tbd",
				ValPHs: []string{injecteePH.String()},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	x := qualsym.New("x")
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
				BindPHs: []string{injecteePH.String()},
				CommPH:  x.String(),
				ContES: &procexp.ExpSpecME{
					Wait: &procexp.WaitSpecME{
						CommPH: x.String(),
						ContES: procexp.ExpSpecME{
							Close: &procexp.CloseSpecME{
								CommPH: spawnerPH.String(),
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
	mainPoolQN := qualsym.New("main-pool-qn")
	mainExecRef, err := s.poolExecAPI.Create(poolexec.ExecSpecME{
		PoolQN: mainPoolQN.String(),
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
	closerChnlPH := qualsym.New("closer")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: closerChnlPH.String(),
				ProcQN: "tbd",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	forwarderChnlPH := qualsym.New("forwarder")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: forwarderChnlPH.String(),
				ProcQN: "tbd",
				ValPHs: []string{closerChnlPH.String()},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// and
	waiterChnlPH := qualsym.New("waiter")
	err = s.procExecAPI.Run(procexec.ExecSpecME{
		PoolID: mainExecRef.ExecID,
		ExecID: mainExecRef.ProcID,
		ProcES: procexp.ExpSpecME{
			Call: &procexp.CallSpecME{
				BindPH: waiterChnlPH.String(),
				ProcQN: "tbd",
				ValPHs: []string{forwarderChnlPH.String()},
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
				CommPH: closerChnlPH.String(),
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
				X: forwarderChnlPH.String(),
				Y: closerChnlPH.String(),
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
				CommPH: forwarderChnlPH.String(),
				ContES: procexp.ExpSpecME{
					Close: &procexp.CloseSpecME{
						CommPH: waiterChnlPH.String(),
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
