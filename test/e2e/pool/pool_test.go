package pool_test

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/ph"
	"smecalculus/rolevod/lib/sym"

	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	poolbnd "smecalculus/rolevod/app/pool/bnd"
	poolroot "smecalculus/rolevod/app/pool/root"
	poolsig "smecalculus/rolevod/app/pool/sig"
	procbnd "smecalculus/rolevod/app/proc/bnd"
	procroot "smecalculus/rolevod/app/proc/root"
	procsig "smecalculus/rolevod/app/proc/sig"
	roleroot "smecalculus/rolevod/app/role/root"
)

var (
	roleAPI    = roleroot.NewAPI()
	procSigAPI = procsig.NewAPI()
	poolSigAPI = poolsig.NewAPI()
	poolAPI    = poolroot.NewAPI()
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
		oneRoleQN := sym.New("one")
		_, err := roleAPI.Create(
			roleroot.Spec{
				RoleQN: oneRoleQN,
				State:  state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerSigQN := sym.New("closer")
		_, err = procSigAPI.Create(
			procsig.Spec{
				X: procbnd.Spec{
					RoleQN: oneRoleQN,
				},
				SigQN: closerSigQN,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterSigQN := sym.New("waiter")
		_, err = procSigAPI.Create(
			procsig.Spec{
				X: procbnd.Spec{
					ChnlPH: ph.Blank,
					RoleQN: oneRoleQN,
				},
				SigQN: waiterSigQN,
				Ys: []procbnd.Spec{
					{RoleQN: oneRoleQN},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSigQN := sym.New("my-pool")
		_, err = poolSigAPI.Create(
			poolsig.Spec{
				SigQN: poolSigQN,
				Exports: []poolbnd.Spec{
					{ProcQN: closerSigQN},
					{ProcQN: waiterSigQN},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolRootImpl, err := poolAPI.Create(
			poolroot.Spec{
				SigQN: poolSigQN,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerMainPH := ph.New("closer")
		closerProcSpec := procroot.Spec{
			PoolID: poolRootImpl.PoolID,
			ProcID: poolRootImpl.ProcID,
			Term: step.CallSpec{
				X:     closerMainPH,
				SigPH: closerSigQN.ToPH(),
			},
		}
		closerProcRef, err := poolAPI.Spawn(closerProcSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterProcSpec := procroot.Spec{
			PoolID: poolRootImpl.PoolID,
			ProcID: poolRootImpl.ProcID,
			Term: step.CallSpec{
				X:     ph.Blank,
				SigPH: waiterSigQN.ToPH(),
				Ys:    []ph.ADT{closerMainPH},
			},
		}
		waiterProcRef, err := poolAPI.Spawn(waiterProcSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closeStepSpec := poolroot.StepSpec{
			PoolID: poolRootImpl.PoolID,
			ProcID: closerProcRef.ProcID,
			Term: step.CloseSpec{
				X: oneRoleQN.ToPH(),
			},
		}
		// when
		err = poolAPI.Take(closeStepSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waitStepSpec := poolroot.StepSpec{
			PoolID: poolRootImpl.PoolID,
			ProcID: waiterProcRef.ProcID,
			Term: step.WaitSpec{
				X: oneRoleQN.ToPH(),
				Cont: step.CloseSpec{
					X: ph.Blank,
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
		lolliRoleSpec := roleroot.Spec{
			RoleQN: "lolli-role",
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
		oneRoleSpec := roleroot.Spec{
			RoleQN: "one-role",
			State:  state.OneSpec{},
		}
		oneRole, err := roleAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		lolliSigSpec := procsig.Spec{
			SigQN: "sig-1",
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
			SigQN: "sig-2",
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
			SigQN: "sig-3",
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
		receiverChnlPH := ph.New("receiver")
		receiver, err := poolAPI.Spawn(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: step.CallSpec{
					X:     receiverChnlPH,
					SigPH: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		messageChnlPH := ph.New("message")
		_, err = poolAPI.Spawn(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: step.CallSpec{
					X:     messageChnlPH,
					SigPH: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		senderChnlPH := ph.New("sender")
		sender, err := poolAPI.Spawn(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: step.CallSpec{
					X:     senderChnlPH,
					SigPH: "tbd",
					Ys:    []ph.ADT{receiverChnlPH, senderChnlPH},
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
		label := core.Label("label-1")
		// and
		withRoleSpec := roleroot.Spec{
			RoleQN: "with-role",
			State: state.WithSpec{
				Choices: map[core.Label]state.Spec{
					label: state.OneSpec{},
				},
			},
		}
		withRole, err := roleAPI.Create(withRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneRoleSpec := roleroot.Spec{
			RoleQN: "one-role",
			State:  state.OneSpec{},
		}
		oneRole, err := roleAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		withSigSpec := procsig.Spec{
			SigQN: "sig-1",
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
			SigQN: "sig-2",
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
		followerPH := ph.New("follower")
		follower, err := poolAPI.Spawn(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: step.CallSpec{
					X:     followerPH,
					SigPH: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		deciderPH := ph.New("decider")
		decider, err := poolAPI.Spawn(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: step.CallSpec{
					Ys:    []ph.ADT{followerPH},
					X:     deciderPH,
					SigPH: "tbd",
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
				Conts: map[core.Label]step.Term{
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
				X: followerPH,
				L: label,
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
			roleroot.Spec{
				RoleQN: "one-role",
				State:  state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procSigAPI.Create(
			procsig.Spec{
				SigQN: "sig-1",
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
				SigQN: "sig-2",
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
				SigQN: "sig-3",
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
		injecteePH := ph.New("injectee")
		_, err = poolAPI.Spawn(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: step.CallSpec{
					X:     injecteePH,
					SigPH: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		spawnerPH := ph.New("spawner")
		spawner, err := poolAPI.Spawn(
			procroot.Spec{
				PoolID: poolImpl.PoolID,
				ProcID: poolImpl.ProcID,
				Term: step.CallSpec{
					Ys:    []ph.ADT{injecteePH},
					X:     spawnerPH,
					SigPH: "tbd",
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		x := ph.New("x")
		spawnSpec := poolroot.StepSpec{
			PoolID: poolImpl.PoolID,
			ProcID: spawner.ProcID,
			Term: step.SpawnSpec{
				SigID: oneSig3.SigID,
				Ys:    []ph.ADT{injecteePH},
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
			roleroot.Spec{
				RoleQN: "one-role",
				State:  state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := procSigAPI.Create(
			procsig.Spec{
				SigQN: "sig-1",
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
				SigQN: "sig-2",
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
				SigQN: "sig-3",
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
		closerChnlPH := ph.New("closer")
		closerSpec := procroot.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: step.CallSpec{
				X:     closerChnlPH,
				SigPH: "tbd",
			},
		}
		closer, err := poolAPI.Spawn(closerSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		forwarderChnlPH := ph.New("forwarder")
		forwarderSpec := procroot.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: step.CallSpec{
				X:     forwarderChnlPH,
				SigPH: "tbd",
				Ys:    []ph.ADT{closerChnlPH},
			},
		}
		forwarder, err := poolAPI.Spawn(forwarderSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterChnlPH := ph.New("waiter")
		waiterSpec := procroot.Spec{
			PoolID: poolImpl.PoolID,
			ProcID: poolImpl.ProcID,
			Term: step.CallSpec{
				X:     waiterChnlPH,
				SigPH: "tbd",
				Ys:    []ph.ADT{forwarderChnlPH},
			},
		}
		waiter, err := poolAPI.Spawn(waiterSpec)
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
