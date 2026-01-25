package poolexec

import (
	"log/slog"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/go-runtime/lib/db"
)

type pgxDAO struct {
	log *slog.Logger
}

func newPgxDAO(log *slog.Logger) *pgxDAO {
	name := slog.String("name", reflect.TypeFor[pgxDAO]().Name())
	return &pgxDAO{log.With(name)}
}

// for compilation purposes
func newRepo() Repo {
	return &pgxDAO{}
}

func (dao *pgxDAO) InsertRec(source db.Source, rec ExecRec) (err error) {
	ds := db.MustConform[db.SourcePgx](source)
	dto := DataFromExecRec(rec)
	args := pgx.NamedArgs{
		"exec_id":     dto.ID,
		"exec_rn":     dto.RN,
		"sup_exec_id": dto.SupID,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertExec, args)
	if err != nil {
		dao.log.Error("execution failed")
		return err
	}
	dao.log.Debug("insertion succeed", slog.Any("execRef", rec.ExecRef))
	return nil
}

func (dao *pgxDAO) InsertLiab(source db.Source, liab Liab) (err error) {
	ds := db.MustConform[db.SourcePgx](source)
	dto := DataFromLiab(liab)
	args := pgx.NamedArgs{
		"exec_id": dto.ID,
		"exec_rn": dto.RN,
		"proc_id": dto.ProcID,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertLiab, args)
	if err != nil {
		dao.log.Error("execution failed")
		return err
	}
	dao.log.Debug("insertion succeed", slog.Any("execRef", liab.ExecRef))
	return nil
}

func (dao *pgxDAO) SelectSubs(source db.Source, ref ExecRef) (ExecSnap, error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("execRef", ref)
	rows, err := ds.Conn.Query(ds.Ctx, selectOrgSnap, ref.ID.String())
	if err != nil {
		dao.log.Error("execution failed", idAttr)
		return ExecSnap{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[execSnapDS])
	if err != nil {
		dao.log.Error("collection failed", idAttr, slog.Any("t", reflect.TypeOf(dto)))
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
			exec_id, title
		from pool_execs`
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
	insertExec = `
		insert into pool_execs (
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

	insertBind = `
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
