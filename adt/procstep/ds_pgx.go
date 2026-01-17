package procstep

import (
	"errors"
	"log/slog"
	"reflect"

	"github.com/jackc/pgx/v5"

	"orglang/go-runtime/lib/db"
	"orglang/go-runtime/lib/lf"

	"orglang/go-runtime/adt/identity"
)

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

func (dao *pgxDAO) InsertRecs(source db.Source, roots ...StepRec) error {
	ds := db.MustConform[db.SourcePgx](source)
	dtos, err := DataFromSemRecs(roots)
	if err != nil {
		dao.log.Error("conversion failed")
		return err
	}
	batch := pgx.Batch{}
	for _, dto := range dtos {
		args := pgx.NamedArgs{
			"id":   dto.ID,
			"kind": dto.K,
			"pid":  dto.PID,
			"vid":  dto.VID,
			"spec": dto.ProcER,
		}
		batch.Queue(insertStep, args)
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	for _, dto := range dtos {
		_, err = br.Exec()
		if err != nil {
			dao.log.Error("execution failed", slog.Any("id", dto.ID), slog.String("q", insertStep))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (dao *pgxDAO) SelectRecByID(source db.Source, rid identity.ADT) (StepRec, error) {
	query := `
		SELECT
			id, kind, pid, vid, spec
		FROM steps
		WHERE id = $1`
	return dao.execute(source, query, rid.String())
}

func (dao *pgxDAO) execute(source db.Source, query string, arg string) (StepRec, error) {
	ds := db.MustConform[db.SourcePgx](source)
	rows, err := ds.Conn.Query(ds.Ctx, query, arg)
	if err != nil {
		dao.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[StepRecDS])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		dao.log.Error("row collection failed")
		return nil, err
	}
	root, err := dataToStepRec(dto)
	if err != nil {
		dao.log.Error("model conversion failed")
		return nil, err
	}
	dao.log.Log(ds.Ctx, lf.LevelTrace, "entity selection succeed", slog.Any("root", root))
	return root, nil
}

const (
	insertStep = `
		insert into pool_steps (
			id, kind, pid, vid, spec
		) values (
			@id, @kind, @pid, @vid, @spec
		)`
)
