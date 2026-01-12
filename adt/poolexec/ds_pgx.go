package poolexec

import (
	"errors"
	"log/slog"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procexec"
	"orglang/orglang/adt/procstep"
	"orglang/orglang/adt/revnum"
	"orglang/orglang/adt/termctx"
)

// Adapter
type pgxDAO struct {
	log *slog.Logger
}

func newPgxDAO(l *slog.Logger) *pgxDAO {
	name := slog.String("name", reflect.TypeFor[pgxDAO]().Name())
	return &pgxDAO{l.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &pgxDAO{}
}

func (dao *pgxDAO) Insert(source db.Source, root ExecRec) (err error) {
	ds := db.MustConform[db.SourcePgx](source)
	dto := DataFromExecRec(root)
	args := pgx.NamedArgs{
		"pool_id":     dto.PoolID,
		"proc_id":     dto.ProcID,
		"sup_pool_id": dto.SupID,
		"rev":         dto.PoolRN,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertRoot, args)
	if err != nil {
		dao.log.Error("execution failed")
		return err
	}
	dao.log.Debug("insertion succeed", slog.Any("poolID", root.ExecID))
	return nil
}

func (dao *pgxDAO) InsertLiab(source db.Source, liab procexec.Liab) (err error) {
	ds := db.MustConform[db.SourcePgx](source)
	dto := DataFromLiab(liab)
	args := pgx.NamedArgs{
		"pool_id": dto.PoolID,
		"proc_id": dto.ProcID,
		"rev":     dto.PoolRN,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertLiab, args)
	if err != nil {
		dao.log.Error("execution failed")
		return err
	}
	dao.log.Debug("insertion succeed", slog.Any("poolID", liab.PoolID))
	return nil
}

func (dao *pgxDAO) SelectProc(source db.Source, procID identity.ADT) (procexec.Cfg, error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("procID", procID)
	chnlRows, err := ds.Conn.Query(ds.Ctx, selectChnls, procID.String())
	if err != nil {
		dao.log.Error("execution failed", idAttr)
		return procexec.Cfg{}, err
	}
	defer chnlRows.Close()
	chnlDtos, err := pgx.CollectRows(chnlRows, pgx.RowToStructByName[epDS])
	if err != nil {
		dao.log.Error("collection failed", idAttr, slog.Any("t", reflect.TypeOf(chnlDtos)))
		return procexec.Cfg{}, err
	}
	chnls, err := DataToEPs(chnlDtos)
	if err != nil {
		dao.log.Error("conversion failed", idAttr)
		return procexec.Cfg{}, err
	}
	stepRows, err := ds.Conn.Query(ds.Ctx, selectSteps, procID.String())
	if err != nil {
		dao.log.Error("execution failed", idAttr)
		return procexec.Cfg{}, err
	}
	defer stepRows.Close()
	stepDtos, err := pgx.CollectRows(stepRows, pgx.RowToStructByName[procstep.StepRecDS])
	if err != nil {
		dao.log.Error("collection failed", idAttr, slog.Any("t", reflect.TypeOf(stepDtos)))
		return procexec.Cfg{}, err
	}
	steps, err := procstep.DataToSemRecs(stepDtos)
	if err != nil {
		dao.log.Error("conversion failed", idAttr)
		return procexec.Cfg{}, err
	}
	dao.log.Debug("selection succeed", idAttr)
	return procexec.Cfg{
		Chnls: termctx.IndexBy(procexec.ChnlPH, chnls),
		Steps: termctx.IndexBy(procstep.ChnlID, steps),
	}, nil
}

func (dao *pgxDAO) UpdateProc(source db.Source, mod procexec.Mod) (err error) {
	if len(mod.Locks) == 0 {
		panic("empty locks")
	}
	ds := db.MustConform[db.SourcePgx](source)
	dto, err := procexec.DataFromMod(mod)
	if err != nil {
		dao.log.Error("conversion failed")
		return err
	}
	// bindings
	bndReq := pgx.Batch{}
	for _, dto := range dto.Binds {
		args := pgx.NamedArgs{
			"proc_id":  dto.ExecID,
			"chnl_ph":  dto.ChnlPH,
			"chnl_id":  dto.ChnlID,
			"state_id": dto.StateID,
			"rev":      dto.PoolRN,
		}
		bndReq.Queue(insertBnd, args)
	}
	if bndReq.Len() > 0 {
		bndRes := ds.Conn.SendBatch(ds.Ctx, &bndReq)
		defer func() {
			err = errors.Join(err, bndRes.Close())
		}()
		for _, dto := range dto.Binds {
			_, err = bndRes.Exec()
			if err != nil {
				dao.log.Error("execution failed", slog.Any("dto", dto))
			}
		}
		if err != nil {
			return err
		}
	}
	// steps
	stepReq := pgx.Batch{}
	for _, dto := range dto.Steps {
		args := pgx.NamedArgs{
			"proc_id": dto.PID,
			"chnl_id": dto.VID,
			"kind":    dto.K,
			"spec":    dto.ProcER,
		}
		stepReq.Queue(insertStep, args)
	}
	if stepReq.Len() > 0 {
		stepRes := ds.Conn.SendBatch(ds.Ctx, &stepReq)
		defer func() {
			err = errors.Join(err, stepRes.Close())
		}()
		for _, dto := range dto.Steps {
			_, err = stepRes.Exec()
			if err != nil {
				dao.log.Error("execution failed", slog.Any("dto", dto))
			}
		}
		if err != nil {
			return err
		}
	}
	// roots
	rootReq := pgx.Batch{}
	for _, dto := range dto.Locks {
		args := pgx.NamedArgs{
			"pool_id": dto.PoolID,
			"rev":     dto.PoolRN,
		}
		rootReq.Queue(updateRoot, args)
	}
	rootRes := ds.Conn.SendBatch(ds.Ctx, &rootReq)
	defer func() {
		err = errors.Join(err, rootRes.Close())
	}()
	for _, dto := range dto.Locks {
		ct, err := rootRes.Exec()
		if err != nil {
			dao.log.Error("execution failed", slog.Any("dto", dto))
		}
		if ct.RowsAffected() == 0 {
			dao.log.Error("update failed")
			return errOptimisticUpdate(revnum.ADT(dto.PoolRN))
		}
	}
	if err != nil {
		return err
	}
	dao.log.Debug("update succeed")
	return nil
}

func (dao *pgxDAO) SelectSubs(source db.Source, poolID identity.ADT) (ExecSnap, error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("poolID", poolID)
	rows, err := ds.Conn.Query(ds.Ctx, selectOrgSnap, poolID.String())
	if err != nil {
		dao.log.Error("execution failed", idAttr)
		return ExecSnap{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[execSnapDS])
	if err != nil {
		dao.log.Error("collection failed", idAttr, slog.Any("struct", reflect.TypeOf(dto)))
		return ExecSnap{}, err
	}
	snap, err := DataToExecSnap(dto)
	if err != nil {
		dao.log.Error("conversion failed")
		return ExecSnap{}, err
	}
	dao.log.Debug("selection succeed", idAttr)
	return snap, nil
}

func (dao *pgxDAO) SelectRefs(source db.Source) ([]ExecRef, error) {
	ds := db.MustConform[db.SourcePgx](source)
	query := `
		select
			pool_id, title
		from pool_roots`
	rows, err := ds.Conn.Query(ds.Ctx, query)
	if err != nil {
		dao.log.Error("execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[execRefDS])
	if err != nil {
		dao.log.Error("collection failed", slog.Any("t", reflect.TypeOf(dtos)))
		return nil, err
	}
	refs, err := DataToExecRefs(dtos)
	if err != nil {
		dao.log.Error("conversion failed")
		return nil, err
	}
	return refs, nil
}

const (
	insertRoot = `
		insert into pool_roots (
			pool_id, title, proc_id, sup_pool_id, rev
		) values (
			@pool_id, @title, @proc_id, @sup_pool_id, @rev
		)`

	insertLiab = `
		insert into pool_liabs (
			pool_id, proc_id, rev
		) values (
			@pool_id, @proc_id, @rev
		)`

	insertBnd = `
		insert into pool_assets (
			pool_id, chnl_key, state_id, ex_pool_id, rev
		) values (
			@pool_id, @chnl_key, @state_id, @ex_pool_id, @rev
		)`

	insertStep = `
		insert into pool_steps (
			proc_id, chnl_id, kind, spec
		) values (
			@proc_id, @chnl_id, @kind, @spec
		)`

	updateRoot = `
		update pool_roots
		set rev = @rev + 1
		where pool_id = @pool_id
			and rev = @rev`

	selectOrgSnap = `
		select
			sup.pool_id,
			sup.title,
			jsonb_agg(json_build_object('pool_id', sub.pool_id, 'title', sub.title)) as subs
		from pool_roots sup
		left join pool_sups rel
			on rel.sup_pool_id = sup.pool_id
		left join pool_roots sub
			on sub.pool_id = rel.pool_id
			and sub.rev = rel.rev
		where sup.pool_id = $1
		group by sup.pool_id, sup.title`

	selectChnls = `
		with bnds as not materialized (
			select distinct on (chnl_ph)
				*
			from proc_bnds
			where proc_id = 'proc1'
			order by chnl_ph, abs(rev) desc
		), liabs as not materialized (
			select distinct on (proc_id)
				*
			from pool_liabs
			where proc_id = 'proc1'
			order by proc_id, abs(rev) desc
		)
		select
			bnd.*,
			prvd.pool_id
		from bnds bnd
		left join liabs liab
			on liab.proc_id = bnd.proc_id
		left join pool_roots prvd
			on prvd.pool_id = liab.pool_id`

	selectSteps = ``
)
