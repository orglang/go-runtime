package pool_test

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	poolbnd "smecalculus/rolevod/app/pool/bnd"
	poolroot "smecalculus/rolevod/app/pool/root"
	poolsig "smecalculus/rolevod/app/pool/sig"
	procbnd "smecalculus/rolevod/app/proc/bnd"
	procroot "smecalculus/rolevod/app/proc/root"
	procsig "smecalculus/rolevod/app/proc/sig"
	"smecalculus/rolevod/app/proc/xact"
	rolesig "smecalculus/rolevod/app/role/sig"
)

var (
	roleAPI    = rolesig.NewAPI()
	procSigAPI = procsig.NewAPI()
	poolSigAPI = poolsig.NewAPI()
	poolAPI    = poolroot.NewAPI()
	procAPI    = procroot.NewAPI()
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
		poolSpec1 := poolroot.Spec{SigQN: "ts1"}
		poolRef1, err := poolAPI.Create(poolSpec1)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec2 := poolroot.Spec{SigQN: "ts2", SupID: poolRef1.PoolID}
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
		mainRoleSN := sym.New("main-role")
		closerSigSN := sym.New("closer-role")
		waiterSigSN := sym.New("waiter-role")
		_, err := roleAPI.Create(
			rolesig.Spec{
				RoleSN: mainRoleSN,
				State: state.UpSpec{
					X: state.WithSpec{
						Choices: map[sym.ADT]state.Spec{
							closerSigSN: state.DownSpec{
								X: state.LinkSpec{
									RoleQN: mainRoleSN,
								},
							},
							waiterSigSN: state.DownSpec{
								X: state.LinkSpec{
									RoleQN: mainRoleSN,
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
		oneRoleSN := sym.New("one")
		_, err = roleAPI.Create(
			rolesig.Spec{
				RoleSN: oneRoleSN,
				State:  state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerSigSpec := procsig.Spec{
			X: procbnd.Spec{
				ChnlPH: sym.New("x"),
				RoleQN: oneRoleSN,
			},
			SigSN: closerSigSN,
		}
		_, err = procSigAPI.Create(closerSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterSigSpec := procsig.Spec{
			X:     procbnd.Spec{ChnlPH: sym.Blank, RoleQN: oneRoleSN},
			SigSN: waiterSigSN,
			Ys: []procbnd.Spec{
				{ChnlPH: sym.New("y"), RoleQN: oneRoleSN},
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
			poolsig.Spec{
				SigSN: mainSigSN,
				X: poolbnd.Spec{
					ConnPH: mainConnPH,
					RoleQN: mainRoleSN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolRef, err := poolAPI.Create(
			poolroot.Spec{
				SigQN: mainSigSN,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerChnlPH := sym.New("closer")
		closerProcSpec := procroot.Spec{
			PoolID: poolRef.PoolID,
			ProcID: poolRef.ProcID,
			Term: xact.CallSpec{
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
		waiterProcSpec := procroot.Spec{
			PoolID: poolRef.PoolID,
			ProcID: poolRef.ProcID,
			Term: xact.CallSpec{
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
		closeStepSpec := poolroot.StepSpec{
			PoolID: poolRef.PoolID,
			ProcID: closerProcRef.ProcID,
			Term: step.CloseSpec{
				X: closerSigSpec.X.ChnlPH,
			},
		}
		// when
		err = poolAPI.Take(closeStepSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waitStepSpec := poolroot.StepSpec{
			PoolID: poolRef.PoolID,
			ProcID: waiterProcRef.ProcID,
			Term: step.WaitSpec{
				X: waiterSigSpec.Ys[0].ChnlPH,
				Cont: step.CloseSpec{
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
		lolliRoleSpec := rolesig.Spec{
			RoleSN: "lolli-role",
			State: state.LolliSpec{
				Y: state.OneSpec{},
				Z: state.OneSpec{},
			},
		}
		lolliRole, err := roleAPI.Create(lolliRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRoleSpec := rolesig.Spec{
			RoleSN: "one-role",
			State:  state.OneSpec{},
		}
		oneRole, err := roleAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		lolliSigSpec := procsig.Spec{
			SigSN: "sig-1",
			X: procbnd.Spec{
				ChnlPH: "chnl-1",
				RoleQN: lolliRole.RoleQN,
			},
		}
		_, err = procSigAPI.Create(lolliSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec1 := procsig.Spec{
			SigSN: "sig-2",
			X: procbnd.Spec{
				ChnlPH: "chnl-2",
				RoleQN: oneRole.RoleQN,
			},
		}
		oneSig1, err := procSigAPI.Create(oneSigSpec1)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec2 := procsig.Spec{
			SigSN: "sig-3",
			Ys:    []procbnd.Spec{lolliSigSpec.X, oneSig1.X},
			X: procbnd.Spec{
				ChnlPH: "chnl-3",
				RoleQN: oneRole.RoleQN,
			},
		}
		_, err = procSigAPI.Create(oneSigSpec2)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			poolroot.Spec{
				SigQN: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		receiverChnlPH := sym.New("receiver")
		receiver, err := procAPI.Create(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: xact.CallSpec{
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
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: xact.CallSpec{
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
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: xact.CallSpec{
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
		recvSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: receiver.ProcID,
			Term: step.RecvSpec{
				X: receiverChnlPH,
				Y: messageChnlPH,
				Cont: step.WaitSpec{
					X: messageChnlPH,
					Cont: step.CloseSpec{
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
		sendSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: sender.ProcID,
			Term: step.SendSpec{
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
		withRoleSpec := rolesig.Spec{
			RoleSN: "with-role",
			State: state.WithSpec{
				Choices: map[sym.ADT]state.Spec{
					label: state.OneSpec{},
				},
			},
		}
		withRole, err := roleAPI.Create(withRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRoleSpec := rolesig.Spec{
			RoleSN: "one-role",
			State:  state.OneSpec{},
		}
		oneRole, err := roleAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		withSigSpec := procsig.Spec{
			SigSN: "sig-1",
			X: procbnd.Spec{
				ChnlPH: "chnl-1",
				RoleQN: withRole.RoleQN,
			},
		}
		withSig, err := procSigAPI.Create(withSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec := procsig.Spec{
			SigSN: "sig-2",
			Ys:    []procbnd.Spec{withSig.X},
			X: procbnd.Spec{
				ChnlPH: "chnl-2",
				RoleQN: oneRole.RoleQN,
			},
		}
		_, err = procSigAPI.Create(oneSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec := poolroot.Spec{
			SigQN: "pool-1",
		}
		poolImpl, err := poolAPI.Create(poolSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		followerPH := sym.New("follower")
		follower, err := procAPI.Create(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: xact.CallSpec{
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
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: xact.CallSpec{
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
		caseSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: follower.ProcID,
			Term: step.CaseSpec{
				X: followerPH,
				Conts: map[sym.ADT]step.Term{
					label: step.CloseSpec{
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
		labSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: decider.ProcID,
			Term: step.LabSpec{
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
		oneRole, err := roleAPI.Create(
			rolesig.Spec{
				RoleSN: "one-role",
				State:  state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procSigAPI.Create(
			procsig.Spec{
				SigSN: "sig-1",
				X: procbnd.Spec{
					ChnlPH: "chnl-1",
					RoleQN: oneRole.RoleQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procSigAPI.Create(
			procsig.Spec{
				SigSN: "sig-2",
				Ys:    []procbnd.Spec{oneSig1.X},
				X: procbnd.Spec{
					ChnlPH: "chnl-2",
					RoleQN: oneRole.RoleQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig3, err := procSigAPI.Create(
			procsig.Spec{
				SigSN: "sig-3",
				Ys:    []procbnd.Spec{oneSig1.X},
				X: procbnd.Spec{
					ChnlPH: "chnl-3",
					RoleQN: oneRole.RoleQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			poolroot.Spec{
				SigQN: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		injecteePH := sym.New("injectee")
		_, err = procAPI.Create(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: xact.CallSpec{
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
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: xact.CallSpec{
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
		spawnSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: spawner.ProcID,
			Term: step.SpawnSpec{
				SigID: oneSig3.SigID,
				Ys:    []sym.ADT{injecteePH},
				X:     x,
				Cont: step.WaitSpec{
					X: x,
					Cont: step.CloseSpec{
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
		oneRole, err := roleAPI.Create(
			rolesig.Spec{
				RoleSN: "one-role",
				State:  state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procSigAPI.Create(
			procsig.Spec{
				SigSN: "sig-1",
				X: procbnd.Spec{
					ChnlPH: "chnl-1",
					RoleQN: oneRole.RoleQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procSigAPI.Create(
			procsig.Spec{
				SigSN: "sig-2",
				Ys:    []procbnd.Spec{oneSig1.X},
				X: procbnd.Spec{
					ChnlPH: "chnl-2",
					RoleQN: oneRole.RoleQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		_, err = procSigAPI.Create(
			procsig.Spec{
				SigSN: "sig-3",
				Ys:    []procbnd.Spec{oneSig1.X},
				X: procbnd.Spec{
					ChnlPH: "chnl-3",
					RoleQN: oneRole.RoleQN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			poolroot.Spec{
				SigQN: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerChnlPH := sym.New("closer")
		closerSpec := procroot.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: xact.CallSpec{
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
		forwarderSpec := procroot.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: xact.CallSpec{
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
		waiterSpec := procroot.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: xact.CallSpec{
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
		closeSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: closer.ProcID,
			Term: step.CloseSpec{
				X: closerChnlPH,
			},
		}
		err = poolAPI.Take(closeSpec)
		if err != nil {
			t.Fatal(err)
		}
		// when
		fwdSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: forwarder.ProcID,
			Term: step.FwdSpec{
				X: forwarderChnlPH,
				Y: closerChnlPH,
			},
		}
		err = poolAPI.Take(fwdSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waitSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: waiter.ProcID,
			Term: step.WaitSpec{
				X: forwarderChnlPH,
				Cont: step.CloseSpec{
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
