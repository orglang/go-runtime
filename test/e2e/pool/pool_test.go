package pool_test

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"smecalculus/rolevod/lib/sym"

	pooldec "smecalculus/rolevod/app/pool/dec"
	pooldef "smecalculus/rolevod/app/pool/def"
	poolexec "smecalculus/rolevod/app/pool/exec"
	procdec "smecalculus/rolevod/app/proc/dec"
	procdef "smecalculus/rolevod/app/proc/def"
	procexec "smecalculus/rolevod/app/proc/exec"
	typedef "smecalculus/rolevod/app/type/def"
)

var (
	poolDecAPI  = pooldec.NewAPI()
	poolExecAPI = poolexec.NewAPI()
	procDecAPI  = procdec.NewAPI()
	procExecAPI = procexec.NewAPI()
	typeDefAPI  = typedef.NewAPI()
	tc          *testCase
)

func TestMain(m *testing.M) {
	ts := testSuite{}
	tc = ts.Setup()
	code := m.Run()
	ts.Teardown()
	os.Exit(code)
}

type testSuite struct {
	db *sql.DB
}

func (ts *testSuite) Setup() *testCase {
	db, err := sql.Open("pgx", "postgres://rolevod:rolevod@localhost:5432/rolevod")
	if err != nil {
		panic(err)
	}
	ts.db = db
	return &testCase{db}
}

func (ts *testSuite) Teardown() {
	err := ts.db.Close()
	if err != nil {
		panic(err)
	}
}

type testCase struct {
	db *sql.DB
}

