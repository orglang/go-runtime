package typedef

import (
	"errors"
	"log/slog"
	"math"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/db"
	"orglang/orglang/lib/lf"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
)

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

func (dao *pgxDAO) Insert(source db.Source, rec DefRec) error {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("defID", rec.DefID)
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entity insertion started", idAttr)
	dto, err := DataFromDefRec(rec)
	if err != nil {
		dao.log.Error("model conversion failed", idAttr)
		return err
	}
	insertRoot := `
		insert into type_def_roots (
			def_id, def_rn, title
		) values (
			@def_id, @def_rn, @title
		)`
	rootArgs := pgx.NamedArgs{
		"def_id": dto.DefID,
		"def_rn": dto.DefRN,
		"title":  dto.Title,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertRoot, rootArgs)
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", insertRoot))
		return err
	}
	insertState := `
		insert into type_term_states (
			def_id, exp_id, from_rn, to_rn
		) values (
			@def_id, @exp_id, @from_rn, @to_rn
		)`
	stateArgs := pgx.NamedArgs{
		"def_id":  dto.DefID,
		"from_rn": dto.DefRN,
		"to_rn":   math.MaxInt64,
		"exp_id":  dto.ExpID,
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertState, stateArgs)
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", insertState))
		return err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entity insertion succeed", idAttr)
	return nil
}

func (dao *pgxDAO) Update(source db.Source, rec DefRec) error {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("defID", rec.DefID)
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entity update started", idAttr)
	dto, err := DataFromDefRec(rec)
	if err != nil {
		dao.log.Error("model conversion failed", idAttr)
		return err
	}
	updateRoot := `
		update type_def_roots
		set def_rn = @def_rn,
			exp_id = @exp_id
		where def_id = @def_id
			and def_rn = @def_rn - 1`
	insertSnap := `
		insert into role_snaps (
			def_id, def_rn, title, exp_id
		) values (
			@def_id, @def_rn, @title, @exp_id
		)`
	args := pgx.NamedArgs{
		"def_id": dto.DefID,
		"def_rn": dto.DefRN,
		"title":  dto.Title,
		"exp_id": dto.ExpID,
	}
	ct, err := ds.Conn.Exec(ds.Ctx, updateRoot, args)
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", updateRoot))
		return err
	}
	if ct.RowsAffected() == 0 {
		dao.log.Error("entity update failed", idAttr)
		return errOptimisticUpdate(rec.DefRN - 1)
	}
	_, err = ds.Conn.Exec(ds.Ctx, insertSnap, args)
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", insertSnap))
		return err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entity update succeed", idAttr)
	return nil
}

