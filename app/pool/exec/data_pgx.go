package exec

import (
	"errors"
	"log/slog"
	"reflect"

	"github.com/jackc/pgx/v5"

	"smecalculus/rolevod/lib/core"
	"smecalculus/rolevod/lib/data"
	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/rn"

	procexec "smecalculus/rolevod/app/proc/exec"
)

// Adapter
type repoPgx struct {
	log *slog.Logger
}

func newRepoPgx(l *slog.Logger) *repoPgx {
	return &repoPgx{l}
}

// for compilation purposes
func newRepo() Repo {
	return &repoPgx{}
}

func (r *repoPgx) Insert(source data.Source, root PoolRec) (err error) {
	ds := data.MustConform[data.SourcePgx](source)
	dto := DataFromPoolRec(root)
	args := pgx.NamedArgs{
		"pool_id":     dto.PoolID,
		"proc_id":     dto.ProcID,
		"sup_pool_id": dto.SupID,
		"rev":         dto.PoolRN,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertRoot, args)
	if err != nil {
		r.log.Error("execution failed")
		return err
	}
	r.log.Debug("insertion succeeded", slog.Any("poolID", root.ExecID))
	return nil
}

func (r *repoPgx) InsertLiab(source data.Source, liab procexec.Liab) (err error) {
	ds := data.MustConform[data.SourcePgx](source)
	dto := DataFromLiab(liab)
	args := pgx.NamedArgs{
		"pool_id": dto.PoolID,
		"proc_id": dto.ProcID,
		"rev":     dto.PoolRN,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertLiab, args)
	if err != nil {
		r.log.Error("execution failed")
		return err
	}
	r.log.Debug("insertion succeeded", slog.Any("poolID", liab.PoolID))
	return nil
}

func (r *repoPgx) SelectProc(source data.Source, procID id.ADT) (procexec.Cfg, error) {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("procID", procID)
	chnlRows, err := ds.Conn.Query(ds.Ctx, selectChnls, procID.String())
	if err != nil {
		r.log.Error("execution failed", idAttr)
		return procexec.Cfg{}, err
	}
	defer chnlRows.Close()
	chnlDtos, err := pgx.CollectRows(chnlRows, pgx.RowToStructByName[epData])
	if err != nil {
		r.log.Error("collection failed", idAttr, slog.Any("t", reflect.TypeOf(chnlDtos)))
		return procexec.Cfg{}, err
	}
	chnls, err := DataToEPs(chnlDtos)
	if err != nil {
		r.log.Error("mapping failed", idAttr)
		return procexec.Cfg{}, err
	}
	stepRows, err := ds.Conn.Query(ds.Ctx, selectSteps, procID.String())
	if err != nil {
		r.log.Error("execution failed", idAttr)
		return procexec.Cfg{}, err
	}
	defer stepRows.Close()
	stepDtos, err := pgx.CollectRows(stepRows, pgx.RowToStructByName[procexec.SemRecData])
	if err != nil {
		r.log.Error("collection failed", idAttr, slog.Any("t", reflect.TypeOf(stepDtos)))
		return procexec.Cfg{}, err
	}
	steps, err := procexec.DataToSemRecs(stepDtos)
	if err != nil {
		r.log.Error("mapping failed", idAttr)
		return procexec.Cfg{}, err
	}
	r.log.Debug("selection succeeded", idAttr)
	return procexec.Cfg{
		Chnls: core.IndexBy(procexec.ChnlPH, chnls),
		Steps: core.IndexBy(procexec.ChnlID, steps),
	}, nil
}

func (r *repoPgx) UpdateProc(source data.Source, mod procexec.Mod) (err error) {
	if len(mod.Locks) == 0 {
		panic("empty locks")
	}
	ds := data.MustConform[data.SourcePgx](source)
	dto, err := procexec.DataFromMod(mod)
	if err != nil {
		r.log.Error("mapping failed")
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
				r.log.Error("execution failed", slog.Any("dto", dto))
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
				r.log.Error("execution failed", slog.Any("dto", dto))
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
			r.log.Error("execution failed", slog.Any("dto", dto))
		}
		if ct.RowsAffected() == 0 {
			r.log.Error("update failed")
			return errOptimisticUpdate(rn.ADT(dto.PoolRN))
		}
	}
	if err != nil {
		return err
	}
	r.log.Debug("update succeeded")
	return nil
}

func (r *repoPgx) SelectSubs(source data.Source, poolID id.ADT) (PoolSnap, error) {
	ds := data.MustConform[data.SourcePgx](source)
	idAttr := slog.Any("poolID", poolID)
	rows, err := ds.Conn.Query(ds.Ctx, selectOrgSnap, poolID.String())
	if err != nil {
		r.log.Error("execution failed", idAttr)
		return PoolSnap{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[poolSnapData])
	if err != nil {
		r.log.Error("collection failed", idAttr, slog.Any("struct", reflect.TypeOf(dto)))
		return PoolSnap{}, err
	}
	snap, err := DataToPoolSnap(dto)
	if err != nil {
		r.log.Error("mapping failed")
		return PoolSnap{}, err
	}
	r.log.Debug("selection succeeded", idAttr)
	return snap, nil
}

func (r *repoPgx) SelectRefs(source data.Source) ([]PoolRef, error) {
	ds := data.MustConform[data.SourcePgx](source)
	query := `
		select
			pool_id, title
		from pool_roots`
	rows, err := ds.Conn.Query(ds.Ctx, query)
	if err != nil {
		r.log.Error("execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[poolRefData])
	if err != nil {
		r.log.Error("collection failed", slog.Any("t", reflect.TypeOf(dtos)))
		return nil, err
	}
	refs, err := DataToPoolRefs(dtos)
	if err != nil {
		r.log.Error("mapping failed")
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