func (tc *testCase) Setup(t *testing.T) {
	tables := []string{
		"aliases",
		"pool_roots", "pool_liabs", "proc_bnds", "proc_steps",
		"sig_roots", "sig_pes", "sig_ces",
		"role_roots", "role_states",
		"states"}
	for _, table := range tables {
		_, err := tc.db.Exec(fmt.Sprintf("truncate table %v", table))
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestCreation(t *testing.T) {

	t.Run("CreateRetreive", func(t *testing.T) {
		// given
		poolSpec1 := poolexec.PoolSpec{PoolQN: "ts1"}
		poolRef1, err := poolExecAPI.Create(poolSpec1)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec2 := poolexec.PoolSpec{PoolQN: "ts2", SupID: poolRef1.ExecID}
		poolRef2, err := poolExecAPI.Create(poolSpec2)
		if err != nil {
			t.Fatal(err)
		}
		// when
		poolSnap1, err := poolExecAPI.Retrieve(poolRef1.ExecID)
		if err != nil {
			t.Fatal(err)
		}
		// then
		if !slices.Contains(poolSnap1.Subs, poolRef2) {
			t.Errorf("unexpected subs in %q; want: %+v, got: %+v",
				poolSpec1.PoolQN, poolRef2, poolSnap1.Subs)
		}
	})
}

func TestTaking(t *testing.T) {

	t.Run("WaitClose", func(t *testing.T) {
		tc.Setup(t)
		// given
		mainTypeSN := sym.New("main-type-sn")
		closerProcSN := sym.New("closer-proc-sn")
		waiterProcSN := sym.New("waiter-proc-sn")
		_, err := typeDefAPI.Create(typedef.TypeSpec{
			TypeSN: mainTypeSN,
			TypeTS: typedef.UpSpec{
				Z: typedef.XactSpec{
					Zs: map[sym.ADT]typedef.TermSpec{
						closerProcSN: typedef.DownSpec{
							Z: typedef.LinkSpec{TypeQN: mainTypeSN},
						},
						waiterProcSN: typedef.DownSpec{
							Z: typedef.LinkSpec{TypeQN: mainTypeSN},
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		mainPoolSN := sym.New("main-pool-sn")
		mainProvisionPH := sym.New("main-provision-ph")
		mainReceptionPH := sym.New("main-reception-ph")
		_, err = poolDecAPI.Create(pooldec.PoolSpec{
			PoolSN: mainPoolSN,
			InsiderProvisionEP: pooldec.ChnlSpec{
				CommPH: mainProvisionPH,
				TypeQN: mainTypeSN,
			},
			InsiderReceptionEP: pooldec.ChnlSpec{
				CommPH: mainReceptionPH,
				TypeQN: mainTypeSN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		mainExecRef, err := poolExecAPI.Create(poolexec.PoolSpec{
			PoolQN: mainPoolSN,
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneTypeSN := sym.New("one-type-sn")
		_, err = typeDefAPI.Create(typedef.TypeSpec{
			TypeSN: oneTypeSN,
			TypeTS: typedef.OneSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerDecSpec := procdec.ProcSpec{
			ProcSN: closerProcSN,
			ProvisionEP: procdec.ChnlSpec{
				CommPH: sym.New("closer-provision-ph"),
				TypeQN: oneTypeSN,
			},
		}
		_, err = procDecAPI.Create(closerDecSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerProcPH := sym.New("closer-proc-ph")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.AcqureSpec{
				CommPH: mainReceptionPH,
				ContTS: procdef.CallSpec{
					CommPH: mainReceptionPH,
					BindPH: closerProcPH,
					ProcSN: closerProcSN,
					ContTS: procdef.ReleaseSpec{
						CommPH: mainReceptionPH,
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.AcceptSpec{
				CommPH: mainProvisionPH,
				ContTS: procdef.SpawnSpec{
					CommPH: mainProvisionPH,
					ProcSN: closerProcSN,
					ContTS: procdef.DetachSpec{
						CommPH: mainProvisionPH,
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.CloseSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterDecSpec := procdec.ProcSpec{
			ProcSN:      waiterProcSN,
			ProvisionEP: procdec.ChnlSpec{CommPH: sym.New("waiter-provision-ph"), TypeQN: oneTypeSN},
			ReceptionEPs: []procdec.ChnlSpec{
				{CommPH: sym.New("closer-reception-ph"), TypeQN: oneTypeSN},
			},
		}
		_, err = procDecAPI.Create(waiterDecSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.AcqureSpec{
				CommPH: mainProvisionPH,
				ContTS: procdef.CallSpec{
					CommPH: mainProvisionPH,
					BindPH: sym.New("waiter-proc-ph"),
					ProcSN: waiterProcSN,
					ValPHs: []sym.ADT{closerProcPH},
					ContTS: procdef.ReleaseSpec{
						CommPH: mainProvisionPH,
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.AcceptSpec{
				CommPH: mainProvisionPH,
				ContTS: procdef.SpawnSpec{
					CommPH: mainProvisionPH,
					ProcSN: waiterProcSN,
					ContTS: procdef.DetachSpec{
						CommPH: mainProvisionPH,
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.WaitSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// when
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: closerExecRef.ExecID,
			ProcTS: procdef.CloseSpec{
				CommPH: closerDecSpec.ProvisionEP.CommPH,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: waiterExecRef.ExecID,
			ProcTS: procdef.WaitSpec{
				CommPH: waiterDecSpec.ReceptionEPs[0].CommPH,
				ContTS: procdef.CloseSpec{
					CommPH: waiterDecSpec.ProvisionEP.CommPH,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("RecvSend", func(t *testing.T) {
		tc.Setup(t)
		// given
		mainTypeSN := sym.New("main-type-sn")
		senderProcSN := sym.New("sender-proc-sn")
		receiverProcSN := sym.New("receiver-proc-sn")
		messageProcSN := sym.New("message-proc-sn")
		_, err := typeDefAPI.Create(typedef.TypeSpec{
			TypeSN: mainTypeSN,
			TypeTS: typedef.UpSpec{
				Z: typedef.XactSpec{
					Zs: map[sym.ADT]typedef.TermSpec{
						senderProcSN: typedef.DownSpec{
							Z: typedef.LinkSpec{TypeQN: mainTypeSN},
						},
						receiverProcSN: typedef.DownSpec{
							Z: typedef.LinkSpec{TypeQN: mainTypeSN},
						},
						messageProcSN: typedef.DownSpec{
							Z: typedef.LinkSpec{TypeQN: mainTypeSN},
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		mainPoolSN := sym.New("main-pool-sn")
		mainProvisionPH := sym.New("main-provision-ph")
		mainReceptionPH := sym.New("main-reception-ph")
		_, err = poolDecAPI.Create(pooldec.PoolSpec{
			PoolSN: mainPoolSN,
			InsiderProvisionEP: pooldec.ChnlSpec{
				CommPH: mainProvisionPH,
				TypeQN: mainTypeSN,
			},
			InsiderReceptionEP: pooldec.ChnlSpec{
				CommPH: mainReceptionPH,
				TypeQN: mainTypeSN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		mainExecRef, err := poolExecAPI.Create(poolexec.PoolSpec{
			PoolQN: mainPoolSN,
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		lolliTypeSN := sym.New("lolli-type-sn")
		_, err = typeDefAPI.Create(typedef.TypeSpec{
			TypeSN: lolliTypeSN,
			TypeTS: typedef.LolliSpec{
				Y: typedef.OneSpec{},
				Z: typedef.OneSpec{},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneTypeSN := sym.New("one-type-sn")
		_, err = typeDefAPI.Create(typedef.TypeSpec{
			TypeSN: oneTypeSN,
			TypeTS: typedef.OneSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		receiverDecSpec := procdec.ProcSpec{
			ProcSN: receiverProcSN,
			ProvisionEP: procdec.ChnlSpec{
				CommPH: sym.New("receiver-provision-ph"),
				TypeQN: lolliTypeSN,
			},
		}
		_, err = procDecAPI.Create(receiverDecSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		messageDecSpec := procdec.ProcSpec{
			ProcSN: messageProcSN,
			ProvisionEP: procdec.ChnlSpec{
				CommPH: sym.New("message-provision-ph"),
				TypeQN: oneTypeSN,
			},
		}
		_, err = procDecAPI.Create(messageDecSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		senderDecSpec := procdec.ProcSpec{
			ProcSN: senderProcSN,
			ProvisionEP: procdec.ChnlSpec{
				CommPH: sym.New("sender-provision-ph"),
				TypeQN: oneTypeSN,
			},
			ReceptionEPs: []procdec.ChnlSpec{
				{CommPH: sym.New("receiver-reception-ph"), TypeQN: lolliTypeSN},
				{CommPH: sym.New("message-reception-ph"), TypeQN: oneTypeSN},
			},
		}
		_, err = procDecAPI.Create(senderDecSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		receiverProcPH := sym.New("receiver-proc-ph")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: mainReceptionPH,
				BindPH: receiverProcPH,
				ProcSN: receiverProcSN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		messageProcPH := sym.New("message-proc-ph")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: mainReceptionPH,
				BindPH: messageProcPH,
				ProcSN: messageProcSN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		senderProcPH := sym.New("sender-proc-ph")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: mainReceptionPH,
				BindPH: senderProcPH,
				ProcSN: senderProcSN,
				ValPHs: []sym.ADT{receiverProcPH, messageProcPH},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		receiverExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.RecvSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// when
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: receiverExecRef.ExecID,
			ProcTS: procdef.RecvSpec{
				CommPH: receiverDecSpec.ProvisionEP.CommPH,
				BindPH: sym.New("message-reception-ph"),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		senderExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.SendSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: senderExecRef.ExecID,
			ProcTS: procdef.SendSpec{
				CommPH: senderDecSpec.ReceptionEPs[0].CommPH,
				ValPH:  senderDecSpec.ReceptionEPs[1].CommPH,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("CaseLab", func(t *testing.T) {
		tc.Setup(t)
		// given
		mainPoolSN := sym.New("main-pool-sn")
		mainExecRef, err := poolExecAPI.Create(poolexec.PoolSpec{
			PoolQN: mainPoolSN,
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		label := sym.ADT("label-1")
		// and
		withRoleSpec := typedef.TypeSpec{
			TypeSN: "with-role",
			TypeTS: typedef.WithSpec{
				Zs: map[sym.ADT]typedef.TermSpec{
					label: typedef.OneSpec{},
				},
			},
		}
		withRole, err := typeDefAPI.Create(withRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRoleSpec := typedef.TypeSpec{
			TypeSN: "one-role",
			TypeTS: typedef.OneSpec{},
		}
		oneRole, err := typeDefAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		withSigSpec := procdec.ProcSpec{
			ProcSN: "sig-1",
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-1",
				TypeQN: withRole.TypeQN,
			},
		}
		withSig, err := procDecAPI.Create(withSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec := procdec.ProcSpec{
			ProcSN:       "sig-2",
			ReceptionEPs: []procdec.ChnlSpec{withSig.X},
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-2",
				TypeQN: oneRole.TypeQN,
			},
		}
		_, err = procDecAPI.Create(oneSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		followerPH := sym.New("follower")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: followerPH,
				ProcSN: "tbd",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		deciderPH := sym.New("decider")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: deciderPH,
				ProcSN: "tbd",
				ValPHs: []sym.ADT{followerPH},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		followerExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.CaseSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// when
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: followerExecRef.ExecID,
			ProcTS: procdef.CaseSpec{
				CommPH: followerPH,
				Conts: map[sym.ADT]procdef.TermSpec{
					label: procdef.CloseSpec{
						CommPH: followerPH,
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		deciderExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.LabSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: deciderExecRef.ExecID,
			ProcTS: procdef.LabSpec{
				CommPH: followerPH,
				Label:  label,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("Spawn", func(t *testing.T) {
		tc.Setup(t)
		// given
		mainPoolSN := sym.New("main-pool-sn")
		mainExecRef, err := poolExecAPI.Create(poolexec.PoolSpec{
			PoolQN: mainPoolSN,
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRole, err := typeDefAPI.Create(
			typedef.TypeSpec{
				TypeSN: "one-role",
				TypeTS: typedef.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procDecAPI.Create(procdec.ProcSpec{
			ProcSN: "sig-1",
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-1",
				TypeQN: oneRole.TypeQN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procDecAPI.Create(procdec.ProcSpec{
			ProcSN:       "sig-2",
			ReceptionEPs: []procdec.ChnlSpec{oneSig1.X},
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-2",
				TypeQN: oneRole.TypeQN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig3, err := procDecAPI.Create(procdec.ProcSpec{
			ProcSN:       "sig-3",
			ReceptionEPs: []procdec.ChnlSpec{oneSig1.X},
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-3",
				TypeQN: oneRole.TypeQN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolExecAPI.Create(poolexec.PoolSpec{
			PoolQN: "pool-1",
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		injecteePH := sym.New("injectee")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: poolImpl.ExecID,
			ExecID: poolImpl.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: injecteePH,
				ProcSN: "tbd",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		spawnerPH := sym.New("spawner")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: poolImpl.ExecID,
			ExecID: poolImpl.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: spawnerPH,
				ProcSN: "tbd",
				ValPHs: []sym.ADT{injecteePH},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		x := sym.New("x")
		// and
		spawnerExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.CallSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// when
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: poolImpl.ExecID,
			ExecID: spawnerExecRef.ExecID,
			ProcTS: procdef.SpawnSpecOld{
				SigID: oneSig3.DecID,
				Ys:    []sym.ADT{injecteePH},
				X:     x,
				Cont: procdef.WaitSpec{
					CommPH: x,
					ContTS: procdef.CloseSpec{
						CommPH: spawnerPH,
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("Fwd", func(t *testing.T) {
		tc.Setup(t)
		// given
		mainPoolSN := sym.New("main-pool-sn")
		mainExecRef, err := poolExecAPI.Create(poolexec.PoolSpec{
			PoolQN: mainPoolSN,
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRole, err := typeDefAPI.Create(typedef.TypeSpec{
			TypeSN: "one-role",
			TypeTS: typedef.OneSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procDecAPI.Create(procdec.ProcSpec{
			ProcSN: "sig-1",
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-1",
				TypeQN: oneRole.TypeQN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procDecAPI.Create(procdec.ProcSpec{
			ProcSN:       "sig-2",
			ReceptionEPs: []procdec.ChnlSpec{oneSig1.X},
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-2",
				TypeQN: oneRole.TypeQN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procDecAPI.Create(procdec.ProcSpec{
			ProcSN:       "sig-3",
			ReceptionEPs: []procdec.ChnlSpec{oneSig1.X},
			ProvisionEP: procdec.ChnlSpec{
				CommPH: "chnl-3",
				TypeQN: oneRole.TypeQN,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerChnlPH := sym.New("closer")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: closerChnlPH,
				ProcSN: "tbd",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		forwarderChnlPH := sym.New("forwarder")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: forwarderChnlPH,
				ProcSN: "tbd",
				ValPHs: []sym.ADT{closerChnlPH},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterChnlPH := sym.New("waiter")
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: mainExecRef.ProcID,
			ProcTS: procdef.CallSpec{
				CommPH: waiterChnlPH,
				ProcSN: "tbd",
				ValPHs: []sym.ADT{forwarderChnlPH},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.CloseSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: closerExecRef.ExecID,
			ProcTS: procdef.CloseSpec{
				CommPH: closerChnlPH,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		forwarderExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.FwdSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// when
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: forwarderExecRef.ExecID,
			ProcTS: procdef.FwdSpec{
				X: forwarderChnlPH,
				Y: closerChnlPH,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterExecRef, err := poolExecAPI.Poll(poolexec.PollSpec{
			PoolID: mainExecRef.ExecID,
			PoolTS: pooldef.WaitSpec{},
		})
		if err != nil {
			t.Fatal(err)
		}
		// and
		err = procExecAPI.Run(procexec.ProcSpec{
			PoolID: mainExecRef.ExecID,
			ExecID: waiterExecRef.ExecID,
			ProcTS: procdef.WaitSpec{
				CommPH: forwarderChnlPH,
				ContTS: procdef.CloseSpec{
					CommPH: waiterChnlPH,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})
}
