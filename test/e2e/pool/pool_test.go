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
	poolxact "smecalculus/rolevod/app/pool/xact"
	procdec "smecalculus/rolevod/app/proc/dec"
	procdef "smecalculus/rolevod/app/proc/def"
	proceval "smecalculus/rolevod/app/proc/eval"
	typedef "smecalculus/rolevod/app/type/def"
)

var (
	typeAPI    = typedef.NewAPI()
	procSigAPI = procdec.NewAPI()
	poolSigAPI = pooldec.NewAPI()
	poolAPI    = pooldef.NewAPI()
	procAPI    = proceval.NewAPI()
	tc         *testCase
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
		poolSpec1 := pooldef.PoolSpec{SigQN: "ts1"}
		poolRef1, err := poolAPI.Create(poolSpec1)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec2 := pooldef.PoolSpec{SigQN: "ts2", SupID: poolRef1.PoolID}
		poolRef2, err := poolAPI.Create(poolSpec2)
		if err != nil {
			t.Fatal(err)
		}
		// when
		poolSnap1, err := poolAPI.Retrieve(poolRef1.PoolID)
		if err != nil {
			t.Fatal(err)
		}
		// then
		if !slices.Contains(poolSnap1.Subs, poolRef2) {
			t.Errorf("unexpected subs in %q; want: %+v, got: %+v",
				poolSpec1.SigQN, poolRef2, poolSnap1.Subs)
		}
	})
}

