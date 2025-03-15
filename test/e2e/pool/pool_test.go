package pool_test

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/ph"

	"smecalculus/rolevod/internal/chnl"
	"smecalculus/rolevod/internal/state"
	"smecalculus/rolevod/internal/step"

	"smecalculus/rolevod/app/pool"
	"smecalculus/rolevod/app/role"
	"smecalculus/rolevod/app/sig"
)

var (
	roleAPI = role.NewAPI()
	sigAPI  = sig.NewAPI()
	poolAPI = pool.NewAPI()
	tc      *testCase
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
		"pool_roots",
		"sig_roots", "sig_pes", "sig_ces",
		"role_roots", "role_states",
		"states", "channels", "steps", "clientships"}
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
		poolSpec1 := pool.Spec{Title: "ts1"}
		poolRoot1, err := poolAPI.Create(poolSpec1)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec2 := pool.Spec{Title: "ts2", SupID: poolRoot1.PoolID}
		poolRoot2, err := poolAPI.Create(poolSpec2)
		if err != nil {
			t.Fatal(err)
		}
		// when
		poolSnap1, err := poolAPI.Retrieve(poolRoot1.PoolID)
		if err != nil {
			t.Fatal(err)
		}
		// then
		extectedSub := pool.ConvertRootToRef(poolRoot2)
		if !slices.Contains(poolSnap1.Subs, extectedSub) {
			t.Errorf("unexpected subs in %q; want: %+v, got: %+v",
				poolSpec1.Title, extectedSub, poolSnap1.Subs)
		}
	})
}

