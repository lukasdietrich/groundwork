package groundwork

import (
	"context"
	"database/sql"
	"fmt"
)

func Sqlite(db *sql.DB, tableName string) Dialect {
	return sqliteDialect{db, tableName}
}

type sqliteDialect struct {
	db        *sql.DB
	tableName string
}

func (d sqliteDialect) Setup(ctx context.Context) error {
	const query = `
		create table if not exists %q (
			"name" varchar  not null ,
			"hash" varchar  not null ,
			"time" datetime not null ,

			primary key ( "name" )
		) ;
	`

	return ignoreResult(d.db.ExecContext(ctx, fmt.Sprintf(query, d.tableName)))
}

func (d sqliteDialect) Begin(ctx context.Context) (Tx, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	return sqliteTx{tx, d.tableName}, err
}

type sqliteTx struct {
	*sql.Tx
	tableName string
}

func (t sqliteTx) Exec(ctx context.Context, query string) error {
	return ignoreResult(t.ExecContext(ctx, query))
}

func (t sqliteTx) Lookup(ctx context.Context, changeset Changeset) (*Changelog, error) {
	const query = `select "hash", "time" from %q where "name" = ? ;`

	changelog := Changelog{
		Name: changeset.Name(),
	}

	row := t.QueryRowContext(ctx, fmt.Sprintf(query, t.tableName), changeset.Name())
	return &changelog, row.Scan(&changelog.Hash, &changelog.Time)
}

func (t sqliteTx) Insert(ctx context.Context, changelog Changelog) error {
	const query = `insert into %q ( "name", "hash", "time" ) values ( ?, ?, ? ) ;`

	return ignoreResult(t.ExecContext(ctx, fmt.Sprintf(query, t.tableName),
		changelog.Name,
		changelog.Hash,
		changelog.Time,
	))
}

func ignoreResult(_ sql.Result, err error) error {
	return err
}
