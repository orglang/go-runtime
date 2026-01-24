package procexec

import (
	"errors"
	"log/slog"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/go-runtime/lib/db"

	"orglang/go-runtime/adt/procbind"
	"orglang/go-runtime/adt/procstep"
	"orglang/go-runtime/adt/revnum"
)

// Adapter
type pgxDAO struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo() Repo {
	return &pgxDAO{}
}

func newPgxDAO(l *slog.Logger) *pgxDAO {
	name := slog.String("name", reflect.TypeFor[pgxDAO]().Name())
	return &pgxDAO{l.With(name)}
}

func (dao *pgxDAO) SelectSnap(source db.Source, execRef ExecRef) (ExecSnap, error) {
	ds := db.MustConform[db.SourcePgx](source)
	refAttr := slog.Any("execRef", execRef)
	chnlRows, err := ds.Conn.Query(ds.Ctx, selectChnls, execRef.ID.String())
	if err != nil {
		dao.log.Error("execution failed", refAttr)
		return ExecSnap{}, err
	}
	defer chnlRows.Close()
	chnlDtos, err := pgx.CollectRows(chnlRows, pgx.RowToStructByName[procbind.BindRecDS])
	if err != nil {
		dao.log.Error("collection failed", refAttr, slog.Any("t", reflect.TypeOf(chnlDtos)))
		return ExecSnap{}, err
	}
	chnls, err := procbind.DataToBindRecs(chnlDtos)
	if err != nil {
		dao.log.Error("conversion failed", refAttr)
		return ExecSnap{}, err
	}
	stepRows, err := ds.Conn.Query(ds.Ctx, selectSteps, execRef.ID.String())
	if err != nil {
		dao.log.Error("execution failed", refAttr)
		return ExecSnap{}, err
	}
	defer stepRows.Close()
	stepDtos, err := pgx.CollectRows(stepRows, pgx.RowToStructByName[procstep.StepRecDS])
	if err != nil {
		dao.log.Error("collection failed", refAttr, slog.Any("t", reflect.TypeOf(stepDtos)))
		return ExecSnap{}, err
	}
	steps, err := procstep.DataToSemRecs(stepDtos)
	if err != nil {
		dao.log.Error("conversion failed", refAttr)
		return ExecSnap{}, err
	}
	dao.log.Debug("selection succeed", refAttr)
	return ExecSnap{
		ChnlBRs: procbind.IndexBy(ChnlPH, chnls),
		ProcSRs: procbind.IndexBy(procstep.ChnlID, steps),
	}, nil
}

func (dao *pgxDAO) UpdateProc(source db.Source, mod ExecMod) (err error) {
	if len(mod.Locks) == 0 {
		panic("empty locks")
	}
	ds := db.MustConform[db.SourcePgx](source)
	dto, err := DataFromMod(mod)
	if err != nil {
		dao.log.Error("conversion failed")
		return err
	}
	// binds
	bindReq := pgx.Batch{}
	for _, dto := range dto.Binds {
		args := pgx.NamedArgs{
			"exec_id":  dto.ExecID,
			"exec_rn":  dto.ExecRN,
			"chnl_ph":  dto.ChnlPH,
			"chnl_id":  dto.ChnlID,
			"state_id": dto.StateID,
		}
		bindReq.Queue(insertBind, args)
	}
	if bindReq.Len() > 0 {
		bindRes := ds.Conn.SendBatch(ds.Ctx, &bindReq)
		defer func() {
			err = errors.Join(err, bindRes.Close())
		}()
		for _, dto := range dto.Binds {
			_, err = bindRes.Exec()
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
			"kind":    dto.K,
			"proc_id": dto.ExecID,
			"chnl_id": dto.ChnlID,
			"proc_er": dto.ProcER,
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
	// execs
	execReq := pgx.Batch{}
	for _, dto := range dto.Locks {
		args := pgx.NamedArgs{
			"exec_id": dto.ID,
			"exec_rn": dto.RN,
		}
		execReq.Queue(updateExec, args)
	}
	execRes := ds.Conn.SendBatch(ds.Ctx, &execReq)
	defer func() {
		err = errors.Join(err, execRes.Close())
	}()
	for _, dto := range dto.Locks {
		ct, err := execRes.Exec()
		if err != nil {
			dao.log.Error("execution failed", slog.Any("dto", dto))
		}
		if ct.RowsAffected() == 0 {
			dao.log.Error("update failed")
			return errOptimisticUpdate(revnum.ADT(dto.RN))
		}
	}
	if err != nil {
		return err
	}
	dao.log.Debug("update succeed")
	return nil
}

const (
	insertBind = `
		insert into proc_binds (
			exec_id, chnl_ph, chnl_id, state_id, exec_rn
		) values (
			@exec_id, @chnl_ph, @chnl_id, @state_id, @exec_rn
		)`

	insertStep = `
		insert into proc_steps (
			exec_id, chnl_id, kind, proc_er
		) values (
			@exec_id, @chnl_id, @kind, @proc_er
		)`

	updateExec = `
		update proc_execs
		set exec_rn = @exec_rn + 1
		where exec_id = @exec_id
			and exec_rn = @exec_rn`

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
		left join pool_execs prvd
			on prvd.pool_id = liab.pool_id`

	selectSteps = ``
)