func TestTaking(t *testing.T) {

	t.Run("WaitClose", func(t *testing.T) {
		tc.Setup(t)
		// given
		mainRoleSN := sym.New("main-role-sn")
		closerSigSN := sym.New("closer-sn")
		waiterSigSN := sym.New("waiter-sn")
		_, err := typeAPI.Create(
			typedef.TypeSpec{
				TypeSN: mainRoleSN,
				TypeTS: typedef.UpSpec{
					X: typedef.WithSpec{
						Choices: map[sym.ADT]typedef.TermSpec{
							closerSigSN: typedef.DownSpec{
								X: typedef.LinkSpec{
									TypeQN: mainRoleSN,
								},
							},
							waiterSigSN: typedef.DownSpec{
								X: typedef.LinkSpec{
									TypeQN: mainRoleSN,
								},
							},
						},
					},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRoleSN := sym.New("one-role")
		_, err = typeAPI.Create(
			typedef.TypeSpec{
				TypeSN: oneRoleSN,
				TypeTS: typedef.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerSigSpec := procdec.SigSpec{
			X: procdec.ChnlSpec{
				ChnlPH: sym.New("x"),
				TypeQN: oneRoleSN,
			},
			SigSN: closerSigSN,
		}
		_, err = procSigAPI.Create(closerSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterSigSpec := procdec.SigSpec{
			X:     procdec.ChnlSpec{ChnlPH: sym.Blank, TypeQN: oneRoleSN},
			SigSN: waiterSigSN,
			Ys: []procdec.ChnlSpec{
				{ChnlPH: sym.New("y"), TypeQN: oneRoleSN},
			},
		}
		_, err = procSigAPI.Create(waiterSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		mainSigSN := sym.New("main-sig")
		mainConnPH := sym.New("main-conn")
		_, err = poolSigAPI.Create(
			pooldec.SigSpec{
				SigSN: mainSigSN,
				X: pooldec.BndSpec{
					ChnlPH: mainConnPH,
					TypeQN: mainRoleSN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolRef, err := poolAPI.Create(
			pooldef.PoolSpec{
				SigQN: mainSigSN,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerChnlPH := sym.New("closer")
		closerProcSpec := proceval.Spec{
			PoolID: poolRef.PoolID,
			ProcID: poolRef.ProcID,
			Term: poolxact.CallSpec{
				MainPH: mainConnPH,
				X:      closerChnlPH,
				SigSN:  closerSigSN,
			},
		}
		closerProcRef, err := procAPI.Create(closerProcSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterProcSpec := proceval.Spec{
			PoolID: poolRef.PoolID,
			ProcID: poolRef.ProcID,
			Term: poolxact.CallSpec{
				MainPH: mainConnPH,
				X:      sym.Blank,
				SigSN:  waiterSigSN,
				Ys:     []sym.ADT{closerChnlPH},
			},
		}
		waiterProcRef, err := procAPI.Create(waiterProcSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closeStepSpec := pooldef.StepSpec{
			PoolID: poolRef.PoolID,
			ProcID: closerProcRef.ProcID,
			Term: procdef.CloseSpec{
				X: closerSigSpec.X.ChnlPH,
			},
		}
		// when
		err = poolAPI.Take(closeStepSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waitStepSpec := pooldef.StepSpec{
			PoolID: poolRef.PoolID,
			ProcID: waiterProcRef.ProcID,
			Term: procdef.WaitSpec{
				X: waiterSigSpec.Ys[0].ChnlPH,
				Cont: procdef.CloseSpec{
					X: sym.Blank,
				},
			},
		}
		// and
		err = poolAPI.Take(waitStepSpec)
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("RecvSend", func(t *testing.T) {
		tc.Setup(t)
		// given
		lolliRoleSpec := typedef.TypeSpec{
			TypeSN: "lolli-role",
			TypeTS: typedef.LolliSpec{
				Y: typedef.OneSpec{},
				Z: typedef.OneSpec{},
			},
		}
		lolliRole, err := typeAPI.Create(lolliRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRoleSpec := typedef.TypeSpec{
			TypeSN: "one-role",
			TypeTS: typedef.OneSpec{},
		}
		oneRole, err := typeAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		lolliSigSpec := procdec.SigSpec{
			SigSN: "sig-1",
			X: procdec.ChnlSpec{
				ChnlPH: "chnl-1",
				TypeQN: lolliRole.TypeQN,
			},
		}
		_, err = procSigAPI.Create(lolliSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec1 := procdec.SigSpec{
			SigSN: "sig-2",
			X: procdec.ChnlSpec{
				ChnlPH: "chnl-2",
				TypeQN: oneRole.TypeQN,
			},
		}
		oneSig1, err := procSigAPI.Create(oneSigSpec1)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec2 := procdec.SigSpec{
			SigSN: "sig-3",
			Ys:    []procdec.ChnlSpec{lolliSigSpec.X, oneSig1.X},
			X: procdec.ChnlSpec{
				ChnlPH: "chnl-3",
				TypeQN: oneRole.TypeQN,
			},
		}
		_, err = procSigAPI.Create(oneSigSpec2)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			pooldef.PoolSpec{
				SigQN: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		receiverChnlPH := sym.New("receiver")
		receiver, err := procAPI.Create(
			proceval.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: poolxact.CallSpec{
					X:     receiverChnlPH,
					SigSN: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		messageChnlPH := sym.New("message")
		_, err = procAPI.Create(
			proceval.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: poolxact.CallSpec{
					X:     messageChnlPH,
					SigSN: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		senderChnlPH := sym.New("sender")
		sender, err := procAPI.Create(
			proceval.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: poolxact.CallSpec{
					X:     senderChnlPH,
					SigSN: "tbd",
					Ys:    []sym.ADT{receiverChnlPH, senderChnlPH},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		recvSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: receiver.ProcID,
			Term: procdef.RecvSpec{
				X: receiverChnlPH,
				Y: messageChnlPH,
				Cont: procdef.WaitSpec{
					X: messageChnlPH,
					Cont: procdef.CloseSpec{
						X: receiverChnlPH,
					},
				},
			},
		}
		// when
		err = poolAPI.Take(recvSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		sendSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: sender.ProcID,
			Term: procdef.SendSpec{
				X: receiverChnlPH,
				Y: messageChnlPH,
			},
		}
		// and
		err = poolAPI.Take(sendSpec)
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("CaseLab", func(t *testing.T) {
		tc.Setup(t)
		// given
		label := sym.ADT("label-1")
		// and
		withRoleSpec := typedef.TypeSpec{
			TypeSN: "with-role",
			TypeTS: typedef.WithSpec{
				Choices: map[sym.ADT]typedef.TermSpec{
					label: typedef.OneSpec{},
				},
			},
		}
		withRole, err := typeAPI.Create(withRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRoleSpec := typedef.TypeSpec{
			TypeSN: "one-role",
			TypeTS: typedef.OneSpec{},
		}
		oneRole, err := typeAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		withSigSpec := procdec.SigSpec{
			SigSN: "sig-1",
			X: procdec.ChnlSpec{
				ChnlPH: "chnl-1",
				TypeQN: withRole.TypeQN,
			},
		}
		withSig, err := procSigAPI.Create(withSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec := procdec.SigSpec{
			SigSN: "sig-2",
			Ys:    []procdec.ChnlSpec{withSig.X},
			X: procdec.ChnlSpec{
				ChnlPH: "chnl-2",
				TypeQN: oneRole.TypeQN,
			},
		}
		_, err = procSigAPI.Create(oneSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec := pooldef.PoolSpec{
			SigQN: "pool-1",
		}
		poolImpl, err := poolAPI.Create(poolSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		followerPH := sym.New("follower")
		follower, err := procAPI.Create(
			proceval.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: poolxact.CallSpec{
					X:     followerPH,
					SigSN: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		deciderPH := sym.New("decider")
		decider, err := procAPI.Create(
			proceval.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: poolxact.CallSpec{
					X:     deciderPH,
					SigSN: "tbd",
					Ys:    []sym.ADT{followerPH},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		caseSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: follower.ProcID,
			Term: procdef.CaseSpec{
				X: followerPH,
				Conts: map[sym.ADT]procdef.TermSpec{
					label: procdef.CloseSpec{
						X: followerPH,
					},
				},
			},
		}
		// when
		err = poolAPI.Take(caseSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		labSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: decider.ProcID,
			Term: procdef.LabSpec{
				X:     followerPH,
				Label: label,
			},
		}
		// and
		err = poolAPI.Take(labSpec)
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("Spawn", func(t *testing.T) {
		tc.Setup(t)
		// given
		oneRole, err := typeAPI.Create(
			typedef.TypeSpec{
				TypeSN: "one-role",
				TypeTS: typedef.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procSigAPI.Create(
			procdec.SigSpec{
				SigSN: "sig-1",
				X: procdec.ChnlSpec{
					ChnlPH: "chnl-1",
					TypeQN: oneRole.TypeQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procSigAPI.Create(
			procdec.SigSpec{
				SigSN: "sig-2",
				Ys:    []procdec.ChnlSpec{oneSig1.X},
				X: procdec.ChnlSpec{
					ChnlPH: "chnl-2",
					TypeQN: oneRole.TypeQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig3, err := procSigAPI.Create(
			procdec.SigSpec{
				SigSN: "sig-3",
				Ys:    []procdec.ChnlSpec{oneSig1.X},
				X: procdec.ChnlSpec{
					ChnlPH: "chnl-3",
					TypeQN: oneRole.TypeQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			pooldef.PoolSpec{
				SigQN: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		injecteePH := sym.New("injectee")
		_, err = procAPI.Create(
			proceval.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: poolxact.CallSpec{
					X:     injecteePH,
					SigSN: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		spawnerPH := sym.New("spawner")
		spawner, err := procAPI.Create(
			proceval.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: poolxact.CallSpec{
					X:     spawnerPH,
					SigSN: "tbd",
					Ys:    []sym.ADT{injecteePH},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		x := sym.New("x")
		spawnSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: spawner.ProcID,
			Term: procdef.SpawnSpec{
				SigID: oneSig3.SigID,
				Ys:    []sym.ADT{injecteePH},
				X:     x,
				Cont: procdef.WaitSpec{
					X: x,
					Cont: procdef.CloseSpec{
						X: spawnerPH,
					},
				},
			},
		}
		// when
		err = poolAPI.Take(spawnSpec)
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("Fwd", func(t *testing.T) {
		tc.Setup(t)
		// given
		oneRole, err := typeAPI.Create(
			typedef.TypeSpec{
				TypeSN: "one-role",
				TypeTS: typedef.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procSigAPI.Create(
			procdec.SigSpec{
				SigSN: "sig-1",
				X: procdec.ChnlSpec{
					ChnlPH: "chnl-1",
					TypeQN: oneRole.TypeQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procSigAPI.Create(
			procdec.SigSpec{
				SigSN: "sig-2",
				Ys:    []procdec.ChnlSpec{oneSig1.X},
				X: procdec.ChnlSpec{
					ChnlPH: "chnl-2",
					TypeQN: oneRole.TypeQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procSigAPI.Create(
			procdec.SigSpec{
				SigSN: "sig-3",
				Ys:    []procdec.ChnlSpec{oneSig1.X},
				X: procdec.ChnlSpec{
					ChnlPH: "chnl-3",
					TypeQN: oneRole.TypeQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			pooldef.PoolSpec{
				SigQN: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerChnlPH := sym.New("closer")
		closerSpec := proceval.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: poolxact.CallSpec{
				X:     closerChnlPH,
				SigSN: "tbd",
			},
		}
		closer, err := procAPI.Create(closerSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		forwarderChnlPH := sym.New("forwarder")
		forwarderSpec := proceval.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: poolxact.CallSpec{
				X:     forwarderChnlPH,
				SigSN: "tbd",
				Ys:    []sym.ADT{closerChnlPH},
			},
		}
		forwarder, err := procAPI.Create(forwarderSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterChnlPH := sym.New("waiter")
		waiterSpec := proceval.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: poolxact.CallSpec{
				X:     waiterChnlPH,
				SigSN: "tbd",
				Ys:    []sym.ADT{forwarderChnlPH},
			},
		}
		waiter, err := procAPI.Create(waiterSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closeSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: closer.ProcID,
			Term: procdef.CloseSpec{
				X: closerChnlPH,
			},
		}
		err = poolAPI.Take(closeSpec)
		if err != nil {
			t.Fatal(err)
		}
		// when
		fwdSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: forwarder.ProcID,
			Term: procdef.FwdSpec{
				X: forwarderChnlPH,
				Y: closerChnlPH,
			},
		}
		err = poolAPI.Take(fwdSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waitSpec := pooldef.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: waiter.ProcID,
			Term: procdef.WaitSpec{
				X: forwarderChnlPH,
				Cont: procdef.CloseSpec{
					X: waiterChnlPH,
				},
			},
		}
		err = poolAPI.Take(waitSpec)
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})
}