func (dao *pgxDAO) SelectRefs(source db.Source) ([]DefRef, error) {
	ds := db.MustConform[db.SourcePgx](source)
	query := `
		SELECT
			def_id, def_rn, title
		FROM type_def_roots`
	rows, err := ds.Conn.Query(ds.Ctx, query)
	if err != nil {
		dao.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[defRefDS])
	if err != nil {
		dao.log.Error("rows collection failed")
		return nil, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entities selection succeed", slog.Any("dtos", dtos))
	return DataToDefRefs(dtos)
}

func (dao *pgxDAO) SelectRecByID(source db.Source, defID identity.ADT) (DefRec, error) {
	ds := db.MustConform[db.SourcePgx](source)
	idAttr := slog.Any("defID", defID)
	rows, err := ds.Conn.Query(ds.Ctx, selectById, defID.String())
	if err != nil {
		dao.log.Error("query execution failed", idAttr, slog.String("q", selectById))
		return DefRec{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[defRecDS])
	if err != nil {
		dao.log.Error("row collection failed", idAttr)
		return DefRec{}, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entity selection succeed", idAttr)
	return DataToDefRec(dto)
}

func (dao *pgxDAO) SelectRecByQN(source db.Source, typeQN qualsym.ADT) (DefRec, error) {
	ds := db.MustConform[db.SourcePgx](source)
	fqnAttr := slog.Any("typeQN", typeQN)
	rows, err := ds.Conn.Query(ds.Ctx, selectByFQN, qualsym.ConvertToString(typeQN))
	if err != nil {
		dao.log.Error("query execution failed", fqnAttr, slog.String("q", selectByFQN))
		return DefRec{}, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[defRecDS])
	if err != nil {
		dao.log.Error("row collection failed", fqnAttr)
		return DefRec{}, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entity selection succeed", fqnAttr)
	return DataToDefRec(dto)
}

func (dao *pgxDAO) SelectRecsByIDs(source db.Source, defIDs []identity.ADT) (_ []DefRec, err error) {
	ds := db.MustConform[db.SourcePgx](source)
	if len(defIDs) == 0 {
		return []DefRec{}, nil
	}
	query := `
		select
			def_id, def_rn, title, exp_id, whole_id
		from type_def_roots
		where def_id = $1`
	batch := pgx.Batch{}
	for _, defID := range defIDs {
		if defID.IsEmpty() {
			return nil, identity.ErrEmpty
		}
		batch.Queue(query, defID.String())
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var dtos []defRecDS
	for _, defID := range defIDs {
		rows, err := br.Query()
		if err != nil {
			dao.log.Error("query execution failed", slog.Any("defID", defID), slog.String("q", query))
		}
		defer rows.Close()
		dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[defRecDS])
		if err != nil {
			dao.log.Error("row collection failed", slog.Any("defID", defID))
		}
		dtos = append(dtos, dto)
	}
	if err != nil {
		return nil, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entities selection succeed", slog.Any("dtos", dtos))
	return DataToDefRecs(dtos)
}

func (dao *pgxDAO) SelectEnv(source db.Source, typeQNs []qualsym.ADT) (map[qualsym.ADT]DefRec, error) {
	recs, err := dao.SelectRecsByQNs(source, typeQNs)
	if err != nil {
		return nil, err
	}
	env := make(map[qualsym.ADT]DefRec, len(recs))
	for i, root := range recs {
		env[typeQNs[i]] = root
	}
	return env, nil
}

func (dao *pgxDAO) SelectRecsByQNs(source db.Source, typeQNs []qualsym.ADT) (_ []DefRec, err error) {
	ds := db.MustConform[db.SourcePgx](source)
	if len(typeQNs) == 0 {
		return []DefRec{}, nil
	}
	batch := pgx.Batch{}
	for _, defQN := range typeQNs {
		batch.Queue(selectByFQN, qualsym.ConvertToString(defQN))
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	var dtos []defRecDS
	for _, defQN := range typeQNs {
		rows, err := br.Query()
		if err != nil {
			dao.log.Error("query execution failed", slog.Any("defQN", defQN), slog.String("q", selectByFQN))
		}
		defer rows.Close()
		dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[defRecDS])
		if err != nil {
			dao.log.Error("row collection failed", slog.Any("defQN", defQN))
		}
		dtos = append(dtos, dto)
	}
	if err != nil {
		return nil, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entities selection succeed", slog.Any("dtos", dtos))
	return DataToDefRecs(dtos)
}

const (
	selectByFQN = `
		select
			rr.def_id,
			rr.def_rn,
			rr.title,
			rs.exp_id,
			null as whole_id
		from type_def_roots rr
		left join aliases a
			on a.id = rr.def_id
			and a.from_rn >= rr.def_rn
			and a.to_rn > rr.def_rn
		left join type_term_states rs
			on rs.def_id = rr.def_id
			and rs.from_rn >= rr.def_rn
			and rs.to_rn > rr.def_rn
		where a.sym = $1`

	selectById = `
		select
			rr.def_id,
			rr.def_rn,
			rr.title,
			rs.exp_id,
			null as whole_id
		from type_def_roots rr
		left join type_term_states rs
			on rs.def_id = rr.def_id
			and rs.from_rn >= rr.def_rn
			and rs.to_rn > rr.def_rn
		where rr.def_id = $1`
)