func TestTaking(t *testing.T) {

	t.Run("WaitClose", func(t *testing.T) {
		tc.Setup(t)
		// given
		oneRoleSpec := role.Spec{
			QN:    "one-role",
			State: state.OneSpec{},
		}
		oneRole, err := roleAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerSigSpec := sig.Spec{
			QN: "closer",
			X: chnl.Spec{
				ChnlPH: "closing-1",
				RoleQN: oneRole.QN,
			},
		}
		closerSig, err := sigAPI.Create(closerSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterSigSpec := sig.Spec{
			QN: "waiter",
			X: chnl.Spec{
				ChnlPH: "closing-2",
				RoleQN: oneRole.QN,
			},
			Ys: []chnl.Spec{
				closerSig.X2,
			},
		}
		waiterSig, err := sigAPI.Create(waiterSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec := pool.Spec{
			Title: "big-deal",
		}
		poolImpl, err := poolAPI.Create(poolSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  closerSig.SigID,
		}
		closer, err := poolAPI.Involve(closerSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  waiterSig.SigID,
			ChnlIDs: []id.ADT{
				closer.ChnlID,
			},
		}
		waiter, err := poolAPI.Involve(waiterSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closeSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: closer.ProcID,
			Term: step.CloseSpec{
				X: closer.ChnlPH,
			},
		}
		// when
		err = poolAPI.Take(closeSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waitSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: waiter.ProcID,
			Term: step.WaitSpec{
				X: closer.ChnlPH,
				Cont: step.CloseSpec{
					X: waiter.ChnlPH,
				},
			},
		}
		// and
		err = poolAPI.Take(waitSpec)
		if err != nil {
			t.Fatal(err)
		}
		// then
		// TODO добавить проверку
	})

	t.Run("RecvSend", func(t *testing.T) {
		tc.Setup(t)
		// given
		lolliRoleSpec := role.Spec{
			QN: "lolli-role",
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
		oneRoleSpec := role.Spec{
			QN:    "one-role",
			State: state.OneSpec{},
		}
		oneRole, err := roleAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		lolliSigSpec := sig.Spec{
			QN: "sig-1",
			X: chnl.Spec{
				ChnlPH: "chnl-1",
				RoleQN: lolliRole.QN,
			},
		}
		lolliSig, err := sigAPI.Create(lolliSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec1 := sig.Spec{
			QN: "sig-2",
			X: chnl.Spec{
				ChnlPH: "chnl-2",
				RoleQN: oneRole.QN,
			},
		}
		oneSig1, err := sigAPI.Create(oneSigSpec1)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec2 := sig.Spec{
			QN: "sig-3",
			X: chnl.Spec{
				ChnlPH: "chnl-3",
				RoleQN: oneRole.QN,
			},
			Ys: []chnl.Spec{
				lolliSigSpec.X,
				oneSig1.X2,
			},
		}
		oneSig2, err := sigAPI.Create(oneSigSpec2)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec := pool.Spec{
			Title: "pool-1",
		}
		poolImpl, err := poolAPI.Create(poolSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		receiverSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  lolliSig.SigID,
		}
		receiver, err := poolAPI.Involve(receiverSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		messageSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  oneSig1.SigID,
		}
		message, err := poolAPI.Involve(messageSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		senderSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  oneSig2.SigID,
			ChnlIDs: []id.ADT{
				receiver.ChnlID,
				message.ChnlID,
			},
		}
		sender, err := poolAPI.Involve(senderSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		recvSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: receiver.ProcID,
			Term: step.RecvSpec{
				X: receiver.ChnlPH,
				Y: message.ChnlPH,
				Cont: step.WaitSpec{
					X: message.ChnlPH,
					Cont: step.CloseSpec{
						X: receiver.ChnlPH,
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
		sendSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: sender.ProcID,
			Term: step.SendSpec{
				X: receiver.ChnlPH,
				Y: message.ChnlPH,
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
		withRoleSpec := role.Spec{
			QN: "with-role",
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
		oneRoleSpec := role.Spec{
			QN:    "one-role",
			State: state.OneSpec{},
		}
		oneRole, err := roleAPI.Create(oneRoleSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		withSigSpec := sig.Spec{
			QN: "sig-1",
			X: chnl.Spec{
				ChnlPH: "chnl-1",
				RoleQN: withRole.QN,
			},
		}
		withSig, err := sigAPI.Create(withSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSigSpec := sig.Spec{
			QN: "sig-2",
			X: chnl.Spec{
				ChnlPH: "chnl-2",
				RoleQN: oneRole.QN,
			},
			Ys: []chnl.Spec{
				withSig.X2,
			},
		}
		oneSig, err := sigAPI.Create(oneSigSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolSpec := pool.Spec{
			Title: "pool-1",
		}
		poolImpl, err := poolAPI.Create(poolSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		followerSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  withSig.SigID,
		}
		follower, err := poolAPI.Involve(followerSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		deciderSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  oneSig.SigID,
			ChnlIDs: []id.ADT{
				follower.ChnlID,
			},
		}
		decider, err := poolAPI.Involve(deciderSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		caseSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: follower.ProcID,
			Term: step.CaseSpec{
				X: follower.ChnlPH,
				Conts: map[core.Label]step.Term{
					label: step.CloseSpec{
						X: follower.ChnlPH,
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
		labSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: decider.ProcID,
			Term: step.LabSpec{
				X: follower.ChnlPH,
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
			role.Spec{
				QN:    "one-role",
				State: state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := sigAPI.Create(
			sig.Spec{
				QN: "sig-1",
				X: chnl.Spec{
					ChnlPH: "chnl-1",
					RoleQN: oneRole.QN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig2, err := sigAPI.Create(
			sig.Spec{
				QN: "sig-2",
				Ys: []chnl.Spec{oneSig1.X2},
				X: chnl.Spec{
					ChnlPH: "chnl-2",
					RoleQN: oneRole.QN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig3, err := sigAPI.Create(
			sig.Spec{
				QN: "sig-3",
				Ys: []chnl.Spec{oneSig1.X2},
				X: chnl.Spec{
					ChnlPH: "chnl-3",
					RoleQN: oneRole.QN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			pool.Spec{
				Title: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		injectee, err := poolAPI.Involve(
			pool.PartSpec{
				PoolID: poolImpl.PoolID,
				SigID:  oneSig1.SigID,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		spawnerSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  oneSig2.SigID,
			ChnlIDs: []id.ADT{
				injectee.ChnlID,
			},
		}
		spawner, err := poolAPI.Involve(spawnerSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		x := ph.New("x")
		// and
		spawnSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: spawner.ProcID,
			Term: step.SpawnSpec{
				X: x,
				Ys: []ph.ADT{
					injectee.ChnlPH,
				},
				Cont: step.WaitSpec{
					X: x,
					Cont: step.CloseSpec{
						X: spawner.ChnlPH,
					},
				},
				SigID: oneSig3.SigID,
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
			role.Spec{
				QN:    "one-role",
				State: state.OneSpec{},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig1, err := sigAPI.Create(
			sig.Spec{
				QN: "sig-1",
				X: chnl.Spec{
					ChnlPH: "chnl-1",
					RoleQN: oneRole.QN,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig2, err := sigAPI.Create(
			sig.Spec{
				QN: "sig-2",
				X: chnl.Spec{
					ChnlPH: "chnl-2",
					RoleQN: oneRole.QN,
				},
				Ys: []chnl.Spec{
					oneSig1.X2,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		oneSig3, err := sigAPI.Create(
			sig.Spec{
				QN: "sig-3",
				X: chnl.Spec{
					ChnlPH: "chnl-3",
					RoleQN: oneRole.QN,
				},
				Ys: []chnl.Spec{
					oneSig1.X2,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		poolImpl, err := poolAPI.Create(
			pool.Spec{
				Title: "pool-1",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closerSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  oneSig1.SigID,
		}
		closer, err := poolAPI.Involve(closerSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		forwarderSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  oneSig2.SigID,
			ChnlIDs: []id.ADT{
				closer.ChnlID,
			},
		}
		forwarder, err := poolAPI.Involve(forwarderSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waiterSpec := pool.PartSpec{
			PoolID: poolImpl.PoolID,
			SigID:  oneSig3.SigID,
			ChnlIDs: []id.ADT{
				forwarder.ChnlID,
			},
		}
		waiter, err := poolAPI.Involve(waiterSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		closeSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: closer.ProcID,
			Term: step.CloseSpec{
				X: closer.ChnlPH,
			},
		}
		err = poolAPI.Take(closeSpec)
		if err != nil {
			t.Fatal(err)
		}
		// when
		fwdSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: forwarder.ProcID,
			Term: step.FwdSpec{
				X: forwarder.ChnlPH,
				Y: closer.ChnlPH,
			},
		}
		err = poolAPI.Take(fwdSpec)
		if err != nil {
			t.Fatal(err)
		}
		// and
		waitSpec := pool.TranSpec{
			PoolID: poolImpl.PoolID,
			ProcID: waiter.ProcID,
			Term: step.WaitSpec{
				X: forwarder.ChnlPH,
				Cont: step.CloseSpec{
					X: waiter.ChnlPH,
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
