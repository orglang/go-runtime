package poolstep

import (
	"errors"
	"log/slog"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/db"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/procexec"
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

func (d *pgxDAO) InsertLiab(source db.Source, liab procexec.Liab) (err error) {
	ds := db.MustConform[db.SourcePgx](source)
	dto := DataFromLiab(liab)
	args := pgx.NamedArgs{
		"pool_id": dto.PoolID,
		"proc_id": dto.ProcID,
		"rev":     dto.PoolRN,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertLiab, args)
	if err != nil {
		d.log.Error("execution failed")
		return err
	}
	d.log.Debug("insertion succeed", slog.Any("poolID", liab.PoolID))
	return nil
}

func (d *pgxDAO) SelectProc(source db.Source, procID identity.ADT) (procexec.Cfg, error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("procID", procID)
	chnlRows, err := ds.Conn.Query(ds.Ctx, selectChnls, procID.String())
	if err != nil {
		d.log.Error("execution failed", idAttr)
		return procexec.Cfg{}, err
	}
	defer chnlRows.Close()
	chnlDtos, err := pgx.CollectRows(chnlRows, pgx.RowToStructByName[epDS])
	if err != nil {
		d.log.Error("collection failed", idAttr, slog.Any("t", reflect.TypeOf(chnlDtos)))
		return procexec.Cfg{}, err
	}
	chnls, err := DataToEPs(chnlDtos)
	if err != nil {
		d.log.Error("conversion failed", idAttr)
		return procexec.Cfg{}, err
	}
	stepRows, err := ds.Conn.Query(ds.Ctx, selectSteps, procID.String())
	if err != nil {
		d.log.Error("execution failed", idAttr)
		return procexec.Cfg{}, err
	}
	defer stepRows.Close()
	stepDtos, err := pgx.CollectRows(stepRows, pgx.RowToStructByName[procexec.SemRecDS])
	if err != nil {
		d.log.Error("collection failed", idAttr, slog.Any("t", reflect.TypeOf(stepDtos)))
		return procexec.Cfg{}, err
	}
	steps, err := procexec.DataToSemRecs(stepDtos)
	if err != nil {
		d.log.Error("conversion failed", idAttr)
		return procexec.Cfg{}, err
	}
	d.log.Debug("selection succeed", idAttr)
	return procexec.Cfg{
		Chnls: termctx.IndexBy(procexec.ChnlPH, chnls),
		Steps: termctx.IndexBy(procexec.ChnlID, steps),
	}, nil
}

func (d *pgxDAO) UpdateProc(source db.Source, mod procexec.Mod) (err error) {
	if len(mod.Locks) == 0 {
		panic("empty locks")
	}
	ds := db.MustConform[db.SourcePgx](source)
	dto, err := procexec.DataFromMod(mod)
	if err != nil {
		d.log.Error("conversion failed")
		return err
	}
	// bindings
	bndReq := pgx.Batch{}
	for _, dto := range dto.Bnds {
		args := pgx.NamedArgs{
			"proc_id":  dto.ProcID,
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
		for _, dto := range dto.Bnds {
			_, err = bndRes.Exec()
			if err != nil {
				d.log.Error("execution failed", slog.Any("dto", dto))
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
			"spec":    dto.TR,
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
				d.log.Error("execution failed", slog.Any("dto", dto))
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
			d.log.Error("execution failed", slog.Any("dto", dto))
		}
		if ct.RowsAffected() == 0 {
			d.log.Error("update failed")
			return errOptimisticUpdate(revnum.ADT(dto.PoolRN))
		}
	}
	if err != nil {
		return err
	}
	d.log.Debug("update succeed")
	return nil
}

const (
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
