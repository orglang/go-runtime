package procexec

import (
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"orglang/orglang/lib/db"
	"orglang/orglang/lib/lf"

	"orglang/orglang/adt/identity"
)

// Adapter
type daoPgx struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo() execRepo {
	return &daoPgx{}
}

func newDaoPgx(l *slog.Logger) *daoPgx {
	return &daoPgx{l}
}

func (d *daoPgx) SelectMain(db.Source, identity.ADT) (MainCfg, error) {
	return MainCfg{}, nil
}

func (d *daoPgx) UpdateMain(db.Source, MainMod) error {
	return nil
}

type daoPgx2 struct {
	log *slog.Logger
}

// for compilation purposes
func newRepo2() SemRepo {
	return &daoPgx2{}
}

func (d *daoPgx2) InsertSem(source db.Source, roots ...SemRec) error {
	ds := db.MustConform[db.SourcePgx](source)
	dtos, err := DataFromSemRecs(roots)
	if err != nil {
		d.log.Error("model mapping failed")
		return err
	}
	batch := pgx.Batch{}
	for _, dto := range dtos {
		args := pgx.NamedArgs{
			"id":   dto.ID,
			"kind": dto.K,
			"pid":  dto.PID,
			"vid":  dto.VID,
			"spec": dto.TR,
		}
		batch.Queue(insertRoot, args)
	}
	br := ds.Conn.SendBatch(ds.Ctx, &batch)
	defer func() {
		err = errors.Join(err, br.Close())
	}()
	for _, dto := range dtos {
		_, err = br.Exec()
		if err != nil {
			d.log.Error("query execution failed", slog.Any("id", dto.ID), slog.String("q", insertRoot))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *daoPgx2) SelectSemByID(source db.Source, rid identity.ADT) (SemRec, error) {
	query := `
		SELECT
			id, kind, pid, vid, spec
		FROM steps
		WHERE id = $1`
	return d.execute(source, query, rid.String())
}

func (d *daoPgx2) execute(source db.Source, query string, arg string) (SemRec, error) {
	ds := db.MustConform[db.SourcePgx](source)
	rows, err := ds.Conn.Query(ds.Ctx, query, arg)
	if err != nil {
		d.log.Error("query execution failed", slog.String("q", query))
		return nil, err
	}
	defer rows.Close()
	dto, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SemRecDS])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		d.log.Error("row collection failed")
		return nil, err
	}
	root, err := dataToSemRec(dto)
	if err != nil {
		d.log.Error("model mapping failed")
		return nil, err
	}
	d.log.Log(ds.Ctx, lf.LevelTrace, "entity selection succeed", slog.Any("root", root))
	return root, nil
}

const (
	insertRoot = `
		insert into pool_steps (
			id, kind, pid, vid, spec
		) values (
			@id, @kind, @pid, @vid, @spec
		)`
)
